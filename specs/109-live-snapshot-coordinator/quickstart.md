# Quickstart & Verification Guide: Live Snapshot Coordinator

**Feature Branch**: `109-live-snapshot-coordinator`  
**Date**: 2026-11-15  
**Spec Reference**: [spec.md](./spec.md) | **Contracts Reference**: [contracts/](./contracts/)

---

## 1. Prerequisites

- Node.js >= 22 LTS
- Local development environment with dependencies installed:
  ```bash
  npm install
  ```

---

## 2. Verification Scenarios

### Scenario 1: Multi-Domain In-Memory Cache Isolation & Retrieval (US1)
Validates that `LiveSnapshotCoordinator` returns distinct, isolated snapshots across all 6 asset domains in $<1\text{ ms}$:
```bash
npx vitest run tests/unit/live-snapshot-coordinator.test.ts
```
**Expected Outcome**: All domains (`rates`, `gold`, `gas`, `stocks`, `indices`, `gcc-stocks`) resolve their respective active snapshots independently from memory.

### Scenario 2: Push-Based Snapshot Ingestion & State Update (US2)
Simulates background push listeners delivering fresh market updates and verifies that memory is atomically updated:
```bash
npx vitest run tests/unit/live-snapshot-ingestion.test.ts
```
**Expected Outcome**: Reads immediately reflect incoming push updates; malformed payloads are rejected while preserving existing valid state.

### Scenario 3: Cold Boot Warming & In-Flight Promise Coalescing (US3)
Simulates cold start and thundering herd conditions:
```bash
npx vitest run tests/unit/live-snapshot-warming.test.ts
```
**Expected Outcome**: In-flight concurrent fetches coalesce to exactly 1 upstream read; cold boots warm from disk in $<500\text{ ms}$.

### Scenario 4: Batch Quotes Resolution via Unified Seam (US4)
Verifies that `QuotesService` resolves mixed portfolio quotes (currencies, gold, indices, regional stocks) from the coordinator:
```bash
npx vitest run tests/unit/quotes-service.test.ts
```
**Expected Outcome**: Quotes resolve in parallel from RAM with zero database reads.

### Scenario 5: Full Regression & Contract Verification
```bash
npm run test:unit
npm run test:contract
```
**Expected Outcome**: 100% green pass rate across all contract snapshots and unit tests.
