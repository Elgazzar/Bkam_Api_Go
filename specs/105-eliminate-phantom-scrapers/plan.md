# Implementation Plan: Eliminate Phantom Scraping Machinery from API Gateway

**Branch**: `105-eliminate-phantom-scrapers` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/105-eliminate-phantom-scrapers/spec.md`

---

## Summary

Refactor the Bkam API Gateway (`Bkam_Api`) to eliminate dead browser scraping machinery, phantom adapter seams, and browser automation dependencies. `RatesSnapshotService` constructor parameters `adapters: BankAdapter[]` and `browserEngine?: ScrapingEngine` will be removed, `src/app.ts` will stop instantiating scraping adapters on startup, 30+ legacy bank adapter files will be deleted from `Bkam_Api`, unused npm dependencies (`playwright`, `cheerio`) will be purged, and the Docker image will be migrated from the heavy Playwright base image to a lean `node:22-alpine` runtime.

---

## Technical Context

**Language/Version**: TypeScript 5.6+ / Node.js >= 22 (ESM modules)

**Primary Dependencies**: Fastify 5.0+, `@fastify/compress`, `@fastify/etag`, `@google-cloud/firestore`, `flat-cache`, `zod`

**Removed Dependencies**: `playwright`, `playwright-stealth`, `cheerio`, `@playwright/test`

**Storage**: Google Cloud Firestore (read-only push listener `onSnapshot`), volatile RAM cache (`SnapshotCache`), persistent local disk cache (`flat-cache`)

**Testing**: Vitest (`npm run test:unit`, `npm run test:contract`)

**Target Platform**: Oracle Cloud Infrastructure (OCI) Ampere A1 (Ubuntu / Docker, 1GB RAM Free Tier)

**Project Type**: Fastify REST API Gateway (pure read gateway)

**Performance Goals**: <10ms for in-memory cached responses, <200ms p95 across platform endpoints

**Constraints**: $0 hosting budget; zero Playwright browser dependencies in production container; zero regressions to existing REST contracts

---

## Constitution Check

*GATE: Pre-design and post-design validation against `.specify/memory/constitution.md`*

- [x] **Living Requirements Review (Principle XVI)**: Reviewed `REQUIREMENTS.md`. This refactor directly fulfills Principle VIII (Independent Deployment Lifecycle) and enforces strict physical and logical decoupling between scraper and API.
- [x] **API-First Contract (Principle I)**: All public JSON contracts (`/v1/rates`, `/v1/rates/:currency`, `/v1/health`, etc.) remain 100% byte-for-byte compatible with existing mobile and web clients.
- [x] **Zero-Read Memory & Disk Caching (Principle II)**: Push-based Firestore listeners (`onSnapshot`) continue to feed the in-memory cache directly with 0 read quota consumption.
- [x] **Exhaustive Commenting (Principle V)**: Every modified or added line of code in `RatesSnapshotService` and `app.ts` will be exhaustively commented.
- [x] **Mandatory Testing and Contract Verification (Principle VII)**: All contract tests (`npm run test:contract`) and unit tests (`npm run test:unit`) will be verified before merge.
- [x] **Independent Deployment Lifecycle (Principle VIII)**: Solidifies deployment independence by eliminating Playwright build steps and switching to `node:22-alpine`.

---

## Project Structure

### Documentation (this feature)

```text
specs/105-eliminate-phantom-scrapers/
├── plan.md              # Implementation plan
├── research.md          # Architectural decisions & research
├── data-model.md        # Refactored class models & data flow
├── contracts/
│   └── snapshot-service-contract.md # Service and bootstrap interfaces
├── quickstart.md        # Runnable verification scenarios
├── checklists/
│   └── requirements.md  # Quality validation checklist
└── tasks.md             # Implementation tasks (generated via /speckit-tasks)
```

### Source Code (`D:\Repo\Bkam\Bkam_Api`)

```text
src/
├── app.ts                                 # Bootstrap: remove createSupportedAdapters
├── services/
│   ├── rates-snapshot-service.ts          # Constructor: remove adapters and browserEngine
│   ├── rates-query-service.ts             # Consumes refactored snapshot service
│   └── ...                                # Unaltered query services
├── cache/
│   ├── snapshot-cache.ts                  # Unaltered in-memory cache
│   └── snapshot-store.ts                  # Unaltered disk snapshot store
└── [DELETED] adapters/
    ├── banks/                             # DELETED: 30+ bank scraping adapters
    ├── engines/                           # DELETED: Playwright / Cheerio engines
    ├── base/                              # DELETED: BankAdapter, ScrapingEngine
    └── registry.ts                        # DELETED: createSupportedAdapters

tests/
├── contract/                              # Remove createSupportedAdapters mocks
├── unit/                                  # Unit tests with simplified constructors
└── [DELETED] fixtures/                    # DELETED: obsolete bank scraping fixtures
```

---

## Implementation Strategy

1. **Step 1: Simplify RatesSnapshotService Constructor**
   - Update `src/services/rates-snapshot-service.ts` to take only `cache`, `store`, and `freshnessService`.
   - Remove all imports and typings for `BankAdapter` and `ScrapingEngine`.

2. **Step 2: Refactor Application Bootstrap in app.ts**
   - Remove `createSupportedAdapters` import and invocation.
   - Instantiate `RatesSnapshotService` directly with `snapshotCache` and `snapshotStore`.

3. **Step 3: Purge Scraping Adapters and Engines**
   - Delete `src/adapters/banks/*`, `src/adapters/engines/*`, `src/adapters/base/*`, `src/adapters/registry.ts`, and `src/services/browser-service.ts`.
   - Remove obsolete scraping fixture tests (`tests/fixtures/*`) and live bank integration tests (`tests/integration/banks/*`).

4. **Step 4: Clean Up Test Suites**
   - Remove `createSupportedAdapters` mocks from `tests/contract/*.test.ts` and `tests/integration/*.test.ts`.
   - Update `tests/unit/rates-snapshot-service.test.ts`.

5. **Step 5: Clean Up Dependencies & Dockerfile**
   - Update `package.json` to remove `playwright`, `playwright-stealth`, `cheerio`, `@playwright/test`, and `"playwright:install"`.
   - Update `Dockerfile` to use `node:22-alpine` and remove Playwright installation steps.

6. **Step 6: Verification**
   - Run `npm run build`, `npm run test:unit`, `npm run test:contract`, and verify live runtime on port 3001.
