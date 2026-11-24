# Phase 0 Research: Dedicated Latest Intraday Timeline Endpoints

**Feature**: `097-latest-intraday-api`  
**Date**: 2026-10-24  
**Status**: Completed  

---

## Technical Context & Decisions

### Decision 1: Dedicated REST Routes vs. Range Query Adaptation
- **Decision**: Introduce explicit, dedicated REST endpoints:
  - `GET /v1/stocks/:symbol/intraday/latest` (supports optional `?market=saudi|uae|qatar|kuwait`)
  - `GET /v1/indices/:symbol/intraday/latest` (supports optional `?market=...`)
  - `GET /v1/gold/:item/intraday/latest` (supports `:item` as `24`, `21`, `18`, `pound`, `international`, or `all`)
  - `GET /v1/rates/:currency/intraday/latest` (supports optional `?bank=...`)
- **Rationale**: Mobile charting components need single-purpose, lightweight responses containing exactly one session's timeline and the latest quote summary. Querying multi-day ranges forces client-side date arithmetic, weekend detection, and high network payload overhead.
- **Alternatives Considered**:
  - *Adding `?latest=true` to existing range endpoints*: Rejected because range endpoints expect `start` and `end` date parameters and return an array of multi-day snapshots (`history: [...]`), which conflicts with the flat single-session envelope required for sparkline widgets.
  - *Generic `/v1/history/latest` endpoint*: Rejected because each asset domain has distinct lookup parameters (e.g., `:symbol` for stocks vs. `:item` for gold vs. `:currency` for bank rates).

---

### Decision 2: Zero-Read Active Session Serving from In-Memory Snapshot Cache
- **Decision**: During active trading sessions, the API resolves intraday points exclusively from the volatile in-memory snapshot store (`IntradaySnapshotService`, singleton `intradaySnapshotService`) in RAM.
- **Rationale**: Active sessions experience rapid quote changes (every 1–5 minutes). The existing `FirestoreListenerService` listens to real-time `onSnapshot` events and synchronizes points directly to `IntradaySnapshotService` in RAM. Serving live requests directly from RAM guarantees sub-millisecond response latency (<0.1ms) and 0 Firestore read costs, satisfying Constitution Principle VII (Strict Cloud Cost Safeguards) and Principle XIII (Intelligent Caching).
- **Alternatives Considered**:
  - *Directly querying Firestore on every API request*: Violates Principle VII; would burn through the 50,000 daily Always Free read quota within minutes under moderate traffic.
  - *Disk cache (`flat-cache`)*: Unsuitable for live, high-frequency intra-session ticks that mutate continuously throughout the active session.

---

### Decision 3: Single-Read Descending Query for Off-Hours Session Backtracking
- **Decision**: When today's in-memory snapshot has no points (weekends, exchange holidays, pre-market morning hours), the system executes a single descending query against Firestore:
  ```typescript
  db.collection(collectionName)
    .orderBy(FieldPath.documentId(), 'desc')
    .limit(1)
    .get();
  ```
- **Rationale**: Daily intraday collection documents are keyed by ISO calendar dates (`YYYY-MM-DD`). In ASCII, ISO date strings sort chronologically. A descending query with `limit(1)` resolves the most recent populated session date in exactly 1 read, regardless of whether yesterday was Friday, Saturday, Sunday, or a multi-day holiday. The result is buffered in `IntradaySnapshotService.setFallbackSnapshot` to serve subsequent off-hours requests with 0 additional database reads.
- **Alternatives Considered**:
  - *Sequential day-by-day probing (`today - 1`, `today - 2`, ...)*: Requires up to 4 sequential Firestore reads on a Sunday morning (checking Saturday, Friday, Thursday), tripling database read consumption.
  - *Static calendar date deduction via market registry*: Exchange schedules can vary due to unscheduled national holidays, mourning periods, or weather closures; querying actual recorded documents guarantees the API serves data that truly exists.

---

### Decision 4: Centralization in `IntradayCoordinator`
- **Decision**: Expand `IIntradayCoordinator` in `api/src/services/intraday-coordinator.ts` with a centralized method `resolveLatestTimeline(options)` to coordinate in-memory checks, descending Firestore queries, fallback buffering, and 7-day lookback bounds.
- **Rationale**: In accordance with Constitution Principle XV (Modular Architecture) and Feature 081 (`081-modular-intraday-logic`), intraday extraction, point normalization, and session rollover logic must live in a single centralized coordinator rather than being duplicated across `StockQueryService`, `IndicesQueryService`, `GoldQueryService`, and `RatesQueryService`.
- **Alternatives Considered**:
  - *Duplicating backtracking logic across query services*: Strongly rejected due to code drift and regression risk.

---

### Decision 5: Adaptive HTTP Caching Policy
- **Decision**: Return adaptive `Cache-Control` response headers based on session state:
  - Active Session (today during open hours): `public, max-age=60`
  - Concluded Session (off-hours / weekend / past date): `public, max-age=900` (15 minutes)
- **Rationale**: Active sessions require short TTLs to show near-real-time quote updates, while concluded historical sessions are immutable until the next market open and can safely be cached longer on reverse proxies and client devices.

---

### Decision 6: Zero-Trade Inactive Security Resiliency
- **Decision**: When a listed security exists in the catalog but recorded 0 price updates in the resolved session, return HTTP `200 OK` with `intraday: []` and its latest known closing price.
- **Rationale**: Mobile charting screens bind both the header card (company name, price, change) and the chart canvas. Returning 404 for an inactive stock breaks the entire screen, whereas returning `intraday: []` allows the client to display the quote metrics with a flat or empty chart gracefully.
