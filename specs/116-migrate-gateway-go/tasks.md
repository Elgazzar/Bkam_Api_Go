# Tasks: Migrate Bkam Financial API Gateway to Idiomatic Go

**Input**: Design documents from `/specs/116-migrate-gateway-go/` (`spec.md`, `plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`)  
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/http-api.md`, `contracts/firestore-push.md`  
**Tests**: Unit tests with `-race` detection, contract integration tests, and platform verification suite `npm run verify` (49 scenarios) per Constitution Principles IV, V, and VII.  
**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

---

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (`[US1]`, `[US2]`, `[US3]`, `[US4]`)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, directory structure, Go module setup, and base configuration.

- [X] T001 Initialize Go module and configure project layout in go.mod
- [X] T002 [P] Install core runtime dependencies (chi/v5, google firestore SDK, singleflight, cron/v3) in go.mod
- [X] T003 [P] Implement environment configuration loader in internal/config/config.go
- [X] T004 [P] Implement standardized HTTP error models and JSON error responder in internal/adapter/http/error.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core domain catalog, disk storage persistence, and Cairo market schedule engine that MUST be complete before ANY user story can be implemented.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T005 [P] Implement bank catalog and currency metadata registry in internal/domain/catalog.go
- [X] T006 [P] Implement thread-safe filesystem JSON persistence adapter with atomic writes in internal/adapter/storage/disk_adapter.go
- [X] T007 Implement MarketScheduleEngine evaluating Africa/Cairo sessions, holidays, and dynamic TTLs in internal/engine/market_schedule.go
- [X] T008 [P] Unit test market schedule engine and holiday rules in internal/engine/market_schedule_test.go
- [X] T009 [P] Unit test filesystem disk adapter persistence and hydration in internal/adapter/storage/disk_adapter_test.go

**Checkpoint**: Foundation ready - catalog, storage adapter, and schedule engine fully operational and tested.

---

## Phase 3: User Story 1 - Real-Time Market Quotes via Push Streaming (Priority: P1) 🎯 MVP

**Goal**: Deliver real-time market quote updates pushed from upstream Firestore snapshot streams (rates, gold, petroleum gas, EGX stocks, GCC equities, indices) directly to client API responses with sub-millisecond memory lookups and zero direct Firestore reads during serving.

**Independent Test**: Publish mock or live snapshot updates to upstream collections/documents and verify that GET requests across `/v1/rates`, `/v1/gold`, `/v1/gas`, `/v1/stocks`, and `/v1/indices` immediately return fresh quotes with updated trend values and metadata.

### Tests for User Story 1

- [X] T010 [P] [US1] Unit test LiveSnapshotCoordinator thread safety and singleflight coalescing in internal/engine/live_snapshot_test.go
- [X] T011 [P] [US1] Unit test RatesProcessor enrichment, trend detection, and availability in internal/adapter/firestore/rates_processor_test.go

### Implementation for User Story 1

- [X] T012 [P] [US1] Create currency rate and bank availability domain models in internal/domain/rates.go
- [X] T013 [P] [US1] Create gold karat and fuel petroleum domain models in internal/domain/commodities.go
- [X] T014 [P] [US1] Create domestic EGX and regional GCC equity domain models in internal/domain/stocks.go
- [X] T015 [P] [US1] Create benchmark index domain models in internal/domain/indices.go
- [X] T016 [US1] Implement LiveSnapshotCoordinator with RWMutex store, singleflight suppression, and async disk persistence in internal/engine/live_snapshot.go
- [X] T017 [US1] Implement RatesProcessor to map raw Firestore bank rate documents into enriched CurrencyRate structures in internal/adapter/firestore/rates_processor.go
- [X] T018 [US1] Implement FirestoreListenerService managing worker goroutines for rates collection and latest quote documents in internal/adapter/firestore/listener.go
- [X] T019 [P] [US1] Implement HTTP handler for currency rates (/v1/rates, /v1/rates/:currency, /v1/banks/:bankCode/rates, /v1/catalog) in internal/adapter/http/rates_handler.go
- [X] T020 [P] [US1] Implement HTTP handler for precious metals and fuel (/v1/gold, /v1/gas) in internal/adapter/http/commodities_handler.go
- [X] T021 [P] [US1] Implement HTTP handler for EGX and GCC equities (/v1/stocks, /v1/stocks/:market, /v1/stocks/regional/catalog) in internal/adapter/http/stock_handler.go
- [X] T022 [P] [US1] Implement HTTP handler for market benchmark indices (/v1/indices, /v1/indices/:indexCode/stocks, /v1/indices/:market) in internal/adapter/http/indices_handler.go
- [X] T023 [US1] Wire LiveSnapshotCoordinator, Firestore listeners, and live quote handlers into Chi router in internal/adapter/http/router.go

**Checkpoint**: User Story 1 (MVP) is fully functional - all real-time market quote endpoints respond directly from memory with zero database reads.

---

## Phase 4: User Story 2 - Authoritative Historical Archives & Range Slicing (Priority: P2)

**Goal**: Query historical single-date daily archives and multi-day date ranges across calendar weeks, resolving active mutating weeks from RAM Tier 1 and past sealed weeks from persistent disk Tier 2 (30-day TTL) with zero upstream database reads during warmed serving.

**Independent Test**: Execute single-date and range queries (`/v1/rates/history/:date`, `/v1/rates/history/range`, `/v1/stocks/history/:date`, `/v1/stocks/:symbol/history/range`, `/v1/gold/history/range`, `/v1/indices/history/range`) for past sealed dates and current active week dates, verifying exact JSON envelopes and sub-15ms response latency.

### Tests for User Story 2

- [X] T024 [P] [US2] Unit test HistoryResolutionEngine week partitioning, Tier 1/Tier 2 lookup, and range assembly in internal/engine/history_resolution_test.go
- [X] T025 [P] [US2] Integration test historical date and range endpoints in internal/adapter/http/history_test.go

### Implementation for User Story 2

- [X] T026 [P] [US2] Create historical weekly bucket, archive entry, and range query models in internal/domain/history.go
- [X] T027 [US2] Implement HistoryResolutionEngine managing Tier 1 RAM active weeks, Tier 2 disk 30-day archives, and range slicing in internal/engine/history_resolution.go
- [X] T028 [US2] Extend FirestoreListenerService to subscribe to active *_history_weekly documents in internal/adapter/firestore/listener.go
- [X] T029 [P] [US2] Implement historical endpoints for currency rates (/v1/rates/history/:date, /v1/rates/history/range) in internal/adapter/http/rates_handler.go
- [X] T030 [P] [US2] Implement historical endpoints for commodities (/v1/gold/history/:date, /v1/gold/history/range) in internal/adapter/http/commodities_handler.go
- [X] T031 [P] [US2] Implement historical endpoints for stocks (/v1/stocks/history/:date, /v1/stocks/:symbol/history/range, /v1/stocks/:market/history/:date, /v1/stocks/:market/:symbol/history/range) in internal/adapter/http/stock_handler.go
- [X] T032 [P] [US2] Implement historical endpoints for benchmark indices (/v1/indices/history/:date, /v1/indices/history/range) in internal/adapter/http/indices_handler.go
- [X] T033 [US2] Wire HistoryResolutionEngine and historical query endpoints into Chi router in internal/adapter/http/router.go

**Checkpoint**: User Stories 1 and 2 are fully functional - both real-time and historical multi-day range queries resolve from memory and disk with zero database queries.

---

## Phase 5: User Story 3 - High-Throughput Batch Quotes & Intraday Timelines (Priority: P3)

**Goal**: Support heterogeneous multi-asset batch quotes (`POST /v1/quotes` and `GET /v1/quotes?items=...`) with item-level freshness markers, and provide intraday timeline sparklines with automatic fallback to the previous completed trading session during off-hours.

**Independent Test**: Execute batch queries requesting mixed asset codes (`CBE_USD,GOLD_21K,EGX_COMI,TASI`) and intraday timeline routes (`/v1/stocks/:symbol/intraday/latest`, `/v1/rates/:currency/intraday/latest`, `/v1/gold/intraday/latest`), confirming item-level status envelopes and graceful off-hours session fallback.

### Tests for User Story 3

- [X] T034 [P] [US3] Unit test IntradayTimelineStore caching and off-hours previous session backtracking in internal/engine/intraday_test.go
- [X] T035 [P] [US3] Unit test batch quote resolution and partial degradation handling in internal/adapter/http/quotes_handler_test.go

### Implementation for User Story 3

- [X] T036 [P] [US3] Create batch quote request, response, and item envelope models in internal/domain/quotes.go
- [X] T037 [P] [US3] Create intraday tick point and timeline series domain models in internal/domain/intraday.go
- [X] T038 [US3] Implement IntradayTimelineStore with thread-safe point maps and previous session backtracking in internal/engine/intraday.go
- [X] T039 [US3] Extend FirestoreListenerService to subscribe to active *_intraday documents in internal/adapter/firestore/listener.go
- [X] T040 [US3] Implement batch quotes HTTP handler supporting POST and GET queries in internal/adapter/http/quotes_handler.go
- [X] T041 [P] [US3] Implement intraday timeline endpoints for rates (/v1/rates/:currency/intraday/latest) in internal/adapter/http/rates_handler.go
- [X] T042 [P] [US3] Implement intraday timeline endpoints for gold (/v1/gold/intraday/latest, /v1/gold/:karat/intraday/latest) in internal/adapter/http/commodities_handler.go
- [X] T043 [P] [US3] Implement intraday timeline endpoints for stocks (/v1/stocks/:symbol/intraday/latest) in internal/adapter/http/stock_handler.go
- [X] T044 [US3] Wire batch quotes and intraday handlers into Chi router in internal/adapter/http/router.go

**Checkpoint**: User Stories 1, 2, and 3 are operational - batch quote resolution and intraday timeline back-tracking operate seamlessly.

---

## Phase 6: User Story 4 - Gateway Runtime, Security, & Operational Telemetry (Priority: P4)

**Goal**: Enforce API key authentication (`x-api-key` header and `?apiKey=` query), deterministic HTTP ETag revalidation (`304 Not Modified`), request tracing (`x-trace-id`), operational telemetry metrics, and scheduled Cairo rollovers (midnight week seal and 08:00 AM intraday rotate).

**Independent Test**: Execute unauthorized requests to verify HTTP `401`, send matching `If-None-Match` headers to verify `304 Not Modified`, probe `/v1/health`, `/v1/version`, and `/v1/metrics`, and trigger scheduled rollovers without service interruption.

### Tests for User Story 4

- [X] T045 [P] [US4] Unit test authentication, ETag hashing, and trace ID middleware in internal/adapter/http/middleware_test.go
- [X] T046 [P] [US4] Unit test Cairo midnight and 08:00 AM cron rollover actions in internal/engine/scheduler_test.go

### Implementation for User Story 4

- [X] T047 [US4] Implement authentication, ETag generation, tracing, and compression middleware in internal/adapter/http/middleware.go
- [X] T048 [P] [US4] Implement health check handler (/v1/health) in internal/adapter/http/health_handler.go
- [X] T049 [P] [US4] Implement version discovery handler (/v1/version, /version) reporting 2.18.20 in internal/adapter/http/version_handler.go
- [X] T050 [P] [US4] Implement operational metrics tracking handler (/v1/metrics) in internal/adapter/http/metrics_handler.go
- [X] T051 [US4] Implement timezone-aware cron scheduler for midnight week sealing and 08:00 AM intraday rotation in internal/engine/scheduler.go
- [X] T052 [US4] Implement application bootstrap, configuration binding, graceful shutdown, and signal trapping in cmd/server/main.go

**Checkpoint**: All 49 platform routes are fully assembled with end-to-end security, caching headers, telemetry, and background scheduling.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Verification against platform gates, line-by-line comment audit, race detection, and performance benchmarking.

- [X] T053 Execute Go race detector suite across all packages (go test -v -race -cover ./...)
- [X] T054 Audit all Go source files to verify 100% line comment coverage per Constitution Principle V
- [X] T055 [P] Implement PowerShell verification test runner for Go gateway in scripts/verify/verify_go.ps1
- [X] T056 Execute full 49-scenario verification gate (npm run verify) against running Go server and ensure exit code 0
- [X] T057 Execute 80-test differential parity comparison (npm run verify:report) against base server and confirm 100% parity
- [X] T058 Validate P99 response latency is strictly under 15ms and update quickstart documentation in specs/116-migrate-gateway-go/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories.
- **User Stories (Phase 3-6)**:
  - **User Story 1 (P1 🎯 MVP)**: Depends on Foundational (Phase 2).
  - **User Story 2 (P2)**: Depends on Foundational (Phase 2) and US1 domain models.
  - **User Story 3 (P3)**: Depends on Foundational (Phase 2) and US1 live engines.
  - **User Story 4 (P4)**: Depends on Foundational (Phase 2) and integrates middleware across all handlers.
- **Polish (Phase 7)**: Depends on completion of all desired user stories.

### User Story Dependencies

```
┌────────────────────────────────────────────────────────┐
│ Phase 1: Setup (Go Module, Config, Error Models)       │
└──────────────────────────┬─────────────────────────────┘
                           │
