# Feature Specification: Migrate Bkam Financial API Gateway to Idiomatic Go

**Feature Branch**: `116-migrate-gateway-go`  
**Created**: 2026-11-16  
**Status**: Draft  
**Input**: User description: "Migrate Bkam Financial API Gateway from Node.js (TypeScript) to Idiomatic Go with Firestore snapshot feature, multi-tier historical cache, and deep module architecture"

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Real-Time Market Quotes via Push Streaming (Priority: P1) 🎯 MVP

As an active consumer of financial market data (bank currency rates, precious metals, petroleum, EGX stocks, and GCC equities), I want real-time price updates pushed from the upstream database to be reflected instantly in gateway API responses so that my application always receives the freshest available market quotes without polling delays or stale intermediary caching.

**Why this priority**: Real-time market quote delivery is the primary core capability of the gateway. Without accurate and instant quote resolution, downstream consumers cannot display authoritative rates or trigger alerts.

**Independent Test**: Can be tested independently by publishing new price events to upstream database collections and verifying that GET queries across `/v1/rates`, `/v1/gold`, `/v1/gas`, `/v1/stocks`, and `/v1/indices` immediately return the updated figures with sub-millisecond in-memory lookups.

**Acceptance Scenarios**:
1. **Given** real-time currency rate updates are published to upstream bank collections, **When** a client sends a GET request to `/v1/rates` or `/v1/rates/:currency`, **Then** the gateway responds with the updated rates, accurate trend indicators (`up` or `down`), and metadata within SLA limits.
2. **Given** real-time price changes occur for gold karats or fuel varieties, **When** a client queries `/v1/gold` or `/v1/gas`, **Then** the gateway serves the latest prices directly from memory with zero database read operations.
3. **Given** regional equity updates are pushed for Saudi, UAE, Qatar, or Kuwait markets, **When** a client queries `/v1/stocks/:market`, **Then** the response reflects the updated quotes normalized to the respective market currency.

---

### User Story 2 - Authoritative Historical Archives & Range Slicing (Priority: P2)

As a financial analyst or mobile charting user, I want to query historical daily archives and multi-day date ranges across calendar weeks so that I can inspect price trends over time without causing database load spikes or experiencing slow response times.

**Why this priority**: Charting and historical queries account for significant traffic during non-trading hours and historical review sessions. Serving active weeks from memory and sealed past weeks from local storage eliminates upstream database read costs while keeping query latency minimal.

**Independent Test**: Can be tested independently by executing single-date and range queries (`/v1/rates/history/:date`, `/v1/rates/history/range`, `/v1/stocks/history/:date`, `/v1/stocks/:symbol/history/range`) for past sealed dates and current active week dates, confirming identical JSON shapes and sub-15ms responses with zero upstream read calls.

**Acceptance Scenarios**:
1. **Given** a requested date falls within the currently mutating calendar week, **When** a client requests a historical single-date or range endpoint, **Then** the gateway retrieves daily quotes directly from its primary in-memory active week cache.
2. **Given** a requested date range spans prior completed calendar weeks, **When** a client requests a historical range endpoint, **Then** the gateway resolves the range by slicing persistent local storage archives with zero upstream database queries.
3. **Given** a date range query ends on the active calendar day, **When** the range is evaluated, **Then** the gateway seamlessly synthesizes and appends the live session quote into the historical time series.

---

### User Story 3 - High-Throughput Batch Quotes & Intraday Timelines (Priority: P3)

As a mobile application loading its home dashboard or detailed ticker views, I want to query heterogeneous asset vectors in a single request and inspect intraday tick progressions so that the client minimizes HTTP round-trips and renders real-time sparklines efficiently.

**Why this priority**: Consolidating requests into batch queries (`/v1/quotes`) drastically reduces mobile device network usage and latency, while intraday timelines enable interactive visualization of intraday volatility.

**Independent Test**: Can be tested independently by executing POST and GET batch requests against `/v1/quotes` with mixed asset keys (e.g. `CBE_USD,GOLD_21K,EGX_COMI,TASI`) and intraday timeline routes (`/v1/stocks/:symbol/intraday/latest`, `/v1/rates/:currency/intraday/latest`), verifying item-level envelope consistency and fallback resolution.

