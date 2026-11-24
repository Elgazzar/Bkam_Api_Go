# Implementation Plan: Unify Fragmented Market Schedules into Deep MarketSchedule Engine

**Branch**: `111-unify-market-schedules` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/111-unify-market-schedules/spec.md`

---

## Summary

Consolidate conflicting market schedule declarations and the expensive 7-day Firestore document scanning loop into a single, deep `MarketScheduleEngine` (`src/services/market-schedule-engine.ts`). The engine hides multi-market calendar configurations (`EGX`, `SA`, `AE`, `QA`, `KW`), timezone offsets, and national holidays behind a minimal 3-method seam (`isMarketOpen`, `getSecondsUntilNextOpen`, `getLatestTradingSessionDate`). By calculating trading session dates mathematically in memory with $O(1)$ complexity, off-hours session backtracking achieves 0 database reads and sub-millisecond execution.

---

## Technical Context

**Language/Version**: TypeScript 5.8 / Node.js 22 LTS  
**Primary Dependencies**: Fastify 5.x, native `Intl.DateTimeFormat` (IANA timezone database), `date-fns`  
**Storage**: Stateless engine (0 database reads, 0 persistent disk files); integrates with `flat-cache` / `CacheService` via TTL calculation  
**Testing**: Vitest (`npm run test:unit`, `npm run test:contract`, `npm run verify:diff`)  
**Target Platform**: Linux (Ubuntu on OCI Ampere A1 ARM64) & Windows Local Dev  
**Project Type**: Monolithic modular REST API gateway service  
**Performance Goals**: $<1\text{ ms}$ session backtracking calculation (0 Firestore reads); $<200\text{ ms}$ p95 API response times  
**Constraints**: Zero database read budget consumption on off-hours session lookups; 100% backward-compatible REST contracts; 100% commented code per Constitution Principle V  
**Scale/Scope**: 5 financial exchanges (`EGX`, `SA`, `AE`, `QA`, `KW`) + domestic banking and commodities schedules  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Plan Conformance | Status |
| :--- | :--- | :--- | :--- |
| **I. API-First Contract** | Stable, versioned JSON contracts; zero payload drift | All existing REST routes preserve identical response shapes, headers, and status codes. | ✅ PASS |
| **II. Zero-Read Memory Caching** | Serve real-time data from RAM; protect cloud DB free tier | Eliminates the 7-day Firestore document scanning loop entirely (0 database reads on off-hours lookups). | ✅ PASS |
| **III. Freshness & Fallbacks** | Clear degradation markers and timestamps | Provides precise off-hours TTL and session freshness indicators. | ✅ PASS |
| **IV. Performance Guarantees** | p95 $<200\text{ ms}$; transport compression | Reduces session resolution latency from >300ms to $<1\text{ ms}$ in memory. | ✅ PASS |
| **V. Exhaustive Commenting** | Every single line explicitly commented | All newly created and modified files will have 100% line comments. | ✅ PASS |
| **VI. Cloud Budget Safeguards** | $0 hosting budget; zero quota burn | Eliminates up to 7 Firestore reads per off-hours intraday request. | ✅ PASS |
| **VII. Mandatory Testing** | Unit, contract, and differential tests passing | Full unit test suite, contract suite, and 30-endpoint differential parity verification required. | ✅ PASS |
| **VIII. Independent Lifecycle** | Decoupled from external daemons | Engine is self-contained within API gateway runtime. | ✅ PASS |

---

## Project Structure

### Documentation (this feature)

```text
specs/111-unify-market-schedules/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan
├── research.md          # Architectural research & problem analysis
├── data-model.md        # Type schemas and entity relationships
├── quickstart.md        # Usage examples and testing guidelines
├── contracts/           # Seam interface definitions
│   └── market-schedule-engine.contract.ts
├── checklists/          # Quality validation checklists
│   └── requirements.md
└── tasks.md             # Implementation tasks (created in next step)
```

### Source Code (repository root)

```text
src/
├── services/
│   ├── market-schedule-engine.ts     # Deep engine encapsulating registry, math, holidays
│   ├── market-schedule-service.ts    # Slim facade delegating to marketScheduleEngine
│   ├── intraday-coordinator.ts       # Refactored to replace 7-day Firestore loop with engine
│   ├── firestore-listener-service.ts # Consumes marketScheduleEngine for off-hours persistence
│   └── history-resolution-engine.ts  # Consumes marketScheduleEngine for TTL calculations
├── config/
│   └── market-registry.ts            # isMarketOpenById delegates to marketScheduleEngine
└── modules/gcc-stocks/lib/
    └── market-schedule.ts            # Re-exports unified engine schedules for backwards compat

tests/
├── unit/
│   ├── market-schedule-engine.test.ts # Exhaustive tests for all markets, rollovers, holidays
│   ├── intraday-coordinator.test.ts   # Verifies zero-read session backtracking
│   └── firestore-listener-service.test.ts
└── contract/
```

**Structure Decision**: Single modular service architecture. Deep `MarketScheduleEngine` encapsulates all scheduling logic in `src/services/market-schedule-engine.ts`, with existing `market-schedule-service.ts` demoted to a thin facade to prevent breaking legacy imports.

---

## Phase 0: Research & Problem Resolution

- **Research Artifact**: [research.md](./research.md)
- **Key Findings**:
  1. `market-schedule-service.ts` declared EGX hours as 09:30–16:00, while `market-registry.ts` declared 09:00–15:30. Resolved to canonical 09:00–15:30 Cairo time.
  2. `intraday-coordinator.ts` scanned Firestore across a 7-day loop. Resolved to deterministic memory calculation via `getLatestTradingSessionDate(market)`.
  3. UAE operates Monday–Friday, while Egypt/Saudi/Qatar/Kuwait operate Sunday–Thursday. Resolved via per-market `activeDays` configuration.

---

## Phase 1: Design Artifacts

- **Data Model**: [data-model.md](./data-model.md) defining `MarketId`, `MarketScheduleConfig`, `MarketScheduleState`.
- **Contracts**: [contracts/market-schedule-engine.contract.ts](./contracts/market-schedule-engine.contract.ts) defining `IMarketScheduleEngine`.
- **Quickstart**: [quickstart.md](./quickstart.md) documenting core usage patterns and test commands.

---

## Complexity Tracking

> Zero constitutional violations detected. Architecture directly simplifies the codebase by eliminating duplicate schedule modules and redundant database reads.
