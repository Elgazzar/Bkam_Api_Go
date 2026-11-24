# API Contract: Latest Stock Intraday Timeline

**Route**: `GET /v1/stocks/:symbol/intraday/latest`  
**Authentication**: Required (`x-api-key` header)  
**Description**: Returns the latest trading session's intraday timeline points and latest quote metrics for a specific stock ticker.

---

## Request

### Path Parameters
| Parameter | Type | Required | Description | Example |
| :--- | :--- | :--- | :--- | :--- |
| `symbol` | `string` | Yes | Ticker symbol code (case-insensitive) | `COMI`, `FWRY`, `2222` |

### Query Parameters
| Parameter | Type | Required | Default | Description | Example |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `market` | `string` | No | `egx` | Regional market exchange identifier (`egx`, `saudi`, `uae`, `qatar`, `kuwait`) | `?market=saudi` |

---

## Response

### Headers
- `Content-Type: application/json`
- `Cache-Control: public, max-age=60` (when session is active today)
- `Cache-Control: public, max-age=900` (when session is concluded/closed)

### Success Response (`200 OK`)
```json
{
  "symbol": "COMI",
  "date": "2026-10-22",
  "currency": "EGP",
  "price": 95.5,
  "change": 1.2,
  "changePercent": 1.27,
  "updatedAt": "2026-10-22T14:30:00.000Z",
  "intraday": [
    {
      "time": "2026-10-22T10:05:00.000Z",
      "price": 95.5
    },
    {
      "time": "2026-10-22T10:10:00.000Z",
      "price": 95.8
    }
  ]
}
```

### Inactive Symbol Response (`200 OK`)
```json
{
  "symbol": "DAPH",
  "date": "2026-10-22",
  "currency": "EGP",
  "price": 12.4,
  "change": 0.0,
  "changePercent": 0.0,
  "updatedAt": "2026-10-22T14:30:00.000Z",
  "intraday": []
}
```

### Error Responses
- `400 Bad Request` (`INVALID_VALIDATION`): Invalid symbol or market parameter.
- `401 Unauthorized` (`UNAUTHORIZED`): Missing or invalid `x-api-key`.
- `404 Not Found` (`NOT_FOUND`): Symbol not registered in catalog.
