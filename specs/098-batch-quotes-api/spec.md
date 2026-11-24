# Feature Specification: Stateless Batch Quotes API Endpoint for Home Favorites

**Feature Branch**: `098-batch-quotes-api`

**Created**: 2026-10-31

**Status**: Ready for Planning

**Input**: User description: "Implement a Stateless Batch Quotes API Endpoint for Home Favorites. We want to introduce a unified, stateless batch endpoint (BFF pattern) to fetch prices for a user's locally stored favorites in a single HTTP request instead of having the mobile app make 4–5 separate network calls (/v1/rates, /v1/gold, /v1/indices, /v1/stocks)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fetch All Home Favorites in a Single Batch Request (Priority: P1)

As a mobile app user opening the app, I want my saved favorites (spanning bank exchange rates, gold prices, market indices, and individual stocks) to load together in a single fast network request so that my home dashboard displays all updated prices instantly without waiting for multiple round-trips.

**Why this priority**: Core value of the feature. Replaces 4–5 parallel HTTP requests with a single batched call, saving mobile bandwidth, reducing network latency, and preventing partial UI loading states.

**Independent Test**: Can be tested independently by sending a batch request containing a mixed list of identifiers (`["CBE_USD", "21k", "EGX30", "COMI"]`) and verifying that valid quotes for all 4 distinct financial domains are returned in a single response payload keyed by identifier.

**Acceptance Scenarios**:

1. **Given** valid identifiers for bank currency (`CBE_USD`), gold (`21k`), index (`EGX30`), and stock (`COMI`), **When** the client submits a batch quotes request with these items via POST or GET, **Then** the system returns latest price data, currency/units, daily change, and update timestamps for each requested item inside the `quotes` map keyed by ID.
2. **Given** a client request with a single item category (e.g., only gold karats or only bank currencies), **When** the client submits the batch request, **Then** the system returns accurate quote data specifically for those items in the `quotes` dictionary without requiring items from other categories.

---

### User Story 2 - Resilient Partial Batch Resolution for Unknown or Stale Identifiers (Priority: P2)

As a mobile client application with potentially outdated locally stored favorite IDs, I want the batch quotes endpoint to gracefully resolve known items even if some requested identifiers are invalid, unlisted, or temporarily unavailable, so that the user's valid favorites still render seamlessly and missing items are explicitly reported.

**Why this priority**: Mobile favorites are stored locally on device disks and may survive catalog changes or typos. A single bad identifier must never cause the entire batch request to fail.

**Independent Test**: Can be tested by sending a payload with valid items alongside nonexistent items (`["CBE_USD", "INVALID_XYZ"]`) and verifying that `CBE_USD` returns valid data in `quotes` while `INVALID_XYZ` is reported in the `unresolved` list without a 500 server error.

**Acceptance Scenarios**:

1. **Given** a request containing valid items and an unrecognized identifier, **When** the client submits the request, **Then** valid items are returned in the `quotes` dictionary and the invalid identifier is returned in the `unresolved` array.
2. **Given** a batch request where none of the requested identifiers are recognized, **When** the client submits the request, **Then** the system responds with a 200 OK containing an empty `quotes` dictionary (`{}`) and all requested IDs listed in the `unresolved` array.

---

### User Story 3 - Client Payload Validation and Batch Sizing Safeguards (Priority: P3)

As a system operator, I want the batch quotes endpoint to validate incoming requests and enforce a maximum batch size limit, so that client applications cannot overwhelm the backend memory or compute layers with excessive identifiers.

**Why this priority**: Protects backend resources, ensures predictable sub-millisecond response times, and prevents denial-of-service or memory pressure.

**Independent Test**: Can be tested by sending empty requests, malformed payloads, or requests exceeding the batch size limit, and verifying appropriate validation error responses.

**Acceptance Scenarios**:

1. **Given** an empty `items` array or missing `items` field, **When** the client submits the request, **Then** the system returns a descriptive 400 Bad Request error envelope.
2. **Given** a request containing more items than the permitted maximum batch threshold (e.g., > 100 items), **When** the client submits the request, **Then** the system rejects the request with a 400 Bad Request indicating the limit.

---

### Edge Cases

- **Duplicate Identifiers**: When a request contains duplicate IDs (e.g. `["21k", "21k", "CBE_USD"]`), the system deduplicates them and returns each quote once in the `quotes` dictionary.
- **Mixed Domain Collisions**: Identifiers are resolved deterministically across domains:
  - Format `<BANK>_<CURRENCY>` matches bank rates (e.g., `CBE_USD`, `MISR_EUR`).
  - Standard gold keys match gold commodities (`24k`, `21k`, `18k`, `pound`, `ounce`).
  - Market index symbols match indices (`EGX30`, `EGX70`, `EGX100`).
  - Remaining valid symbols match equities/stocks (e.g., `COMI`, `1120`, `EMAAR`).
