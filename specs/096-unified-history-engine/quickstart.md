# Quickstart & E2E Validation Guide: Unified History Resolution Engine

This guide describes how to validate the **History Resolution Engine** locally using unit and functional tests, verifying tiered caching from RAM and `flat-cache` without database regressions.

---

## Prerequisites

1. Node.js >= 20.0.0
2. Local repository checked out on branch `096-unified-history-engine`
3. Dependencies installed in `api/`:
   ```bash
   cd api && npm install
   ```

---

## Test Suite Execution

### 1. Unit Tests
Verify that `HistoryResolutionEngine` resolves single-date snapshots and ranges across tiers:
```bash
npm run test:unit -- api/tests/unit/services/history-resolution-engine.test.ts
```

**Expected Results**:
- `getDateSnapshot` hits `ActiveWeeklyBucketStore` in `<1ms` when querying a date in the current ISO week.
- `getDateSnapshot` hits `cacheService` (`flat-cache`) on sealed past weeks.
- `getDateSnapshot` records negative cache on missing dates and returns `null` immediately on subsequent calls without touching Firestore.
- `getDateRange` slices ISO weeks across multi-month intervals and aggregates sorted daily entries.
- Bank Rates range filtering by `bank` and `currency` operates correctly in memory.

### 2. E2E API Regression Tests
Run the Fastify server functional and E2E test suite:
```bash
npm run test:e2e
```

**Expected Results**:
- All historical endpoints (`/v1/rates/history/*`, `/v1/gold/history/*`, `/v1/stocks/history/*`, `/v1/indices/history/*`, `/v1/stocks/:market/history/*`) return valid status codes and conforming JSON schemas.
- `GET /v1/gold` correctly injects `yesterdayBuyPrice` and `yesterdaySellPrice` without errors.
- `GET /v1/rates/history_weekly` is monitored in `FirestoreListenerService` during server startup.

---

## Manual Endpoint Verification (curl)

Start the API server locally:
```bash
npm run dev
```

1. **Verify Bank Rates Single-Date & Active Week**:
   ```bash
   curl -H "x-api-key: your-api-key" http://localhost:3000/v1/rates/history/2026-10-21
   ```
   *Expected*: HTTP 200 with daily rates snapshot. Response header indicates rapid cache resolution.

2. **Verify Bank Rates Multi-Day Range**:
   ```bash
   curl -H "x-api-key: your-api-key" "http://localhost:3000/v1/rates/history/range?start=2026-09-01&end=2026-10-21&bank=cib&currency=usd"
   ```
   *Expected*: HTTP 200 with array of chronologically sorted CIB USD rates.

3. **Verify Gold Single-Date Query**:
   ```bash
   curl -H "x-api-key: your-api-key" http://localhost:3000/v1/gold/history/2026-10-21
   ```
   *Expected*: HTTP 200 returning `karats` dictionary (24k, 21k, 18k, pound).

4. **Verify Market Holiday Negative Caching**:
   ```bash
   # Query a known weekend or Egyptian banking holiday (e.g., Friday)
   curl -H "x-api-key: your-api-key" http://localhost:3000/v1/rates/history/2026-10-16
   ```
   *Expected*: HTTP 404 on initial request, and instant 0ms HTTP 404 on subsequent requests (negative cache hit).
