# Research & Technical Decisions: Deepen Intraday Coordinator

**Branch**: `107-deepen-intraday-coordinator` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

---

## Technical Problem Analysis

Prior to this feature, `IntradayCoordinator` was a shallow module. While it managed the volatile memory buffers (`intradaySnapshotService`) and descending Firestore queries (`orderBy(FieldPath.documentId(), 'desc').limit(1)`), its interface only returned raw symbol-to-points dictionaries via `resolveLatestTimeline`.

As a direct consequence of this shallowness, 4 domain query services (`RatesQueryService`, `GoldQueryService`, `StockQueryService`, and `IndicesQueryService`) each leaked over 150 lines of duplicate orchestration boilerplate:
1. **Live vs. Concluded Session Branching**: Each service checked `sessionResult.isToday`, fetching live RAM quotes for active sessions or calling `HistoryResolutionEngine.getDateSnapshot` for past sessions.
2. **Previous Price Baselining (Feature 100)**: Each service implemented bespoke extraction logic to find previous session closing prices (`yesterdayCloseBuy`/`oldBuyRate`, `yesterdayBuyPrice`, `yesterdayClosePrice`/`oldPrice`, `yesterdayClose`/`previousClose`).
3. **Symbol Point Matching & Normalization**: Each service parsed points maps, handled case-insensitivity and karat aliases (`24` vs `24k`), and formatted `{ time, price }` arrays.
4. **Envelope Construction**: Each service manually packaged `{ data: ..., isToday: boolean }` and attached timestamps.

By applying the **deletion test**, deleting this shallow seam concentrates complexity cleanly inside `IntradayCoordinator` rather than scattering it across 4 callers.

---

## Architectural Decisions

### Decision 1: Unified High-Leverage Seam (`getLatestIntraday`)

- **Decision**: Deepen `IntradayCoordinator` with a single primary entry point:
  ```typescript
  async getLatestIntraday<T = any>(
    assetClass: AssetClass,
    target?: string,
    options?: LatestIntradayOptions
  ): Promise<LatestIntradayResult<T>>
  ```
- **Rationale**: Callers now need only specify *what* asset class and identifier they want. All session state evaluation, live memory extraction, off-hours session backtracking, previous-price baselining, and envelope serialization are completely encapsulated inside `IntradayCoordinator`.
- **Alternatives Considered**:
  - *Keep `resolveLatestTimeline` and add separate helper services*: Rejected because it introduces more shallow modules and preserves caller coordination leakage.
  - *Place intraday quote synthesis inside `HistoryResolutionEngine`*: Rejected because `HistoryResolutionEngine` is focused on date-indexed historical archives and daily calendar boundaries, whereas `IntradayCoordinator` specializes in high-frequency volatile ticks, intra-day rollover, and 7-day lookback session backtracking.

---

### Decision 2: Centralized Previous Price Baselining (Feature 100 Consolidation)

- **Decision**: Consolidate all previous closing price baseline extraction rules into `IntradayCoordinator`:
  - **FX Rates**: `previousBuyPrice` from `yesterdayCloseBuy ?? oldBuyRate ?? buy`, `previousSellPrice` from `yesterdayCloseSell ?? oldSellRate ?? sell`.
  - **Gold**: `previousBuyPrice` from `yesterdayBuyPrice ?? buy`, `previousSellPrice` from `yesterdaySellPrice ?? sell`.
  - **Stocks**: `previousPrice` from `yesterdayClosePrice ?? previousClosePrice ?? oldPrice ?? price`.
  - **Indices**: `previousValue` from `yesterdayClose ?? previousClose ?? value`.
- **Rationale**: Centralizing this logic ensures consistency, eliminates off-hours baselining regressions (such as BUG-009 where overnight rates returned 0 baseline), and reduces caller query service code size by 4x.
- **Alternatives Considered**:
  - *Require clients to calculate baselines from intraday points*: Rejected because clients need the baseline to calculate change percentages even when active points are empty.

---

### Decision 3: Clean Seam Composition with `HistoryResolutionEngine`

- **Decision**: For concluded sessions (`isToday: false`), `IntradayCoordinator` delegates historical quote retrieval directly to `historyResolutionEngine.getDateSnapshot(assetClass, sessionDate, options)`.
- **Rationale**: Leverages the newly deepened `HistoryResolutionEngine` (Feature 106). `IntradayCoordinator` does not duplicate weekly bucket queries or flat-cache disk lookups; it relies on `HistoryResolutionEngine` to provide the finalized daily closing quotes for the backtracked session date.
- **Alternatives Considered**:
  - *Direct Firestore queries in `IntradayCoordinator`*: Rejected as an architectural regression violating Principle VII (0 unnecessary Firestore reads) and bypassing `flat-cache`.

---

### Decision 4: Non-Breaking Contract Preservation & Caching Propagation

- **Decision**: Response bodies returned to client routes remain 100% byte-for-byte backwards-compatible. In addition, `LatestIntradayResult` exposes non-enumerable caching metadata (`_isToday`, `_cacheTtlSeconds`) allowing Fastify route handlers to set:
  - Active Today: `Cache-Control: public, max-age=60`
  - Concluded Historical Session: `Cache-Control: public, max-age=3600`
- **Rationale**: Adheres to Constitution Principle I (API-First Contract) and Principle XIII (Intelligent Caching).

