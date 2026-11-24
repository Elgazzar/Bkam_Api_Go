# Research: Deepen Historical Engine & Demote Leaky Firestore Storage Layer

**Feature Branch**: `110-deepen-history-engine`  
**Date**: 2026-11-15  
**Spec**: [spec.md](./spec.md)

---

## 1. Research Objectives

Investigate the architectural consolidation of historical date range slicing and storage caching into `HistoryResolutionEngine`, evaluate the decommissioning of `HistoryCacheStore`, and define the stateless seam for `firestore-service.ts`.

---

## 2. Architectural Findings & Decisions

### Decision 1: Deepen `HistoryResolutionEngine.getDateRange` to Absorb Prefix Caching and Dynamic Slicing

- **Context**: 
  Currently, `StockQueryService.fetchStockPriceHistoryRange`, `GoldQueryService.fetchGoldPriceHistoryRange`, and `IndicesQueryService.fetchIndexPriceHistoryRange` each contain ~100-150 lines of identical temporal logic:
  - Checking `todayCairo` and `yesterdayCairo`.
  - Checking pure historical range disk cache (`{domain}_range:{symbol}:{start}:{end}`).
  - Splitting ranges extending into today into an immutable prefix (`{domain}_range_prefix:{symbol}:{start}:{yesterday}`) and dynamic live today.
  - Iterating dates and synthesizing yesterday's rate defensively if absent.
  - Appending today's live rate and intraday points from live push stores.
- **Decision**: 
  Absorb 100% of range slicing, flat-cache persistence, and live today decoration into `HistoryResolutionEngine.getDateRange(assetClass, options)`.
  ```typescript
  export interface RangeOptions {
    startDate: string;
    endDate: string;
    symbol?: string;
    currency?: string;
    bank?: string;
    market?: GccMarket;
    liveSnapshot?: any;
  }
  ```
- **Rationale**:
  - *Locality*: All range caching, prefix caching (`flat-cache`), date chunking, and intraday stitching concentrate in exactly ONE module.
  - *Leverage*: One single method serves rates, gold, stocks, indices, and GCC equities.
  - *Deletion*: Deletes >300 lines of duplicate range code across `StockQueryService`, `GoldQueryService`, and `IndicesQueryService`.
- **Alternatives Considered**:
  - *Keep range slicing in query services*: Rejected. Callers repeatedly drift, miss edge cases, and maintain inconsistent cache keys.
  - *Extract a separate `RangeHelper` utility*: Rejected. A stateless helper without access to engine caches leaks storage state back to callers.

---

### Decision 2: Standardize on Single Disk Cache (`flat-cache` via `CacheService`) and Deprecate `HistoryCacheStore`

- **Context**:
  `src/cache/history-cache-store.ts` manages ad-hoc individual JSON files in `var/history-cache/*.json` without TTL expiration, atomic lock handling, or metadata tracking. Meanwhile, `src/cache/cacheService.ts` wraps `flat-cache` in `var/cache/` with standardized 30-day TTL support, automatic serialization, and disk flushing.
- **Decision**:
  Completely eliminate `HistoryCacheStore` and delete `src/cache/history-cache-store.ts`. All sealed historical snapshots and ranges will be stored exclusively in `CacheService` (`flat-cache`).
- **Rationale**:
  - Eliminates conflicting disk cache formats and prevents duplicate writes on every historical request.
  - Reduces disk I/O and centralizes cache pruning under `flat-cache`.
- **Alternatives Considered**:
  - *Keep `HistoryCacheStore` for individual daily files*: Rejected. Maintaining two distinct disk caching engines in the same service adds cognitive load and causes cache incoherence.

---

### Decision 3: Demote `firestore-service.ts` to a Stateless Storage Adapter

- **Context**:
  `firestore-service.ts` has accumulated multiple stateful in-memory structures:
  - `weeklyBucketCache`: In-flight promise map and TTL timeouts.
  - `missingDatesCache`: In-memory negative cache map.
  - `activeWeeklyBucketStore`: Accessing live push store.
  - `historyCacheStore`: Disk cache file reader/writer.
  - Calling `marketScheduleService` to calculate dynamic TTLs.
  At the same time, `HistoryResolutionEngine` maintains its own `missingDatesCache` and checks `activeWeeklyBucketStore` and `cacheService` before delegating to `firestore-service.ts`, resulting in duplicate checks on every single request.
- **Decision**:
  Strip all caches (`missingDatesCache`, `weeklyBucketCache`, `historyCacheStore`) and schedule checks from `firestore-service.ts`.
  Convert it into a pure, stateless database reader:
  - `fetchWeeklyBucketDocument(collectionName: string, weekKey: string): Promise<Record<string, any> | null>`
  - `fetchDailyDocument(collectionName: string, date: string): Promise<any | null>`
  All negative caching, in-memory deduplication, and disk cache lookups are owned exclusively by `HistoryResolutionEngine`.
- **Rationale**:
  - Transforms `firestore-service.ts` into a clean, testable storage adapter.
  - Eliminates double negative caching and double disk checking.
- **Alternatives Considered**:
  - *Keep negative caching in `firestore-service.ts`*: Rejected. Caching belongs at the engine/coordinator layer, not inside the raw cloud storage adapter.

---

### Decision 4: Single Authoritative In-Memory Negative Cache in `HistoryResolutionEngine`

- **Context**:
  Currently, both `HistoryResolutionEngine` and `firestore-service.ts` instantiate `new Map<string, number>()` named `missingDatesCache`.
- **Decision**:
  Consolidate into `HistoryResolutionEngine.missingDatesCache`. When a weekly bucket document exists in Firestore but a requested date is absent from that bucket:
  - The date is identified as a non-trading day (weekend or official holiday).
  - The engine sets a negative cache entry for 24 hours (or 5 minutes if querying active today).
  - The engine returns `null` immediately without attempting legacy daily lookups.
- **Rationale**:
  - Guarantees exactly one negative cache lookup path.
  - Eliminates duplicate map allocations and TTL expiration handling.

---

## 3. Summary of Refactoring Impact

| Component | Before | After | Lines of Code Delta |
| :--- | :--- | :--- | :--- |
| `src/services/history-resolution-engine.ts` | Single-date engine with partial range support; duplicates checks with firestore-service | Deep coordinator owning single-date + multi-day ranges, flat-cache keys, and single negative cache | ~ +150 lines (encapsulating range logic) |
| `src/services/firestore-service.ts` | 381 lines; stateful with 4 cache layers (`weeklyBucketCache`, `missingDatesCache`, `historyCacheStore`, `activeWeeklyBucketStore`) | ~120 lines; stateless Firestore document fetcher | ~ -260 lines |
| `src/cache/history-cache-store.ts` | 40 lines; separate JSON disk file reader/writer | **DELETED** | -40 lines |
| `src/services/stock-query-service.ts` | 393 lines; 120 lines of range slicing and prefix cache keys | Delegating facade invoking `historyResolutionEngine.getDateRange` | ~ -110 lines |
| `src/services/gold-query-service.ts` | Duplicated range prefix and slicing | Delegating facade invoking `historyResolutionEngine.getDateRange` | ~ -90 lines |
| `src/services/indices-query-service.ts` | Duplicated range prefix and slicing | Delegating facade invoking `historyResolutionEngine.getDateRange` | ~ -90 lines |
| **Total Code Reduction** | Complex multi-tier duplication | Deep module with single seam | **Net deletion of >400 lines of boilerplate** |
