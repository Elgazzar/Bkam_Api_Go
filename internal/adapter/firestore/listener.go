// Package firestore implements the inbound real-time push streaming adapter connecting Google Cloud Firestore snapshot streams to domain engines.
package firestore

// Import standard library and Google Cloud Firestore packages.
import (
	// context handles lifecycle cancellation and timeout control.
	"context"
	// fmt formats error and trace strings.
	"fmt"
	// log provides standard logging output for listener lifecycle events.
	"log"
	// sync coordinates listener goroutine termination during graceful shutdown.
	"sync"
	// time provides backoff sleep durations and timestamping.
	"time"

	// firestore provides the official Google Cloud Firestore SDK client.
	"cloud.google.com/go/firestore"
	// google.golang.org/api/iterator defines iteration constants.
	"google.golang.org/api/iterator"

	// domain defines core financial entities.
	"bkam-api/internal/domain"
	// engine provides the LiveSnapshotCoordinator receiving pushed updates.
	"bkam-api/internal/engine"
)

// FirestoreListenerService manages long-lived goroutine workers listening to Firestore snapshot streams.
type FirestoreListenerService struct {
	// client points to the active Google Cloud Firestore client instance.
	client *firestore.Client
	// coordinator receives and distributes real-time market updates.
	coordinator *engine.LiveSnapshotCoordinator
	// historyEngine receives and caches active weekly historical buckets.
	historyEngine *engine.HistoryResolutionEngine
	// intradayStore receives and caches intraday timelines for today's trading session.
	intradayStore *engine.IntradayTimelineStore
	// processor converts raw bank collection documents into enriched RateSnapshot envelopes.
	processor *RatesProcessor
	// cancel cancels the root context to terminate all background worker goroutines on shutdown.
	cancel context.CancelFunc
	// wg tracks active listener goroutines to guarantee clean shutdown.
	wg sync.WaitGroup
}

// NewFirestoreListenerService constructs a new FirestoreListenerService instance.
func NewFirestoreListenerService(
	client *firestore.Client,
	coordinator *engine.LiveSnapshotCoordinator,
	historyEngine *engine.HistoryResolutionEngine,
	intradayStore *engine.IntradayTimelineStore,
) *FirestoreListenerService {
	// Initialize struct with dependencies.
	return &FirestoreListenerService{
		// Store client pointer.
		client: client,
		// Store coordinator pointer.
		coordinator: coordinator,
		// Store history resolution engine pointer.
		historyEngine: historyEngine,
		// Store intraday timeline store pointer.
		intradayStore: intradayStore,
		// Construct rates processor.
		processor: NewRatesProcessor(),
	}
}

