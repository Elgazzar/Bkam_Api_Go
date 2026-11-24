# Feature Specification: Unified Market Catalog

**Feature Branch**: `108-unified-market-catalog`

**Created**: 2026-11-15

**Status**: Draft

**Input**: User description: "3. Deepen Catalog Module into Unified Market Catalog. Involved Modules: src/services/catalog-service.ts, src/config/stock-catalog.ts, src/config/index-catalog.ts, src/config/market-registry.ts, src/modules/gcc-stocks/catalogs/*.ts, src/services/indices-query-service.ts, src/services/stock-query-service.ts, src/routes/stock-routes.ts, src/routes/indices-routes.ts. Problem: The existing Catalog module is a shallow 10-line module leaving static inventories raw, forcing callers to duplicate symbol normalization, index constituent scanning, and validation loops. Solution: Deepen Catalog into a single module that encapsulates all platform inventory behind expressive operations for lookup, constituent membership, snapshot enrichment, and validation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Universal Security and Asset Resolution (Priority: P1)

Downstream API callers, quote processors, and history resolution engines need to discover and look up any supported security (domestic EGX equities, regional GCC equities across Saudi Arabia, UAE, Qatar, and Kuwait), currency, bank, commodity, or benchmark index by its canonical code, ticker symbol, or alias through a single query interface without requiring prior knowledge of which internal catalog partition houses the asset.

**Why this priority**: Fast, unified asset identification is the prerequisite for all quote routing, history fetching, and validation across the platform.

**Independent Test**: Can be verified by querying domestic and regional symbols, numeric ticker codes, currency codes, and benchmark indices with varying case and aliases, verifying that each query resolves the correct canonical metadata.

**Acceptance Scenarios**:

1. **Given** a valid security ticker symbol (e.g., `"COMI"`, `"1010"`, or `"QNBK"`), **When** a lookup request is issued, **Then** the catalog returns the canonical security definition including official symbol, localized names (Arabic and English), exchange, currency, and sector.
2. **Given** a symbol with alternative aliasing or formatting (e.g., lowercase `"comi"`, padded `" 1010 "`, or legacy suffix `"COMI.CA"`), **When** a lookup request is issued, **Then** the catalog normalizes the input and resolves the exact canonical security.
3. **Given** an unknown or unsupported symbol or asset identifier, **When** a lookup request is issued, **Then** the catalog clearly indicates that the asset does not exist without throwing internal unhandled exceptions.

---

### User Story 2 - Benchmark Index Constituent Resolution (Priority: P1)

Mobile applications and institutional clients tracking market indices (such as EGX30, EGX70, TASI, ADX15, DFMGI, QE20) require an authoritative, instant list of all constituent member securities belonging to any benchmark index, including member weighting, sector grouping, and display metadata.

**Why this priority**: Index constituent tracking is a primary user journey for mobile market overview screens. Currently, callers execute multi-hop nested iterations across disparate catalog arrays.

**Independent Test**: Can be verified by requesting constituents for a known benchmark index (e.g., `"EGX30"` or `"TASI"`) and confirming that the response returns the complete, correctly structured set of member securities.

**Acceptance Scenarios**:

1. **Given** a recognized benchmark index code (e.g., `"EGX30"` or `"DFMGI"`), **When** constituents are requested, **Then** the catalog returns all member securities currently registered to that index, along with their canonical tickers, English and Arabic names, and exchange details.
2. **Given** an index code from a specific regional market, **When** constituents are requested, **Then** only securities belonging to that exchange and index membership are returned.
3. **Given** an invalid or unrecognized index code, **When** constituents are requested, **Then** the catalog rejects the request with an unambiguous indication that the index code is invalid.

---

### User Story 3 - Snapshot Metadata and Visual Asset Enrichment (Priority: P2)

Real-time price streams and historical quote archives contain raw trading figures (e.g., symbol, price, change, volume) but often omit presentation metadata such as corporate branding logos, localized company names, sectors, or benchmark index tags. Downstream endpoints require consistent, automated enrichment of raw quote snapshots before delivery to clients.

**Why this priority**: Eliminates repetitive mapping loops across all query services, guaranteeing that client applications receive standardized logos and localized names across every asset class.

**Independent Test**: Can be verified by passing a bare quote record containing only a symbol and price, and confirming that the returned payload is enriched with official logo URLs, localized Arabic/English names, sector, and associated benchmark indices.

**Acceptance Scenarios**:

1. **Given** a raw price snapshot for a registered equity or bank, **When** metadata enrichment is requested, **Then** the catalog attaches the verified logo URL, canonical names, and applicable index affiliations while preserving original pricing metrics.
2. **Given** a price snapshot for an asset without a registered logo, **When** metadata enrichment is requested, **Then** the catalog provides a standard fallback logo reference without corrupting the snapshot envelope.

---

### User Story 4 - Centralized Domain Validation and Parameter Filtering (Priority: P2)

