# Research: Expose Previous Day Price in Latest Intraday Endpoints

**Feature Branch**: `100-intraday-previous-price`  
**Date**: 2026-11-05  
**Spec**: [Feature Specification](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/100-intraday-previous-price/spec.md)  
**Status**: Completed  

---

## 1. Executive Summary & Problem Formulation

Mobile client applications (specifically Android native apps utilizing MPAndroidChart, Vico, and Jetpack Compose Canvas) need a fixed baseline reference line to render intraday price charts effectively. In financial charting, this baseline represents the previous trading session's closing price:
- **Baseline Drawing**: Horizontal dotted line across the chart viewport indicating whether the security is currently trading at a net gain or net loss.
- **Dynamic Shading / Coloration**: Intraday curve and fill gradients dynamically shift to green (above baseline) or red (below baseline).
- **Single-Request Mobile Efficiency**: Previously, mobile clients had to make two independent HTTP requests—one to `/v1/*/intraday/latest` and a second historical query to `/v1/*/history/:yesterday`—doubling network latency, mobile battery drain, and exposing the system to Firestore read quota exhaustion if clients performed ad-hoc history queries.

By exposing top-level previous session price fields directly within all four latest intraday endpoints (`/v1/stocks/:symbol/intraday/latest`, `/v1/indices/:symbol/intraday/latest`, `/v1/rates/:currency/intraday/latest`, and `/v1/gold/:item/intraday/latest`), the API delivers all chart rendering data in a single network round-trip. Crucially, all previous prices are resolved strictly from volatile in-memory caches, memory-resident snapshot metadata, and static catalogs, guaranteeing **0 additional Firestore database reads** (satisfying Principle VII).

---

## 2. In-Memory Data Sources & Resolution Strategies

### 2.1 Equities (Egyptian EGX & Regional GCC Equities)
- **Endpoint**: `GET /v1/stocks/:symbol/intraday/latest?market={egx|saudi|uae|qatar|kuwait}`
- **New Field**: `previousClose: number`
- **Data Source in Memory**:
  - `StockPriceItem` (in `liveSnap.stocks[symbol]`) and `GccStockRate` (in `liveSnap.stocks[symbol]`) already store `previousClose: number`.
  - For Egyptian equities, `stockCatalog` entries also include a static `previousClose` seed value.
- **Resolution Precedence & Fallback Chain**:
  1. `quoteData.previousClose` (if defined and positive number)
  2. `Number((quoteData.price - quoteData.change).toFixed(4))` (if `quoteData.price` and `quoteData.change` are present)
  3. `catalogEntry.previousClose` (from in-memory catalog)
  4. `quoteData.price`
  5. Fallback: `0`
- **Database Reads**: **0**. Sourced from volatile RAM snapshot (`StockQueryService.fetchLatestStockPrices` / `gccStockQueryService.fetchLatestStockPrices`) and catalog constants in memory.

### 2.2 Benchmark Indices (EGX30, TASI, FADX15, etc.)
- **Endpoint**: `GET /v1/indices/:symbol/intraday/latest?market={egx|saudi|uae|qatar}`
- **New Field**: `previousClose: number`
- **Data Source in Memory**:
  - `IndexRate` (in `liveSnap.indices[symbol]`) stores `previousClose: number`.
- **Resolution Precedence & Fallback Chain**:
  1. `quoteData.previousClose` (if defined and positive number)
  2. `Number((quoteData.price - quoteData.change).toFixed(4))` (if `quoteData.price` and `quoteData.change` are present)
  3. `quoteData.price`
  4. Fallback: `0`
- **Database Reads**: **0**. Sourced from volatile RAM snapshot (`IndicesQueryService.fetchLatestIndexPrices`).

### 2.3 Bank Currency Exchange Rates
- **Endpoint**: `GET /v1/rates/:currency/intraday/latest?bank={BANK_CODE}`
- **New Fields**: `previousBuyPrice: number`, `previousSellPrice: number`
- **Data Source in Memory**:
  - Each `CurrencyRate` object in `liveSnapshot.rates` contains:
    - `yesterdayCloseBuy: number | null | undefined`
    - `yesterdayCloseSell: number | null | undefined`
- **Resolution Logic**:
  - **Single Bank (`?bank=cbe`)**:
    - `match = liveSnapshot.rates.find(r => r.bankCode === targetBank && r.currencyCode === upperCurrency)`
    - `previousBuyPrice = (typeof match?.yesterdayCloseBuy === "number" && match.yesterdayCloseBuy > 0) ? match.yesterdayCloseBuy : (buy || 0)`
    - `previousSellPrice = (typeof match?.yesterdayCloseSell === "number" && match.yesterdayCloseSell > 0) ? match.yesterdayCloseSell : (sell || 0)`
  - **Consolidated Market Average (no bank filter)**:
    - Filter all bank quotes for `upperCurrency`.
    - Extract valid positive `yesterdayCloseBuy` and `yesterdayCloseSell` numbers.
    - If valid quotes exist, compute arithmetic average rounded to 2 decimal places (`toFixed(2)`).
    - If no previous close is recorded, fall back to consolidated `buy` and `sell` rates.
  - **Concluded Past Sessions (`!sessionResult.isToday`)**:
    - `HistoryResolutionEngine.getDateSnapshot("rates", sessionResult.date)` resolves the day snapshot from `ActiveWeeklyBucketStore` (RAM) or `flat-cache` (disk, 30-day TTL).
    - If historical bank rate contains previous close or if live snapshot has closing baselines, extract accordingly; otherwise fall back to session `buy` and `sell`.
