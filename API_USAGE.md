# 🔌 Bekam Fe El-Bank Scrapper — Complete API Reference (v2.5.0)

Welcome to the official API integration reference manual for the Bekam Fe El-Bank platform. The platform serves Egypt's market indices including real-time exchange rates across 28 domestic banks, daily historical exchange archives, gold commodities, regulated domestic gas prices, and systemic availability metadata.

---

## 🔑 Authentication

All endpoints—except the public health checker—require a valid API key configured in the system. Authentication can be supplied using either of the following two mechanisms:

### 1. HTTP Request Header (Recommended)
Include the `x-api-key` header with your verified security token:
```bash
x-api-key: local-dev-key
```

### 2. URL Query Parameter
Alternatively, append the key to the query string using the `apiKey` or `apikey` parameter:
```text
?apikey=local-dev-key
```

---

## 🗺️ API Endpoint Blueprint

| Route | Method | Authentication | Caching Policy | Purpose |
|:---|:---:|:---:|:---:|:---|
| `/v1/rates` | `GET` | **Required** | `1-Minute RAM TTL` | Fetch all current exchange rates (filters supported) |
| `/v1/rates/:currency` | `GET` | **Required** | `1-Minute RAM TTL` | Retrieve rates for a single currency code |
| `/v1/rates/history/:date` | `GET` | **Required** | `1-Hour RAM TTL` | Query Cairo banking close historical exchange rates |
| `/v1/rates/history/range` | `GET` | **Required** | `1-Hour RAM TTL` | Query historical exchange rate range over calendar dates |
| `/v1/gold` | `GET` | **Required** | `1-Minute RAM TTL` | Fetch current domestic Egyptian Gold prices |
| `/v1/gas` | `GET` | **Required** | `10-Minute RAM TTL` | Fetch Egypt regulated fuel and petroleum prices |
| `/v1/stocks` | `GET` | **Required** | `Real-Time Cache` | Fetch EGX stock prices (`?country=saudi|uae|qatar` supported) |
| `/v1/stocks/saudi` | `GET` | **Required** | `1-Minute RAM TTL` | Fetch Saudi Tadawul stock quotes in SAR |
| `/v1/stocks/uae` | `GET` | **Required** | `1-Minute RAM TTL` | Fetch UAE DFM & ADX stock quotes in AED |
| `/v1/stocks/qatar` | `GET` | **Required** | `1-Minute RAM TTL` | Fetch Qatar QSE stock quotes in QAR |
| `/v1/stocks/regional/catalog` | `GET` | **Required** | *Static* | Discover regional GCC stock markets & constituent lists |
| `/v1/stocks/:market/history/:date` | `GET` | **Required** | `1-Hour RAM TTL` | Fetch daily closing snapshot for regional market |
| `/v1/stocks/history/:date` | `GET` | **Required** | `1-Hour RAM TTL` | Fetch daily closing snapshot for EGX30 stock prices |
| `/v1/stocks/:symbol/history/range` | `GET` | **Required** | `1-Hour RAM TTL` | Fetch historical price range for an individual stock symbol (ISO week batched) |
| `/v1/indices` | `GET` | **Required** | `1-Minute RAM TTL` | Fetch regional benchmark indices (`?market=egx\|saudi\|uae\|qatar`) |
| `/v1/indices/:code/stocks` | `GET` | **Required** | `1-Minute RAM TTL` | Fetch constituent equities for an index code with quotes |
| `/v1/indices/saudi` | `GET` | **Required** | `1-Minute RAM TTL` | Shorthand alias for Saudi Tadawul benchmark indices |
| `/v1/indices/uae` | `GET` | **Required** | `1-Minute RAM TTL` | Shorthand alias for UAE DFM/ADX benchmark indices |
| `/v1/indices/qatar` | `GET` | **Required** | `1-Minute RAM TTL` | Shorthand alias for Qatar QSE benchmark indices |
| `/v1/catalog` | `GET` | **Required** | *Static* | Retrieve registered bank catalog metadata and codes |
| `/v1/health` | `GET` | *Public* | *None* | Get operational and health checks for scrapers |

---

## 💰 1. Current Exchange Rates (`GET /v1/rates`)

Retrieves the latest exchange rates retrieved from live bank scraping workflows.

