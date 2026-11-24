# Research & Architectural Analysis: Unify Real-Time Ingestion and Active Weekly Caching

**Feature**: `115-unify-realtime-ingestion`  
**Date**: 2026-11-16  
**Status**: Completed  

---

## 1. Technical Unknowns & Architectural Investigation

### Research Question 1: How should `HistoryResolutionEngine` manage the active week in Tier 1 RAM without `ActiveWeeklyBucketStore`?

- **Current State**:
  `ActiveWeeklyBucketStore` (`src/services/active-weekly-bucket-store.ts`) is a 76-line wrapper around a generic JavaScript `Map<string, any>`. It maps composite keys (`${collectionName}/${weekKey}`) to daily document dictionaries.
  It is queried exclusively by `HistoryResolutionEngine` in three places:
  1. Single-date historical lookup: `activeWeeklyBucketStore.get(collectionName, weekKey)`
  2. Range historical lookup: `activeWeeklyBucketStore.get(weeklyCol, weekKey)`
  3. Weekly bucket fetch caching: `activeWeeklyBucketStore.set(weeklyCol, weekKey, weekData)`
  It is populated and rolled over exclusively by `FirestoreListenerService` in `startWeeklyListeners()` and `rolloverWeeklyListeners()`.

- **Finding & Decision**:
  `ActiveWeeklyBucketStore` has no business existing as a standalone service. Tier 1 memory storage for weekly buckets is an internal implementation detail of `HistoryResolutionEngine`.
  `HistoryResolutionEngine` will maintain an internal `activeWeeklyBuckets: Map<string, Record<string, any>>` and expose clean, explicit methods on its interface:
  - `getActiveWeeklyBucket(collectionName: string, weekKey: string): Record<string, any> | null`
  - `setActiveWeeklyBucket(collectionName: string, weekKey: string, data: Record<string, any>): void`
  - `deleteActiveWeeklyBucket(collectionName: string, weekKey: string): boolean`
  - `clearActiveWeeklyBuckets(): void`
  - `archiveActiveWeekToDisk(collectionName: string, weekKey: string): void`
  This concentrates weekly memory management inside the history engine, satisfies the deletion test (complexity concentrates, not moves), and eliminates an artificial cross-service seam.

---

### Research Question 2: How should `FirestoreListenerService` interact with `LiveSnapshotCoordinator` and `HistoryResolutionEngine`?

- **Current State**:
  `FirestoreListenerService` accepts 7 dependencies in its constructor:
  `ratesService`, `goldService`, `gasService`, `stockService`, `indicesService`, `gccStockService`, `coordinator`.
  On every real-time Firestore event, it performs a circular dual-write:
  ```typescript
  this.coordinator.updateSnapshot("gold", data);
  if (this.goldService) {
    this.goldService.updateCache(data); // Which literally executes liveSnapshotCoordinator.updateSnapshot("gold", data) AGAIN!
  }
  ```
  Furthermore, `GasQueryService`, `RatesSnapshotService`, `GoldQueryService`, `StockQueryService`, and `IndicesQueryService` do not perform any custom caching in `updateCache`—they all forward directly to `LiveSnapshotCoordinator`.

- **Finding & Decision**:
  `FirestoreListenerService` is an inbound push adapter in a Ports & Adapters architecture. Its role is strictly to listen to Firestore real-time snapshot streams and dispatch updates to the core domain engines:
  1. **Live snapshots** (rates, gold, gas, egx stocks, gcc stocks, indices) $\rightarrow$ `this.coordinator.updateSnapshot(...)`
  2. **Intraday ticks** (rates, gold, stocks, indices) $\rightarrow$ `intradaySnapshotService.updateCategory(...)`
  3. **Weekly consolidated buckets** (rates, stocks, indices, gold) $\rightarrow$ `this.historyEngine.setActiveWeeklyBucket(...)`
  The constructor is streamlined to accept only `coordinator` and `historyEngine` (with defaults to their respective singletons). The 5 query service injections and the duplicate `updateCache()` calls are eliminated.

---

### Research Question 3: How should `SnapshotCache` and `SnapshotStore` be eliminated from `RatesSnapshotService`?

- **Current State**:
  `SnapshotCache` (`src/cache/snapshot-cache.ts`, 27 lines) and `SnapshotStore` (`src/cache/snapshot-store.ts`, 46 lines) were demoted during Feature 109 when `LiveSnapshotCoordinator` was introduced. Their source code explicitly states:
  `// Legacy SnapshotCache adapter delegating in-memory caching to LiveSnapshotCoordinator (Feature 109)`
  `// Legacy SnapshotStore adapter delegating disk persistence to LiveSnapshotCoordinator (Feature 109)`
  They are used only to construct `RatesSnapshotService` in `src/app.ts`, `src/routes/health-routes.ts`, and `tests/unit/rates-snapshot-service.test.ts`.

- **Finding & Decision**:
  `RatesSnapshotService` can inject `LiveSnapshotCoordinator` directly (or default to the `liveSnapshotCoordinator` singleton).
  - In-memory retrieval: `this.coordinator.getSnapshot<RateSnapshot>("rates")`
  - In-memory & disk update: `this.coordinator.updateSnapshot("rates", snapshot)`
  - Delete `src/cache/snapshot-cache.ts` and `src/cache/snapshot-store.ts`.
  - Update `src/app.ts`, `src/routes/health-routes.ts`, and test harnesses to pass/use `LiveSnapshotCoordinator`.

---

### Research Question 4: How should `market-schedule-service.ts` be eliminated?

- **Current State**:
  `src/services/market-schedule-service.ts` (62 lines) provides `MarketScheduleService` and `marketScheduleService`, which wraps `marketScheduleEngine`.
  `git grep` reveals zero usages of `marketScheduleService` across all source code in `src/`. Only `tests/unit/market-schedule-engine.test.ts` imported it in an unused test import.

- **Finding & Decision**:
  Safely delete `src/services/market-schedule-service.ts`.
  Update `tests/unit/market-schedule-engine.test.ts` to test `marketScheduleEngine` directly.

---

## 2. Summary of Architectural Decisions & Impact

| Item | Old Architecture | New Deep Architecture | Impact & Benefit |
|---|---|---|---|
| **Active Week RAM Store** | `ActiveWeeklyBucketStore` (76-line wrapper around `Map`) | Internal Tier 1 cache in `HistoryResolutionEngine` | Concentrates history caching in engine; deletes shallow file |
| **Push Listener Dispatch** | Dual-write: updates coordinator + calls 5 query services | Pure Ports & Adapters: listener dispatches directly to coordinators | Eliminates duplicate updates; listener constructor simplified |
| **Rates Cache Shims** | `SnapshotCache` & `SnapshotStore` pass-through shims | `RatesSnapshotService` directly uses `LiveSnapshotCoordinator` | Deletes 2 legacy shims (73 lines); removes redundant disk writes |
| **Market Schedule Facade** | `market-schedule-service.ts` (62 lines dead code) | Direct use of authoritative `marketScheduleEngine` | Deletes 62 lines of dead code |

**Total Dead / Shallow Code Eliminated**: 4 files, ~211 lines of code.  
**Public API Regressions**: Zero (0). All 49 verification endpoints and Zod contracts remain identical.