┌──────────────────────────▼─────────────────────────────┐
│ Phase 2: Foundational (Catalog, DiskAdapter, Schedule) │
└──────────────────────────┬─────────────────────────────┘
                           │
       ┌───────────────────┴───────────────────┐
       ▼                                       ▼
┌───────────────────────────────┐   ┌───────────────────────────────┐
│ Phase 3: User Story 1 (P1)    │   │ Phase 6: User Story 4 (P4)    │
│ Real-Time Push Quotes (MVP)   │   │ Security, Telemetry & Cron    │
└──────────────┬────────────────┘   └──────────────┬────────────────┘
               │                                   │
       ┌───────┴───────┐                           │
       ▼               ▼                           │
┌──────────────┐ ┌──────────────┐                  │
│ Phase 4: US2 │ │ Phase 5: US3 │                  │
│ Historical   │ │ Batch Quotes │                  │
│ & Ranges     │ │ & Intraday   │                  │
└──────┬───────┘ └──────┬───────┘                  │
       │                │                          │
       └───────┬────────┴──────────────────────────┘
               │
┌──────────────▼────────────────────────────────────────┐
│ Phase 7: Polish & Verification (npm run verify)        │
└────────────────────────────────────────────────────────┘
```

---

## Parallel Opportunities

- **Phase 1**: `T002`, `T003`, `T004` can run concurrently once `T001` initializes the module.
- **Phase 2**: `T005` (catalog) and `T006` (disk adapter) can run concurrently. Tests `T008` and `T009` can run in parallel.
- **Phase 3 (US1)**:
  - Unit tests `T010` and `T011` can be authored concurrently.
  - Domain models `T012`, `T013`, `T014`, `T015` can be created in parallel.
  - Handlers `T019`, `T020`, `T021`, `T022` can be implemented in parallel once `T016` (coordinator) is in place.
- **Phase 4 (US2)**: Handlers `T029`, `T030`, `T031`, `T032` can be implemented in parallel once `T027` (engine) is complete.
- **Phase 5 (US3)**: Models `T036` and `T037` can be created in parallel; handlers `T041`, `T042`, `T043` can be created in parallel.
- **Phase 6 (US4)**: Handlers `T048`, `T049`, `T050` can be created concurrently with middleware `T047`.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (`T001`-`T004`).
2. Complete Phase 2: Foundational (`T005`-`T009`).
3. Complete Phase 3: User Story 1 (`T010`-`T023`).
4. **STOP and VALIDATE**: Verify live quote endpoints (`/v1/rates`, `/v1/gold`, `/v1/gas`, `/v1/stocks`, `/v1/indices`) respond in < 1ms from memory.

### Incremental Delivery

1. Setup + Foundational $\rightarrow$ Core models and persistence ready.
2. Add User Story 1 $\rightarrow$ Test real-time push quotes independently $\rightarrow$ MVP Achieved.
3. Add User Story 2 $\rightarrow$ Historical multi-tier range resolution operational.
4. Add User Story 3 $\rightarrow$ Batch quotes and intraday timeline sparklines operational.
5. Add User Story 4 $\rightarrow$ Security, caching headers, background rollovers, and telemetry wired.
6. Phase 7: Polish & Mandatory Verification $\rightarrow$ `npm run verify` exits with code `0`.
