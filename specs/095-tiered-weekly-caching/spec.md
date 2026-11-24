# Feature Specification: Tiered Weekly Caching & API Performance Optimization

**Feature Branch**: `095-tiered-weekly-caching`

**Created**: 2026-10-24

**Status**: Draft

**Input**: User description: "implement the grill-with-docs session document here 0005-tiered-weekly-caching-and-api-optimization.md"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Instant Historical Stock Chart Loading on Mobile Networks (Priority: P1)

A mobile app user opens an individual stock page (e.g., Commercial International Bank - COMI) and views 1-month, 3-month, or 6-month historical price charts. The chart data downloads in a fraction of a second over cellular networks, rendering smoothly without spinning loaders.

**Why this priority**: Fast chart loading directly impacts user engagement and app responsiveness. Multi-month queries are the heaviest endpoints on the API and must load instantly.

**Independent Test**: Request `GET /v1/stocks/COMI/history/range?start=2026-05-01&end=2026-10-20` over a simulated 3G/4G cellular connection. The response is transparently compressed, under 35KB, and returns in under 300ms.

**Acceptance Scenarios**:

1. **Given** a requested date range where the end date is in the past (`end < todayCairo`), **When** the client queries `GET /v1/stocks/:symbol/history/range`, **Then** the response is served from the persistent range result cache in under 20ms, returns `Cache-Control: public, max-age=86400, immutable`, and contains only closing price records without intraday arrays.
2. **Given** a client that sends the `Accept-Encoding: gzip` or `br` header, **When** any API response payload exceeds 1KB, **Then** the server compresses the payload using the requested encoding, achieving an 80%+ payload size reduction compared to raw JSON.

---

### User Story 2 - Real-Time Today Stitching with Zero-Stale Quotes (Priority: P2)

An equity trader requests historical charts ending on the active trading day (`end >= todayCairo`) during market hours. The trader sees all historical days resolved from cache, seamlessly stitched together with the live market price and real-time intraday tick points for today.

**Why this priority**: Financial users require sub-second chart rendering while demanding guaranteed 100% price accuracy for the current trading day.

**Independent Test**: Query `GET /v1/stocks/COMI/history/range?start=2026-05-01&end=todayCairo`. The historical days match past daily closes, while today's record contains the latest live price and intraday points, with `Cache-Control: public, max-age=60`.

**Acceptance Scenarios**:

1. **Given** a range query ending on the active calendar date, **When** the server processes the request, **Then** it retrieves the sealed historical prefix (`[start, yesterdayCairo]`) from persistent cache, extracts today's live price and intraday points from memory, merges them into a single sorted timeline, and serves the response in under 50ms.
2. **Given** consecutive queries for the same range ending today within 60 seconds, **When** no new market price change has occurred, **Then** the client receives the cached response with zero database reads.

---

### User Story 3 - Off-Hours Server Reboot with Zero Database Reads (Priority: P3)

The API service restarts in the evening or over the weekend (Friday/Saturday) for maintenance or deployment. When incoming requests arrive for the active calendar week, the server immediately serves the weekly consolidated bucket from portable disk storage without making database reads.

**Why this priority**: Preserves external database read quotas and ensures fast cold-start availability during off-hours, when ~85% of weekly clock time occurs.

**Independent Test**: Restart the API process on a weekend or after market close (e.g. 20:00 Cairo). Query `GET /v1/stocks/:symbol/history/range`. The weekly consolidated bucket is read from disk with zero external database queries.

**Acceptance Scenarios**:

1. **Given** the current time is outside operating trading hours (e.g. after 16:00 Cairo on Sunday–Thursday, or any time on Friday/Saturday), **When** a weekly bucket document for the current week is fetched, **Then** it is read from disk cache if present; if fetched from the database, it is saved to disk cache with an expiration set to 09:30 AM Cairo of the next trading day.
2. **Given** the current time is during active trading hours (Sunday–Thursday 09:30–16:00 Cairo), **When** the active week document is fetched, **Then** disk caching is strictly bypassed and data is held in memory with a 5-second deduplication TTL to prevent stale disk writes.

---

### User Story 4 - Instant Chart Tab Switching via Conditional Revalidation (Priority: P4)

A user navigates between different stock tabs in a mobile app, periodically returning to a previously loaded chart. The client checks if the chart has updated using standard HTTP conditional revalidation, receiving an immediate confirmation without re-downloading the chart data.

**Why this priority**: Minimizes cellular bandwidth and battery consumption on mobile devices while delivering instant 0ms visual transitions for previously viewed screens.

**Independent Test**: Request `GET /v1/stocks/COMI/history/range`, extract the returned `ETag` header, and repeat the request with `If-None-Match: <etag>`. Verify the response is HTTP `304 Not Modified` with an empty response body.

**Acceptance Scenarios**:

1. **Given** an API response containing an `ETag` header, **When** a client repeats the request with `If-None-Match` matching the current payload hash, **Then** the server responds with status `304 Not Modified` and an empty payload body in under 10ms.
2. **Given** a stock whose price changes due to a new trading session, **When** the client sends `If-None-Match` with the old ETag, **Then** the server detects the hash mismatch and returns status `200 OK` with the updated JSON payload and a new ETag.

---

### Edge Cases

