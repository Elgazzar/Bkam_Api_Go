# REST Contract Compatibility: Live Snapshot Coordinator

**Feature Branch**: `109-live-snapshot-coordinator`  
**Date**: 2026-11-15  
**Spec Reference**: [spec.md](../spec.md)

---

## 1. Zero-Regression REST Contract Guarantee

All client-facing REST endpoints maintain 100% backwards-compatible JSON response structures, HTTP status codes, and Cache-Control headers when consuming `LiveSnapshotCoordinator`:

| Endpoint | Method | Upstream Domain Key | Client Response Shape | Cache-Control |
|:---|:---|:---|:---|:---|
| `/v1/rates` | `GET` | `rates` | `RateSnapshot` (`rates: CurrencyRate[]`, `availability`) | `public, max-age=60` |
| `/v1/gold` | `GET` | `gold` | `GoldPriceDocument` (`rates`, `international`) | `public, max-age=60` |
| `/v1/gas` | `GET` | `gas` | `GasPriceSnapshot` (`rates`) | `public, max-age=3600` |
| `/v1/stocks` | `GET` | `stocks:egx` | `StockPriceSnapshot` (`stocks: Record<string, StockItem>`) | `public, max-age=60` |
| `/v1/stocks/:market` | `GET` | `gcc-stocks:${market}` | `GccStockPriceSnapshot` (`stocks`) | `public, max-age=60` |
| `/v1/indices` | `GET` | `indices:egx` (or `?market`) | `IndexPriceSnapshot` (`indices`) | `public, max-age=60` |
| `/v1/quotes` | `POST` / `GET` | Multi-domain parallel | `BatchQuoteResponse` (`quotes`, `unresolved`) | `public, max-age=60` |

---

## 2. Invalidation & Freshness Metadata

All endpoints serve standardized metadata fields:
- `updatedAt`: ISO 8601 string of official source update.
- `scrapedAt`: ISO 8601 string of platform retrieval.
- `freshness`: `"current"` during active trading sessions; `"stale"` or `"degraded"` if upstream push failed.
- `sourceTraceId`: Trace ID inherited from coordinator or push listener.
