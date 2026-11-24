# Feature Specification: Eliminate Phantom Scraping Machinery from API Gateway

**Feature Branch**: `105-eliminate-phantom-scrapers`

**Created**: 2026-11-15

**Status**: Draft

**Input**: User description: "1. Eliminate the Phantom Scraping Machinery from API Gateway
Files :
src/services/rates-snapshot-service.ts • src/adapters/registry.ts • src/adapters/banks/* • src/services/browser-service.ts • src/app.ts
Problem: The `RatesSnapshotService` constructor mandates `BankAdapter[]` and `ScrapingEngine`, yet its implementation never invokes them—relying 100% on pre-populated Firestore snapshots from the serverless scraper.
Solution: Delete the phantom adapter seam and dead browser scraper tree from `Bkam_Api`, concentrating the module interface entirely on Firestore snapshot ingestion and RAM cache delivery."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Clean Standalone API Gateway Startup & Memory Footprint (Priority: P1)

As an API operator and cloud infrastructure maintainer, I want the Bkam API server to initialize immediately without instantiating or configuring 30+ commercial bank scraping adapters or launching Playwright background browser contexts, so that container startup is fast, lightweight, and completely insulated from headless browser crashes, memory leaks, or missing system dependencies on low-memory environments (such as OCI Free Tier 1GB RAM instances).

**Why this priority**: The API Gateway is strictly a high-performance read gateway serving pre-populated data from Google Firestore. Instantiating scraping machinery that is never called wastes memory, slows boot times, and couples the API container to browser automation tools.

**Independent Test**: Can be tested by starting the API Gateway (`npm run dev` or `npm start`) without Chromium or Playwright installed on the host machine. The server boots cleanly, logs zero browser initialization messages, listens on port `3001`, and serves all `/v1/rates` endpoints from in-memory cache and Firestore push listeners.

**Acceptance Scenarios**:

1. **Given** the API Gateway starts up via `app.ts`, **When** `buildApp()` executes, **Then** no bank scraping adapters are instantiated, no browser context is launched, and the server begins listening on the configured port immediately.
2. **Given** an incoming request to `GET /v1/health`, **When** the health check is evaluated, **Then** the service reports status `200 OK` (or `degraded` if Firestore is connecting) with sub-10ms response latency without evaluating any browser health.

---

### User Story 2 - Decoupled Snapshot Ingestion Interface (Priority: P1)

As an API developer, I want `RatesSnapshotService` to present a focused, minimal interface centered exclusively on Firestore real-time snapshot ingestion and in-memory cache delivery (`getSnapshot`, `refresh`, `updateCache`), without accepting unused `BankAdapter[]` or `ScrapingEngine` constructor parameters across a phantom seam.

**Why this priority**: Constructor parameters that are never called represent dead seams that confuse developers, complicate testing, and violate the core design principle that interfaces should accurately describe required dependencies.

**Independent Test**: Can be tested by inspecting the TypeScript constructor of `RatesSnapshotService` and verifying that it requires only its true dependencies (`SnapshotCache`, `SnapshotStore`, `FreshnessService`), with zero references to `BankAdapter` or `ScrapingEngine`.

**Acceptance Scenarios**:

1. **Given** `RatesSnapshotService`, **When** inspecting its constructor signature, **Then** it accepts only cache and freshness dependencies, with zero parameters for bank scrapers or browser engines.
2. **Given** Firestore push listener updates via `updateCache()`, **When** new bank rate snapshots arrive from Firestore, **Then** the service normalizes, enriches, and stores the rates in the active in-memory cache without calling any external scrapers.
3. **Given** an explicit refresh request via `refresh(traceId)`, **When** invoked, **Then** the service queries Firestore directly via `fetchRatesFromFirestore()` and returns the enriched snapshot.

---

### User Story 3 - Elimination of Dead Scraping Dependencies and Artifacts (Priority: P2)

As a DevOps engineer and repository maintainer, I want the API Gateway repository (`Bkam_Api`) to be purged of obsolete scraping adapters (`src/adapters/banks/*`), scraping engines (`src/adapters/engines/*`), browser services (`src/services/browser-service.ts`), and unused scraper dependencies (`playwright`, `playwright-stealth`, `cheerio`), so that Docker images are lean and CI/CD pipelines run test suites without browser download steps.

**Why this priority**: Unused dependencies bloat the container image by hundreds of megabytes, introduce unnecessary security surface area, and cause test failures in environments lacking OS-level browser libraries.

**Independent Test**: Can be tested by running `npm test` and `npm run test:unit` in `Bkam_Api` on an environment where `npx playwright install` was never executed. All test suites pass cleanly with 0 failures.

**Acceptance Scenarios**:

1. **Given** the `Bkam_Api` repository, **When** examining `src/adapters/`, **Then** all 30+ commercial bank scraping adapters and browser engine abstractions have been removed.
2. **Given** `package.json`, **When** dependencies are inspected, **Then** `playwright`, `playwright-stealth`, and `cheerio` are removed from runtime dependencies.
3. **Given** the CI/CD test workflow (`.github/workflows/test-api.yml`), **When** tests execute, **Then** all unit and contract tests pass without installing Chromium or browser dependencies.

---

### Edge Cases

- **Firestore Temporary Disconnect**: If Google Firestore is momentarily unreachable during startup, the snapshot module loads the latest valid disk-persisted snapshot from `var/snapshots/latest-snapshot.json` and serves requests in degraded mode without crashing.
- **Legacy Test Suite Expectations**: Tests that previously mocked `createSupportedAdapters` or passed mock `BankAdapter` instances are refactored to pass pure snapshot data directly, ensuring tests verify domain behavior across the real seam rather than testing phantom browser abstractions.
- **Off-Market & Weekend Hours**: When banks do not publish new rates over weekends or official banking holidays, the snapshot module preserves cached rates and availability records with proper timestamps and warnings, unaffected by the removal of direct bank scraping code.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST remove `adapters: BankAdapter[]` and `browserEngine?: ScrapingEngine` from the `RatesSnapshotService` constructor, properties, and internal references.
- **FR-002**: The system MUST remove the `createSupportedAdapters()` call and all browser scraper initialization logic from `src/app.ts`.
- **FR-003**: The system MUST delete all individual bank scraping adapter files (`src/adapters/banks/*.ts`) from the API Gateway codebase (`D:\Repo\Bkam\Bkam_Api`).
- **FR-004**: The system MUST delete `src/adapters/registry.ts`, `src/adapters/engines/*`, and `src/services/browser-service.ts` from the API Gateway codebase.
- **FR-005**: The system MUST remove `playwright`, `playwright-stealth`, and `cheerio` from runtime dependencies in `Bkam_Api/package.json`.
- **FR-006**: The system MUST update all unit tests and contract tests in `Bkam_Api` that previously depended on `createSupportedAdapters` or mock bank adapters, ensuring all tests assert against the real `RatesSnapshotService` interface.
- **FR-007**: The system MUST preserve 100% backward compatibility of all public HTTP endpoints (`/v1/rates`, `/v1/rates/:currency`, `/v1/banks/:bank/rates`, `/v1/health`, `/v1/catalog`), returning identical JSON structures and status codes.
- **FR-008**: The system MUST maintain zero Firestore read costs during active serving by retaining the push-based listener architecture (`onSnapshot`) feeding the in-memory snapshot cache.

---

### Key Entities

- **RateSnapshot**: The unified platform model containing active multi-currency exchange rates, bank availability records, operational warnings, and retrieval timestamps.
- **SnapshotCache**: In-memory cache holding the volatile `RateSnapshot` to deliver sub-10ms response times.
- **FirestoreRateDocument**: Pre-scraped exchange rate documents stored in the Firestore `rates` collection by the external serverless scraper.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The API server startup time decreases and requires zero browser subprocesses or Chromium dependencies.
- **SC-002**: Over 3,000 lines of dead scraping code and 35+ redundant adapter files are eliminated from `Bkam_Api`.
- **SC-003**: The production Docker container image size decreases significantly by eliminating Playwright and browser dependencies.
- **SC-004**: 100% of unit tests (`npm run test:unit`) and contract tests (`npm run test:contract`) pass cleanly without requiring browser binaries.
- **SC-005**: All public endpoints continue to produce identical responses, preserving 100% contract fidelity with mobile and web clients.

---

## Assumptions

- Bank rate data collection and web scraping are 100% owned and executed by the independent serverless scraper in `D:\Repo\Bekam_Fe_El_bank_Scrapper`, writing normalized records into Google Firestore.
- The Bkam API Gateway service is strictly a read gateway and will never perform direct website scraping or browser automation in production.
- Client applications rely exclusively on the standardized REST contract defined in `API_ENDPOINTS_GUIDE.md` and are unaffected by internal module refactoring.

