# Tasks: Expose Previous Day Price in Latest Intraday Endpoints

**Input**: Design documents from `/specs/100-intraday-previous-price/`  
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/intraday-latest.openapi.json`, `quickstart.md`  
**Tests**: Integration tests in `api/tests/integration/latest-intraday-routes.test.ts` are mandatory per Constitution Principle XII and Spec User Stories.

## Format: `- [ ] [TaskID] [P?] [Story?] Description with file path`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., [US1], [US2], [US3], [US4])
- File paths are relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Baseline verification of existing test suite and execution environment

- [X] T001 Verify baseline passing state of existing latest intraday route tests by running `cd api; npx vitest run tests/integration/latest-intraday-routes.test.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared response typing and schema contracts for previous session price baselines

**⚠️ CRITICAL**: Must be completed before user story implementation

- [X] T002 [P] Review and verify response envelope types in `api/src/models/stock-rate.ts` and `api/src/models/index-rate.ts` for `previousClose` support
- [X] T003 [P] Review and verify response envelope types in `api/src/models/currency-rate.ts` and `api/src/models/gold-rate.ts` for `previousBuyPrice` and `previousSellPrice` support

**Checkpoint**: Foundational contracts verified - user story implementation can begin

---

## Phase 3: User Story 1 - Equities & Benchmark Indices Intraday Chart Baselines (Priority: P1) 🎯 MVP

**Goal**: Expose top-level `previousClose` in `GET /v1/stocks/:symbol/intraday/latest` and `GET /v1/indices/:symbol/intraday/latest` across Egyptian (EGX) and regional GCC markets (Saudi, UAE, Qatar, Kuwait) sourced from memory with 0 database reads.

**Independent Test**: Query `GET /v1/stocks/COMI/intraday/latest` and `GET /v1/indices/EGX30/intraday/latest`; assert `previousClose` is present as a positive floating-point number.

### Tests for User Story 1
- [X] T004 [P] [US1] Add assertions for `previousClose` across Egyptian stocks, regional GCC stocks, and benchmark indices in `api/tests/integration/latest-intraday-routes.test.ts`

### Implementation for User Story 1
- [X] T005 [P] [US1] Enrich `fetchLatestStockIntraday` with in-memory `previousClose` resolution and fallback chain (`quoteData.previousClose` -> `price - change` -> `catalogEntry.previousClose` -> `price`) in `api/src/services/stock-query-service.ts`
- [X] T006 [P] [US1] Enrich `fetchLatestIndexIntraday` with in-memory `previousClose` resolution and fallback chain (`quoteData.previousClose` -> `price - change` -> `price`) in `api/src/services/indices-query-service.ts`
- [X] T007 [US1] Execute and verify integration tests for User Story 1 with `cd api; npx vitest run tests/integration/latest-intraday-routes.test.ts`

**Checkpoint**: User Story 1 fully functional and verified independently (MVP achieved)

---

## Phase 4: User Story 2 - Bank Exchange Rates Intraday Chart Baselines (Priority: P1)

**Goal**: Expose top-level `previousBuyPrice` and `previousSellPrice` in `GET /v1/rates/:currency/intraday/latest` for both single bank queries (`?bank=cbe`) and consolidated market averages sourced strictly from volatile RAM with 0 database reads.

**Independent Test**: Query `GET /v1/rates/USD/intraday/latest?bank=cbe` and `GET /v1/rates/USD/intraday/latest`; assert `previousBuyPrice` and `previousSellPrice` are present as positive floating-point numbers.

### Tests for User Story 2
- [X] T008 [P] [US2] Add assertions for `previousBuyPrice` and `previousSellPrice` for single bank and consolidated rates in `api/tests/integration/latest-intraday-routes.test.ts`

### Implementation for User Story 2
- [X] T009 [US2] Enrich `fetchLatestRateIntraday` with in-memory `previousBuyPrice` and `previousSellPrice` resolution from `CurrencyRate.yesterdayCloseBuy`/`yesterdayCloseSell` and consolidated averages in `api/src/services/rates-query-service.ts`
- [X] T010 [US2] Execute and verify integration tests for User Story 2 with `cd api; npx vitest run tests/integration/latest-intraday-routes.test.ts`

**Checkpoint**: User Stories 1 and 2 fully functional and verified independently

---

## Phase 5: User Story 3 - Gold Commodities Intraday Chart Baselines (Priority: P1)

