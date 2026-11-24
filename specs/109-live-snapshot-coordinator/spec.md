# Feature Specification: Live Snapshot Coordinator

**Feature Branch**: `109-live-snapshot-coordinator`

**Created**: 2026-11-15

**Status**: Implemented

**Input**: User description: "Deepen Live Snapshot In-Memory Storage into LiveSnapshotCoordinator. Involved Modules: src/services/rates-snapshot-service.ts • src/cache/snapshot-cache.ts • src/cache/snapshot-store.ts • src/services/gold-query-service.ts • src/services/gas-query-service.ts • src/services/stock-query-service.ts • src/services/indices-query-service.ts • src/modules/gcc-stocks/services/gcc-stock-query-service.ts • src/services/firestore-listener-service.ts • src/services/quotes-service.ts. Problem: Every asset domain maintains its own ad-hoc in-memory cache slot and duplicate Firestore fallback logic, forcing listeners, quotes services, and intraday engines to inject and juggle 6 separate services. Solution: Consolidate live quote storage and push ingestion into a single deep LiveSnapshotCoordinator module behind an expressive three-method interface (getSnapshot, updateSnapshot, fetchFresh)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Universal Real-Time Quote Retrieval (Priority: P1)

Downstream API callers, route handlers, batch quotes resolution pipelines, and intraday timeline engines require instantaneous access to the current live price snapshot for any supported financial asset category (foreign currency exchange rates, gold karats and bullion, fuel gas products, domestic EGX equities, regional GCC equities across Saudi Arabia, UAE, Qatar, Kuwait, and benchmark indices) through a single unified query interface, receiving verified high-freshness quotes purely from in-memory storage without needing to couple to domain-specific query service instances.

**Why this priority**: Instant live quote resolution is the central gateway capability for every real-time consumer screen. Centralizing this operation behind a single interface eliminates disjointed per-domain cache lookups and ensures uniform sub-millisecond delivery.

**Independent Test**: Can be verified by issuing snapshot retrieval requests across all asset categories (rates, gold, gas, stocks, indices, and regional GCC equities) and confirming that each domain returns the current active price snapshot with standard freshness metadata in under 1 millisecond.

**Acceptance Scenarios**:

1. **Given** a warm in-memory cache, **When** an application requests the latest price snapshot for a registered asset domain (e.g., currency rates, gold karats, or Egyptian equities), **Then** the coordinator returns the active price snapshot with freshness indicators and source timestamps.
2. **Given** a request specifying a regional market (e.g., Saudi Tadawul or UAE equities), **When** retrieved from the coordinator, **Then** the market-specific live price snapshot is returned accurately.
3. **Given** a query for an unsupported or unknown asset domain, **When** evaluated by the coordinator, **Then** a structured error is returned indicating the domain is unrecognized without throwing unhandled exceptions.

---

### User Story 2 - Push-Based Real-Time Snapshot Ingestion (Priority: P1)

Background real-time event listeners and database push subscriptions need to ingest live price updates into a single centralized in-memory storage coordinator, instantly updating cached market prices across all asset classes with zero polling delay, eliminating the need to manage separate update routines across 6 distinct query services.

**Why this priority**: Real-time push updates are required to maintain zero-read operation under active HTTP serving conditions. A single ingestion interface eliminates synchronization gaps and circular dependencies between background listeners and domain query modules.

**Independent Test**: Can be verified by pushing mock snapshot updates into the coordinator for any domain and verifying that subsequent read queries immediately return the updated data with refreshed timestamps.

**Acceptance Scenarios**:

1. **Given** an active real-time data stream, **When** an upstream push update delivers a new price snapshot for an asset domain, **Then** the coordinator replaces the active in-memory snapshot and subsequent reads immediately reflect the updated prices.
2. **Given** an incoming push update containing malformed or empty price collections, **When** ingested by the coordinator, **Then** the coordinator rejects the corrupted payload, preserves the existing valid snapshot, and flags a degraded status marker.
3. **Given** simultaneous real-time updates arriving for different asset domains, **When** processed by the coordinator, **Then** all domains update concurrently without memory race conditions or cross-domain interference.

---

### User Story 3 - Resilient On-Demand Warming and Fail-Safe Fallbacks (Priority: P2)

When the API service boots up cold or encounters a cache miss during early initialization, the coordinator must transparently hydrate its in-memory storage from local persistent snapshot files or on-demand upstream queries, ensuring zero service disruption, sub-second cold starts, and seamless graceful degradation during upstream network outages.

**Why this priority**: High availability and crash resilience require that services never serve 500 errors during startup or transient upstream connectivity disruptions.

**Independent Test**: Can be verified by clearing all in-memory caches, initiating quote requests, and confirming that the coordinator automatically warms from persistent disk storage or upstream sources and returns valid snapshots.

**Acceptance Scenarios**:

1. **Given** a cold startup with empty memory caches, **When** a client requests a quote snapshot, **Then** the coordinator transparently executes an on-demand retrieval, populates in-memory storage, and returns the snapshot with complete pricing data.
2. **Given** an upstream database timeout or network disruption during an on-demand refresh, **When** a stale or persisted snapshot is available in memory or on disk, **Then** the coordinator gracefully returns the last known snapshot marked with degraded freshness rather than failing the client request.
3. **Given** multiple concurrent client requests arriving while an asset domain's cache is warming from an upstream source, **When** handled by the coordinator, **Then** concurrent fetches are coalesced into a single in-flight operation preventing redundant upstream requests.
4. **Given** an explicit force-refresh request by an authorized caller, **When** received, **Then** the coordinator bypasses in-memory caching, retrieves a fresh upstream snapshot, updates the in-memory store, and returns the fresh payload.

