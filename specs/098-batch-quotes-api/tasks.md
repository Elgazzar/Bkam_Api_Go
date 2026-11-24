# Tasks: Stateless Batch Quotes API Endpoint for Home Favorites

**Input**: Design documents from `specs/098-batch-quotes-api/`
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [data-model.md](./data-model.md), [contracts/quotes-api.yaml](./contracts/quotes-api.yaml), [quickstart.md](./quickstart.md)

## Organization & Execution Rules
- Every task follows the required checklist format: `- [ ] [TaskID] [P?] [Story?] Description with file path`.
- Tasks marked `[P]` are parallelizable (different files, no blocking dependency).
- Story tasks are labeled with their respective story tag (`[US1]`, `[US2]`, `[US3]`).
- Setup, Foundational, and Polish phases have NO story label.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Define domain models, Zod validation schemas, and type contracts.

- [X] T001 [P] Create Quote domain interfaces and Zod validation schemas in api/src/models/quote.ts

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core service and routing infrastructure that MUST be complete before user story logic begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 Implement QuotesService skeleton with dependency injection for rates, gold, indices, stocks, and GCC stocks in api/src/services/quotes-service.ts
- [X] T003 [P] Implement registerQuotesRoutes Fastify plugin skeleton with API key verification in api/src/routes/quotes-routes.ts
- [X] T004 Wire QuotesService and registerQuotesRoutes into Fastify application in api/src/app.ts

**Checkpoint**: Foundation ready - QuotesService and route handlers are registered and ready for story implementations.

---

## Phase 3: User Story 1 - Fetch All Home Favorites in a Single Batch Request (Priority: P1) 🎯 MVP

**Goal**: Mobile clients can submit a list of favorite identifiers (bank currencies, gold, indices, stocks) and receive a unified dictionary of quotes in a single network round-trip via POST or GET.

**Independent Test**: Send a batch request containing valid items across all 4 domains (`CBE_USD`, `21k`, `EGX30`, `COMI`, `1120`) and verify all return with price, change, and freshness metadata in `< 50ms` with zero Firestore reads.

### Tests for User Story 1
- [X] T005 [P] [US1] Create unit test suite for multi-domain quote resolution in api/tests/unit/quotes-service.test.ts
- [X] T006 [P] [US1] Create end-to-end integration test suite for POST and GET /v1/quotes in api/tests/e2e/quotes.e2e.test.ts
(Note: implemented in api/tests/e2e/quotes.e2e.ts adhering to local e2e framework conventions)

### Implementation for User Story 1
- [X] T007 [US1] Implement bank currency quote resolution (`<BANK>_<CURRENCY>`) via RatesSnapshotService in api/src/services/quotes-service.ts
- [X] T008 [US1] Implement gold commodity quote resolution (`24k`, `21k`, `18k`, `pound`, `ounce`) via GoldQueryService in api/src/services/quotes-service.ts
- [X] T009 [US1] Implement benchmark market index resolution (`EGX30`, `EGX70`, `EGX100`, etc.) via IndicesQueryService in api/src/services/quotes-service.ts
- [X] T010 [US1] Implement equity/stock resolution for EGX and GCC markets via StockQueryService and GccStockQueryService in api/src/services/quotes-service.ts
- [X] T011 [US1] Implement POST /v1/quotes handler with JSON body parsing in api/src/routes/quotes-routes.ts
- [X] T012 [US1] Implement GET /v1/quotes handler with comma-separated query string parsing, Cache-Control, and ETag support in api/src/routes/quotes-routes.ts

**Checkpoint**: User Story 1 is fully functional and delivers the core MVP value.

---

## Phase 4: User Story 2 - Resilient Partial Batch Resolution for Unknown or Stale Identifiers (Priority: P2)

**Goal**: Unrecognized, unlisted, or deprecated identifiers are captured in an `unresolved` array while all valid quotes are returned normally, preventing request failures.

**Independent Test**: Submit a payload mixing valid and invalid identifiers (`["CBE_USD", "INVALID_XYZ"]`) and verify `200 OK` with `CBE_USD` in `quotes` and `"INVALID_XYZ"` in `unresolved`.

