# Phase 0 Research: Stateless Batch Quotes API Endpoint

**Feature Branch**: `098-batch-quotes-api`
**Date**: 2026-10-31

## Context & Objectives
Mobile clients currently need to issue 4–5 individual network calls (`/v1/rates`, `/v1/gold`, `/v1/indices`, `/v1/stocks`, plus GCC endpoints) to hydrate a user's locally stored favorites dashboard. This produces avoidable mobile radio active time, increased cellular latency, and potential partial UI loading states. 

The goal of this feature is to introduce a stateless, unified Backend-for-Frontend (BFF) quotes endpoint that resolves a batch of mixed asset identifiers in a single round-trip with zero database reads.

---

## Technical Decisions & Findings

### Decision 1: Domain Identifier Parsing & Disambiguation Strategy

- **Decision**: Deterministic category-based pattern matching with priority resolution order.
  - **Pattern 1: Bank Rates**: Matched if string contains an underscore dividing a bank code and currency code (`^[A-Z0-9]+_[A-Z]{3}$` case-insensitive), e.g. `CBE_USD`, `MISR_EUR`, `NBE_SAR`.
  - **Pattern 2: Gold Commodities**: Matched against canonical gold keys (`24k`, `21k`, `18k`, `pound`, `ounce` case-insensitive).
  - **Pattern 3: Market Indices**: Matched against index registry (`getIndexEntry(id)`) or canonical index codes (`EGX30`, `EGX70`, `EGX100`, `TASI`, etc.).
  - **Pattern 4: Equities / Stocks**: Matched against EGX equity snapshot symbols (e.g. `COMI`, `FWRY`, `EAST`) and regional GCC stock catalogs/snapshots (e.g. `1120`, `2222`, `EMAAR`).
- **Rationale**: Clean regex and dictionary keys avoid ambiguous clashes. Bank currencies have unique underscore separators; gold karats have well-defined naming; indices and stocks are indexed by unique alphanumeric symbols.
- **Alternatives Considered**:
  - *Explicit prefixed strings (e.g., `rate:CBE_USD`, `gold:21k`, `stock:COMI`)*: Rejected because client mobile apps already store raw symbol strings in local databases, and requiring prefixes would break backward compatibility with stored user favorites.

### Decision 2: Zero-Read In-Memory Resolution Architecture

- **Decision**: Resolve all quotes strictly from the existing in-memory service instances managed by `FirestoreListenerService` (`RatesSnapshotService`, `GoldQueryService`, `IndicesQueryService`, `StockQueryService`, `GccStockQueryService`).
- **Rationale**:
  - `FirestoreListenerService` already maintains active push listeners (`onSnapshot`) for bank rates, gold, EGX stocks, EGX indices, and GCC stocks.
  - Querying these in-memory stores executes synchronously in sub-millisecond time ($<1\text{ ms}$) without invoking Firestore `.get()` reads.
  - Complies strictly with Constitution Principle VII (Zero Cloud Cost) and Principle XIII (Intelligent Caching).
- **Alternatives Considered**:
  - *Direct Firestore queries*: Rejected because issuing multi-document reads on every mobile dashboard load would quickly breach Firestore's 50,000 free daily read quota.
  - *Flat-cache disk reads*: Rejected because live prices are already retained in RAM; disk I/O would add unnecessary latency.

### Decision 3: HTTP Protocol & Payload Schema

- **Decision**:
  - Support both `POST /v1/quotes` (body: `{"items": ["..."]}`) and `GET /v1/quotes?items=CBE_USD,21k,EGX30`.
  - Enforce max batch size of 100 items per request.
  - Return quotes in a key-value dictionary (`quotes: { [id]: QuoteItem }`) alongside an `unresolved: string[]` list and a server `timestamp`.
- **Rationale**:
  - Satisfies user requirement A1 (keyed dictionary for $O(1)$ client lookup), A2 (POST and GET support for caching and testing), and A3 (isolated `unresolved` array for partial success).
  - Capping batch size at 100 items prevents memory exhaustion or abuse.
- **Alternatives Considered**:
  - *Returning pure array*: Rejected by user decision A1; requires client to iterate through list to locate specific favorites.
  - *Failing the whole batch on one bad item (all-or-nothing)*: Rejected because mobile local storage may hold obsolete IDs, and crashing the entire home screen is unacceptable UX.

### Decision 4: Price Change & Freshness Calculation

- **Decision**:
  - **Bank rates**: Calculate change from yesterday's closing rate if available in `intradayCoordinator` or snapshot history; otherwise return current buy/sell rates.
  - **Gold**: Calculate change from yesterday's price (`yesterdayBuyPrice` / `yesterdaySellPrice`).
  - **Indices**: Extract `change` and `changePercent` directly from index snapshot data.
  - **Stocks**: Extract `change` and `changePercent` directly from live equity snapshot data.
- **Rationale**: Standardizes quote metrics across different asset classes into a uniform `QuoteItem` contract.

---

## Summary of Architectural Best Practices
1. **Validation**: Use Zod schemas to validate `items` array format and bounds ($1 \le \text{length} \le 100$).
2. **Defensive Isolation**: Wrap each item resolution step in try/catch blocks; failure in one domain must never affect another.
3. **HTTP ETag & Cache-Control**: Emit `Cache-Control: public, max-age=15` and weak ETags for `GET` requests to benefit mobile caching.