### Query Parameters
- `bank` (optional): Filter results to a specific bank code (e.g. `CIB`, `NBE`, `MISR`). Validated against catalog.
- `currency` (optional): Filter results to a specific three-letter currency code (e.g. `USD`, `EUR`, `SAR`). Validated against catalog.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/rates?bank=CIB&currency=USD"
```

### Example Response Shape (200 OK)
```json
{
  "requestedAt": "2026-05-20T17:15:00.000Z",
  "freshness": "current",
  "requestFilters": {
    "bank": "CIB",
    "currency": "USD"
  },
  "rates": [
    {
      "name_ar": "البنك التجاري الدولي",
      "name_en": "Commercial International Bank",
      "logo": "https://bekamfeelbank.web.app/images/banks/CIB.png",
      "bankCode": "CIB",
      "currencyCode": "USD",
      "buyRate": 47.15,
      "sellRate": 47.25,
      "oldBuyRate": 47.10,
      "oldSellRate": 47.20,
      "retrievedAt": "2026-05-20T17:00:00.000Z",
      "bankUpdatedAt": "2026-05-20T16:45:00.000Z",
      "freshness": "current",
      "sourceTraceId": "tr-abc123xyz",
      "dailyHighBuy": 47.20,
      "dailyHighBuyAt": "2026-05-20T09:30:00.000Z",
      "dailyHighSell": 47.30,
      "dailyHighSellAt": "2026-05-20T10:15:00.000Z"
    }
  ],
  "availability": [
    {
      "bankCode": "CIB",
      "status": "SUCCESS",
      "lastAttemptAt": "2026-05-20T17:00:00.000Z",
      "errorMessage": null
    }
  ],
  "warnings": []
}
```

---

## 💱 2. Exchange Rates by Path Currency (`GET /v1/rates/:currency`)

Shorthand path parameter REST endpoint to query a single currency across all active banks.

### Path Parameters
- `currency` (required): The 3-character ISO currency code (e.g. `usd`, `eur`, `gbp`). Case-insensitive.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/rates/eur"
```

### Example Response Shape (200 OK)
```json
{
  "requestedAt": "2026-05-20T17:16:00.000Z",
  "freshness": "current",
  "requestFilters": {
    "currency": "EUR"
  },
  "rates": [
    {
      "name_ar": "البنك التجاري الدولي",
      "name_en": "Commercial International Bank",
      "logo": "https://bekamfeelbank.web.app/images/banks/CIB.png",
      "bankCode": "CIB",
      "currencyCode": "EUR",
      "buyRate": 50.85,
      "sellRate": 51.10,
      "oldBuyRate": null,
      "oldSellRate": null,
      "retrievedAt": "2026-05-20T17:00:00.000Z",
      "bankUpdatedAt": null,
      "freshness": "current"
    }
  ],
  "availability": [],
  "warnings": []
}
```

---

## 📅 3. Daily Historical Rates (`GET /v1/rates/history/:date`)

Queries compiled daily closing exchange snapshots. To minimize cost, calls are cached in RAM for exactly **1 hour** on the server.

### Path Parameters
- `date` (required): Target history date formatted strictly as `YYYY-MM-DD`. Cannot be in the future.

### Query Parameters
- `bank` (optional): Filter the response snapshot to a single bank (e.g., `CIB`). Case-insensitive.
- `currency` (optional): Filter the response snapshot to a single currency (e.g., `USD`). Case-insensitive.

### Example Requests
```bash
# Get complete daily close snapshot
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/rates/history/2026-05-20"

# Filter daily close snapshot to a specific bank and currency
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/rates/history/2026-05-20?bank=CIB&currency=USD"
```

### Example Response Shape (200 OK)
```json
{
  "date": "2026-05-20",
  "archivedAt": "2026-05-20T17:00:00.000Z",
  "rates": {
    "CIB": {
      "name_ar": "البنك التجاري الدولي",
      "name_en": "Commercial International Bank",
      "logo": "https://bekamfeelbank.web.app/images/banks/CIB.png",
      "bankId": "CIB",
      "status": "SUCCESS",
      "timestamp": "2026-05-20T17:00:00.000Z",
      "rates": {
        "USD": { "buy": 47.15, "sell": 47.25 }
      }
    }
  }
}
```

