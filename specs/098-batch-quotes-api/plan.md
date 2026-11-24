# Implementation Plan: Stateless Batch Quotes API Endpoint

**Branch**: `098-batch-quotes-api` | **Date**: 2026-10-31 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/098-batch-quotes-api/spec.md`

## Summary

Implement a unified, stateless Backend-for-Frontend (BFF) quotes endpoint (`POST /v1/quotes` and `GET /v1/quotes?items=...`) allowing mobile clients to retrieve quotes for a list of locally stored favorite asset identifiers across four financial domains (Bank Exchange Rates, Gold Karats/Coins, Market Indices, and Equities/Stocks) in a single HTTP request. The endpoint operates purely in-memory over existing pushed snapshots (`FirestoreListenerService`) with zero Firestore reads, returning an indexed map of quotes and an isolated list of unresolved identifiers.

## Technical Context

**Language/Version**: TypeScript 5.4+ / Node.js 20+ (ES Modules)  
**Primary Dependencies**: Fastify 4.x, Zod 3.x, @google-cloud/firestore (native SDK via existing instances)  
**Storage**: None (pure in-memory lookups from active pushed cache instances in RAM, zero disk/database writes)  
**Testing**: Vitest (`npm run test:unit`), Supertest/Fastify injection (`npm run test:e2e`)  
**Target Platform**: Node.js backend hosted on Oracle Cloud Infrastructure (OCI) Always Free VM  
**Project Type**: REST Web Service (Fastify modular routing plugin)  
**Performance Goals**: $< 50\text{ ms}$ processing time for batches up to 100 items; sub-millisecond in-memory resolution  
**Constraints**: Zero Firestore reads during request lifecycle; strict $0.00 cloud budget adherence; maximum 100 items per request  
**Scale/Scope**: Serves all active mobile home screen instances fetching user favorites  

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Living Requirements Review (Principle XVI)**: Read `.specify/memory/REQUIREMENTS.md`. The new BFF endpoint adds a complementary read-only query route without breaking or modifying existing `/v1/rates`, `/v1/gold`, `/v1/indices`, or `/v1/stocks` routes.
- [x] **API-First Contract (Principle I)**: Request and response schemas, error envelopes, and HTTP query parameters are formally defined in [contracts/quotes-api.yaml](./contracts/quotes-api.yaml) and [data-model.md](./data-model.md).
- [x] **Cloud Cost Safeguards (Principle VII)**: In-memory resolution from `RatesSnapshotService`, `GoldQueryService`, `IndicesQueryService`, `StockQueryService`, and `GccStockQueryService` ensures 0 Firestore reads or writes per request.
- [x] **Exhaustive Commenting (Principle VI)**: Every added or modified source line in `quotes-routes.ts`, `quotes-service.ts`, and test files will be explicitly and clearly commented.
- [x] **E2E Testing (Principle XII)**: Comprehensive unit tests and end-to-end integration tests (`tests/e2e/quotes.e2e.test.ts`) will be implemented and validated.
- [x] **Intelligent Caching & Intraday Lifecycle (Principle XIII)**: Quotes are served purely from in-memory push caches with standard `Cache-Control` headers and weak ETag support for GET queries.
- [x] **Documentation Sync (Principles X, XI, XV, XVI)**: Changes will be documented in `ARCHITECTURE.md`, `README.md`, `MODULAR_COMPONENTS.md`, and `REQUIREMENTS.md`.
- [x] **Bug Tracking & Ledger (Principle XVII)**: Any issues or regressions encountered during development will be documented in `bugs.md`.

## Project Structure

### Documentation (this feature)

```text
specs/098-batch-quotes-api/
├── plan.md              # This file (/speckit-plan output)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── quotes-api.yaml  # OpenAPI 3.0 specification
└── checklists/
    └── requirements.md  # Quality checklist
```

### Source Code Layout

```text
api/
├── src/
│   ├── models/
│   │   └── quote.ts              # Domain types & Zod validation schemas
│   ├── routes/
│   │   └── quotes-routes.ts      # Fastify routing plugin for POST/GET /v1/quotes
│   ├── services/
│   │   └── quotes-service.ts     # Core QuotesService resolving items from in-memory caches
│   └── app.ts                    # Register QuotesService & registerQuotesRoutes
└── tests/
    ├── unit/
    │   └── quotes-service.test.ts  # Unit tests for multi-domain resolution and edge cases
    └── e2e/
        └── quotes.e2e.test.ts      # End-to-end route tests for POST and GET /v1/quotes
```

**Structure Decision**: Add `quotes-service.ts` under `api/src/services/` and `quotes-routes.ts` under `api/src/routes/`, registered via dependency injection in `api/src/app.ts`. This follows the exact modular pattern of existing domain routes.

## Complexity Tracking

*No violations of constitution principles identified. No complex dependencies or infrastructure added.*
