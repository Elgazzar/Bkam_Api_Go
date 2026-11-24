# Implementation Plan: Deepen History Resolution Engine

**Branch**: `106-deepen-history-resolution` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/106-deepen-history-resolution/spec.md`

---

## Summary

Deepen the existing `HistoryResolutionEngine` (`src/services/history-resolution-engine.ts`) to absorb active trading day synthesis, Africa/Cairo timezone calculations, intraday timeline coordination, and multi-day live-range stitching behind a single, high-leverage interface (`getDateSnapshot` and `getDateRange`). This eliminates over 350 lines of duplicate branching boilerplate across 4 caller query services (`RatesQueryService`, `GoldQueryService`, `StockQueryService`, and `IndicesQueryService`), concentrating date-boundary synthesis into one deep module while preserving 100% backwards-compatibility and sub-200ms latency.

---

## Technical Context

**Language/Version**: TypeScript 5.8 / Node.js 22 LTS  
**Primary Dependencies**: Fastify 5.x, Zod 3.24, @google-cloud/firestore 8.6, flat-cache 6.1  
**Storage**: Tiered caching hierarchy (Active Week RAM store -> Flat-Cache disk storage -> Firestore weekly buckets `rates_history_weekly/YYYY-Www`)  
**Testing**: Vitest 4.1 (unit/integration) + custom modular verification suite (`tsx scripts/verify/index.ts`, `tsx scripts/verify/diff.ts`)  
**Target Platform**: Linux server (Node.js 22 Alpine in Docker container, ARM64 / AMD64)  
**Project Type**: Standalone REST API Gateway (`Bkam_Api`)  
**Performance Goals**: Sub-200ms p95 latency for in-memory and cached responses; 0 database reads for active today queries  
**Constraints**: Zero breaking changes to external REST endpoints; strict Always Free tier compliance ($0 cloud budget); exhaustive comments on every modified line (Constitution Principle V)  
**Scale/Scope**: 4 financial asset classes (FX, Gold, Equities, Indices) across 5 markets (Egypt, Saudi Arabia, UAE, Qatar, Kuwait)  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Living Requirements Review (Principle XVI)**: Read `.specify/memory/REQUIREMENTS.md`; verified that deepening `HistoryResolutionEngine` fulfills FR-API-008, FR-API-009, and FR-API-012 with zero regressions to existing requirements.
- [x] **API-First Contract (Principle I)**: Request and response shapes for `/v1/rates/history/*`, `/v1/gold/history/*`, `/v1/stocks/history/*`, and `/v1/indices/history/*` remain 100% identical; no breaking contract changes.
- [x] **Cloud Cost Safeguards (Principle VII)**: Active session queries continue to read 100% from in-memory snapshots and `intradayCoordinator` with zero on-demand Firestore reads ($0 budget preserved).
- [x] **Exhaustive Commenting (Principle V)**: Every single line of newly authored or modified source code will be explicitly commented explaining its rationale, logic, and runtime behavior.
- [x] **E2E & Parity Testing (Principle XII & XIII)**: All 36 verification scenarios in `verify:remote` and all 30 differential comparison endpoints in `verify:diff` must pass 100% green before merge.
- [x] **Intelligent Caching & Intraday Lifecycle (Principle XIII)**: Dynamic active sessions return `Cache-Control: public, max-age=60`, while sealed past dates return `Cache-Control: public, max-age=3600`.
- [x] **Documentation Sync (Principles X, XI, XV, XVI)**: Plans include updating `ARCHITECTURE.md`, `README.md`, `MODULAR_COMPONENTS.md`, and `REQUIREMENTS.md`.
- [x] **Bug Tracking & Ledger (Principle XVII)**: Any issues encountered during implementation will be logged in `specs/106-deepen-history-resolution/bugs.md`.

---

## Project Structure

### Documentation (this feature)

```text
specs/106-deepen-history-resolution/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit-plan output)
├── research.md          # Architectural research & technical decisions
├── data-model.md        # Entities, contracts, and lifecycle diagrams
├── quickstart.md        # Runnable verification guide
├── contracts/           # Interface contracts
│   ├── history-resolution-engine.contract.ts
│   └── query-service.contract.ts
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (`D:\Repo\Bkam\Bkam_Api\src`)

```text
src/
├── services/
│   ├── history-resolution-engine.ts    # [DEEPEN] Absorbs isToday check, live RAM synthesis, & range stitching
│   ├── rates-query-service.ts          # [SIMPLIFY] Delegates getRatesHistory & range to engine
│   ├── gold-query-service.ts           # [SIMPLIFY] Delegates fetchGoldPriceHistory & range to engine
│   ├── stock-query-service.ts          # [SIMPLIFY] Delegates fetchStockPriceHistory & range to engine
│   ├── indices-query-service.ts        # [SIMPLIFY] Delegates fetchIndexPriceHistory & range to engine
│   ├── intraday-coordinator.ts         # Coordinates volatile intra-session tick points
│   └── active-weekly-bucket-store.ts   # In-memory store for ongoing calendar week
└── routes/
    ├── rates-routes.ts                 # Uses engine cache metadata for response headers
    ├── gold-routes.ts                  # Uses engine cache metadata for response headers
    ├── stock-routes.ts                 # Uses engine cache metadata for response headers
    └── indices-routes.ts               # Uses engine cache metadata for response headers
```

**Structure Decision**: Monolithic in-process refactoring inside `D:\Repo\Bkam\Bkam_Api\src\services`. No new packages, services, or runtime dependencies introduced.

---

## Complexity Tracking

*No violations or unjustified complexity introduced. All principles satisfied.*
