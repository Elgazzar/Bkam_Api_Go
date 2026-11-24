# Implementation Plan: Live Snapshot Coordinator

**Branch**: `109-live-snapshot-coordinator` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/109-live-snapshot-coordinator/spec.md`

---

## Summary

Consolidate fragmented live quote storage and push ingestion into a single, high-leverage `LiveSnapshotCoordinator` module (`src/services/live-snapshot-coordinator.ts`). The coordinator replaces ad-hoc `cachedPrices: T | null` slots, duplicate fallback Firestore queries (`doc("latest").get()`), and redundant error logging across 6 separate query services (`RatesSnapshotService`, `GoldQueryService`, `GasQueryService`, `StockQueryService`, `IndicesQueryService`, `GccStockQueryService`) behind an expressive 3-method interface (`getSnapshot`, `updateSnapshot`, `fetchFresh`). It encapsulates atomic in-memory multi-domain RAM storage, promise coalescing against thundering herds, disk persistence for zero-latency cold boots, and fail-safe degraded execution.

---

## Technical Context

**Language/Version**: TypeScript 5.8 / Node.js 22 LTS  
**Primary Dependencies**: Fastify 5.x, Zod 3.24, @google-cloud/firestore 8.6, flat-cache 6.1  
**Storage**: In-memory multi-domain RAM Map (`LiveSnapshotCoordinator` singleton), backed by local JSON disk files (`var/snapshots/`) for cold-start warming  
**Testing**: Vitest 4.1 (`npm run test:unit`), contract tests (`npm run test:contract`)  
**Target Platform**: Linux server (Node.js 22 Alpine in Docker container, ARM64 / AMD64 on OCI Ampere A1)  
**Project Type**: Standalone REST API Gateway (`Bkam_Api`)  
**Performance Goals**: Sub-0.01ms snapshot retrieval from RAM (100% compliance with SC-001 and Zero-Read Principle II)  
**Constraints**: 100% backwards-compatible REST endpoints; zero external database reads during active serving conditions; exhaustive commenting on every newly authored or modified line (Constitution Principle V)  
**Scale/Scope**: 6 asset domains (`rates`, `gold`, `gas`, `stocks`, `indices`, `gcc-stocks`), ~13 composite domain keys, <15MB RAM footprint  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **API-First Contract (Principle I)**: All existing REST endpoints (`/v1/rates`, `/v1/gold`, `/v1/gas`, `/v1/stocks`, `/v1/indices`, `/v1/quotes`) maintain 100% contract schema stability. Documented in [coordinator-routes.contract.md](./contracts/coordinator-routes.contract.md).
- [x] **Zero-Read Memory Caching (Principle II)**: All active quote snapshots reside in RAM, populated strictly via real-time WebSocket listeners (`onSnapshot`) with zero Firestore reads during active request serving.
- [x] **Freshness, Traceability, and Fallbacks (Principle III)**: Snapshots retain trace IDs, official source timestamps, and dynamic freshness markers (`current`, `stale`, `degraded`) when degraded fail-safe fallbacks are served.
- [x] **Performance and SLA Guarantees (Principle IV)**: In-memory hash map lookup executes in <0.01ms, easily meeting the p95 <200ms gateway latency SLA.
- [x] **Exhaustive Code Commenting (Principle V)**: Every single line of newly authored and refactored code will be commented detailing its logic, rationale, and runtime behavior.
- [x] **Cloud Cost Safeguards (Principle VI)**: Zero cloud infrastructure cost increase; runs strictly within OCI Always Free tier footprint (<15MB heap).
- [x] **Mandatory Testing & Contract Verification (Principle VII)**: Unit tests (`live-snapshot-coordinator.test.ts`), route integration tests, and contract verification must pass 100% green before merge.
- [x] **Independent Deployment Lifecycle (Principle VIII)**: No external scraping dependencies introduced; clean in-process domain refactoring.

---

## Project Structure

### Documentation (this feature)

```text
specs/109-live-snapshot-coordinator/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan (this file)
├── research.md          # Architectural research & technical decisions
├── data-model.md        # Domain entities, key mapping & lifecycle
├── quickstart.md        # Runnable verification guide
├── contracts/           # Interface contracts
│   ├── live-snapshot-coordinator.contract.ts
│   └── coordinator-routes.contract.md
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (`D:\Repo\Bkam\Bkam_Api\src`)

```text
src/
├── services/
│   ├── live-snapshot-coordinator.ts    # [NEW] Deep LiveSnapshotCoordinator module
│   ├── firestore-listener-service.ts   # [SIMPLIFY] Pushes all updates to liveSnapshotCoordinator
│   ├── quotes-service.ts               # [SIMPLIFY] Resolves all quotes from liveSnapshotCoordinator
│   ├── rates-snapshot-service.ts       # [SIMPLIFY] Delegates memory caching to coordinator
│   ├── gold-query-service.ts           # [SIMPLIFY] Delegates live caching & fetch to coordinator
│   ├── gas-query-service.ts            # [SIMPLIFY] Delegates live caching & fetch to coordinator
│   ├── stock-query-service.ts          # [SIMPLIFY] Delegates live caching & fetch to coordinator
│   ├── indices-query-service.ts        # [SIMPLIFY] Delegates live caching & fetch to coordinator
│   └── gcc-stocks/services/
│       └── gcc-stock-query-service.ts  # [SIMPLIFY] Delegates GCC live caching to coordinator
├── cache/
│   ├── snapshot-cache.ts               # [RETIRE] Absorbed by LiveSnapshotCoordinator
│   └── snapshot-store.ts               # [RETIRE] Absorbed by LiveSnapshotCoordinator
└── app.ts                              # [WIRE] Register LiveSnapshotCoordinator singleton
```

**Structure Decision**: In-process deepening inside `src/services/live-snapshot-coordinator.ts`. The legacy shallow cache wrappers in `src/cache/snapshot-cache.ts` and `snapshot-store.ts` are retired/deprecated, while domain query services retain their existing external interfaces as thin delegations to preserve 100% backward compatibility.

---

## Complexity Tracking

*No violations or unjustified complexity introduced. All principles satisfied.*
