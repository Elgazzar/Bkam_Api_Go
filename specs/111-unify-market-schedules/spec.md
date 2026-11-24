# Feature Specification: Unify Fragmented Market Schedules into Deep MarketSchedule Engine

**Feature Branch**: `111-unify-market-schedules`  
**Created**: 2026-11-15  
**Status**: Draft  
**Input**: User description: "Unify Fragmented Market Schedules into Deep MarketSchedule Engine\nProblem\nTwo disjoint schedule modules declare conflicting EGX trading hours, while intraday services scan Firestore across a 7-day loop to find previous session dates.\n\nSolution\nConsolidate all domestic and regional markets into a deep MarketScheduleEngine that provides immediate session dates without storage scanning."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Authoritative Multi-Market Trading Status & Cache TTL (Priority: P1) 🎯 MVP

As an API client or caching middleware querying market quotes across domestic (Egyptian Exchange) and regional financial markets (Saudi Tadawul, Dubai Financial Market / Abu Dhabi Securities Exchange, Qatar Stock Exchange, Boursa Kuwait), I want immediate and authoritative market session status (`open` vs `closed`) and precise remaining durations until the next opening bell, so that client applications display accurate market badges and server caches compute exact off-hours time-to-live (TTL) expiration timestamps without conflicting schedules.

**Why this priority**: Conflicting schedule declarations across disjoint modules lead to inconsistent cache invalidation, premature disk cache expirations, and mismatched market open indicators across API consumers. Unifying all market definitions into a single authoritative engine eliminates operational ambiguity.

**Independent Test**: Can be tested by evaluating market status and seconds until next open for any supported market identifier (e.g. `EGX`, `SA`, `AE`, `QA`, `KW`) across active trading sessions, after-hours evenings, weekends, and holidays, verifying deterministic and unified boolean status and non-negative TTL outputs.

**Acceptance Scenarios**:

1. **Given** the current clock time is within the official trading hours of the Egyptian Stock Exchange (e.g., Sunday 11:00 AM Cairo time), **When** the schedule engine is queried for `EGX`, **Then** the engine reports `isMarketOpen = true` and `secondsUntilNextOpen = 0`.
2. **Given** the current clock time is outside trading hours (e.g., Thursday 6:00 PM Cairo time), **When** the schedule engine is queried for `EGX`, **Then** the engine reports `isMarketOpen = false` and returns the exact number of seconds until Sunday 09:30 AM Cairo time.
3. **Given** regional markets with differing operating calendars (e.g., UAE operating Monday–Friday vs. Egypt/Saudi/Qatar/Kuwait operating Sunday–Thursday), **When** schedule status is evaluated on a Sunday, **Then** the engine correctly identifies Saudi, Egypt, Qatar, and Kuwait as active trading days while reporting UAE as closed for weekend.

---

### User Story 2 - Instant Session Backtracking Without Storage Scanning (Priority: P2)

As an intraday coordinator or financial time-series service resolving the latest trading session timeline during off-hours, weekends, or midnight rollovers, I want the schedule engine to immediately determine the exact calendar date (`YYYY-MM-DD`) of the current or most recent trading session, so that the service can retrieve the target session document in a single point lookup without looping or scanning cloud database collections across a 7-day window.

**Why this priority**: Eliminates redundant cloud database query loops (which previously scanned up to 7 documents on cold-cache or off-hours requests), protecting cloud database read budgets, reducing latency from >500ms to <1ms, and preventing lookback timeouts.

**Independent Test**: Can be tested by invoking session date resolution for various markets during midnight rollovers (e.g. Sunday 01:00 AM before market open), weekends (e.g. Saturday afternoon), and multi-day holiday closures, verifying that it instantly returns the exact calendar date of the preceding trading day with zero cloud database queries.

**Acceptance Scenarios**:

1. **Given** a query executed during an active trading session (e.g., Tuesday 12:00 PM in Cairo), **When** requesting the latest trading session date for `EGX`, **Then** the engine returns today's calendar date (`YYYY-MM-DD`).
2. **Given** a query executed on a trading day but prior to market open (e.g., Tuesday 07:00 AM in Cairo), **When** requesting the latest trading session date for `EGX`, **Then** the engine returns yesterday's calendar date (Monday's date).
3. **Given** a query executed over the weekend (e.g., Friday or Saturday in Cairo), **When** requesting the latest trading session date for `EGX`, **Then** the engine returns the preceding Thursday's calendar date.
4. **Given** a query executed over the UAE weekend (e.g., Sunday in Dubai), **When** requesting the latest trading session date for `AE`, **Then** the engine returns the preceding Friday's calendar date.

---

### User Story 3 - Comprehensive Holiday Calendar & Special Session Support (Priority: P3)

As an operations administrator and automated scheduling daemon, I want the market schedule engine to recognize official national holidays, public market closures, and Ramadan shortened trading sessions, so that automated archival tasks, weekly bucket consolidation, and cache warming jobs do not misclassify official holidays as unexpected data collection failures or negative cache anomalies.

**Why this priority**: Financial exchanges frequently observe scheduled secular and religious holidays (e.g., Eid holidays, National Days, New Year). Without an integrated holiday calendar, services assume trading sessions occurred and trigger false-positive alerts or attempt redundant scrapes.

**Independent Test**: Can be tested by evaluating schedule state on configured national holiday dates for Egypt and regional GCC countries, verifying that the engine classifies the date as non-trading and backtracks to the last open business day.

**Acceptance Scenarios**:

1. **Given** a target calendar date recognized as a national market holiday in Egypt, **When** evaluating market open status, **Then** the engine reports `isMarketOpen = false` and calculates the time until the next regular business day opening session.
2. **Given** an intraday query on an official holiday, **When** resolving the latest session date, **Then** the engine skips the holiday and returns the preceding open session date.

