# Tasks: Unify Real-Time Ingestion and Active Weekly Caching

**Input**: Design documents from `specs/115-unify-realtime-ingestion/`  
**Prerequisites**: `spec.md`, `plan.md`, `research.md`, `data-model.md`, `contracts/ingestion-api.md`, `quickstart.md`  
**Organization**: Tasks are grouped by foundational preparation, prioritized user stories (US1 through US3), concluding with polish, verification, and constitution compliance.

---

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Associated user story (US1, US2, US3)

---

## Phase 1: Foundational (Tier 1 Weekly Memory Storage in HistoryResolutionEngine)

**Purpose**: Deepen `HistoryResolutionEngine` to natively own active calendar week in-memory caching, preparing the seam for `ActiveWeeklyBucketStore` elimination.

- [X] T001 Deepen `HistoryResolutionEngine` in `src/services/history-resolution-engine.ts` by adding internal `activeWeeklyBuckets: Map<string, Record<string, any>>` and public Tier 1 methods: `getActiveWeeklyBucket`, `setActiveWeeklyBucket`, `deleteActiveWeeklyBucket`, `clearActiveWeek`, and `clearActiveWeeklyBuckets`
- [X] T002 Update internal weekly resolution calls in `src/services/history-resolution-engine.ts` (lines 676, 972, 1261) to query and populate `this.activeWeeklyBuckets` directly instead of referencing external `activeWeeklyBucketStore`
- [X] T003 [P] Add unit tests in `tests/unit/history-resolution-engine.test.ts` verifying `HistoryResolutionEngine` Tier 1 active weekly bucket memory operations (get, set, delete, clearWeek, and clearAll)

**Checkpoint**: `HistoryResolutionEngine` natively manages its Tier 1 memory storage; ready for push listener integration.

---

## Phase 2: User Story 1 - Streamlined Real-Time Market Quote Ingestion (Priority: P1) 🎯 MVP

**Goal**: Transform `FirestoreListenerService` into a lean Ports & Adapters inbound push listener that updates `LiveSnapshotCoordinator` and `HistoryResolutionEngine` directly, eliminating circular query service calls.

- [X] T004 [US1] Refactor `FirestoreListenerService` constructor in `src/services/firestore-listener-service.ts` to accept only `LiveSnapshotCoordinator` and `HistoryResolutionEngine`, removing the 5 query service parameters (`ratesService`, `goldService`, `gasService`, `stockService`, `indicesService`, and `gccStockService`)
- [X] T005 [US1] Eliminate circular `.updateCache()` calls across all snapshot listener handlers (`rates`, `gold_prices`, `gas_prices`, `stock_prices`, `index_prices`, `gcc_stock_prices_*`) in `src/services/firestore-listener-service.ts`, updating `LiveSnapshotCoordinator` directly
- [X] T006 [US1] Update `FirestoreListenerService.startWeeklyListeners()` and `rolloverWeeklyListeners()` in `src/services/firestore-listener-service.ts` to call `this.historyEngine.setActiveWeeklyBucket()` and `this.historyEngine.clearActiveWeek()` instead of `activeWeeklyBucketStore`
- [X] T007 [US1] Update `src/app.ts` listener instantiation to instantiate `FirestoreListenerService(liveSnapshotCoordinator, historyResolutionEngine)`
- [X] T008 [P] [US1] Update `tests/unit/firestore-listener-service.test.ts` to reflect the streamlined constructor and verify direct coordinator and history engine update dispatches

**Checkpoint**: Real-time push ingestion is direct and non-circular. Inbound adapter strictly delegates to core engines.

---

## Phase 3: User Story 2 - Authoritative Active Calendar Week Resolution & Store Elimination (Priority: P2)

**Goal**: Complete elimination of `src/services/active-weekly-bucket-store.ts` across the entire codebase.

- [X] T009 [US2] Update unit test suites that reference `activeWeeklyBucketStore` to use `historyResolutionEngine` Tier 1 methods:
  - `tests/unit/gold-query-service.test.ts`
  - `tests/unit/indices-query-service.test.ts`
  - `tests/unit/stock-query-service.test.ts`
  - `tests/unit/history-resolution-engine.test.ts`
- [X] T010 [US2] Delete obsolete shallow module `src/services/active-weekly-bucket-store.ts` (76 lines)
- [X] T011 [US2] Delete obsolete test file `tests/unit/active-weekly-bucket-store.test.ts` (47 lines)

**Checkpoint**: `ActiveWeeklyBucketStore` is completely eradicated. Historical queries resolve through unified engine memory.

---

## Phase 4: User Story 3 - Clean Seam Boundary & Legacy Facade Removal (Priority: P3)

**Goal**: Purge dead facades and legacy snapshot shims.

- [X] T012 [US3] Refactor `RatesSnapshotService` constructor in `src/services/rates-snapshot-service.ts` to inject `LiveSnapshotCoordinator` directly, removing dependencies on `SnapshotCache` and `SnapshotStore`
- [X] T013 [US3] Update `src/routes/health-routes.ts` to replace `SnapshotCache` import with `LiveSnapshotCoordinator`
- [X] T014 [US3] Update `tests/unit/rates-snapshot-service.test.ts` and `tests/integration/stale-fallback.integration.test.ts` to inject `LiveSnapshotCoordinator`
- [X] T015 [US3] Delete legacy shims `src/cache/snapshot-cache.ts` (27 lines) and `src/cache/snapshot-store.ts` (46 lines)
- [X] T016 [US3] Remove unused import of `marketScheduleService` in `tests/unit/market-schedule-engine.test.ts`
- [X] T017 [US3] Delete dead facade `src/services/market-schedule-service.ts` (62 lines)

**Checkpoint**: All 4 targeted shallow/shim/dead code files deleted; callers use authoritative core engines directly.

---

## Phase 5: Polish, Constitution Compliance & Verification

**Purpose**: Enforce all constitutional requirements, verify build, and run the verification suites.

- [X] T018 Verify 100% comment coverage across all modified and newly authored code lines per Constitution Principle V
- [X] T019 Bump application version from `2.18.19` to `2.18.20` in `package.json`, `src/routes/version-routes.ts`, and `tests/unit/version-routes.test.ts` per Constitution Principle IX
- [X] T020 Run `npm run build` to verify clean TypeScript compilation with zero type errors
- [X] T021 Run `npm run test:unit` to verify all 320+ unit tests pass
- [X] T022 Run `npm run verify` to verify all 49 platform deployment scenarios pass with exit code `0`
- [X] T023 Run `npm run verify:report` to confirm 100% differential parity and contract conformance

---

## Dependencies & Execution Sequence

```
Phase 1: Foundational (T001 - T003)
   │
   ▼
Phase 2: User Story 1 (T004 - T008) 🎯 MVP
   │
   ▼
Phase 3: User Story 2 (T009 - T011)
   │
   ▼
Phase 4: User Story 3 (T012 - T017)
   │
   ▼
Phase 5: Polish & Verification (T018 - T023)
```
