# Feature Specification: Dedicated Latest Intraday Timeline Endpoints

**Feature Branch**: `097-latest-intraday-api`

**Created**: 2026-10-24

**Status**: Draft

**Input**: User description: "create the specification for interday apis which we concluded during grill-with-docs session herer 0007-dedicated-latest-intraday-timeline-endpoints.md also make sure that during trading days we normally get the interday prices from snapshot which caches in ram I am not sure if there is a service for that"

## Clarifications

### Session 2026-10-24
- Q: Does the API support optional stock or bank and currency same as original API? → A: Keep single-asset contract (:symbol and :currency mandatory, bank optional via ?bank=, gold item optional) to prevent multi-megabyte payload bloat and optimize for sparkline/detail chart rendering.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Active Trading Day Intraday Timeline Retrieval (Priority: P1)

During active financial market trading sessions (e.g. 10:00 AM - 2:30 PM Cairo time for EGX, 10:00 AM - 3:00 PM Riyadh time for Tadawul), mobile and web clients query the latest intraday price movement timeline for a specific financial symbol or asset to display intraday sparkline and detail charts. The system immediately resolves and serves today's accumulated intra-session timeline directly from the volatile in-memory snapshot cache, returning the complete timeline along with current price, change, and freshness metadata in a single lightweight round-trip without querying database storage or downloading multi-day historical ranges.

**Why this priority**: This is the primary user journey for mobile financial charting. Under the previous architecture, mobile clients had to request broad date ranges (`/history/range`), search through multi-day payloads to locate today's date, and parse nested arrays. Serving the current session directly from memory provides instant chart rendering and zero database read overhead.

**Independent Test**: Can be fully tested by submitting a request to `GET /v1/stocks/COMI/intraday/latest` during an active trading session and verifying that the API returns today's date, live price, change metrics, and the intra-session point timeline in under 10ms with zero database queries.

**Acceptance Scenarios**:

1. **Given** an ongoing market session where price movements have been recorded today, **When** a client sends `GET /v1/stocks/:symbol/intraday/latest`, **Then** the system returns HTTP 200 OK with the resolved session `date` matching today, current quote metrics, and an `intraday` array containing all recorded timestamp-price pairs served directly from the in-memory snapshot cache.
2. **Given** an active trading session, **When** a client requests intraday timelines for benchmark indices (`/v1/indices/:symbol/intraday/latest`), gold karats (`/v1/gold/:item/intraday/latest`), or bank currencies (`/v1/rates/:currency/intraday/latest`), **Then** each respective domain resolves its intraday timeline from the in-memory snapshot store.
3. **Given** an active trading session where quotes are continuously mutating, **When** the response is delivered, **Then** the response includes `Cache-Control: public, max-age=60` to ensure mobile clients refresh data frequently while preventing server overload.

---

### User Story 2 - Off-Hours, Weekend, and Holiday Session Backtracking (Priority: P2)

When clients request intraday timelines outside market trading hours (e.g. Friday and Saturday weekends in Egypt and Saudi Arabia, Saturday and Sunday weekends for UAE markets, exchange holidays, or early morning hours before market open), today's intraday dataset is empty. Instead of returning an empty array or a 404 error, the system automatically backtracks to the most recent active trading session (up to 7 calendar days prior) and returns that session's complete timeline, clearly identifying the resolved trading date.

**Why this priority**: Mobile users frequently check charts in the evening, on weekends, or during holidays. If the API returned empty points or 404s, mobile apps would display blank charts or error banners. Backtracking guarantees a seamless charting experience 24/7 without requiring the client to understand market calendars or write complex fallback logic.

**Independent Test**: Can be tested by invoking `GET /v1/stocks/:symbol/intraday/latest` on a weekend or pre-market hour and verifying that the response returns the preceding trading day's date (e.g. Thursday's date for EGX), the complete intraday points recorded during that session, and an extended cache header (`max-age=900`).

**Acceptance Scenarios**:

1. **Given** a weekend, market holiday, or pre-opening hour where today has no recorded intraday points, **When** a client queries `GET /v1/stocks/:symbol/intraday/latest`, **Then** the system automatically queries the most recent recorded intraday session document (within a 7-day lookback window) and returns that session's `date` and complete `intraday` timeline.
2. **Given** an off-hours request where the latest session has been resolved from persistent storage, **When** subsequent client requests arrive during the same closed period, **Then** the system serves the resolved session from an in-memory buffer without repeating the database lookup.
3. **Given** a concluded, immutable market session is served, **When** the response headers are evaluated, **Then** the HTTP response includes `Cache-Control: public, max-age=900` (15 minutes).

---

### User Story 3 - Inactive Asset and Zero-Trade Graceful Resiliency (Priority: P3)

When a client queries the latest intraday timeline for a validly listed security or commodity that experienced zero trade executions or price shifts during the latest trading session (such as an illiquid small-cap stock or a stable currency pair), the system returns a standard HTTP 200 OK response with an empty timeline array (`intraday: []`) and its last known closing price, rather than throwing a 404 Not Found error.

**Why this priority**: Consistent response envelopes prevent client crashes. If inactive securities returned 404, mobile client applications would fail to render the screen header and company metadata, degrading user experience for less active stocks.

**Independent Test**: Can be tested by querying an illiquid catalog stock with zero trades during the latest session and verifying that the API returns HTTP 200 OK with `intraday: []`, valid latest closing price, and the resolved session date.

