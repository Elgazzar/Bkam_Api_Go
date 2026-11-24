# Feature Specification: Deepen Intraday Coordinator

**Feature Branch**: `107-deepen-intraday-coordinator`

**Created**: 2026-11-15

**Status**: Draft

**Input**: User description: "Candidate 01 - Deepen Intraday Coordinator (Absorb Quote Synthesis & Session Backtracking). Problem: IntradayCoordinator is a shallow module that only returns raw points maps, forcing 4 caller modules to duplicate 600+ lines of quote extraction, session backtracking, previous-price baselining, and envelope packaging. Solution: Deepen IntradayCoordinator behind a single high-leverage interface method (getLatestIntraday) that returns the complete, finalized intraday envelope with quotes, points, baselines, and caching metadata."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Active Session Intraday Delivery (Priority: P1) 🎯 MVP

As a mobile application or financial web client,
I want to query real-time intraday price movement timelines for any financial instrument (bank currency, gold karat, regional equity stock, or market index) using a single standardized endpoint,
So that my application can render interactive daily volatility charts and price baselines without custom client-side assembly.

**Why this priority**: Core intraday experience for users during active market hours. Delivering live session ticks, current quotes, and previous session baselines in a unified response is essential for chart rendering across all 4 asset domains.

**Independent Test**: Can be fully tested by querying the `/v1/*/intraday/latest` endpoints during active trading hours for Bank FX (e.g. `USD`), Gold (e.g. `21k`), Equities (e.g. `COMI`), and Benchmark Indices (e.g. `EGX30`). All responses return complete daily tick series, current price quotes, and previous closing price baselines with zero client-side assembly.

**Acceptance Scenarios**:
1. **Given** an ongoing trading session, **When** a user requests the latest intraday timeline for a specific asset, **Then** the system returns the chronological sequence of price shifts recorded during the session.
2. **Given** an active trading session, **When** the intraday timeline is delivered, **Then** the response includes the active quote (current price, absolute change, percentage change) alongside the previous session's closing baseline price.
3. **Given** a request for an asset that has recorded zero price movements during the active session, **When** the timeline is requested, **Then** the system returns an empty timeline points array accompanied by current baseline quote prices rather than failing.

---

### User Story 2 - Automated Session Backtracking During Off-Hours (Priority: P2)

As a financial charting widget or market observer,
I want the latest intraday timeline to automatically display the most recent trading session when queried outside market operating hours (overnight, weekends, or statutory market holidays),
So that users viewing charts off-hours still see meaningful price movement patterns from the last active trading session instead of blank or broken charts.

**Why this priority**: Crucial for off-hours user engagement. Markets are closed overnight, on Fridays/Saturdays (Egypt), and on public holidays; charts must seamlessly display the concluding active session.

**Independent Test**: Request the latest intraday timeline for an asset when markets are closed (e.g. during weekend hours). Verify that the system delivers the complete timeline from the most recent active session, indicates that the session is concluded (`isToday: false`), and completes within the standard sub-200ms latency ceiling.

**Acceptance Scenarios**:
1. **Given** the market is closed and today has recorded no trading activity, **When** a user requests the latest intraday timeline, **Then** the system backtracks to the most recently concluded trading session (up to a 7-day lookback window) and delivers its sealed timeline.
2. **Given** a backtracked past session, **When** the response is delivered, **Then** it clearly indicates the session date and flags that the data represents a concluded historical session (`isToday: false`).
3. **Given** an asset with no recorded sessions within the 7-day lookback window, **When** the timeline is requested, **Then** the system safely returns an empty timeline with baseline prices rather than throwing an internal error.

---

### User Story 3 - Consolidated Multi-Asset Views & Resilient Identifier Resolution (Priority: P3)

As an API consumer building overview dashboards,
I want to query consolidated intraday views (such as all gold karats at once, or general bank currency rates without specifying a single bank) and use flexible identifiers (such as ticker aliases or karat variations),
So that my dashboard can fetch multi-asset timelines in a single round-trip without strict client-side parameter preprocessing.

**Why this priority**: Enhances client developer experience and efficiency, reducing round-trips for multi-karat gold boards and accommodating common ticker symbol variations (e.g., dual exchange tickers or karat abbreviations).

**Independent Test**: Query the gold intraday endpoint with `?item=all` and verify all karats (24, 21, 18, pound, international) are returned in a consolidated dictionary. Query an equity stock using its exchange alias (e.g., `BDC` or `BQDC`) and verify identical quote and timeline resolution.

**Acceptance Scenarios**:
1. **Given** a request for all gold commodities (`item=all`), **When** processed, **Then** the system returns timelines and quote baselines for all domestic karats (24k, 21k, 18k, pound) and international spot gold in a single envelope.
2. **Given** a request for a currency without a bank parameter, **When** processed, **Then** the system resolves benchmark bank timelines or aggregates available bank shifts for that currency.
3. **Given** a request referencing an equity or index alias, **When** processed, **Then** the system maps the alias to the canonical ticker and returns the expected intraday data.

