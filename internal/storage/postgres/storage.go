// Package postgres provides PostgreSQL 17 persistence and real-time LISTEN/NOTIFY in-memory cache invalidation.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"bkam-api/internal/config"
	"bkam-api/internal/domain"
	"bkam-api/internal/engine"
)

// Storage manages PostgreSQL connection pooling, cold boot hydration, and real-time LISTEN/NOTIFY invalidation.
type Storage struct {
	pool        *pgxpool.Pool
	cache       *MemoryCache
	coordinator *engine.LiveSnapshotCoordinator
	processor   *RatesProcessor
	connStr     string

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewStorage initializes a new PostgreSQL storage instance with a configured pgx connection pool.
func NewStorage(ctx context.Context, cfg *config.Config, coordinator *engine.LiveSnapshotCoordinator) (*Storage, error) {
	connStr := cfg.GetDatabaseConnString()
	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid postgres connection config: %w", err)
	}

	// Architectural pool constraints
	poolCfg.MaxConns = 15
	poolCfg.MinConns = 2
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}

	// Ping with 3-second timeout to verify connectivity
	pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
	defer pingCancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to postgres (%s): %w", cfg.DBHost, err)
	}

	return &Storage{
		pool:        pool,
		cache:       NewMemoryCache(),
		coordinator: coordinator,
		processor:   NewRatesProcessor(),
		connStr:     connStr,
	}, nil
}

// NewStorageWithPool constructs a Storage instance wrapping an existing connection pool (useful for tests).
func NewStorageWithPool(pool *pgxpool.Pool, coordinator *engine.LiveSnapshotCoordinator) *Storage {
	return &Storage{
		pool:        pool,
		cache:       NewMemoryCache(),
		coordinator: coordinator,
		processor:   NewRatesProcessor(),
	}
}

// GetCache returns the thread-safe in-memory cache.
func (s *Storage) GetCache() *MemoryCache {
	return s.cache
}

// GetPool returns the underlying pgxpool.Pool.
func (s *Storage) GetPool() *pgxpool.Pool {
	return s.pool
}

// Hydrate populates the in-memory cache and coordinator on API boot by querying all rows from live_snapshots.
func (s *Storage) Hydrate(ctx context.Context) error {
	if s.pool == nil {
		return errors.New("postgres connection pool is nil")
	}

	query := `SELECT domain, sub_key, payload, updated_at FROM live_snapshots`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query live_snapshots for hydration: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var domainName, subKey string
		var payload []byte
		var updatedAt time.Time

		if err := rows.Scan(&domainName, &subKey, &payload, &updatedAt); err != nil {
			log.Printf("[PostgresStorage] Warning: Failed to scan live_snapshots row: %v", err)
			continue
		}

		s.cache.Set(&LiveSnapshot{
			Domain:    domainName,
			SubKey:    subKey,
			Payload:   payload,
			UpdatedAt: updatedAt,
		})

		s.dispatchToCoordinator(domainName, subKey, payload, updatedAt)
		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating live_snapshots rows: %w", err)
	}

	log.Printf("[PostgresStorage] Hydrated %d snapshot(s) from live_snapshots into memory cache.", count)
	return nil
}

// StartListener begins listening for 'snapshot_update' asynchronous PostgreSQL notifications.
func (s *Storage) StartListener(parentCtx context.Context) {
	if s.pool == nil {
		log.Println("[PostgresListener] Skipping listener start: connection pool is nil.")
		return
	}

	ctx, cancel := context.WithCancel(parentCtx)
	s.ctx = ctx
	s.cancel = cancel

	s.wg.Add(1)
	go s.runListenerLoop(ctx)
}

// Stop terminates the listener goroutine and closes the connection pool.
func (s *Storage) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	if s.pool != nil {
		s.pool.Close()
	}
	log.Println("[PostgresStorage] Stopped listener and closed connection pool cleanly.")
}