**Acceptance Scenarios**:
1. **Given** a batch request specifying multiple asset identifiers across currencies, commodities, and equities, **When** `/v1/quotes` is invoked, **Then** the response contains resolved quote envelopes for each valid asset identifier with individual freshness markers.
2. **Given** a client requests intraday tick timeline points for a stock or currency, **When** today's trading session has not yet commenced or is in off-hours, **Then** the gateway automatically backtracks and serves the concluding intraday points of the most recent completed trading session.

---

### User Story 4 - High-Performance Gateway Runtime & Operational Telemetry (Priority: P4)

As a platform operator and systems administrator, I want the gateway to enforce API key security, maintain sub-15ms response latency under heavy concurrent load, provide accurate operational metrics, and safely handle timezone-aligned rollover transitions with zero service interruption.

**Why this priority**: Operational reliability, SLA compliance, security barriers, and resource efficiency are essential for maintaining 24/7 uptime on compute resources while guaranteeing consistent service delivery.

**Independent Test**: Can be tested independently by executing load tests, triggering midnight and 8:00 AM scheduled rollovers, probing `/v1/metrics`, `/v1/health`, and `/v1/version`, and verifying conditional HTTP revalidation (`304 Not Modified`).

**Acceptance Scenarios**:
1. **Given** an incoming request with an invalid or missing API key, **When** the gateway receives the request, **Then** it rejects the request with HTTP `401 Unauthorized` without executing downstream domain logic.
2. **Given** an un-mutated cached resource with a matching client `If-None-Match` header, **When** evaluated, **Then** the gateway responds with HTTP `304 Not Modified` and zero payload body.
3. **Given** the clock reaches midnight (00:01 AM) and 08:00 AM Cairo operational boundaries, **When** the scheduled rollover executes, **Then** the prior calendar week is sealed to persistent storage, intraday listeners rotate, and in-memory caches re-align without dropped requests.

---

### Edge Cases

- **Cold Boot Before Upstream Push**: When the server restarts before the first upstream database push event is received, it must immediately hydrate its memory state from local persistent disk storage so that it never serves empty responses or crashes.
- **Upstream Connection Interruption**: If the real-time push streaming connection drops, the gateway must continue serving cached in-memory quotes, mark payloads as degraded if expiry thresholds pass, and automatically re-establish streams with backoff.
- **Midnight Week Transition Collisions**: If a range query arrives precisely while a calendar week boundary transitions, read-write locking must prevent race conditions or partial slice corruption.
- **Thundering Herd Concurrent Requests**: When multiple concurrent requests query a cold or missing historical date, the gateway must coalesce requests onto a single execution promise/flight to prevent duplicate disk reads or upstream queries.
- **Empty or Incomplete Scraped Data**: If a pushed document contains partial bank rate data or missing trend values, the engine must self-heal missing fields using historical closing values and fallback baselines rather than failing the request.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST implement an idiomatic, high-concurrency Go application providing 100% contract and route parity with all 49 existing endpoints of the Bkam API Gateway.
- **FR-002**: The system MUST subscribe to upstream database push streams using real-time snapshot listeners managed within isolated, cancellable worker goroutines.
- **FR-003**: The real-time listener adapter MUST dispatch pushed updates directly into the live snapshot coordinator and history resolution engine without circular or duplicate intermediate service mutations.
- **FR-004**: The system MUST maintain a two-tier historical cache architecture:
  - **Tier 1 (RAM)**: In-memory thread-safe dictionary caching consolidated weekly buckets for the active mutating ISO calendar week (`YYYY-Www`).
  - **Tier 2 (Disk / Local KV)**: Persistent local storage caching sealed past calendar weeks with a 30-day retention period (TTL).
- **FR-005**: The system MUST resolve warmed historical date archives and multi-day ranges with zero (0) upstream database read operations.
- **FR-006**: The system MUST enforce single-flight request coalescing across all cache-miss queries and live fetches to eliminate thundering herd concurrency spikes.
- **FR-007**: The system MUST evaluate trading session hours, market holidays, and dynamic cache-control TTLs using the authoritative `Africa/Cairo` operational timezone (`*time.Location`).
- **FR-008**: The system MUST automatically execute boundary rollovers:
  - Weekly rollover at 00:01 AM Cairo time on Mondays: seals the prior week to persistent storage and re-targets active week listeners.
  - Daily rollover at 08:00 AM Cairo time: rotates intraday document listeners to the new operational business date.