### Tests for User Story 2
- [X] T013 [P] [US2] Add unit test cases for unresolved identifiers and all-unresolved batches in api/tests/unit/quotes-service.test.ts
- [X] T014 [P] [US2] Add E2E test cases for partial resolution and empty results in api/tests/e2e/quotes.e2e.test.ts
(Note: implemented in api/tests/e2e/quotes.e2e.ts)

### Implementation for User Story 2
- [X] T015 [US2] Implement unresolved item tracking and domain fallback isolation in api/src/services/quotes-service.ts
- [X] T016 [US2] Ensure defensive try/catch error containment per item resolution in api/src/services/quotes-service.ts

**Checkpoint**: User Stories 1 and 2 work independently with full error containment.

---

## Phase 5: User Story 3 - Client Payload Validation and Batch Sizing Safeguards (Priority: P3)

**Goal**: Validate incoming request bounds, deduplicate client identifiers, and enforce maximum batch limits ($\le 100$ items) to protect server resources.

**Independent Test**: Submit empty requests (`[]`), malformed structures, and payloads exceeding 100 items, and verify `400 Bad Request` validation responses.

### Tests for User Story 3
- [X] T017 [P] [US3] Add unit test cases for deduplication and boundary constraints in api/tests/unit/quotes-service.test.ts
- [X] T018 [P] [US3] Add E2E test cases for 400 Bad Request on empty payloads and >100 items in api/tests/e2e/quotes.e2e.test.ts
(Note: implemented in api/tests/e2e/quotes.e2e.ts)

### Implementation for User Story 3
- [X] T019 [US3] Implement identifier deduplication and whitespace normalization in api/src/services/quotes-service.ts
- [X] T020 [US3] Enforce Zod validation middleware and 400 Bad Request error handling in api/src/routes/quotes-routes.ts

**Checkpoint**: All three user stories are functionally complete, resilient, and guarded against payload abuse.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates, exhaustive commenting, and compliance verification across all project standards.

- [X] T021 [P] Update living system requirements in .specify/memory/REQUIREMENTS.md and REQUIREMENTS.md per Principle XVI
- [X] T022 [P] Update architecture documentation in .specify/memory/ARCHITECTURE.md and MODULAR_COMPONENTS.md per Principles X & XV
- [X] T023 [P] Update user-facing endpoint documentation in README.md per Principle X
- [X] T024 Verify exhaustive commenting on every newly added line of code across all modified files per Principle VI
- [X] T025 Execute api unit and E2E test suites (npm run test:unit && npm run test:e2e) per Principle XII
- [X] T026 Execute quickstart validation scenarios from specs/098-batch-quotes-api/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies
- **Phase 1 (Setup)**: No dependencies — start immediately.
- **Phase 2 (Foundational)**: Depends on Phase 1 (T001) — BLOCKS all user stories.
- **Phase 3 (User Story 1 - P1)**: Depends on Phase 2 (T002, T003, T004).
- **Phase 4 (User Story 2 - P2)**: Depends on Phase 3 completion.
- **Phase 5 (User Story 3 - P3)**: Depends on Phase 4 completion.
- **Phase 6 (Polish)**: Depends on all user story phases (Phases 3, 4, 5).

### Parallel Opportunities
- T001 (Setup) is independent.
- T003 (Route skeleton) can run in parallel with T002 (Service skeleton).
- T005 (Unit tests) and T006 (E2E tests) can be authored in parallel before implementation.
- T013 and T014 can run in parallel.
- T017 and T018 can run in parallel.
- T021, T022, and T023 (Documentation updates) can be executed in parallel.

---

## Implementation Strategy: MVP First

1. Complete **Phase 1 (Setup)** and **Phase 2 (Foundational)** to establish models and routing scaffolds.
2. Complete **Phase 3 (User Story 1)**: This delivers the Minimum Viable Product (MVP) enabling mobile apps to fetch home favorites in a single call.
3. Validate User Story 1 against `quickstart.md` Scenario 1 & 2.
4. Complete **Phase 4 (User Story 2)** for partial resolution resilience.
5. Complete **Phase 5 (User Story 3)** for payload validation and batch size safety.
6. Complete **Phase 6 (Polish)** for full documentation sync and E2E verification.
