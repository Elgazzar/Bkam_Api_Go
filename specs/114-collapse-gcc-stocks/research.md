# Phase 0 Research: Collapse Artificial GCC Stocks Boundary into Stock Module

**Feature**: `specs/114-collapse-gcc-stocks`

---

## 1. Architectural Decisions

### Decision 1: Relocation of Regional Stock Catalogs
- **Decision**: Relocate the static security definition files from `src/modules/gcc-stocks/catalogs/` (`saudi-catalog.ts`, `uae-catalog.ts`, `qatar-catalog.ts`, `kuwait-catalog.ts`) to `src/config/catalogs/`.
- **Rationale**: The static catalog definitions represent canonical domain configuration (symbols, ISINs, company names, sectors, trading hours). They are consumed during application boot by `MarketCatalog` (`src/services/catalog-service.ts`) to build the in-memory multi-key index. Moving them to `src/config/catalogs/` standardizes them alongside other domain catalogs without requiring an artificial `modules/gcc-stocks/` hierarchy.
- **Alternatives Considered**:
  - *Keep in `src/modules/gcc-stocks/catalogs/`*: Rejected because the entire goal is to delete `src/modules/gcc-stocks/`.
  - *Merge all GCC stocks into `src/config/stock-catalog.ts`*: Rejected because keeping per-exchange catalog files (`saudi-catalog.ts`, `uae-catalog.ts`, etc.) provides clean modular separation and easier maintenance when an exchange updates its constituent list.

---

### Decision 2: Query Service Interface Deepening
- **Decision**: Deepen `StockQueryService` (`src/services/stock-query-service.ts`) to natively handle all regional equity markets via an optional `market` parameter (defaulting to `'egx'`).
- **Signature Design**:
  ```typescript
  export class StockQueryService {
    // Synchronous RAM snapshot lookup across any market partition (<0.01ms)
    public getCachedPrices(market: string = "egx"): StockPriceSnapshot | null;

    // Push listener cache update across any market partition
    public updateCache(data: StockPriceSnapshot, market: string = "egx"): void;

    // Fetch latest prices with automated warming and thundering-herd deduplication
    public async fetchLatestStockPrices(market: string = "egx"): Promise<StockPriceSnapshot | null>;

    // Fetch single-date historical daily closing summary via HistoryResolutionEngine
    public async fetchStockPriceHistory(date: string, market: string = "egx"): Promise<StockPriceSnapshot | null>;

    // Fetch multi-week historical range with flat-cache prefixing via HistoryResolutionEngine
    public async fetchStockPriceHistoryRange(symbol: string, startDate: string, endDate: string, market: string = "egx"): Promise<any[]>;

    // Fetch latest trading session intraday timeline via IntradayCoordinator
    public async fetchLatestStockIntraday(symbol: string, market: string = "egx"): Promise<any>;
  }
  ```
- **Rationale**: `LiveSnapshotCoordinator`, `HistoryResolutionEngine`, and `IntradayCoordinator` already accept market identifiers. `StockQueryService` acts as the single authoritative deep module coordinating equity data across all 5 exchanges.
- **Elimination of Duplication**: `gcc-stock-query-service.ts` contained ~150 lines of manual catalog enrichment loops. In the unified architecture, `MarketCatalog.enrichSnapshot` handles enrichment centrally in O(1) time.

---

### Decision 3: Route Ergonomics and Backward Compatibility
- **Decision**: Maintain 100% backward compatibility for all existing REST routes in `src/routes/stock-routes.ts`.
- **Supported Endpoints**:
  1. `GET /v1/stocks`: Accepts `?market=saudi|uae|qatar|kuwait|egx` (defaults to `egx`). Also accepts ISO country codes `?market=sa|ae|qa|kw|eg`.
  2. Legacy Aliases:
     - `GET /v1/stocks/saudi`
     - `GET /v1/stocks/uae`
     - `GET /v1/stocks/qatar`
     - `GET /v1/stocks/kuwait`
     - `GET /v1/stocks/regional/catalog`
  3. Historical Endpoints:
     - `GET /v1/stocks/history/:date` (with optional `?market=...`)
     - `GET /v1/stocks/:market/history/:date` (legacy path alias)
     - `GET /v1/stocks/:symbol/history/range` (with optional `?market=...`)
     - `GET /v1/stocks/:market/:symbol/history/range` (legacy path alias)
  4. Intraday Endpoints:
     - `GET /v1/stocks/:symbol/intraday/latest` (with optional `?market=...`)
     - `GET /v1/stocks/:market/:symbol/intraday/latest` (legacy path alias)
- **Rationale**: Ensures zero breakage for mobile app clients and guarantees 100% pass on differential parity checks (`npm run verify:diff`).

---

### Decision 4: Purge Dead Scraper Daemons (Constitution Principle VIII)
- **Decision**: Completely delete `src/modules/gcc-stocks/adapters/` (`saudi-stock-adapter.ts`, `uae-stock-adapter.ts`, `qatar-stock-adapter.ts`), `gcc-firestore-service.ts`, and `standalone-test.ts`.
- **Rationale**: The API gateway serves data; external scrapers run in isolated background jobs or external pipelines. Leaving scraper classes in the API codebase violates Constitution Principle VIII ("The API service MUST remain entirely decoupled from external scraping daemons") and causes confusion.
- **Alternatives Considered**: Retaining them in a `scrapers/` folder was rejected because scraper code does not belong in this repository.

---

### Decision 5: Elimination of Dynamic Import Loader in `app.ts`
- **Decision**: Remove lines 172–220 of `src/app.ts` which dynamically loaded `src/modules/gcc-stocks/index.js` via `try/catch import()`.
- **Rationale**: Dynamic module loading was used as a workaround for circular and missing build dependencies. With `stock-routes.ts` handling all routes statically, route registration is deterministic and fully typed.