---

### User Story 4 - Multi-Domain Batch Quote Resolution (Priority: P2)

Client applications requesting mixed watchlists containing diverse financial assets (such as USD bank rates, 21K gold, the EGX30 index, and regional Saudi equities) receive all requested quotes through an atomic resolution pipeline that queries the unified live snapshot coordinator, eliminating sequential multi-service fan-outs.

**Why this priority**: Mobile apps display unified dashboard overviews combining currencies, gold, indices, and equities in a single screen. Resolving these from one memory store eliminates multi-hop latency and simplifies consumer aggregation.

**Independent Test**: Can be verified by passing a mixed list of asset identifiers (currency codes, commodity names, equity tickers, and index codes) to the batch quotes resolver and verifying that all items resolve with exact live pricing in a single response envelope.

**Acceptance Scenarios**:

1. **Given** a batch request containing asset symbols across foreign exchange, gold, stock indices, and equities, **When** evaluated by the quotes pipeline, **Then** all recognized symbols are resolved from the live snapshot coordinator and returned in a consolidated response.
2. **Given** a batch request where some symbols are unregistered or currently inactive, **When** processed, **Then** all valid symbols are resolved successfully and invalid symbols are collected in a clear unresolved list without failing the entire batch.

---

### Edge Cases

- **Cold Boot Before Listeners Connect**: During container startup before background real-time push listeners establish active WebSocket connections, the coordinator must load disk-persisted snapshots synchronously or on demand, guaranteeing zero 500 error responses on early health probes and API calls.
- **Upstream Network Failures during Refresh**: If a force-refresh request encounters an upstream timeout or connection drop, the coordinator must retain the previous valid in-memory snapshot and return it with degraded freshness indicators (`freshness: "degraded"`).
- **Thundering Herd on Cache Miss**: When an un-cached domain experiences a burst of simultaneous incoming queries, the coordinator must coalesce in-flight retrieval promises so that exactly one upstream query is dispatched.
- **Regional Disambiguation**: Where asset categories span multiple sovereign jurisdictions (such as equities and indices across Egypt, Saudi Arabia, UAE, Qatar, and Kuwait), market qualification parameters must isolate and return the appropriate regional market state without symbol collision.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a single, centralized in-memory live snapshot coordinator encapsulating active price snapshots across all supported financial domains (foreign currency exchange rates, gold karats and bullion, fuel gas products, domestic EGX stocks, regional GCC stocks, and benchmark indices).
- **FR-002**: The system MUST expose a standardized query operation (`getSnapshot`) that delivers the active in-memory price snapshot for any requested asset domain and regional market with sub-millisecond retrieval latency.
- **FR-003**: The system MUST expose a centralized push ingestion operation (`updateSnapshot`) allowing real-time event listeners to push fresh price snapshots directly into memory across all domains.
- **FR-004**: The system MUST expose an on-demand retrieval operation (`fetchFresh`) that queries upstream persistent storage to hydrate or force-refresh memory when requested or uninitialized.
- **FR-005**: The system MUST encapsulate fail-safe degraded execution, serving the last known valid snapshot with appropriate freshness indicators when upstream retrieval fails.
- **FR-006**: The system MUST support persistent disk snapshots to enable immediate zero-latency warming on process cold boot before network listeners connect.
- **FR-007**: The system MUST deduplicate concurrent in-flight upstream fetches for the same domain during cache miss scenarios.
- **FR-008**: The system MUST eliminate per-domain ad-hoc cache variables and duplicate fallback logic across individual query services, establishing a single authoritative in-memory quote seam.

### Key Entities *(include if feature involves data)*

- **Live Snapshot Envelope**: Normalized container representing the current active market pricing state for an asset domain, containing domain identifier, optional market code, currency, last updated timestamp, retrieval timestamp, freshness status (`current`, `stale`, `degraded`), source trace identification, and domain-specific price dictionary.
- **Asset Domain**: Canonical classification of financial assets managed by the coordinator, including `rates` (foreign exchange), `gold` (precious metals), `gas` (fuel commodities), `stocks` (domestic equities), `indices` (benchmark market indices), and `gcc-stocks` (regional equities).
- **Freshness State**: Machine-readable indicator specifying the timeliness of an active snapshot (`current` when fresh, `stale` when aged, or `degraded` when serving a fail-safe fallback).
- **Batch Quote Item**: Normalized individual price quote record resolved from the active snapshot, containing symbol, canonical title, current price/rate, previous close/old price, net change, percentage change, and visual logo URL.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of live quote retrievals for warm domains resolve from in-memory storage in under 1 millisecond without database reads.
- **SC-002**: Cold boot recovery hydrates active snapshots from local persistence or authoritative fallback within 500 milliseconds of service startup.
- **SC-003**: Elimination of redundant cache management logic across 6 separate query service modules, consolidating all live state behind a single interface.
- **SC-004**: Real-time push listeners depend on exactly one centralized snapshot ingestion interface instead of maintaining connections to 6 separate domain services.
- **SC-005**: 100% of existing contract tests and unit tests continue to pass without contract regressions or schema drift.

## Assumptions

- The set of supported asset domains (currencies, gold, gas, domestic stocks, regional stocks, benchmark indices) is bounded and known at runtime.
- In-memory snapshots are held in Node.js process RAM and require less than 15MB total heap allocation across all active domains.
- Real-time push listeners via Firestore WebSocket streams remain the primary ingestion mechanism for sub-second quote freshness.
- Existing REST endpoint contracts (`/v1/rates`, `/v1/gold`, `/v1/gas`, `/v1/stocks`, `/v1/indices`, `/v1/quotes`) maintain 100% backwards-compatible JSON response shapes.