**Acceptance Scenarios**:

1. **Given** a security that exists in the market catalog but recorded zero price movements during the latest session, **When** a client queries its latest intraday endpoint, **Then** the system returns HTTP 200 OK containing the session `date`, current `price`, and `intraday: []`.
2. **Given** an invalid or unrecognized symbol that does not exist in the market catalog, **When** a client queries the endpoint, **Then** the system rejects the request with HTTP 400 Bad Request or 404 Not Found and a clear validation message.

---

### Edge Cases

- **Session Opening Rollover**: When the clock crosses the market opening bell and the first new trade ticks arrive, the system MUST seamlessly switch from serving the buffered previous session to serving today's live accumulating session without requiring service restarts.
- **Extended Multi-Day Holiday (Beyond 7 Days)**: If an exchange is closed for an exceptional period exceeding 7 consecutive calendar days (e.g. extended national emergency or market halt), the system MUST gracefully fall back to the most recent daily closing quote with `intraday: []` rather than hanging or failing.
- **Regional Market Schedules**: Different regional markets follow different operating hours and weekend days (e.g. EGX and Tadawul trade Sunday–Thursday, while UAE DFM and ADX trade Monday–Friday). The resolver MUST respect the specific market parameter (`?market=`) when determining trading sessions and catalogs.
- **Cold Server Startup During Off-Hours**: If the server starts cold on a weekend with an empty memory cache, the first request MUST execute a single database lookup to identify the latest session, buffer the result, and immediately serve subsequent requests from memory.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose dedicated, standardized RESTful endpoints for retrieving the latest intraday timeline across all platform asset classes:
  - **Stocks**: `GET /v1/stocks/:symbol/intraday/latest` (supporting optional query `?market=saudi|uae|qatar|kuwait`, defaulting to EGX)
  - **Benchmark Indices**: `GET /v1/indices/:symbol/intraday/latest` (supporting optional query `?market=...`)
  - **Gold**: `GET /v1/gold/:item/intraday/latest` (supporting items `24`, `21`, `18`, `pound`, `international`, or `all`) and `GET /v1/gold/intraday/latest`
  - **Bank Rates**: `GET /v1/rates/:currency/intraday/latest` (supporting optional query `?bank=...`)
- **FR-002**: The system MUST format all latest intraday timeline responses into a uniform, backward-compatible JSON envelope containing:
  - `symbol` (or asset identifier code)
  - `date` (ISO date string `YYYY-MM-DD` of the resolved trading session)
  - `currency` (ISO currency code, e.g. `EGP`, `SAR`, `AED`, `USD`)
  - `price` (numerical latest quote or closing price)
  - `change` (numerical price difference from previous session close)
  - `changePercent` (percentage price difference)
  - `updatedAt` (ISO-8601 timestamp string of the rate)
  - `intraday` (array of objects, each containing `time: string` in ISO-8601 format and `price: number`)
- **FR-003**: During active market trading sessions when today's in-memory snapshot store contains recorded price shifts, the system MUST resolve the timeline points directly from volatile RAM without performing database read operations.
- **FR-004**: When today's in-memory snapshot store contains no recorded intraday points (e.g. weekends, holidays, pre-opening hours), the system MUST query the most recent recorded intraday session document from persistent storage using a single descending query capped at a 7-day lookback window.
- **FR-005**: Once an off-hours session is resolved from persistent storage, the system MUST buffer the resolved timeline in memory to serve subsequent queries without repeated database read consumption.
- **FR-006**: When a catalog security has zero trade events in the resolved session, the system MUST return HTTP 200 OK with `intraday: []` and the latest closing price.
- **FR-007**: The system MUST emit adaptive HTTP `Cache-Control` response headers:
  - `public, max-age=60` during active open market sessions
  - `public, max-age=900` (15 minutes) for concluded, immutable past sessions
- **FR-008**: All latest intraday endpoints MUST enforce API key verification via the platform's standard authentication pre-handler.

### Key Entities

- **Latest Intraday Response**: The client-facing payload containing the resolved session date, price summary, and intra-session price points array.
- **Intraday Point**: An individual intra-session rate reading characterized by a precise timestamp (`time`) and a numerical price value (`price`).
- **Latest Trading Session**: The most recent active or completed financial market operating period that recorded price movements.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Mobile and web clients can retrieve the complete intraday timeline and current quote for any tracked asset in a single HTTP request, eliminating multi-day range downloads.
- **SC-002**: At least 95% of active trading session intraday requests resolve in under 10 milliseconds from in-memory snapshot cache.
- **SC-003**: 100% of weekend, holiday, and off-hours requests successfully resolve the preceding active trading session's timeline without returning empty datasets or 404 errors.
- **SC-004**: Valid catalog securities with zero trades return HTTP 200 OK with empty timeline arrays 100% of the time, causing zero client-side crashes.
- **SC-005**: Off-hours database reads are restricted to at most 1 read operation per session rollover, strictly preserving the platform's Always Free database quota limits.

## Assumptions

- Client applications render intraday charts from an array of timestamp-price pairs (`[{ time, price }]`) sorted in ascending chronological order.
- High-frequency intraday tick points are retained for active and immediate prior sessions; archival daily closing quotes are persisted separately in weekly historical summaries.
- Market schedules, timezones, and trading calendars defined in the platform's market registry accurately reflect regional exchange operations.
