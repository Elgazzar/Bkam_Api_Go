# API Contract: Latest Gold Intraday Timeline

**Routes**:
- Single item: `GET /v1/gold/:item/intraday/latest`
- All items: `GET /v1/gold/intraday/latest`

**Authentication**: Required (`x-api-key` header)  
**Description**: Returns the latest trading session's intraday timeline points for gold commodities (24k, 21k, 18k, gold pound, international ounce).

---

## Request

### Path Parameters (`/v1/gold/:item/intraday/latest`)
| Parameter | Type | Required | Description | Example |
| :--- | :--- | :--- | :--- | :--- |
| `item` | `string` | Yes | Target gold commodity or karat | `24`, `21`, `18`, `pound`, `international`, `all` |

---

## Response

### Headers
- `Content-Type: application/json`
- `Cache-Control: public, max-age=60` (active today)
- `Cache-Control: public, max-age=900` (concluded/closed)

### Single Karat Success Response (`200 OK`)
```json
{
  "item": "21",
  "date": "2026-10-24",
  "currency": "EGP",
  "buy": 3720.0,
  "sell": 3740.0,
  "updatedAt": "2026-10-24T15:30:00.000Z",
  "intraday": [
    {
      "time": "2026-10-24T10:15:00.000Z",
      "price": 3715.0
    },
    {
      "time": "2026-10-24T12:00:00.000Z",
      "price": 3720.0
    }
  ]
}
```

### All Karats Success Response (`200 OK`)
```json
{
  "date": "2026-10-24",
  "currency": "EGP",
  "updatedAt": "2026-10-24T15:30:00.000Z",
  "karats": {
    "24": {
      "buy": 4251.0,
      "sell": 4274.0,
      "intraday": [
        { "time": "2026-10-24T10:15:00.000Z", "price": 4245.0 }
      ]
    },
    "21": {
      "buy": 3720.0,
      "sell": 3740.0,
      "intraday": [
        { "time": "2026-10-24T10:15:00.000Z", "price": 3715.0 }
      ]
    },
    "18": {
      "buy": 3188.0,
      "sell": 3205.0,
      "intraday": []
    }
  },
  "pound": {
    "buy": 29760.0,
    "sell": 29920.0,
    "intraday": []
  },
  "international": {
    "buy": 2735.5,
    "sell": 2736.0,
    "currency": "USD",
    "intraday": []
  }
}
```
