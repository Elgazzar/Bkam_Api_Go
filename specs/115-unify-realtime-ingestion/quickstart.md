# Quickstart & Verification Guide: Unify Real-Time Ingestion and Active Weekly Caching

**Feature**: `115-unify-realtime-ingestion`  
**Date**: 2026-11-16  

---

## 1. Overview

Feature 115 unifies real-time push ingestion and active calendar week historical caching by:
1. Absorbing `ActiveWeeklyBucketStore` directly into `HistoryResolutionEngine` (Tier 1 Memory Store).
2. Deleting obsolete compatibility shims (`SnapshotCache`, `SnapshotStore`) and dead facades (`market-schedule-service.ts`).
3. Streamlining `FirestoreListenerService` into a clean Ports & Adapters inbound push listener.

---

## 2. Verification Steps

### Step 1: TypeScript Build & Type Check
Verify that all import paths, constructor signatures, and removed shims compile with zero errors:
```bash
npm run build
```
Expected outcome: `tsc -p tsconfig.build.json` exits with code `0`.

### Step 2: Unit Test Suite
Execute the Vitest unit test suite to verify that `HistoryResolutionEngine`, `RatesSnapshotService`, `FirestoreListenerService`, and all query services pass:
```bash
npm run test:unit
```
Expected outcome: All 37 test files and 320+ unit tests pass.

### Step 3: Deletion Test Validation
Confirm that the 4 targeted shallow/shim files have been removed from the repository:
```bash
# Verify files no longer exist:
test -f src/services/active-weekly-bucket-store.ts && echo "FAILED" || echo "PASSED: active-weekly-bucket-store.ts deleted"
test -f src/cache/snapshot-cache.ts && echo "FAILED" || echo "PASSED: snapshot-cache.ts deleted"
test -f src/cache/snapshot-store.ts && echo "FAILED" || echo "PASSED: snapshot-store.ts deleted"
test -f src/services/market-schedule-service.ts && echo "FAILED" || echo "PASSED: market-schedule-service.ts deleted"
```

### Step 4: Full Deployment & Parity Verification
Run the 49-scenario verification suite against the running gateway:
```bash
npm run verify
```
Expected outcome: **All 49 scenarios pass** with exit code `0`.

Run the full HTML audit report and differential parity check:
```bash
npm run verify:report
```
Expected outcome: **80/80 checks pass (100%)** across Gate 1 and Gate 2.