- **Database Reads**: **0**. Sourced from volatile RAM `SnapshotCache` / `ratesSnapshotService` and cached weekly buckets.

### 2.4 Gold Commodities (Domestic Karats & International Spot Gold)
- **Endpoint**: `GET /v1/gold/:item/intraday/latest` and `GET /v1/gold/intraday/latest`
- **New Fields**: `previousBuyPrice: number`, `previousSellPrice: number` (top-level on single items, and on each nested karat in multi-karat envelope)
- **Data Source in Memory**:
  - `GoldRateItem` in `liveGold.rates[karatKey]` contains `yesterdayBuyPrice: number | null` and `yesterdaySellPrice: number | null` (populated by `gold-routes.ts` via `goldQueryService.fetchGoldPriceHistory(yesterday)` which resolves in `<1ms` from `ActiveWeeklyBucketStore` in RAM or `flat-cache` on disk).
  - International spot gold (`liveGold.international`) contains `price` and `oldPrice`.
- **Resolution Logic**:
  - **Domestic Karats (`24`, `21`, `18`, `pound`)**:
    - `previousBuyPrice = (typeof q.yesterdayBuyPrice === "number" && q.yesterdayBuyPrice > 0) ? q.yesterdayBuyPrice : q.buy`
    - `previousSellPrice = (typeof q.yesterdaySellPrice === "number" && q.yesterdaySellPrice > 0) ? q.yesterdaySellPrice : q.sell`
  - **International Spot Gold (`international`)**:
    - For spot gold, buy and sell are equivalent to spot price:
    - `previousBuyPrice = (typeof intl.oldPrice === "number" && intl.oldPrice > 0) ? intl.oldPrice : (intl.buy ?? intl.price ?? 0)`
    - `previousSellPrice = (typeof intl.oldPrice === "number" && intl.oldPrice > 0) ? intl.oldPrice : (intl.sell ?? intl.price ?? 0)`
  - **Multi-Karat Envelope (`all`)**:
    - Each nested object in `karats` (`"24"`, `"21"`, `"18"`), as well as `"pound"` and `"international"`, will receive `previousBuyPrice` and `previousSellPrice`.
- **Database Reads**: **0**. Sourced from volatile RAM `liveGold` or `ActiveWeeklyBucketStore`.

---

## 3. Alternatives Evaluated

| Alternative | Pros | Cons | Verdict |
|:---|:---|:---|:---|
| **Client-Side Historical Fetching (Android App)** | No backend changes required. | Doubles mobile HTTP round-trips; increases latency; complex calendar logic on client (Sun-Thu vs Mon-Fri trade calendars); high risk of Firestore quota explosion. | **Rejected** |
| **Separate Baseline Endpoint (`GET /v1/*/intraday/baseline`)** | Isolates baseline data from timeline payload. | Still requires 2 round-trips; fragments API contract; increases server route maintenance. | **Rejected** |
| **Injecting Previous Close as an Artificial Intraday Point (`time: "00:00"`)** | No envelope schema change. | Corrupts time-series data; falsifies trading volume/timestamps; breaks charts expecting strict chronological market session bounds. | **Rejected** |
| **Top-Level Envelope Properties on Existing Endpoints (Chosen Approach)** | 1 single network request; zero extra database reads; clean semantic separation between quote baseline and intraday points; 100% backward-compatible. | Minor schema addition across 4 response shapes. | **Accepted** |

---

## 4. Architectural Invariants & Cost Boundary Verification

1. **Principle VII Compliance (Always Free Tier)**:
   - Sourcing strictly leverages in-memory snapshots (`liveSnap`, `liveGold`, `ratesSnapshotService.getSnapshot()`) and static catalog metadata.
   - When resolving concluded past sessions, the system relies on `HistoryResolutionEngine`, which queries `ActiveWeeklyBucketStore` in volatile RAM or `.cache/api-cache.json` with 30-day TTL.
   - Total new Firestore reads per intraday query: **0**.
2. **Principle VI Compliance (Exhaustive Commenting)**:
   - Every modified line in `stock-query-service.ts`, `indices-query-service.ts`, `rates-query-service.ts`, and `gold-query-service.ts` will have an explicit explaining comment.
3. **Principle XIII Compliance (Intraday Memory Segregation)**:
   - No intraday tick points are persisted to disk; only volatile response envelopes in RAM are populated.

