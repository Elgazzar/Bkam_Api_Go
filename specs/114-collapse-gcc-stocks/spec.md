# Feature Specification: Collapse Artificial GCC Stocks Boundary into Stock Module

**Feature Branch**: `114-collapse-gcc-stocks`

**Created**: 2026-11-15

**Status**: Draft

**Input**: User description: "1. Collapse the Artificial GCC Stocks Boundary into Stock Module. Affected Files & Modules: src/modules/gcc-stocks/services/gcc-stock-query-service.ts (621 lines — duplicate queries & enrichment), src/modules/gcc-stocks/routes/gcc-stock-routes.ts (717 lines — duplicate routing), src/modules/gcc-stocks/index.ts (86 lines — dynamic loader wrapper), src/modules/gcc-stocks/adapters/* (dead scraper code violating Principle VIII), src/services/stock-query-service.ts & src/routes/stock-routes.ts (absorbing seam). Problem: GCC stocks was isolated in an artificial sub-module loaded via dynamic string imports, replicating over 1,400 lines of query and route logic while leaking dead scraping adapters into the API repo. Solution: Eliminate src/modules/gcc-stocks/ entirely by absorbing regional equity queries and routes directly into the authoritative StockQueryService and stock-routes.ts."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Real-Time Quotes Across Domestic & Regional Exchanges (Priority: P1)

As a financial mobile application or web portal developer,
I need a single, consistent REST interface to query live equity prices across all supported exchanges (Egyptian Exchange, Saudi Tadawul, Dubai Financial Market, Abu Dhabi Securities Exchange, Qatar Stock Exchange, and Boursa Kuwait),
So that client applications can retrieve identical real-time quote schemas without diverging data structures or special-case regional routing logic.

**Why this priority**:
The core purpose of the gateway is high-performance real-time quote delivery. Unifying all 5 regional exchanges under one authoritative equity interface guarantees identical payload envelopes, zero schema divergence, and instantaneous RAM response times (<150ms).

**Independent Test**:
Can be fully tested by querying `GET /v1/stocks?market=saudi`, `GET /v1/stocks?market=uae`, `GET /v1/stocks?market=qatar`, `GET /v1/stocks?market=kuwait`, and `GET /v1/stocks` (defaulting to Egypt), verifying that all responses return `200 OK` with valid Zod-validated snapshot schemas, correct currency designations (SAR, AED, QAR, KWD, EGP), and active freshness indicators.

**Acceptance Scenarios**:
1. **Given** the gateway is operating with pre-warmed memory snapshots, **When** a client sends `GET /v1/stocks?market=saudi`, **Then** the system returns `200 OK` with the full Saudi Tadawul stock snapshot containing currency `SAR`, country `SA`, valid company metadata, and sub-150ms latency.
2. **Given** a client requests quotes for Egypt without specifying a market parameter, **When** the client sends `GET /v1/stocks`, **Then** the system defaults to the Egyptian Exchange (EGX) and delivers `200 OK` with currency `EGP`.
3. **Given** an invalid or unknown market parameter (e.g. `?market=london`), **When** the request is evaluated, **Then** the gateway responds with `400 Bad Request` and a structured error message listing valid supported markets.

---

### User Story 2 - Full Backward Compatibility for Existing Regional Route Paths (Priority: P1)

As an existing consumer of the Bkam API (e.g., Android app v2.x and legacy web dashboards),
I need existing regional route aliases (`/v1/stocks/saudi`, `/v1/stocks/uae`, `/v1/stocks/qatar`, `/v1/stocks/kuwait`, and `/v1/stocks/regional/catalog`) to continue functioning identically with 100% structural parity,
So that no client app updates or configuration changes are required when the underlying architecture is unified.

**Why this priority**:
Per Constitution Principle I (API-First Contract), breaking changes to client-facing routes are strictly forbidden. Existing mobile clients actively query these paths in production and must experience zero regressions.

**Independent Test**:
Can be fully tested by running `npm run verify:diff` against baseline port 3000, confirming that all 5 regional stock routes and catalog endpoints achieve `✔ MATCH` with zero structural divergence and identical HTTP status codes.

