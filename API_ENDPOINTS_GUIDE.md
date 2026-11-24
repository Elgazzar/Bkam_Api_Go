# 📖 Bekam Fe El-Bank — Complete API Endpoint & URL Guide (v2.17.1)

> **API Version**: `2.17.1`  
> **Target Audience**: Mobile app developers, web clients, integration partners, and data consumers.  
> **Base URLs**:
> - **Local Development**: `http://localhost:3001`
> - **Production**: `https://api.bekamfeelbank.com`

---

## 📑 Table of Contents
1. [Authentication & Request Format](#-authentication--request-format)
2. [Bank Exchange Rates API](#-1-bank-exchange-rates-api)
   - [Live Exchange Rates](#11-live-exchange-rates)
   - [Single Bank Dedicated Endpoint](#12-single-bank-dedicated-endpoint)
   - [Single Currency Dedicated Endpoint](#13-single-currency-dedicated-endpoint)
   - [Latest Intraday Rates Timeline](#14-latest-intraday-rates-timeline)
   - [Historical Exchange Rates](#15-historical-exchange-rates)
   - [Historical Range Query](#16-historical-range-query)
   - [Bank Catalog & Metadata](#17-bank-catalog--metadata)
3. [Gold & Precious Metals API](#-2-gold--precious-metals-api)
   - [Live Gold Prices (Domestic & International)](#21-live-gold-prices)
   - [Historical Gold Prices](#22-historical-gold-prices)
   - [Historical Gold Price Range](#23-historical-gold-price-range)
   - [Latest Intraday Gold Timeline](#24-latest-intraday-gold-timeline)
4. [Domestic Petroleum & Fuel API](#-3-domestic-petroleum--fuel-api)
   - [Current Regulated Fuel Prices](#31-current-regulated-fuel-prices)
5. [Regional Equities & Stock Markets API](#-4-regional-equities--stock-markets-api)
   - [Egyptian Stock Exchange (EGX)](#41-egyptian-stock-exchange-egx)
   - [Saudi Stock Exchange (Tadawul)](#42-saudi-stock-exchange-tadawul)
   - [United Arab Emirates Financial Markets (DFM & ADX)](#43-united-arab-emirates-financial-markets-dfm--adx)
   - [Qatar Stock Exchange (QSE)](#44-qatar-stock-exchange-qse)
   - [Boursa Kuwait (KSE)](#45-boursa-kuwait-kse)
   - [Historical Stock Snapshots](#46-historical-stock-snapshots)
   - [Individual Stock Historical Range](#47-individual-stock-historical-range)
   - [Latest Stock Intraday Timeline](#48-latest-stock-intraday-timeline)
   - [Regional Stock Metadata Catalog](#49-regional-stock-metadata-catalog)
6. [Benchmark Indices API](#-5-benchmark-indices-api)
   - [Live Benchmark Indices](#51-live-benchmark-indices)
   - [Index Constituent Equities Reverse-Lookup](#52-index-constituent-equities-reverse-lookup)
   - [Historical Index Prices](#53-historical-index-prices)
   - [Latest Benchmark Index Intraday Timeline](#54-latest-benchmark-index-intraday-timeline)
7. [Stateless Batch Quotes API (Home Favorites)](#-7-stateless-batch-quotes-api-home-favorites)
8. [System Monitoring & Health API](#-8-system-monitoring--health-api)
9. [Comprehensive Reference Cheat Sheets](#-9-comprehensive-reference-cheat-sheets)
   - [Supported Banks (28 Banks)](#91-supported-banks-28-banks)
   - [Supported Currencies (18 Currencies)](#92-supported-currencies-18-currencies)
   - [Benchmark Indices (10 Indices)](#93-benchmark-indices-10-indices)
   - [EGX Thematic Groups](#94-egx-thematic-groups)

---

## 🔑 Authentication & Request Format

All endpoints (except `GET /v1/health`) require authentication. You can provide your API key in **one of two ways**:

### Option 1: HTTP Header (Recommended)
```http
x-api-key: your-api-key-here
```

### Option 2: Query Parameter
Append `?apikey=your-api-key-here` or `?apiKey=your-api-key-here` to any request URL:
```text
https://api.bekamfeelbank.com/v1/rates?apikey=your-api-key-here
```

---

## 🏦 1. Bank Exchange Rates API

Real-time buy and sell exchange rates across **28 Egyptian commercial and investment banks** for **18 foreign currencies**, updated dynamically.

### 1.1 Live Exchange Rates
| Purpose | URL / Query Example | cURL Example |
|:---|:---|:---|
| **All Banks & All Currencies** | `GET /v1/rates` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates"` |
| **Filter by Bank** | `GET /v1/rates?bank=CIB` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates?bank=CIB"` |
| **Filter by Currency** | `GET /v1/rates?currency=USD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates?currency=USD"` |
| **Filter by Bank & Currency** | `GET /v1/rates?bank=NBE&currency=EUR` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates?bank=NBE&currency=EUR"` |
| **Filter by Arab Currencies** | `GET /v1/rates?currency=SAR` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates?currency=SAR"` |
| **Filter by Kuwaiti Dinar** | `GET /v1/rates?currency=KWD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates?currency=KWD"` |
| **Filter by UAE Dirham** | `GET /v1/rates?currency=AED` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates?currency=AED"` |

### 1.2 Single Bank Dedicated Endpoint
Dedicated route retrieving all currency exchange rates for a single target financial institution:
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Commercial International Bank** | `GET /v1/banks/CIB/rates` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/banks/CIB/rates"` |
| **National Bank of Egypt** | `GET /v1/banks/NBE/rates` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/banks/NBE/rates"` |
| **Banque Misr** | `GET /v1/banks/MISR/rates` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/banks/MISR/rates"` |
| **QNB Alahli** | `GET /v1/banks/QNB/rates` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/banks/QNB/rates"` |
| **Single Bank + Currency Filter** | `GET /v1/banks/CIB/rates?currency=USD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/banks/CIB/rates?currency=USD"` |

### 1.3 Single Currency Dedicated Endpoint
Dedicated route retrieving exchange rates across all 28 banks for a single specified currency:
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **US Dollar (USD)** | `GET /v1/rates/USD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/USD"` |
| **Euro (EUR)** | `GET /v1/rates/EUR` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/EUR"` |
| **Saudi Riyal (SAR)** | `GET /v1/rates/SAR` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/SAR"` |
| **British Pound (GBP)** | `GET /v1/rates/GBP` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/GBP"` |
| **Single Currency + Bank Filter**| `GET /v1/rates/USD?bank=CIB` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/USD?bank=CIB"` |

### 1.4 Latest Intraday Rates Timeline
Dedicated endpoint retrieving the latest trading session's intraday price timeline for a specific foreign currency pair across all banks (consolidated average) or filtered to a single financial institution:
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Consolidated USD Intraday Timeline** | `GET /v1/rates/USD/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/USD/intraday/latest"` |
| **CIB USD Intraday Timeline** | `GET /v1/rates/USD/intraday/latest?bank=cib` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/USD/intraday/latest?bank=cib"` |
| **EUR Intraday Timeline** | `GET /v1/rates/EUR/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/EUR/intraday/latest"` |
| **SAR Intraday Timeline** | `GET /v1/rates/SAR/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/SAR/intraday/latest"` |

> **Adaptive Caching**: Returns `Cache-Control: public, max-age=60` during active trading sessions today, and `Cache-Control: public, max-age=900` during off-hours / weekends.  
> **Baseline Pricing (Feature 100)**: Response includes `previousBuyPrice` and `previousSellPrice` representing the preceding banking session's close to enable mobile charting baselines and percentage change calculations in a single network round-trip with 0 database reads.

```json
{
  "currency": "USD",
  "bank": "cbe",
  "date": "2026-11-05",
  "buy": 48.65,
  "sell": 48.75,
  "previousBuyPrice": 48.60,
  "previousSellPrice": 48.70,
  "updatedAt": "2026-11-05T14:30:00.000Z",
  "intraday": [
    { "time": "09:30", "buy": 48.60, "sell": 48.70 },
    { "time": "12:15", "buy": 48.65, "sell": 48.75 }
  ]
}
```

### 1.5 Historical Exchange Rates
Fetch the official closing snapshot recorded for any past calendar date (`YYYY-MM-DD`):
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **All Banks for a Specific Date** | `GET /v1/rates/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/history/2026-10-22"` |
| **Historical Date + Currency Filter** | `GET /v1/rates/history/:date?currency=USD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/history/2026-10-22?currency=USD"` |
| **Historical Date + Bank Filter** | `GET /v1/rates/history/:date?bank=CIB` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/history/2026-10-22?bank=CIB"` |
| **Historical Date + Bank & Currency** | `GET /v1/rates/history/:date?bank=CIB&currency=USD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/history/2026-10-22?bank=CIB&currency=USD"` |

### 1.6 Historical Range Query
Fetch exchange rates across a date window for timeline graphing and historical trend analysis:
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Multi-Day Date Range** | `GET /v1/rates/history/range?startDate=YYYY-MM-DD&endDate=YYYY-MM-DD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/history/range?startDate=2026-10-01&endDate=2026-10-22"` |
| **Range Filtered by Bank & Currency** | `GET /v1/rates/history/range?startDate=...&endDate=...&bank=CIB&currency=USD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/rates/history/range?startDate=2026-10-01&endDate=2026-10-22&bank=CIB&currency=USD"` |

### 1.7 Unified Market Catalog & Platform Metadata
Discover all registered banks, currencies, commodities, benchmark indices, and inventory counts:
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Unified Market Catalog** | `GET /v1/catalog` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/catalog"` |

*Returns complete discoverable platform inventory: `banks`, `currencies`, `commodities`, `indices`, and `counts` (`totalBanks`, `totalCurrencies`, `totalCommodities`, `totalIndices`, `totalSecurities`). HTTP responses are cached with `Cache-Control: public, max-age=3600`.*

---

## 🪙 2. Gold & Precious Metals API

Real-time domestic Egyptian bullion and coin rates in EGP (24K, 22K, 21K, 18K, 14K, 12K, 9K, Gold Ounce, Gold Pound, BTC/Ingot coins) plus **International Spot Gold in USD/oz**.

### 2.1 Live Gold Prices
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Current Gold Prices (Local + Global)** | `GET /v1/gold` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold"` |

*Returns domestic buy/sell prices, yesterday's comparison close, 24-hour change %, and live TradingView international spot price.*

### 2.2 Historical Gold Prices
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Historical Gold by Single Date** | `GET /v1/gold/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/history/2026-10-22"` |

### 2.3 Historical Gold Price Range
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Historical Gold Range (Multi-Day)** | `GET /v1/gold/history/range?startDate=YYYY-MM-DD&endDate=YYYY-MM-DD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/history/range?startDate=2026-10-01&endDate=2026-10-22"` |

### 2.4 Latest Intraday Gold Timeline
Retrieve the latest trading session's intraday price points for gold commodities (24k, 21k, 18k, gold pound, international spot):
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **All Karats & Commodities Intraday** | `GET /v1/gold/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/intraday/latest"` |
| **Explicit All Karats** | `GET /v1/gold/all/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/all/intraday/latest"` |
| **21 Karat Gold Intraday** | `GET /v1/gold/21/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/21/intraday/latest"` |
| **24 Karat Gold Intraday** | `GET /v1/gold/24/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/24/intraday/latest"` |
| **18 Karat Gold Intraday** | `GET /v1/gold/18/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/18/intraday/latest"` |
| **Gold Pound Intraday** | `GET /v1/gold/pound/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/pound/intraday/latest"` |
| **International Spot Gold Intraday** | `GET /v1/gold/international/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gold/international/intraday/latest"` |

> **Adaptive Caching**: Returns `Cache-Control: public, max-age=60` during active trading sessions today, and `Cache-Control: public, max-age=900` during off-hours / weekends.  
> **Baseline Pricing (Feature 100)**: Response includes `previousBuyPrice` and `previousSellPrice` on single commodity responses and within each nested karat (`24`, `21`, `18`, `pound`, `international`) in the all-commodities envelope.

```json
{
  "item": "21",
  "date": "2026-11-05",
  "currency": "EGP",
  "buy": 3650.0,
  "sell": 3670.0,
  "previousBuyPrice": 3630.0,
  "previousSellPrice": 3650.0,
  "updatedAt": "2026-11-05T14:30:00.000Z",
  "intraday": [
    { "time": "10:15", "price": 3645.0 },
    { "time": "13:00", "price": 3650.0 }
  ]
}
```

---

## ⛽ 3. Domestic Petroleum & Fuel API

Official regulated domestic Egyptian petroleum and fuel prices denominated in EGP per liter / cubic meter.

### 3.1 Current Regulated Fuel Prices
| Purpose | URL Pattern | Example Request |
|:---|:---|:---|
| **Domestic Gas & Petroleum Prices** | `GET /v1/gas` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/gas"` |

*Covers Octane Benzene 95, Benzene 92, Benzene 80, Diesel (Solar), Kerosene, and Compressed Natural Gas (CNG).*

---

## 📈 4. Regional Equities & Stock Markets API

Unified multi-market equity quoting engine serving the **Egyptian Stock Exchange (EGX)** and GCC markets: **Saudi Arabia (Tadawul)**, **United Arab Emirates (DFM & ADX)**, and **Qatar (QSE)**.

### 4.1 Egyptian Stock Exchange (EGX)
Currency: `EGP` | Exchange: `EGX` | Operating Hours: `Sunday - Thursday, 09:00 - 15:30 Cairo Time`  
*Includes newly integrated EGX listings: Banque du Caire (`BQDC`, alias `BDC`) and MNT-Halan (`HALN`, alias `MNT`).*

| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **All EGX Equities (102 Stocks)** | `GET /v1/stocks` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks"` |
| **Explicit Country Parameter** | `GET /v1/stocks?country=egypt` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=egypt"` |
| **Filter: EGX 30 Blue-Chips Index** | `GET /v1/stocks?index=EGX30` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?index=EGX30"` |
| **Filter: EGX 70 EWI Small/Mid-Caps** | `GET /v1/stocks?index=EGX70` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?index=EGX70"` |
| **Filter: EGX 100 EWI Broad Market** | `GET /v1/stocks?index=EGX100` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?index=EGX100"` |
| **Filter: EGX 33 Sharia Index** | `GET /v1/stocks?index=EGX33` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?index=EGX33"` |
| **Filter: Sharia Group Equities** | `GET /v1/stocks?group=Sharia` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?group=Sharia"` |
| **Filter: Dividend Aristocrats** | `GET /v1/stocks?group=Dividend` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?group=Dividend"` |
| **Filter: Technology Equities** | `GET /v1/stocks?group=Technology` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?group=Technology"` |
| **Bypass Cache (Force Fresh Live Scrape)** | `GET /v1/stocks?forceRefresh=true` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?forceRefresh=true"` |
| **Combined Index & Fresh Scrape** | `GET /v1/stocks?index=EGX30&forceRefresh=true` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?index=EGX30&forceRefresh=true"` |

### 4.2 Saudi Stock Exchange (Tadawul)
Currency: `SAR` | Exchange: `TADAWUL` | Operating Hours: `Sunday - Thursday, 09:00 - 16:00 Riyadh Time`

> **Symbol Standardization & Preserved Numeric Identifiers**:
> All Saudi Tadawul stock symbols are standardized to uppercase alphanumeric codes (e.g. `symbol: "ARAMCO"`, `symbol: "ALRAJHI"`, `symbol: "SEC"`, `symbol: "ZAIN"`). The official Tadawul 4-digit code is preserved in `numericCode` (e.g. `numericCode: "2222"` for Aramco, `numericCode: "5110"` for Saudi Electricity Co.). The returned `stocks` dictionary covers all **271 listed Main Market equities and REITs** with zero duplicate entries.
> 
> **Dynamic Multi-Field Sorting**:
> Use `?sortBy=<field>&sortOrder=<asc|desc>` to sort returned quotes:
> - `sortBy=price`: Numerical sort by current quote price (default order: `desc`).
> - `sortBy=change` / `sortBy=changePercent`: Numerical sort by percentage change (highest gainers first with `desc`, biggest losers first with `asc`).
> - `sortBy=volume`: Numerical sort by cumulative trading volume (`desc`).
> - `sortBy=name`: Alphabetical sort by English corporate name (`asc` or `desc`).
> - `sortBy=index`: Ranked by benchmark index tier (`TADAWUL_LARGE_CAP` / `TASI50` / `MT30` -> `TADAWUL_MID_CAP` -> `TADAWUL_SMALL_CAP` -> other).

| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **All Saudi Equities & REITs (271 Stocks)** | `GET /v1/stocks/saudi` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi"` |
| **Via Country Parameter** | `GET /v1/stocks?country=saudi` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=saudi"` |
| **Via Country Code `sa`** | `GET /v1/stocks?country=sa` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=sa"` |
| **Filter: Broad Market TASI (271 Stocks)** | `GET /v1/stocks/saudi?index=TASI` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=TASI"` |
| **Filter: MT30 Blue-Chips Index (30 Stocks)** | `GET /v1/stocks/saudi?index=MT30` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=MT30"` |
| **Filter: TASI 50 Benchmark Index (50 Stocks)** | `GET /v1/stocks/saudi?index=TASI50` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=TASI50"` |
| **Filter: Tadawul Large Cap (40 Stocks)** | `GET /v1/stocks/saudi?index=TADAWUL_LARGE_CAP` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=TADAWUL_LARGE_CAP"` |
| **Filter: Tadawul Mid Cap (60 Stocks)** | `GET /v1/stocks/saudi?index=TADAWUL_MID_CAP` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=TADAWUL_MID_CAP"` |
| **Filter: Tadawul Small Cap (171 Stocks)** | `GET /v1/stocks/saudi?index=TADAWUL_SMALL_CAP` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=TADAWUL_SMALL_CAP"` |
| **Filter: TASI Shariah Index (243 Compliant Equities)** | `GET /v1/stocks/saudi?index=TASI_SHARIAH` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=TASI_SHARIAH"` |
| **Sort: Top Gainers (Highest % Change First)** | `GET /v1/stocks/saudi?sortBy=changePercent&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?sortBy=changePercent&sortOrder=desc"` |
| **Sort: Most Active by Volume** | `GET /v1/stocks/saudi?sortBy=volume&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?sortBy=volume&sortOrder=desc"` |
| **Sort: By Index Tier (Large Cap > Mid Cap > Small Cap)** | `GET /v1/stocks/saudi?sortBy=index&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?sortBy=index&sortOrder=desc"` |
| **Filter by Sector / Group (Banking)** | `GET /v1/stocks/saudi?group=Banking` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?group=Banking"` |
| **Filter by Sector / Group (Energy)** | `GET /v1/stocks/saudi?group=Energy` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?group=Energy"` |
| **Combined Index & Sort** | `GET /v1/stocks/saudi?index=TASI50&sortBy=price&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?index=TASI50&sortBy=price&sortOrder=desc"` |
| **Force Fresh Scrape** | `GET /v1/stocks/saudi?forceRefresh=true` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi?forceRefresh=true"` |

### 4.3 United Arab Emirates Financial Markets (DFM & ADX)
Currency: `AED` | Exchanges: `DFM`, `ADX` | Operating Hours: `Monday - Friday, 09:00 - 16:00 Dubai Time`

> **Constituent Inventory (158 Total Equities)**:
> Tracks all 158 listed equities and REITs across Abu Dhabi Securities Exchange (ADX, 89 stocks) and Dubai Financial Market (DFM & NASDAQ Dubai, 69 stocks), with zero duplicate entries.
>
> **Dynamic Multi-Field Sorting**:
> Use `?sortBy=<field>&sortOrder=<asc|desc>`:
> - `sortBy=price`: Numerical sort by current quote price (default order: `desc`).
> - `sortBy=change` / `sortBy=changePercent`: Numerical sort by percentage change (`desc` or `asc`).
> - `sortBy=volume`: Numerical sort by cumulative trading volume (`desc`).
> - `sortBy=name`: Alphabetical sort by English corporate name (`asc` or `desc`).
> - `sortBy=index`: Ranked by benchmark index tier (`FADX15` -> `DFMGI` -> `FADGI`).

| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **All UAE Equities (DFM + ADX - 158 Stocks)** | `GET /v1/stocks/uae` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae"` |
| **Via Country Parameter** | `GET /v1/stocks?country=uae` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=uae"` |
| **Via Country Code `ae`** | `GET /v1/stocks?country=ae` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=ae"` |
| **Filter by Dubai Financial Market (69 Stocks)** | `GET /v1/stocks/uae?exchange=DFM` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?exchange=DFM"` |
| **Filter by Abu Dhabi Securities Exchange (89 Stocks)** | `GET /v1/stocks/uae?exchange=ADX` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?exchange=ADX"` |
| **Filter by DFM General Index (69 Stocks)** | `GET /v1/stocks/uae?index=DFMGI` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?index=DFMGI"` |
| **Filter by FTSE ADX General Index (89 Stocks)** | `GET /v1/stocks/uae?index=FADGI` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?index=FADGI"` |
| **Filter by FTSE ADX 15 Index (15 Blue Chips)** | `GET /v1/stocks/uae?index=FADX15` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?index=FADX15"` |
| **Sort: Top Gainers (% Change)** | `GET /v1/stocks/uae?sortBy=changePercent&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?sortBy=changePercent&sortOrder=desc"` |
| **Sort: Most Active by Volume** | `GET /v1/stocks/uae?sortBy=volume&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?sortBy=volume&sortOrder=desc"` |
| **Force Fresh Scrape** | `GET /v1/stocks/uae?forceRefresh=true` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae?forceRefresh=true"` |

### 4.4 Qatar Stock Exchange (QSE)
Currency: `QAR` | Exchange: `QSE` | Operating Hours: `Sunday - Thursday, 08:30 - 14:15 Doha Time`

> **Constituent Inventory (55 Total Equities)**:
> Tracks all 55 listed equities on the Qatar Stock Exchange.
>
> **Dynamic Multi-Field Sorting**:
> Use `?sortBy=<field>&sortOrder=<asc|desc>` (`price`, `change`, `changePercent`, `volume`, `name`, `index`).

| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **All Qatar Equities (QSE - 55 Stocks)** | `GET /v1/stocks/qatar` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar"` |
| **Via Country Parameter** | `GET /v1/stocks?country=qatar` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=qatar"` |
| **Via Country Code `qa`** | `GET /v1/stocks?country=qa` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=qa"` |
| **Filter by QE 20 Index (20 Blue Chips)** | `GET /v1/stocks/qatar?index=QSE_INDEX` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar?index=QSE_INDEX"` |
| **Filter by QSE All Share Index (55 Stocks)** | `GET /v1/stocks/qatar?index=QSE_ALL` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar?index=QSE_ALL"` |
| **Sort: Top Gainers (% Change)** | `GET /v1/stocks/qatar?sortBy=changePercent&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar?sortBy=changePercent&sortOrder=desc"` |
| **Sort: Most Active by Volume** | `GET /v1/stocks/qatar?sortBy=volume&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar?sortBy=volume&sortOrder=desc"` |
| **Force Fresh Scrape** | `GET /v1/stocks/qatar?forceRefresh=true` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar?forceRefresh=true"` |

### 4.5 Boursa Kuwait (KSE)
Currency: `KWD` | Exchange: `KSE` | Operating Hours: `Sunday - Thursday, 09:00 - 12:30 Kuwait Time`

> **Constituent Inventory (136 Total Equities)**:
> Tracks all 136 listed equities on Boursa Kuwait, spanning the Premier Market (35 blue chips) and Main Market (101 growth equities).
>
> **Dynamic Multi-Field Sorting**:
> Use `?sortBy=<field>&sortOrder=<asc|desc>` (`price`, `change`, `changePercent`, `volume`, `name`, `index`).

| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **All Kuwait Equities (136 Stocks)** | `GET /v1/stocks/kuwait` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait"` |
| **Via Country Parameter** | `GET /v1/stocks?country=kuwait` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=kuwait"` |
| **Via Country Code `kw`** | `GET /v1/stocks?country=kw` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks?country=kw"` |
| **Filter by Premier Market Index (35 Stocks)** | `GET /v1/stocks/kuwait?index=BK_PREMIER` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait?index=BK_PREMIER"` |
| **Filter by Main Market Index (101 Stocks)** | `GET /v1/stocks/kuwait?index=BK_MAIN` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait?index=BK_MAIN"` |
| **Filter by All Share Index (136 Stocks)** | `GET /v1/stocks/kuwait?index=BK_ALL_SHARE` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait?index=BK_ALL_SHARE"` |
| **Sort: Top Gainers (% Change)** | `GET /v1/stocks/kuwait?sortBy=changePercent&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait?sortBy=changePercent&sortOrder=desc"` |
| **Sort: Most Active by Volume** | `GET /v1/stocks/kuwait?sortBy=volume&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait?sortBy=volume&sortOrder=desc"` |
| **Sort: By Index Tier (Premier > Main)** | `GET /v1/stocks/kuwait?sortBy=index&sortOrder=desc` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait?sortBy=index&sortOrder=desc"` |
| **Force Fresh Scrape** | `GET /v1/stocks/kuwait?forceRefresh=true` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait?forceRefresh=true"` |

### 4.6 Historical Stock Snapshots
Retrieve full historical market closing prices for any previous calendar day (`YYYY-MM-DD`):
| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **EGX Historical Snapshot** | `GET /v1/stocks/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/history/2026-10-22"` |
| **EGX Historical Filtered by Index** | `GET /v1/stocks/history/:date?index=EGX30` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/history/2026-10-22?index=EGX30"` |
| **Saudi Historical Snapshot** | `GET /v1/stocks/saudi/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi/history/2026-10-22"` |
| **Saudi Historical Filtered by Index** | `GET /v1/stocks/saudi/history/:date?index=TASI` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi/history/2026-10-22?index=TASI"` |
| **UAE Historical Snapshot** | `GET /v1/stocks/uae/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae/history/2026-10-22"` |
| **UAE Historical Filtered by Exchange** | `GET /v1/stocks/uae/history/:date?exchange=DFM` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/uae/history/2026-10-22?exchange=DFM"` |
| **Qatar Historical Snapshot** | `GET /v1/stocks/qatar/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar/history/2026-10-22"` |
| **Qatar Historical Filtered by Index** | `GET /v1/stocks/qatar/history/:date?index=QSE_INDEX`| `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/qatar/history/2026-10-22?index=QSE_INDEX"`|
| **Kuwait Historical Snapshot** | `GET /v1/stocks/kuwait/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait/history/2026-10-22"` |
| **Kuwait Historical Filtered by Index** | `GET /v1/stocks/kuwait/history/:date?index=BK_PREMIER`| `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait/history/2026-10-22?index=BK_PREMIER"`|

### 4.7 Individual Stock Historical Range
Retrieve historical price history across a date window for an individual stock ticker symbol:
| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **CIB (COMI) Range** | `GET /v1/stocks/COMI/history/range?start=YYYY-MM-DD&end=YYYY-MM-DD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/COMI/history/range?start=2026-10-01&end=2026-10-22"` |
| **Saudi Aramco (by Symbol or Numeric Code)** | `GET /v1/stocks/saudi/:symbol/history/range?start=...&end=...` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/saudi/ARAMCO/history/range?start=2026-10-01&end=2026-10-22"` *(or `/2222/`)* |
| **National Bank of Kuwait (NBK)** | `GET /v1/stocks/kuwait/:symbol/history/range?start=...&end=...` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/kuwait/NBK/history/range?start=2026-10-01&end=2026-10-22"` |

### 4.8 Latest Stock Intraday Timeline
Retrieve the latest trading session's intraday price points and latest quote metrics for an individual stock ticker across regional markets:
| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **CIB (COMI) Latest Intraday (EGX)** | `GET /v1/stocks/COMI/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/COMI/intraday/latest"` |
| **Fawry (FWRY) Latest Intraday (EGX)** | `GET /v1/stocks/FWRY/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/FWRY/intraday/latest"` |
| **Saudi Aramco (Tadawul)** | `GET /v1/stocks/ARAMCO/intraday/latest?market=saudi` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/ARAMCO/intraday/latest?market=saudi"` |
| **Saudi Aramco by Numeric Code** | `GET /v1/stocks/2222/intraday/latest?market=saudi` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/2222/intraday/latest?market=saudi"` |
| **Emaar Properties (DFM / UAE)** | `GET /v1/stocks/EMAAR/intraday/latest?market=uae` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/EMAAR/intraday/latest?market=uae"` |
| **Qatar National Bank (QNBK / QSE)** | `GET /v1/stocks/QNBK/intraday/latest?market=qatar` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/QNBK/intraday/latest?market=qatar"` |
| **National Bank of Kuwait (NBK / KSE)** | `GET /v1/stocks/NBK/intraday/latest?market=kuwait` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/NBK/intraday/latest?market=kuwait"` |

> **Graceful Inactive Security Handling**: If a catalog security had 0 trades executed during the resolved session, the endpoint returns `200 OK` with `intraday: []` and the latest closing quote (never a 404).  
> **Adaptive Caching**: Returns `Cache-Control: public, max-age=60` during active trading sessions today, and `Cache-Control: public, max-age=900` during off-hours / weekends.  
> **Baseline Pricing (Feature 100)**: Response includes `previousClose` representing the previous trading session's official closing price for rendering chart limit lines with 0 database reads.

```json
{
  "symbol": "COMI",
  "date": "2026-11-05",
  "currency": "EGP",
  "price": 84.5,
  "previousClose": 83.0,
  "change": 1.5,
  "changePercent": 1.81,
  "updatedAt": "2026-11-05T14:30:00.000Z",
  "intraday": [
    { "time": "10:05", "price": 83.5 },
    { "time": "11:30", "price": 84.5 }
  ]
}
```

### 4.9 Regional Stock Metadata Catalog
Discover supported equities, company codes, ISINs, sectors, logos, and trading schedules:
| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **All Regional Equities (Full GCC - 620 Stocks)** | `GET /v1/stocks/regional/catalog` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/regional/catalog"` |
| **Saudi Catalog (271 Stocks)** | `GET /v1/stocks/regional/catalog?country=saudi` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/regional/catalog?country=saudi"` |
| **UAE Catalog (158 Stocks)** | `GET /v1/stocks/regional/catalog?country=uae` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/regional/catalog?country=uae"` |
| **Qatar Catalog (55 Stocks)** | `GET /v1/stocks/regional/catalog?country=qatar` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/regional/catalog?country=qatar"` |
| **Kuwait Catalog (136 Stocks)** | `GET /v1/stocks/regional/catalog?country=kuwait` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/stocks/regional/catalog?country=kuwait"` |

---

## 📊 5. Benchmark Indices API

Real-time price quotes, point changes, and constituent equity reverse-lookups for official regional financial benchmark indices.

### 5.1 Live Benchmark Indices
| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **EGX Indices (Default)** | `GET /v1/indices` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices"` |
| **Explicit Egypt Indices** | `GET /v1/indices?market=egx` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices?market=egx"` |
| **Saudi Tadawul Indices (TASI, MT30, TASI50, etc.)** | `GET /v1/indices/saudi` *(or `?market=saudi`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/saudi"` |
| **UAE Indices (DFMGI, FADGI, FADX15)** | `GET /v1/indices/uae` *(or `?market=uae`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/uae"` |
| **Qatar Indices (QSE_INDEX, QSE_ALL)** | `GET /v1/indices/qatar` *(or `?market=qatar`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/qatar"` |
| **Kuwait Indices (BK_ALL_SHARE, BK_PREMIER, BK_MAIN)** | `GET /v1/indices/kuwait` *(or `?market=kuwait`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/kuwait"` |
| **Force Fresh Scrape** | `GET /v1/indices?forceRefresh=true` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices?forceRefresh=true"` |

### 5.2 Index Constituent Equities Reverse-Lookup
Get the exact list of constituent member companies that make up a benchmark index along with their real-time prices. Accessible via both `/v1/indices/:code/stocks` and the canonical alias `/v1/indices/:code/constituents`:
| Target Index | Dedicated URL Pattern | Example Request |
|:---|:---|:---|
| **EGX 30 Blue-Chips** | `GET /v1/indices/EGX30/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/EGX30/constituents"` |
| **EGX 70 EWI Small/Mid-Caps** | `GET /v1/indices/EGX70/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/EGX70/constituents"` |
| **EGX 100 EWI Broad Composite** | `GET /v1/indices/EGX100/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/EGX100/constituents"` |
| **EGX 33 Sharia Index** | `GET /v1/indices/EGX33/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/EGX33/constituents"` |
| **Saudi TASI All Share Index** | `GET /v1/indices/TASI/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/TASI/constituents"` |
| **Saudi MSCI Tadawul 30 (MT30)** | `GET /v1/indices/MT30/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/MT30/constituents"` |
| **Dubai DFM General Index** | `GET /v1/indices/DFMGI/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/DFMGI/constituents"` |
| **Abu Dhabi FTSE ADX General** | `GET /v1/indices/FADGI/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/FADGI/constituents"` |
| **Abu Dhabi FTSE ADX 15** | `GET /v1/indices/FADX15/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/FADX15/constituents"` |
| **Qatar QE Index** | `GET /v1/indices/QSE_INDEX/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/QSE_INDEX/constituents"` |
| **Qatar All Share Index** | `GET /v1/indices/QSE_ALL/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/QSE_ALL/constituents"` |
| **Kuwait Premier Market (35 Stocks)** | `GET /v1/indices/BK_PREMIER/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/BK_PREMIER/constituents"` |
| **Kuwait Main Market (101 Stocks)** | `GET /v1/indices/BK_MAIN/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/BK_MAIN/constituents"` |
| **Kuwait All Share Index (136 Stocks)** | `GET /v1/indices/BK_ALL_SHARE/constituents` *(or `/stocks`)* | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/BK_ALL_SHARE/constituents"` |

### 5.3 Historical Index Prices
| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **Historical Index by Single Date** | `GET /v1/indices/history/:date` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/history/2026-10-22"` |
| **Historical Index Range** | `GET /v1/indices/history/range?startDate=YYYY-MM-DD&endDate=YYYY-MM-DD` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/history/range?startDate=2026-10-01&endDate=2026-10-22"` |

### 5.4 Latest Benchmark Index Intraday Timeline
Retrieve the latest trading session's intraday price points and latest quote metrics for an official regional benchmark index:
| Purpose | URL / Query Pattern | Example Request |
|:---|:---|:---|
| **EGX 30 Benchmark Latest Intraday** | `GET /v1/indices/EGX30/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/EGX30/intraday/latest"` |
| **EGX 70 EWI Latest Intraday** | `GET /v1/indices/EGX70/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/EGX70/intraday/latest"` |
| **EGX 100 EWI Latest Intraday** | `GET /v1/indices/EGX100/intraday/latest` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/EGX100/intraday/latest"` |
| **Saudi TASI All Share Intraday** | `GET /v1/indices/TASI/intraday/latest?market=saudi` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/TASI/intraday/latest?market=saudi"` |
| **Dubai DFM General Index Intraday** | `GET /v1/indices/DFMGI/intraday/latest?market=uae` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/DFMGI/intraday/latest?market=uae"` |
| **Abu Dhabi FTSE ADX 15 Intraday** | `GET /v1/indices/FADX15/intraday/latest?market=uae` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/FADX15/intraday/latest?market=uae"` |
| **Qatar QE Index Intraday** | `GET /v1/indices/QSE_INDEX/intraday/latest?market=qatar` | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/indices/QSE_INDEX/intraday/latest?market=qatar"` |

> **Adaptive Caching**: Returns `Cache-Control: public, max-age=60` during active trading sessions today, and `Cache-Control: public, max-age=900` during off-hours / weekends.  
> **Baseline Pricing (Feature 100)**: Response includes `previousClose` representing the benchmark index's previous closing level with 0 database reads.

```json
{
  "symbol": "EGX30",
  "date": "2026-11-05",
  "currency": "EGP",
  "price": 30500.25,
  "previousClose": 30250.0,
  "change": 250.25,
  "changePercent": 0.83,
  "updatedAt": "2026-11-05T14:30:00.000Z",
  "intraday": [
    { "time": "10:05", "price": 30300.0 },
    { "time": "12:00", "price": 30500.25 }
  ]
}
```

---

---

## ⚡ 7. Stateless Batch Quotes API (Home Favorites)

> **Detailed Integration Guide**: See [docs/BATCH_QUOTES_GUIDE.md](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/docs/BATCH_QUOTES_GUIDE.md) for full Kotlin/Retrofit Android snippets and Swift/iOS implementations.

Resolves live quotes for arbitrary user favorites across **all asset domains** (Bank Currencies `<BANK>_<CURRENCY>`, Gold Carats `21k`, `pound`, Indices `EGX30`, `TASI`, and Stocks `COMI`, `1120`, `ARAMCO`) in a **single network request** with **0 Firestore reads**, served synchronously in $<0.01\text{ ms}$ via the unified `LiveSnapshotCoordinator`.

| Endpoint | Method | Auth | Description | Example Request |
|:---|:---:|:---:|:---|:---|
| `/v1/quotes` | `POST` | **Required** | Submit JSON array of item IDs (`{"items": [...]}`) to receive a unified keyed map of quotes | `curl -X POST "http://localhost:3001/v1/quotes" -H "Content-Type: application/json" -H "x-api-key: $KEY" -d '{"items": ["CBE_USD", "21k", "pound", "EGX30", "COMI", "1120"]}'` |
| `/v1/quotes` | `GET` | **Required** | Comma-separated query list (`?items=...`) with automatic 15-second Cache-Control and ETag validation | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/quotes?items=CBE_USD,21k,EGX30,COMI"` |

---

## 🩺 8. System Monitoring & Health API

| Endpoint | Method | Auth | Description | Example Request |
|:---|:---:|:---:|:---|:---|
| `/v1/health` | `GET` | **Public** | Service liveness, Firestore connectivity, and scraper subsystem status | `curl "http://localhost:3001/v1/health"` |
| `/v1/metrics` | `GET` | **Required** | Real-time traffic metrics, route hits, response latency, and error breakdown | `curl -H "x-api-key: $KEY" "http://localhost:3001/v1/metrics"` |

---

## 📚 9. Comprehensive Reference Cheat Sheets

### 9.1 Supported Banks (28 Banks)
Use any of these codes in `?bank={CODE}` or `/v1/banks/{CODE}/rates`:

| Code | English Name | Arabic Name |
|:---|:---|:---|
| `CIB` | Commercial International Bank | البنك التجاري الدولي |
| `NBE` | National Bank of Egypt | البنك الأهلي المصري |
| `MISR` | Banque Misr | بنك مصر |
| `QNB` | QNB Alahli | بنك قطر الوطني الأهلي |
| `HDB` | Housing and Development Bank | بنك التعمير والإسكان |
| `CBE` | Central Bank of Egypt | البنك المركزي المصري |
| `BDC` | Banque du Caire | بنك القاهرة |
| `EGBANK` | Egyptian Gulf Bank | البنك المصري الخليجي |
| `NBK` | National Bank of Kuwait - Egypt | بنك الكويت الوطني - مصر |
| `HSBC` | HSBC Bank Egypt | بنك إتش إس بي سي مصر |
| `ENBD` | Emirates NBD Egypt | بنك الإمارات دبي الوطني |
| `AAIB` | Arab African International Bank | البنك العربي الأفريقي الدولي |
| `ALEX` | Bank of Alexandria | بنك الإسكندرية |
| `ADIB` | Abu Dhabi Islamic Bank - Egypt | مصرف أبوظبي الإسلامي - مصر |
| `FAISAL` | Faisal Islamic Bank of Egypt | بنك فيصل الإسلامي المصري |
| `ALBARAKA` | Al Baraka Bank Egypt | بنك البركة مصر |
| `ADCB` | Abu Dhabi Commercial Bank Egypt | بنك أبوظبي التجاري مصر |
| `ABK` | Al Ahli Bank of Kuwait - Egypt | البنك الأهلي الكويتي - مصر |
| `SAIB` | Société Arabe Internationale de Banque | بنك الشركة المصرفية العربية الدولية |
| `AIB` | Arab International Bank | المصرف العربي الدولي |
| `NXT` | Next Bank (formerly AI Bank) | بنك نكست |
| `SCB` | Suez Canal Bank | بنك قناة السويس |
| `KFH` | Kuwait Finance House Egypt | بيت التمويل الكويتي |
| `FABMISR` | First Abu Dhabi Bank Misr | بنك أبوظبي الأول مصر |
| `UB` | The United Bank of Egypt | المصرف المتحد |
| `EALB` | Egyptian Arab Land Bank | البنك العقاري المصري العربي |
| `CAE` | Crédit Agricole Egypt | كريدي أجريكول مصر |
| `IDB` | Industrial Development Bank | بنك التنمية الصناعية |

---

### 7.2 Supported Currencies (18 Currencies)
Use any of these ISO codes in `?currency={CODE}` or `/v1/rates/{CODE}`:

| Code | Currency Name | Symbol / Denomination |
|:---|:---|:---|
| `USD` | United States Dollar | $ |
| `EUR` | Euro | € |
| `GBP` | British Pound Sterling | £ |
| `SAR` | Saudi Riyal | ر.س |
| `AED` | United Arab Emirates Dirham | د.إ |
| `KWD` | Kuwaiti Dinar | د.ك |
| `QAR` | Qatari Riyal | ر.ق |
| `BHD` | Bahraini Dinar | .د.ب |
| `OMR` | Omani Rial | ر.ع. |
| `JOD` | Jordanian Dinar | د.ا |
| `CAD` | Canadian Dollar | C$ |
| `CHF` | Swiss Franc | CHF |
| `JPY` | Japanese Yen | ¥ |
| `AUD` | Australian Dollar | A$ |
| `CNY` | Chinese Yuan Renminbi | ¥ |
| `SEK` | Swedish Krona | kr |
| `NOK` | Norwegian Krone | kr |
| `DKK` | Danish Krone | kr. |

---

### 7.3 Benchmark Indices (15 Indices)
Use any of these codes in `?index={CODE}` or `/v1/indices/{CODE}/stocks`:

| Index Code | Country | Exchange | Full Name (EN) | Full Name (AR) | Tracked Constituents |
|:---|:---:|:---:|:---|:---|:---:|
| `EGX30` | 🇪🇬 Egypt | EGX | EGX 30 Blue-Chip Index | مؤشر إي جي إكس 30 | 30 |
| `EGX70` | 🇪🇬 Egypt | EGX | EGX 70 EWI Small/Mid-Cap Index | مؤشر إي جي إكس 70 متساوي الأوزان | 70 |
| `EGX100` | 🇪🇬 Egypt | EGX | EGX 100 EWI Broad Market Index | مؤشر إي جي إكس 100 متساوي الأوزان | 100 |
| `EGX33` | 🇪🇬 Egypt | EGX | EGX 33 Sharia Compliant Index | مؤشر إي جي إكس 33 للشريعة | 33 |
| `TASI` | 🇸🇦 Saudi Arabia | Tadawul | Tadawul All Share Index | مؤشر السوق الرئيسية تاسي | 271 |
| `MT30` | 🇸🇦 Saudi Arabia | Tadawul | MSCI Tadawul 30 Index | مؤشر إم إس سي آي تداول 30 | 30 |
| `TASI50` | 🇸🇦 Saudi Arabia | Tadawul | TASI 50 Index | مؤشر تاسي 50 | 50 |
| `TADAWUL_LARGE_CAP` | 🇸🇦 Saudi Arabia | Tadawul | Tadawul Large Cap Index | مؤشر تداول للشركات الكبيرة | 40 |
| `TADAWUL_MID_CAP` | 🇸🇦 Saudi Arabia | Tadawul | Tadawul Mid Cap Index | مؤشر تداول للشركات المتوسطة | 60 |
| `TADAWUL_SMALL_CAP` | 🇸🇦 Saudi Arabia | Tadawul | Tadawul Small Cap Index | مؤشر تداول للشركات الصغيرة | 171 |
| `TASI_SHARIAH` | 🇸🇦 Saudi Arabia | Tadawul | TASI Shariah Index | مؤشر تاسي الإسلامي للشريعة | 243 |
| `DFMGI` | 🇦🇪 UAE | DFM | DFM General Index | المؤشر العام لسوق دبي المالي | 13 |
| `FADGI` | 🇦🇪 UAE | ADX | FTSE ADX General Index | مؤشر فوتسي سوق أبوظبي العام | 15 |
| `FADX15` | 🇦🇪 UAE | ADX | FTSE ADX 15 Index | مؤشر فوتسي سوق أبوظبي 15 | 15 |
| `QSE_INDEX` | 🇶🇦 Qatar | QSE | Qatar Exchange General Index | مؤشر بورصة قطر العام | 20 |

---

### 7.4 EGX Thematic Groups
Use any of these values in `GET /v1/stocks?group={GROUP}`:

| Group Name | Description | Key Companies Included |
|:---|:---|:---|
| `EGX30` | 30 largest and most liquid blue chips | COMI, EAST, TMGH, HRHO, FWRY, SWDY, EKHO |
| `EGX70` | 70 medium and small cap growth stocks | ISPH, CIEB, ORAS, ESRS, DSCW |
| `EGX100` | Complete 100 composite index companies | All EGX30 + EGX70 constituents |
| `Sharia` | Sharia-compliant companies (Islamic finance) | ABUK, MFPC, ESRS, ALUM, HELI, ORAS |
| `Dividend` | High dividend-paying equities | COMI, EAST, ABUK, AMOC, SKPC |
| `Technology` | Tech, payments, and telecom firms | FWRY, ETEL, EFIH, RAYA |