- **Single-Day Range for Today**: When `start === end === todayCairo`, the historical prefix is empty. The server serves today's live rate and in-memory intraday points directly from memory without writing to the persistent historical range cache.
- **Weekend / Holiday Midnight Rollover**: When today is a non-trading day (Friday/Saturday or public holiday), today's intraday is empty; the coordinator falls back to the previous trading session from memory, and the range query preserves the previous session's closing rates.
- **Client Without Compression Support**: If a client does not send an `Accept-Encoding` header (or requests `identity`), the server responds with standard uncompressed JSON without error.
- **Cache Corruption / Disk Read Failure**: If a cached entry in disk storage is malformed or unreadable, the provider safely catches the error, logs a diagnostic warning, and falls back to database retrieval without throwing a 500 error.
- **Intraday Points Segregation**: Intraday tick points are strictly excluded from persistent disk caches and weekly bucket documents, existing only in ephemeral memory.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001 (Session-Aware Active Week Disk Caching)**: The system MUST gate persistent disk caching for the active calendar week (`currentWeekKey`). During market trading hours (Sunday–Thursday 09:30–16:00 Cairo), disk writes MUST be bypassed and the active week MUST be cached only in memory with a 5-second TTL. Outside trading hours (all day Friday/Saturday and Sunday–Thursday 16:00–09:30 Cairo), the active week MUST be written to persistent disk cache with an expiration timestamp synchronized to 09:30 AM Cairo of the next trading day.
- **FR-002 (Immutable Sealed Week Disk Caching)**: The system MUST permanently cache past, sealed weekly bucket documents (`weekKey !== currentWeekKey`) in persistent disk cache (`flat-cache`) with a 30-day TTL, resolving subsequent lookups across process restarts with zero database reads.
- **FR-003 (Range Result Cache for Historical Queries)**: The system MUST maintain a symbol-scoped historical range result cache in persistent disk storage. For ranges ending in the past (`end < todayCairo`), the complete symbol time-series array MUST be cached for 30 days. Repeat requests for the same symbol and range MUST resolve in under 20ms with zero document iteration.
- **FR-004 (Dynamic Historical Prefix Stitching)**: For range queries ending on or after the active calendar day (`end >= todayCairo`), the system MUST cache only the closed historical prefix (`[start, yesterdayCairo]`) in disk storage. The system MUST dynamically extract today's live quote and real-time intraday tick points from memory at request time and stitch them onto the historical prefix.
- **FR-005 (Strict Intraday RAM-Only Lifecycle)**: Intraday 5-minute tick points MUST exist exclusively in memory for the active trading session and MUST NEVER be serialized into persistent disk caches or consolidated weekly documents. Intraday memory containers MUST be cleared upon daily session rollover.
- **FR-006 (Payload Compression)**: The API gateway MUST enable HTTP response compression for all responses exceeding 1KB, supporting standard Gzip, Deflate, and Brotli encodings, reducing JSON payload transfer size by at least 80% for supporting clients.
- **FR-007 (Conditional HTTP Requests & ETags)**: The API gateway MUST generate weak entity tags (`ETag`) for all JSON responses and evaluate `If-None-Match` request headers. If the payload has not changed, the server MUST return HTTP status `304 Not Modified` with an empty body.
- **FR-008 (Dynamic Cache-Control Headers)**: The range history endpoint MUST apply dynamic cache headers:
  - Historical ranges ending before today (`end < todayCairo`): `Cache-Control: public, max-age=86400, immutable`.
  - Ranges ending on or after today (`end >= todayCairo`): `Cache-Control: public, max-age=60`.

---

### Key Entities

- **Weekly Bucket**: A consolidated document storing daily closing quotes for an ISO 8601 calendar week (`YYYY-Www`), containing daily closing snapshots for all trading sessions in that week.
- **Active Week**: The current ISO calendar week whose trading sessions are ongoing and mutable.
- **Sealed Week**: A concluded past ISO calendar week whose trading sessions are permanently complete and immutable.
- **Range Result Cache**: A pre-computed, symbol-specific historical quote time-series stored in persistent disk storage to bypass multi-document iteration across weekly buckets.
- **Intraday Timeline**: High-frequency intra-session quotes (5-minute tick points) maintained strictly in memory for the active trading session and deleted upon session rollover.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001 (Payload Size Reduction)**: 6-month historical stock range response payloads over cellular networks are reduced by at least 80% (from ~300KB uncompressed down to <50KB compressed).
- **SC-002 (Cached Query Latency)**: Warm historical range queries for previously fetched stock ranges resolve in under 20ms (a 90%+ latency reduction compared to un-cached multi-document iteration).
- **SC-003 (Off-Hours Zero Database Reads)**: 100% of historical weekly bucket reads for the active week during weekend or evening server restarts hit the local disk cache with zero database reads.
- **SC-004 (Bandwidth Elimination on Navigation)**: Repeated chart views with `If-None-Match` return HTTP `304 Not Modified` with 0 payload bytes transferred and under 15ms roundtrip time.
- **SC-005 (Zero Stale Quotes During Market Hours)**: Exactly 0% of range queries during active trading sessions serve stale daily quotes for the current trading day.

---

## Assumptions

- **Target Operating Schedule**: Egyptian Stock Exchange (EGX) trading hours are Sunday through Thursday, 10:00 AM to 2:30 PM Cairo time (with pre-market / auction windows spanning 09:30 to 15:30 Cairo). Friday and Saturday are full weekend market closures.
- **Client Transport Compatibility**: Mobile clients (Android OkHttp, iOS URLSession, web browsers) support standard HTTP Gzip compression and conditional `ETag` validation without requiring custom application code changes.
- **Storage Boundary**: Local persistent disk storage is available via Node.js filesystem (`.cache/` directory) and survives standard process reboots on the host environment.
- **Backward Compatibility**: Existing response JSON structures for `GET /v1/stocks/:symbol/history/range` remain unchanged, ensuring 100% backward compatibility for existing client applications.
