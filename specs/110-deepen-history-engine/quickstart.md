# Quickstart & Verification Guide: Deepen Historical Engine

**Feature Branch**: `110-deepen-history-engine`  
**Date**: 2026-11-15  
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

---

## 1. Prerequisites

- Node.js 22+
- Dependencies installed: `npm install`
- Working tree on branch `110-deepen-history-engine`

---

## 2. Test Execution Commands

### 2.1 Unit & Contract Test Suite
Verify that all unit and contract tests continue to pass with 0 regressions:
```powershell
npm run test:unit
npm run test:contract
```

### 2.2 Dedicated History Engine Tests
Execute unit tests covering `HistoryResolutionEngine`:
```powershell
npx vitest run tests/unit/history-resolution-engine.test.ts
```

### 2.3 Domain Query Service Tests
Verify that callers (`StockQueryService`, `GoldQueryService`, `IndicesQueryService`, `RatesQueryService`) work correctly after delegating range queries to the engine:
```powershell
npx vitest run tests/unit/stock-query-service.test.ts
npx vitest run tests/unit/gold-query-service.test.ts
npx vitest run tests/unit/indices-query-service.test.ts
```

### 2.4 Verify Post-Deployment Parity
Verify 100% parity against baseline server (port 3000 vs port 3001) across all historical endpoints:
```powershell
npm run verify:diff -- --base http://84.235.253.78:3000 --target http://84.235.253.78:3001 --key local-dev-key
```

---

## 3. Key Verification Scenarios

### Scenario 1: Pure Historical Range from Disk Cache
- **Request**: `GET /v1/stocks/COMI/history/range?start=2026-10-01&end=2026-10-15`
- **Expected Outcome**:
  - First request computes range from weekly buckets and writes to `flat-cache` under key `stock_range:COMI:2026-10-01:2026-10-15`.
  - Subsequent requests resolve from `flat-cache` in $<10\text{ ms}$ with zero Firestore reads.
  - HTTP `Cache-Control: public, max-age=3600`.

### Scenario 2: Range Spanning into Dynamic Today
- **Request**: `GET /v1/gold/history/range?start=2026-11-10&end=2026-11-15` (assuming today is `2026-11-15`)
- **Expected Outcome**:
  - Past days [2026-11-10 to 2026-11-14] resolved from disk prefix cache (`gold_range_prefix:...`).
  - Active today (`2026-11-15`) injected from live quotes and intraday points.
  - HTTP `Cache-Control: public, max-age=60`.

### Scenario 3: Non-Trading Date Negative Cache
- **Request**: `GET /v1/stocks/history/2026-11-13` (an Egyptian Friday / weekend)
- **Expected Outcome**:
  - First request detects absence in weekly bucket document, records entry in `missingDatesCache` for 24h, and returns empty/404.
  - Repeated request resolves immediately as `null` from `missingDatesCache` without Firestore reads.
