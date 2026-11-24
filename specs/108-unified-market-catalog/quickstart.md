# Quickstart & Verification Guide: Unified Market Catalog

**Branch**: `108-unified-market-catalog` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

This guide documents runnable validation scenarios that verify the deepened `MarketCatalog` module and ensure zero regressions across existing HTTP endpoints.

---

## Prerequisites & Setup

1. **Working Directory**: `d:\Repo\Bkam\Bkam_Api`
2. **Environment**: Node.js 22 LTS, npm dependencies installed.
3. **Contracts**: Refer to [market-catalog.contract.ts](./contracts/market-catalog.contract.ts) and [data-model.md](./data-model.md).

---

## Verification Scenarios

### Scenario 1: Universal Security & Alias Lookup (Unit Test)
Verifies that domestic and regional securities resolve by ticker symbol, numeric code, or alias with automatic normalization.

```bash
npm run test:unit tests/unit/market-catalog.test.ts
```

**Expected Results**:
- `getSecurity("COMI")` returns EGX Commercial International Bank profile.
- `getSecurity(" comi ")` normalizes whitespace/casing and returns the same profile.
- `getSecurity("COMI.CA")` strips legacy exchange suffix and returns the same profile.
- `getSecurity("1010")` returns Tadawul Al Rajhi Bank profile.
- `getSecurity("BDC")` resolves via alias to Banque du Caire (`BQDC`).
- `getSecurity("UNKNOWN_SECURITY")` returns `undefined`.

---

### Scenario 2: Benchmark Index Constituent Resolution (Integration Test)
Verifies that index constituent queries return authoritative member lists in O(1) time without live array traversal.

```bash
npm run test:unit tests/integration/indices-constituents.test.ts
```

**Expected Results**:
- `getConstituents("EGX30")` returns exactly 30 EGX blue-chip securities.
- `getConstituents("TASI")` returns all registered Tadawul constituent equities.
- Constituent security objects include canonical symbols, localized Arabic and English names, and verified logo URLs.
- Invalid index codes return empty array or trigger standard `INVALID_VALIDATION` rejection.

---

### Scenario 3: Snapshot Metadata Enrichment (Unit Test)
Verifies that raw quote snapshots are enriched with official branding logos and index tags.

```bash
npm run test:unit tests/unit/stock-query-service.test.ts
```

**Expected Results**:
- Quotes missing logos receive verified HTTPS logo URLs from the catalog.
- Quotes missing index tags are populated with authoritative benchmark affiliations.
- Pre-existing pricing figures (e.g. `price`, `change`, `changePercent`) are untouched.

---

### Scenario 4: Parameter Validation Consistency (Contract & Route Tests)
Verifies that routes enforce unified parameter validation.

```bash
npm run test:contract
npm run test:unit tests/unit/catalog-validation.test.ts
```

**Expected Results**:
- `GET /v1/stocks?group=EGX30` returns `200 OK` with filtered equities.
- `GET /v1/stocks?group=InvalidGroup` returns `400 Bad Request` with `INVALID_VALIDATION`.
- `GET /v1/indices/INVALID/constituents` returns `400 Bad Request`.
- All contract schema tests pass with 0 schema violations.

---

## Full Regression Suite

```bash
npm run test:unit
```
*Criteria*: All 27+ test suites must pass 100% green before merge.
