# Feature Specification: Deepen History Resolution Engine

**Feature Branch**: `106-deepen-history-resolution`

**Created**: 2026-11-15

**Status**: Draft

**Input**: User description: "1. Deepen History Resolution Engine (Absorb Today vs Past Date Synthesis) - Problem: HistoryResolutionEngine has a shallow interface that only resolves past dates, forcing 4 query modules to leak Cairo timezone arithmetic, live RAM stitching, and intraday coordinator calls. Solution: Deepen HistoryResolutionEngine to absorb live today synthesis and intraday stitching behind a single interface taking assetClass and date."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Single-Date Market Snapshot Retrieval (Priority: P1)

As an API consumer (mobile application or web client),
I want to query market snapshots for any valid calendar date using a single consistent date parameter—regardless of whether the market session is currently active or concluded in the past—
So that my application can seamlessly display daily prices without client-side branching between live and historical views.

**Why this priority**: Core value of the platform. Eliminating the conceptual rift between "today's ongoing session" and "past archived sessions" simplifies all client integrations and provides a unified interface across all financial asset classes.

**Independent Test**: Can be fully tested by querying the date endpoint for yesterday (historical) and today (active) across Bank FX, Gold, Stocks, and Indices, confirming both return contract-compliant snapshots with identical top-level structures.

**Acceptance Scenarios**:
1. **Given** a valid past calendar date, **When** a user requests that date snapshot for any asset class, **Then** the system returns finalized historical closing prices from archive storage.
2. **Given** the current active operational date, **When** a user requests that date snapshot for any asset class, **Then** the system returns the latest live prices combined with intra-session timeline points.
3. **Given** a request specifying query filters (e.g., currency code `USD` or bank code `NBE`), **When** the query is executed for either today or a past date, **Then** the returned rates are accurately filtered while preserving metadata.

---

### User Story 2 - Continuous Multi-Day Historical Range Across Session Boundaries (Priority: P2)

As a financial charting widget or market analyst,
I want to query historical price ranges that span from past historical dates up to and including the current active day,
So that trend visualizations display a continuous, unbroken timeline from historical closing prices into the current trading session.

**Why this priority**: Essential for interactive charts and volatility analysis. Charting tools should not require two separate network requests to show historical trends plus today's current quote.

**Independent Test**: Request a date range starting 3 days ago and ending today. Verify that the response contains chronological daily data points where past days contain closed values and the final point contains today's active prices.

**Acceptance Scenarios**:
1. **Given** a date range spanning from a past date up to today, **When** the range query is requested, **Then** the system returns an ordered array combining finalized historical archives with today's live session quote.
2. **Given** a date range entirely in the past, **When** the range query is requested, **Then** the system returns immutable historical records.
3. **Given** non-trading days within the requested range (e.g., weekends or statutory market holidays), **When** the range query is requested, **Then** the system cleanly accounts for closed market sessions without corrupting adjacent active session data.

---

### User Story 3 - Operational Timezone Consistency and Caching Policies (Priority: P3)

As an API gateway operator,
I want date resolution to automatically enforce market operational timezone boundaries (Africa/Cairo) and apply optimal caching lifetimes (short-lived for active sessions, permanent for concluded archives),
So that bandwidth and database resources are conserved without serving stale data during active trading hours.

**Why this priority**: Protects cloud resource limits and ensures compliance with Always-Free tier budgets while guaranteeing real-time freshness during market hours.

**Independent Test**: Inspect response caching headers for today's date query versus a past date query. Verify that today's response permits short-term revalidation, while past date responses specify immutable long-term caching.

**Acceptance Scenarios**:
1. **Given** a request for today's active session, **When** the response is delivered, **Then** it includes dynamic caching directives allowing frequent client updates.
2. **Given** a request for a concluded past date, **When** the response is delivered, **Then** it includes aggressive long-term public caching directives.
3. **Given** a request with a future date parameter, **When** validated by the system, **Then** the request is rejected immediately with a clear validation error before any database or storage operations occur.

---

### Edge Cases

