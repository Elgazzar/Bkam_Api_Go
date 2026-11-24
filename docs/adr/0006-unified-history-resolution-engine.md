# 6. Unified History Resolution Engine

## Context
Historical queries across asset classes (Bank Rates, Gold, Stocks, Indices, and GCC regional markets) evolved independently, leading to significant architectural fragmentation:
1. **Missing RAM Cache for Bank Rates**: `rates_history_weekly` was omitted from `watchedWeeklyCollections` in [`FirestoreListenerService`](../../api/src/services/firestore-listener-service.ts), meaning Bank Rates Active Week snapshots were not maintained in volatile RAM, unlike Stocks, Indices, and Gold.
2. **Duplicated & Inconsistent Retrieval Logic**: Date math, Cairo timezone boundaries, ISO week slicing, and multi-tier lookups were copied across `RatesQueryService`, `StockQueryService`, `IndicesQueryService`, `GoldQueryService`, and `GccStockQueryService`.
3. **Gold Single-Date Workaround**: Gold lacked a first-class single-date history lookup, routing `GET /v1/gold/history/:date` through `fetchGoldPriceHistoryRange(date, date)` and extracting the first element.
4. **Intraday vs. History Confusion**: Historical endpoints were defensively attempting to recover and inject high-frequency intraday tick points into daily history documents, blurring the boundary between live session snapshots and immutable historical archives.
5. **Range Caching Trade-Offs**: Stocks and Indices use pre-computed single-symbol Range Result Caches (`stock_range:${symbol}:${start}:${end}`), but naively applying this to Bank Rates would trigger cache key combinatorial explosion due to arbitrary permutations of query parameters (`bank`, `currency`).

## Decision
We established a centralized, modular **History Resolution Engine** with strict domain segregation:

1. **Strict Segregation of History from Intraday Snapshots**:
   - The **Intraday Timeline** is strictly a volatile live session snapshot, maintained in memory via real-time Firestore `onSnapshot` listeners and discarded on daily session rollover.
   - The **History Resolution Engine** strictly retrieves daily closing summaries (open, high, low, close, extremes) from ISO **Weekly Buckets** (`*_history_weekly`). It never recovers or injects intraday tick points.

2. **Active Week RAM Caching Parity for Bank Rates**:
   - Register `rates_history_weekly` in `watchedWeeklyCollections` in `FirestoreListenerService`.
   - The Active Week for Bank Rates is kept synchronized in volatile RAM (`ActiveWeeklyBucketStore`), ensuring `<0.1ms` response times with 0 Firestore reads.

3. **Universal Engine Contract**:
   The engine provides two standard entry points across all asset classes and regional markets:
   - `getDateSnapshot(assetClass: AssetClass, date: string, market?: string): Promise<any | null>`:
     Resolves a single date's closing snapshot:
     1. Check Active Week in RAM (`ActiveWeeklyBucketStore`) if date falls in current ISO week.
     2. Check negative cache for non-trading days/official holidays.
     3. Check Sealed Week in disk cache (`flat-cache`).
     4. Fallback to Firestore weekly bucket (persisting to `flat-cache` for 30 days).
     5. Fallback to legacy daily collection if pre-migration weekly bucket does not exist.
   - `getDateRange(assetClass: AssetClass, options: RangeOptions): Promise<any[]>`:
     Resolves a multi-day range by slicing ISO calendar weeks across the interval, resolving each week from RAM or `flat-cache`, and sorting the resulting daily entries.

4. **Gold Endpoint Parity**:
   - Refactor `GET /v1/gold/history/:date` and yesterday's delta lookup in `GET /v1/gold` to call `getDateSnapshot('gold', date)` directly, eliminating the range workaround.

5. **Direct Weekly Bucket Slicing for Bank Rates Ranges**:
   - Bank Rates history range queries resolve by slicing the ~26 Weekly Buckets in RAM and `flat-cache`, filtering in memory. We deliberately avoid secondary pre-computed `rates_range:*` cache keys to prevent cache key explosion across arbitrary `bank` and `currency` filter combinations.

6. **Regional Market Extensibility**:
   - The engine accepts a parameterized `market` identifier, enabling GCC regional equities (Saudi, UAE, Qatar) to reuse the exact same tiered caching pipeline without code duplication.

## Consequences
- Single, authoritative pipeline for all historical data retrieval across all asset classes.
- Bank Rates Active Week queries resolve in `<0.1ms` from RAM with zero database read quota consumption.
- Clean separation between volatile intraday session data and immutable historical daily archives.
- Zero risk of cache key explosion in `flat-cache` from arbitrary bank and currency query combinations.
- GCC stocks, Gold, EGX equities, Indices, and Bank Rates share identical caching semantics, error handling, and negative caching rules.
