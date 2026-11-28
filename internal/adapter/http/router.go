// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library and external Chi router packages.
import (
	// net/http provides HTTP request and handler primitives.
	"net/http"
	// strings provides string inspection for metric classification.
	"strings"

	// chi provides lightweight, high-performance HTTP routing.
	"github.com/go-chi/chi/v5"
	// chi middleware provides standard recoverer, request ID, and logger.
	chimw "github.com/go-chi/chi/v5/middleware"

	// engine provides domain coordinators and schedule calculators.
	"bkam-api/internal/engine"
)

// RouterDependencies aggregates all handlers and engines needed to assemble the HTTP gateway router.
type RouterDependencies struct {
	// APIKey is the configured platform secret for API key authentication.
	APIKey string
	// Coordinator supplies in-memory live quote envelopes.
	Coordinator *engine.LiveSnapshotCoordinator
	// MarketSchedule supplies dynamic TTL and market session evaluations.
	MarketSchedule *engine.MarketScheduleEngine
	// QuotesHandler handles multi-domain batch quote queries.
	QuotesHandler *QuotesHandler
	// RatesHandler handles FX rates and catalog endpoints.
	RatesHandler *RatesHandler
	// CommoditiesHandler handles gold and gas endpoints.
	CommoditiesHandler *CommoditiesHandler
	// StockHandler handles domestic and regional equity endpoints.
	StockHandler *StockHandler
	// IndicesHandler handles market benchmark index endpoints.
	IndicesHandler *IndicesHandler
	// HealthHandler handles health assessment probes.
	HealthHandler *HealthHandler
	// VersionHandler handles version discovery endpoints.
	VersionHandler *VersionHandler
	// MetricsHandler handles operational request analytics.
	MetricsHandler *MetricsHandler
}

