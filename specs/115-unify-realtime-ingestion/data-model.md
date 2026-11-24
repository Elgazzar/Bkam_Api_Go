# Data Model: Unify Real-Time Ingestion and Active Weekly Caching

**Feature**: `115-unify-realtime-ingestion`  
**Date**: 2026-11-16  
**Status**: Draft  

---

## 1. Domain Entities & Storage Layout

### 1.1 In-Memory Active Weekly Bucket (`HistoryResolutionEngine` Tier 1)

The active calendar week bucket is held directly in memory inside `HistoryResolutionEngine` as a private map, keyed by the collection and ISO week string:

```typescript
// Canonical key structure: `${collectionName}/${weekKey}`
// Example: "rates_history_weekly/2026-W47"
// Example: "stock_prices_history_weekly/2026-W47"
type ActiveWeeklyBucketKey = `${string}/${string}`;

interface ActiveWeeklyBucketPayload {
  // Map of daily calendar dates (YYYY-MM-DD) to daily closing rate snapshots
  [date: string]: Record<string, any>;
}
```

#### Lifecycle State Transitions:

```
[Server Cold Boot]
       │
       ▼
Check disk cache (`flat-cache`) for active week key
       │
       ├─► [Found in disk] ──► seed in-memory Tier 1 (`setActiveWeeklyBucket`)
       │
       └─► [Not found] ────► Tier 1 initialized as null (awaits Firestore push)
                                      │
                                      ▼
                      [Real-time Firestore Push Received]
                                      │
                                      ▼
                      atomic write to Tier 1 (`setActiveWeeklyBucket`)
                                      │
                                      ▼
                      [Sunday 00:00 Cairo Weekly Rollover]
                                      │
                                      ├─► Read prior active week from Tier 1
                                      ├─► Persist to `flat-cache` (TTL 30 days)
                                      └─► Clear prior week from Tier 1
```

---

## 2. Service Interface Contracts

### 2.1 `IHistoryResolutionEngine` Tier 1 Weekly Store Extension

The interface and implementation of `HistoryResolutionEngine` are deepened to absorb the methods previously scattered across `ActiveWeeklyBucketStore`:

```typescript
export interface IHistoryResolutionEngine {
  // Existing contract methods:
  getDateSnapshot(assetClass: AssetClass, date: string, options?: DateSnapshotOptions): Promise<any | null>;
  getDateRange(assetClass: AssetClass, options: RangeOptions): Promise<any[]>;
  resolveWeeklyCollection(assetClass: AssetClass, market?: GccMarket): string;
  resolveLegacyDailyCollection(assetClass: AssetClass, market?: GccMarket): string;

  // NEW Tier 1 Active Week Memory Store Methods:
  getActiveWeeklyBucket(collectionName: string, weekKey: string): Record<string, any> | null;
  setActiveWeeklyBucket(collectionName: string, weekKey: string, data: Record<string, any>): void;
  deleteActiveWeeklyBucket(collectionName: string, weekKey: string): boolean;
  clearActiveWeeklyBuckets(): void;
  clearActiveWeek(weekKey: string): void;
}
```

### 2.2 `FirestoreListenerService` Refactored Interface

The `FirestoreListenerService` transforms from a 7-dependency multiplexer into a clean Ports & Adapters inbound push listener:

```typescript
export class FirestoreListenerService {
  constructor(
    // Centralized LiveSnapshotCoordinator (defaults to singleton)
    coordinator: LiveSnapshotCoordinator = liveSnapshotCoordinator,
    // Centralized HistoryResolutionEngine (defaults to singleton)
    historyEngine: HistoryResolutionEngine = historyResolutionEngine
  );

  // Starts all Firestore onSnapshot subscriptions
  public start(): void;

  // Stops and cleanly unbinds all real-time listeners
  public stop(): void;

  // Attaches active week listeners for monitored weekly collections
  public startWeeklyListeners(): void;
}
```

### 2.3 `RatesSnapshotService` Constructor Refactoring

`RatesSnapshotService` drops the legacy `SnapshotCache` and `SnapshotStore` shims, injecting `LiveSnapshotCoordinator` directly:

```typescript
export class RatesSnapshotService {
  constructor(
    // Primary centralized coordinator for memory and disk snapshot persistence
    coordinator: LiveSnapshotCoordinator = liveSnapshotCoordinator,
    // Optional storage directory for legacy fallback compatibility
    storageDir?: string
  );

  // Directly delegates in-memory retrieval:
  public getSnapshot(): RateSnapshot | null {
    return this.coordinator.getSnapshot<RateSnapshot>("rates");
  }

  // Directly delegates in-memory and disk persistence:
  public async updateCache(rawDocs: any[]): Promise<void> {
    // Process and enrich rates...
    this.coordinator.updateSnapshot("rates", processedSnapshot);
  }
}
```

---

## 3. Files to Delete (Deletion Test Analysis)

| File | Lines | Rationale for Deletion |
|---|---|---|
| `src/services/active-weekly-bucket-store.ts` | 76 | Shallow wrapper around `Map`. Absorbed by `HistoryResolutionEngine`. |
| `src/cache/snapshot-cache.ts` | 27 | Zombie shim forwarding to `LiveSnapshotCoordinator`. |
| `src/cache/snapshot-store.ts` | 46 | Zombie shim forwarding to `LiveSnapshotCoordinator`. |
| `src/services/market-schedule-service.ts` | 62 | Dead facade over `marketScheduleEngine` with zero consumers. |
| `tests/unit/active-weekly-bucket-store.test.ts` | 47 | Redundant test file testing the deleted shallow store; superseded by `tests/unit/history-resolution-engine.test.ts`. |