**Acceptance Scenarios**:
1. **Given** legacy client requests to `GET /v1/stocks/saudi`, **When** processed by the unified route handler, **Then** the response matches the legacy endpoint output byte-for-byte in structure, schema keys, and status code `200 OK`.
2. **Given** a request to `GET /v1/stocks/regional/catalog`, **When** evaluated, **Then** the gateway returns `200 OK` with the multi-market security inventory grouped by country.

---

### User Story 3 - Historical Archive & Range Queries Across Any Market (Priority: P2)

As a technical analyst or chart-rendering client,
I need to query single-date historical snapshots and multi-week historical ranges for any stock on any supported exchange using a unified syntax,
So that chart components can display historical trends without having to branch into different endpoint conventions depending on the exchange.

**Why this priority**:
Consolidates the fragmented historical routing paths (`/v1/stocks/:market/history/:date` vs `/v1/stocks/history/:date?market=...`) into a coherent, standardized historical query interface powered by the centralized tiered history cache.

**Independent Test**:
Can be tested by executing historical queries for a past date across all supported markets (e.g. `GET /v1/stocks/history/2026-11-12?market=saudi`) and verifying `200 OK` with daily closing summaries served directly from disk cache or memory buckets.

**Acceptance Scenarios**:
1. **Given** a valid ISO date in the past, **When** a client queries `GET /v1/stocks/history/2026-11-12?market=saudi`, **Then** the system returns `200 OK` with the Saudi market historical closing snapshot.
2. **Given** a historical range query for a specific ticker symbol (e.g. `GET /v1/stocks/1120/history/range?start=2026-11-01&end=2026-11-12&market=saudi`), **When** evaluated, **Then** the system returns `200 OK` with the sorted daily historical records without fetching raw intraday tick series.

---

### User Story 4 - Dedicated Single-Hop Intraday Timelines for Regional Securities (Priority: P3)

As a mobile app rendering live trading charts for GCC equities (e.g. Al Rajhi Bank, Emaar, Qatar National Bank),
I need to query the latest intraday tick points via `GET /v1/stocks/:symbol/intraday/latest?market=...`,
So that the chart renders today's active session or gracefully backtracks to the most recent concluded trading session with zero client-side weekend math.

**Why this priority**:
ADR 0007 established dedicated latest intraday timeline endpoints. Ensuring GCC equities are fully unified under this single seam guarantees consistent chart behavior across domestic and international assets.

**Independent Test**:
Can be tested by querying `GET /v1/stocks/1120/intraday/latest?market=saudi` and verifying `200 OK` with the standardized intraday envelope containing `intraday` tick points and session metadata.

**Acceptance Scenarios**:
1. **Given** an active or recently concluded trading session for a regional security, **When** queried via `GET /v1/stocks/:symbol/intraday/latest?market=saudi`, **Then** the gateway delivers the tick timeline with appropriate `Cache-Control` headers (60s during open market, 900s when closed).
2. **Given** a valid catalog symbol that experienced zero trades during the session, **When** queried, **Then** the system returns `200 OK` with `intraday: []` rather than `404 Not Found`.

---

### Edge Cases

- **Case 1: Country Code vs Market Name Alias Resolution**:
  What happens when a caller passes `?market=SA` (country code) instead of `?market=saudi` (market identifier)?
  The system MUST normalize both inputs equivalently to the canonical market definition.
- **Case 2: Weekend / Closed Session Temporal Alignment**:
  How does the system handle queries for GCC markets on Sundays (when EGX is closed, but Tadawul is active) or on Fridays (when GCC is closed, but international markets operate)?
  The unified engine MUST evaluate trading status according to the target market's specific IANA timezone and calendar rules via the centralized market schedule engine.
- **Case 3: Unwarmed / Cold Boot Snapshot Access**:
  What happens when a regional equity query arrives during application startup before Firestore listener synchronization?
  The system MUST serve the pre-warmed snapshot hydrated from local disk cache, maintaining sub-150ms latency with zero direct database reads.
- **Case 4: Residual Scraper Execution Prevention**:
  Can external scraping daemons or background cron scripts trigger execution inside the API process?
  By removing all residual scraper adapters (`TradingViewSaudiAdapter`, `uae-stock-adapter`, `qatar-stock-adapter`), the API process is completely isolated from scraping operations per Constitution Principle VIII.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST serve real-time stock snapshots for all supported exchanges (EGX, Saudi, UAE, Qatar, Kuwait) through the authoritative `StockQueryService`.
