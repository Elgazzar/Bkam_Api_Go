# Data Model: Deepen Historical Engine & Demote Leaky Firestore Storage Layer

**Feature Branch**: `110-deepen-history-engine`  
**Date**: 2026-11-15  
**Spec**: [spec.md](./spec.md)

---

## 1. Domain Entities & Type Definitions

### 1.1 `AssetClass` & `GccMarket`
Canonical domain identifiers for supported asset categories and regional exchanges:
```typescript
export type AssetClass = 'rates' | 'gold' | 'stocks' | 'indices';
export type GccMarket = 'saudi' | 'uae' | 'qatar' | 'kuwait';
```

---

### 1.2 `RangeOptions`
Input configuration for resolving multi-day historical time-series ranges:
```typescript
export interface RangeOptions {
  // ISO calendar start date formatted as YYYY-MM-DD (inclusive)
  startDate: string;
  // ISO calendar end date formatted as YYYY-MM-DD (inclusive)
  endDate: string;
  // Optional ticker symbol filter (e.g. 'COMI', 'ARAMCO', 'EGX30')
  symbol?: string;
  // Optional currency code filter for bank exchange rates (e.g. 'USD', 'EUR')
  currency?: string;
  // Optional bank short code filter (e.g. 'NBE', 'CIB')
  bank?: string;
  // Optional regional market exchange (e.g. 'saudi', 'uae', 'qatar', 'kuwait')
  market?: GccMarket;
  // Optional pre-resolved live snapshot payload to avoid redundant fetches in tests
  liveSnapshot?: any;
}
```

---

### 1.3 `RangePoint`
Standardized point observation in a historical time series:
```typescript
export interface RangePoint {
  // Calendar date formatted as YYYY-MM-DD
  date: string;
  // Primary price/rate observation
  price: number;
  // Numeric change compared to previous session
  change?: number;
  // Percentage change compared to previous session
  changePercent?: number;
  // ISO timestamp of the quote
  updatedAt?: string;
  // Optional symbol identifier (for stocks/indices)
  symbol?: string;
  // Optional currency identifier (for bank exchange rates)
  currency?: string;
  // Optional bank identifier (for bank exchange rates)
  bank?: string;
  // Optional live intraday movement points if this observation is dynamic today
  intraday?: Array<{
    timestamp: string;
    price: number;
  }>;
}
```

---

### 1.4 `ResolvedDateSnapshot`
Standardized envelope returned for single-date queries:
```typescript
export interface ResolvedDateSnapshot<T = any> {
  // Canonical data payload dictionary, or null if market was closed
  data: T | null;
  // Boolean flag indicating if requested date matches active Cairo trading day
  isToday: boolean;
  // Calendar date formatted as YYYY-MM-DD
  date: string;
  // Recommended public HTTP Cache-Control max-age in seconds
  cacheTtlSeconds: number;
}
```

---

## 2. Persistent & In-Memory Cache Key Registry

All caching is centralized exclusively in `HistoryResolutionEngine` using the shared `CacheService` (`flat-cache`) and internal maps:

| Cache Type | Key Pattern | TTL | Scope / Location |
| :--- | :--- | :--- | :--- |
| **Pure Historical Range** | `{asset}_range:{symbol}:{startDate}:{endDate}` | 30 days (2,592,000s) | Disk (`flat-cache`) |
| **Historical Range Prefix** | `{asset}_range_prefix:{symbol}:{startDate}:{yesterday}` | 30 days (2,592,000s) | Disk (`flat-cache`) |
| **Sealed Past Date Snapshot** | `{asset}_history:{date}` or `{asset}_{market}_history:{date}` | 30 days (2,592,000s) | Disk (`flat-cache`) |
| **Missing / Negative Date** | `{collectionName}_{date}` | 24 hours (past) / 5 mins (today) | In-Memory `missingDatesCache` |
| **Inflight Request Promise** | `{collectionName}_{date}` | Lifecycle of promise | In-Memory `inflightPromises` |

*Note*: The legacy `var/history-cache/*.json` format is completely superseded by `flat-cache`.

---

## 3. Database Adapter Seam Types

Stateless interfaces exposed by `firestore-service.ts`:
```typescript
export interface IFirestoreHistoryAdapter {
  // Fetch raw weekly bucket document dictionary from Firestore
  fetchWeeklyBucketDocument(collectionName: string, weekKey: string): Promise<Record<string, any> | null>;
  // Fetch raw legacy daily document from Firestore
  fetchDailyDocument(collectionName: string, date: string): Promise<any | null>;
}
```
