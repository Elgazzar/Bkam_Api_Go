# Phase 0 Research: Unified History Resolution Engine

## Executive Summary
This document investigates and consolidates architectural decisions for implementing the **History Resolution Engine** in compliance with ADR 0006, Principle XIII (Intelligent Caching, TTL Management, and Intraday Lifecycle), Principle XV (Modular Architecture & Component Registry), and the explicit user requirement that the engine must utilize the project's centralized `CacheService` layer.

---

## 1. Engine Placement, Interface & Modularity

### Context
Currently, historical retrieval logic (ISO week calculation, Cairo date math, negative caching, 3-tier fallback, and yesterday recovery) is duplicated across:
- `api/src/services/rates-query-service.ts`
- `api/src/services/gold-query-service.ts`
- `api/src/services/stock-query-service.ts`
- `api/src/services/indices-query-service.ts`
- `api/src/modules/gcc-stocks/services/gcc-stock-query-service.ts`

### Decision
Extract a dedicated service: [`api/src/services/history-resolution-engine.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/history-resolution-engine.ts) implementing interface `IHistoryResolutionEngine`. The class will be instantiated as a singleton `historyResolutionEngine` and injected into the existing query services.

### Rationale
- **Loose Coupling (Principle XV)**: Encapsulates all tiered storage mechanics (RAM Active Week, `CacheService` disk cache, Firestore weekly documents, negative caching).
- **Separation of Concerns**: Query services focus purely on catalog validation and client DTO formatting, delegating all historical data resolution to the engine.

### Alternatives Considered
- *Expanding `firestore-service.ts`*: Rejected. `firestore-service.ts` is an infrastructure client wrapper for database operations, not a multi-tier cache coordinator.
- *Abstract Base Query Service*: Rejected. Inheritance introduces tight coupling across disparate financial domains. Composition via an injected engine is cleaner and more testable.

---

## 2. Integration with Centralized `CacheService` Layer

### Context
The user explicitly specified: *"also the engine should use the cache service layer"*. The project provides [`api/src/cache/cacheService.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/cache/cacheService.ts), which wraps `FlatCacheProvider` on disk and implements `ICacheProvider`.

### Decision
`HistoryResolutionEngine` will directly inject `cacheService`:
```typescript
import { cacheService } from '../cache/cacheService.js';
```
All persistent caching operations will invoke:
- `cacheService.get<T>(cacheKey)`
- `cacheService.set<T>(cacheKey, data, ttlSeconds)`

### Cache Keys & TTLs
- **Sealed Weeks**: Key `${collection}_weekly/${weekKey}` with TTL = `30 * 24 * 60 * 60` (30 days).
- **Active Week (Off-Hours)**: Key `${collection}_weekly/${currentWeekKey}` with TTL = `secondsUntilNextMarketOpen`.
- **Pure Past Range (Stocks/Indices Single-Symbol)**: Key `${asset}_range:${symbol}:${start}:${end}` with TTL = `30 * 24 * 60 * 60` (30 days).

---

## 3. Active Week RAM Synchronization for Bank Rates

### Context
In [`api/src/services/firestore-listener-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/firestore-listener-service.ts#L28-L32), `watchedWeeklyCollections` currently monitors:
- `stock_prices_history_weekly`
- `index_prices_history_weekly`
- `gold_prices_history_weekly`

`rates_history_weekly` is absent. Meanwhile, `serverless-scraper` already writes daily snapshots into `rates_history_weekly/{weekKey}` whenever bank rates shift.

### Decision
Add `"rates_history_weekly"` to `watchedWeeklyCollections` in `FirestoreListenerService`.

### Rationale
- Completes Principle XIII compliance for Bank Rates: Active Week queries resolve directly from [`ActiveWeeklyBucketStore`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/active-weekly-bucket-store.ts) in volatile memory with **`<0.1ms` latency and 0 Firestore database reads**.
- Ensures cold-boot pre-seeding from `flat-cache` and automated ISO week rollover sealing applies uniformly to bank rates.

---

## 4. Range Query Strategy for Bank Rates (Combinatorial Key Avoidance)

### Context
Unlike Stock/Index history ranges (which query a single ticker symbol, e.g. `COMI`), Bank Rates ranges accept multi-parameter filters (`bank`, `currency`). Caching the filtered output of every possible permutation in `flat-cache` would produce hundreds of cache keys.

### Decision
Bank Rates range queries will slice the ~26 **Weekly Buckets** directly from `ActiveWeeklyBucketStore` (in RAM) and `cacheService` (on disk) across the requested interval, applying filtering and sorting in memory. No secondary `rates_range:*` entries will be written to disk.

### Rationale
- Slicing and filtering ~26 weekly documents in memory takes **`<2ms`**.
- Eliminates cache key explosion on disk.

---

## 5. Strict Segregation of Intraday Timeline from History

### Context
Previously, query services attempted to recover yesterday's intraday points via `intradayCoordinator.recoverYesterdayPoints` and inject them into historical daily archives, causing bloat and coupling live trading session logic to past history.

### Decision
Per GLOSSARY.md and ADR 0006:
- The **Intraday Timeline** is strictly a volatile live session snapshot maintained in memory via `intradaySnapshotService`.
- The **History Resolution Engine** retrieves strictly daily closing summaries (open, high, low, close, extremes) from Weekly Buckets. It never fetches, recovers, or injects intraday tick arrays.

---

## 6. Gold Single-Date Endpoint Parity

### Context
`GET /v1/gold/history/:date` in [`gold-routes.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/routes/gold-routes.ts#L109) and the yesterday price delta in `GET /v1/gold` call `fetchGoldPriceHistoryRange(date, date)` and extract `[0]`.

### Decision
Refactor `GoldQueryService` to expose `fetchGoldPriceHistory(date)` which directly invokes `historyResolutionEngine.getDateSnapshot('gold', date)`. Use this method for both the single-date route and yesterday's delta calculation in `GET /v1/gold`.

---

## 7. GCC Regional Markets Parameterization

### Context
`GccStockQueryService` duplicates manual date math and weekly document fetching for Saudi, UAE, and Qatar equities.

### Decision
Support `market?: 'saudi' | 'uae' | 'qatar'` in the engine, mapping dynamically to:
- `stock_prices_saudi_history_weekly`
- `stock_prices_uae_history_weekly`
- `stock_prices_qatar_history_weekly`

All regional weekly buckets are fetched through `cacheService` and cached identically.
