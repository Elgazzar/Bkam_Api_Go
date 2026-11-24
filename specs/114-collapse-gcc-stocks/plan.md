# Implementation Plan: Collapse Artificial GCC Stocks Boundary into Stock Module

**Branch**: `114-collapse-gcc-stocks` | **Date**: 2026-11-15 | **Spec**: [specs/114-collapse-gcc-stocks/spec.md](spec.md)

**Input**: Feature specification from `specs/114-collapse-gcc-stocks/spec.md`

---

## Summary

Eliminate the artificial `src/modules/gcc-stocks/` sub-module (>1,400 lines of duplicate queries, routes, and dead scraper daemons) by absorbing regional equity queries and routes directly into the authoritative `StockQueryService` and `stock-routes.ts`. Reaps direct leverage from existing deep core engines (`MarketCatalog`, `LiveSnapshotCoordinator`, `HistoryResolutionEngine`, and `IntradayCoordinator`), removes dynamic string imports, purges non-compliant scraper daemons per Constitution Principle VIII, and guarantees 100% backward compatibility for all existing regional endpoints.

---

## Technical Context

**Language/Version**: TypeScript 5.8+ / Node.js 20+ (ESM modules)  
**Primary Dependencies**: Fastify 5.x, Zod 3.x, `@fastify/compress`, `@fastify/etag`  
**Storage**: In-memory RAM caches (`LiveSnapshotCoordinator`), persistent disk cache (`flat-cache`), Firestore collections (`stock_prices_snapshot`, `gcc_stock_prices_*`) via push listeners  
**Testing**: Vitest 4.x (`npm run test:unit`), Contract tests (`npm run test:contract`), Canonical verification (`npm run verify`), Differential parity (`npm run verify:diff`)  
**Target Platform**: Linux ARM64 (Oracle Cloud Infrastructure Ampere A1)  
**Project Type**: Fastify REST API Gateway  
**Performance Goals**: <150ms RAM latency ceiling for all warmed multi-market stock routes  
**Constraints**: 100% backward-compatible HTTP contracts, Zero direct Firestore reads during active serving (Principle II), 100% line comment coverage (Principle V), Zero scraper daemon code (Principle VIII)  
**Scale/Scope**: 5 equity exchanges (EGX, Tadawul, DFM/ADX, QSE, Boursa Kuwait) covering 1,180+ securities  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Compliance Status | Justification / Architectural Verification |
| :--- | :--- | :--- |
| **I. API-First Contract** | **PASS** | 100% backward-compatible route aliases (`/v1/stocks/saudi`, etc.) preserved. Payloads match existing Zod schemas byte-for-byte. |
| **II. Zero-Read Caching** | **PASS** | All regional stock queries served from `LiveSnapshotCoordinator` in-memory partitions or `flat-cache` historical buckets. 0 direct Firestore reads. |
| **III. Freshness & Fallbacks** | **PASS** | Snapshots contain `freshness`, `sourceTraceId`, `updatedAt`, and graceful cached fallback on network disruption. |
| **IV. Performance & SLA** | **PASS** | All stock queries resolve in <150ms (typical <1ms in-memory) with Brotli/Gzip compression and ETag revalidation. |
| **V. Exhaustive Commenting** | **PASS** | 100% line comment coverage mandated for all modified and added code lines. |
| **VI. Cloud Budget Safeguards** | **PASS** | Zero incremental hosting cost; zero Firestore read billing consumption. |
| **VII. Mandatory Testing** | **PASS** | Verification requires exit code `0` on `npm run verify` (37/37) and `npm run verify:diff` (31/31). |
| **VIII. Independent Deployment** | **PASS** | Purges all dead scraper daemon adapters (`TradingViewSaudiAdapter`, etc.) from the API repository. |
| **IX. Mandatory Version Incrementing** | **PASS** | Application version will be incremented from `2.18.10` to `2.18.11` in `package.json` upon completion. |

*Gate Evaluation: ALL 9 PRINCIPLES PASS WITH ZERO VIOLATIONS.*

---

## Project Structure

### Documentation (this feature)

```text
specs/114-collapse-gcc-stocks/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan (this file)
├── research.md          # Phase 0 architectural decisions
├── data-model.md        # Phase 1 data entities and Zod schemas
├── quickstart.md        # Phase 1 validation and execution scenarios
├── contracts/
│   └── stock-api.md     # Phase 1 REST contract definitions
└── checklists/
    └── requirements.md  # Specification quality checklist
```

### Source Code Migration Layout

```text
src/
├── config/
│   ├── catalogs/                     # Relocated regional stock catalogs
│   │   ├── saudi-catalog.ts          # Relocated from src/modules/gcc-stocks/catalogs/
│   │   ├── uae-catalog.ts            # Relocated from src/modules/gcc-stocks/catalogs/
│   │   ├── qatar-catalog.ts          # Relocated from src/modules/gcc-stocks/catalogs/
│   │   └── kuwait-catalog.ts         # Relocated from src/modules/gcc-stocks/catalogs/
│   ├── stock-catalog.ts              # Existing EGX stock definitions
│   └── index-catalog.ts
├── routes/
│   └── stock-routes.ts               # Absorbs all regional routes and legacy aliases
├── services/
│   ├── stock-query-service.ts        # Deepened: handles all 5 exchanges via core engines
│   ├── catalog-service.ts            # MarketCatalog updating import paths
│   ├── live-snapshot-coordinator.ts  # Native multi-market quote storage
│   ├── history-resolution-engine.ts  # Native multi-market history
│   └── intraday-coordinator.ts       # Native multi-market intraday
└── app.ts                            # Purges dynamic module loader (lines 172-220)

DELETED DIRECTORIES:
└── src/modules/gcc-stocks/           # Completely eliminated (>1,400 lines deleted)
```

**Structure Decision**:
Eliminates `src/modules/gcc-stocks/` entirely. Security definitions move to `src/config/catalogs/`. All query logic concentrates in `src/services/stock-query-service.ts`. All routing concentrates in `src/routes/stock-routes.ts`.

---

## Complexity Tracking

> **Constitution Check has zero violations. No exceptions required.**

| Violation | Why Needed | Simpler Alternative Rejected Because |
| :--- | :--- | :--- |
| *None* | *None* | *None* |
