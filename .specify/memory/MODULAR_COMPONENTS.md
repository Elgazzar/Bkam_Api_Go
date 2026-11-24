# Modular Components Registry

This document catalogs all architectural subsystems, abstractions, and features implemented as modular components within the Bekam Fe El Bank project, as mandated by **Principle XV** of the project constitution.

---

## Architecture Principles for Modularity
1. **Interface Segregation & Loose Coupling**: Components MUST expose contract interfaces (`I...`) or base abstractions so that consumers depend on behavior, not concrete implementations.
2. **Plug-and-Play Swappability**: Changing a provider (e.g., swapping local cache for Redis, or switching a market data provider) MUST only require adding/updating a modular class/file and registering it in a factory/singleton, leaving core business logic and HTTP routes completely untouched.
3. **Isolated Fault Domains**: Failures in one modular subsystem MUST be contained within its boundary and handled defensively without crashing parent orchestrators (Principle XIV).

---

## Active Modular Components

### 1. Cache Service & Storage Providers
- **Purpose**: Provides persistent and in-memory caching for API responses, historical archives, and rate snapshots with configurable TTL.
- **Contract Interface**: `ICacheProvider` in [`api/src/cache/ICacheProvider.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/cache/ICacheProvider.ts)
- **Active Implementations**:
  - `FlatCacheProvider` in [`api/src/cache/FlatCacheProvider.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/cache/FlatCacheProvider.ts) — File-backed persistent disk cache (`flat-cache`).