- **Midnight Session Rollover**: What happens at 00:00:00 Africa/Cairo time when today becomes yesterday? The system MUST dynamically evaluate calendar dates against Africa/Cairo time so active intraday memory points seal and transition gracefully into daily archive retrieval.
- **Future Date Rejection**: How does the system handle dates in the future (e.g., tomorrow)? The system MUST reject future dates with HTTP 400 Bad Request and a descriptive validation error message.
- **Missing Historical Day**: What happens if an older date has no recorded trading activity (e.g., unexpected data gap or holiday)? The system MUST return HTTP 404 with a standard error code explaining that no archive exists for that date.
- **Missing Intraday Points**: What happens if today's active session is queried before any price shifts occur? The system MUST deliver the baseline morning rates with an empty or initial intraday timeline rather than failing.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a single unified resolution interface capable of retrieving daily market snapshots for any valid calendar date across all supported asset classes (Bank FX, Gold, Equities, Benchmark Indices).
- **FR-002**: The system MUST absorb calendar day boundary evaluation (determining whether a queried date represents the current active session or a finalized past session) centrally within the history resolution capability.
- **FR-003**: When resolving the current active day, the system MUST retrieve live in-memory quotes, coordinate intra-session timeline points, and synthesize a complete daily snapshot.
- **FR-004**: When resolving a past calendar day, the system MUST retrieve the corresponding daily archive from tiered storage (in-memory bucket cache, disk cache, or cloud storage buckets).
- **FR-005**: Date range queries spanning past dates through the active day MUST synthesize a continuous, chronological timeline combining archived closing points with today's live quotes.
- **FR-006**: The system MUST evaluate all operational dates and session transitions strictly according to the official financial market timezone (`Africa/Cairo`).
- **FR-007**: The system MUST reject any queried date occurring in the future with a standardized 400 Bad Request error.
- **FR-008**: Single-date and date-range resolutions MUST support asset-specific filtering (such as currency codes, bank codes, and equity ticker symbols) with consistent filtering semantics across both active and historical dates.
- **FR-009**: Responses for the active trading day MUST carry short-lived cache headers (maximum 60 seconds), whereas responses for sealed historical dates MUST carry aggressive long-term cache headers (minimum 3600 seconds).
- **FR-010**: Existing public REST API route signatures, JSON contracts, and query parameters MUST remain 100% backwards-compatible with zero breaking changes for existing API consumers.

---

### Key Entities

- **Market Snapshot**: A structured representation of asset prices on a given date, containing price values, direction trends, timestamps, and optional intraday timeline points.
- **Trading Date**: An ISO-8601 calendar date (`YYYY-MM-DD`) evaluated within the `Africa/Cairo` operational timezone context.
- **Asset Class**: One of four financial instrument domains tracked by the gateway: `rates` (foreign exchange), `gold` (precious metals), `stocks` (equities), or `indices` (benchmark market indices).
- **Intraday Timeline**: A sequence of timestamped intra-session price readings recorded during an active market session.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of single-date queries for today's active session return contract-compliant snapshots with intraday timelines without requiring callers to implement date-checking logic.
- **SC-002**: Multi-day date range requests spanning historical dates through today complete within the standard 200ms latency ceiling for in-memory and cached responses.
- **SC-003**: 0% regression across all 36 verification scenarios and 100% parity across all 30 differential comparison endpoints.
- **SC-004**: Eliminates 100% of duplicated date-checking and intraday-stitching boilerplate across caller query modules, concentrating date-boundary synthesis into one deep module.

---

## Assumptions

- Operating timezone for Egyptian financial markets is permanently anchored to `Africa/Cairo`.
- Concluded past weekly buckets and daily archives are immutable and do not change after the calendar week is sealed.
- Callers requesting today's active session expect intra-session timeline points when available, without needing separate query parameters.
- Public client API route URLs (`/v1/rates/history/:date`, `/v1/gold/history/:date`, `/v1/stocks/history/:date`, etc.) remain identical; only internal architectural depth and responsibility boundaries change.