---

## 📅 4. Historical Rates for Date Range (`GET /v1/rates/history/range`)

Serves a chronological array of filtered historical rates snapshots for a specific date range. Highly optimized for mobile charting libraries.

### Query Parameters
- `start` (required): Starting calendar date formatted strictly as `YYYY-MM-DD`. Cannot be in the future.
- `end` (required): Ending calendar date formatted strictly as `YYYY-MM-DD`. Cannot be in the future. Must be after or equal to `start`.
- `bank` (optional): Filter the range snapshots to a single bank (e.g. `CIB`). Case-insensitive.
- `currency` (optional): Filter the range snapshots to a single currency (e.g. `USD`). Case-insensitive.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/rates/history/range?start=2026-05-18&end=2026-05-19&bank=CIB&currency=USD"
```

### Example Response Shape (200 OK)
```json
[
  {
    "date": "2026-05-18",
    "archivedAt": "2026-05-18T17:00:00.000Z",
    "rates": {
      "CIB": {
        "name_ar": "البنك التجاري الدولي",
        "name_en": "Commercial International Bank",
        "logo": "https://bekamfeelbank.web.app/images/banks/CIB.png",
        "bankId": "CIB",
        "status": "SUCCESS",
        "timestamp": "2026-05-18T17:00:00.000Z",
        "rates": {
          "USD": { "buy": 47.10, "sell": 47.20 }
        }
      }
    }
  },
  {
    "date": "2026-05-19",
    "archivedAt": "2026-05-19T17:00:00.000Z",
    "rates": {
      "CIB": {
        "name_ar": "البنك التجاري الدولي",
        "name_en": "Commercial International Bank",
        "logo": "https://bekamfeelbank.web.app/images/banks/CIB.png",
        "bankId": "CIB",
        "status": "SUCCESS",
        "timestamp": "2026-05-19T17:00:00.000Z",
        "rates": {
          "USD": { "buy": 47.15, "sell": 47.25 }
        }
      }
    }
  }
]
```

---

## 🏆 5. Egyptian Gold Prices (`GET /v1/gold`)

Exposes real-time Egyptian market prices for standard gold commodities (Karats 24, 21, 18, and Gold Coin/Pound). Serves raw values in EGP.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/gold"
```

### Example Response Shape (200 OK)
```json
{
  "retrievedAt": "2026-05-20T17:15:30.000Z",
  "prices": {
    "karat_24": 3650.5,
    "karat_21": 3194.25,
    "karat_18": 2738.0,
    "gold_coin": 25554.0
  },
  "history": {
    "lastUpdated": "2026-05-20T17:00:00.000Z"
  }
}
```

---

## ⛽ 6. Domestic Fuel & Gas Prices (`GET /v1/gas`)

Serves Cairo regulated energy fuel commodities extracted from the Ministry of Petroleum, normalized to standard `"EGP/Liter"` and `"EGP/Cylinder"` indices.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/gas"
```

### Example Response Shape (200 OK)
```json
{
  "retrievedAt": "2026-05-20T17:15:30.000Z",
  "prices": {
    "benzene_95": { "price": 13.5, "unit": "EGP/Liter" },
    "benzene_92": { "price": 12.5, "unit": "EGP/Liter" },
    "benzene_80": { "price": 10.0, "unit": "EGP/Liter" },
    "solar": { "price": 8.25, "unit": "EGP/Liter" },
    "kerosene": { "price": 8.25, "unit": "EGP/Liter" },
    "lpg_cylinder": { "price": 75.0, "unit": "EGP/Cylinder" }
  },
  "history": {
    "lastUpdated": "2026-05-20T17:00:00.000Z"
  }
}
```

---

## 📈 7. EGX Stock Prices (`GET /v1/stocks`)

Retrieves real-time market prices for Egyptian Exchange equities scraped from the TradingView Scanner API. Prices include directional trend indicators powered by a real-time push cache.

### Query Parameters
- `index` (optional): Filter stocks by canonical benchmark index code (e.g. `EGX30`, `EGX70`, `EGX100`, `EGX33`). Case-insensitive. Strictly takes precedence over `group`.
- `group` (optional): Legacy taxonomy filter (e.g. `EGX30`, `Dividend`, `Technology`). Supported for 100% backward compatibility.
- `country` (optional): Regional routing alias (`saudi`, `uae`, `qatar`).

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/stocks?index=EGX30"
```