- **Coordinator / Singleton**:
  - `CacheService` in [`api/src/cache/cacheService.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/cache/cacheService.ts) — Delegates to the active `ICacheProvider`.
- **How to Replace**: Implement a new class adhering to `ICacheProvider` (e.g., `RedisCacheProvider.ts`) and pass it to `CacheService.getInstance(new RedisCacheProvider())`.

---

### 2. Bank Currency Scraper Adapters
- **Purpose**: Scrapes real-time exchange rates for individual Egyptian banks and normalizes them into a unified domain model.
- **Contract / Base Class**:
  - `BaseAdapter` in [`api/src/adapters/base-adapter.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/base-adapter.ts)
  - `BankAdapter` in [`api/src/models/bank-adapter.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/models/bank-adapter.ts)
- **Active Implementations** ([`api/src/adapters/banks/`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/banks)):
  - `NbeAdapter` ([`api/src/adapters/banks/nbe.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/banks/nbe.ts))
  - `CibAdapter` ([`api/src/adapters/banks/cib.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/banks/cib.ts))
  - `BdcAdapter` ([`api/src/adapters/banks/bdc.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/banks/bdc.ts))
  - `QnbAdapter` ([`api/src/adapters/banks/qnb.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/banks/qnb.ts))
  - `BmAdapter` ([`api/src/adapters/banks/bm.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/banks/bm.ts))
  - `AaibAdapter` ([`api/src/adapters/banks/aaib.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/banks/aaib.ts))
  - *(And 15+ additional bank adapters)*
- **Registry & Factory**:
  - `createSupportedAdapters()` in [`api/src/adapters/registry.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/registry.ts)
- **How to Replace**: Modify or add an adapter file under `api/src/adapters/banks/` and update its instantiation in `registry.ts`.

---

### 3. Commodity & Market Asset Adapters
- **Purpose**: Independent adapters for scraping non-bank commodities, fuel, precious metals, and stocks.
- **Active Implementations**:
  - **BTC Gold Adapter**: [`api/src/adapters/btc.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/btc.ts) (Local gold bullion prices)
  - **GoldEra Gold Adapter**: [`api/src/adapters/goldera.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/goldera.ts) (Fallback local gold prices)
  - **Petroleum Gas Adapter**: [`api/src/adapters/petroleum.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/petroleum.ts) (Egyptian fuel/gas rates)
  - **TradingView Adapter**: [`api/src/adapters/tradingview.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/tradingview.ts) (International spot gold & EGX stock indices)
  - **Mubasher Stocks Adapter**: [`api/src/adapters/mubasher.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/mubasher.ts) (EGX stock quotes)
  - **Ta3weem Aggregator Adapter**: [`api/src/adapters/ta3weem.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/adapters/ta3weem.ts) (Failover aggregator for bank rate extraction)
- **How to Replace**: Replace the target adapter file without affecting unrelated commodities or routes.

---

### 4. Metrics Tracking & Storage Layer
- **Purpose**: Records daily API traffic volume categorized by asset type with debounced background persistence.
- **Contract Interface**: `IMetricsStore` in [`api/src/services/metrics-store.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/metrics-store.ts)
- **Active Implementations**:
  - `FlatCacheMetricsStore` in [`api/src/services/metrics-store.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/metrics-store.ts) (Disk persistence using `CacheService`)
  - `MemoryMetricsStore` in [`api/src/services/metrics-store.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/metrics-store.ts) (In-memory storage for unit tests)
- **Coordinator / Service**:
  - `MetricsService` in [`api/src/services/metrics-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/metrics-service.ts)
- **How to Replace**: Implement `IMetricsStore` (e.g. `PostgresMetricsStore.ts` or `FirestoreMetricsStore.ts`) and inject it into `new MetricsService(new PostgresMetricsStore())`.

---

### 5. Crash Loop Protection (CrashGuard)
- **Purpose**: Detects rapid consecutive container/process restarts and enforces exponential backoff delays before database access to protect Firestore read quotas.
- **Implementation**: [`api/src/services/crash-loop-detector.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/crash-loop-detector.ts)
- **How to Replace**: Self-contained module with zero external database dependencies; can be swapped or customized directly.

---

### 6. Concurrency Control Worker Pool
- **Purpose**: Limits simultaneous scraping worker executions to prevent host RAM/CPU exhaustion and CDN bot triggers.
- **Implementation**: `WorkerPool` in [`api/src/lib/pool.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/lib/pool.ts)
- **How to Replace**: Independent concurrency queue primitive that can be replaced with alternative queue managers (e.g., BullMQ, p-limit) without altering scraper adapters.

---

### 7. Real-time Firestore Listeners & In-Memory Stores (Intraday & Active Weekly Bucket)
- **Purpose**: Listens for real-time Firestore database snapshots (`onSnapshot`) and maintains live intraday price shifts and the active mutating weekly bucket document in memory for 0-latency HTTP delivery with zero Firestore polling reads. Includes automated ISO week rollover sealing to `flat-cache` (30-day TTL) and cold-boot pre-seeding.
- **Active Implementations**:
  - `FirestoreListenerService` in [`api/src/services/firestore-listener-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/firestore-listener-service.ts) — Event-driven subscription manager.
  - `IntradaySnapshotService` in [`api/src/services/intraday-snapshot-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-snapshot-service.ts) — In-memory intraday aggregator with dual-buffer (`currentSnapshot` and `previousSnapshot`) retention.
  - `IntradayCoordinator` in [`api/src/services/intraday-coordinator.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-coordinator.ts) — Centralized session backtracking, multi-symbol point lookup, and previous day baseline resolution (Feature 100) across all asset classes with zero database read charges.
  - `ActiveWeeklyBucketStore` in [`api/src/services/active-weekly-bucket-store.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/active-weekly-bucket-store.ts) — In-memory volatile store for the active mutating weekly consolidated document.
- **How to Replace**: Decoupled from HTTP routes; can be switched to alternative push protocols (e.g. Redis Pub/Sub, WebSockets) without affecting client-facing route controllers.

---

### 8. Stealth HTTP Request Client
- **Purpose**: Centralized request helper spoofing browser headers, TLS configurations, and User-Agents to prevent anti-bot blocks, with support for caller-specified header suppression (empty string / undefined via `buildRequestHeaders`).
- **Implementation**: `createHttpClient` in [`api/src/lib/http-client.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/lib/http-client.ts) and [`serverless-scraper/src/lib/http-client.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/lib/http-client.ts)
- **How to Replace**: Can be replaced with alternative HTTP clients (e.g. Undici, Got, Axios with proxy rotation) without altering adapter parsing logic.

---

### 9. Intraday Coordinator & Timeline Orchestration Layer
- **Purpose**: Centralizes and coordinates all intraday price movement timeline resolution, single-read cold-cache fallback, midnight session rollover, historical archive point recovery, and heterogeneous point normalization across all asset domains (`rates`, `gold`, `stocks`, `indices`). Any modifications to intraday timeline orchestration occur in **exactly one location** per Constitution Principle XV.
- **Contract Interface**: `IIntradayCoordinator` in [`api/src/services/intraday-types.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-types.ts)
- **Active Implementation / Singleton**:
  - `IntradayCoordinator` in [`api/src/services/intraday-coordinator.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-coordinator.ts)
- **Consuming Services**:
  - `RatesQueryService` ([`api/src/services/rates-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/rates-query-service.ts))
  - `GoldQueryService` ([`api/src/services/gold-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/gold-query-service.ts))
  - `StockQueryService` ([`api/src/services/stock-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/stock-query-service.ts))
  - `IndicesQueryService` ([`api/src/services/indices-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/indices-query-service.ts))
- **How to Replace**: Implement a class adhering to `IIntradayCoordinator` and inject or re-assign `intradayCoordinator`. All query services invoke only methods defined on `IIntradayCoordinator` (`resolveTimeline`, `recoverYesterdayPoints`, `lookupSymbolPoints`, `normalizePoints`).

---

### 10. Bank Failure Tracker & Scraper Configuration Layer
- **Purpose**: Consolidates consecutive failed scraping cycle tracking, transient failure tolerance, automated aggregator escalation thresholds, direct website self-healing, declarative fallback-only routing (`fallbackOnlyBanks`: `["hsbc", "enbd", "aaib", "adib", "faisal", "abk", "saib"]`), and runtime configuration parameters in a single, authoritative module per Constitution Principle XV. All configuration parameters (`SCRAPER_CONFIG_PARAMETERS`, `consecutiveFailuresBeforeFallback`, `fallbackOnlyBanks`) are maintained in **exactly one location** within `IBankFailureTracker` so users and operators can easily review and adjust settings.
- **Contract Interface**: `IBankFailureTracker` in [`serverless-scraper/src/services/bank-failure-tracker.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/bank-failure-tracker.ts)
- **Active Implementation / Singleton**:
  - `BankFailureTracker` / `bankFailureTracker` in [`serverless-scraper/src/services/bank-failure-tracker.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/bank-failure-tracker.ts)
- **Configuration Bridge**:
  - `scraper-config.ts` in [`serverless-scraper/src/config/scraper-config.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/config/scraper-config.ts) (re-exports centralized schemas, types, and loader from `bank-failure-tracker.ts`)
  - `scraper.config.json` in [`serverless-scraper/scraper.config.json`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/scraper.config.json) (on-disk override)
- **Consuming Services**:
  - Scraper Orchestrator ([`serverless-scraper/src/worker.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/worker.ts))
  - Firestore Service ([`serverless-scraper/src/services/firestore-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/firestore-service.ts))
- **How to Replace**: Implement a class adhering to `IBankFailureTracker` and re-assign `bankFailureTracker`. All scraper lifecycle operations invoke only contract methods (`recordSuccess`, `recordFailure`, `evaluate`, `getConfig`, `getThreshold`, `setConfig`, `isFallbackOnly`, `getFallbackOnlyBanks`).

---

### 11. Regional GCC Stock Scrapers & API Module (`gcc-stocks`)
- **Purpose**: Collects, normalizes, persists, and exposes real-time and historical equity quotes for Saudi Arabia (Tadawul), UAE (DFM & ADX), and Qatar (QSE) with zero-touch file-level portability.
- **Zero-Touch Portability Mechanism**: All feature code resides entirely within `src/modules/gcc-stocks/`. Core services (`api/src/app.ts` and `serverless-scraper/src/worker.ts`) load this feature dynamically using path variable indirection (`const gccStocksModulePath = ...; await import(gccStocksModulePath)`). Removing the `gcc-stocks/` directory disables the feature cleanly with zero compile-time breakages and zero runtime exceptions.
- **Module Structure**:
  - **Constituent Catalogs**: [`api/src/modules/gcc-stocks/catalogs/`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/catalogs/) (`saudi-catalog.ts`, `uae-catalog.ts`, `qatar-catalog.ts`)
  - **Market Scraper Adapters**: [`api/src/modules/gcc-stocks/adapters/`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/adapters/) (`saudi-stock-adapter.ts`, `uae-stock-adapter.ts`, `qatar-stock-adapter.ts` polling TradingView scanners)
  - **Trading Schedule Engine**: [`api/src/modules/gcc-stocks/lib/market-schedule.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/lib/market-schedule.ts) (Operating hours and weekend calendars for AST and GST timezones)
  - **Isolated Persistence Service**: [`api/src/modules/gcc-stocks/services/gcc-firestore-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/services/gcc-firestore-service.ts) (In-memory change detection, zero-write cost guard, partitioned weekly history)
  - **Query & Disk Cache Service**: [`api/src/modules/gcc-stocks/services/gcc-stock-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/services/gcc-stock-query-service.ts) (In-memory snapshots and disk cache)
  - **Fastify Route Registrations**: [`api/src/modules/gcc-stocks/routes/gcc-stock-routes.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/routes/gcc-stock-routes.ts) (`/v1/stocks/saudi`, `/uae`, `/qatar`, `/regional/catalog`, `/history`, and `/v1/stocks?country=...` alias)
- **How to Disable**: Delete or rename `src/modules/gcc-stocks/` from `api` and `serverless-scraper`. Core Egyptian stock endpoints continue operating unchanged.

---

### 12. Push Notification Subsystem: Two-Stage Transport & Strategy Pipeline
- **Purpose**: Provides a decoupled, provider-agnostic push notification infrastructure for dispatching volatility alerts across currencies, gold commodities, and equities, separating domain topic formatting from pure transport delivery per ADR-0002.
- **Contract Interfaces & Models**:
  - `INotificationProvider`, `NotificationDispatch`, and `PriceAlertEvent` in [`api/src/notifications/INotificationProvider.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/notifications/INotificationProvider.ts)
  - `IAlertTopicStrategy` in [`api/src/notifications/IAlertTopicStrategy.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/notifications/IAlertTopicStrategy.ts)
- **Active Strategy Implementations**:
  - `CuriosityGapTopicStrategy` in [`api/src/notifications/CuriosityGapTopicStrategy.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/notifications/CuriosityGapTopicStrategy.ts) — Encapsulates curiosity-gap topic construction and integer boundary calculation (`getCrossedBoundaries`), resolving domain events into transport-ready `NotificationDispatch` payloads.
- **Active Transport Provider Implementations**:
  - `FcmNotificationProvider` in [`api/src/notifications/FcmNotificationProvider.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/notifications/FcmNotificationProvider.ts) — Pure Firebase Cloud Messaging transport client delivering pre-formatted dispatches, with optional `admin.AppOptions` initialization and ambient credential fallback.
  - `MockNotificationProvider` in [`api/src/notifications/MockNotificationProvider.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/notifications/MockNotificationProvider.ts) — In-memory capture of `NotificationDispatch[]` payloads and error simulation for offline unit/integration testing with zero external cloud dependencies.
- **Coordinator / Singleton Orchestrator**:
  - `NotificationService` in [`api/src/notifications/notificationService.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/notifications/notificationService.ts) — Central orchestrator managing the active strategy (`setTopicStrategy`) and active provider (`setProvider`) with defensive error containment (`sendAlert`).
- **Backward Compatibility Facade**:
  - `fcmAlerts.ts` in [`api/src/services/fcmAlerts.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/fcmAlerts.ts) and [`serverless-scraper/src/services/fcmAlerts.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/fcmAlerts.ts) — Preserves 100% legacy signature compatibility (`sendPriceChangeAlert`, `getCrossedBoundaries`), delegating under the hood to `NotificationService`.
- **How to Replace**:
  - **Replace Provider (Transport)**: Implement a class adhering to `INotificationProvider` (e.g. `WebhookNotificationProvider`, `OneSignalNotificationProvider`) accepting `NotificationDispatch` and activate it via `NotificationService.getInstance().setProvider(new CustomProvider())`.
  - **Replace Strategy (Formatting)**: Implement a class adhering to `IAlertTopicStrategy` (e.g. `StandardAlertTopicStrategy` with direct prices in titles) and activate it via `NotificationService.getInstance().setTopicStrategy(new CustomStrategy())`.

---

### 13. Unified Parameterized Market Scraper Engine & Registry (`TradingViewMarketAdapter`)
- **Purpose**: Consolidates all regional equity and index market scrapers (Egypt EGX, EGX indices, Saudi Arabia Tadawul, UAE DFM/ADX, Qatar QSE) into a single, reusable parameterized scraper engine driven by a centralized declarative Market Registry, inline calendar schedule gating, and isolated fault boundaries per Constitution Principle XV and ADR 0001.
- **Contract Interface & Types**:
  - `IMarketAdapter`, `MarketDefinition`, `TradingSchedule`, `StockCatalogItem` in [`serverless-scraper/src/adapters/base/market-adapter.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/adapters/base/market-adapter.ts)
- **Active Implementations**:
  - `TradingViewMarketAdapter` in [`serverless-scraper/src/adapters/markets/tradingview-market-adapter.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/adapters/markets/tradingview-market-adapter.ts) — Reusable scanner engine querying `https://scanner.tradingview.com/${scannerSlug}/scan`.
  - `MARKET_REGISTRY` in [`serverless-scraper/src/config/market-registry.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/config/market-registry.ts) — Central declarative market definitions and catalogs.
  - `isScheduleActive` in [`serverless-scraper/src/modules/gcc-stocks/lib/market-schedule.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/modules/gcc-stocks/lib/market-schedule.ts) — Timezone-aware calendar gating engine.
  - `MarketOrchestrator` / `marketOrchestrator` in [`serverless-scraper/src/services/market-orchestrator.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/market-orchestrator.ts) — Isolated execution runner with 20s timeouts and error containment.
  - `saveMarketPrices` in [`serverless-scraper/src/services/market-persistence-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/market-persistence-service.ts) — Zero-write cost-boundary Firestore persistence.
- **Backward Compatibility Facades**:
  - `TradingViewEgxAdapter`, `TradingViewIndicesAdapter`, `TradingViewSaudiAdapter`, `TradingViewUaeAdapter`, `TradingViewQatarAdapter` derive as lightweight subclasses of `TradingViewMarketAdapter` to preserve 100% legacy import compatibility.
- **How to Onboard New Markets**: Add a single declarative `MarketDefinition` entry to `MARKET_REGISTRY` in `market-registry.ts` with its trading schedule, currency, and stock catalog. No new scraper classes or deployment changes are required.

---

### 14. Declarative Benchmark Index Registry & Regional Index Ingestion (`IndexRegistry` & `IndicesQueryService`)
- **Purpose**: Establishes a single authoritative declarative catalog of benchmark financial indices across Egypt and the GCC (EGX, Tadawul, DFM/ADX, QSE), enforces mandatory index membership across 100% of equities, and ingests/serves real-time index benchmarks and constituents with segregated storage and multi-market caching.
- **Contract Interface & Models**:
  - `IndexDefinition` and `IndexEntry` in [`api/src/config/index-catalog.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/config/index-catalog.ts) and [`serverless-scraper/src/config/index-catalog.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/config/index-catalog.ts)
  - Canonical Index Codes: `EGX30`, `EGX70`, `EGX100`, `EGX33`, `TASI`, `MT30`, `DFMGI`, `FADGI`, `FADX15`, `QSE_INDEX`.
  - Helper functions: `isValidIndexCode(code)`, `getIndexEntry(code)`, `getIndexCatalog(market?)`.
- **API Routes & Filtering**:
  - Constituent Reverse-Lookup: `GET /v1/indices/:code/stocks` via [`api/src/routes/indices-routes.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/routes/indices-routes.ts) resolves equities matching an index code with in-memory caching and catalog fallback.
  - Multi-Market Index Listing: `GET /v1/indices?market=egx|saudi|uae|qatar` (defaults to `egx` for 100% backward compatibility) with regional alias routes (`/v1/indices/saudi`, `/v1/indices/uae`, `/v1/indices/qatar`).
  - Stock Filtering: `?index=` filtering across `GET /v1/stocks`, `GET /v1/stocks/saudi`, `GET /v1/stocks/uae`, `GET /v1/stocks/qatar`, `GET /v1/stocks/:market/history/:date` with precedence over legacy `?group=`.
- **Scraper Ingestion & Market Registry Integration**:
  - Declarative Index Markets: `SA_INDICES`, `AE_INDICES`, and `QA_INDICES` registered in `MARKET_REGISTRY` in [`serverless-scraper/src/config/market-registry.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/config/market-registry.ts).
  - Segregated Storage: Persists to Firestore collections `index_prices`, `index_prices_saudi`, `index_prices_uae`, `index_prices_qatar` via `saveMarketPrices` in [`serverless-scraper/src/services/market-persistence-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/market-persistence-service.ts) protected by change-hash zero-write guards.
- **Catalog Integrity Enforcement**:
  - 100% of equities in all regional catalogs specify `indices: string[]` (`indices.length >= 1`), mirrored in `groups` for legacy backward compatibility. Automated CI test suites (`catalog-validation.test.ts`, `catalog-index-integrity.test.ts`) fail builds upon any unmapped or missing index code.
- **How to Add a Benchmark Index**:
  - Add the index entry to `INDEX_CATALOG` in `index-catalog.ts` in both `api/` and `serverless-scraper/`.
  - Add the index's TradingView symbol to the respective market definition in `market-registry.ts` if live scraping is required.

---

### 15. Regional Stock Catalogs & Alphanumeric Symbol Standardization
- **Purpose**: Governs canonical constituent catalogs, symbol naming standardization, and deduplicated dictionary serialization across regional stock markets (Saudi Arabia, UAE, Qatar) per Constitution Principle XV and ADR 0004.
- **Contract Interface & Models**:
  - `GccStockCatalogEntry`, `StockQuote`, `GccStockPriceItem` in [`api/src/modules/gcc-stocks/models/gcc-stock-types.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/models/gcc-stock-types.ts) and [`serverless-scraper/src/modules/gcc-stocks/models/gcc-stock-types.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/modules/gcc-stocks/models/gcc-stock-types.ts)
  - `numericCode?: string` preserves official numeric exchange ticker (e.g. `"2222"` for Aramco, `"1120"` for Al Rajhi).
- **Active Implementations**:
  - **Saudi Stock Catalog (30 Equities)**: [`api/src/modules/gcc-stocks/catalogs/saudi-catalog.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/catalogs/saudi-catalog.ts) — 100% complete constituent coverage for `MT30` and `TASI`. Symbols are uppercase alphanumeric codes (e.g. `ARAMCO`, `ALRAJHI`, `SNB`).
  - **UAE Stock Catalog (28 Equities)**: [`api/src/modules/gcc-stocks/catalogs/uae-catalog.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/catalogs/uae-catalog.ts) — 15 `FADX15` constituents (ADX) and 13 liquid constituents (DFM).
  - **Qatar Stock Catalog (20 Equities)**: [`api/src/modules/gcc-stocks/catalogs/qatar-catalog.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/catalogs/qatar-catalog.ts) — 100% complete constituent coverage for `QSE_INDEX` (QE 20).
- **Deduplication Engine**:
  - `TradingViewMarketAdapter` strictly indexes parsed quotes by canonical `quoteItem.symbol` into `stocksMap`, eliminating duplicate dictionary insertions.
  - `GccStockQueryService.enrichSnapshotWithCatalog` dynamically deduplicates cached or stored dictionary entries by mapping quotes to catalog `symbol`.
- **How to Add a Constituent Equity**:
  - Add the equity with full metadata (`symbol`, `code`, optional `numericCode`, `tvSymbol`, `indices`, `groups`, `logo`) to the respective catalog in both `api/` and `serverless-scraper/`.

---

### 16. Batched Weekly Historical Query Engine & Tiered Cache Layer
- **Purpose**: Batches multi-month historical range queries across consolidated ISO weekly bucket documents (`{collection}_weekly/{YYYY-Www}`), reducing Firestore read operations by >85%, eliminates redundant legacy fallback queries on closed market days (weekends & public holidays), and provides tiered memory caching (24-hour TTL for past immutable weeks, 5-second TTL for active week) along with negative caching for missing past dates.
- **Contract Interface & Helpers**:
  - `getISOWeekRange(startDate, endDate)` and `getISOWeekString(date)` in [`api/src/lib/date-utils.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/lib/date-utils.ts)
  - `fetchWeeklyBucketDocument(collectionName, weekKey)` in [`api/src/services/firestore-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/firestore-service.ts)
  - `fetchWeeklyFallbackHistory(collectionName, date)` in [`api/src/services/firestore-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/firestore-service.ts)
- **Active Implementations**:
  - `StockQueryService.fetchStockPriceHistoryRange` in [`api/src/services/stock-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/stock-query-service.ts)
  - `StockQueryService.fetchStockPriceHistory` in [`api/src/services/stock-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/stock-query-service.ts) (with negative caching)
  - `GccStockQueryService.fetchStockPriceHistory` in [`api/src/modules/gcc-stocks/services/gcc-stock-query-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/modules/gcc-stocks/services/gcc-stock-query-service.ts) (with negative caching)
  - `IntradayCoordinator.resolveTimeline` in [`api/src/services/intraday-coordinator.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-coordinator.ts) (closed-market fast-path)
- **How to Use**:
  - To query time series across date ranges, invoke `getISOWeekRange` to identify target weekly buckets, fetch via `fetchWeeklyBucketDocument`, and extract daily snapshots directly from the dictionary, bypassing day-by-day sequential queries.

---

### 17. Unified History Resolution Engine (`HistoryResolutionEngine` - ADR-0006, Feature 106)
- **Purpose**: Provides a single authoritative coordinator across all asset classes (Bank Rates, Gold, EGX Stocks, EGX Indices, GCC Equities) mediating single-date snapshots and multi-day range queries via tiered storage (Active Week RAM -> Negative Cache -> Persistent Disk Cache via `CacheService` -> Firestore Weekly Bucket -> Legacy Daily Fallback). Deepened under Feature 106 behind a single clean seam (`getDateSnapshot`, `getDateRange`) to absorb Africa/Cairo timezone calculations, active today live RAM + intraday coordinator synthesis, yesterday recovery fallback, multi-day live session appending, centralized future date validation rejection (400 Bad Request), and standardized caching metadata propagation.
- **Contract Interface & Types**:
  - `IHistoryResolutionEngine`, `AssetClass`, `GccMarket`, `RangeOptions`, `DateSnapshotOptions`, `ResolvedDateSnapshot` in [`api/src/services/history-resolution-engine.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/history-resolution-engine.ts)
- **Active Implementations**:
  - `HistoryResolutionEngine` in [`api/src/services/history-resolution-engine.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/history-resolution-engine.ts)
- **Coordinator / Service Integration**:
  - `historyResolutionEngine` singleton injected into `RatesQueryService`, `StockQueryService`, `IndicesQueryService`, `GoldQueryService`, and `GccStockQueryService`. All 4 caller query services delegate directly to `historyResolutionEngine.getDateSnapshot` and `historyResolutionEngine.getDateRange`, eliminating duplicated Cairo timezone math, branch conditions, and live RAM stitching.
  - Snapshot responses are decorated with non-enumerable caching metadata (`_isToday`, `_cacheTtlSeconds`) allowing Fastify route handlers to set HTTP `Cache-Control` (`max-age=60` for active today vs `max-age=3600` for past dates) without altering the serialized JSON envelope.
- **How to Replace / Extend**:
  - To support new asset classes or geographic markets, add collection mapping entries to `WEEKLY_COLLECTION_MAP` and `LEGACY_COLLECTION_MAP` within `HistoryResolutionEngine`, register weekly listeners in `FirestoreListenerService` if real-time active week RAM synchronization is required, and configure active live provider setters (`setRatesSnapshotProvider`, `setGoldPricesProvider`, etc.) in `app.ts`.

---

### 18. Latest Intraday Timeline Coordinator (`IntradayCoordinator` - ADR-0007, Feature 107)
- **Purpose**: Authoritative resolution coordinator for the "last trading day" intraday timelines across Stocks, Indices, Gold, and Bank Rates. Deepened under Feature 107 behind a single high-leverage interface method (`getLatestIntraday`) to absorb live quote extraction, session backtracking, previous closing price baselining (Feature 100), ticker alias normalization, consolidated all-karat payloads (`item=all`), and standardized cache lifetime metadata (`max-age=60` for active today vs `max-age=3600` for backtracked past sessions). Eliminates over 900 lines of duplicate boilerplate across `RatesQueryService`, `GoldQueryService`, `StockQueryService`, and `IndicesQueryService`. Bypasses database queries during active market hours by serving live tick series directly from RAM (`IntradaySnapshotService`, <1ms latency, 0 Firestore reads), and provides single-read descending Firestore backtracking with a 7-day lookback ceiling for off-hours, weekends, and holidays.
- **Contract Interface & Types**:
  - `IIntradayCoordinator`, `AssetClass`, `GccMarket`, `LatestIntradayOptions`, `LatestIntradayResult`, `ResolveLatestTimelineOptions`, `LatestIntradayTimelineResult` in [`api/src/services/intraday-types.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-types.ts)
- **Active Implementations**:
  - `IntradayCoordinator` in [`api/src/services/intraday-coordinator.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-coordinator.ts)
  - `IntradaySnapshotService` in [`api/src/services/intraday-snapshot-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/intraday-snapshot-service.ts)
- **Service Integration**:
  - `intradayCoordinator` singleton consumed by `StockQueryService`, `IndicesQueryService`, `GoldQueryService`, and `RatesQueryService`.
- **How to Use / Extend**:
  - Call `intradayCoordinator.getLatestIntraday(assetClass, target, options)` to resolve the complete finalized intraday envelope with quotes, points, and baselines. Domain query services act as slim caller facades delegating 100% of quote synthesis and backtracking to this seam.

---

### 19. Stateless Batch Quotes Service (`QuotesService` - Feature 098)
- **Purpose**: Unified BFF (Backend-For-Frontend) service resolving cross-domain favorites (Bank Currencies, Gold Carats, Benchmark Indices, Egyptian & GCC Stocks) in a single stateless request. Resolves exclusively from in-memory push-synchronized caches, maintaining zero Firestore read overhead.
- **Contract Interface & Types**:
  - `IQuotesService`, `QuoteItem`, `BatchQuoteResponse`, `QuoteCategory` in [`api/src/models/quote.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/models/quote.ts)
- **Active Implementations**:
  - `QuotesService` in [`api/src/services/quotes-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/services/quotes-service.ts)
- **Controller & Routes**:
  - `registerQuotesRoutes` in [`api/src/routes/quotes-routes.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/src/routes/quotes-routes.ts) (`POST /v1/quotes` & `GET /v1/quotes?items=...`)
- **Service Integration**:
  - Injected with `RatesSnapshotService`, `GoldQueryService`, `StockQueryService`, `IndicesQueryService`, and `GccStockQueryService` in `api/src/app.ts`.
- **How to Use / Extend**:
  - Call `quotesService.resolveQuotes(items)` with any array of identifiers (e.g. `["CBE_USD", "21k", "EGX30", "COMI", "1120"]`). To support new domain categories, add identifier resolution helpers in `QuotesService` following the non-throwing fallback pattern.

---

### 20. Modular Fallback Currency Providers (`IFallbackCurrencyProvider` & `BkamApiProvider`)
- **Purpose**: Provides modular secondary rate recovery for banks that fail direct scraping or run in fallback-only mode (e.g., HSBC), strictly decoupled from core scraping orchestrators. Selected dynamically via runtime Firestore configuration (`configurations/bkamapi` document with field `enabled: 1 | 0`).
- **Contract Interface & Models**:
  - `IFallbackCurrencyProvider` and `FallbackRecoveryResult` in [`serverless-scraper/src/adapters/base/fallback-provider.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/adapters/base/fallback-provider.ts).
- **Active Implementations**:
  - `BkamApiProvider` in [`serverless-scraper/src/adapters/providers/bkamapi/bkamapi-provider.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/adapters/providers/bkamapi/bkamapi-provider.ts) — Ingests exchange rates from upstream Firebase Realtime Database endpoints (`currencies-prices/{currencyId}.json`), maps 37 banks and 18 currencies via centralized catalogs (`bkamapi-mapping.ts`), sanitizes numeric strings, and filters out banks marked `hidden: true` or `stop: true`.
  - `Ta3weemAdapter` in [`serverless-scraper/src/adapters/banks/ta3weem-adapter.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/adapters/banks/ta3weem-adapter.ts) — Web scraper fallback aggregator used when Bkamapi is disabled (`enabled: 0`) or as secondary failover if Bkamapi encounters operational exceptions.
- **Dynamic Configuration & Toggle**:
  - `getBkamApiConfig` in [`serverless-scraper/src/adapters/providers/bkamapi/bkamapi-config.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/adapters/providers/bkamapi/bkamapi-config.ts) — Zero-read on-demand loader checking `configurations/bkamapi`. Evaluated strictly when `escalationQueue.length > 0`.
- **How to Replace / Swappability**:
  - Implement a new provider adhering to `IFallbackCurrencyProvider` (e.g. `CentralBankApiProvider.ts`).
  - Update the fallback resolver in `worker.ts` to instantiate the new provider.
  - To decommission Bkamapi, simply delete `src/adapters/providers/bkamapi/` and set `enabled: 0`; direct scrapers and core orchestrator logic remain 100% unaffected.

---

### 21. Scraper Relational Database Persistence Service (`SupabaseDatabaseService` - Feature 102)
- **Purpose**: Provides a decoupled, secondary relational storage engine for the background scraper service (`serverless-scraper`), persisting all scraped financial assets into strongly-typed PostgreSQL tables in Supabase immediately following primary Firestore batch commits. Replaces the legacy Redis dual-write cache with durable, structured SQL tables.
- **Contract Interface & Models**:
  - `ISupabaseDatabaseService` and relational row interfaces in [`serverless-scraper/src/services/supabase-types.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/supabase-types.ts).
- **Active Implementation**:
  - `SupabaseDatabaseService` in [`serverless-scraper/src/services/supabase-service.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/src/services/supabase-service.ts) — Connects via the official `@supabase/supabase-js` client over stateless HTTPS (PostgREST), mapping market domain objects into `bank_rates_latest`, `bank_rates_history`, `gold_prices_latest`, `gold_prices_history`, `gas_prices_latest`, `stock_quotes_latest`, `stock_quotes_history`, `index_quotes_latest`, and `index_quotes_history`.
- **Fault Boundary & Resilience (Principle XIV)**:
  - All public methods are wrapped in defensive `try/catch` error guards. If Supabase is unreachable, times out, or fails due to network latency, structured warnings are logged without throwing errors, and primary Firestore persistence continues completely unaffected.
- **How to Replace / Swappability**:
  - Implement a new class adhering to `ISupabaseDatabaseService` (e.g. `PostgresDirectService.ts`, `BigQueryDatabaseService.ts`).
  - Wire the replacement service into `firestore-service.ts`, `gcc-firestore-service.ts`, and `worker.ts`.
  - Core bank adapters, market scrapers, and Firestore persistence routines remain completely untouched.---

### 23. Centralized Live Snapshot In-Memory Coordinator (`LiveSnapshotCoordinator` - Feature 109)
- **Purpose**: Deep in-memory storage, warming, and push ingestion module consolidating real-time quotes across all 6 asset domains (`rates`, `gold`, `gas`, `stocks`, `indices`, `gcc-stocks`). Replaces fragmented ad-hoc in-memory slots, redundant Firestore fallback queries, and multi-service listener injection with an expressive, three-method contract interface (`getSnapshot`, `updateSnapshot`, `fetchFresh`).
- **Contract Interface & Domain Types**:
  - `ILiveSnapshotCoordinator`, `LiveSnapshotEnvelope`, `AssetDomain`, `FreshnessStatus`, and `DomainStorageKey` in [`src/models/live-snapshot-types.ts`](file:///d:/Repo/Bkam/Bkam_Api/src/models/live-snapshot-types.ts).
- **Active Implementation**:
  - `LiveSnapshotCoordinator` in [`src/services/live-snapshot-coordinator.ts`](file:///d:/Repo/Bkam/Bkam_Api/src/services/live-snapshot-coordinator.ts) — Encapsulates native Map memory partitions, `DiskSnapshotAdapter` file persistence (`var/snapshots/`), single-flight promise pooling against thundering herds, and graceful degraded stale serving.
- **Core Operations**:
  - `getSnapshot(domain, market?)`: Synchronous non-blocking in-memory retrieval of active quotes in $<0.01\text{ ms}$ with zero database I/O.
  - `updateSnapshot(domain, data, market?)`: Centralized push ingestion sink for real-time Firestore listeners with atomic pointer assignment and asynchronous disk persistence.
  - `fetchFresh(domain, market?, forceRefresh?)`: Authoritative upstream fetch with single-flight promise deduplication and degraded fallback recovery.
  - `warm()`: Rapid pre-seeding of in-memory cache from persistent disk storage on server boot ($<500\text{ ms}$).
- **How to Replace / Swappability**:
  - Downstream callers (`QuotesService`, `FirestoreListenerService`, query services, routes) depend strictly on `ILiveSnapshotCoordinator`.
  - Swapping local Map RAM storage for Redis or a clustered cache requires zero changes to route handlers or listener subscribers.
---

### 23. Centralized Live Snapshot In-Memory Coordinator (`LiveSnapshotCoordinator` - Feature 109)
- **Purpose**: Deep in-memory storage, warming, and push ingestion module consolidating real-time quotes across all 6 asset domains (`rates`, `gold`, `gas`, `stocks`, `indices`, `gcc-stocks`). Replaces fragmented ad-hoc in-memory slots, redundant Firestore fallback queries, and multi-service listener injection with an expressive, three-method contract interface (`getSnapshot`, `updateSnapshot`, `fetchFresh`).
- **Contract Interface & Domain Types**:
  - `ILiveSnapshotCoordinator`, `LiveSnapshotEnvelope`, `AssetDomain`, `FreshnessStatus`, and `DomainStorageKey` in [`src/models/live-snapshot-types.ts`](file:///d:/Repo/Bkam/Bkam_Api/src/models/live-snapshot-types.ts).
- **Active Implementation**:
  - `LiveSnapshotCoordinator` in [`src/services/live-snapshot-coordinator.ts`](file:///d:/Repo/Bkam/Bkam_Api/src/services/live-snapshot-coordinator.ts) — Encapsulates native Map memory partitions, `DiskSnapshotAdapter` file persistence (`var/snapshots/`), single-flight promise pooling against thundering herds, and graceful degraded stale serving.
- **Core Operations**:
  - `getSnapshot(domain, market?)`: Synchronous non-blocking in-memory retrieval of active quotes in $<0.01\text{ ms}$ with zero database I/O.
  - `updateSnapshot(domain, data, market?)`: Centralized push ingestion sink for real-time Firestore listeners with atomic pointer assignment and asynchronous disk persistence.
  - `fetchFresh(domain, market?, forceRefresh?)`: Authoritative upstream fetch with single-flight promise deduplication and degraded fallback recovery.
  - `warm()`: Rapid pre-seeding of in-memory cache from persistent disk storage on server boot ($<500\text{ ms}$).
- **How to Replace / Swappability**:
  - Downstream callers (`QuotesService`, `FirestoreListenerService`, query services, routes) depend strictly on `ILiveSnapshotCoordinator`.
  - Swapping local Map RAM storage for Redis or a clustered cache requires zero changes to route handlers or listener subscribers.