// Start launches all background listener worker goroutines under the provided root context.
func (s *FirestoreListenerService) Start(parentCtx context.Context) {
	// If client is nil (e.g. running in offline test mode), skip starting listeners.
	if s.client == nil {
		// Log informative notice.
		log.Println("[FirestoreListenerService] Firestore client is nil; running in offline/cached mode.")
		// Return early.
		return
	}

	// Create cancellable child context for worker lifecycles.
	ctx, cancel := context.WithCancel(parentCtx)
	// Store cancel function.
	s.cancel = cancel

	// 1. Launch worker listening to "rates" collection.
	s.wg.Add(1)
	// Start rates worker goroutine.
	go s.listenRatesCollection(ctx)

	// 2. Launch worker listening to "gold_prices/latest" document.
	s.wg.Add(1)
	// Start gold worker goroutine.
	go s.listenSingleDocument(ctx, "gold_prices", "latest", func(data map[string]any) {
		// Construct gold snapshot.
		var snap domain.GoldSnapshot
		// Deserialize data map into snap struct.
		if decodeMap(data, &snap) == nil {
			// Update coordinator memory and disk.
			s.coordinator.UpdateSnapshot("gold", &snap)
		}
	})

	// 3. Launch worker listening to "gas_prices/latest" document.
	s.wg.Add(1)
	// Start gas worker goroutine.
	go s.listenSingleDocument(ctx, "gas_prices", "latest", func(data map[string]any) {
		// Construct gas snapshot.
		var snap domain.GasSnapshot
		// Deserialize map.
		if decodeMap(data, &snap) == nil {
			// Update coordinator.
			s.coordinator.UpdateSnapshot("gas", &snap)
		}
	})

	// 4. Launch worker listening to "stock_prices/latest" (EGX stocks).
	s.wg.Add(1)
	// Start EGX stocks worker goroutine.
	go s.listenSingleDocument(ctx, "stock_prices", "latest", func(data map[string]any) {
		// Construct stock snapshot.
		var snap domain.StockSnapshot
		// Set market.
		snap.Market = "egx"
		// Deserialize.
		if decodeMap(data, &snap) == nil {
			// Update coordinator.
			s.coordinator.UpdateSnapshot("stocks", &snap, "egx")
		}
	})

	// 5. Launch worker listening to "index_prices/latest" (EGX benchmark indices).
	s.wg.Add(1)
	// Start EGX indices worker goroutine.
	go s.listenSingleDocument(ctx, "index_prices", "latest", func(data map[string]any) {
		// Construct index snapshot.
		var snap domain.IndexSnapshot
		// Set market.
		snap.Market = "egx"
		// Deserialize.
		if decodeMap(data, &snap) == nil {
			// Update coordinator.
			s.coordinator.UpdateSnapshot("indices", &snap, "egx")
		}
	})

	// 6. Launch workers for regional GCC stock markets.
	gccMarkets := []string{"saudi", "uae", "qatar", "kuwait"}
	// Iterate through GCC markets.
	for _, mkt := range gccMarkets {
		// Capture market variable for closure.
		targetMkt := mkt
		// Collection name e.g. "stock_prices_saudi".
		colName := fmt.Sprintf("stock_prices_%s", targetMkt)
		// Increment WaitGroup.
		s.wg.Add(1)
		// Launch worker.
		go s.listenSingleDocument(ctx, colName, "latest", func(data map[string]any) {
			// Construct stock snapshot.
			var snap domain.StockSnapshot
			// Set market.
			snap.Market = targetMkt
			// Deserialize.
			if decodeMap(data, &snap) == nil {
				// Update coordinator.
				s.coordinator.UpdateSnapshot("gcc-stocks", &snap, targetMkt)
			}
		})
	}

	// 7. Launch workers for regional GCC benchmark indices.
	for _, mkt := range gccMarkets {
		// Capture market variable.
		targetMkt := mkt
		// Collection name e.g. "index_prices_saudi".
		colName := fmt.Sprintf("index_prices_%s", targetMkt)
		// Increment WaitGroup.
		s.wg.Add(1)
		// Launch worker.
		go s.listenSingleDocument(ctx, colName, "latest", func(data map[string]any) {
			// Construct index snapshot.
			var snap domain.IndexSnapshot
			// Set market.
			snap.Market = targetMkt
			// Deserialize.
			if decodeMap(data, &snap) == nil {
				// Update coordinator.
				s.coordinator.UpdateSnapshot("indices", &snap, targetMkt)
			}
		})
	}

	// 8. Launch listeners for active ISO calendar week historical buckets if historyEngine configured.
	if s.historyEngine != nil {
		// Calculate current ISO week string e.g. "2026-W47".
		currentWeekKey := engine.GetISOWeekString(time.Now().UTC())
		// List of core historical weekly collections.
		historyCollections := []string{
			"rates_history_weekly",
			"gold_prices_history_weekly",
			"stock_prices_history_weekly",
			"index_prices_history_weekly",
			"stock_prices_saudi_history_weekly",
			"stock_prices_uae_history_weekly",
			"stock_prices_qatar_history_weekly",
			"stock_prices_kuwait_history_weekly",
		}

		// Subscribe listener worker for each active weekly collection.
		for _, col := range historyCollections {
			// Capture collection string.
			targetCol := col
			// Increment WaitGroup.
			s.wg.Add(1)
			// Launch single document snapshot worker for the active week.
			go s.listenSingleDocument(ctx, targetCol, currentWeekKey, func(data map[string]any) {
				// Store active weekly bucket in Tier 1 RAM and Tier 2 disk.
				s.historyEngine.SetActiveWeeklyBucket(targetCol, currentWeekKey, data)
			})
		}
	}

	// 9. Launch listeners for active today's intraday documents if intradayStore configured.
	if s.intradayStore != nil {
		// Cairo location timezone.
		cairoLoc := time.FixedZone("Africa/Cairo", 2*3600)
		// Current time in Cairo.
		nowCairo := time.Now().In(cairoLoc)
		// Today's Cairo calendar date in YYYY-MM-DD.
		todayDate := nowCairo.Format("2006-01-02")
		// List of intraday subscriptions with target domain and market.
		intradaySubscriptions := []struct {
			col      string
			category string
			market   string
		}{
			{"rates_intraday", "rates", "default"},
			{"gold_prices_intraday", "gold", "default"},
			{"stock_prices_intraday", "stocks", "egx"},
			{"index_prices_intraday", "indices", "egx"},
			{"stock_prices_saudi_intraday", "stocks", "saudi"},
			{"stock_prices_uae_intraday", "stocks", "uae"},
			{"stock_prices_qatar_intraday", "stocks", "qatar"},
			{"stock_prices_kuwait_intraday", "stocks", "kuwait"},
		}

		// Subscribe listener worker for each active intraday collection.
		for _, sub := range intradaySubscriptions {
			// Capture item variables.
			targetCol := sub.col
			cat := sub.category
			mkt := sub.market
			// Increment WaitGroup.
			s.wg.Add(1)
			// Launch snapshot worker for today's intraday document.
			go s.listenSingleDocument(ctx, targetCol, todayDate, func(data map[string]any) {
				// Update in-memory intraday store.
				s.intradayStore.UpdateIntraday(cat, mkt, todayDate, data)
			})
		}
	}
}

// Stop cancels the context and waits for all listener worker goroutines to terminate cleanly.
func (s *FirestoreListenerService) Stop() {
	// If cancel function exists, invoke it.
	if s.cancel != nil {
		// Cancel root listener context.
		s.cancel()
	}
	// Wait for all goroutines to return.
	s.wg.Wait()
}

