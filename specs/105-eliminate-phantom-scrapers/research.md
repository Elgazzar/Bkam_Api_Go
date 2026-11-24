# Research & Architectural Decisions: Eliminate Phantom Scraping Machinery

**Feature**: `105-eliminate-phantom-scrapers`

**Date**: 2026-11-15

---

## 1. RatesSnapshotService Seam Simplification

### Decision
Remove `adapters: BankAdapter[]` and `browserEngine?: ScrapingEngine` from the `RatesSnapshotService` constructor and internal state. The constructor signature will be simplified to:
```typescript
constructor(
  private readonly cache: SnapshotCache,
  private readonly store: SnapshotStore,
  private readonly freshnessService = new FreshnessService()
)
```

### Rationale
- Deep inspection of `src/services/rates-snapshot-service.ts` proved that `this.adapters` and `this.browserEngine` were never referenced or invoked anywhere in the 470-line implementation.
- All live rate data is ingested via Firestore snapshots:
  - `refresh(traceId)` calls `fetchRatesFromFirestore()`.
  - `updateCache(firestoreRates)` is called directly by `FirestoreListenerService` via real-time WebSocket push (`onSnapshot`).
- Eliminating these parameters satisfies the Codebase Design principle: *"The interface is the test surface. Callers and tests cross the same seam."*

### Alternatives Considered
- **Keep optional parameters (`adapters?: BankAdapter[]`)**: Rejected. Keeping phantom parameters preserves hypothetical seams that deceive maintainers into believing the API service can fall back to browser scraping.
- **Extract a new interface without modifying the class**: Rejected. The class itself is the module; modifying its constructor directly concentrates leverage and cleans up all call sites.

---

## 2. Scraping Machinery & Adapter Purge from API Repository

### Decision
Delete the following scraping-specific files and directories from `D:\Repo\Bkam\Bkam_Api`:
1. `src/adapters/banks/*.ts` (all 30+ commercial bank scraping adapters).
2. `src/adapters/engines/*.ts` (Playwright and Cheerio scraping engine wrappers).
3. `src/adapters/registry.ts` (`createSupportedAdapters` function).
4. `src/adapters/base/bank-adapter.ts`, `scraping-engine.ts`, `fallback-provider.ts`, `market-adapter.ts`.
5. `src/services/browser-service.ts` (Playwright browser launcher).
6. Obsolete scraper fixture and live integration tests: `tests/fixtures/*.test.ts`, `tests/integration/banks/*.test.ts`, and scraper-only utility scripts.

### Rationale
- Following Feature 104 (Repository Separation), data collection and scraping are 100% owned by `D:\Repo\Bekam_Fe_El_bank_Scrapper`.
- In `Bkam_Api`, the scraper integration tests (`tests/integration/banks/*.test.ts`) were failing with `Executable doesn't exist: chrome-headless-shell.exe` because Playwright was never intended to run in the API container.
- Removing these files purges over 3,000 lines of dead code and eliminates all brittle mock boilerplate (`createSupportedAdapters: () => ({ adapters: [], browserEngine: {} })`) from test suites.

### Alternatives Considered
- **Keep adapter files as passive references**: Rejected. Passive unused code leads to rot, causes TypeScript compilation friction, and pollutes search results.

---

## 3. Docker Image Optimization: Node 22 Slim vs Playwright Base

### Decision
Switch `D:\Repo\Bkam\Bkam_Api\Dockerfile` from:
```dockerfile
FROM mcr.microsoft.com/playwright:v1.49.1-jammy
```
to a lightweight official Node runtime:
```dockerfile
FROM node:22-alpine
```
Remove all apt-get certificate workarounds, browser font packages, and `npx playwright install chromium --with-deps` commands from the build pipeline.

### Rationale
- The API container is a lightweight Fastify gateway performing memory and disk reads from Firestore. It requires zero graphical libraries, X11 dependencies, or Chromium binaries.
- Image size decreases from ~1.5GB to ~150MB (90% reduction).
- Container cold-start time and memory consumption on OCI Free Tier (1GB RAM) are vastly improved, completely eliminating Error 137 (OOM killed).

### Alternatives Considered
- **Node 22 Slim (Debian)**: Evaluated as a fallback if Alpine musl compatibility issues arose. However, since all API dependencies (`fastify`, `@google-cloud/firestore`, `flat-cache`, `zod`) are pure JavaScript or pre-compiled standard native bindings, `node:22-alpine` is the optimal choice.

---

## 4. Package Dependency Purge

### Decision
Remove the following packages from `D:\Repo\Bkam\Bkam_Api\package.json`:
- `playwright`
- `playwright-stealth`
- `cheerio`
- `@playwright/test`
- Script `"playwright:install"`

### Rationale
- None of these packages are imported by `src/app.ts`, `src/routes/*`, `src/services/*`, or `src/cache/*`.
- Eliminates 50+ transitive npm packages and removes known security audit advisories associated with browser automation dependencies.

---

## 5. Test Suite Modernization & Cleanup

### Decision
1. In `src/app.ts`, remove `createSupportedAdapters` import and instantiate `RatesSnapshotService` directly with `snapshotCache` and `snapshotStore`.
2. In all contract and integration tests (`tests/contract/*.test.ts`, `tests/integration/*.test.ts`), remove the mock definition for `createSupportedAdapters`.
3. In `tests/unit/rates-snapshot-service.test.ts`, update test instantiations to pass only the real cache and store dependencies.
4. Verify `npm run test:unit` and `npm run test:contract` pass 100% cleanly in `Bkam_Api`.
