# Feature Specification: Unify Real-Time Ingestion and Active Weekly Caching

**Feature Branch**: `115-unify-realtime-ingestion`  
**Created**: 2026-11-16  
**Status**: Draft  
**Input**: User description: "Unify Real-Time Ingestion and Active Weekly Caching (Eliminate ActiveWeeklyBucketStore & Shims)"

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Streamlined Real-Time Market Quote Ingestion (Priority: P1)

As a consumer of real-time financial market data (bank rates, precious metals, petroleum, equity securities, and benchmark indices), I want updates pushed from the upstream database to be applied directly to the authoritative live and intraday data stores so that live endpoints immediately reflect new market prices without redundant intermediate updates or circular caching delays.

**Why this priority**: Real-time accuracy is the core product promise of the API gateway. Ingestion must be lean, deterministic, and direct without passing through circular wrappers or multiplexers.

**Independent Test**: Can be tested independently by pushing simulated asset snapshots and verifying that all asset endpoints immediately serve the updated quotes in sub-millisecond memory lookups with zero redundant internal mutations.

**Acceptance Scenarios**:
1. **Given** an upstream real-time price change occurs for an asset domain, **When** the real-time listener receives the update notification, **Then** the payload is dispatched directly to the authoritative domain coordinator without routing through legacy intermediate wrappers.
2. **Given** multiple asset price updates occur concurrently, **When** the updates are received, **Then** each asset domain processes its update independently and idempotently without triggering duplicate refresh cycles.

---

### User Story 2 - Authoritative Active Calendar Week Historical Resolution (Priority: P2)

As an analytical chart or historical data client, I want multi-day and single-date historical queries for the active calendar week to be served directly from the historical engine's primary in-memory tier so that active week queries achieve instant sub-millisecond response times without relying on an external intermediate memory bucket store.

**Why this priority**: Active week historical queries represent the highest volume of charting queries during operating hours. Eliminating the intermediate memory store and unifying Tier 1 memory storage directly inside the historical resolution engine strengthens data locality and eliminates cross-service synchronization gaps.

**Independent Test**: Can be tested independently by querying active week historical single-date and date-range endpoints and confirming they resolve accurately from memory in < 150ms without initiating external database reads.

**Acceptance Scenarios**:
1. **Given** a query for a trading date within the current active calendar week, **When** the historical query is received, **Then** the engine retrieves the daily closing quote directly from its primary in-memory tier.
2. **Given** a weekly document update is pushed from the database for the active week, **When** the update is delivered, **Then** the historical engine updates its in-memory active week tier atomically.
3. **Given** a calendar week transition occurs at the weekly rollover boundary, **When** the new calendar week commences, **Then** the previous week is sealed and persisted to disk storage while the new week begins tracking in memory.

---

### User Story 3 - Clean Seam Boundary & Legacy Facade Removal (Priority: P3)

As a system maintainer and developer, I want all obsolete compatibility shims, dead facades, and artificial single-purpose wrappers removed from the codebase so that the system presents clean, unambiguous architectural seams that are easy to navigate, understand, and test.

**Why this priority**: Eliminating dead code and pass-through shims reduces maintenance friction, prevents developer confusion, and guarantees that unit tests evaluate real system interfaces rather than pass-through abstractions.

**Independent Test**: Can be tested by running the entire unit and integration test suites, ensuring that all 49 platform verification scenarios pass and no legacy shim classes or dead functions remain in use.

**Acceptance Scenarios**:
1. **Given** callers and test suites interacting with live market coordinators, **When** querying or warming snapshots, **Then** all callers interact directly with the authoritative coordinator rather than legacy cache facades.
2. **Given** services querying trading schedule calendars and operational hours, **When** checking market status, **Then** services use the authoritative schedule engine directly without referencing dead intermediate facades.

---

### Edge Cases