- **Offline / Degraded Provider**: When one underlying data asset has temporary stale data, the system returns the latest cached quote data with freshness timestamps rather than failing the entire batch.
- **Query Parameter String Parsing**: When using `GET /v1/quotes?items=CBE_USD,21k,EGX30`, whitespace is trimmed and empty entries are ignored.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a stateless endpoint to accept a list of requested asset identifiers and return their latest market quotes in a single response.
- **FR-002**: System MUST NOT require user authentication or persist client favorites on the backend.
- **FR-003**: System MUST support cross-domain identifier resolution for:
  - Bank exchange rates formatted as `<BANK>_<CURRENCY>` (e.g., `CBE_USD`, `MISR_EUR`, `NBE_SAR`)
  - Gold commodities and karats (e.g., `24k`, `21k`, `18k`, `pound`, `ounce`)
  - Stock market indices (e.g., `EGX30`, `EGX70`, `EGX100`)
  - Equities and stocks across Egyptian and GCC markets (e.g., `COMI`, `FWRY`, `1120`, `2222`)
- **FR-004**: System MUST resolve all requested quotes using existing in-memory/push-cached state with zero external scraping or redundant database reads during the request lifecycle.
- **FR-005**: System MUST enforce a maximum batch size limit (up to 100 items per request) and reject requests exceeding this threshold with a 400 Bad Request.
- **FR-006**: System MUST return quotes structured as a dictionary / map keyed by the requested item identifier under `quotes` (e.g., `{"quotes": {"CBE_USD": {...}, "21k": {...}}, "unresolved": [...]}`), enabling O(1) key lookups for client applications.
- **FR-007**: System MUST support both `POST /v1/quotes` (with JSON body `{"items": [...]}`) and `GET /v1/quotes?items=CBE_USD,21k,EGX30` (comma-separated query string) to support flexible integration and standard HTTP caching.
- **FR-008**: System MUST report any unrecognized, invalid, or unlisted identifiers in a separate `unresolved` string array in the response (e.g., `"unresolved": ["INVALID_ID"]`), while successfully returning all valid items in `quotes`.
- **FR-009**: System MUST return normalized quote metadata for each resolved item, including item identifier, asset category (`currency`, `gold`, `index`, `stock`), display name, current price (buy/sell or last/close), change amount, change percentage, and timestamp of last update.

### Key Entities *(include if feature involves data)*

- **BatchQuoteRequest**: Represents the client request payload containing an array of item identifier strings (via JSON body for POST or comma-separated list for GET).
- **BatchQuoteResponse**: Represents the unified response envelope delivering:
  - `quotes`: Key-value map of resolved `QuoteItem` objects keyed by item identifier.
  - `unresolved`: Array of identifier strings that could not be resolved or found.
  - `timestamp`: Server response timestamp (ISO 8601).
- **QuoteItem**: Normalized data representation of a single resolved asset quote, containing:
  - `id`: The requested identifier (e.g., `CBE_USD`, `21k`, `EGX30`, `COMI`).
  - `category`: Asset domain (`currency` | `gold` | `index` | `stock`).
  - `name`: Human-readable display name (e.g., "US Dollar - Central Bank of Egypt", "21K Gold").
  - `price`: Primary numeric price or quote value.
  - `buyPrice`: Optional buy rate (for currencies and gold).
  - `sellPrice`: Optional sell rate (for currencies and gold).
  - `change`: Absolute price change value.
  - `changePercent`: Percentage price change value.
  - `currency`: Currency unit (e.g., `EGP`, `SAR`, `AED`, `QAR`, `KWD`).
  - `updatedAt`: ISO 8601 timestamp of latest price update.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Mobile clients can retrieve price updates for all favorite assets in a single network round-trip instead of 4–5 separate requests.
- **SC-002**: Batch quote requests for up to 50 items are processed and served from memory in under 50 milliseconds (excluding client network transit).
- **SC-003**: 100% of valid identifiers across the 4 core domains (bank rates, gold, indices, stocks) resolve correctly in a single combined payload.
- **SC-004**: System executes zero new Firestore document reads when serving batch quotes, leveraging existing in-memory cache layers.
- **SC-005**: 0% total request failure rate when client requests include invalid or deprecated identifiers alongside valid items (unresolved items isolated to `unresolved` list).

## Assumptions

- Clients will store user favorites locally on device storage (e.g. SQLite, Room, SharedPreferences) and provide their active list with each request.
- Asset identifiers follow existing naming conventions already used by client and backend services (e.g., bank codes like `CBE`, gold karats like `21k`, stock symbols like `COMI`).
- The batch quotes endpoint operates strictly within existing cache lifecycles without triggering ad-hoc scraping jobs.
