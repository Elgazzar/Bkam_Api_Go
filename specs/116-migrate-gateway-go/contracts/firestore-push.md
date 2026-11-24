# Inbound Push Stream Contracts: Migrate Bkam Financial API Gateway to Idiomatic Go

**Feature**: `116-migrate-gateway-go`  
**Date**: 2026-11-16  
**Status**: Completed  

---

## 1. Push Adapter Architecture (`FirestoreListenerService`)

The inbound push adapter converts Firestore push snapshot streams into thread-safe domain engine mutations with zero circular intermediate callbacks.

```text
Firestore Push Stream (SDK Iterators)
        │
        ├─► [Collection: rates] ───────────────► RatesProcessor.ProcessAndEnrich(rawDocs)
        │                                             │
        │                                             ▼
        │                                        LiveSnapshotCoordinator.UpdateSnapshot("rates", data)
        │
        ├─► [Doc: gold_prices/latest] ─────────► LiveSnapshotCoordinator.UpdateSnapshot("gold", data)
        │
        ├─► [Doc: gas_prices/latest] ──────────► LiveSnapshotCoordinator.UpdateSnapshot("gas", data)
        │
        ├─► [Doc: stock_prices/latest] ────────► LiveSnapshotCoordinator.UpdateSnapshot("stocks", data, "egx")
        │
        ├─► [Doc: index_prices/latest] ────────► LiveSnapshotCoordinator.UpdateSnapshot("indices", data, "egx")
        │
        ├─► [Doc: stock_prices_{mkt}/latest] ──► LiveSnapshotCoordinator.UpdateSnapshot("gcc-stocks", data, mkt)
        │
        ├─► [Doc: *_intraday/:today] ──────────► IntradayTimelineStore.UpdateCategory(cat, points, today)
        │
        └─► [Doc: *_history_weekly/:weekKey] ──► HistoryResolutionEngine.SetActiveWeeklyBucket(col, weekKey, data)
```

---

## 2. Inbound Stream Subscriptions

| Source Path | Stream Type | Trigger Frequency | Target Coordinator Method |
|---|---|---|---|
| `rates` | Collection | Real-time on any bank rate change | `RatesProcessor` $\rightarrow$ `LiveSnapshotCoordinator.UpdateSnapshot("rates", ...)` |
| `gold_prices/latest` | Document | Real-time scraper push | `LiveSnapshotCoordinator.UpdateSnapshot("gold", ...)` |
| `gas_prices/latest` | Document | Real-time scraper push | `LiveSnapshotCoordinator.UpdateSnapshot("gas", ...)` |
| `stock_prices/latest` | Document | Trading session tick pushes | `LiveSnapshotCoordinator.UpdateSnapshot("stocks", ..., "egx")` |
| `index_prices/latest` | Document | Trading session tick pushes | `LiveSnapshotCoordinator.UpdateSnapshot("indices", ..., "egx")` |
| `stock_prices_saudi/latest` | Document | Real-time push | `LiveSnapshotCoordinator.UpdateSnapshot("gcc-stocks", ..., "saudi")` |
| `stock_prices_uae/latest` | Document | Real-time push | `LiveSnapshotCoordinator.UpdateSnapshot("gcc-stocks", ..., "uae")` |
| `stock_prices_qatar/latest` | Document | Real-time push | `LiveSnapshotCoordinator.UpdateSnapshot("gcc-stocks", ..., "qatar")` |
| `stock_prices_kuwait/latest` | Document | Real-time push | `LiveSnapshotCoordinator.UpdateSnapshot("gcc-stocks", ..., "kuwait")` |
| `rates_intraday/{today}` | Document | Minute intervals | `IntradayTimelineStore.UpdateCategory("rates", points, today)` |
| `gold_prices_intraday/{today}`| Document | Minute intervals | `IntradayTimelineStore.UpdateCategory("gold", points, today)` |
| `stock_prices_intraday/{today}`| Document | Minute intervals | `IntradayTimelineStore.UpdateCategory("stocks", points, today)` |
| `index_prices_intraday/{today}`| Document | Minute intervals | `IntradayTimelineStore.UpdateCategory("indices", points, today)` |
| `rates_history_weekly/{week}` | Document | Hourly/Daily consolidations | `HistoryResolutionEngine.SetActiveWeeklyBucket("rates_history_weekly", week, data)` |
| `stock_prices_history_weekly/{week}` | Document | Daily consolidations | `HistoryResolutionEngine.SetActiveWeeklyBucket("stock_prices_history_weekly", week, data)` |
| `index_prices_history_weekly/{week}` | Document | Daily consolidations | `HistoryResolutionEngine.SetActiveWeeklyBucket("index_prices_history_weekly", week, data)` |
| `gold_prices_history_weekly/{week}`  | Document | Daily consolidations | `HistoryResolutionEngine.SetActiveWeeklyBucket("gold_prices_history_weekly", week, data)` |

---

## 3. Reliability & Lifecycle Guarantees

1. **Context Cancellation**: On server shutdown, cancellation of the root `context.Context` terminates all iterator loops immediately without hanging.
2. **Exponential Backoff**: If an iterator returns an unexpected transport error (e.g. transient network blip), the worker sleeps with exponential backoff (1s, 2s, 4s up to 15s) and reconnects.
3. **Non-Blocking Ingestion**: Handlers never block the iterator loop; memory mutations execute under microsecond locks, and any disk persistence operations are offloaded to asynchronous background goroutines.
4. **Cold Boot Pre-Seeding**: Before the first Firestore snapshot arrives, both `LiveSnapshotCoordinator` and `HistoryResolutionEngine` hydrate their in-memory state from the local persistent disk cache directory.
