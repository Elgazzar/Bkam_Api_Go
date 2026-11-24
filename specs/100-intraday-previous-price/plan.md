# Implementation Plan: Expose Previous Day Price in Latest Intraday Endpoints

**Branch**: `100-intraday-previous-price` | **Date**: 2026-11-05 | **Spec**: [specs/100-intraday-previous-price/spec.md](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/100-intraday-previous-price/spec.md)

**Input**: Feature specification from `/specs/100-intraday-previous-price/spec.md`

---

## Summary

Enhance all four latest intraday endpoints (`GET /v1/stocks/:symbol/intraday/latest`, `GET /v1/indices/:symbol/intraday/latest`, `GET /v1/rates/:currency/intraday/latest`, and `GET /v1/gold/:item/intraday/latest`) to expose top-level previous trading session closing price fields (`previousClose` for stocks and benchmark indices; `previousBuyPrice` and `previousSellPrice` for bank currency rates and gold commodities).

This technical approach enables mobile clients (such as Android apps using MPAndroidChart or Jetpack Compose Canvas) to render an intraday price chart with an accurate baseline limit line and percentage change using exactly 1 HTTP network request instead of 2. All previous day prices will be derived strictly from existing volatile in-memory caches, memory-resident snapshot metadata, and static catalog constants (`SnapshotCache`, `intradayCoordinator`, `stockCatalog`, `saudiStockCatalog`, `CurrencyRate.yesterdayCloseBuy`/`yesterdayCloseSell`, `liveGold.rates[k].yesterdayBuyPrice`/`yesterdaySellPrice`), guaranteeing **0 additional Firestore database reads** (strictly complying with Principle VII).

---

## Technical Context

**Language/Version**: TypeScript 5.x / Node.js 22 LTS  
**Primary Dependencies**: Fastify 5.x, Zod 3.x, Vitest 4.x  
**Storage**: Volatile RAM snapshot caches (`SnapshotCache`, `intradayCoordinator`, `ActiveWeeklyBucketStore`), file-backed disk cache (`flat-cache`), 0 new Firestore database reads  
**Testing**: Vitest integration test suite (`api/tests/integration/latest-intraday-routes.test.ts`), contract verification  
**Target Platform**: Node.js 22 on Linux (Oracle Cloud Infrastructure Ampere A1 Always Free VM) / Windows local development  
**Project Type**: RESTful JSON API web service  
**Performance Goals**: Sub-10ms response time from in-memory cache; zero additional database queries  
**Constraints**: Strict Always Free Tier boundary ($0 budget); backward compatibility across all endpoints; exhaustive commenting on every newly added or modified line of code (Principle VI)  
**Scale/Scope**: 4 query services and routes: `StockQueryService`, `IndicesQueryService`, `RatesQueryService`, `GoldQueryService`  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Living Requirements Review (Principle XVI)**: Read `.specify/memory/REQUIREMENTS.md` before implementation; verified that the enhancement extends `FR-INTRA-001` through `FR-INTRA-004` without breaking any existing baseline behaviors or contracts.
- [x] **API-First Contract (Principle I)**: Request and response shapes, OpenAPI schemas, and client compatibility verified. Non-breaking additions of `previousClose`, `previousBuyPrice`, and `previousSellPrice`.
- [x] **Cloud Cost Safeguards (Principle VII)**: Sourcing strictly occurs from volatile RAM caches (`liveSnap`, `liveGold`, `ratesSnapshotService`) and in-memory catalogs; 0 new Firestore reads or writes.
- [x] **Exhaustive Commenting (Principle VI)**: Every newly introduced or modified line of code in query services and test suites will be explicitly commented.
- [x] **E2E Testing (Principle XII)**: Vitest integration test suite (`latest-intraday-routes.test.ts`) updated and executed.
- [x] **Intelligent Caching & Intraday Lifecycle (Principle XIII)**: Intraday timeline tick points remain strictly isolated to volatile memory (`intradayCoordinator` in RAM); sealed past weeks remain in `flat-cache`.
- [x] **Documentation Sync (Principles X, XI, XV, XVI)**: Plan includes updating `ARCHITECTURE.md`, `README.md`, `MODULAR_COMPONENTS.md`, `REQUIREMENTS.md`, and `API_ENDPOINTS_GUIDE.md`.
- [x] **Bug Tracking & Ledger (Principle XVII)**: Any defect, compilation error, or test anomaly encountered during implementation will be logged in `bugs.md`.

---

## Project Structure

### Documentation (this feature)

```text
specs/100-intraday-previous-price/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
│   └── intraday-latest.openapi.json
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code Layout

```text
api/
├── src/
│   ├── services/
│   │   ├── stock-query-service.ts    # Enriches stock intraday response with previousClose
│   │   ├── indices-query-service.ts  # Enriches index intraday response with previousClose
│   │   ├── rates-query-service.ts    # Enriches rate intraday response with previousBuyPrice & previousSellPrice
│   │   └── gold-query-service.ts     # Enriches gold intraday response with previousBuyPrice & previousSellPrice
│   ├── routes/
│   │   ├── stock-routes.ts
│   │   ├── indices-routes.ts
│   │   ├── rates-routes.ts
│   │   └── gold-routes.ts
│   └── models/
│       ├── stock-rate.ts
│       ├── index-rate.ts
│       ├── currency-rate.ts
│       └── gold-rate.ts
└── tests/
    └── integration/
        └── latest-intraday-routes.test.ts # Comprehensive integration test coverage for all 4 endpoints
```

**Structure Decision**: Single API project modifying existing query services in `api/src/services/` and expanding integration assertions in `api/tests/integration/latest-intraday-routes.test.ts`.

---

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*No violations. All design choices strictly adhere to the project constitution.*
