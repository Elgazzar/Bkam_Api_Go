# Implementation Plan: Unified History Resolution Engine

**Branch**: `096-unified-history-engine` | **Date**: 2026-10-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/096-unified-history-engine/spec.md` and ADR 0006 (`docs/adr/0006-unified-history-resolution-engine.md`).

## Summary

Implement a centralized, modular `HistoryResolutionEngine` in `api/src/services/history-resolution-engine.ts` that unifies historical market quote retrieval across all asset classes (Bank Rates, Gold, EGX Stocks, EGX Indices, and GCC regional stocks). The engine integrates directly with the existing `CacheService` layer (`ICacheProvider` / `FlatCacheProvider`), synchronizes `rates_history_weekly` in volatile RAM via `FirestoreListenerService`, decouples live Intraday Timelines from historical archives, gives Gold first-class single-date resolution parity, and eliminates combinatorial cache key explosion for Bank Rates ranges by directly slicing Weekly Buckets in memory and disk cache.

## Technical Context

**Language/Version**: TypeScript 5.x / Node.js >= 20.0.0 (ESM modules)

**Primary Dependencies**: Fastify 4.x, `@google-cloud/firestore`, `flat-cache` (via `CacheService`), `zod`

**Storage**:
- Volatile RAM (Tier 1): `ActiveWeeklyBucketStore` (active calendar week), `intradaySnapshotService` (today's active ticks), `missingDatesCache` (negative lookup cache)
- Persistent Disk (Tier 2 & Tier 3): `CacheService` (`FlatCacheProvider` writing to `api/var/cache/` or `.cache/`)
- Cloud Database: Google Cloud Firestore (`*_history_weekly/{YYYY-Www}` collections)

**Testing**: Vitest (`npm run test:unit`, `npm run test:e2e`)

**Target Platform**: Linux (Oracle Cloud Infrastructure Ampere A1 ARM compute instance & local dev)

**Project Type**: Web Service API (Fastify REST backend)

**Performance Goals**:
- Active week historical queries: `<1ms` from RAM with 0 database reads
- Sealed past week queries: `<5ms` from `CacheService` disk cache
- 180-day multi-month range queries: `<50ms` from sliced weekly buckets in disk cache
- Negative cache hits for closed market days: `<0.1ms`

**Constraints**:
- Must strictly operate within $0.00 Always Free Tier limits on Google Cloud Platform and Oracle Cloud (Principle VII)
- Zero raw intraday tick points permitted inside disk caches or historical weekly documents (Principle XIII)
- Every added or modified line of source code must be exhaustively commented (Principle VI)

**Scale/Scope**:
- 5 Asset domains: Bank Rates, Gold, EGX Stocks, EGX Indices, GCC Equities (Saudi, UAE, Qatar)
- Over 90+ financial tickers and 30+ domestic banks

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Living Requirements Review (Principle XVI)**: Read `.specify/memory/REQUIREMENTS.md` before implementation; ensure new features or bug fixes do not regress or conflict with existing system/functional requirements.
- [x] **API-First Contract (Principle I)**: Request/response shapes, versioning, error envelopes, and client compatibility verified. All existing routes remain 100% backward compatible.
- [x] **Cloud Cost Safeguards (Principle VII)**: Firestore reads/writes and compute stay within Always Free tier boundaries ($0 budget). Active week RAM listener prevents read spikes.
- [x] **Exhaustive Commenting (Principle VI)**: Every new or modified line of code will include clear explanatory comments detailing logic, rationale, and runtime expectations.
- [x] **E2E Testing (Principle XII)**: Functional and End-to-End tests will be executed and verified before merge (`npm run test:unit` and `npm run test:e2e`).
- [x] **Intelligent Caching & Intraday Lifecycle (Principle XIII)**: High-frequency intraday tick points are strictly isolated to volatile memory (`intradaySnapshotService` in RAM) and purged upon session rollover; active week is synchronized via push-based `onSnapshot` real-time listener in RAM with zero polling reads; sealed past weeks and historical ranges are cached to `flat-cache` on disk (30-day TTL).
- [x] **Documentation Sync (Principles X, XI, XV, XVI)**: Plans include updating `ARCHITECTURE.md`, `README.md`, `MODULAR_COMPONENTS.md`, and `REQUIREMENTS.md`.
- [x] **Pull Request Exclusivity (Principle XVII)**: Feature branch `096-unified-history-engine` created; all changes will be merged into master exclusively through an approved Pull Request after automated CI/CD and review checks pass.

## Project Structure

### Documentation (this feature)

```text
specs/096-unified-history-engine/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (decisions & trade-offs)
├── data-model.md        # Phase 1 output (entities & cache keys)
├── quickstart.md        # Phase 1 output (validation & run guide)
├── contracts/           # Phase 1 output (engine & API contracts)
│   ├── history-resolution-engine.contract.ts
│   └── api-routes.contract.md
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code Changes

```text
api/
├── src/
│   ├── cache/
│   │   └── cacheService.ts                # Injected into HistoryResolutionEngine
│   ├── services/
│   │   ├── history-resolution-engine.ts   # NEW: Centralized modular history coordinator
│   │   ├── firestore-listener-service.ts  # Add rates_history_weekly to watchedWeeklyCollections
│   │   ├── rates-query-service.ts         # Delegate getRatesHistory/Range to engine
│   │   ├── gold-query-service.ts          # Delegate history & yesterday close to engine
│   │   ├── stock-query-service.ts         # Delegate history to engine
│   │   └── indices-query-service.ts       # Delegate history to engine
│   ├── modules/
│   │   └── gcc-stocks/
│   │       └── services/
│   │           └── gcc-stock-query-service.ts # Delegate regional history to engine
│   └── routes/
│       └── gold-routes.ts                 # Update single-date gold route to call engine directly
└── tests/
    ├── unit/
    │   └── services/
    │       └── history-resolution-engine.test.ts # NEW: Unit tests for engine resolution
    └── e2e/
```

## Complexity Tracking

> *No constitutional violations. All implementations adhere strictly to the established 4-tier caching architecture and modular component standards.*
