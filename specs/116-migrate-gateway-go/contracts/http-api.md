# Inbound HTTP API Contracts: Migrate Bkam Financial API Gateway to Idiomatic Go

**Feature**: `116-migrate-gateway-go`  
**Date**: 2026-11-16  
**Status**: Completed  

---

## 1. Public Endpoint Surface (49 Routes)

All routes must strictly maintain the existing URL schemes, query parameters, status codes, and JSON response envelopes.

### 1.1 Authentication & System Discovery

| Method | Path | Auth Required | Description | Status Code |
|---|---|:---:|---|:---:|
| `GET` | `/v1/health` | No | Liveness probe returning operational state and timestamp | `200 OK` |
| `GET` | `/v1/version` | No | Semantic version (`2.18.20`), runtime environment, uptime | `200 OK` |
| `GET` | `/version` | No | Public discovery alias for `/v1/version` | `200 OK` |
| `GET` | `/v1/metrics` | Yes | Request counts, route distributions, p95 latencies | `200 OK` |
| `GET` | `/v1/catalog` | Yes | Catalog of supported banks, currencies, localized metadata | `200 OK` |

---

### 1.2 Currency Exchange Rates

| Method | Path | Description | Status Code |
|---|---|---|:---:|
| `GET` | `/v1/rates` | Full bank currency rate snapshot | `200 OK` |
| `GET` | `/v1/rates?currency={code}` | Filtered rates query by currency | `200 OK` |
| `GET` | `/v1/rates/:currency` | Path-based filtered currency rates | `200 OK` |
| `GET` | `/v1/banks/:bankCode/rates` | Rates offered by a specific bank | `200 OK` |
| `GET` | `/v1/rates/history/:date` | Historical rates for single calendar date (`YYYY-MM-DD`) | `200 OK` |
| `GET` | `/v1/rates/history/range` | Multi-day date range query (`start`, `end`, `currency`, `bank`) | `200 OK` |
| `GET` | `/v1/rates/:currency/intraday/latest` | Intraday price movement timeline points | `200 OK` |

---

### 1.3 Precious Metals & Petroleum

| Method | Path | Description | Status Code |
|---|---|---|:---:|
| `GET` | `/v1/gold` | Latest gold prices across all karats | `200 OK` |
| `GET` | `/v1/gold/intraday/latest` | Intraday timeline points for all gold karats | `200 OK` |
| `GET` | `/v1/gold/:karat/intraday/latest` | Intraday timeline for specific karat (e.g. `24`, `21`, `18`) | `200 OK` |
| `GET` | `/v1/gold/history/:date` | Historical gold prices for specific date | `200 OK` |
| `GET` | `/v1/gold/history/range` | Historical gold range (`start`, `end`, `karat`) | `200 OK` |
| `GET` | `/v1/gas` | Latest petroleum prices (benzene 80/92/95, diesel, CNG) | `200 OK` |

---

### 1.4 Egyptian Equities & Benchmark Indices

| Method | Path | Description | Status Code |
|---|---|---|:---:|
| `GET` | `/v1/stocks` | Full EGX stock quotes snapshot | `200 OK` |
| `GET` | `/v1/stocks/:symbol/intraday/latest` | Intraday ticker timeline (e.g. `COMI`) | `200 OK` |
| `GET` | `/v1/stocks/history/:date` | Historical EGX daily archive | `200 OK` |
| `GET` | `/v1/stocks/:symbol/history/range` | Multi-day historical range for specific stock | `200 OK` |
| `GET` | `/v1/indices` | Latest EGX index benchmarks (EGX30, EGX70, EGX100) | `200 OK` |
| `GET` | `/v1/indices/:indexCode/stocks` | Index constituent equities list | `200 OK` |
| `GET` | `/v1/indices/history/:date` | Historical index benchmark archive | `200 OK` |
| `GET` | `/v1/indices/history/range` | Multi-day historical index range | `200 OK` |

---

### 1.5 Regional GCC Stock Markets

| Method | Path | Market Exchange | Currency | Status Code |
|---|---|---|---|:---:|
| `GET` | `/v1/stocks/saudi` | Saudi Tadawul | `SAR` | `200 OK` |
| `GET` | `/v1/stocks/uae` | UAE (DFM & ADX) | `AED` | `200 OK` |
| `GET` | `/v1/stocks/qatar` | Qatar (QSE) | `QAR` | `200 OK` |
| `GET` | `/v1/stocks/kuwait` | Kuwait (KSE) | `KWD` | `200 OK` |
| `GET` | `/v1/stocks/regional/catalog` | GCC equity catalog | Multi | `200 OK` |
| `GET` | `/v1/stocks/:market/history/:date` | Regional daily archive | Respective | `200 OK` |
| `GET` | `/v1/stocks/:market/:symbol/history/range` | Regional multi-day range | Respective | `200 OK` |
| `GET` | `/v1/indices/:market` | Regional market benchmarks | Respective | `200 OK` |

---

### 1.6 Batch Quotes API

| Method | Path | Description | Status Code |
|---|---|---|:---:|
| `POST` | `/v1/quotes` | Batch resolution from JSON array: `{"items": ["CBE_USD", "GOLD_21K", "EGX_COMI"]}` | `200 OK` |
| `GET` | `/v1/quotes?items=...` | Batch resolution from comma-separated query string | `200 OK` |

---

## 2. Standard Headers & Error Contracts

### 2.1 Request Headers
- `x-api-key`: API key authentication token.
- `x-trace-id` (optional): Distributed tracing token. If absent, gateway generates one.
- `If-None-Match`: Client-cached ETag string for conditional `304 Not Modified` checks.

### 2.2 Response Headers
- `Content-Type`: `application/json; charset=utf-8`
- `ETag`: `W/"{hash}"` (deterministic weak or strong entity tag).
- `Cache-Control`: `public, max-age={ttl}` (computed dynamically by `MarketScheduleEngine`).
- `x-trace-id`: Propagated or generated trace identifier.
- `x-server-id`: Cluster instance identifier (configured via `SERVER_ID`).

### 2.3 Error Envelope (`HttpError`)
```json
{
  "code": "INVALID_VALIDATION",
  "message": "Unsupported currency code: XYZ",
  "details": {},
  "timestamp": "2026-11-16T14:00:00.000Z"
}
```
Authentication failures return:
```json
{
  "code": "UNAUTHORIZED",
  "message": "Missing or invalid API key.",
  "timestamp": "2026-11-16T14:00:00.000Z"
}
```
with HTTP status `401 Unauthorized`.
