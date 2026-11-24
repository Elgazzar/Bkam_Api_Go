# Implementation Plan: Dissolve Shadow Caches and Firestore Fallbacks in Query Modules

**Branch**: `112-dissolve-shadow-caches` | **Date**: 2026-11-15 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/112-dissolve-shadow-caches/spec.md`

---

## Summary

Consolidate all live quote resolution across domain query modules (`StockQueryService`, `GoldQueryService`, `GasQueryService`, `IndicesQueryService`) into the authoritative `LiveSnapshotCoordinator`, stripping duplicate private shadow cache fields (`cachedPrices`, `cachedRegionalPrices`) and redundant direct Firestore fallback calls. Streamline `QuotesService` to inject and query exclusively `LiveSnapshotCoordinator`, dropping 5 redundant domain service dependencies while enforcing single-flight thundering herd protection and 100% backward compatibility across all 30 public API gateway endpoints.

---

## Technical Context

**Language/Version**: TypeScript 5.6 / Node.js >= 22 (ECMAScript Modules `type: "module"`)  
**Primary Dependencies**: Fastify 5.0, `@google-cloud/firestore`, `pino`, `vitest`, `zod`  
**Storage**: In-memory `Map` RAM caches, local filesystem JSON persistence (`./var/snapshots`), Google Cloud Firestore  
**Testing**: Vitest (`npm run test:unit`, `npm run test:contract`), differential parity runner (`npm run verify:diff`)  
**Target Platform**: Linux (OCI Ampere A1 ARM64 VM, Oracle Linux 9) / Windows development environment  
**Project Type**: Fastify REST API Gateway Web Service  
**Performance Goals**: $<200\text{ ms}$ p95 response latency on all cached endpoints; zero duplicate Firestore reads during concurrent cache misses  
**Constraints**: Strict Always Free tier envelope ($0 budget, $<50,000$ Firestore daily reads); Constitution Principle V (100% line comment coverage); 100% backward-compatible payload parity across all 30 routes  
**Scale/Scope**: 6 domain query services refactored, 1 centralized snapshot authority, 30 public API routes verified  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status | Compliance Verification |
| :--- | :--- | :---: | :--- |
| **I. API-First Contract** | Stable, versioned JSON contracts with zero silent drift | **PASS** | All existing REST endpoint responses, envelopes, and status codes remain 100% preserved. |
| **II. Zero-Read Caching** | Serve real-time quotes directly from RAM; $<50\text{k}$ reads/day | **PASS** | Shadow caches dissolved; quotes resolved from `LiveSnapshotCoordinator` RAM; single-flight coalescing protects Firestore on cold misses. |
| **III. Freshness & Fallbacks** | Source identification, timestamps, and degraded disk fallbacks | **PASS** | Stale in-memory and disk persistence recovery preserved in `LiveSnapshotCoordinator.fetchFresh`. |
| **IV. Performance & SLA** | p95 latency $<200\text{ ms}$; Brotli/gzip compression; ETags | **PASS** | Direct RAM lookups execute in $<1\text{ ms}$; promise coalescing reduces thundering herd latency. |
| **V. Exhaustive Commenting** | Every single line of code explicitly commented | **PASS** | 100% line comment coverage enforced across all modified and newly created files. |
| **VI. Resource Safeguards** | Strict $0 hosting budget on OCI Always Free | **PASS** | Eliminates uncoordinated duplicate Firestore reads during cache misses, conserving free-tier quota. |
| **VII. Mandatory Testing** | Unit tests, contract verification, and build validation | **PASS** | Verified via `test:unit` (310+ tests), `test:contract` (9 tests), and `verify:diff` (30/30 endpoints). |
| **VIII. Independent Lifecycle** | Decoupled from scraping daemons; independent deployment | **PASS** | Pure gateway refactoring; no external scraping dependencies modified. |

---

## Project Structure

### Documentation (this feature)

```text
specs/112-dissolve-shadow-caches/
├── spec.md              # Feature specification
├── plan.md              # This implementation plan
├── research.md          # Architectural analysis and decisions
├── data-model.md        # Architecture diagrams, seams, and state transitions
├── quickstart.md        # Verification and test execution guide
├── contracts/           # Interface contracts
│   └── live-snapshot-coordinator.contract.ts
└── checklists/
    └── requirements.md  # Specification quality checklist
```

### Source Code Impact

```text
src/
├── app.ts                                    # Simplifies QuotesService instantiation (1 dependency)
├── services/
│   ├── quotes-service.ts                     # Injects only LiveSnapshotCoordinator; drops 5 domain dependencies
│   ├── gold-query-service.ts                 # Deletes cachedPrices; delegates fetchLatestGoldPrices to coordinator
│   ├── gas-query-service.ts                  # Deletes cachedPrices; dissolves 95 lines of duplicate boilerplate
│   ├── stock-query-service.ts                # Deletes cachedPrices; delegates fetchLatestStockPrices to coordinator
│   ├── indices-query-service.ts              # Deletes cachedPrices & cachedRegionalPrices; delegates to coordinator
│   └── live-snapshot-coordinator.ts          # Authoritative single-store for all live market domain snapshots
tests/
└── unit/
    ├── quotes-service.test.ts                # Aligns QuotesService test suite with single authority injection
    ├── quotes-service-coordinator.test.ts    # Verifies multi-asset batch quotes directly from coordinator
    ├── gold-query-service.test.ts            # Verifies zero-shadow-cache behavior in GoldQueryService
    ├── stock-query-service.test.ts           # Verifies zero-shadow-cache behavior in StockQueryService
    └── live-snapshot-coordinator.test.ts     # Verifies single-flight pooling and multi-domain snapshot handling
```

---

## Complexity Tracking

> **No violations. All design choices strictly align with Constitution Principles I–VIII and ADR 0005.**

| Violation | Why Needed | Simpler Alternative Rejected Because |
| :--- | :--- | :--- |
| *None* | *N/A* | *N/A* |
