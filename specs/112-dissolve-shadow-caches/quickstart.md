# Quickstart & Verification Guide: Dissolve Shadow Caches

**Feature Branch**: `112-dissolve-shadow-caches`  
**Date**: 2026-11-15  

---

## 1. Overview & Verification Goals

This guide documents runnable scenarios verifying that:
1. `QuotesService` successfully resolves multi-asset quotes with only `LiveSnapshotCoordinator` injected.
2. `GoldQueryService` and `GasQueryService` serve live quotes without private shadow cache fields (`cachedPrices`).
3. `StockQueryService` and `IndicesQueryService` serve live quotes without private shadow caches.
4. Single-flight promise pooling prevents duplicate upstream Firestore reads during concurrent cache misses.
5. All 30 public API gateway endpoints achieve 100% differential parity with zero regressions.

---

## 2. Test Execution Commands

### Scenario 1: Unit Test Validation

Run the dedicated unit test suites for `QuotesService`, `StockQueryService`, `GoldQueryService`, and `LiveSnapshotCoordinator`:

```bash
# Run quotes service unit tests
npx vitest run tests/unit/quotes-service.test.ts tests/unit/quotes-service-coordinator.test.ts

# Run commodity and stock query service unit tests
npx vitest run tests/unit/gold-query-service.test.ts tests/unit/stock-query-service.test.ts tests/unit/live-snapshot-coordinator.test.ts

# Run all unit tests
npm run test:unit
```

**Expected Outcome**: All 36+ test files pass (310+ unit tests) with 0 failures.

---

### Scenario 2: Contract Schema Validation

Validate that all contract schemas match client expectations:

```bash
npm run test:contract
```

**Expected Outcome**: All 3 contract test files pass (9 tests).

---

### Scenario 3: TypeScript Build Compilation

Verify that removing shadow cache fields and unused dependencies introduces zero compiler errors:

```bash
npm run build
```

**Expected Outcome**: `tsc -p tsconfig.build.json` terminates with exit code 0.

---

### Scenario 4: Differential Parity Verification

Run side-by-side verification between the baseline production server (port 3000) and the candidate server (port 3001):

```bash
npm run verify:diff -- --base http://84.235.253.78:3000 --target http://84.235.253.78:3001 --key local-dev-key
```

**Expected Outcome**:
- All 30/30 endpoints report `✔ MATCH`.
- Zero payload schema drift or value discrepancies.
