# Research & Technical Decisions: Dissolve Shadow Caches and Firestore Fallbacks

**Feature Branch**: `112-dissolve-shadow-caches`  
**Date**: 2026-11-15  
**Status**: Completed  

---

## 1. Architectural Problem Analysis

### 1.1 Duplicated State & Shadow Caching
Currently, several domain query services maintain their own private in-memory cache fields:
- `StockQueryService`: `private cachedPrices: StockPriceSnapshot | null = null;`
- `GoldQueryService`: `private cachedPrices: GoldPriceDocument | null = null;`
- `GasQueryService`: `private cachedPrices: GasPriceSnapshot | null = null;`
- `IndicesQueryService`: `private cachedPrices: IndexPriceSnapshot | null = null;` and `private cachedRegionalPrices: Record<string, IndexPriceSnapshot | null> = {};`
- `GccStockQueryService`: `private cachedPrices: Record<GccCountry, GccStockPriceSnapshot | null>;`

At the same time, Feature 109 introduced `LiveSnapshotCoordinator` as the centralized in-memory and disk-persisted snapshot authority. When real-time listener push events arrive, services update both their local `cachedPrices` and `liveSnapshotCoordinator.updateSnapshot()`. Furthermore, accessor methods like `getCachedPrices()` and `fetchLatest...()` perform defensive checks like:
```ts
const inMemorySnapshot = this.cachedPrices || liveSnapshotCoordinator.getSnapshot<...>(...);
```
This violates the Single Source of Truth principle (ADR 0005, Constitution Principle II). Memory is allocated twice for the same snapshot objects, garbage collection pressure is doubled, and updates can become desynchronized if a listener pushes directly to one without the other.

### 1.2 Redundant Direct Database Fallbacks & Thundering Herd
When a cache miss occurs (e.g. during cold boot or explicit force refresh), each domain query module independently executes direct Firestore queries:
```ts
const docRef = db.collection("gold_prices").doc("latest");
const docSnapshot = await docRef.get();
```
Because these fallbacks are implemented independently in each query service:
1. They bypass `LiveSnapshotCoordinator`'s built-in **single-flight promise pooling** (`inFlightFetches.get(key)`).
2. Under concurrent requests during cold boots, 10 or 20 simultaneous requests to `/v1/gold` or `/v1/stocks` trigger 10 or 20 simultaneous Firestore reads for `latest`, rapidly exhausting free-tier quotas.
3. They duplicate disk persistence fallback logic (`diskAdapter.load()`), making domain query services tightly coupled to Firestore `db` instances instead of being lightweight query presenters.

### 1.3 Bloated Dependency Injection in `QuotesService`
`QuotesService` currently takes 6 constructor arguments:
```ts
constructor(
  private readonly ratesService?: RatesSnapshotService,
  private readonly goldService?: GoldQueryService,
  private readonly indicesService?: IndicesQueryService,
  private readonly stockService?: StockQueryService,
  private readonly gccStockService?: any,
  private readonly coordinator: LiveSnapshotCoordinator = liveSnapshotCoordinator
)
```
Inside `QuotesService`, it queries `this.coordinator.getSnapshot(...)` first, but falls back to each injected domain service if null. Because `LiveSnapshotCoordinator` already holds all domain snapshots (`rates`, `gold`, `gas`, `stocks`, `indices`, `gcc-stocks`), `QuotesService` does not need any of the 5 domain services!

---

## 2. Technical Decisions & Consolidation Matrix

### Decision 1: Slim Down `QuotesService` to Single Authority Dependency
- **Decision**: Update `QuotesService` constructor to accept exclusively `(coordinator: LiveSnapshotCoordinator = liveSnapshotCoordinator)`. For backward compatibility with legacy test suites, keep optional stub/mock parameters or allow them to pass stubs to `coordinator`.
- **Rationale**: `LiveSnapshotCoordinator` is already the unified in-memory snapshot store across all financial domains. Eliminating the 5 domain dependencies decouples `QuotesService`, simplifies service instantiation in `src/app.ts`, and prevents redundant object graph construction.
- **Alternatives Considered**: Keeping optional service parameters alongside coordinator was rejected because it perpetuates multiple sources of truth and confusing fallback precedence.