---

### Edge Cases

- **Midnight Rollover Boundary**: At 00:00:00 local time on a scheduled trading day before the opening bell, the latest session MUST resolve to the prior business day rather than the un-started current day.
- **Mid-Session Evaluation**: Right at the exact opening second (e.g., 09:30:00.000) or closing second (e.g., 15:30:00.000), boundary inequalities must deterministically classify the state without jitter.
- **Timezone Transition & Daylight Saving**: The engine MUST evaluate dates and times strictly within each market's respective native IANA timezone identifier (e.g., `Africa/Cairo`, `Asia/Riyadh`, `Asia/Dubai`, `Asia/Qatar`, `Asia/Kuwait`), handling seasonal UTC offset changes (such as Egypt's summer time shifts between UTC+2 and UTC+3) automatically via standard timezone formatters.
- **Unrecognized Market Identifier**: When an unrecognized or null market identifier is passed, the engine MUST safely fall back to the primary domestic market (`EGX`) with a warning log rather than throwing an unhandled exception.
- **Commodity & Foreign Exchange Assets**: For non-equity assets that trade 24/7 or follow international banking schedules (e.g., Gold, Currency Exchange), the engine MUST provide standardized scheduling profiles that reflect banking days (Sunday–Thursday in Egypt) while accommodating continuous spot pricing.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a single, authoritative, centralized schedule coordinator (`MarketScheduleEngine`) that consolidates all market operating hours, calendar definitions, and time-zone rules across domestic and regional exchanges.
- **FR-002**: `MarketScheduleEngine` MUST expose a unified 3-method public contract interface:
  - `isMarketOpen(market?: string, referenceDate?: Date): boolean`
  - `getSecondsUntilNextOpen(market?: string, referenceDate?: Date): number`
  - `getLatestTradingSessionDate(market?: string, referenceDate?: Date): string`
- **FR-003**: System MUST resolve conflicting EGX schedule declarations by establishing a single authoritative schedule: active trading and pre/post-trading buffer window spanning 09:00 to 15:30 Cairo local time on Sunday through Thursday.
- **FR-004**: System MUST encapsulate multi-market configurations for all supported regional exchanges:
  - **Egypt (`EGX`, `EGX_INDICES`)**: `Africa/Cairo` timezone, active days Sunday–Thursday, operating window 09:00–15:30.
  - **Saudi Arabia (`SA`, `SA_INDICES`)**: `Asia/Riyadh` timezone, active days Sunday–Thursday, operating window 09:00–16:00.
  - **United Arab Emirates (`AE`, `AE_INDICES`)**: `Asia/Dubai` timezone, active days Monday–Friday, operating window 09:00–16:00.
  - **Qatar (`QA`, `QA_INDICES`)**: `Asia/Qatar` timezone, active days Sunday–Thursday, operating window 08:30–14:15.
  - **Kuwait (`KW`, `KW_INDICES`)**: `Asia/Kuwait` timezone, active days Sunday–Thursday, operating window 08:00–13:30.
- **FR-005**: `MarketScheduleEngine.getLatestTradingSessionDate` MUST calculate the latest active or completed trading session date in zero (0) cloud database reads, replacing all 7-day Firestore document scanning loops.
- **FR-006**: `IntradayCoordinator` MUST delegate session backtracking directly to `MarketScheduleEngine.getLatestTradingSessionDate`, completely deleting the iterative Firestore `docRefs` search and descending collection query loops.
- **FR-007**: `FirestoreListenerService` and `HistoryResolutionEngine` MUST consume schedule states and off-hours TTL calculations exclusively through `MarketScheduleEngine`, deprecating duplicate local calculation helpers.
- **FR-008**: The schedule engine MUST support an extensible holiday calendar registry capable of evaluating multi-day official market closures per country.
- **FR-009**: All existing downstream REST API endpoint contracts and payload formats MUST remain 100% backward-compatible with zero response drift.

---

### Key Entities

- **MarketScheduleState**: Comprehensive operational state envelope for a specific financial exchange, capturing current local date, day of week, decimal hour, active trading status, next opening epoch timestamp, and remaining seconds.
- **MarketOperatingRules**: Static configuration record defining an exchange's canonical identifier, IANA timezone string, active operating days of the week, daily opening time, daily closing time, and official holidays.
- **TradingSessionWindow**: Temporal boundaries representing the start and end timestamps of a specific market's trading day in UTC and local time.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% elimination of cloud database read operations during off-hours intraday session backtracking (0 Firestore reads vs. up to 7 reads previously).
- **SC-002**: Intraday session resolution latency during off-hours reduced from >300ms to <2ms in memory.
- **SC-003**: 100% consistency in EGX market open/closed status across all consumers, with zero conflicting schedule definitions in the repository.
- **SC-004**: All unit tests, contract tests, and differential parity checks pass with 100% success rate across all 30 public API endpoints.

---

## Assumptions

- **Market Calendars**: The trading week definitions (Sunday–Thursday for Egypt, Saudi Arabia, Qatar, and Kuwait; Monday–Friday for the UAE) follow standard national banking and exchange regulations in the respective countries.
- **Timezone Offsets**: System runtime evaluates standard and daylight saving time offsets using official IANA timezone databases via standard JavaScript/Node `Intl.DateTimeFormat` APIs.
- **Scraper Ingestion Buffer**: The configured schedule boundaries include standard pre-market auction and delayed feed arrival margins (e.g. 15-minute delayed market feeds) to ensure caches remain fresh through closing trade settlement.
