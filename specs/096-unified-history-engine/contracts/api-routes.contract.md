# Historical API Routes Contract

All endpoints require verified API key authorization via preHandler `verifyApiKey`.

---

## 1. Bank Rates History

### `GET /v1/rates/history/:date`
- **Params**: `date` (format: `YYYY-MM-DD`)
- **Query Params**:
  - `bank?: string` (e.g. `cib`)
  - `currency?: string` (e.g. `USD`)
- **Headers**:
  - `Cache-Control: public, max-age=60` (if date is today)
  - `Cache-Control: public, max-age=86400, immutable` (if date is past)
- **Response `200 OK`**:
  ```json
  {
    "date": "2026-10-21",
    "archivedAt": "2026-10-21T15:00:00.000Z",
    "rates": {
      "CIB": {
        "USD": {
          "buy": 48.55,
          "sell": 48.65,
          "trend": "up"
        }
      }
    }
  }
  ```
- **Response `404 Not Found`**:
  ```json
  {
    "statusCode": 404,
    "error": "NOT_FOUND",
    "message": "No historical rates found for date: 2026-10-21"
  }
  ```

### `GET /v1/rates/history/range`
- **Query Params**:
  - `start: string` (format: `YYYY-MM-DD`)
  - `end: string` (format: `YYYY-MM-DD`)
  - `bank?: string`
  - `currency?: string`
- **Response `200 OK`**: Array of historical rates snapshots, sorted by date ascending.

---

## 2. Gold History

### `GET /v1/gold/history/:date`
- **Params**: `date` (format: `YYYY-MM-DD`)
- **Response `200 OK`**:
  ```json
  {
    "date": "2026-10-21",
    "updatedAt": "2026-10-21T15:00:00.000Z",
    "karats": {
      "24": { "buy": 4200, "sell": 4220 },
      "21": { "buy": 3675, "sell": 3695 },
      "18": { "buy": 3150, "sell": 3170 },
      "pound": { "buy": 29400, "sell": 29560 }
    },
    "international": {
      "priceOunceUsd": 2720.50
    }
  }
  ```

### `GET /v1/gold/history/range`
- **Query Params**:
  - `start: string`
  - `end: string`
- **Response `200 OK`**: Array of daily gold snapshots, sorted by date ascending.

---

## 3. Stocks History

### `GET /v1/stocks/history/:date`
- **Params**: `date`
- **Query Params**: `group?: string`, `index?: string`
- **Response `200 OK`**: Daily stock snapshot payload.

### `GET /v1/stocks/:symbol/history/range`
- **Params**: `symbol` (e.g. `COMI`)
- **Query Params**: `start: string`, `end: string`
- **Response `200 OK`**:
  ```json
  {
    "symbol": "COMI",
    "startDate": "2026-05-01",
    "endDate": "2026-10-21",
    "history": [
      {
        "date": "2026-05-02",
        "price": 85.50,
        "change": 1.20,
        "changePercent": 1.42
      }
    ]
  }
  ```

---

## 4. Indices History

### `GET /v1/indices/history/:date`
- **Params**: `date`
- **Response `200 OK`**: Daily index snapshot payload.

### `GET /v1/indices/:symbol/history/range`
- **Params**: `symbol` (e.g. `EGX30`)
- **Query Params**: `start: string`, `end: string`
- **Response `200 OK`**: Array of historical index points for the requested symbol.

---

## 5. GCC Equities History

### `GET /v1/stocks/:market/history/:date`
- **Params**: `market` (`saudi` | `uae` | `qatar`), `date`

### `GET /v1/stocks/:market/:symbol/history/range`
- **Params**: `market`, `symbol`
- **Query Params**: `start`, `end`