// runListenerLoop manages a dedicated connection to LISTEN snapshot_update with exponential backoff reconnection.
func (s *Storage) runListenerLoop(ctx context.Context) {
	defer s.wg.Done()

	backoff := 500 * time.Millisecond
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Acquire dedicated connection for LISTEN
		conn, err := s.pool.Acquire(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[PostgresListener] Acquire connection error: %v; retrying in %v...", err, backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
				continue
			}
		}

		_, err = conn.Exec(ctx, "LISTEN snapshot_update")
		if err != nil {
			conn.Release()
			if ctx.Err() != nil {
				return
			}
			log.Printf("[PostgresListener] Failed to LISTEN snapshot_update: %v; retrying in %v...", err, backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
				continue
			}
		}

		// Reset backoff on successful subscription
		backoff = 500 * time.Millisecond
		log.Println("[PostgresListener] Subscribed to 'snapshot_update' notifications.")

		// Notification listening loop
		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				conn.Release()
				if ctx.Err() != nil {
					return
				}
				log.Printf("[PostgresListener] WaitForNotification error: %v; reconnecting...", err)
				break
			}

			var notif struct {
				Domain string `json:"domain"`
				SubKey string `json:"sub_key"`
			}
			if err := json.Unmarshal([]byte(notification.Payload), &notif); err != nil {
				log.Printf("[PostgresListener] Malformed notification payload '%s': %v", notification.Payload, err)
				continue
			}

			// Invalidate and refresh cache for this domain/sub_key
			if err := s.refreshSnapshot(ctx, notif.Domain, notif.SubKey); err != nil {
				log.Printf("[PostgresListener] Error refreshing snapshot (domain=%s, sub_key=%s): %v", notif.Domain, notif.SubKey, err)
			} else {
				log.Printf("[PostgresListener] Real-time invalidation applied: domain=%s, sub_key=%s", notif.Domain, notif.SubKey)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			backoff = min(backoff*2, maxBackoff)
		}
	}
}

// refreshSnapshot fetches the latest payload for (domain, sub_key) and updates the in-memory cache.
func (s *Storage) refreshSnapshot(ctx context.Context, domainName, subKey string) error {
	query := `SELECT domain, sub_key, payload, updated_at FROM live_snapshots WHERE domain = $1 AND sub_key = $2`
	row := s.pool.QueryRow(ctx, query, domainName, subKey)

	var d, sk string
	var payload []byte
	var updatedAt time.Time

	if err := row.Scan(&d, &sk, &payload, &updatedAt); err != nil {
		return err
	}

	s.cache.Set(&LiveSnapshot{
		Domain:    d,
		SubKey:    sk,
		Payload:   payload,
		UpdatedAt: updatedAt,
	})

	s.dispatchToCoordinator(d, sk, payload, updatedAt)
	return nil
}

// dispatchToCoordinator decodes the JSONB payload and populates domain snapshot structs in the coordinator.
func (s *Storage) dispatchToCoordinator(domainName, subKey string, payload []byte, updatedAt time.Time) {
	if s.coordinator == nil {
		return
	}

	normDomain := strings.ToLower(strings.TrimSpace(domainName))
	normSubKey := strings.ToLower(strings.TrimSpace(subKey))

	switch normDomain {
	case "rates":
		s.dispatchRates(payload, updatedAt)
	case "gold":
		s.dispatchGold(payload, updatedAt)
	case "gas":
		s.dispatchGas(payload, updatedAt)
	case "stocks":
		s.dispatchStocks(subKey, payload, updatedAt)
	case "indices":
		s.dispatchIndices(subKey, payload, updatedAt)
	default:
		// Fallback: store raw unmarshaled generic structure
		var raw any
		if json.Unmarshal(payload, &raw) == nil {
			s.coordinator.UpdateSnapshot(normDomain, raw, normSubKey)
		}
	}
}

func (s *Storage) dispatchRates(payload []byte, updatedAt time.Time) {
	// Attempt parsing as pre-built RateSnapshot envelope
	var directSnap domain.RateSnapshot
	if err := json.Unmarshal(payload, &directSnap); err == nil && len(directSnap.Rates) > 0 {
		if directSnap.RequestedAt.IsZero() {
			directSnap.RequestedAt = updatedAt
		}
		s.coordinator.UpdateSnapshot("rates", &directSnap)
		return
	}

	// Parse as map of bank codes to rate records
	var bankMap map[string]RawBankDocument
	if err := json.Unmarshal(payload, &bankMap); err == nil && len(bankMap) > 0 {
		docs := make([]RawBankDocument, 0, len(bankMap))
		for code, doc := range bankMap {
			if doc.BankID == "" && doc.BankCode == "" {
				doc.BankID = code
			} else if doc.BankID == "" {
				doc.BankID = doc.BankCode
			}
			if doc.Status == "" {
				doc.Status = "SUCCESS"
			}
			docs = append(docs, doc)
		}
		snap := s.processor.ProcessRawDocuments(docs, fmt.Sprintf("pg-%d", updatedAt.UnixNano()))
		snap.RequestedAt = updatedAt
		s.coordinator.UpdateSnapshot("rates", snap)
	}
}

