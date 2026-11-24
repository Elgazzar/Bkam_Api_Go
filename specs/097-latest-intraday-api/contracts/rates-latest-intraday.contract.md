# API Contract: Latest Bank Rate Intraday Timeline

**Route**: `GET /v1/rates/:currency/intraday/latest`  
**Authentication**: Required (`x-api-key` header)  
**Description**: Returns the latest trading session's intraday timeline points for a specific foreign currency (e.g. USD, EUR, SAR).

---

## Request

### Path Parameters
| Parameter | Type | Required | Description | Example |
| :--- | :--- | :--- | :--- | :--- |
| `currency` | `string` | Yes | 3-letter ISO currency code | `USD`, `EUR`, `SAR` |

### Query Parameters
| Parameter | Type | Required | Default | Description | Example |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `bank` | `string` | No | Consolidated | Optional bank short code to filter points to a single financial institution | `?bank=cib` |

---

## Response

### Headers
- `Content-Type: application/json`
- `Cache-Control: public, max-age=60` (active today)
- `Cache-Control: public, max-age=900` (concluded/closed)

### Filtered by Bank Success Response (`200 OK`)
```json
{
  "currency": "USD",
  "bank": "cib",
  "date": "2026-10-24",
  "buy": 48.65,
  "sell": 48.75,
  "updatedAt": "2026-10-24T14:00:00.000Z",
  "intraday": [
    {
      "time": "2026-10-24T09:30:00.000Z",
      "price": 48.60
    },
    {
      "time": "2026-10-24T11:15:00.000Z",
      "price": 48.65
    }
  ]
}
```

### Consolidated Currency Success Response (`200 OK`)
```json
{
  "currency": "USD",
  "date": "2026-10-24",
  "buy": 48.62,
  "sell": 48.72,
  "updatedAt": "2026-10-24T14:00:00.000Z",
  "intraday": [
    {
      "time": "2026-10-24T09:30:00.000Z",
      "price": 48.60
    },
    {
      "time": "2026-10-24T11:15:00.000Z",
      "price": 48.62
    }
  ]
}
```

### Error Responses
- `400 Bad Request` (`INVALID_FILTER`): Invalid currency code or unrecognized bank identifier.
- `401 Unauthorized` (`UNAUTHORIZED`): Missing or invalid `x-api-key`.
- `404 Not Found` (`NOT_FOUND`): Currency not supported.
