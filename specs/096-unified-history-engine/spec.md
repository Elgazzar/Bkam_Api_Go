# Feature Specification: Unified History Resolution Engine

**Feature Branch**: `096-unified-history-engine`

**Created**: 2026-10-24

**Status**: Draft

**Input**: User description: "Implement 0006-unified-history-resolution-engine.md ,also the engine should use the cache service layer"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Historical Single-Date Quote Retrieval (Priority: P1)

As a financial app user or mobile consumer, I want to retrieve historical market quotes for any asset class (Bank Rates, Gold, Stocks, Indices, GCC Equities) on a specific calendar date so that I can inspect closing prices and verify past market conditions.

**Why this priority**: Single-date history lookups power the primary daily archive views across all client applications. Unifying the resolution pipeline under a single engine guarantees uniform caching, eliminates redundant database reads, and fixes existing gaps (such as Bank Rates active week caching and Gold range workarounds).

**Independent Test**: Can be fully tested by querying `GET /v1/rates/history/:date`, `GET /v1/gold/history/:date`, `GET /v1/stocks/history/:date`, and `GET /v1/indices/history/:date` for past dates, verifying consistent payload structure, sub-millisecond retrieval from cache, and identical negative caching on holidays.

**Acceptance Scenarios**:

1. **Given** a user requests closing quotes for a date within the active calendar week, **When** the API processes the request, **Then** it serves the daily closing snapshot directly from volatile RAM (`ActiveWeeklyBucketStore`) in `<1ms` with 0 Firestore document reads.
2. **Given** a user requests closing quotes for a sealed past week, **When** the request is received, **Then** the engine retrieves the weekly bucket from the persistent cache service layer (`flat-cache`) without database access.
3. **Given** a cache miss on a sealed past week, **When** the weekly bucket is fetched from Firestore, **Then** the engine caches the document in the cache service layer for 30 days and extracts the requested date's closing snapshot.
4. **Given** a user queries a date on a weekend or official market holiday, **When** the date is not found in the weekly bucket, **Then** the engine returns a 404 response and caches the negative lookup to prevent future database queries.

---

### User Story 2 - Multi-Day Historical Range Querying (Priority: P2)

As a financial analyst or mobile chart viewer, I want to request historical quotes across a multi-day or multi-month date range for any asset (e.g. Bank Rates with optional filters, Gold across all carats, or specific Stocks/Indices) so that I can render interactive price trend charts.

**Why this priority**: Charting is a core user engagement capability. Multi-day range queries previously suffered from inconsistent implementations (Bank Rates looping 180 times per-day vs. Stocks slicing weekly buckets).

**Independent Test**: Can be fully tested by querying `GET /v1/rates/history/range`, `GET /v1/gold/history/range`, `GET /v1/stocks/:symbol/history/range`, and `GET /v1/indices/:symbol/history/range` across multi-month intervals (e.g., 180 days) and verifying that all dates are sliced from consolidated weekly buckets via the cache service layer.

**Acceptance Scenarios**:

1. **Given** a user requests a multi-day range for Bank Rates with bank or currency filters, **When** the range spans multiple weeks, **Then** the engine slices the consolidated weekly buckets from the cache service layer and RAM, applies the filters in memory, and returns the sorted history in `<50ms` without issuing 180 individual date queries.
2. **Given** a user requests a historical range for Gold, **When** the query executes, **Then** the engine resolves the date interval across weekly buckets in the cache service layer and returns all carat quotes chronologically.
3. **Given** a user requests a historical range for a Stock or Index symbol, **When** the query executes, **Then** the engine retrieves the symbol's quotes from the weekly buckets via the cache service layer.

---

### User Story 3 - Gold Endpoint Parity & Yesterday Close Resolution (Priority: P3)

As a gold market observer, I want single-date gold queries and live gold yesterday price lookups to resolve via a dedicated, first-class single-date method rather than an artificial range query workaround, so that response times are minimized and architecture is clean.