func (s *Storage) dispatchGold(payload []byte, updatedAt time.Time) {
	// Attempt pre-built GoldSnapshot envelope
	var directSnap domain.GoldSnapshot
	if err := json.Unmarshal(payload, &directSnap); err == nil && len(directSnap.Rates) > 0 {
		if directSnap.UpdatedAt.IsZero() {
			directSnap.UpdatedAt = updatedAt
		}
		s.coordinator.UpdateSnapshot("gold", &directSnap)
		return
	}

	// Parse scraper payload: {"lastUpdated": "...", "carats": {"24k": {"buy": 4200, "sell": 4250}}, "spotUsd": 2650.5}
	type scraperGold struct {
		LastUpdated any `json:"lastUpdated"`
		Carats      map[string]struct {
			Buy     float64  `json:"buy"`
			Sell    float64  `json:"sell"`
			OldBuy  *float64 `json:"oldBuy,omitempty"`
			OldSell *float64 `json:"oldSell,omitempty"`
		} `json:"carats"`
		SpotUSD *float64 `json:"spotUsd,omitempty"`
	}

	var sg scraperGold
	if err := json.Unmarshal(payload, &sg); err == nil && len(sg.Carats) > 0 {
		rates := make(map[string]domain.GoldKaratPrice, len(sg.Carats))
		for karat, price := range sg.Carats {
			trend := domain.TrendUp
			if price.OldBuy != nil && price.Buy < *price.OldBuy {
				trend = domain.TrendDown
			}
			rates[karat] = domain.GoldKaratPrice{
				Buy:     price.Buy,
				Sell:    price.Sell,
				OldBuy:  price.OldBuy,
				OldSell: price.OldSell,
				Trend:   trend,
			}
		}
		snap := &domain.GoldSnapshot{
			Source:        "https://market.isagha.com/prices",
			Currency:      "EGP",
			Logo:          "https://bekamfeelbank.web.app/images/gold.png",
			UpdatedAt:     updatedAt,
			ScrapedAt:     updatedAt,
			Freshness:     "current",
			SourceTraceID: fmt.Sprintf("pg-%d", updatedAt.UnixNano()),
			Rates:         rates,
		}
		if sg.SpotUSD != nil {
			snap.International = &domain.InternationalGoldQuote{
				Price:    *sg.SpotUSD,
				Currency: "USD",
			}
		}
		s.coordinator.UpdateSnapshot("gold", snap)
	}
}

func (s *Storage) dispatchGas(payload []byte, updatedAt time.Time) {
	// Attempt pre-built GasSnapshot envelope
	var directSnap domain.GasSnapshot
	if err := json.Unmarshal(payload, &directSnap); err == nil && len(directSnap.Rates) > 0 {
		if directSnap.UpdatedAt.IsZero() {
			directSnap.UpdatedAt = updatedAt
		}
		s.coordinator.UpdateSnapshot("gas", &directSnap)
		return
	}

	// Parse scraper payload: {"lastUpdated": "...", "fuels": {"octane92": 15.25, "diesel": 13.50}}
	type scraperGas struct {
		LastUpdated any            `json:"lastUpdated"`
		Fuels       map[string]any `json:"fuels"`
	}

	var sg scraperGas
	if err := json.Unmarshal(payload, &sg); err == nil && len(sg.Fuels) > 0 {
		rates := make(map[string]domain.GasPriceItem, len(sg.Fuels))
		for k, v := range sg.Fuels {
			var price float64
			switch val := v.(type) {
			case float64:
				price = val
			case map[string]any:
				if p, ok := val["price"].(float64); ok {
					price = p
				}
			}
			rates[k] = domain.GasPriceItem{
				Price: price,
				Unit:  "liter",
				Trend: domain.TrendUp,
			}
		}
		snap := &domain.GasSnapshot{
			Source:        "https://www.petroleum.gov.eg",
			Currency:      "EGP",
			Logo:          "https://bekamfeelbank.web.app/images/gas.png",
			UpdatedAt:     updatedAt,
			ScrapedAt:     updatedAt,
			Freshness:     "current",
			SourceTraceID: fmt.Sprintf("pg-%d", updatedAt.UnixNano()),
			Rates:         rates,
		}
		s.coordinator.UpdateSnapshot("gas", snap)
	}
}

