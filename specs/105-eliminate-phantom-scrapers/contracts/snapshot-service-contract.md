# Interface Contract: RatesSnapshotService

**Feature**: `105-eliminate-phantom-scrapers`

**Date**: 2026-11-15

---

## 1. Class Interface Definition

```typescript
export class RatesSnapshotService {
  /**
   * Initializes RatesSnapshotService with only true runtime dependencies.
   * Eliminates phantom adapters and scraping engines.
   *
   * @param cache In-memory volatile L1 cache layer
   * @param store Persistent disk L2 fallback layer
   * @param freshnessService Evaluation service for stale fallback detection
   */
  constructor(
    private readonly cache: SnapshotCache,
    private readonly store: SnapshotStore,
    private readonly freshnessService = new FreshnessService()
  );

  /**
   * Synchronous accessor to retrieve the currently active snapshot from volatile memory.
   * Applies optional bank and currency query filters in-memory with sub-1ms response latency.
   *
   * @param filters Optional filtering parameters (bank, currency)
   * @returns Filtered RateSnapshot
   */
  getSnapshot(filters?: RateQuery): RateSnapshot;

  /**
   * On-demand snapshot refresh executing a direct query against Google Firestore.
   * Used for initial cache warming and fallback cycles.
   *
   * @param traceId Correlation trace ID for diagnostic logging
   * @returns Enriched RateSnapshot
   */
  refresh(traceId: string): Promise<RateSnapshot>;

  /**
   * Push-based cache update invoked by FirestoreListenerService upon receiving
   * real-time Firestore onSnapshot collection events.
   *
   * @param firestoreRates Raw Firestore document array
   */
  updateCache(firestoreRates: any[]): Promise<void>;

  /**
   * Normalizes raw Firestore document records into the unified RateSnapshot model.
   *
   * @param firestoreRates Raw document records
   * @param traceId Correlation trace ID
   * @returns Fully normalized and enriched RateSnapshot
   */
  processFirestoreRates(firestoreRates: any[], traceId: string): Promise<RateSnapshot>;
}
```

---

## 2. Fastify Application Bootstrap Contract (`src/app.ts`)

```typescript
// Refactored bootstrap: Zero scraper imports or adapter instantiations
export async function buildApp(customMetricsService?: MetricsService): Promise<AppContext> {
  const app = Fastify({ loggerInstance: logger });

  // 1. Storage & Cache Layers
  const snapshotCache = new SnapshotCache();
  const snapshotStore = new SnapshotStore(env.SNAPSHOT_DIR);

  // 2. Snapshot Service: Initialized directly without adapters
  const snapshotService = new RatesSnapshotService(
    snapshotCache,
    snapshotStore
  );

  // 3. Query Services
  const catalogService = new CatalogService();
  const ratesQueryService = new RatesQueryService(catalogService, snapshotService);
  const goldQueryService = new GoldQueryService();
  const gasQueryService = new GasQueryService();
  const stockQueryService = new StockQueryService();
  const indicesQueryService = new IndicesQueryService();

  // 4. Background & Listener Daemons
  const refreshJob = new RefreshRatesJob(snapshotService, goldQueryService, gasQueryService);
  const firestoreListenerService = new FirestoreListenerService(
    snapshotService,
    goldQueryService,
    gasQueryService,
    stockQueryService,
    indicesQueryService
  );

  // ... route registrations ...
  return {
    app,
    refreshJob,
    snapshotService,
    goldQueryService,
    gasQueryService,
    stockQueryService,
    indicesQueryService,
    firestoreListenerService,
    metricsService: customMetricsService || defaultMetricsService,
    quotesService: new QuotesService()
  };
}
```