- **Cold Boot Before First Real-Time Push**: What happens if the gateway boots before the database push listener delivers the first active week document? The system must fall back to persisted disk cache or cold-fetch the active week document on-demand and seed its primary in-memory tier.
- **Weekly Boundary Transition**: What happens when a request arrives exactly as the active calendar week rolls over? The historical engine must gracefully finalize the prior week, promote the new week key as active, and prevent concurrent read stampedes.
- **Transient Listener Disconnection**: How does the system handle temporary upstream connection drops? In-memory state remains intact and continues serving cached quotes, and listeners re-attach with clean state synchronization upon reconnection.
- **Missing or Empty Weekly Document**: What happens if an upstream document snapshot is received empty or deleted? The in-memory tier must clear or mark the affected week as uninitialized without throwing uncaught exceptions.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST ingest real-time market snapshots from the database and deliver them directly to the primary live snapshot coordinator without passing through redundant intermediate query service delegates.
- **FR-002**: The system MUST ingest active calendar week historical bucket documents and store them directly in the primary in-memory tier of the historical resolution engine.
- **FR-003**: The historical resolution engine MUST manage the active calendar week in-memory state internally, supporting atomic updates, retrieval, deletion, and cold-boot pre-seeding from persistent disk storage.
- **FR-004**: The real-time listener adapter MUST require only the direct core domain coordinators for live snapshots, intraday timelines, and historical resolution, eliminating circular or duplicate cache updates.
- **FR-005**: The system MUST eliminate artificial pass-through shims and dead facades that wrap underlying coordinators without adding domain behavior.
- **FR-006**: In-memory active week lookups and live snapshot queries MUST resolve with zero (0) external database read operations under warmed operating conditions.
- **FR-007**: Existing public API contracts and JSON response shapes across all 49 platform verification endpoints MUST remain 100% backward-compatible with zero structural drift.
- **FR-008**: All modified and newly authored code lines MUST include 100% comprehensive documentation comments detailing logic, rationale, and runtime behavior.
- **FR-009**: The application version MUST be incremented in compliance with semantic versioning prior to completion.

---

### Key Entities

- **Active Calendar Week Bucket**: An in-memory dictionary mapping ISO calendar dates (`YYYY-MM-DD`) to daily asset closing summaries for the current active ISO week (`YYYY-Www`).
- **Inbound Real-Time Push Adapter**: A lightweight transport adapter that subscribes to external database push events and dispatches unwrapped data payloads directly to domain coordinators.
- **Live Snapshot Coordinator**: The centralized authoritative in-memory and disk persistence coordinator responsible for live market quote envelopes across all asset classes.
- **Historical Resolution Engine**: The unified multi-tiered coordinator responsible for single-date archives and multi-day ranges across all asset domains.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of real-time push updates are reflected immediately across all asset endpoints without duplicate internal update calls.
- **SC-002**: Warmed active week historical and live quote queries resolve from memory in under 150ms latency (maintaining SLA ceilings).
- **SC-003**: 100% of existing unit tests (324+ tests) and integration test suites pass with zero regressions.
- **SC-004**: All 49 platform deployment verification scenarios in `scripts/verify/` pass with exit code `0`.
- **SC-005**: At least 4 shallow/shim/dead code files are safely deleted, eliminating over 200 lines of redundant indirection without shifting complexity to callers.

---

## Assumptions

- **Timezone Consistency**: The active calendar week boundary continues to follow the Africa/Cairo business calendar and standard ISO week numbering (`YYYY-Www`).
- **Disk Cache Format**: Sealed past weekly buckets continue to be persisted in local disk cache (`flat-cache`) with standard 30-day retention policies.
- **No Client Contract Changes**: Mobile and web clients require no endpoint or schema changes; this refactor strictly deepens internal module architecture.
- **Core Engine Availability**: `LiveSnapshotCoordinator`, `HistoryResolutionEngine`, and `IntradayCoordinator` remain the three authoritative domain engines of the platform.
