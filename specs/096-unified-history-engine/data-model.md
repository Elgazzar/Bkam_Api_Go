# Phase 1 Data Model: Unified History Resolution Engine

## Storage Models & Collections

### 1. Consolidated ISO Weekly Bucket Document (`*_history_weekly/{weekKey}`)

Each weekly document groups all daily market snapshots for an ISO 8601 calendar week (`YYYY-Www`). Intraday tick points are strictly excluded.

```typescript
// Keyed by ISO calendar week (e.g., "2026-W43")
interface WeeklyBucketDocument<T = any> {
  // Keyed by calendar date in YYYY-MM-DD format (e.g. "2026-10-21")
  [date: string]: T;
}
```

#### Daily Snapshot Types per Asset Class

- **Bank Rates Daily Entry** (`rates_history_weekly/{weekKey}`):
  ```typescript
  interface DailyRatesSnapshot {
    date: string;                     // "2026-10-21"
    archivedAt: string;               // ISO timestamp
    rates: Record<string, {           // Keyed by lower-case bankCode (e.g., "cib")
      [currencyCode: string]: {       // Upper-case currency (e.g., "USD")
        buy: number;
        sell: number;
        oldBuy?: number | null;
        oldSell?: number | null;
        trend?: 'up' | 'down';
        dailyHighBuy?: number | null;
        dailyHighSell?: number | null;
        dailyLowBuy?: number | null;
        dailyLowSell?: number | null;
      };
    }>;
  }
  ```

- **Gold Daily Entry** (`gold_prices_history_weekly/{weekKey}`):
  ```typescript
  interface DailyGoldSnapshot {
    date: string;
    updatedAt: string;
    rates: {
      "24k": { buy: number; sell: number };
      "21k": { buy: number; sell: number };
      "18k"?: { buy: number; sell: number };
      "pound": { buy: number; sell: number };
      [custom: string]: any;
    };
    international?: {
      priceOunceUsd: number;
      updatedAt?: string;
    };
  }
  ```

- **Stocks Daily Entry** (`stock_prices_history_weekly/{weekKey}`):
  ```typescript
  interface DailyStockSnapshot {
    date: string;
    source: string;
    currency: string;
    updatedAt: string;
    scrapedAt?: string;
    stocks: Record<string, {          // Keyed by symbol (e.g., "COMI")
      symbol: string;
      nameAr?: string;
      nameEn?: string;
      price: number;
      open?: number;
      high?: number;
      low?: number;
      close?: number;
      change: number;
      changePercent: number;
    }>;
  }
  ```

- **Indices Daily Entry** (`index_prices_history_weekly/{weekKey}`):
  ```typescript
  interface DailyIndexSnapshot {
    date: string;
    source: string;
    currency: string;
    updatedAt: string;
    scrapedAt?: string;
    indices: Record<string, {         // Keyed by symbol (e.g., "EGX30")
      code: string;
      name: string;
      price: number;
      change: number;
      changePercent: number;
    }>;
  }
  ```

- **GCC Regional Equities Daily Entry** (`stock_prices_{saudi|uae|qatar}_history_weekly/{weekKey}`):
  ```typescript
  interface DailyGccStockSnapshot {
    date: string;
    market: 'saudi' | 'uae' | 'qatar';
    currency: string;
    updatedAt: string;
    scrapedAt?: string;
    stocks: Record<string, {
      symbol: string;
      name: string;
      price: number;
      change: number;
      changePercent: number;
    }>;
  }
  ```

---

## 2. In-Memory RAM Structures (Tier 1)

### `ActiveWeeklyBucketStore`
Maintains volatile in-memory cache of the currently active mutating calendar week (`currentWeekKey`) pushed via Firestore `onSnapshot`.

```typescript
// Memory map keyed by: collectionName -> weekKey -> WeeklyBucketDocument
Map<string, Map<string, Record<string, any>>>
```

### `NegativeDatesCache`
Prevents repeated database reads when querying known non-trading days (weekends, official holidays).

```typescript
// Keyed by "${collectionName}_${date}" -> expiration timestamp (ms)
Map<string, number>
// TTL: 24 hours for past dates, 5 minutes for dynamic today
```

---

## 3. Persistent Disk Cache (`flat-cache` via `CacheService`) (Tier 2 & 3)

| Cache Key Pattern | Tier | Purpose | TTL | Invalidation / Eviction |
| :--- | :--- | :--- | :--- | :--- |
| `${col}_weekly/${weekKey}` | Tier 3 | Sealed Past Week document | 30 Days (`2,592,000s`) | Immutable; auto-eviction on expiry |
| `${col}_weekly/${currentWeekKey}` | Tier 1/3 | Active Week during off-hours | Dynamic (`until 09:30 AM`) | Market open on next trading day |
| `stock_range:${sym}:${start}:${end}` | Tier 2 | Pure past stock historical range | 30 Days | Immutable |
| `index_range:${sym}:${start}:${end}` | Tier 2 | Pure past index historical range | 30 Days | Immutable |
| `gold_range:${start}:${end}` | Tier 2 | Pure past gold historical range | 30 Days | Immutable |

*Note: Bank Rates deliberately omits Tier 2 pre-computed range keys to avoid combinatorial cache key explosion.*