---

## Edge Cases

- **Midnight Rollover (00:00 Africa/Cairo)**: When the operational calendar date rolls over at midnight before morning trading begins (typically 08:00–10:00 AM), today's memory buffer is empty. The system MUST automatically backtrack to yesterday's concluded session until today's opening rates are recorded.
- **Unrecorded Trading Days / Extended Holidays**: If a market remains closed for an extended holiday (e.g. Eid or long weekends), the system MUST backtrack day-by-day up to the 7-day ceiling to find the last trading session.
- **Asset with Zero Shifts**: If an equity or currency has not fluctuated since market open, the system MUST return the current morning price with an empty `intraday` array, preserving `previousPrice` baseline for calculation.
- **Invalid Asset Identifier**: If a user queries an unsupported symbol, bank, or karat (e.g., `invalid_symbol`), the system MUST reject the request with HTTP 400 Bad Request and a clear validation error envelope before attempting database queries.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a single, centralized intraday resolution capability capable of resolving the latest intraday timeline across all 4 supported asset domains: Bank Exchange Rates, Gold Commodities, Equities (Domestic and Regional GCC), and Benchmark Indices.
- **FR-002**: The intraday resolution capability MUST absorb live quote retrieval, intraday timeline point matching, previous closing price baselining, and response envelope construction behind a single interface.
- **FR-003**: When resolving an ongoing trading session, the system MUST serve intraday timeline points directly from in-memory stores with zero on-demand database reads ($0 cloud cost safeguard).
- **FR-004**: When resolving outside trading hours or before session open, the system MUST automatically backtrack to the latest recorded historical session within a 7-day ceiling and retrieve the concluding quotes from historical storage.
- **FR-005**: All intraday responses MUST include a previous session closing price baseline (`previousPrice`, `previousValue`, or `previousBuyPrice`/`previousSellPrice`) to enable client-side trend and volatility calculations.
- **FR-006**: The system MUST support consolidated multi-item queries, including all-commodity gold queries (`item=all`) and multi-bank currency lookups.
- **FR-007**: The system MUST support canonical asset identifier resolution, including karat formats (`24` vs `24k`), bank identifiers (case-insensitive), and stock ticker aliases.
- **FR-008**: The system MUST return standardized response envelopes indicating session freshness (`isToday: true` for ongoing sessions, `isToday: false` for concluded backtracked sessions).
- **FR-009**: Public REST API route signatures (`/v1/*/intraday/latest`), HTTP methods, query parameters, and JSON response bodies MUST remain 100% backwards-compatible with zero breaking changes.
- **FR-010**: Caller query modules (`RatesQueryService`, `GoldQueryService`, `StockQueryService`, `IndicesQueryService`) MUST delegate latest intraday operations directly to the deepened coordinator, eliminating duplicate quote synthesis and backtracking boilerplate.

---

### Key Entities

- **Intraday Timeline**: A chronological series of timestamped price readings (`time`, `price`) recorded during a single trading session.
- **Intraday Quote Envelope**: A unified representation of an asset's latest intraday state, containing current quote rates (price, change, change percentage), previous session closing baseline, session date, session timestamp, and the associated intraday timeline array.
- **Session Freshness State**: A boolean indicator (`isToday`) distinguishing ongoing active market sessions from concluded historical sessions.
- **Asset Domain**: One of the four platform financial categories: `rates` (foreign exchange), `gold` (precious metals), `stocks` (equities), or `indices` (market benchmarks).

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of latest intraday requests across all 4 asset domains resolve through the centralized intraday coordinator without caller quote-stitching logic.
- **SC-002**: Active session intraday queries complete within a 50ms latency ceiling by serving 100% from in-memory cache with zero database queries.
- **SC-003**: 0% regression across all 36 verification scenarios and 100% parity across all 30 differential comparison endpoints.
- **SC-004**: Eliminates over 600 lines of duplicate quote resolution, session backtracking, and baseline extraction boilerplate across the 4 caller query modules.

---

## Assumptions

- Operating market timezone for financial session boundaries is `Africa/Cairo`.
- The maximum permissible lookback window for off-hours session backtracking is 7 calendar days; sessions older than 7 days are treated as stale/inactive.
- Active trading sessions continue to be populated in volatile RAM by real-time push listeners with zero client polling database reads.
- Public route signatures (`GET /v1/rates/intraday/latest`, `GET /v1/gold/intraday/latest`, `GET /v1/stocks/:symbol/intraday/latest`, `GET /v1/indices/:code/intraday/latest`) remain completely unchanged.