- **FR-009**: The system MUST support heterogeneous batch quote queries (`POST /v1/quotes` and `GET /v1/quotes?items=...`), returning normalized quote envelopes with item-level freshness indicators.
- **FR-010**: The system MUST provide graceful intraday session fallback, serving the previous completed trading session's timeline points when current intraday data is not yet available.
- **FR-011**: The system MUST enforce API key authentication across all data endpoints via `x-api-key` request header and `?apiKey=` query parameter, while allowing unauthenticated access to `/v1/health` and `/v1/version`.
- **FR-012**: The system MUST generate deterministic HTTP `ETag` headers and evaluate `If-None-Match` conditional requests, returning `304 Not Modified` when content is unchanged.
- **FR-013**: The system MUST maintain thread safety with zero data races (`go test -race` passing cleanly) across all shared in-memory stores and coordinators.
- **FR-014**: All source code files MUST adhere to the Deep Module philosophy, exposing simple, expressive interfaces backed by substantive domain logic while avoiding shallow pass-through facades.
- **FR-015**: Every line of code MUST feature exhaustive documentation comments detailing logic, rationale, and runtime behavior.

---

### Key Entities

- **Live Snapshot Envelope**: Unified domain entity representing the latest market state for an asset class (bank rates, gold, petroleum, EGX stocks, GCC equities, indices), containing timestamp, source trace ID, freshness status, currency, and asset-specific payloads.
- **Active Weekly Bucket**: An in-memory data dictionary mapping ISO calendar dates (`YYYY-MM-DD`) to daily asset summaries for the currently mutating ISO week (`YYYY-Www`).
- **Sealed Historical Archive**: Immutable daily snapshot record or weekly consolidated file stored in persistent local storage with a 30-day lifecycle.
- **Intraday Timeline Point**: Time-series tick recording timestamp, price, change, and trading volume for charting sparklines and intraday volatility.
- **Market Schedule State**: Evaluation envelope indicating whether a market exchange (EGX, Tadawul, DFM, ADX, QSE, KSE) is currently open, holiday status, and seconds remaining until the next scheduled trading session.
- **Batch Quote Vector**: Consolidated multi-asset envelope containing resolved quotes for requested asset keys across all supported financial domains.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% parity across all 49 existing API gateway endpoints, verified by passing the full automated verification test suite with zero status code or structural payload deviations.
- **SC-002**: P99 response latency for all warmed in-memory quote and historical queries is reduced to **under 15 milliseconds** (substantially outperforming the existing 150ms ceiling).
- **SC-003**: Standard warmed HTTP request handling generates **zero (0) upstream database read operations**, strictly protecting cloud tier budgets.
- **SC-004**: System handles at least 5,000 concurrent HTTP requests per second on a 4-core ARM environment without latency degradation or dropped connections.
- **SC-005**: Zero data races detected under full concurrency verification using `go test -race`.
- **SC-006**: Cold boot time from process start to fully warmed serving state completes in **under 500 milliseconds**.

---

## Assumptions

- **Upstream Data Protocol**: Upstream database collections and documents follow the canonical schema structure currently utilized by the production scrapers (`rates`, `gold_prices/latest`, `gas_prices/latest`, `stock_prices/latest`, `index_prices/latest`, regional GCC stock collections, and `*_history_weekly` collections).
- **Host Architecture**: Deployment environment is an OCI Ampere A1 Compute instance running Linux ARM64 with Go 1.23+ runtime installed.
- **Storage Availability**: A local persistent filesystem directory (configured via `SNAPSHOT_DIR`) is available for caching sealed weekly historical archives and cold-boot live snapshots.
- **Timezone Availability**: The host operating system provides standard IANA timezone database data including `Africa/Cairo` and regional Gulf timezones.
- **Client Backwards Compatibility**: Existing client applications will transition to the Go gateway seamlessly via reverse proxy DNS routing with zero client-side code changes required.
