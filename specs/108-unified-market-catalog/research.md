# Architectural & Technical Research: Unified Market Catalog

**Branch**: `108-unified-market-catalog` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

---

## Technical Decisions

### Decision 1: Pre-Indexed Multi-Key In-Memory Catalog Architecture
- **Decision**: Deepen `CatalogService` into a unified `MarketCatalog` module initialized as an in-memory singleton that pre-indexes all platform entities (banks, currencies, commodities, domestic EGX stocks, GCC equities, and benchmark indices) into specialized O(1) hash maps at startup.
- **Rationale**:
  - Eliminates repetitive O(N) array scans (`.find()`, `.filter()`) and per-request `Map` allocations across `StockQueryService`, `IndicesQueryService`, and `GccStockQueryService`.
  - Lookups resolve in `<0.01ms` from volatile RAM, 100% satisfying Constitution Principle II (Zero-Read Memory Caching) and Success Criterion SC-001.
  - Memory footprint for all ~600 entities across Egypt and GCC is less than 2MB, easily accommodated on the OCI Ampere A1 24GB instance.
- **Alternatives Considered**:
  - *On-demand linear array scans (current approach)*: Rejected due to duplicate iteration logic in 5 different files and unnecessary GC pressure from transient `new Map()` instantiations.
  - *SQLite / Embedded key-value database*: Rejected as premature over-engineering; platform inventory is read-only and static at runtime.

---

### Decision 2: Normalization, Disambiguation & Alias Resolution Strategy
- **Decision**: Implement a multi-stage symbol resolution pipeline inside `MarketCatalog.getSecurity(identifier, market?)`:
  1. **Sanitization**: Trim leading/trailing whitespace, convert to uppercase, and strip legacy exchange suffixes (e.g., `COMI.CA` &rarr; `COMI`).
  2. **Market-Qualified Match**: If `market` is supplied, check the composite index `MARKET:SYMBOL` (e.g., `SA:1010` or `EG:COMI`).
  3. **Universal Symbol/Code Match**: Check primary ticker symbol index, numeric code index (e.g. `1010` for Al Rajhi Bank), and retail alias registry (e.g., `BDC` &rarr; `BQDC`).
  4. **Domestic Defaulting**: If market is omitted and a collision occurs across exchanges, resolve to the domestic EGX entity to preserve 100% backwards compatibility with legacy client queries.
- **Rationale**:
  - Solves client input inconsistencies (mixed case, retail abbreviations, whitespace) cleanly at the catalog seam.
  - Ensures seamless interoperability between alphanumeric symbols and Tadawul 4-digit numeric codes.
- **Alternatives Considered**:
  - *Strict case-sensitive exact matching*: Rejected because mobile clients frequently pass varying case in deep links and search queries.
  - *Fuzzy string matching (Levenshtein)*: Rejected to prevent accidental resolution of distinct ticker symbols (e.g., `ETEL` vs `ESRS`).

---

### Decision 3: Pre-Aggregated Benchmark Index Constituent Indexing
- **Decision**: Construct a reverse constituent index `Map<string, SecurityDefinition[]>` at startup that maps every uppercase benchmark index code (`EGX30`, `EGX70`, `EGX100`, `TASI`, `ADX15`, `DFMGI`, `QE20`, `BKP`) directly to its constituent security objects.
- **Rationale**:
  - `catalog.getConstituents(indexCode)` becomes an instantaneous O(1) pointer return.
  - Completely eliminates the 150-line constituent scanning loop in `indices-query-service.ts` (lines 42–195).
  - Seamlessly handles cross-border index queries with identical ergonomics.
- **Alternatives Considered**:
  - *Scanning stock arrays on every constituent request*: Rejected; wastes CPU cycles on every index constituent API hit.

---

### Decision 4: Standardized Snapshot Metadata Enrichment Pipeline
- **Decision**: Expose a unified enrichment operation `MarketCatalog.enrichSnapshot(assetClass, snapshot, market?)` that populates missing branding logos, localized Arabic/English names, sectors, and benchmark index associations using authoritative catalog metadata.
- **Rationale**:
  - Consolidates scattered `enrichSnapshotWithCatalog` implementations from `StockQueryService` (lines 60–91) and `GccStockQueryService` (lines 85–150).
  - Provides a single point of modification for corporate asset branding and logo updates.
- **Alternatives Considered**:
  - *Enriching during database ingestion/scraping only*: Rejected because Firestore documents may have been archived prior to metadata additions; live in-memory enrichment ensures clients always receive the latest branding.

---

### Decision 5: Centralized Domain Validation and Parameter Guards
- **Decision**: Expose `MarketCatalog.validate(type, value, market?): ValidationResult` with dedicated validator methods (`isValidBank`, `isValidCurrency`, `isValidIndex`, `isValidStockGroup`, `isValidSecurity`).
- **Rationale**:
  - Replaces ad-hoc `VALID_STOCK_GROUPS.includes()` and `getIndexEntry()` checks in `stock-routes.ts` and `indices-routes.ts`.
  - Provides uniform validation outcomes: `{ valid: boolean, canonicalValue?: string, error?: string }`.
  - Ensures Fastify routes reject invalid requests with standardized 400 Bad Request error envelopes before querying database or cache layers.
- **Alternatives Considered**:
  - *Relying solely on Zod route schemas*: Zod validates string formats and non-empty constraints, but runtime membership in dynamic market catalogs belongs at the catalog seam.