### Decision 2: Dissolve Shadow Caches in `GoldQueryService` and `GasQueryService`
- **Decision**: Remove `private cachedPrices` from `GoldQueryService` and `GasQueryService`.
  - `getCachedPrices()` becomes: `return liveSnapshotCoordinator.getSnapshot<...>(domain);`
  - `updateCache(data)` becomes: `liveSnapshotCoordinator.updateSnapshot(domain, data);`
  - `fetchLatest...Prices(forceRefresh = false)` becomes: `return liveSnapshotCoordinator.fetchFresh<...>(domain, undefined, forceRefresh);`
- **Rationale**: `LiveSnapshotCoordinator.fetchFresh(domain, market, forceRefresh)` already contains:
  1. Cache-first check (`getSnapshot(domain)`).
  2. Single-flight promise pooling (`inFlightFetches.get(key)`).
  3. Upstream Firestore retrieval (`db.collection(...).doc("latest").get()`).
  4. Stale in-memory degraded fallback.
  5. Stale disk-persistence recovery.
  Delegating to `fetchFresh` eliminates ~80 lines of duplicate boilerplate per service and enforces single-flight protection.

### Decision 3: Dissolve Shadow Caches in `StockQueryService` and `IndicesQueryService`
- **Decision**: Remove `cachedPrices` from `StockQueryService`, and `cachedPrices` / `cachedRegionalPrices` from `IndicesQueryService`.
  - `StockQueryService.fetchLatestStockPrices(forceRefresh)` delegates directly to `liveSnapshotCoordinator.fetchFresh<StockPriceSnapshot>("stocks", "egx", forceRefresh)`.
  - `IndicesQueryService.fetchLatestIndexPrices(marketOrForceRefresh, forceRefresh)` normalizes arguments and delegates directly to `liveSnapshotCoordinator.fetchFresh<IndexPriceSnapshot>("indices", market, effectiveForceRefresh)`.
  - In `IndicesQueryService.fetchConstituentStocks()`, constituent live stocks are resolved directly from `liveSnapshotCoordinator.getSnapshot("stocks", "egx") || await liveSnapshotCoordinator.fetchFresh("stocks", "egx", forceRefresh)`.
- **Rationale**: Centralizes all stock and index quote retrieval, ensuring consistent catalog metadata enrichment through `marketCatalog.enrichSnapshot()` and eliminating uncoordinated Firestore queries.

### Decision 4: Backward Compatibility with Unit Tests
- **Decision**: In unit tests where existing tests may mock or instantiate query services without a live Firestore connection, `liveSnapshotCoordinator.updateSnapshot(domain, mockData, market)` will be used in `beforeEach` to seed test fixtures into volatile RAM, or query services will preserve clean stub compatibility.
- **Rationale**: Ensures all 36 unit test files (310+ tests) continue to pass with 0 test regressions.

---

## 3. Comparison Matrix

| Component | Before (Feature 111) | After (Feature 112) |
| :--- | :--- | :--- |
| `QuotesService` | 6 constructor parameters (`ratesService`, `goldService`, `indicesService`, `stockService`, `gccStockService`, `coordinator`) | 1 constructor parameter (`coordinator: LiveSnapshotCoordinator = liveSnapshotCoordinator`) |
| `StockQueryService` | `private cachedPrices` + direct `db.collection("stock_prices").doc("latest").get()` | 0 shadow caches; delegates 100% to `liveSnapshotCoordinator` |
| `GoldQueryService` | `private cachedPrices` + direct `db.collection("gold_prices").doc("latest").get()` | 0 shadow caches; delegates 100% to `liveSnapshotCoordinator` |
| `GasQueryService` | `private cachedPrices` + 95 lines of boilerplate direct Firestore reads | 0 shadow caches; 1-line delegations to `liveSnapshotCoordinator` |
| `IndicesQueryService` | `private cachedPrices` + `cachedRegionalPrices` + direct Firestore reads | 0 shadow caches; delegates 100% to `liveSnapshotCoordinator` |
| Single-Flight Protection | Only applied when calling `LiveSnapshotCoordinator.fetchFresh` directly | Applied universally across all domain query service calls |
