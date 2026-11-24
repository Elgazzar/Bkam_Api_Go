# Quickstart & Verification Guide: Deepen Intraday Coordinator

**Branch**: `107-deepen-intraday-coordinator` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

---

## Prerequisites

1. Active git branch `107-deepen-intraday-coordinator` across both repositories:
   - `D:\Repo\Bekam_Fe_El_bank_Scrapper`
   - `D:\Repo\Bkam\Bkam_Api`
2. Node.js 22 LTS installed.
3. Clean TypeScript compilation:
   ```bash
   cd D:\Repo\Bkam\Bkam_Api
   npm run build
   ```

---

## Runnable Verification Scenarios

### Scenario 1: Bank FX Rates Latest Intraday
Query the latest USD exchange rate intraday timeline for National Bank of Egypt (NBE):
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/rates/intraday/latest?currency=USD&bank=NBE" -i
```
- **Expected Outcome**:
  - HTTP `200 OK`
  - Headers: `Cache-Control: public, max-age=60` (if active today) or `max-age=3600` (if off-hours backtracked)
  - JSON Body contains `currency: "USD"`, `bank: "NBE"`, `buy`, `sell`, `previousBuyPrice`, `previousSellPrice`, `intraday: [...]`.

---

### Scenario 2: Gold Commodities Latest Intraday
Query the consolidated gold board for all karats:
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/gold/intraday/latest?item=all" -i
```
- **Expected Outcome**:
  - HTTP `200 OK`
  - JSON Body contains `karats: { "24": { buy, sell, previousBuyPrice, previousSellPrice, intraday }, "21": ..., "18": ..., "pound": ... }` and `international: ...`.

---

### Scenario 3: Equities Latest Intraday (Domestic EGX & Regional GCC)
Query Egypt's benchmark commercial bank (`COMI`) and Saudi Aramco (`ARAMCO`):
```bash
# Egypt EGX
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/stocks/COMI/intraday/latest" -i

# Saudi Arabia (Tadawul)
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/stocks/saudi/ARAMCO/intraday/latest" -i
```
- **Expected Outcome**:
  - HTTP `200 OK`
  - JSON Body contains `symbol`, `price`, `change`, `changePercent`, `previousPrice`, and chronological `intraday: [{ time, price }]`.

---

### Scenario 4: Benchmark Indices Latest Intraday
Query the primary EGX30 benchmark index:
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/indices/EGX30/intraday/latest" -i
```
- **Expected Outcome**:
  - HTTP `200 OK`
  - JSON Body contains `code: "EGX30"`, `value`, `change`, `changePercent`, `previousValue`, and `intraday: [{ time, price }]`.

---

## Automated Test Suites

### Unit Test Execution
Execute the full Vitest suite in `Bkam_Api`:
```bash
cd D:\Repo\Bkam\Bkam_Api
npm run test:unit
```
- **Target**: 100% pass across all unit test suites (230+ tests).

### Remote Verification & Differential Parity
Run the 36-scenario verification suite and 30-endpoint differential comparison against baseline:
```bash
cd D:\Repo\Bkam\Bkam_Api
npm run verify:remote -- --url http://84.235.253.78:3001 --key local-dev-key
npm run verify:diff -- --base http://84.235.253.78:3000 --target http://84.235.253.78:3001 --key local-dev-key
```
- **Target**: 36/36 scenarios passed, 30/30 differential endpoints matched with 100% parity.

