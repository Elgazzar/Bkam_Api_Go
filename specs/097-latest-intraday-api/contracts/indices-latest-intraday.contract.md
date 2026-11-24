# API Contract: Latest Index Intraday Timeline

**Route**: `GET /v1/indices/:symbol/intraday/latest`  
**Authentication**: Required (`x-api-key` header)  
**Description**: Returns the latest trading session's intraday timeline points and latest quote metrics for a recognized benchmark index.

---

## Request

### Path Parameters
| Parameter | Type | Required | Description | Example |
| :--- | :--- | :--- | :--- | :--- |
| `symbol` | `string` | Yes | Benchmark index code (case-insensitive) | `EGX30`, `TASI`, `DFMGI` |

### Query Parameters
| Parameter | Type | Required | Default | Description | Example |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `market` | `string` | No | `egx` | Regional market exchange identifier (`egx`, `saudi`, `uae`, `qatar`) | `?market=saudi` |

---

## Response

### Headers
- `Content-Type: application/json`
- `Cache-Control: public, max-age=60` (active today)
- `Cache-Control: public, max-age=900` (concluded/closed)

### Success Response (`200 OK`)
```json
{
  "symbol": "EGX30",
  "date": "2026-10-22",
  "currency": "EGP",
  "price": 30540.2,
  "change": 185.4,
  "changePercent": 0.61,
  "updatedAt": "2026-10-22T14:30:00.000Z",
  "intraday": [
    {
      "time": "2026-10-22T10:05:00.000Z",
      "price": 30360.0
    },
    {
      "time": "2026-10-22T10:10:00.000Z",
      "price": 30385.5
    }
  ]
}
```

### Error Responses
- `400 Bad Request` (`INVALID_VALIDATION`): Invalid index code.
- `401 Unauthorized` (`UNAUTHORIZED`): Missing or invalid `x-api-key`.
- `404 Not Found` (`NOT_FOUND`): Index not recognized.
