# Research & Technical Decisions: Live Snapshot Coordinator

**Feature Branch**: `109-live-snapshot-coordinator`  
**Date**: 2026-11-15  
**Spec Reference**: [spec.md](./spec.md)

---

## 1. Multi-Domain In-Memory Storage Architecture

### Decision
Store all real-time market snapshots in a centralized in-memory hash map (`Map<string, LiveDomainSnapshot>`) keyed by composite domain identifiers (`${domain}` or `${domain}:${market}`), managed by a singleton `LiveSnapshotCoordinator` class.

### Rationale
- **Sub-0.01ms Access Time**: JavaScript `Map.get()` executes in constant $O(1)$ time in RAM, satisfying Constitution Principle II (Zero-Read Caching) and SC-001 ($<1\text{ ms}$ SLA).
- **Memory Efficiency**: The entire active price catalog across currencies (37 banks), precious metals (4 karats + ounce + coins), gas products (6 fuels), EGX equities (102 stocks), regional GCC equities (620 stocks across 4 countries), and benchmark indices (20 indices) occupies under 8MB of Node.js heap memory, well within the 24GB OCI Ampere Always Free envelope.
- **Atomic Pointer Updates**: In JavaScript's single-threaded event loop, replacing a snapshot reference (`this.snapshots.set(key, newSnapshot)`) is atomic. Readers never observe half-written quotes or locked collections.

### Alternatives Considered
- **Status Quo (Per-Service Ad-Hoc Fields)**: Retaining `cachedPrices: T | null` across 6 separate query service files. *Rejected*: Causes high coupling in listeners/quotes, code duplication across fallback paths, and violates the deletion test.
- **External Redis / Memcached**: Deploying an external cache daemon. *Rejected*: Adds operational complexity, network latency hops (1-2ms vs 0.01ms RAM), and violates the $0 direct infrastructure budget constraint (Principle VI).

---

## 2. Interface Design and Seam Placement

### Decision
Expose an expressive, high-leverage 3-method public interface on `ILiveSnapshotCoordinator`:

```typescript
export interface ILiveSnapshotCoordinator {
  // Synchronous, non-blocking in-memory snapshot retrieval
  getSnapshot<T = any>(domain: AssetDomain, market?: string): T | null;

  // Real-time push update ingestion point for background listeners
  updateSnapshot(domain: AssetDomain, data: any, market?: string): void;

  // Authoritative fallback query to Firestore with promise coalescing and stale fallback
  fetchFresh<T = any>(domain: AssetDomain, market?: string, forceRefresh?: boolean): Promise<T | null>;
}
```

### Rationale
- **High Depth / Small Interface**: Three generic methods absorb hundreds of lines of per-domain cache check, Firestore collection querying, degraded fallback handling, and error logging.
- **Interface Decoupling**: Downstream consumers (`QuotesService`, `IntradayCoordinator`, `HistoryResolutionEngine`, and route handlers) only learn one interface instead of importing 6 query services.
- **Push Listener Simplification**: `FirestoreListenerService` binds to `liveSnapshotCoordinator.updateSnapshot(domain, data, market)` as a single sink, eliminating 6 injected service dependencies.

### Alternatives Considered
- **Per-Domain Methods (`getRates()`, `getGold()`, `getStocks()`, etc.)**: *Rejected*: Creates a wide, shallow interface with 18+ methods that must be updated whenever a new regional market or asset type is introduced.

---

## 3. Cold-Boot Warming and Persistent Disk Storage

### Decision
Consolidate the legacy `SnapshotStore` (which currently only writes bank rates to disk) into a unified `DiskSnapshotAdapter` inside the coordinator. On push update, write snapshots asynchronously to `var/snapshots/{domain}_{market}.json`. On startup, warm memory from disk before Firestore listeners establish active WebSocket connections.

### Rationale
- **Zero-Latency Startup**: Server instances immediately answer `/v1/health` and live traffic upon startup with valid pricing snapshots, eliminating startup 503/500 errors.
- **Absorption of Shallow Modules**: Retires `src/cache/snapshot-cache.ts` (13 lines) and `src/cache/snapshot-store.ts` (25 lines), concentrating disk persistence within the coordinator.

### Alternatives Considered
- **Firestore-Only Cold Boot**: Fetch all 6 collections from Firestore on container launch. *Rejected*: Slower cold starts (~1.5s), consumes initial Firestore read quotas, and fails if the container restarts during a Firestore network glitch.

---

## 4. In-Flight Request Deduplication (Thundering Herd Guard)

### Decision
Implement promise coalescing inside `fetchFresh()` using an in-flight cache map (`inFlightFetches: Map<string, Promise<any>>`).

### Rationale
If 50 concurrent requests arrive during a cold-cache miss or force-refresh:
1. The first request creates the upstream Firestore fetch promise and registers it in `inFlightFetches`.
2. Requests 2 through 50 receive the identical existing promise and await its completion.
3. Once settled, the entry is removed from `inFlightFetches`.
Exactly one database read occurs, strictly protecting Firestore free-tier read quotas (<50,000 reads/day per Constitution Principle II).

---

## 5. Migration and Backward Compatibility

### Decision
Implement the coordinator as a new singleton in `src/services/live-snapshot-coordinator.ts`, then systematically rewire consumers:
1. `FirestoreListenerService`: replace 6 domain query service dependencies with `liveSnapshotCoordinator`.
2. `QuotesService`: read all live quotes via `liveSnapshotCoordinator.getSnapshot()`.
3. `StockQueryService`, `IndicesQueryService`, `GoldQueryService`, `GasQueryService`, `GccStockQueryService`: replace internal `cachedPrices` with delegations to `liveSnapshotCoordinator`.
4. `RatesSnapshotService`: delegate memory caching and warming to `liveSnapshotCoordinator`.

All existing public class methods on existing query services remain intact as backward-compatible delegate facades so that external routes and existing test suites experience zero breaking changes.
