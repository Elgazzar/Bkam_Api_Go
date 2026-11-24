# Feature Specification: Deepen Historical Engine & Demote Leaky Firestore Storage Layer

**Feature Branch**: `110-deepen-history-engine`  
**Created**: 2026-11-15  
**Status**: Draft  
**Input**: User description: "Deepen Historical Engine & Demote Leaky Firestore Storage Layer\nsrc/services/history-resolution-engine.ts • src/services/firestore-service.ts • src/cache/history-cache-store.ts • src/cache/cacheService.ts • src/services/stock-query-service.ts\nProblem\nHistorical queries cascade through two redundant negative caches, two conflicting disk cache formats, and duplicated ~100-line range slicers in calling modules.\n\nSolution\nDeepen HistoryResolutionEngine to own range slicing and flat-cache persistence while demoting firestore-service.ts to a stateless database adapter."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Historical Multi-Day Range Resolution (Priority: P1)

As an API client querying historical price trends for equities, gold, benchmark indices, or foreign exchange rates over a custom date range (e.g. past 30 days or spanning into the active trading day), I want continuous time-series snapshots returned with sub-200ms response times, automatic caching of immutable past ranges, and seamless real-time stitching of the active session, so that mobile and web charts render without gaps or duplicate database lookups.

**Why this priority**: Range queries are the primary mechanism for financial charts and analytics across all client applications. Duplicated range slicing across domain callers currently leads to inconsistent prefix caching, fragile yesterday/today boundary math, and uncoordinated database reads.

**Independent Test**: Can be fully tested by requesting a multi-day range spanning both past sealed weeks and the active trading day for any asset (e.g., `GET /v1/stocks/COMI/history/range` or `GET /v1/gold/history/range`) and verifying that past days are served from persistent cache while today includes live quotes.

**Acceptance Scenarios**:

1. **Given** a requested date range entirely in the past (e.g., `2026-10-01` to `2026-10-15`), **When** the client submits the historical range request, **Then** the system returns the chronological array of daily quotes from disk cache without issuing database reads, with HTTP `Cache-Control: public, max-age=3600`.
2. **Given** a requested date range that starts in the past and extends through the active trading day, **When** the client submits the request, **Then** the system resolves the immutable past prefix from disk cache and appends the active session quote and intraday points, with HTTP `Cache-Control: public, max-age=60`.
3. **Given** repeated requests for the same date range and symbol, **When** subsequent calls are received, **Then** the complete range resolves directly from the unified persistent disk cache in $<10\text{ ms}$.

---

### User Story 2 - Consolidated Single-Date History & Negative Caching (Priority: P2)

As an API consumer requesting historical quotes for a specific date (e.g., a past banking day or an official holiday), I want immediate responses with single-point negative caching for non-trading days and single-source disk caching for trading days, so that closed market days do not trigger redundant cloud database reads or cascading cache lookups.

**Why this priority**: Eliminates redundant negative cache checking and dual disk formats across the historical engine and storage layers, preventing quota consumption on invalid or non-trading dates.

**Independent Test**: Can be fully tested by querying a known closed market date (e.g., an Egyptian Friday or weekend date) and verifying that the first request returns a 404/empty snapshot and subsequent requests for that date resolve immediately without hitting the cloud database.

**Acceptance Scenarios**:

1. **Given** a historical request for a past trading date present in weekly bucket storage, **When** the snapshot is fetched, **Then** the record is cached in the primary persistent disk cache and returned with complete quote details.
2. **Given** a historical request for a known non-trading date (weekend or official holiday) where the weekly bucket document exists but contains no entry for that date, **When** the request is processed, **Then** the system records a negative cache entry with a 24-hour expiration and returns `null` without attempting legacy daily database lookups.
3. **Given** a historical date that previously resulted in a negative cache entry, **When** repeated requests arrive, **Then** the system resolves immediately as `null` from the single centralized negative cache without issuing cloud database reads.

---

### User Story 3 - Stateless Database Adapter Isolation (Priority: P3)

As a system maintainer and infrastructure engineer, I want the cloud database adapter to perform pure, stateless reads of weekly bucket and legacy daily documents without maintaining local cache state, file stores, or scheduling calculations, so that storage adapters have zero state leakage, are isolated fault domains, and can be tested without complex file-system mocks.

**Why this priority**: Consolidates architectural locality so that all caching policies, TTLs, and cache invalidation reside in exactly one deep coordinator (`HistoryResolutionEngine`), demoting the storage adapter to a simple read bridge.

**Independent Test**: Can be tested by invoking the database adapter functions directly with simulated database client stubs and verifying they return raw document payloads without creating disk files, writing to disk folders, or managing internal cache maps.

**Acceptance Scenarios**:

