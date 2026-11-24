# Implementation Plan: Dedicated Latest Intraday Timeline Endpoints

**Branch**: `097-latest-intraday-api` | **Date**: 2026-10-24 | **Spec**: [specs/097-latest-intraday-api/spec.md](spec.md)

**Input**: Feature specification from `/specs/097-latest-intraday-api/spec.md` and Architectural Decision Record [0007-dedicated-latest-intraday-timeline-endpoints.md](../../docs/adr/0007-dedicated-latest-intraday-timeline-endpoints.md).

## Summary

Implement dedicated RESTful endpoints across all platform asset classes (`/v1/stocks/:symbol/intraday/latest`, `/v1/indices/:symbol/intraday/latest`, `/v1/gold/:item/intraday/latest`, `/v1/rates/:currency/intraday/latest`) to serve the latest trading session's intraday timeline and current quote metrics in a single lightweight round-trip. The solution prioritizes zero-read volatile memory resolution via `IntradaySnapshotService` during active market hours, and provides single-read descending Firestore backtracking (`orderBy(FieldPath.documentId(), 'desc').limit(1)`) during weekends, holidays, and pre-market periods with adaptive HTTP caching (`max-age=60` active vs. `max-age=900` closed).

## Technical Context

**Language/Version**: TypeScript 5.x / Node.js 20+ (ES Modules)  
**Primary Dependencies**: Fastify 4.x, Zod (validation schemas), `@google-cloud/firestore`  
**Storage**: Google Cloud Firestore (`*_intraday` collections), Volatile RAM (`IntradaySnapshotService`), Disk Cache (`cacheService` / `flat-cache`)  
**Testing**: Vitest / Node Test Runner (Unit & Integration tests in `api/tests/unit/` and `api/tests/integration/`)  
**Target Platform**: Linux server on Oracle Cloud Infrastructure (OCI Ampere A1 ARM64 Compute instance)  
**Project Type**: Fastify Web API Service (`api/`)  
**Performance Goals**: <5ms response time for active in-memory session queries, <50ms for off-hours single-read database resolution  
**Constraints**: Zero database reads during active sessions, at most 1 database read per off-hours session rollover, strict Always Free cloud quota compliance ($0 budget)  
**Scale/Scope**: 4 asset domains (Stocks, Indices, Gold, Bank Rates), 5 regional exchanges (EGX, Saudi Tadawul, UAE DFM/ADX, Qatar QSE, Kuwait KSE), supporting ~10,000 daily mobile app requests  

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Living Requirements Review (Principle XVI)**: Read `.specify/memory/REQUIREMENTS.md`; verified `FR-CCH-002` (Zero-Latency Live Intraday Data from in-memory caches) and `FR-SYS-004` (API Key Security Guard). No regressions or conflicts.
- [x] **API-First Contract (Principle I)**: Request and response shapes standardized across all 4 asset domains with backward-compatible JSON contracts and explicit freshness metadata.
- [x] **Cloud Cost Safeguards (Principle VII)**: Active sessions use 0 Firestore reads (in-memory `IntradaySnapshotService`). Off-hours use a single descending query (1 read) buffered in RAM, remaining well within the 50,000 daily free read ceiling.
- [x] **Exhaustive Commenting (Principle VI)**: Every line of modified or newly created source code will feature complete, explicit explanatory comments detailing logic and rationale.
- [x] **E2E Testing (Principle XII)**: Unit and integration tests will validate both active market and off-hours backtracking scenarios before PR review.
- [x] **Intelligent Caching & Intraday Lifecycle (Principle XIII)**: Intraday points are kept in volatile memory and purged/isolated from permanent weekly history buckets; adaptive HTTP headers (`max-age=60` vs. `max-age=900`) protect reverse proxies and origin servers.
- [x] **Documentation Sync (Principles X, XI, XV, XVI)**: Plans include updating `API_ENDPOINTS_GUIDE.md`, `MODULAR_COMPONENTS.md`, `GLOSSARY.md` (completed), and `REQUIREMENTS.md`.
- [x] **Pull Request Exclusivity (Principle XVII)**: Feature branch `097-latest-intraday-api` created; changes will be submitted exclusively via an approved Pull Request.

## Project Structure

### Documentation (this feature)

```text
specs/097-latest-intraday-api/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output: decisions, rationale, alternatives
├── data-model.md        # Phase 1 output: entities, schemas, state transitions
├── quickstart.md        # Phase 1 output: runnable validation scenarios
└── contracts/           # Phase 1 output: interface and route contracts
    ├── stocks-latest-intraday.contract.md
    ├── indices-latest-intraday.contract.md
    ├── gold-latest-intraday.contract.md
    └── rates-latest-intraday.contract.md
```

### Source Code (repository root)

```text
api/
├── src/
│   ├── routes/
│   │   ├── stock-routes.ts         # Add GET /v1/stocks/:symbol/intraday/latest
│   │   ├── indices-routes.ts       # Add GET /v1/indices/:symbol/intraday/latest
│   │   ├── gold-routes.ts          # Add GET /v1/gold/:item/intraday/latest & /v1/gold/intraday/latest
│   │   └── rates-routes.ts         # Add GET /v1/rates/:currency/intraday/latest
│   ├── services/
│   │   ├── intraday-types.ts       # Declare ResolveLatestTimelineOptions & LatestTimelineResult
│   │   ├── intraday-coordinator.ts # Implement resolveLatestTimeline with RAM check & 1-read Firestore backtracking
│   │   ├── intraday-snapshot-service.ts # In-memory buffer for current and previous session snapshots
│   │   ├── stock-query-service.ts  # Add fetchLatestStockIntraday delegate
│   │   ├── indices-query-service.ts# Add fetchLatestIndexIntraday delegate
│   │   ├── gold-query-service.ts   # Add fetchLatestGoldIntraday delegate
│   │   └── rates-query-service.ts  # Add fetchLatestRateIntraday delegate
└── tests/
    ├── unit/
    │   ├── intraday-coordinator.test.ts # Test resolveLatestTimeline (RAM hit vs. Firestore backtracking)
    │   ├── stock-query-service.test.ts  # Test latest intraday stock query
    │   ├── indices-query-service.test.ts# Test latest intraday index query
    │   ├── gold-query-service.test.ts   # Test latest intraday gold query
    │   └── rates-query-service.test.ts  # Test latest intraday rates query
    └── integration/
        └── latest-intraday-routes.test.ts # E2E HTTP integration tests across all 4 endpoints
```

**Structure Decision**: Standard Fastify modular architecture with routes in `api/src/routes/` delegating to domain query services in `api/src/services/`, orchestrating cross-cutting intraday logic through `IntradayCoordinator`.

## Complexity Tracking

> **No violations**. All architecture choices conform strictly to Constitution principles.