### Example Response Shape (200 OK)
```json
{
  "source": "https://www.tradingview.com/markets/egypt/market-quotes/",
  "currency": "EGP",
  "updatedAt": "2026-05-31T14:30:00.000Z",
  "scrapedAt": "2026-05-31T14:32:00.000Z",
  "stocks": {
    "COMI": {
      "symbol": "COMI.CA",
      "code": "COMI",
      "nameAr": "البنك التجاري الدولي",
      "nameEn": "Commercial International Bank",
      "price": 75.50,
      "change": 1.20,
      "changePercent": 1.61,
      "previousClose": 74.30,
      "dayHigh": 76.00,
      "dayLow": 74.10,
      "volume": 1500000,
      "indices": ["EGX30", "EGX100"],
      "groups": ["EGX30", "EGX100"],
      "logo": "https://bekamfeelbank.web.app/images/stocks/COMI.png",
      "oldPrice": 74.30,
      "trend": "up"
    }
  },
  "status": "SUCCESS"
}
```

---

## 📅 8. Daily Historical Stock Prices (`GET /v1/stocks/history/:date`)

Queries archived daily closing stock snapshots. Cached in RAM for 1 hour.

### Path Parameters
- `date` (required): Target history date formatted strictly as `YYYY-MM-DD`. Cannot be in the future.

### Query Parameters
- `index` (optional): Filter constituents by benchmark index code (`EGX30`, `EGX70`, `EGX100`, `EGX33`). Takes precedence over `group`.
- `group` (optional): Legacy filter parameter.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/stocks/history/2026-05-31?index=EGX30"
```

---

## 📈 8.1 Single-Stock Historical Price Range (`GET /v1/stocks/:symbol/history/range`)

Queries chronological historical closing quotes and intraday shifts for an individual stock symbol across a calendar date range. Internally utilizes ISO week batching against consolidated weekly buckets (`stock_prices_history_weekly/{YYYY-Www}`) to resolve multi-month ranges in under 1.5s cold / 300ms warm with zero redundant reads on non-trading days.

### Path Parameters
- `symbol` (required): Target stock symbol or ticker code (e.g. `COMI`, `FWRY`, `ETEL`). Case-insensitive.

### Query Parameters
- `start` (required): Start date formatted strictly as `YYYY-MM-DD`. Cannot be in the future.
- `end` (required): End date formatted strictly as `YYYY-MM-DD`. Must be greater than or equal to `start`.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/stocks/COMI/history/range?start=2026-05-01&end=2026-05-31"
```

### Example Response Shape (200 OK)
```json
{
  "symbol": "COMI",
  "startDate": "2026-05-01",
  "endDate": "2026-05-31",
  "history": [
    {
      "date": "2026-05-03",
      "symbol": "COMI",
      "price": 75.50,
      "change": 1.20,
      "changePercent": 1.61,
      "updatedAt": "2026-05-03T14:30:00.000Z",
      "intraday": [
        { "time": "2026-05-03T10:05:00Z", "price": 74.80 },
        { "time": "2026-05-03T14:25:00Z", "price": 75.50 }
      ]
    }
  ]
}
```

---

## 🌍 9. Regional GCC Stock Markets (Saudi, UAE, Qatar)

Access live quotes, constituents, and daily history for Saudi Arabia (Tadawul), UAE (DFM and ADX), and Qatar (QSE) equity markets.

### Supported Endpoints

- **Saudi Arabia (Tadawul)**: `GET /v1/stocks/saudi` (or `GET /v1/stocks?country=saudi`)
  - Currency: `SAR` | Exchange: `TADAWUL`
  - Query Filters: `?index=TASI|MT30` (takes precedence), `?group=TASI|Energy|Banking`
- **UAE (DFM & ADX)**: `GET /v1/stocks/uae` (or `GET /v1/stocks?country=uae`)
  - Currency: `AED` | Exchanges: `DFM`, `ADX`
  - Query Filters: `?index=DFMGI|FADGI|FADX15` (takes precedence), `?exchange=dfm|adx`, `?group=DFMGI|ADXGI`
