# 5. Tiered Weekly History Caching and API Optimization

## Context
Client applications fetching multi-month historical equity charts (e.g., 6 months spanning 181 calendar days) via `GET /v1/stocks/:symbol/history/range` experienced excessive latency (5 to 15+ seconds).

Investigation revealed three fundamental bottlenecks:
1. **Query Fanout & Negative Caching Gaps**: Historical range queries issued individual daily lookups across all 181 dates instead of leveraging consolidated ISO weekly buckets (`stock_prices_history_weekly/{YYYY-Www}`). Unnecessary secondary fallbacks were triggered for non-trading days (weekends and Egyptian market holidays).
2. **Current Week Disk Caching Dilemma**: While past weeks are sealed and immutable, the active calendar week mutates continuously during trading sessions. Naively saving the active week to `flat-cache` risks serving stale prices after server restarts, whereas bypassing disk entirely forces Firestore reads on every server boot even when the market is closed (evenings and the 48-hour Friday/Saturday weekend).
3. **HTTP Transport & Intraday Bloat**: Historical responses lacked HTTP compression and weak ETags, sending large uncompressed JSON payloads over mobile networks. Furthermore, high-frequency intraday tick points—meant only for live trading charts—were being queried across historical ranges, increasing payload size by over 500%.

## Decision
We established a tiered caching, storage, and transport optimization architecture:

1. **Session-Aware Disk Caching for Active Week**:
   - Sealed past weeks (`weekKey !== currentWeekKey`) are permanently cached to `flat-cache` on disk (30-day TTL).
   - The active week (`weekKey === currentWeekKey`) strictly bypasses disk writes during active trading hours (Sunday–Thursday 10:00 AM – 3:00 PM Cairo) and relies on in-memory listeners and 5-second promise deduplication.
   - Outside operating hours (evenings, nights, and all day Friday and Saturday), the active week is persisted to `flat-cache` with an expiration timestamp synchronized to the next market open (09:30 AM Cairo of the next trading day).

2. **Strict Intraday Segregation from Disk Storage**:
   - High-frequency intraday tick points exist strictly in memory (`intradaySnapshotService` in RAM) for the current active trading session.
   - Historical weekly buckets and disk caches (`flat-cache`) store only daily closing prices (`date`, `symbol`, `price`, `change`, `changePercent`, `updatedAt`).
   - Intraday tick points are never written to `flat-cache` and are deleted from active memory upon daily session rollover.

3. **Range Result Cache with Dynamic Today Stitching**:
   - For multi-month range queries (`GET /v1/stocks/:symbol/history/range`), the closed historical prefix (`[startDate, yesterdayCairo]`) is cached in `flat-cache` under a symbol-scoped key (`stock_range:{symbol}:{start}:{yesterday}`).
   - If the range includes dynamic today, today's live closing price and active intraday points are stitched from RAM at request time. Repeat historical requests resolve in `<1ms` without re-evaluating weekly documents.

4. **HTTP Transport Optimization**:
   - Install `@fastify/compress` with a 1KB threshold to apply transparent Gzip and Brotli compression, shrinking payload sizes by 85–90% for mobile clients (e.g., Android OkHttp).
   - Register `@fastify/etag` to support conditional `If-None-Match` requests, returning `304 Not Modified` with zero payload bytes when clients revalidate.
   - Apply dynamic `Cache-Control` headers: `public, max-age=86400, immutable` for historical ranges ending before today, and `public, max-age=60` for ranges ending today.

## Consequences
- Multi-month stock history queries drop from 5–15s cold / 3s warm down to <300ms cold and <1ms warm.
- Server cold starts and restarts outside trading hours (nights and weekends) achieve 100% cache hits from disk with zero Firestore reads.
- Mobile client bandwidth consumption drops by ~90% due to transparent Gzip/Brotli compression, with instant 0ms chart loading via HTTP 304 revalidation.
- Zero risk of serving stale market quotes during live trading sessions.
