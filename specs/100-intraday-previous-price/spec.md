# Feature Specification: Expose Previous Day Price in Latest Intraday Endpoints

**Feature Branch**: `100-intraday-previous-price`  
**Created**: 2026-11-05  
**Status**: Draft  
**Input**: User description: "Update all GET /v1/*/intraday/latest endpoints (Stocks, Indices, Gold, and Bank Rates) to include top-level previous day price fields (previousClose for stocks/indices, and previousBuyPrice / previousSellPrice for gold and bank rates) sourced from existing in-memory caches with 0 new database reads, and update tests and API documentation."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Equities & Benchmark Indices Intraday Chart Baselines (Priority: P1)

As an Android mobile app developer, I want `GET /v1/stocks/:symbol/intraday/latest` and `GET /v1/indices/:symbol/intraday/latest` to return a top-level `previousClose` attribute alongside the current price and intraday points, so that the mobile app can render a horizontal baseline reference line and calculate price movement percentages in a single HTTP request without secondary historical queries.

**Why this priority**: Mobile charts (e.g. MPAndroidChart, Vico, Jetpack Compose Canvas) require a previous closing reference line to establish chart color (green for gains, red for losses) and baseline geometry. Eliminating the second round-trip halves latency and battery consumption on mobile devices.

**Independent Test**: Query `GET /v1/stocks/COMI/intraday/latest` and `GET /v1/indices/EGX30/intraday/latest`; verify the returned JSON envelope includes `previousClose` as a positive floating-point number.

**Acceptance Scenarios**:
1. **Given** an active or historical trading session for an Egyptian equity (`COMI`), **When** requesting `GET /v1/stocks/COMI/intraday/latest`, **Then** the JSON payload contains `previousClose: <number>` reflecting the closing price of the preceding trading session.
2. **Given** an active or historical session for a regional GCC stock (`1120` in Saudi market), **When** requesting `GET /v1/stocks/1120/intraday/latest?market=saudi`, **Then** the JSON payload contains `previousClose: <number>` denominated in SAR.
3. **Given** an active or historical session for a benchmark index (`EGX30`, `TASI`), **When** requesting `GET /v1/indices/:symbol/intraday/latest`, **Then** the payload contains `previousClose: <number>` representing the benchmark's previous close.

---

### User Story 2 - Bank Exchange Rates Intraday Chart Baselines (Priority: P1)

As an Android mobile app developer, I want `GET /v1/rates/:currency/intraday/latest` to include `previousBuyPrice` and `previousSellPrice` at the top level of the response, so that currency exchange rate charts can render previous closing buy and sell reference lines for individual banks or consolidated market averages.

**Why this priority**: Currency volatility requires comparing live bank rates against the previous banking session's close. Returning previous rates atomically avoids making secondary historical range queries.

**Independent Test**: Query `GET /v1/rates/USD/intraday/latest?bank=cbe` and `GET /v1/rates/USD/intraday/latest`; verify the response envelope contains `previousBuyPrice` and `previousSellPrice`.

**Acceptance Scenarios**:
1. **Given** a bank-filtered currency query (`/v1/rates/USD/intraday/latest?bank=cbe`), **When** the endpoint is invoked, **Then** the payload contains `previousBuyPrice` and `previousSellPrice` matching CBE's previous closing exchange rates.
2. **Given** a consolidated currency query without bank filter (`/v1/rates/USD/intraday/latest`), **When** the endpoint is invoked, **Then** the payload contains `previousBuyPrice` and `previousSellPrice` reflecting the consolidated market average previous close.

---

### User Story 3 - Gold Commodities Intraday Chart Baselines (Priority: P1)

As an Android mobile app developer, I want `GET /v1/gold/:item/intraday/latest` and `GET /v1/gold/intraday/latest` to return `previousBuyPrice` and `previousSellPrice`, so that domestic gold karat charts and international spot gold charts can render baseline trends and net daily changes.

**Why this priority**: Domestic gold (`21k`, `24k`, `18k`, `pound`) and international spot gold (`USD/oz`) fluctuate continuously. Single-call delivery of current, previous, and intraday tick points is critical for high-performance mobile chart rendering.

**Independent Test**: Query `GET /v1/gold/21/intraday/latest` and `GET /v1/gold/intraday/latest`; verify `previousBuyPrice` and `previousSellPrice` are present at the top level of single-item responses and within each karat object in the multi-karat response.

**Acceptance Scenarios**:
1. **Given** a single karat request (`GET /v1/gold/21/intraday/latest`), **When** invoked, **Then** the response includes `previousBuyPrice` and `previousSellPrice`.
2. **Given** an international spot gold request (`GET /v1/gold/international/intraday/latest`), **When** invoked, **Then** the response includes `previousBuyPrice` and `previousSellPrice` denominated in USD.
3. **Given** the multi-karat summary endpoint (`GET /v1/gold/intraday/latest`), **When** invoked, **Then** each nested item (`24`, `21`, `18`, `pound`, `international`) contains `previousBuyPrice` and `previousSellPrice`.

---

### User Story 4 - Zero-Cost In-Memory Resolution & Documentation (Priority: P2)

As a backend platform engineer, I want all previous day reference prices to be resolved strictly from existing volatile in-memory caches, memory-resident snapshot metadata, and catalog constants without executing new database read operations, backed by comprehensive API documentation and integration tests.

**Why this priority**: Protects the platform's Firestore Always Free Tier quota (Principle VII) while maintaining strict architectural consistency across all four asset classes.

