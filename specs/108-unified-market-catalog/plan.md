# Implementation Plan: Unified Market Catalog

**Branch**: `108-unified-market-catalog` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/108-unified-market-catalog/spec.md`

---

## Summary

Deepen the existing `CatalogService` (`src/services/catalog-service.ts`) from a shallow 10-line pass-through into a unified, high-leverage `MarketCatalog` module. The deepened module encapsulates all platform static inventories (banks, currencies, commodities, domestic EGX equities, GCC equities across Saudi Arabia, UAE, Qatar, and Kuwait, and benchmark indices) behind an expressive in-memory interface. It pre-indexes entities at startup for O(1) multi-key lookup, absorbs repetitive symbol normalization and alias resolution, centralizes benchmark index constituent membership, unifies snapshot metadata enrichment, and standardizes parameter validation across HTTP routes.

---

## Technical Context

**Language/Version**: TypeScript 5.8 / Node.js 22 LTS  
**Primary Dependencies**: Fastify 5.x, Zod 3.24, @google-cloud/firestore 8.6, flat-cache 6.1  
**Storage**: In-memory multi-key hash maps in RAM (`MarketCatalog` singleton)  
**Testing**: Vitest 4.1 (`npm run test:unit`), contract tests (`npm run test:contract`)  
**Target Platform**: Linux server (Node.js 22 Alpine in Docker container, ARM64 / AMD64 on OCI Ampere A1)  
**Project Type**: Standalone REST API Gateway (`Bkam_Api`)  
**Performance Goals**: Sub-0.01ms lookup times in memory (100% compliance with SC-001 and Zero-Read Principle II)  
**Constraints**: 100% backwards-compatible REST endpoints; zero external database reads during catalog queries; exhaustive commenting on every modified line (Constitution Principle V)  
**Scale/Scope**: ~600 securities, 21 benchmark indices, 36 banks, 28 currencies, 7 commodities across 5 countries  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **API-First Contract (Principle I)**: All existing REST endpoints (`/v1/catalog`, `/v1/stocks`, `/v1/indices`, `/v1/rates`) maintain 100% contract schema stability. Documented in [catalog-routes.contract.md](./contracts/catalog-routes.contract.md).
- [x] **Zero-Read Memory Caching (Principle II)**: Catalog inventory and constituent sets reside entirely in RAM hash maps, incurring exactly zero (0) Firestore database reads.
- [x] **Freshness, Traceability, and Fallbacks (Principle III)**: Enriched quotes retain trace IDs and source freshness markers while receiving authoritative brand logos and localized names.
- [x] **Performance and SLA Guarantees (Principle IV)**: O(1) hash map access executes in <0.01ms, easily meeting the p95 <200ms gateway latency SLA.
- [x] **Exhaustive Code Commenting (Principle V)**: Every single line of newly authored and refactored code will be commented with its rationale and runtime behavior.
- [x] **Cloud Cost Safeguards (Principle VI)**: Zero cloud infrastructure cost increase; runs strictly within OCI Always Free tier footprint (<2MB added heap).
- [x] **Mandatory Testing & Contract Verification (Principle VII)**: Unit tests (`market-catalog.test.ts`), route integration tests, and contract verification must pass 100% green before merge.
- [x] **Independent Deployment Lifecycle (Principle VIII)**: No external scraping dependencies introduced; clean in-process domain refactoring.

---

## Project Structure

### Documentation (this feature)

```text
specs/108-unified-market-catalog/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan (this file)
├── research.md          # Architectural research & technical decisions
├── data-model.md        # Domain entities, indexes & complexity analysis
├── quickstart.md        # Runnable verification guide
├── contracts/           # Interface contracts
│   ├── market-catalog.contract.ts
│   └── catalog-routes.contract.md
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (`D:\Repo\Bkam\Bkam_Api\src`)

```text
src/
├── services/
│   ├── catalog-service.ts              # [DEEPEN] Transforms into deep MarketCatalog module
│   ├── stock-query-service.ts          # [SIMPLIFY] Delegates enrichment & filtering to catalog
│   ├── indices-query-service.ts        # [SIMPLIFY] Delegates constituent lookup to catalog
│   └── rates-query-service.ts          # [SIMPLIFY] Uses catalog for bank/currency validation
├── config/
│   ├── stock-catalog.ts                # Static EGX stock definitions (encapsulated)
│   ├── index-catalog.ts                # Static benchmark index definitions (encapsulated)
│   └── market-registry.ts              # Operating exchange registry
├── modules/gcc-stocks/
│   ├── catalogs/                       # Static GCC catalog files (encapsulated)
│   └── services/
│       └── gcc-stock-query-service.ts  # [SIMPLIFY] Delegates GCC enrichment & lookup to catalog
└── routes/
    ├── catalog-routes.ts               # Serves unified inventory from MarketCatalog
    ├── stock-routes.ts                 # Uses catalog validation for group & symbol params
    └── indices-routes.ts               # Uses catalog for index validation & constituents
```

**Structure Decision**: In-process deepening inside `src/services/catalog-service.ts` (aliased as `MarketCatalog`). Existing static catalog files in `src/config/` and `src/modules/gcc-stocks/catalogs/` remain internal data sources encapsulated by the deep catalog module, without disrupting git history.

---

## Complexity Tracking

*No violations or unjustified complexity introduced. All principles satisfied.*
