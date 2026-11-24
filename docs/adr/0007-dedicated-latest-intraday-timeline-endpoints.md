# 7. Dedicated Latest Intraday Timeline Endpoints

## Context
Mobile clients (specifically the Android app) need to render intra-day price charts (Intraday Timelines) for the most recently active market session. Previously, the client had to query a broad historical date range (`GET /v1/<asset>/history/range`), inspect the returned multi-day list to find the most recent trading date with populated data, and then extract the intraday tick series. This introduced several problems:
1. **Network & Payload Bloat**: Clients downloaded multi-day historical snapshots just to locate and render today's or the last trading day's intraday ticks.
2. **Client Complexity**: Client-side logic had to understand market trading schedules, weekends (e.g., Friday/Saturday for EGX and Tadawul vs. Saturday/Sunday for UAE and international), and holiday closures to determine which date represented the latest trading session.
3. **Rollover & Weekend Gaps**: During weekends, exchange holidays, or pre-market hours before opening auctions, today's intraday document is empty. Existing APIs either returned empty intraday arrays or required multi-hop fallback logic.

## Decision
We establish dedicated **Latest Intraday Timeline Endpoints** across all platform asset classes (Stocks, Indices, Gold, and Bank Rates), powered by a centralized resolution method in [`IntradayCoordinator`](../../api/src/services/intraday-coordinator.ts):

1. **Standardized REST Endpoints**:
   - **Stocks**: `GET /v1/stocks/:symbol/intraday/latest` (supports optional `?market=saudi|uae|qatar|kuwait`)
   - **Benchmark Indices**: `GET /v1/indices/:symbol/intraday/latest` (supports optional `?market=...`)
   - **Gold**: `GET /v1/gold/:item/intraday/latest` (where `:item` is `24`, `21`, `18`, `pound`, `international`, or `all`)
   - **Bank Rates**: `GET /v1/rates/:currency/intraday/latest` (supports optional `?bank=...`)

2. **Standardized Response Envelope**:
   Matches the existing history item contract expected by client charts:
   ```json
   {
     "symbol": "COMI",
     "date": "2026-10-22",
     "currency": "EGP",
     "price": 95.5,
     "change": 1.2,
     "changePercent": 1.27,
     "updatedAt": "2026-10-22T14:30:00.000Z",
     "intraday": [
       { "time": "2026-10-22T10:05:00.000Z", "price": 95.5 },
       { "time": "2026-10-22T10:10:00.000Z", "price": 95.8 }
     ]
   }
   ```

3. **Single-Read Backtracking Session Resolution**:
   - If today's active in-memory snapshot (`IntradaySnapshotService`) has populated intraday points, serve today immediately from RAM (<0.1ms, 0 reads).
   - If today has no points (weekend, pre-market, or holiday), query the respective Firestore intraday collection (`*_intraday`) ordered by document ID descending with `limit(1)`. Because daily document IDs are ISO dates (`YYYY-MM-DD`), this resolves the latest recorded session in a single database read.
   - Guard with a 7-day maximum lookback limit. If no session exists within 7 days, fall back to the most recent daily closing quote with `intraday: []`.

4. **Zero-Trade Inactive Security Resiliency**:
   - If a symbol exists in the catalog but had zero trades during the resolved session, return `200 OK` with `intraday: []` and the latest closing price. Never return `404` for valid catalog securities.

5. **Adaptive HTTP Caching**:
   - Active open session: `Cache-Control: public, max-age=60`
   - Concluded / closed session: `Cache-Control: public, max-age=900` (15 minutes)

## Consequences
- Android and web clients can fetch the exact intraday timeline for the latest trading session in a single, lightweight HTTP request without date calculation or range queries.
- Zero extra reads during active sessions (served from memory), and exactly 1 Firestore read during weekends/pre-market hours (subsequently buffered in memory).
- Consistent REST ergonomics across all four asset classes.
