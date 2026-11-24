# Feature Specification: Dissolve Shadow Caches and Firestore Fallbacks in Query Modules

**Feature Branch**: `112-dissolve-shadow-caches`  
**Created**: 2026-11-15  
**Status**: Draft  
**Input**: User description: "Dissolve Shadow Caches and Firestore Fallbacks in Query Modules. Problem: Domain query modules maintain shadow cache fields and duplicate direct Firestore fallback reads already handled by LiveSnapshotCoordinator. Solution: Strip shadow caches and duplicate fallback logic from query modules, delegating all quote reads directly to LiveSnapshotCoordinator."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single-Authority Quote Resolution in Batch Quotes (Priority: P1) 🎯 MVP

When client applications and background consumers request real-time multi-asset market quotes via the batch quotes endpoint, the system resolves quotes from a single authoritative snapshot store without initializing or invoking redundant domain-specific query service instances.

**Why this priority**: Eliminates 5 redundant injected dependencies in the batch quotes subsystem, ensures memory consistency across all quote lookups, and provides immediate MVP value by establishing a clean, unified quote retrieval seam.

**Independent Test**: Request quotes across multiple asset categories (currency rates, gold carats, equities, and indices) via the batch quotes service; verify that all requested quotes resolve with accurate prices and zero duplicate snapshot instantiation.

**Acceptance Scenarios**:

1. **Given** requests for arbitrary multi-asset identifiers (e.g. `CBE_USD`, `GOLD_21`, `COMI`, `EGX30`), **When** batch quotes are resolved, **Then** all live prices return accurately from the central snapshot store without delegating to secondary domain query instances.
2. **Given** an unpopulated or missing live quote in memory, **When** batch quotes query the central snapshot store, **Then** the request gracefully falls back to degraded or disk-persisted quotes without triggering redundant uncoordinated database reads.

---

### User Story 2 - Dissolve Shadow Caches in Commodity and Fuel Services (Priority: P2)

When API consumers fetch latest prices for domestic gold carats and petroleum products, the commodity and fuel query modules read live prices directly from the central snapshot coordinator rather than maintaining private shadow cache variables and isolated database fallback queries.

**Why this priority**: Gold and fuel query modules currently duplicate in-memory caching fields and redundant direct database fallback routines that consume unnecessary memory and bypass central single-flight concurrency safeguards.

**Independent Test**: Invoke gold and gas query endpoints under cold cache and concurrent load conditions; verify identical price payloads with zero shadow cache fields and unified database fallback coordination.

**Acceptance Scenarios**:

1. **Given** a request for latest gold prices, **When** query execution occurs, **Then** prices return from the central snapshot store without maintaining duplicate local cache fields.
2. **Given** a request for latest petroleum prices, **When** query execution occurs, **Then** fuel rates return from the central snapshot store with zero boilerplate database fallback code in the fuel service.
3. **Given** real-time listener push updates for gold or fuel, **When** new market rates arrive, **Then** updates mutate only the central snapshot store without synchronizing redundant local service variables.

---

### User Story 3 - Dissolve Shadow Caches in Equity and Benchmark Index Services (Priority: P3)

When API consumers fetch latest equity prices and benchmark index levels for domestic and regional markets, the equity and index query services delegate 100% of live snapshot retrieval to the central snapshot coordinator, eliminating local dictionary shadow caches and uncoordinated database gets.

**Why this priority**: Ensures all domestic (EGX) and regional (Tadawul, DFM, ADX, QSE, KSE) stock and index queries hit a single consolidated store, guaranteeing data consistency and preventing duplicate database reads across concurrent requests.

**Independent Test**: Query latest stock prices and benchmark index values across domestic and regional markets; verify that all live quotes are served from the central snapshot store and that constituent stock lookups reuse central snapshots.

**Acceptance Scenarios**:

1. **Given** a request for latest stock prices, **When** the query service executes, **Then** live stock quotes resolve from the central snapshot store with automatic catalog metadata enrichment.
2. **Given** a request for benchmark index levels, **When** the query service executes, **Then** index levels resolve from the central snapshot store without querying local duplicate caches.
3. **Given** concurrent requests for stock and index quotes during cold cache conditions, **When** database retrieval is needed, **Then** all concurrent requests coalesce into a single coordinated read through central single-flight protection.

---

### Edge Cases

- **Cold System Boot with Unhydrated Memory**: When the gateway boots cold before any background listener delivers updates, requests to any query service must seamlessly recover from disk snapshots or perform a single coordinated upstream fetch without raising unhandled errors.
- **Concurrent Identical Fetch Requests (Thundering Herd)**: When multiple concurrent client requests hit cold cache simultaneously for the same asset domain, all requests must coalesce into a single upstream read via the central coordinator rather than each domain service firing its own database query.
- **Degraded Execution During Database Outages**: When upstream database connectivity is interrupted, all query services must serve stale or disk-persisted fallback snapshots with appropriate degraded freshness headers rather than failing with server 500 errors.
- **Partial or Malformed Snapshot Payloads**: When incoming push updates or database documents contain partial or malformed rate entries, the central coordinator must defensively validate the payload before ingestion, preventing corruption across all dependent query endpoints.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST consolidate all live market quote retrieval across all financial domains behind the single central snapshot coordinator.
- **FR-002**: System MUST eliminate private shadow cache fields (`cachedPrices`, `cachedRegionalPrices`) from individual domain query modules (`StockQueryService`, `GoldQueryService`, `GasQueryService`, `IndicesQueryService`).
- **FR-003**: System MUST eliminate duplicate direct database fallback queries from domain query modules, routing on-demand fetches through central coordinator routines.
- **FR-004**: System MUST simplify the quotes aggregation service by removing unnecessary domain service dependencies and injecting only the central snapshot coordinator.
- **FR-005**: System MUST ensure that real-time listener push updates hydrate only the central snapshot store without requiring secondary synchronization with domain query modules.
- **FR-006**: System MUST maintain 100% backward compatibility for all existing public API responses, query filters, catalog enrichments, and sorting parameters.
- **FR-007**: System MUST provide single-flight promise pooling for all on-demand domain snapshot fetches to prevent duplicate concurrent database reads under load.
- **FR-008**: System MUST preserve stale fallback capabilities from persistent disk storage across all query services during network interruptions.

---

### Key Entities *(include if feature involves data)*

- **Live Market Snapshot**: Standardized, domain-agnostic envelope containing live market rates, currency qualifiers, freshness state, telemetry trace identifiers, and source timestamps.
- **Batch Quote Request**: Collection of arbitrary financial asset identifiers across banking, precious metals, equities, and benchmark indices to be resolved into standardized quote items.
- **Domain Query Adapter**: Lightweight, specialized module responsible for domain-specific query parameter validation, filtering, and response serialization without maintaining independent in-memory state.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of live market quote queries resolve from the single authoritative snapshot coordinator, eliminating all shadow cache storage fields across domain services.
- **SC-002**: Zero duplicate database reads occur during concurrent cache misses for the same asset domain, verified by single-flight promise pooling.
- **SC-003**: Injected dependency count in the quotes service decreases from 6 dependencies to exactly 1 authoritative dependency.
- **SC-004**: 100% backward compatibility and response parity across all 30 public API gateway endpoints, verified with 0 schema or value regressions.

---

## Assumptions

- The central snapshot coordinator introduced in Feature 109 is fully functional, supports all asset domains (`rates`, `gold`, `gas`, `stocks`, `indices`, `gcc-stocks`), and maintains persistent disk caching and single-flight pooling.
- Historical date range queries and weekly bucket slicing remain the responsibility of the history resolution engine and are not impacted by live snapshot cache dissolution.
- Catalog enrichment and index constituent filtering remain domain responsibilities delegated to the unified market catalog module.