func (s *Storage) dispatchStocks(subKey string, payload []byte, updatedAt time.Time) {
	market := normalizeMarket(subKey)

	// Attempt pre-built StockSnapshot envelope
	var directSnap domain.StockSnapshot
	if err := json.Unmarshal(payload, &directSnap); err == nil && len(directSnap.Stocks) > 0 {
		if directSnap.UpdatedAt.IsZero() {
			directSnap.UpdatedAt = updatedAt
		}
		if directSnap.Market != "" {
			market = normalizeMarket(directSnap.Market)
		}
		s.routeStockSnapshot(&directSnap, market, subKey)
		return
	}

	// Parse scraper payload: {"market": "EGX", "lastUpdated": "...", "quotes": [...]}
	type scraperStocks struct {
		Market      string                  `json:"market"`
		LastUpdated any                     `json:"lastUpdated"`
		Quotes      []domain.StockPriceItem `json:"quotes"`
	}

	var ss scraperStocks
	if err := json.Unmarshal(payload, &ss); err == nil && len(ss.Quotes) > 0 {
		if ss.Market != "" {
			market = normalizeMarket(ss.Market)
		}
		stocksMap := make(map[string]domain.StockPriceItem, len(ss.Quotes))
		for _, q := range ss.Quotes {
			sym := q.Symbol
			if sym == "" {
				sym = q.Code
			}
			if sym != "" {
				stocksMap[sym] = q
			}
		}

		currency := "EGP"
		switch market {
		case "saudi":
			currency = "SAR"
		case "uae":
			currency = "AED"
		case "qatar":
			currency = "QAR"
		case "kuwait":
			currency = "KWD"
		}

		snap := &domain.StockSnapshot{
			Source:        "bkam-gateway",
			Market:        market,
			Currency:      currency,
			UpdatedAt:     updatedAt,
			ScrapedAt:     updatedAt,
			Freshness:     "current",
			SourceTraceID: fmt.Sprintf("pg-%d", updatedAt.UnixNano()),
			Stocks:        stocksMap,
		}
		s.routeStockSnapshot(snap, market, subKey)
	}
}

func (s *Storage) routeStockSnapshot(snap *domain.StockSnapshot, market, rawSubKey string) {
	if market == "egx" {
		s.coordinator.UpdateSnapshot("stocks", snap, "egx")
		return
	}

	// Regional GCC equities
	s.coordinator.UpdateSnapshot("gcc-stocks", snap, market)

	// Register aliases for regional tickers (e.g. tadawul, adx, dfm, qse, kse)
	normRaw := strings.ToLower(strings.TrimSpace(rawSubKey))
	if normRaw != "" && normRaw != market {
		s.coordinator.UpdateSnapshot("gcc-stocks", snap, normRaw)
	}
}

func (s *Storage) dispatchIndices(subKey string, payload []byte, updatedAt time.Time) {
	market := normalizeMarket(subKey)

	var directSnap domain.IndexSnapshot
	if err := json.Unmarshal(payload, &directSnap); err == nil && len(directSnap.Indices) > 0 {
		if directSnap.UpdatedAt.IsZero() {
			directSnap.UpdatedAt = updatedAt
		}
		s.coordinator.UpdateSnapshot("indices", &directSnap, market)
		return
	}

	type scraperIndices struct {
		Market  string                  `json:"market"`
		Indices []domain.IndexPriceItem `json:"indices"`
	}

	var si scraperIndices
	if err := json.Unmarshal(payload, &si); err == nil && len(si.Indices) > 0 {
		idxMap := make(map[string]domain.IndexPriceItem, len(si.Indices))
		for _, item := range si.Indices {
			code := item.Code
			if code == "" {
				code = item.IndexCode
			}
			if code != "" {
				idxMap[code] = item
			}
		}
		snap := &domain.IndexSnapshot{
			Source:        "bkam-gateway",
			Market:        market,
			UpdatedAt:     updatedAt,
			ScrapedAt:     updatedAt,
			Freshness:     "current",
			SourceTraceID: fmt.Sprintf("pg-%d", updatedAt.UnixNano()),
			Indices:       idxMap,
		}
		s.coordinator.UpdateSnapshot("indices", snap, market)
	}
}

func normalizeMarket(m string) string {
	lower := strings.ToLower(strings.TrimSpace(m))
	switch lower {
	case "egx", "egypt":
		return "egx"
	case "tadawul", "saudi", "ksa":
		return "saudi"
	case "adx", "dfm", "uae":
		return "uae"
	case "qse", "qatar":
		return "qatar"
	case "kse", "kuwait":
		return "kuwait"
	default:
		if lower == "" {
			return "egx"
		}
		return lower
	}
}
