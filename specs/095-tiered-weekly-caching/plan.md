# Implementation Plan: Tiered Weekly Caching & API Performance Optimization

**Branch**: `095-tiered-weekly-caching` | **Date**: 2026-10-24 | **Spec**: [spec.md](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/095-tiered-weekly-caching/spec.md)

**Input**: Feature specification from `/specs/095-tiered-weekly-caching/spec.md`

## Summary

Implement a multi-tier caching, transport, and validation architecture to make API loading times faster and optimize storage for weekly consolidated buckets and the active calendar week.
1. **Session-Aware Active Week Disk Caching**: Gate `flat-cache` persistence of the active week to market-closed periods (evenings, nights, and the 48-hour Friday/Saturday weekend), setting an expiration timestamp synchronized to the next market open (09:30 AM Cairo of the next trading day). During trading hours (Sunday–Thursday 09:30–16:00 Cairo), disk writes are strictly bypassed and memory caching is used with 5-second deduplication.
2. **Range Result Cache & Historical Prefix Stitching**: Persist finalized symbol time-series in `flat-cache` for immutable past ranges (30-day TTL). For ranges ending today, cache the closed historical prefix (`[start, yesterdayCairo]`) in `flat-cache`, dynamically stitching today's live price and in-memory intraday points in RAM at request time.
3. **HTTP Transport Acceleration**: Install and register `@fastify/compress` (1KB threshold) for transparent Gzip and Brotli compression (shrinking payload sizes by 85–90% for Android OkHttp and web clients).
4. **Conditional Requests & Dynamic Cache-Control**: Register `@fastify/etag` to support HTTP 304 Not Modified validation, applying `Cache-Control: public, max-age=86400, immutable` for historical ranges ending before today, and `max-age=60` for ranges ending today.

## Technical Context

**Language/Version**: TypeScript 5.6, Node.js >= 22 (ES modules)  
**Primary Dependencies**: Fastify 5.0, `@fastify/compress` ^9.0, `@fastify/etag` ^6.0, `flat-cache` ^6.1, `firebase-admin` ^13.10  
**Storage**: Cloud Firestore (`stock_prices_history_weekly`, `stock_prices`), Local Disk Cache (`FlatCacheProvider` in `.cache/api-cache`), RAM (`intradaySnapshotService`, `weeklyBucketCache`)  
**Testing**: Vitest (`npm run test:unit`, `npm run test:integration`)  
**Target Platform**: Node.js Linux/Windows server, Fastify Gateway  
**Project Type**: REST API Gateway / Web Service  
**Performance Goals**:
- Historical range queries resolve in `<20ms` warm / `<300ms` cold.
- Response payloads compressed by `>=80%` (from ~300KB down to <50KB).
- Off-hours process restarts achieve 100% cache hits from disk with 0 database reads.
- Repeated chart navigations return `304 Not Modified` with 0 payload bytes in `<10ms`.
**Constraints**:
- Intraday points MUST NEVER be cached to disk; they exist strictly in RAM for the active session.
- Zero breaking changes to client API contracts.
- Strictly adhere to Constitution Principle VI (exhaustive comments on every new line).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Living Requirements Review (Principle XVI)**: Read `.specify/memory/REQUIREMENTS.md` before implementation; ensure new features or bug fixes do not regress or conflict with existing system/functional requirements. (New FR-STK-007 through FR-STK-010 defined and aligned).
- [x] **API-First Contract (Principle I)**: Request/response shapes, versioning, error envelopes, and client compatibility verified. Zero breaking changes to `GET /v1/stocks/:symbol/history/range`.
- [x] **Cloud Cost Safeguards (Principle VII)**: Off-hours session-aware disk caching preserves Always Free Firestore read quotas during maintenance and deployments.
- [x] **Exhaustive Commenting (Principle VI)**: Every new or modified line of code will include clear explanatory comments.
- [x] **E2E Testing (Principle XII)**: Unit and integration test suites will validate compression, ETags, 304 responses, and cache tiering.
- [x] **Documentation Sync (Principles X, XI, XV, XVI)**: Plans include updating `ARCHITECTURE.md`, `README.md`, `MODULAR_COMPONENTS.md`, and `REQUIREMENTS.md`.

## Project Structure

### Documentation (this feature)

```text
specs/095-tiered-weekly-caching/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output: decisions, rationale, alternatives
├── data-model.md        # Phase 1 output: cache entities, models, session temporal logic
├── contracts/           # Phase 1 output: HTTP contract for stock history range endpoint
│   └── stock-history-range.contract.md
├── quickstart.md        # Phase 1 output: validation scenarios and testing instructions
└── checklists/
    └── requirements.md  # Spec quality validation checklist
```

### Source Code Layout

```text
api/
├── package.json                               # Add @fastify/compress, @fastify/etag
├── src/
│   ├── app.ts                                 # Register @fastify/compress and @fastify/etag
│   ├── routes/
│   │   └── stock-routes.ts                    # Dynamic Cache-Control headers (immutable past vs 60s today)
│   ├── services/
│   │   ├── firestore-service.ts               # Session-aware active week flat-cache gating & expiration calculation
│   │   └── stock-query-service.ts             # Range result caching and dynamic today prefix stitching
│   └── cache/
│       ├── FlatCacheProvider.ts               # TTL expiration handling and disk persistence
│       └── cacheService.ts                    # Cache service singleton facade
└── tests/
    ├── unit/
    │   ├── firestore-service.test.ts          # Unit tests for session-aware active week caching
    │   └── stock-query-service.test.ts        # Unit tests for range result caching and prefix stitching
    └── integration/
        └── stock-history.test.ts              # Integration tests for compression, ETags, and 304 responses
```

**Structure Decision**: Multi-file Fastify gateway architecture. All caching and transport enhancements are localized to `api/src/services`, `api/src/routes`, `api/src/app.ts`, and corresponding unit/integration test suites.

## Complexity Tracking

> **No Constitution violations detected. No complexity exceptions required.**
