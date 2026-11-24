# Implementation Plan: Deepen Intraday Coordinator

**Branch**: `107-deepen-intraday-coordinator` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/107-deepen-intraday-coordinator/spec.md`

---

## Summary

Deepen the existing `IntradayCoordinator` (`src/services/intraday-coordinator.ts`) to absorb quote extraction, previous-price baselining (Feature 100), off-hours session backtracking, and response envelope formatting behind a single, high-leverage interface method (`getLatestIntraday`). This eliminates over 600 lines of duplicate branching and baselining boilerplate across 4 caller query modules (`RatesQueryService`, `GoldQueryService`, `StockQueryService`, and `IndicesQueryService`), concentrating intraday session lifecycles into one deep module while preserving 100% backwards-compatibility and sub-50ms latency.

---

## Technical Context

**Language/Version**: TypeScript 5.8 / Node.js 22 LTS  
**Primary Dependencies**: Fastify 5.x, Zod 3.24, @google-cloud/firestore 8.6, flat-cache 6.1  
**Storage**: Volatile in-memory store (`intradaySnapshotService` in RAM) + Firestore descending query fallback (`limit(1)`) + `HistoryResolutionEngine`  
**Testing**: Vitest 4.1 (unit/integration) + custom modular verification suite (`npm run verify:remote`, `npm run verify:diff`)  
**Target Platform**: Linux server (Node.js 22 Alpine in Docker container, ARM64 / AMD64)  
**Project Type**: Standalone REST API Gateway (`Bkam_Api`)  
**Performance Goals**: Sub-50ms latency for active today queries (100% from RAM); 0 on-demand database reads during active sessions  
**Constraints**: Zero breaking changes to public REST endpoints; strict Always Free tier compliance ($0 cloud budget); exhaustive comments on every modified line (Constitution Principle V)  
**Scale/Scope**: 4 financial asset classes (FX, Gold, Equities, Indices) across 5 markets (Egypt, Saudi Arabia, UAE, Qatar, Kuwait)  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Living Requirements Review (Principle XVI)**: Read `.specify/memory/REQUIREMENTS.md`; verified that deepening `IntradayCoordinator` fulfills FR-API-009, FR-API-010, and FR-API-013 with zero regressions to existing requirements.
- [x] **API-First Contract (Principle I)**: Request and response shapes for `/v1/rates/intraday/latest`, `/v1/gold/intraday/latest`, `/v1/stocks/:symbol/intraday/latest`, and `/v1/indices/:code/intraday/latest` remain 100% identical; no breaking contract changes.
- [x] **Cloud Cost Safeguards (Principle VII)**: Active session queries continue to read 100% from volatile memory with zero on-demand Firestore reads ($0 budget preserved).
- [x] **Exhaustive Commenting (Principle V)**: Every single line of newly authored or modified source code will be explicitly commented explaining its rationale, logic, and runtime behavior.
- [x] **E2E & Parity Testing (Principle XII & XIII)**: All 36 verification scenarios in `verify:remote` and all 30 differential comparison endpoints in `verify:diff` must pass 100% green before merge.
- [x] **Intelligent Caching & Intraday Lifecycle (Principle XIII)**: Dynamic active sessions return `Cache-Control: public, max-age=60`, while concluded backtracked sessions return `Cache-Control: public, max-age=3600`. Intraday tick points strictly reside in RAM.
- [x] **Documentation Sync (Principles X, XI, XV, XVI)**: Plans include updating `ARCHITECTURE.md`, `README.md`, `MODULAR_COMPONENTS.md`, and `REQUIREMENTS.md`.
- [x] **Bug Tracking & Ledger (Principle XVII)**: Any issues encountered during implementation will be logged in `specs/107-deepen-intraday-coordinator/bugs.md`.

---

## Project Structure

### Documentation (this feature)

```text
specs/107-deepen-intraday-coordinator/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit-plan output)
├── research.md          # Architectural research & technical decisions
├── data-model.md        # Entities, contracts, and lifecycle diagrams
├── quickstart.md        # Runnable verification guide
├── contracts/           # Interface contracts
│   ├── intraday-coordinator.contract.ts
│   └── intraday-envelope.contract.ts
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (`D:\Repo\Bkam\Bkam_Api\src`)

```text
src/
├── services/
│   ├── intraday-coordinator.ts         # [DEEPEN] Absorbs quote synthesis, session backtracking & baselines
│   ├── intraday-types.ts               # Adds LatestIntradayOptions & LatestIntradayResult interfaces
│   ├── rates-query-service.ts          # [SIMPLIFY] Delegates fetchLatestRatesIntraday to coordinator
│   ├── gold-query-service.ts           # [SIMPLIFY] Delegates fetchLatestGoldIntraday to coordinator
│   ├── stock-query-service.ts          # [SIMPLIFY] Delegates fetchLatestStockIntraday to coordinator
│   ├── indices-query-service.ts        # [SIMPLIFY] Delegates fetchLatestIndexIntraday to coordinator
│   ├── history-resolution-engine.ts    # Provides historical daily snapshots for concluded sessions
│   └── intraday-snapshot-service.ts    # Volatile in-memory store for ongoing session tick points
└── routes/
    ├── rates-routes.ts                 # Consumes coordinator caching metadata for headers
    ├── gold-routes.ts                  # Consumes coordinator caching metadata for headers
    ├── stock-routes.ts                 # Consumes coordinator caching metadata for headers
    └── indices-routes.ts               # Consumes coordinator caching metadata for headers
```

**Structure Decision**: Monolithic in-process refactoring inside `D:\Repo\Bkam\Bkam_Api\src\services`. No new packages, services, or runtime dependencies introduced.

---

## Complexity Tracking

*No violations or unjustified complexity introduced. All principles satisfied.*