HTTP route handlers validating incoming query parameters (such as bank codes, currency codes, market exchange identifiers, stock groups, and index codes) require an authoritative validation mechanism to ensure malformed or out-of-scope parameters are rejected at the edge before triggering unnecessary storage or cache operations.

**Why this priority**: Consolidates validation rules currently scattered across route handlers, ensuring uniform error messaging and boundary enforcement.

**Independent Test**: Can be verified by evaluating valid and invalid parameters across all asset categories and confirming that validation outcomes and error classifications are consistent.

**Acceptance Scenarios**:

1. **Given** an incoming request with query parameters (e.g., `bank=cib`, `currency=USD`, `market=saudi`), **When** parameter validation is executed, **Then** the catalog confirms validity and outputs normalized parameter representations.
2. **Given** an invalid or out-of-catalog parameter value (e.g., `bank=FAKEBANK`, `index=INVALID99`), **When** parameter validation is executed, **Then** the catalog flags the violation with a specific validation failure reason.
3. **Given** a list of securities and a target filter (such as an index code or sector group), **When** catalog filtering is applied, **Then** only matching securities adhering to the filter criteria are returned.

---

### Edge Cases

- **Mixed-case and whitespace inputs**: How does the system handle symbols like `" comi "` or `"qnbk"`? The catalog must automatically trim whitespace and normalize casing to uppercase canonical tickers.
- **Ambiguous symbols across regional exchanges**: What happens when two different exchanges share an identical ticker symbol? The catalog must allow optional market qualification (e.g., symbol + market) to disambiguate collisions while providing sensible domestic defaults when market is omitted.
- **Securities with multiple legacy aliases**: How are securities with ticker changes or dual codes (such as Tadawul numeric code `"1010"` vs mnemonic `"RIBL"`) resolved? The catalog must maintain an alias registry that resolves both identifiers to the identical canonical entity.
- **Empty or null parameter inputs**: The catalog must return structured validation errors rather than throwing runtime reference errors.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The catalog MUST maintain a comprehensive inventory encompassing all supported Egyptian banks, currencies, commodities (gold karats and fuel categories), domestic EGX equities, GCC regional equities (Saudi Arabia, UAE, Qatar, Kuwait), and benchmark indices.
- **FR-002**: The catalog MUST provide a universal security lookup capability that resolves an asset by ticker symbol, numeric code, or alias, returning its complete canonical profile.
- **FR-003**: The catalog MUST support case-insensitive, whitespace-trimmed symbol lookups and resolve known symbol aliases to their canonical primary representation.
- **FR-004**: The catalog MUST provide an authoritative index constituent lookup operation that returns all securities belonging to a specified benchmark index.
- **FR-005**: The catalog MUST provide metadata enrichment operations that augment raw quote data with official localized names (Arabic and English), sector classifications, branding logo URLs, and index tags.
- **FR-006**: The catalog MUST provide centralized validation operations for all domain parameters, including bank codes, currency pairs, exchange markets, stock groups, and benchmark index codes.
- **FR-007**: The catalog MUST support catalog filtering operations to extract subsets of securities based on exchange, sector group, or index affiliation without requiring callers to implement bespoke array iterations.
- **FR-008**: The catalog MUST encapsulate all underlying inventory representations as private implementation details, exposing only high-level query, resolution, enrichment, and validation operations.

### Key Entities *(include if feature involves data)*

- **Security Profile**: Represents a single tradable asset or equity, containing canonical symbol, numeric code (where applicable), exchange/market, localized English name, localized Arabic name, sector, currency, logo URL, and associated benchmark indices.
- **Benchmark Index**: Represents a market index (e.g., EGX30, TASI), containing index code, market exchange, localized titles, description, and the list of member security symbols.
- **Catalog Asset Category**: Represents supported asset classifications, including banks, currencies, commodities, domestic stocks, regional stocks, and indices.
- **Validation Outcome**: Represents the result of parameter validation, containing validity flag, normalized canonical value, and optional error code/message.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of asset, symbol, and constituent lookups resolve in under 1 millisecond from memory without external I/O.
- **SC-002**: Elimination of duplicate traversal and filtering loops across all query and route modules, concentrating all catalog access behind a single module interface.
- **SC-003**: 100% of supported securities across domestic and GCC markets resolve correctly regardless of input case or leading/trailing whitespace.
- **SC-004**: Parameter validation produces unified error messages across all HTTP endpoints, achieving 100% consistency in client-facing validation responses.
- **SC-005**: Existing contract and integration test suites achieve 100% pass rate without regressions in endpoint contracts.

## Assumptions

- The inventory of supported banks, currencies, commodities, equities, and indices remains relatively static during runtime and is safely held in memory as the primary authoritative reference.
- Symbol alias mappings (e.g., numeric codes vs mnemonic tickers) are deterministic and defined within the catalog data set.
- Existing REST endpoint URL schemas and client-facing response formats will remain completely backwards compatible.
- The catalog module runs in-process with zero external network or database dependencies.