**Independent Test**: Execute integration test suites verifying that endpoints return within <10ms and produce zero Firestore read operations during active market sessions.

**Acceptance Scenarios**:
1. **Given** live active trading sessions, **When** latest intraday endpoints are queried, **Then** previous day prices are resolved from memory in <10ms with 0 Firestore reads.
2. **Given** off-hours or weekend sessions, **When** backtracking to the preceding session, **Then** previous day prices are derived from buffered memory objects.
3. **Given** the updated API contracts, **When** consumers inspect `API_ENDPOINTS_GUIDE.md`, **Then** the new fields, sample JSON payloads, and TypeScript/Kotlin types are clearly documented.

---

### Edge Cases

- **Missing Previous Close on Newly Listed Securities**: If an equity or security was recently listed (e.g. IPO) or has no historical record, `previousClose` MUST gracefully default to the security's current quote or catalog reference value without returning null or throwing an error.
- **Illiquid Securities with Zero Intraday Trades**: When `intraday: []`, `previousClose` MUST still be returned accurately so that client charts can display an unchanged flat baseline.
- **Session Rollover During Off-Hours/Weekends**: When the intraday coordinator backtracks to a preceding trading session (e.g. Sunday query backtracking to Thursday session), the previous price MUST represent the session immediately preceding that resolved session.
- **Null / Unrecorded Banking Rates**: If a specific bank did not quote a currency on the previous day, `previousBuyPrice` and `previousSellPrice` should fall back to the bank's earliest recorded rate or consolidate safely without failing the request.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST include a top-level numeric `previousClose` field in the response envelope of `GET /v1/stocks/:symbol/intraday/latest` across all supported regional markets (EGX, Saudi, UAE, Qatar, Kuwait).
- **FR-002**: The system MUST include a top-level numeric `previousClose` field in the response envelope of `GET /v1/indices/:symbol/intraday/latest` across all supported benchmark indices.
- **FR-003**: The system MUST include top-level numeric `previousBuyPrice` and `previousSellPrice` fields in the response envelope of `GET /v1/rates/:currency/intraday/latest`.
- **FR-004**: When `?bank=` is provided in `GET /v1/rates/:currency/intraday/latest`, the previous rates MUST correspond to the specified bank; when omitted, they MUST correspond to the consolidated market average.
- **FR-005**: The system MUST include top-level numeric `previousBuyPrice` and `previousSellPrice` fields in `GET /v1/gold/:item/intraday/latest` for all single commodity items (`24`, `21`, `18`, `pound`, `international`).
- **FR-006**: The system MUST include `previousBuyPrice` and `previousSellPrice` within each nested karat object (`24`, `21`, `18`, `pound`, `international`) in `GET /v1/gold/intraday/latest` and `GET /v1/gold/all/intraday/latest`.
- **FR-007**: Sourcing of all previous day price fields MUST occur strictly from volatile RAM caches (`SnapshotCache`, `intradayCoordinator`, `stockCatalog`, `catalogEntry.previousClose`, `yesterdayCloseBuy`/`yesterdayCloseSell`), incurring **0 new database reads**.
- **FR-008**: The system MUST preserve backward compatibility with all existing fields (`price`, `buy`, `sell`, `change`, `changePercent`, `updatedAt`, `intraday`).
- **FR-009**: The system MUST update API contracts and documentation in `API_ENDPOINTS_GUIDE.md` reflecting the new fields across all four asset classes.

---

### Key Entities

- **Stock Intraday Response**: Contains `symbol`, `date`, `currency`, `price`, `previousClose`, `change`, `changePercent`, `updatedAt`, and `intraday` tick array.
- **Index Intraday Response**: Contains `symbol`, `date`, `currency`, `price`, `previousClose`, `change`, `changePercent`, `updatedAt`, and `intraday` tick array.
- **Bank Rate Intraday Response**: Contains `currency`, optional `bank`, `date`, `buy`, `sell`, `previousBuyPrice`, `previousSellPrice`, `updatedAt`, and `intraday` tick array.
- **Gold Intraday Response**: Contains `item` (or `karats` dictionary), `date`, `currency`, `buy`, `sell`, `previousBuyPrice`, `previousSellPrice`, `updatedAt`, and `intraday` tick array.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of successful responses to `GET /v1/*/intraday/latest` across Stocks, Indices, Gold, and Bank Rates contain the specified previous day price fields.
- **SC-002**: Mobile client applications can render an intraday price chart with an accurate baseline reference line and percentage change using exactly 1 HTTP network request.
- **SC-003**: 0 additional Firestore read operations are incurred during both active trading sessions and weekend/off-hours backtracking.
- **SC-004**: Endpoint response time remains under 10ms when serving cached active sessions from memory.
- **SC-005**: 100% test pass rate across integration tests validating the presence and mathematical validity of the new fields.

---

## Assumptions

- Previous day rates and closing prices are already recorded and accessible in the system's in-memory models (`yesterdayCloseBuy`/`yesterdayCloseSell` on `CurrencyRate`, `previousClose` on stock catalog items, and `yesterdayBuyPrice`/`yesterdaySellPrice` in gold snapshots).
- Client applications will use `previousClose` or `previousBuyPrice`/`previousSellPrice` to draw baseline limit lines and compute display percentages.
- The field names follow established platform naming conventions (`previousClose` for equities/indices, and `previousBuyPrice`/`previousSellPrice` for two-way commodity/currency quotes).