1. **Given** an adapter request for a weekly bucket document, **When** the document exists in the database, **Then** the adapter returns the raw document dictionary and updates zero internal cache state.
2. **Given** an adapter request for an absent document, **When** the database returns a non-existent snapshot, **Then** the adapter returns `null` directly, leaving negative caching decisions to the calling engine.

---

### Edge Cases

- **Future Dates**: When a date range or snapshot specifies a date beyond the current operational calendar day in Cairo, the system MUST reject the request with HTTP `400 Bad Request` before consulting any cache or database.
- **Inverted Date Ranges**: When `startDate > endDate`, the system MUST reject the request immediately with HTTP `400 Bad Request`.
- **Pre-Migration Dates**: For historical dates prior to weekly bucket aggregation where weekly documents do not exist, the engine MUST seamlessly fall back to legacy daily document reads via the database adapter and persist the resulting snapshot into the unified disk cache.
- **Corrupted Disk Cache**: If a disk cache entry is corrupted or unparseable, the engine MUST safely evict the invalid entry, re-fetch from storage, and re-persist without failing the client request.
- **Active Week Mutation**: During an active trading week, dates belonging to the current calendar week MUST reflect active intraday movements and not be sealed permanently to disk until the week has completed.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a single, centralized historical coordinator (`HistoryResolutionEngine`) responsible for resolving all single-date snapshots and multi-day range queries across foreign exchange rates, precious metals, equities, and benchmark indices.
- **FR-002**: `HistoryResolutionEngine` MUST encapsulate all multi-day range slicing, date iteration, and dynamic session stitching, accepting query parameters (`assetClass`, `startDate`, `endDate`, `symbol`, `market`) and returning chronological time-series arrays.
- **FR-003**: System MUST standardize all persistent disk caching on a single shared key-value cache (`flat-cache` via `CacheService`), completely eliminating the secondary file-based disk store (`HistoryCacheStore`) and the `var/history-cache/` directory.
- **FR-004**: System MUST maintain exactly one authoritative in-memory negative cache for non-trading dates and missing archives within `HistoryResolutionEngine`, removing duplicate negative cache maps from the database adapter.
- **FR-005**: `HistoryResolutionEngine` MUST cache sealed past ranges under deterministic keys (`{asset}_range:{symbol}:{startDate}:{endDate}` and `{asset}_range_prefix:...`) with a 30-day retention ceiling.
- **FR-006**: When a date range extends into the active trading day, `HistoryResolutionEngine` MUST automatically split the query into a cached immutable historical prefix and a dynamic live session suffix decorated with real-time intraday tick points.
- **FR-007**: Domain query services (`StockQueryService`, `GoldQueryService`, `IndicesQueryService`, `RatesQueryService`) MUST delegate range requests directly to `HistoryResolutionEngine`, deleting their internal duplicate range slicing, prefix caching, and yesterday synthesis implementations.
- **FR-008**: The cloud database adapter (`firestore-service.ts`) MUST be demoted to a stateless document reader exposing clean read methods (`fetchWeeklyBucketDocument`, `fetchDailyDocument`), stripped of internal disk stores, memory caches, negative caches, and market schedule dependencies.
- **FR-009**: All responses MUST preserve 100% backward-compatible payload shapes, status codes, and HTTP cache headers (`Cache-Control`, `ETag`) across existing REST endpoints.

---

### Key Entities

- **ResolvedDateSnapshot**: Standardized envelope encapsulating the daily market data payload for a specific calendar date along with temporal metadata (`isToday`, `date`, `cacheTtlSeconds`).
- **DateRangeQuery**: Parameterized specification of a time-series request including asset class, canonical symbol/market, start date, and end date.
- **RangePoint**: Standardized individual time-series observation containing calendar date, closing price, percentage change, timestamp, and optional intraday tick series.
- **WeeklyBucketDocument**: Immutable or mutating consolidated document aggregating all daily snapshots for a single ISO calendar week (`YYYY-Www`).

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of historical date range and single-date archive endpoints maintain identical contract and response structure across baseline (port 3000) and target deployments.
- **SC-002**: P95 response latency for cached historical range queries remains under 50ms across all asset domains.
- **SC-003**: Under repeated historical queries, direct cloud database read operations for sealed dates remain at zero (0).
- **SC-004**: Total lines of code dedicated to range slicing, prefix caching, and negative caching across domain query services and storage adapters are reduced by over 300 lines through consolidation into the history coordinator.
- **SC-005**: Exactly one persistent disk cache provider and exactly one negative cache instance exist for historical queries across the entire codebase.

---

## Assumptions

- Historical records prior to the introduction of weekly buckets remain accessible via legacy daily documents in the cloud database.
- Standard ISO week numbering (`YYYY-Www`) governs all weekly bucket partitioning.
- The shared `CacheService` (`flat-cache`) provides reliable atomic file-backed persistence with configurable TTL support.
- All date calculations and boundary checks operate within the official financial market timezone (`Africa/Cairo`).
