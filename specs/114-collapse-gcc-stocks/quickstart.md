# Phase 1: Quickstart Validation Guide

**Feature**: `specs/114-collapse-gcc-stocks`

---

## Prerequisites
- Node.js v20+
- Running baseline server on port 3000 (or candidate deployment on port 3001)
- Valid local/test API key (default: `local-dev-key`)

---

## Verification Scenarios

### Scenario 1: Real-Time Multi-Exchange Verification
Verify that `GET /v1/stocks?market=...` resolves all regional stock exchanges with correct currency envelopes and sub-150ms RAM latency.

```bash
# Query Saudi Tadawul
curl -s -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks?market=saudi" | jq '{market: .market, currency: .currency, stock_count: (.data.stocks | length)}'

# Query UAE (DFM & ADX)
curl -s -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks?market=uae" | jq '{market: .market, currency: .currency, stock_count: (.data.stocks | length)}'

# Query Qatar (QSE)
curl -s -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks?market=qatar" | jq '{market: .market, currency: .currency, stock_count: (.data.stocks | length)}'

# Query Kuwait (Boursa Kuwait)
curl -s -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks?market=kuwait" | jq '{market: .market, currency: .currency, stock_count: (.data.stocks | length)}'

# Query Default (EGX)
curl -s -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks" | jq '{currency: .currency, stock_count: (.data.stocks | length)}'
```

**Expected Outcome**:
Each query returns HTTP `200 OK` with non-empty stock collections matching their national currencies (`SAR`, `AED`, `QAR`, `KWD`, `EGP`).

---

### Scenario 2: Legacy Route Parity Revalidation
Verify that legacy route aliases continue to function identically with zero regression.

```bash
# Verify legacy path aliases
curl -s -o /dev/null -w "%{http_code}\n" -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks/saudi"
curl -s -o /dev/null -w "%{http_code}\n" -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks/uae"
curl -s -o /dev/null -w "%{http_code}\n" -H "x-api-key: local-dev-key" "http://localhost:3000/v1/stocks/regional/catalog"
```

**Expected Outcome**:
All return HTTP `200 OK`.

---

### Scenario 3: Automated Verification Suites
Execute the full regression and differential parity suite:

```bash
# 1. Run canonical verification suite (37 scenarios)
npm run verify

# 2. Run differential parity runner (31 endpoints)
npm run verify:diff

# 3. Run full HTML audit report generator (68 tests)
npm run verify:report
```

**Expected Outcome**:
- `npm run verify` exits with code `0` (37/37 passed).
- `npm run verify:diff` exits with code `0` (31/31 passed).
- `reports/verification-report.html` reports 100% pass rate across all 68 tests.
