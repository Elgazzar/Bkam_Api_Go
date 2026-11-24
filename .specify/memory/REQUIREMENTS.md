# Bkam API Gateway - Living Requirements

This living requirements document tracks the functional requirements for the Bkam API Gateway service.

---

## Functional Requirements Catalog

### 1. Currency Rates (`/v1/rates`)
- **FR-API-001**: System MUST serve latest bank currency exchange rates with sub-200ms p95 latency.
- **FR-API-002**: System MUST support query parameter filters: `?bank=` (case-insensitive) and `?currency=` (case-insensitive).
- **FR-API-003**: System MUST reject unsupported currency codes on `/v1/rates/:currency` with `400 Bad Request` and machine-readable error envelope.

### 2. Commodities (`/v1/gold`, `/v1/gas`)
- **FR-API-004**: System MUST serve latest Egyptian gold prices (`24k`, `21k`, `18k`, `pound`) and international spot gold in EGP and USD.
- **FR-API-005**: System MUST serve domestic fuel rates (Benzene 95, 92, 80, Solar, Kerosene, LPG) cached in RAM.

### 3. Equities and Indices (`/v1/stocks`)
- **FR-API-006**: System MUST serve real-time quotes across supported markets: Egypt (EGX), Saudi Arabia (Tadawul), UAE (ADX/DFM), Qatar (QSE), and Kuwait (Boursa Kuwait).
- **FR-API-007**: System MUST support dynamic multi-field sorting (`?sortBy=price|change|volume|name|index`) and benchmark filtering.
- **FR-API-008**: System MUST serve historical price series ranges using ISO weekly bucket queries (`rates_history_weekly/{YYYY-Www}`) stitched with today's live memory quote.

### 4. Zero-Read Caching & Performance
- **FR-API-009**: System MUST serve active session intraday timeline endpoints (`/v1/*/intraday/latest`) strictly from volatile memory without querying Firestore.
- **FR-API-010**: System MUST evaluate HTTP ETags and return `304 Not Modified` when requested resource representations match `If-None-Match`.

### 5. Unified Batch Quotes (`/v1/quotes`)
- **FR-API-011**: System MUST support unified batch quotes across all four asset classes in a single request via `POST /v1/quotes` or `GET /v1/quotes?items=...`.

### 6. Decoupled Read Gateway Architecture
- **FR-API-012**: System MUST initialize as a standalone read gateway without instantiating direct web scrapers or browser contexts, relying 100% on pre-populated Firestore snapshots and push listeners.

### 7. Unified Market History Resolution (`/v1/*/history/*`)
- **FR-API-013**: System MUST resolve single-date and multi-day historical market snapshots across FX rates, gold commodities, equities, and benchmark indices through a centralized deep engine (`HistoryResolutionEngine`), transparently absorbing Africa/Cairo operational timezone calculation, active today live synthesis (RAM + intraday timeline), multi-day range live stitching, standardized cache lifetime metadata (`max-age=60` for active today vs `max-age=3600` for sealed past), and centralized future date validation rejection (400 Bad Request).

### 8. Centralized Deep Intraday Coordination (`/v1/*/intraday/latest`)
- **FR-API-014**: System MUST resolve latest intraday price timelines and quote metrics across all 4 platforms (FX rates, gold karats, stock equities, benchmark indices) through a centralized deep coordinator (`IntradayCoordinator.getLatestIntraday`), transparently absorbing in-memory volatile snapshot extraction (0 database reads during active sessions), quote extraction, previous closing price baselining (Feature 100), off-hours session backtracking (capped at 7 days), ticker alias mapping, consolidated all-item payloads (`item=all`), and standardized cache headers (`max-age=60` for active today vs `max-age=3600` for backtracked past sessions).