- **Qatar (QSE)**: `GET /v1/stocks/qatar` (or `GET /v1/stocks?country=qatar`)
  - Currency: `QAR` | Exchange: `QSE`
  - Query Filters: `?index=QSE_INDEX` (takes precedence), `?group=QSE_INDEX|Banking`
- **Catalog Discovery**: `GET /v1/stocks/regional/catalog`
  - Returns supported regional markets, currencies, operating hours, and constituent inventories.
- **Regional Daily History**: `GET /v1/stocks/:market/history/:date`
  - Parameters: `:market` (`saudi`, `uae`, `qatar`), `:date` (`YYYY-MM-DD`).
  - Query Filters: `?index=...`, `?group=...`
- **Single-Stock History Range**: `GET /v1/stocks/:market/:symbol/history/range`
  - Parameters: `:market` (`saudi`, `uae`, `qatar`), `:symbol` (`2222`, `EMAAR`, etc.).
  - Query Parameters: `start=YYYY-MM-DD&end=YYYY-MM-DD`. Cached in RAM for 1 hour.

### Example Request (Saudi Arabia)
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/stocks/saudi?index=TASI"
```

### Example Response Shape (200 OK)
```json
{
  "country": "SA",
  "exchange": "TADAWUL",
  "source": "https://www.tradingview.com/markets/saudi-arabia/market-quotes/",
  "currency": "SAR",
  "updatedAt": "2026-10-20T12:00:00.000Z",
  "scrapedAt": "2026-10-20T12:00:05.120Z",
  "stocks": {
    "ARAMCO": {
      "symbol": "2222",
      "code": "ARAMCO",
      "exchange": "TADAWUL",
      "nameAr": "أرامكو السعودية",
      "nameEn": "Saudi Aramco",
      "price": 27.50,
      "change": 0.35,
      "changePercent": 1.28,
      "previousClose": 27.15,
      "dayHigh": 27.80,
      "dayLow": 27.10,
      "volume": 12500000,
      "currency": "SAR",
      "indices": ["TASI"],
      "groups": ["TASI", "Energy"],
      "logo": "https://bekamfeelbank.web.app/images/stocks/saudi/ARAMCO.png",
      "oldPrice": 27.15,
      "trend": "up"
    }
  }
}
```

---

## 📊 10. Benchmark Indices & Constituents (`GET /v1/indices`)

Access live quotes for canonical benchmark financial indices across Egypt and the GCC, or reverse-lookup constituent equities by index code.

### Canonical Benchmark Indices

| Code | Market | Exchange | Currency | Description |
|:---|:---:|:---:|:---:|:---|
| `EGX30` | Egypt | EGX | EGP | Top 30 most liquid Egyptian equities |
| `EGX70` | Egypt | EGX | EGP | Mid & small cap benchmark |
| `EGX100` | Egypt | EGX | EGP | Broad composite index (EGX30 + EGX70) |
| `EGX33` | Egypt | EGX | EGP | Sharia-compliant benchmark index |
| `TASI` | Saudi Arabia | Tadawul | SAR | Tadawul All Share broad composite index |
| `MT30` | Saudi Arabia | Tadawul | SAR | MSCI Tadawul 30 index |
| `DFMGI` | UAE | DFM | AED | Dubai Financial Market General Index |
| `FADGI` | UAE | ADX | AED | FTSE ADX General Index |
| `FADX15` | UAE | ADX | AED | FTSE ADX 15 blue-chip index |
| `QSE_INDEX` | Qatar | QSE | QAR | Qatar Stock Exchange 20 benchmark |

### Multi-Market Index Quotes (`GET /v1/indices`)
- **Query Parameters**:
  - `market` (optional): `egx` (default), `saudi`, `uae`, or `qatar`. Omitting returns Egyptian indices for 100% backward compatibility.
- **Regional Alias Endpoints**:
  - `GET /v1/indices/saudi` (equivalent to `?market=saudi`)
  - `GET /v1/indices/uae` (equivalent to `?market=uae`)
  - `GET /v1/indices/qatar` (equivalent to `?market=qatar`)

### Example Request (Regional Benchmark Indices)
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/indices?market=saudi"
```