// listenRatesCollection manages the continuous snapshot stream for the "rates" bank collection.
func (s *FirestoreListenerService) listenRatesCollection(ctx context.Context) {
	// Defer decrement of WaitGroup when worker exits.
	defer s.wg.Done()

	// Loop with exponential backoff on unexpected stream disconnects.
	for {
		// Check if context has been cancelled.
		select {
		case <-ctx.Done():
			// Exit worker goroutine.
			return
		default:
			// Proceed with subscription.
		}

		// Subscribe to snapshot iterator for "rates" collection.
		iter := s.client.Collection("rates").Snapshots(ctx)

		// Continuously receive snapshot events.
		for {
			// Pull next query snapshot from iterator.
			qSnap, err := iter.Next()
			// Handle iteration errors or context cancellation.
			if err != nil {
				// If cancelled by context, exit cleanly.
				if err == iterator.Done || ctx.Err() != nil {
					// Stop iteration.
					iter.Stop()
					// Return from worker.
					return
				}
				// Log stream warning.
				log.Printf("[FirestoreListener] Rates collection stream error: %v; reconnecting...", err)
				// Stop iterator.
				iter.Stop()
				// Break out of inner loop to trigger backoff reconnect.
				break
			}

			// Slice to hold raw bank document representations.
			rawDocs := make([]RawBankDocument, 0, qSnap.Size)
			// Iterate over individual document snapshots using the DocumentIterator.
			docIter := qSnap.Documents
			// Loop through documents until iterator is exhausted.
			for {
				// Pull next document snapshot.
				docSnap, err := docIter.Next()
				// Check for completion or error.
				if err == iterator.Done {
					// All documents processed.
					break
				}
				// Break on unexpected document error.
				if err != nil {
					// Break loop.
					break
				}

				// Initialize raw doc.
				var rawDoc RawBankDocument
				// Extract bankId from document ID if not in payload.
				rawDoc.BankID = docSnap.Ref.ID
				// Populate fields from document map.
				if err := docSnap.DataTo(&rawDoc); err == nil {
					// Fallback bank ID.
					if rawDoc.BankID == "" {
						// Assign doc ID.
						rawDoc.BankID = docSnap.Ref.ID
					}
					// Append to rawDocs list.
					rawDocs = append(rawDocs, rawDoc)
				}
			}

			// Generate unique trace ID for this push ingestion.
			traceID := fmt.Sprintf("push-%d", time.Now().UnixNano())
			// Process and enrich rates against catalog.
			snapshot := s.processor.ProcessRawDocuments(rawDocs, traceID)
			// Push enriched snapshot to coordinator memory and disk.
			s.coordinator.UpdateSnapshot("rates", snapshot)
		}

		// Sleep with backoff before reconnecting to prevent log flooding.
		select {
		case <-ctx.Done():
			// Terminate.
			return
		case <-time.After(15 * time.Second):
			// Proceed to reconnect.
		}
	}
}

// listenSingleDocument manages a continuous snapshot stream for a single document (collection/docID).
func (s *FirestoreListenerService) listenSingleDocument(
	ctx context.Context,
	collection string,
	docID string,
	onUpdate func(data map[string]any),
) {
	// Decrement WaitGroup on worker return.
	defer s.wg.Done()

	// Outer reconnect loop.
	for {
		// Check cancellation.
		select {
		case <-ctx.Done():
			// Return.
			return
		default:
			// Continue.
		}

		// Subscribe to document snapshot iterator.
		iter := s.client.Collection(collection).Doc(docID).Snapshots(ctx)

		// Inner event loop.
		for {
			// Read next snapshot.
			docSnap, err := iter.Next()
			// Check for errors.
			if err != nil {
				// Clean exit on context cancel.
				if err == iterator.Done || ctx.Err() != nil {
					// Stop iterator.
					iter.Stop()
					// Return.
					return
				}
				// Log error.
				log.Printf("[FirestoreListener] Doc %s/%s stream error: %v; reconnecting...", collection, docID, err)
				// Stop iterator.
				iter.Stop()
				// Reconnect.
				break
			}

			// Check if document exists.
			if docSnap.Exists() {
				// Extract data map.
				data := docSnap.Data()
				// Dispatch update to callback function.
				if data != nil && onUpdate != nil {
					// Invoke callback.
					onUpdate(data)
				}
			}
		}

		// Sleep with backoff before reconnecting to prevent log flooding.
		select {
		case <-ctx.Done():
			// Return.
			return
		case <-time.After(15 * time.Second):
			// Reconnect.
		}
	}
}

// decodeMap helper deserializes a generic map[string]any into a typed Go struct pointer.
func decodeMap(input map[string]any, output any) error {
	// Import encoding/json locally or use json round-trip for robust field mapping.
	// Convert input map to JSON bytes.
	bytes, err := domainToJSON(input)
	// Check error.
	if err != nil {
		// Return error.
		return err
	}
	// Unmarshal JSON bytes into destination struct.
	return domainFromJSON(bytes, output)
}
