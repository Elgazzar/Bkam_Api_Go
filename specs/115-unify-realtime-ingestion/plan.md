# Implementation Plan: Unify Real-Time Ingestion and Active Weekly Caching

**Branch**: `115-unify-realtime-ingestion` | **Date**: 2026-11-16 | **Spec**: [specs/115-unify-realtime-ingestion/spec.md](spec.md)

**Input**: Feature specification from `specs/115-unify-realtime-ingestion/spec.md`

---

## Summary

Unify real-time Firestore push ingestion and active calendar week historical caching by eliminating artificial intermediate buffering modules and legacy shims. Specifically:
1. Absorb `ActiveWeeklyBucketStore` directly into `HistoryResolutionEngine` as its Tier 1 in-memory store.
2. Refactor `FirestoreListenerService` into a lean Ports & Adapters inbound push listener, dropping 5 unused query service injections and eliminating circular cache updates.
3. Eliminate legacy `SnapshotCache` and `SnapshotStore` shims, injecting `LiveSnapshotCoordinator` directly into `RatesSnapshotService`.
4. Delete dead legacy facade `market-schedule-service.ts`.
5. Maintain 100% backward compatibility for all 49 platform verification scenarios with zero database read increase.

---

## Technical Context

**Language/Version**: TypeScript 5.8+ / Node.js 22+ (ESM modules)  
**Primary Dependencies**: Fastify 5.x, Zod 3.x, `@google-cloud/firestore`, `node-cron`  
**Storage**: In-memory Tier 1 RAM buckets, persistent disk cache (`flat-cache`), Firestore collections via real-time push listeners  
**Testing**: Vitest 4.x (`npm run test:unit`), Contract verification (`npm run verify`), HTML audit suite (`npm run verify:report`)  
**Target Platform**: Linux ARM64 (Oracle Cloud Infrastructure Ampere A1)  
**Project Type**: Fastify REST API Gateway  
**Performance Goals**: <150ms RAM latency ceiling for all warmed routes (sub-millisecond typical)  
**Constraints**: Zero direct Firestore reads during active serving (Principle II), 100% line comment coverage (Principle V), SemVer version bump (Principle IX), Zero scraper daemons (Principle VIII)  
**Scale/Scope**: Real-time push subscriptions across 5 asset domains (Rates, Gold, Gas, Stocks, Indices) and 4 weekly bucket collections  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Compliance Status | Justification / Architectural Verification |
| :--- | :--- | :--- |
| **I. API-First Contract** | **PASS** | Internal architectural refactoring only. All 49 existing HTTP endpoints and Zod schemas remain 100% identical. |
| **II. Zero-Read Caching** | **PASS** | Tier 1 active weekly bucket in-memory caching and live snapshot memory tiers remain intact. Zero Firestore reads during active serving. |
| **III. Freshness & Fallbacks** | **PASS** | Freshness indicators, timestamps, and disk fallback persistence remain preserved. |
| **IV. Performance & SLA** | **PASS** | Eliminates circular dual-write updates, reducing event loop latency. Sub-150ms latency ceiling maintained. |
| **V. Exhaustive Commenting** | **PASS** | 100% line comment coverage required for all modified and newly authored code lines. |
| **VI. Cloud Budget Safeguards** | **PASS** | $0 direct hosting cost; zero incremental database reads. |
| **VII. Mandatory Testing** | **PASS** | Full Vitest unit test suite (324+ tests) and 49-scenario verification suite (`npm run verify`) must exit with code `0`. |
| **VIII. Independent Deployment** | **PASS** | Preserves standalone API service boundaries with automated GitHub Actions CI/CD to OCI VM. |
| **IX. Mandatory Version Incrementing** | **PASS** | Application version will be incremented from `2.18.19` to `2.18.20` upon completion. |

*Gate Evaluation: ALL 9 PRINCIPLES PASS WITH ZERO VIOLATIONS.*

---

## Project Structure

### Documentation (this feature)

```text
specs/115-unify-realtime-ingestion/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan (this file)
├── research.md          # Phase 0 architectural analysis
├── data-model.md        # Phase 1 data entities and interfaces
├── quickstart.md        # Phase 1 verification instructions
├── contracts/
│   └── ingestion-api.md # Phase 1 inbound push & internal contracts
└── checklists/
    └── requirements.md  # Specification quality checklist
```

### Source Code Refactoring Layout

```text
src/
├── cache/
│   ├── cacheService.ts
│   ├── FlatCacheProvider.ts
│   ├── ICacheProvider.ts
│   ├── snapshot-cache.ts             # TO BE DELETED (27 lines)
│   └── snapshot-store.ts             # TO BE DELETED (46 lines)
├── services/
│   ├── active-weekly-bucket-store.ts # TO BE DELETED (76 lines - absorbed by HistoryResolutionEngine)
│   ├── market-schedule-service.ts    # TO BE DELETED (62 lines dead facade)
│   ├── market-schedule-engine.ts     # Authoritative schedule engine (retained)
│   ├── history-resolution-engine.ts  # Deepened: absorbs Tier 1 active weekly memory store
│   ├── firestore-listener-service.ts # Deepened: streamlined Ports & Adapters inbound listener
│   ├── live-snapshot-coordinator.ts  # Injected directly into RatesSnapshotService
│   ├── rates-snapshot-service.ts     # Constructor refactored: drops SnapshotCache/Store
│   └── ...
├── app.ts                            # Updates service instantiation & removes deleted imports
└── routes/
    └── health-routes.ts              # Updates SnapshotCache import to LiveSnapshotCoordinator

tests/
├── unit/
│   ├── active-weekly-bucket-store.test.ts  # TO BE DELETED (superseded by history-resolution-engine.test.ts)
│   ├── firestore-listener-service.test.ts # Updated to test clean push listener
│   ├── history-resolution-engine.test.ts  # Updated to test Tier 1 memory store directly
│   ├── market-schedule-engine.test.ts     # Removes dead marketScheduleService import
│   ├── rates-snapshot-service.test.ts     # Updates constructor mock setup
│   └── ...
└── integration/
    └── stale-fallback.integration.test.ts # Removes legacy SnapshotCache/Store imports
```

---

## Complexity Tracking

> **Constitution Check has zero violations. No exceptions required.**
