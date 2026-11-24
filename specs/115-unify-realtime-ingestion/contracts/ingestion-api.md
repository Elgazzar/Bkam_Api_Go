# Inbound & Internal Contracts: Unify Real-Time Ingestion and Active Weekly Caching

**Feature**: `115-unify-realtime-ingestion`  
**Date**: 2026-11-16  
**Status**: Draft  

---

## 1. Public API Contract Compatibility Guarantee

This architectural feature is an internal structural deepening and does **NOT** alter any external HTTP endpoints, query parameters, or response payloads.

- **Zero Breaking Changes**: All 49 existing endpoints verified by `npm run verify` retain identical JSON structures, status codes, and caching headers.
- **Header Parity**: `Cache-Control`, `ETag`, `X-Cache`, and `x-trace-id` headers remain identical.
- **Latency Guarantee**: All in-memory lookups remain sub-millisecond; multi-route warmed latency remains strictly below the 150ms SLA ceiling.

---

## 2. Inbound Push Adapter Contract (`FirestoreListenerService`)

The inbound push adapter translates Firestore real-time collection/document streams into domain engine method calls.

```
Firestore Push Event
        │
        ├─► [Collection: rates] ───────────────► RatesSnapshotService.updateCache(rawDocs)
        │                                             │
        │                                             ▼
        │                                        LiveSnapshotCoordinator.updateSnapshot("rates", data)
        │
        ├─► [Document: gold_prices/latest] ────► LiveSnapshotCoordinator.updateSnapshot("gold", data)
        │
        ├─► [Document: gas_prices/latest] ─────► LiveSnapshotCoordinator.updateSnapshot("gas", data)
        │
        ├─► [Document: stock_prices/latest] ───► LiveSnapshotCoordinator.updateSnapshot("stocks", data, "egx")
        │
        ├─► [Document: index_prices/latest] ───► LiveSnapshotCoordinator.updateSnapshot("indices", data, "egx")
        │
        ├─► [Document: stock_prices_*/latest] ─► LiveSnapshotCoordinator.updateSnapshot("gcc-stocks", data, market)
        │
        ├─► [Collection: *_intraday/:today] ───► intradaySnapshotService.updateCategory(cat, points, today)
        │
        └─► [Collection: *_history_weekly/:week] ► HistoryResolutionEngine.setActiveWeeklyBucket(col, week, data)
```

### Event Handling Rules:

1. **Existence Verification**: Before dispatching to a domain coordinator, the listener verifies `docSnapshot.exists` and validates that expected root properties are present.
2. **Atomic Ingestion**: Ingestion into `HistoryResolutionEngine` or `LiveSnapshotCoordinator` is atomic and non-blocking.
3. **Rollover Execution**: At 8:00 AM Cairo time (for intraday listeners) and midnight Sunday Cairo time (for weekly bucket listeners), listeners automatically detach old listeners, archive sealed data to disk cache, and re-subscribe to current business dates.

---

## 3. Internal Engine Interface Contract (`HistoryResolutionEngine`)

The historical engine exposes the following Tier 1 memory contract:

```typescript
export interface HistoryResolutionEngineWeeklyTier {
  /**
   * Retrieves an active weekly consolidated bucket from Tier 1 memory.
   * @param collectionName Canonical weekly collection name (e.g. 'rates_history_weekly')
   * @param weekKey ISO calendar week key (e.g. '2026-W47')
   * @returns Daily closing dictionary or null if not in memory
   */
  getActiveWeeklyBucket(collectionName: string, weekKey: string): Record<string, any> | null;

  /**
   * Atomically sets an active weekly consolidated bucket in Tier 1 memory.
   * @param collectionName Canonical weekly collection name
   * @param weekKey ISO calendar week key
   * @param data Daily closing dictionary
   */
  setActiveWeeklyBucket(collectionName: string, weekKey: string, data: Record<string, any>): void;

  /**
   * Deletes an active weekly consolidated bucket from Tier 1 memory.
   * @param collectionName Canonical weekly collection name
   * @param weekKey ISO calendar week key
   * @returns True if deleted, false if not found
   */
  deleteActiveWeeklyBucket(collectionName: string, weekKey: string): boolean;

  /**
   * Clears all active weekly buckets for a specific calendar week.
   * Used during weekly rollover.
   * @param weekKey ISO calendar week key
   */
  clearActiveWeek(weekKey: string): void;

  /**
   * Clears all Tier 1 in-memory weekly buckets.
   * Used for test isolation and teardown.
   */
  clearActiveWeeklyBuckets(): void;
}
```