- **FR-002**: The system MUST preserve 100% backward-compatible HTTP route support for legacy paths:
  - `GET /v1/stocks/saudi`
  - `GET /v1/stocks/uae`
  - `GET /v1/stocks/qatar`
  - `GET /v1/stocks/kuwait`
  - `GET /v1/stocks/regional/catalog`
- **FR-003**: The system MUST accept a parameterized `?market=` query parameter on `GET /v1/stocks` accepting both market aliases (`saudi`, `uae`, `qatar`, `kuwait`, `egx`) and ISO country codes (`sa`, `ae`, `qa`, `kw`, `eg`).
- **FR-004**: The system MUST serve all real-time equity snapshots directly from in-memory RAM cache with zero (0) direct Firestore database reads during active request handling.
- **FR-005**: The system MUST route single-date historical equity queries (`GET /v1/stocks/history/:date`) and multi-day range queries (`GET /v1/stocks/:symbol/history/range`) across all regional markets through the centralized `HistoryResolutionEngine`.
- **FR-006**: The system MUST provide dedicated latest intraday timeline queries (`GET /v1/stocks/:symbol/intraday/latest`) for regional securities through the centralized `IntradayCoordinator`.
- **FR-007**: The system MUST eliminate the duplicate `GccStockQueryService` (621 lines) and `gcc-stock-routes.ts` (717 lines), centralizing all stock query and route handling into single authoritative files.
- **FR-008**: The system MUST eliminate dynamic string-based module imports (`import("../../../services/...")`) and restore compile-time static type resolution throughout the stock subsystem.
- **FR-009**: The system MUST completely purge all scraper daemon adapters (`TradingViewSaudiAdapter`, `uae-stock-adapter.ts`, `qatar-stock-adapter.ts`) and scraper orchestration runners from the repository to comply strictly with Constitution Principle VIII.
- **FR-010**: All returned stock payloads MUST include standard freshness indicators (`freshness: "current" | "stale" | "degraded"`), data retrieval timestamps, and trace IDs (`x-trace-id`).
- **FR-011**: All code lines added or modified MUST include 100% explanatory commenting in accordance with Constitution Principle V.
- **FR-012**: The application version in `package.json` MUST be incremented in accordance with Constitution Principle IX before merging.

---

### Key Entities *(include if feature involves data)*

- **SecurityDefinition**: Normalized metadata model for tradable securities encapsulating ticker symbol, ISIN, sector, exchange, country, currency, TradingView ticker, and company name (Arabic and English).
- **StockPriceSnapshot**: Standardized collection envelope wrapping current equity prices, daily changes, percentages, trade volumes, and market timestamps.
- **IntradaySession**: Minute-by-minute trading tick series recorded during a single market session, associated with session date and concluding closing price.
- **MarketSchedule**: Authoritative operational parameters defining market trading days, local IANA timezones, and opening/closing bell hours.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: **100% Differential Parity**: All 31 differential parity endpoints in `npm run verify:diff` achieve `✔ MATCH` status with zero contract regressions against the baseline server.
- **SC-002**: **100% Verification Gate Pass**: All 37 scenarios in `npm run verify` pass with exit code `0`.
- **SC-003**: **Zero Direct Database Reads**: 100% of real-time stock endpoint queries are resolved from volatile RAM cache or local disk storage.
- **SC-004**: **Codebase Simplification**: Over 1,400 lines of redundant routing, query, and scraper code are permanently deleted from `src/modules/gcc-stocks/`.
- **SC-005**: **Performance Guarantee**: Response latency for all warmed stock endpoints remains strictly below the 150ms SLA ceiling.
- **SC-006**: **Constitution Compliance**: Zero scraping code remains within the API codebase, satisfying Principle VIII.

---

## Assumptions

- **Upstream Push Synchronization**: Background data push listeners (`FirestoreListenerService`) continue to deliver live Firestore updates into `LiveSnapshotCoordinator` across all market partitions (`sa`, `ae`, `qa`, `kw`, `egx`).
- **Catalog Static Availability**: Security definitions for GCC markets in `src/config/catalogs/` (or consolidated security models) remain available to hydrate `MarketCatalog` in memory.
- **Existing Client Contracts**: Clients expect identical JSON property names and types on existing endpoints; no response properties are renamed or removed.
