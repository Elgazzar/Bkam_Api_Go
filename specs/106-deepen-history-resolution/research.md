# Research & Architectural Decisions: Deepen History Resolution Engine

**Feature**: `106-deepen-history-resolution`  
**Date**: 2026-11-15  
**Domain**: API Gateway / Market History Resolution

---

## 1. Context & Architectural Problem

Currently, `HistoryResolutionEngine` (`src/services/history-resolution-engine.ts`) implements an interface (`IHistoryResolutionEngine`) that only resolves historical past dates from storage tiers (Active Week RAM store -> flat-cache disk storage -> Firestore weekly buckets).

Consequently, every caller service (`RatesQueryService`, `GoldQueryService`, `StockQueryService`, and `IndicesQueryService`) must duplicate a shallow branching pattern:
1. Call `getCurrentCairoDate()` to obtain the Africa/Cairo operational calendar date.
2. Check `const isToday = (date === todayCairo)`.
3. If `isToday`:
   - Fetch the live in-memory snapshot for the asset class.
   - Invoke `intradayCoordinator.resolveTimeline(...)`.
   - Stitch intraday points to live quotes.
   - Synthesize and format the active day response.
4. If `!isToday`:
   - Call `historyResolutionEngine.getDateSnapshot(assetClass, date, market)`.
   - Re-format or pass through the archive response.
5. In date range queries:
   - Call `historyResolutionEngine.getDateRange(assetClass, options)`.
   - Check if the range spans through today (`endDate >= todayCairo`).
   - If so, manually fetch today's live synthesized snapshot and append it to the range array.

This design violates **locality** (date boundary logic is scattered across 4 services) and makes `HistoryResolutionEngine` **shallow** (callers must do the heavy lifting of date-type discernment).

---

## 2. Technical Decisions & Design Choices

### Decision 1: Deepen `IHistoryResolutionEngine` to Absorb Active Day Synthesis

- **Decision**: Update `HistoryResolutionEngine` to be the single entry point for resolving any calendar date snapshot (`getDateSnapshot(assetClass, date, options)`) and any date range (`getDateRange(assetClass, options)`), absorbing:
  - Evaluation of `getCurrentCairoDate()`
  - Detection of `isToday`
  - Coordination with `intradayCoordinator` for live intra-session timeline points
  - Synthesis of live today quotes into daily snapshot formats
  - Multi-day range stitching (automatically appending today's live session when `endDate >= todayCairo`)
- **Rationale**:
  - **High Locality**: Session rollover, timezone arithmetic, and live-vs-historical synthesis live in exactly one module.
  - **High Leverage**: One interface serves all 4 financial asset classes (FX rates, gold commodities, equities, benchmark indices).
  - **Interface Simplification**: The 4 query services reduce from ~700-line monolithic files by eliminating over 350 lines of duplicate branching boilerplate.
- **Alternatives Considered**:
  - *Keep `HistoryResolutionEngine` separate and introduce a new `HistoryCoordinator` wrapper*: Rejected because it introduces yet another shallow wrapper module without reducing total moving parts (violates deletion test).
  - *Keep branching inside each query service*: Rejected because any bug fix (like time zone handling or midnight rollover) must be manually applied across 4 separate files, leading to architectural drift.

---

### Decision 2: Preserve Backwards-Compatible Caller Contracts and Filters

- **Decision**: `getDateSnapshot` will accept an optional `options` parameter supporting:
  - `market?: GccMarket` (for regional GCC equities & indices)
  - `bank?: string` (for bank-isolated rate snapshots)
  - `currency?: string` (for currency-filtered snapshots)
  - `symbol?: string` (for ticker-isolated equity snapshots)
- **Rationale**: Callers already pass these filters to query services; absorbing them directly into the engine preserves existing HTTP route behaviors without modifying public client APIs.
- **Alternatives Considered**:
  - *Return raw unfiltered snapshots and require callers to filter*: Rejected because filtering in the engine allows disk and range cache optimizations to be applied centrally.

---

### Decision 3: Cache Lifetime Policy Encapsulation

- **Decision**: `getDateSnapshot` returns an enriched result metadata envelope indicating whether the resolved snapshot is dynamic (`isToday: true`, cache TTL 60s) or immutable (`isToday: false`, cache TTL 3600s).
- **Rationale**: Allows Fastify routes to set standard HTTP `Cache-Control` headers (`max-age=60` vs `max-age=3600`) directly from the resolution engine's returned flag without re-computing timezone dates in the route layer.
- **Alternatives Considered**:
  - *Let routes check date string equality*: Rejected because it re-introduces date parsing in the HTTP layer.

---

## 3. Risk Assessment & Mitigations

| Risk | Impact | Mitigation Strategy |
|---|---|---|
| Regressing existing live or historical route responses | High | Run full verification suite (36 scenarios) and differential parity comparator (30 endpoints) on Port 3000 vs 3001. |
| Circular dependencies between `HistoryResolutionEngine` and Query Services | Medium | `HistoryResolutionEngine` delegates live in-memory lookups to dedicated snapshot caches (`SnapshotService`, `SnapshotCache`) rather than calling back into high-level query services. |
| Midnight Cairo rollover edge case | Medium | Covered by existing `intradayCoordinator` fallback tests that verify dynamic lookback ceiling. |