// NewRouter constructs, configures, and returns the root chi.Mux router with all platform routes registered.
func NewRouter(deps *RouterDependencies) *chi.Mux {
	// Allocate new Chi router mux.
	r := chi.NewRouter()

	// Register standard panic recovery middleware.
	r.Use(chimw.Recoverer)

	// Register distributed tracing middleware.
	r.Use(TraceMiddleware)

	// Register ETag revalidation middleware.
	r.Use(ETagMiddleware)

	// Register operational metrics tracking middleware if MetricsHandler configured.
	if deps.MetricsHandler != nil {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				path := strings.ToLower(req.URL.Path)
				cat := "other"
				if strings.HasPrefix(path, "/v1/rates") || strings.HasPrefix(path, "/v1/banks") {
					cat = "banks"
				} else if strings.HasPrefix(path, "/v1/gold") {
					cat = "gold"
				} else if strings.HasPrefix(path, "/v1/stocks") {
					cat = "stocks"
				} else if strings.HasPrefix(path, "/v1/indices") {
					cat = "indices"
				} else if strings.HasPrefix(path, "/v1/gas") {
					cat = "gas"
				}
				deps.MetricsHandler.RecordRequest(cat)
				next.ServeHTTP(w, req)
			})
		})
	}

	// Register API key authentication middleware.
	if deps.APIKey != "" {
		r.Use(AuthMiddleware(deps.APIKey))
	}

	// System Health check endpoint.
	if deps.HealthHandler != nil {
		r.Get("/v1/health", deps.HealthHandler.GetHealth)
	}

	// System Version discovery endpoints.
	if deps.VersionHandler != nil {
		r.Get("/v1/version", deps.VersionHandler.GetVersion)
		r.Get("/version", deps.VersionHandler.GetVersion)
	}

	// Operational Metrics analytics endpoint.
	if deps.MetricsHandler != nil {
		r.Get("/v1/metrics", deps.MetricsHandler.GetMetrics)
	}

	// API Discovery and Platform Catalog.
	r.Get("/v1/catalog", deps.RatesHandler.GetCatalog)

	// Multi-Domain Batch Quotes endpoints.
	if deps.QuotesHandler != nil {
		// POST /v1/quotes JSON batch.
		r.Post("/v1/quotes", deps.QuotesHandler.PostQuotes)
		// GET /v1/quotes query parameter batch.
		r.Get("/v1/quotes", deps.QuotesHandler.GetQuotes)
	}

	// Foreign Exchange Rates endpoints.
	r.Get("/v1/rates", deps.RatesHandler.GetRates)
	r.Get("/v1/rates/history/range", deps.RatesHandler.GetRatesHistoryRange)
	r.Get("/v1/rates/history/{date}", deps.RatesHandler.GetRatesHistoryByDate)
	r.Get("/v1/rates/{currency}/intraday/latest", deps.RatesHandler.GetRatesIntradayLatest)
	r.Get("/v1/rates/{currency}", deps.RatesHandler.GetRatesByCurrency)
	r.Get("/v1/banks/{bankCode}/rates", deps.RatesHandler.GetRatesByBank)
	// Bank historical range endpoint.
	r.Get("/v1/banks/{bankCode}/history/range", deps.RatesHandler.GetRatesHistoryRange)
	// Bank historical date endpoint.
	r.Get("/v1/banks/{bankCode}/history/{date}", deps.RatesHandler.GetRatesHistoryByDate)
	// Bank history default alias endpoint.
	r.Get("/v1/banks/{bankCode}/history", deps.RatesHandler.GetRatesHistoryRange)

	// Commodities endpoints (Gold and Petroleum).
	r.Get("/v1/gold", deps.CommoditiesHandler.GetGold)
	r.Get("/v1/gold/history/range", deps.CommoditiesHandler.GetGoldHistoryRange)
	r.Get("/v1/gold/history/{date}", deps.CommoditiesHandler.GetGoldHistoryByDate)
	r.Get("/v1/gold/intraday/latest", deps.CommoditiesHandler.GetGoldIntradayLatest)
	r.Get("/v1/gold/{karat}/intraday/latest", deps.CommoditiesHandler.GetGoldKaratIntradayLatest)
	r.Get("/v1/gas", deps.CommoditiesHandler.GetGas)

	// Equities endpoints (Domestic EGX and Regional GCC).
	r.Get("/v1/stocks", deps.StockHandler.GetEGXStocks)
	r.Get("/v1/stocks/regional/catalog", deps.StockHandler.GetRegionalCatalog)
	// Domestic stock market history range.
	r.Get("/v1/stocks/history/range", deps.StockHandler.GetStockHistoryRange)
	r.Get("/v1/stocks/history/{date}", deps.StockHandler.GetStockHistoryByDate)
	// Domestic stock single symbol range or regional market-wide range.
	r.Get("/v1/stocks/{symbol}/history/range", deps.StockHandler.GetStockHistoryRange)
	r.Get("/v1/stocks/{symbol}/intraday/latest", deps.StockHandler.GetStockIntradayLatest)
	r.Get("/v1/stocks/{market}/history/{date}", deps.StockHandler.GetRegionalStockHistoryByDate)
	r.Get("/v1/stocks/{market}/{symbol}/history/range", deps.StockHandler.GetRegionalStockHistoryRange)
	r.Get("/v1/stocks/saudi", deps.StockHandler.GetSaudiStocks)
	r.Get("/v1/stocks/uae", deps.StockHandler.GetUAEStocks)
	r.Get("/v1/stocks/qatar", deps.StockHandler.GetQatarStocks)
	r.Get("/v1/stocks/kuwait", deps.StockHandler.GetKuwaitStocks)

	// Benchmark Indices endpoints.
	r.Get("/v1/indices", deps.IndicesHandler.GetIndices)
	r.Get("/v1/indices/history/range", deps.IndicesHandler.GetIndicesHistoryRange)
	r.Get("/v1/indices/history/{date}", deps.IndicesHandler.GetIndicesHistoryByDate)
	r.Get("/v1/indices/{symbol}/history/range", deps.IndicesHandler.GetIndicesHistoryRange)
	r.Get("/v1/indices/{code}/stocks", deps.IndicesHandler.GetConstituentStocks)
	r.Get("/v1/indices/{code}/constituents", deps.IndicesHandler.GetConstituentStocks)
	r.Get("/v1/indices/{market}", deps.IndicesHandler.GetMarketIndices)

	// API route aliases matching /v1 endpoints
	r.Route("/api", func(api chi.Router) {
		api.Get("/rates", deps.RatesHandler.GetRates)
		api.Get("/rates/{currency}", deps.RatesHandler.GetRatesByCurrency)
		api.Get("/gold", deps.CommoditiesHandler.GetGold)
		api.Get("/gas", deps.CommoditiesHandler.GetGas)
		api.Get("/stocks", deps.StockHandler.GetEGXStocks)
		api.Get("/indices", deps.IndicesHandler.GetIndices)
		api.Get("/catalog", deps.RatesHandler.GetCatalog)
		api.Get("/health", deps.HealthHandler.GetHealth)
		api.Get("/version", deps.VersionHandler.GetVersion)
	})

	// Catch-all 404 handler for undefined endpoints matching API error schema.
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		// Respond with 404 NOT_FOUND JSON envelope.
		RespondError(w, http.StatusNotFound, "NOT_FOUND", "Route not found.", nil)
	})

	// Return configured router.
	return r
}
