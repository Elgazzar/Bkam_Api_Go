# Quickstart Guide: Latest Intraday Timeline Endpoints

**Feature**: `097-latest-intraday-api`  
**Date**: 2026-10-24  
**Status**: Ready for Implementation  

---

## Prerequisites

1. Set up local development environment:
   ```bash
   cd api
   npm install
   ```
2. Ensure environment variables are loaded (`.env` or local defaults):
   ```bash
   export API_KEY="test-api-key"
   ```

---

## Validation Scenarios

### Scenario 1: Query Active Stock Intraday Timeline (Live Session)
Execute during trading hours (or when today's intraday snapshot is populated in RAM):
```bash
curl -s -H "x-api-key: $API_KEY" \
  "http://localhost:3000/v1/stocks/COMI/intraday/latest" | jq .
```
**Expected Outcome**:
- Status: `200 OK`
- Header: `Cache-Control: public, max-age=60`
- Response contains `date` matching today's Cairo date, current `price`, `change`, `changePercent`, and a populated `intraday` array:
  ```json
  {
    "symbol": "COMI",
    "date": "2026-10-24",
    "currency": "EGP",
    "price": 95.5,
    "change": 1.2,
    "changePercent": 1.27,
    "intraday": [
      { "time": "2026-10-24T10:05:00.000Z", "price": 95.5 }
    ]
  }
  ```

---

### Scenario 2: Weekend / Holiday Session Backtracking (Off-Hours)
Execute when today has no trades (e.g. Saturday or prior to 10:00 AM Cairo):
```bash
curl -s -H "x-api-key: $API_KEY" \
  "http://localhost:3000/v1/stocks/COMI/intraday/latest" | jq .
```
**Expected Outcome**:
- Status: `200 OK`
- Header: `Cache-Control: public, max-age=900`
- Response contains `date` of the previous active trading day (e.g., Thursday's date), and the full intraday timeline recorded during that session.
- Sub-millisecond response on subsequent queries due to in-memory fallback buffering.

---

### Scenario 3: Query Inactive Catalog Stock (Zero Trades)
Query a valid listed security with no recorded trades during the resolved session:
```bash
curl -s -H "x-api-key: $API_KEY" \
  "http://localhost:3000/v1/stocks/DAPH/intraday/latest" | jq .
```
**Expected Outcome**:
- Status: `200 OK` (NOT 404)
- `intraday: []`
- Valid `price` and metadata returned.

---

### Scenario 4: Benchmark Indices, Gold & Bank Rates Verification
Verify the remaining three domains:
```bash
# 1. Benchmark Index
curl -s -H "x-api-key: $API_KEY" \
  "http://localhost:3000/v1/indices/EGX30/intraday/latest" | jq .

# 2. Single Gold Karat
curl -s -H "x-api-key: $API_KEY" \
  "http://localhost:3000/v1/gold/21/intraday/latest" | jq .

# 3. Bank FX Rate (USD for CIB)
curl -s -H "x-api-key: $API_KEY" \
  "http://localhost:3000/v1/rates/USD/intraday/latest?bank=cib" | jq .
```

---

### Scenario 5: Automated Test Suite Execution
Run the unit and integration test suites:
```bash
cd api
npm run test:unit
npm run test:integration
```
**Expected Outcome**: All tests pass with 100% green status.