**Goal**: Expose `previousBuyPrice` and `previousSellPrice` in `GET /v1/gold/:item/intraday/latest` (single items: `24`, `21`, `18`, `pound`, `international`) and in `GET /v1/gold/intraday/latest` (multi-karat summary envelope) sourced from memory with 0 database reads.

**Independent Test**: Query `GET /v1/gold/21/intraday/latest`, `GET /v1/gold/international/intraday/latest`, and `GET /v1/gold/intraday/latest`; assert `previousBuyPrice` and `previousSellPrice` are present at the top level of single items and within each nested karat object.

### Tests for User Story 3
- [X] T011 [P] [US3] Add assertions for `previousBuyPrice` and `previousSellPrice` for single gold items (`21`, `international`) and multi-karat envelopes in `api/tests/integration/latest-intraday-routes.test.ts`

### Implementation for User Story 3
- [X] T012 [US3] Enrich `fetchLatestGoldIntraday` with in-memory `previousBuyPrice` and `previousSellPrice` resolution for single karats, gold pound, international spot gold, and multi-karat envelopes in `api/src/services/gold-query-service.ts`
- [X] T013 [US3] Execute and verify integration tests for User Story 3 with `cd api; npx vitest run tests/integration/latest-intraday-routes.test.ts`

**Checkpoint**: User Stories 1, 2, and 3 fully functional and verified independently

---

## Phase 6: User Story 4 - Zero-Cost In-Memory Resolution & Documentation (Priority: P2)

**Goal**: Document updated response shapes, fields, TypeScript types, and cURL examples across all four asset classes in public API documentation with 0 database read guarantees.

**Independent Test**: Review `API_ENDPOINTS_GUIDE.md` to ensure `previousClose`, `previousBuyPrice`, and `previousSellPrice` are fully documented with accurate JSON response examples.

### Implementation for User Story 4
- [X] T014 [P] [US4] Update `API_ENDPOINTS_GUIDE.md` documenting `previousClose` for `/v1/stocks/:symbol/intraday/latest` and `/v1/indices/:symbol/intraday/latest`
- [X] T015 [P] [US4] Update `API_ENDPOINTS_GUIDE.md` documenting `previousBuyPrice` and `previousSellPrice` for `/v1/rates/:currency/intraday/latest` and `/v1/gold/:item/intraday/latest`
- [X] T016 [US4] Verify response latency remains <10ms and 0 Firestore read charges occur during execution

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation synchronization, constitution compliance, and full regression test execution

- [X] T017 [P] Update functional requirements in `.specify/memory/REQUIREMENTS.md` and `REQUIREMENTS.md` per Principle XVI
- [X] T018 [P] Update architectural documentation in `.specify/memory/ARCHITECTURE.md` and `ARCHITECTURE.md` per Principles X & XI
- [X] T019 [P] Update modular component catalog in `.specify/memory/MODULAR_COMPONENTS.md` per Principle XV
- [X] T020 [P] Update root `README.md` with release notes and endpoint guides per Principle X
- [X] T021 Verify exhaustive code commenting on every newly added or modified line in query services and test files per Principle VI
- [X] T022 Execute full test suite `cd api; npm test` to ensure zero regressions across all services
- [X] T023 Run manual verification scenarios per `specs/100-intraday-previous-price/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1 (Equities & Indices): Can proceed immediately after Phase 2
  - US2 (Bank Rates): Can proceed independently or in parallel with US1
  - US3 (Gold Commodities): Can proceed independently or in parallel with US1/US2
  - US4 (Documentation & Performance): Depends on US1, US2, and US3 implementation
- **Polish (Phase 7)**: Depends on all user stories being complete

### Parallel Opportunities

- T002 & T003 in Foundational Phase can run in parallel
- T005 (`stock-query-service.ts`) & T006 (`indices-query-service.ts`) in US1 can run in parallel
- T008 (US2 test) & T011 (US3 test) can run in parallel
- T014 & T015 in US4 can run in parallel
- T017, T018, T019, T020 in Polish Phase can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch test and service updates for Equities and Indices in parallel:
Task: "Enrich fetchLatestStockIntraday in api/src/services/stock-query-service.ts"
Task: "Enrich fetchLatestIndexIntraday in api/src/services/indices-query-service.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify tests pass)
2. Complete Phase 2: Foundational (types check)
3. Complete Phase 3: User Story 1 (Equities & Indices `previousClose`)
4. **STOP and VALIDATE**: Run `latest-intraday-routes.test.ts` to confirm US1 works independently
5. Proceed to US2, US3, US4 incrementally
