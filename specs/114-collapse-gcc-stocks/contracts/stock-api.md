# Public API Contract: Unified Stock Endpoints

**Base Path**: `/v1/stocks`
**Authentication**: Required (`x-api-key` header)

---

## 1. Real-Time Snapshot Endpoints

### Primary Endpoint: `GET /v1/stocks`
Retrieves live stock quotes for the requested exchange, defaulting to EGX.

- **Query Parameters**:
  - `market` (optional, default: `egx`): `egx` | `saudi` | `uae` | `qatar` | `kuwait` | `eg` | `sa` | `ae` | `qa` | `kw`
- **Request Headers**:
  - `x-api-key`: string (Required)
  - `If-None-Match`: string (Optional ETag)
- **Response Headers**:
  - `ETag`: W/"..."
  - `Cache-Control`: `public, max-age=60`
  - `x-trace-id`: string
- **Status Codes**:
  - `200 OK`: Valid snapshot returned
  - `304 Not Modified`: ETag matched, payload empty
  - `400 Bad Request`: Unknown market code
  - `401 Unauthorized`: Missing or invalid API key

```json
{
  "date": "2026-11-15",
  "market": "saudi",
  "country": "SA",
  "currency": "SAR",
  "updatedAt": "2026-11-15T15:30:00.000Z",
  "scrapedAt": "2026-11-15T15:30:02.123Z",
  "freshness": "current",
  "sourceTraceId": "push-gcc-stocks-179477...",
  "data": {
    "country": "SA",
    "currency": "SAR",
    "updatedAt": "2026-11-15T15:30:00.000Z",
    "stocks": {
      "1120": {
        "symbol": "1120",
        "nameAr": "مصرف الراجحي",
        "nameEn": "Al Rajhi Bank",
        "price": 88.5,
        "change": 0.5,
        "changePercent": 0.57,
        "trend": "up",
        "volume": 2450000,
        "turnover": 216825000,
        "sector": "Banks",
        "benchmarkIndices": ["TASI"],
        "updatedAt": "2026-11-15T15:30:00.000Z"
      }
    }
  }
}
```

---

### Legacy Aliases (Zero Regression)
These aliases route directly to the unified stock engine:
- `GET /v1/stocks/saudi` $\rightarrow$ Equivalent to `GET /v1/stocks?market=saudi`
- `GET /v1/stocks/uae` $\rightarrow$ Equivalent to `GET /v1/stocks?market=uae`
- `GET /v1/stocks/qatar` $\rightarrow$ Equivalent to `GET /v1/stocks?market=qatar`
- `GET /v1/stocks/kuwait` $\rightarrow$ Equivalent to `GET /v1/stocks?market=kuwait`
- `GET /v1/stocks/regional/catalog` $\rightarrow$ Returns multi-market security inventory

---

## 2. Historical & Range Endpoints

### Single-Date Historical: `GET /v1/stocks/history/:date`
- **Path Parameter**: `date` (`YYYY-MM-DD`)
- **Query Parameter**: `market` (optional, default: `egx`)
- **Legacy Alias**: `GET /v1/stocks/:market/history/:date`
- **Response**: `200 OK` with `StockPriceSnapshot` representing daily closing prices.

### Multi-Week Range: `GET /v1/stocks/:symbol/history/range`
- **Path Parameter**: `symbol` (e.g. `COMI`, `1120`)
- **Query Parameters**:
  - `start`: `YYYY-MM-DD`
  - `end`: `YYYY-MM-DD`
  - `market`: optional market identifier
- **Legacy Alias**: `GET /v1/stocks/:market/:symbol/history/range`
- **Response**: `200 OK` with array of daily closing records:
```json
[
  {
    "date": "2026-11-10",
    "symbol": "1120",
    "price": 87.2,
    "change": -0.3,
    "changePercent": -0.34,
    "updatedAt": "2026-11-10T15:30:00.000Z"
  }
]
```

---

## 3. Dedicated Latest Intraday Timeline

### Latest Session Ticks: `GET /v1/stocks/:symbol/intraday/latest`
- **Path Parameter**: `symbol`
- **Query Parameter**: `market` (optional, default: `egx`)
- **Legacy Alias**: `GET /v1/stocks/:market/:symbol/intraday/latest`
- **Response**: `200 OK` with intraday tick series:
```json
{
  "symbol": "1120",
  "date": "2026-11-15",
  "currency": "SAR",
  "price": 88.5,
  "change": 0.5,
  "changePercent": 0.57,
  "updatedAt": "2026-11-15T15:30:00.000Z",
  "isToday": true,
  "intraday": [
    { "time": "2026-11-15T10:00:00.000Z", "price": 88.0 },
    { "time": "2026-11-15T10:05:00.000Z", "price": 88.2 },
    { "time": "2026-11-15T15:30:00.000Z", "price": 88.5 }
  ]
}
```