### Example Response Shape (200 OK)
```json
{
  "source": "https://www.tradingview.com/markets/saudi-arabia/market-quotes/",
  "currency": "SAR",
  "market": "saudi",
  "updatedAt": "2026-10-22T10:30:00.000Z",
  "indices": {
    "TASI": {
      "symbol": "TASI",
      "code": "TASI",
      "nameAr": "مؤشر تاسي",
      "nameEn": "Tadawul All Share Index",
      "price": 12150.40,
      "change": 45.20,
      "changePercent": 0.37,
      "previousClose": 12105.20,
      "dayHigh": 12180.00,
      "dayLow": 12090.50,
      "currency": "SAR",
      "logo": "https://bekamfeelbank.web.app/images/indices/TASI.png",
      "trend": "up"
    }
  },
  "status": "SUCCESS"
}
```

### Index Constituent Lookup (`GET /v1/indices/:code/stocks`)
Resolves all constituent equities belonging to a canonical index code enriched with current price quotes, percentage change, and logo URLs. Falls back dynamically to catalog metadata when markets are closed.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/indices/EGX30/stocks"
```

### Example Response Shape (200 OK)
```json
{
  "indexCode": "EGX30",
  "market": "egx",
  "currency": "EGP",
  "count": 39,
  "updatedAt": "2026-10-22T14:30:00.000Z",
  "stocks": [
    {
      "symbol": "COMI.CA",
      "code": "COMI",
      "nameAr": "البنك التجاري الدولي",
      "nameEn": "Commercial International Bank",
      "price": 75.50,
      "change": 1.20,
      "changePercent": 1.61,
      "previousClose": 74.30,
      "dayHigh": 76.00,
      "dayLow": 74.10,
      "volume": 1500000,
      "indices": ["EGX30", "EGX100"],
      "groups": ["EGX30", "EGX100"],
      "logo": "https://bekamfeelbank.web.app/images/stocks/COMI.png",
      "trend": "up"
    }
  ]
}
```

---

## 📁 11. Metadata Bank Catalog (`GET /v1/catalog`)

Fetches catalog descriptors and parameters listing all 28 supported banks. Perfect for dynamically generating filter buttons or lookup indices on frontends.

### Example Request
```bash
curl -H "x-api-key: local-dev-key" "http://localhost:3001/v1/catalog"
```

### Example Response Shape (200 OK)
```json
{
  "banks": [
    {
      "code": "CIB",
      "name": "Commercial International Bank",
      "name_ar": "البنك التجاري الدولي",
      "name_en": "Commercial International Bank",
      "logo": "https://bekamfeelbank.web.app/images/banks/CIB.png",
      "sourceUrl": "https://www.cibeg.com/en/personal/exchange-rates",
      "enabled": true,
      "availability": "current"
    },
    {
      "code": "NXT",
      "name": "NXT Bank",
      "name_ar": "بنك نكست",
      "name_en": "NXT Bank",
      "logo": "https://bekamfeelbank.web.app/images/banks/NXT.png",
      "sourceUrl": "https://www.nxt.com.eg/en/rates/",
      "enabled": true,
      "availability": "current"
    }
  ],
  "currencies": [
    { "code": "USD", "name": "US Dollar" },
    { "code": "EUR", "name": "Euro" }
  ]
}
```

---

## 🏥 12. Health Monitor Check (`GET /v1/health`)

Public health checker used for status evaluation. Requires zero authentication keys.

### Example Request
```bash
curl "http://localhost:3001/v1/health"
```

### Example Response Shape (200 OK)
```json
{
  "status": "healthy",
  "uptimeSeconds": 15482.4,
  "timestamp": "2026-05-20T17:15:30.000Z",
  "cacheSize": 1,
  "system": {
    "memoryUsageMB": 42.15,
    "cpuPercentage": 0.8
  }
}
```

---

## 🚫 Error Responses

When standard validation schema parameters fail, or an resource is absent, the API resolves standard HTTP status errors wrapped in a structured error envelope:

### Example: Unsupported Bank/Currency Filter (404 / 400)
```json
{
  "statusCode": 404,
  "error": {
    "code": "BANK_NOT_FOUND",
    "message": "Unsupported bank filter: FAKE_BANK."
  }
}
```

### Example: Missing or Invalid API Key (401 Unauthorized)
```json
{
  "statusCode": 401,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid API key."
  }
}
```
