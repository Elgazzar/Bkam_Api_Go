# HTTP Contract: Stock History Range Endpoint

**Endpoint**: `GET /v1/stocks/:symbol/history/range`  
**Authentication**: Header `x-api-key: <token>`

---

## Request Parameters

### Path Parameters
- `symbol` (string, required): Canonical equity ticker symbol (e.g., `COMI`, `TMGH`, `SWDY`).

### Query Parameters
- `start` (string, required): Start date formatted as `YYYY-MM-DD`.
- `end` (string, required): End date formatted as `YYYY-MM-DD`. Must be greater than or equal to `start`.

### Request Headers
- `x-api-key` (string, required): Valid API authentication key.
- `Accept-Encoding` (string, optional): Client supported compression formats (e.g. `gzip, deflate, br`).
- `If-None-Match` (string, optional): Weak entity tag from previous response (e.g. `W/"123-abc..."`).

---

## Response Headers

- `ETag` (string): Weak entity tag for conditional HTTP revalidation (e.g. `W/"a1b2c3d4..."`).
- `Content-Encoding` (string, conditional): `gzip` or `br` when payload exceeds 1KB and client sends matching `Accept-Encoding`.
- `Cache-Control` (string):
  - Pure historical ranges (`end < todayCairo`): `public, max-age=86400, immutable`
  - Ranges ending on or after today (`end >= todayCairo`): `public, max-age=60`

---

## Response Status Codes

### 200 OK (Data Returned)
```json
{
  "symbol": "COMI",
  "startDate": "2026-05-01",
  "endDate": "2026-10-24",
  "history": [
    {
      "date": "2026-05-03",
      "symbol": "COMI",
      "price": 82.5,
      "change": 1.25,
      "changePercent": 1.54,
      "updatedAt": "2026-05-03T14:30:00.000Z"
    },
    {
      "date": "2026-10-24",
      "symbol": "COMI",
      "price": 88.0,
      "change": 0.5,
      "changePercent": 0.57,
      "updatedAt": "2026-10-24T12:00:00.000Z",
      "intraday": [
        { "time": "10:05", "price": 87.8 },
        { "time": "10:10", "price": 88.0 }
      ]
    }
  ]
}
```

### 304 Not Modified (Conditional Hit)
- **Condition**: Client sends `If-None-Match` header matching the current response ETag.
- **Payload Body**: Empty (`0` bytes).
- **Latency**: `<10ms` roundtrip.

### 400 Bad Request (Validation Error)
```json
{
  "statusCode": 400,
  "error": "INVALID_VALIDATION",
  "message": "Invalid range parameters. Dates must be formatted as YYYY-MM-DD, and start must be before or equal to end.",
  "details": []
}
```
