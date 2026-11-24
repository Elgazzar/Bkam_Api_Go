# Data Model: Tiered Weekly Caching & API Performance Optimization

## Cache Storage Entities (`flat-cache`)

### 1. Weekly Bucket Document Cache Entry
- **Cache Key**: `${collectionName}_weekly/${weekKey}` (e.g. `stock_prices_history_weekly/2026-W42`)
- **Storage Target**: Portable disk cache (`.cache/api-cache`) via `FlatCacheProvider`
- **Payload Schema**:
  ```typescript
  Record<string, {
    date: string;               // ISO date key YYYY-MM-DD
    stocks: Record<string, {    // Stock quotes keyed by symbol
      symbol: string;
      price: number;
      change: number;
      changePercent: number;
      previousClose?: number;
    }>;
    updatedAt?: string | number;
    scrapedAt?: string | number;
  }>
  ```
- **TTL Strategy**:
  - **Sealed Past Week** (`weekKey !== currentWeekKey`): 30 days (`30 * 24 * 60 * 60` seconds).
  - **Active Week** (`weekKey === currentWeekKey`) during market hours: Bypassed from disk, cached only in RAM (5-second deduplication).
  - **Active Week** outside market hours: Persisted to disk with dynamic TTL = `secondsUntilNextMarketOpen` (expires at 09:30 AM Cairo next trading day).

---

### 2. Range Result Cache Entry (Pure Historical Range)
- **Cache Key**: `stock_range:${symbol}:${startDate}:${endDate}` (e.g. `stock_range:COMI:2026-05-01:2026-10-20`)
- **Applicability**: Requests where `endDate < todayCairo` (pure immutable past).
- **Storage Target**: `FlatCacheProvider`
- **Payload Schema**: `StockRangeItem[]`
- **TTL**: 30 days (`30 * 24 * 60 * 60` seconds).

---

### 3. Historical Prefix Cache Entry (Range Ending Today)
- **Cache Key**: `stock_range_prefix:${symbol}:${startDate}:${yesterdayCairo}` (e.g. `stock_range_prefix:COMI:2026-05-01:2026-10-23`)
- **Applicability**: Requests where `endDate >= todayCairo` and `startDate <= yesterdayCairo`.
- **Storage Target**: `FlatCacheProvider`
- **Payload Schema**: `StockRangeItem[]` (excluding today).
- **TTL**: 30 days (`30 * 24 * 60 * 60` seconds).

---

## Response Item Model (`StockRangeItem`)

```typescript
export interface StockRangeItem {
  date: string;                       // YYYY-MM-DD
  symbol: string;                     // Canonical ticker symbol (e.g. "COMI")
  price: number;                      // Daily closing price (or live price for today)
  change: number;                     // Nominal change from previous trading session
  changePercent: number;              // Percentage change from previous trading session
  updatedAt: string | number;         // Timestamp of quote record
  intraday?: Array<{                  // ONLY present on today's record; omitted for historical days
    time: string;                     // HH:MM
    price: number;
  }>;
}
```

---

## Market Session Temporal Model

```typescript
export interface MarketScheduleState {
  currentCairoDate: string;           // YYYY-MM-DD in Africa/Cairo
  currentWeekKey: string;             // ISO calendar week YYYY-Www in Africa/Cairo
  isTradingHours: boolean;            // True if Sun–Thu between 09:30 and 16:00 Cairo
  nextMarketOpenTimestampMs: number;  // Epoch ms of next trading day at 09:30 Cairo
  secondsUntilNextMarketOpen: number; // Duration in seconds until next market open
}
```
