# Data Model: Deepen Intraday Coordinator

**Branch**: `107-deepen-intraday-coordinator` | **Date**: 2026-11-15 | **Spec**: [spec.md](./spec.md)

---

## Entity Relationships & Seam Architecture

```mermaid
classDiagram
    class IntradayCoordinator {
        +getLatestIntraday(assetClass, target, options) Promise~LatestIntradayResult~
        +resolveLatestTimeline(options) Promise~LatestIntradayTimelineResult~
        +lookupSymbolPoints(pointsMap, symbol) IntradayPoint[]
        +normalizePoints(raw) IntradayPoint[]
        -resolveLiveRatesIntraday(...)
        -resolveLiveGoldIntraday(...)
        -resolveLiveStocksIntraday(...)
        -resolveLiveIndicesIntraday(...)
    }

    class LatestIntradayOptions {
        +string bank
        +string currency
        +string market
        +any liveSnapshot
    }

    class LatestIntradayResult {
        +any data
        +boolean isToday
        +string sessionDate
        +number cacheTtlSeconds
    }

    class RatesIntradayPayload {
        +string currency
        +string bank
        +number buy
        +number sell
        +number previousBuyPrice
        +number previousSellPrice
        +IntradayPoint[] intraday
        +string updatedAt
    }

    class GoldIntradayPayload {
        +string date
        +Record~string, GoldKaratItem~ karats
        +GoldKaratItem international
        +string updatedAt
    }

    class StockIntradayPayload {
        +string symbol
        +string name_en
        +string name_ar
        +number price
        +number change
        +number changePercent
        +number previousPrice
        +IntradayPoint[] intraday
        +string updatedAt
    }

    class IndexIntradayPayload {
        +string code
        +string name_en
        +string name_ar
        +number value
        +number change
        +number changePercent
        +number previousValue
        +IntradayPoint[] intraday
        +string updatedAt
    }

    IntradayCoordinator ..> LatestIntradayOptions : receives
    IntradayCoordinator ..> LatestIntradayResult : returns
    LatestIntradayResult *-- RatesIntradayPayload : wraps
    LatestIntradayResult *-- GoldIntradayPayload : wraps
    LatestIntradayResult *-- StockIntradayPayload : wraps
    LatestIntradayResult *-- IndexIntradayPayload : wraps
```

---

## Interface Definitions & Types

### 1. `LatestIntradayOptions`
```typescript
export interface LatestIntradayOptions {
  // Optional bank short code filter for bank exchange rates (e.g. 'NBE', 'CIB')
  bank?: string;
  // Optional currency code filter for bank rates (e.g. 'USD', 'EUR')
  currency?: string;
  // Optional regional GCC equity exchange (e.g. 'saudi', 'uae', 'qatar', 'kuwait')
  market?: GccMarket;
  // Optional pre-resolved live snapshot payload to avoid redundant fetches or for unit testing
  liveSnapshot?: any;
}
```

### 2. `LatestIntradayResult<T>`
```typescript
export interface LatestIntradayResult<T = any> {
  // Domain payload matching exact client JSON route shape
  data: T;
  // Boolean flag indicating if data represents active today session
  isToday: boolean;
  // Resolved trading session calendar date (YYYY-MM-DD)
  sessionDate: string;
  // Standard HTTP Cache-Control max-age lifetime (60 for today, 3600 for past)
  cacheTtlSeconds: number;
}
```

---

## State Transition & Session Lifecycles

```mermaid
stateDiagram-v2
    [*] --> CheckMemoryBuffer: getLatestIntraday(assetClass, target)
    
    CheckMemoryBuffer --> ActiveSessionToday: Memory Buffer has Today's Ticks
    CheckMemoryBuffer --> CheckColdFirestore: Memory Buffer is Empty

    CheckColdFirestore --> ActiveSessionToday: Today doc exists in Firestore
    CheckColdFirestore --> BacktrackSession: Today doc is empty (Closed / Off-Hours)

    BacktrackSession --> ConcludedSession: Query Firestore limit(1) descending (<= 7 days)
    BacktrackSession --> DegradedEmpty: Exceeds 7-day ceiling

    ActiveSessionToday --> ExtractLiveQuotes: Query In-Memory Snapshot
    ExtractLiveQuotes --> AttachLiveBaselines: YesterdayClose / OldPrice
    AttachLiveBaselines --> ReturnEnvelope: isToday: true, max-age: 60s

    ConcludedSession --> QueryHistoryEngine: getDateSnapshot(sessionDate)
    QueryHistoryEngine --> AttachHistoricalBaselines: Previous Session Close
    AttachHistoricalBaselines --> ReturnEnvelope: isToday: false, max-age: 3600s

    DegradedEmpty --> ReturnEnvelope: isToday: false, empty points
```

