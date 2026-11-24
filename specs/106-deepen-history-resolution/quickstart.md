# Quickstart Validation Guide: Deepen History Resolution Engine

**Feature**: `106-deepen-history-resolution`  
**Date**: 2026-11-15

This quickstart guide documents the step-by-step verification procedures to validate that the deepened `HistoryResolutionEngine` works end-to-end across active session queries, historical archives, and multi-day ranges.

---

## 1. Prerequisites & Environment

Ensure you are in the `Bkam_Api` repository with Node.js 22+ and dependencies installed:

```powershell
cd D:\Repo\Bkam\Bkam_Api
git status # Confirm branch is on 106-deepen-history-resolution
```

---

## 2. Validation Scenarios

### Scenario 1: Unit & Integration Verification

Run the existing unit test suite to verify internal engine contracts, timezone calculations, and intraday coordinations:

```bash
npm run test:unit
```

**Expected Outcome:**
- All 27 test files pass (225+ tests green).
- Tests for `history-resolution-engine.test.ts`, `gold-query-service.test.ts`, `rates-query-service.test.ts`, and `intraday-coordinator.test.ts` pass cleanly without mock regressions.

---

### Scenario 2: Active Session (Today) Date Query Validation

Query the history endpoint for today's active date across foreign exchange, gold, and equities:

```powershell
# Using the live server or local development instance
npm run verify:remote -- --url http://84.235.253.78:3001 --key local-dev-key --category rates
npm run verify:remote -- --url http://84.235.253.78:3001 --key local-dev-key --category commodities
npm run verify:remote -- --url http://84.235.253.78:3001 --key local-dev-key --category stocks
```

**Expected Outcome:**
- Endpoints return HTTP 200 OK with synthesized active rates and intraday timeline points.
- Response header includes `Cache-Control: public, max-age=60`.

---

### Scenario 3: Concluded Past Date Query Validation

Query the history endpoint for a finalized historical date (e.g. `2026-11-12`):

```bash
curl -i -H "x-api-key: local-dev-key" http://84.235.253.78:3001/v1/rates/history/2026-11-12
curl -i -H "x-api-key: local-dev-key" http://84.235.253.78:3001/v1/gold/history/2026-11-12
curl -i -H "x-api-key: local-dev-key" http://84.235.253.78:3001/v1/stocks/history/2026-11-12
```

**Expected Outcome:**
- Endpoints return HTTP 200 OK with archived closing snapshots.
- Response header includes `Cache-Control: public, max-age=3600`.

---

### Scenario 4: Multi-Day Date Range Spanning into Today

Query a multi-day range ending today:

```bash
curl -i -H "x-api-key: local-dev-key" "http://84.235.253.78:3001/v1/rates/history/range?start=2026-11-10&end=2026-11-15&currency=USD"
```

**Expected Outcome:**
- Returns an array of daily points where historical days contain archived records and the final entry seamlessly contains today's live synthesized quote.

---

### Scenario 5: Full Regression & Differential Parity Verification

Run both the 36-scenario verification runner and the 30-endpoint differential comparator against the baseline server:

```bash
# 1. Full schema suite (36 scenarios)
npm run verify:remote -- --url http://84.235.253.78:3001 --key local-dev-key

# 2. Differential parity comparator (30 endpoints vs port 3000)
npm run verify:diff -- --base http://84.235.253.78:3000 --target http://84.235.253.78:3001 --key local-dev-key
```

**Expected Outcome:**
- `verify:remote`: 36/36 scenarios pass (0 failures).
- `verify:diff`: 30/30 endpoints achieve 100% parity (0 diffs/regressions).

