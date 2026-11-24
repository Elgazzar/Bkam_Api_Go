# Research: Tiered Weekly Caching & API Performance Optimization

## Decision 1: Session-Aware Active Week Disk Caching

### Context
In [`firestore-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/firestore-service.ts#L87-L135), the active calendar week (`weekKey === currentWeekKey`) strictly bypasses `flat-cache` on disk to avoid persisting mutating quotes. However, the Egyptian Stock Exchange (EGX) trades for only ~25 hours out of 168 hours in a week (~15% of the time). During evenings, nights, and the 48-hour Friday/Saturday weekend, the current week's closing quotes are static. Bypassing disk completely forces Firestore queries on every process restart during off-hours.

### Decision
Implement **Session-Aware Gated Disk Caching**:
1. **Trading Window Detection**: In `Africa/Cairo` timezone:
   - Trading days are Sunday (0) through Thursday (4).
   - Friday (5) and Saturday (6) are weekend closures.
   - Operating trading window on trading days is 09:30 to 16:00 Cairo time (accounting for auction and post-session closing periods).
2. **Behavior During Trading Hours (Sun–Thu 09:30–16:00 Cairo)**:
   - The active week strictly bypasses disk reads and writes (`FlatCacheProvider`).
   - Requests use in-memory `weeklyBucketCache` with 5-second promise deduplication and real-time Firestore listeners.
3. **Behavior Outside Trading Hours (Evenings, Nights & Weekends)**:
   - When fetched from Firestore, the active week is saved to `flat-cache` with an expiration timestamp calculated as **09:30 AM Cairo of the next trading day**:
     - Sunday–Wednesday after 16:00: expires next day at 09:30 AM Cairo.
     - Thursday after 16:00: expires Sunday at 09:30 AM Cairo (covers the entire weekend).
     - Friday & Saturday: expires Sunday at 09:30 AM Cairo.
     - Sunday–Thursday before 09:30: expires today at 09:30 AM Cairo.
4. **Cache Invalidation**:
   - At 09:30 AM Cairo on trading days, the disk cache naturally expires. Subsequent queries during trading hours bypass disk, ensuring zero risk of stale data during market sessions.

### Alternatives Considered
- *Always bypass disk for active week (previous implementation)*: Wasted up to 50 Firestore reads across server reboots on weekends and evenings. Rejected.
- *Fixed 5-minute sliding TTL on disk*: Required disk write churn every 5 minutes and could still write stale snapshots during market sessions if a reboot occurred mid-session. Rejected in favor of calendar-aware session gating.

---

## Decision 2: Range Result Cache with Dynamic Today Stitching

### Context
In [`StockQueryService.fetchStockPriceHistoryRange`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/stock-query-service.ts#L350-L450), multi-month queries iterate over 180 individual calendar dates, extracting stock quote maps, checking intraday point dictionaries, and constructing result objects. Even with weekly documents cached, this CPU iteration and object allocation costs 15–30ms per request.

### Decision
Implement a secondary **Range Result Cache** in `FlatCacheProvider`:
1. **Pure Historical Ranges (`end < todayCairo`)**:
   - Key: `stock_range:${symbol}:${startDate}:${endDate}`
   - Result: Finalized, sorted array of `{ date, symbol, price, change, changePercent, updatedAt }`.
   - TTL: 30 days (`30 * 24 * 60 * 60` seconds).
   - Resolution: Immediate cache return in `<1ms` without weekly document iteration.
2. **Ranges Ending Today (`end >= todayCairo`)**:
   - Split range into:
     - Sealed historical prefix: `[startDate, yesterdayCairo]`
     - Dynamic today: `todayCairo`
   - Cache key for prefix: `stock_range_prefix:${symbol}:${startDate}:${yesterdayCairo}` (cached for 30 days).
   - At request time, fetch prefix in <1ms, retrieve today's quote from memory (`cachedPrices`), look up today's live intraday points from `intradayCoordinator`, append to prefix, and return.
   - Historical days have zero intraday arrays; only today's item contains live intraday points.

### Alternatives Considered
- *Re-iterating weekly documents on every request*: Wasted CPU and memory allocations on static historical dates. Rejected.
- *Caching today's full response on disk*: Violated requirement to keep mutating intraday points strictly in memory. Rejected.

---

## Decision 3: HTTP Response Compression via `@fastify/compress`

### Context
Multi-month stock history responses contain 90+ trading day objects, reaching 200KB to 500KB of raw JSON. On cellular connections in Egypt, downloading uncompressed payloads causes 1,500ms to 3,000ms latency.

### Decision
Install `@fastify/compress` (compatible with Fastify v5):
- Register in `api/src/app.ts` with `threshold: 1024` (1KB) and `encodings: ['gzip', 'deflate', 'br']`.
- Android OkHttp clients automatically send `Accept-Encoding: gzip` and transparently decompress response bodies in bytecode (<2ms CPU), reducing payload size by 85–90% (~30KB) with zero Android app changes.
- Clients lacking `Accept-Encoding` receive standard uncompressed JSON.

---

## Decision 4: Conditional HTTP Requests (`@fastify/etag`) & Dynamic `Cache-Control`

### Context
`stock-routes.ts` currently sends hardcoded `Cache-Control: public, max-age=3600` for all ranges indiscriminately. Fastify does not compute weak ETags without a plugin, preventing mobile clients from validating whether charts have changed.

### Decision
1. Install `@fastify/etag` (compatible with Fastify v5) and register globally in `api/src/app.ts`.
2. Update `api/src/routes/stock-routes.ts`:
   - If `end < todayCairo`: `reply.header("Cache-Control", "public, max-age=86400, immutable");`
   - If `end >= todayCairo`: `reply.header("Cache-Control", "public, max-age=60");`
3. When clients send `If-None-Match`, Fastify evaluates the weak ETag and returns `304 Not Modified` with 0 payload bytes in <10ms.
