# REST API Route Contracts: Catalog & Validation

**Feature**: `108-unified-market-catalog` | **Date**: 2026-11-15

---

## 1. Catalog Discovery Endpoint

### `GET /v1/catalog`
Returns the complete discovery inventory of supported banks, currencies, commodities, benchmark indices, and regional market summaries.

#### Response Headers:
- `Content-Type: application/json; charset=utf-8`
- `Cache-Control: public, max-age=86400, stale-while-revalidate=3600`

#### Response Body:
```json
{
  "banks": [
    {
      "code": "NBE",
      "name": "National Bank of Egypt",
      "name_ar": "البنك الأهلي المصري",
      "name_en": "National Bank of Egypt",
      "logo": "https://bekamfeelbank.web.app/images/banks/NBE.png",
      "sourceUrl": "https://www.nbe.com.eg/en/ExchangeRate.aspx",
      "enabled": true,
      "availability": "current"
    }
  ],
  "currencies": [
    {
      "code": "USD",
      "name": "US Dollar",
      "name_ar": "دولار أمريكي",
      "name_en": "US Dollar",
      "symbol": "$",
      "flag": "🇺🇸"
    }
  ],
  "commodities": [
    {
      "code": "24",
      "category": "gold",
      "nameAr": "عيار 24",
      "nameEn": "24 Karat Gold",
      "unit": "gram"
    }
  ],
  "indices": [
    {
      "symbol": "EGX30",
      "code": "EGX30",
      "exchange": "EGX",
      "country": "EG",
      "currency": "EGP",
      "nameAr": "مؤشر إي جي إكس 30",
      "nameEn": "EGX 30 Index",
      "logo": "https://bekamfeelbank.web.app/images/indices/egx30.png",
      "constituentSymbols": ["COMI", "TMGH", "MFPC"]
    }
  ],
  "counts": {
    "totalBanks": 36,
    "totalCurrencies": 28,
    "totalCommodities": 7,
    "totalIndices": 21,
    "totalSecurities": 580
  }
}
```

---

## 2. Benchmark Index Constituents Endpoint

### `GET /v1/indices/:code/constituents`
Returns constituent member equities for a benchmark index.

#### Request Parameters:
- Path parameter `:code` (string, case-insensitive): e.g. `EGX30`, `egx70`, `TASI`

#### Success Response (`200 OK`):
```json
{
  "indexCode": "EGX30",
  "nameAr": "مؤشر إي جي إكس 30",
  "nameEn": "EGX 30 Index",
  "exchange": "EGX",
  "currency": "EGP",
  "count": 30,
  "updatedAt": "2026-11-15T12:00:00.000Z",
  "constituents": {
    "COMI": {
      "symbol": "COMI",
      "code": "COMI",
      "nameAr": "البنك التجاري الدولي",
      "nameEn": "Commercial International Bank",
      "sector": "Banking",
      "price": 95.5,
      "change": 1.2,
      "changePercent": 1.27,
      "logo": "https://bekamfeelbank.web.app/images/stocks/COMI.png"
    }
  }
}
```

#### Error Response (`400 Bad Request`):
When index code does not exist in catalog:
```json
{
  "code": "INVALID_VALIDATION",
  "error": "Bad Request",
  "message": "Invalid index code: UNKNOWN_INDEX. Must be one of: EGX30, EGX70, EGX100, TASI, ADX15, DFMGI, QE20, ..."
}
```

---

## 3. Stock Screening Group Validation Contract

### `GET /v1/stocks?group=:group`
Filters domestic equities by recognized benchmark group.

#### Valid Values:
`EGX30`, `EGX70`, `EGX100`, `Sharia`, `Dividend`, `Technology`

#### Error Response (`400 Bad Request`):
When group is unrecognized:
```json
{
  "code": "INVALID_VALIDATION",
  "error": "Bad Request",
  "message": "Invalid stock group: InvalidGroup. Must be one of: EGX30, EGX70, EGX100, Sharia, Dividend, Technology"
}
```