**Why this priority**: Previously, `GET /v1/gold/history/:date` and `GET /v1/gold` (for yesterday's price delta) called `fetchGoldPriceHistoryRange(date, date)` and extracted the first array item. Unifying this under `getDateSnapshot('gold', date)` eliminates redundant range logic and array allocations.

**Independent Test**: Can be tested by invoking `GET /v1/gold/history/:date` directly and inspecting `GET /v1/gold` to confirm yesterday's buy/sell prices populate correctly via `getDateSnapshot`.

**Acceptance Scenarios**:

1. **Given** a client requests a specific date via `GET /v1/gold/history/:date`, **When** the route executes, **Then** it delegates to `getDateSnapshot('gold', date)` and returns the snapshot directly.
2. **Given** a client queries live gold prices via `GET /v1/gold`, **When** yesterday's price delta is needed, **Then** it resolves yesterday's closing snapshot via `getDateSnapshot('gold', yesterdayDate)` without invoking range machinery.

---

### User Story 4 - Regional Equities (GCC Markets) History Integration (Priority: P4)

As an equities investor trading on GCC markets (Saudi Tadawul, UAE DFM/ADX, Qatar QSE), I want historical regional stock quotes to resolve through the same high-speed, tiered caching pipeline as domestic EGX equities.

**Why this priority**: Ensures architectural uniformity across all geographic markets supported by the platform, removing duplicate manual week calculations from `gcc-stock-query-service.ts`.

**Independent Test**: Can be tested by requesting `GET /v1/stocks/:market/history/:date` and `GET /v1/stocks/:market/:symbol/history/range` for a GCC market (e.g., `saudi`) and verifying that the engine handles regional weekly buckets correctly.

**Acceptance Scenarios**:

1. **Given** a user requests historical quotes for a GCC market stock, **When** the request is processed, **Then** the engine uses the parameterized market identifier to resolve the corresponding regional weekly bucket from the cache service layer.

---

### Edge Cases

- **Market Closed / Holiday Dates**: When an requested date within an existing weekly document is absent, it represents a closed market day. The engine caches this negative result (24h for past dates, 5m for today) and returns 404 immediately on subsequent requests.
- **Active Calendar Week Queries**: During live trading sessions, active week queries resolve directly from volatile RAM (`ActiveWeeklyBucketStore`) with 0 database reads and bypass disk writes.
- **Pre-Migration Historic Dates**: If an old historical date precedes the weekly bucket schema, the engine gracefully falls back to legacy daily Firestore collections, caches the resulting document in the cache service layer, and returns the snapshot.
- **Server Cold Boot Outside Trading Hours**: On restart, the active week document is pre-seeded into RAM from the cache service layer (`flat-cache`) to ensure 0 cold-start latency before listeners bind.
- **Intraday Segregation Invariant**: Intraday tick curves (`*_intraday`) are strictly volatile and MUST NOT be merged, stitched, or persisted into historical weekly documents or disk cache stores by the history engine.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a centralized, modular `HistoryResolutionEngine` that mediates all historical single-date and range queries across all supported asset classes (Bank Rates, Gold, EGX Stocks, EGX Indices, and GCC Stocks).
- **FR-002**: System MUST integrate the `HistoryResolutionEngine` with the centralized `CacheService` layer (`ICacheProvider` / `FlatCacheProvider`) for all disk-based caching of sealed weekly documents.
- **FR-003**: System MUST register `rates_history_weekly` in the real-time listener service (`watchedWeeklyCollections` in `FirestoreListenerService`) to synchronize the active calendar week's bank rates in volatile RAM (`ActiveWeeklyBucketStore`).
- **FR-004**: System MUST enforce strict segregation between live Intraday Timelines and historical data: the `HistoryResolutionEngine` MUST strictly retrieve daily closing summaries (open, high, low, close, extremes) from weekly buckets without recovering or injecting raw intraday tick arrays.
- **FR-005**: System MUST expose a standardized method `getDateSnapshot(assetClass: AssetClass, date: string, market?: string)` that resolves a single calendar date through the tiered hierarchy: Active Week RAM -> negative cache -> persistent cache service layer -> Firestore weekly bucket -> legacy daily fallback.
- **FR-006**: System MUST expose a standardized method `getDateRange(assetClass: AssetClass, options: RangeOptions)` that slices ISO calendar weeks across the interval, batch-resolves weekly documents from RAM and the cache service layer, and aggregates sorted daily entries.
- **FR-007**: System MUST resolve Bank Rates history ranges by slicing weekly buckets in memory and disk cache without persisting secondary combinatorial `rates_range:*` cache keys.
- **FR-008**: System MUST route `GET /v1/gold/history/:date` and live gold yesterday price lookups through `getDateSnapshot('gold', date)`, replacing the previous range query workaround.
- **FR-009**: System MUST maintain an in-memory negative cache (`missingDatesCache`) for dates missing from weekly documents (weekends and holidays) to prevent redundant database lookups.
- **FR-010**: System MUST support regional market parameters (e.g., `saudi`, `uae`, `qatar`) so GCC equity history utilizes the identical engine pipeline without code duplication.
- **FR-011**: System MUST update `.specify/memory/ARCHITECTURE.md`, `.specify/memory/REQUIREMENTS.md`, and `MODULAR_COMPONENTS.md` to catalog the `HistoryResolutionEngine` per Principle X and Principle XV.

### Key Entities

- **Weekly Bucket**: Consolidated storage document in Firestore containing all daily market snapshots for an ISO 8601 calendar week (`YYYY-Www`).
- **Active Week**: The current ISO calendar week whose trading sessions are ongoing and whose weekly bucket is actively synchronized in volatile RAM (`ActiveWeeklyBucketStore`).
- **Sealed Week**: A concluded past ISO calendar week whose trading sessions are permanently complete, immutable, and persisted in `flat-cache` via `CacheService` with a 30-day TTL.
- **History Resolution Engine**: Centralized coordinator responsible for resolving historical market quotes across storage tiers (Active Week RAM store, `flat-cache` disk storage via `CacheService`, and Firestore weekly buckets).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of historical single-date and range queries across Bank Rates, Gold, EGX Stocks, EGX Indices, and GCC Stocks resolve through the centralized `HistoryResolutionEngine`.
- **SC-002**: Historical queries touching the active calendar week resolve from volatile RAM in under 1 millisecond with zero Firestore document read charges.
- **SC-003**: Sealed past week queries resolve from the local cache service layer in under 5 milliseconds on cache hits with zero database reads.
- **SC-004**: Multi-month historical range queries (e.g. 180 calendar days) resolve in under 50 milliseconds from the cache service layer without executing 180 individual daily disk lookups.
- **SC-005**: Zero raw intraday tick points are written to `flat-cache` disk storage or embedded inside historical weekly documents.
- **SC-006**: Zero Firestore database reads occur when querying known market holidays and weekend dates that have been negatively cached.

## Assumptions

- Bank rate weekly buckets (`rates_history_weekly/{weekKey}`) are already created and updated by the background scraper (`serverless-scraper`), matching the pattern of gold and stock weekly buckets.
- The `CacheService` singleton (`api/src/cache/cacheService.ts`) backed by `FlatCacheProvider` remains the primary disk persistence mechanism for historical documents.
- All date inputs for historical queries are formatted as ISO 8601 strings (`YYYY-MM-DD`) and evaluated using the `Africa/Cairo` operational timezone for domestic assets, and respective regional market timezones for GCC assets.
- Live rate endpoints (`GET /v1/rates`, `GET /v1/gold`, etc.) continue to serve real-time data from their respective snapshot services in volatile memory.
