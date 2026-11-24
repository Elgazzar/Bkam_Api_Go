# Phase 1 Data Model: Dedicated Latest Intraday Timeline Endpoints

**Feature**: `097-latest-intraday-api`  
**Date**: 2026-10-24  
**Status**: Completed  

---

## 1. Domain Entities & Interfaces

### 1.1 `IntradayPoint`
Represents an individual intra-session price change observation event.
```typescript
export interface IntradayPoint {
  // ISO-8601 formatted timestamp string (e.g. "2026-10-22T10:05:00.000Z")
  time: string;
  // Numerical price or exchange rate reading at the specified timestamp
  price: number;
}
```

### 1.2 `LatestIntradayTimelineResult`
Internal resolution outcome returned by the centralized `IntradayCoordinator`.
```typescript
export interface LatestIntradayTimelineResult {
  // The resolved ISO calendar date of the session (YYYY-MM-DD)
  date: string;
  // Dictionary mapping symbols, karats, or bank keys to their normalized point arrays
  pointsMap: Record<string, any>;
  // Boolean flag indicating whether the resolved session is currently open today
  isToday: boolean;
}
```

### 1.3 `ResolveLatestTimelineOptions`
Parameter envelope passed to `IntradayCoordinator.resolveLatestTimeline`.
```typescript
export interface ResolveLatestTimelineOptions {
  // Financial asset domain
  category: 'rates' | 'gold' | 'stocks' | 'indices';
  // Target collection name in Firestore (e.g., 'stock_prices_intraday', 'rates_intraday')
  collectionName: string;
  // Optional regional market code for regional equities
  market?: string;
  // Asynchronous delegate performing direct single-read descending query in Firestore
  fetchLatestFirestoreSession?: (collectionName: string) => Promise<{ date: string; pointsMap: Record<string, any> } | null>;
}
```

### 1.4 Client Response Envelopes

#### A. Stock Latest Intraday Response
```typescript
export interface StockLatestIntradayResponse {
  symbol: string;               // e.g. "COMI"
  date: string;                 // ISO date: "2026-10-22"
  currency: string;             // ISO currency: "EGP"
  price: number;                // e.g. 95.5
  change: number;               // e.g. 1.2
  changePercent: number;        // e.g. 1.27
  updatedAt: string;            // ISO-8601 timestamp: "2026-10-22T14:30:00.000Z"
  intraday: IntradayPoint[];    // Array of intra-session ticks
}
```

#### B. Index Latest Intraday Response
```typescript
export interface IndexLatestIntradayResponse {
  symbol: string;               // e.g. "EGX30"
  date: string;                 // ISO date: "2026-10-22"
  currency: string;             // ISO currency: "EGP"
  price: number;                // e.g. 30540.2
  change: number;               // e.g. 185.4
  changePercent: number;        // e.g. 0.61
  updatedAt: string;            // ISO-8601 timestamp
  intraday: IntradayPoint[];    // Array of intra-session ticks
}
```

#### C. Gold Latest Intraday Response
```typescript
export interface GoldLatestIntradayItemResponse {
  item: string;                 // e.g. "21", "24", "18", "pound", "international"
  date: string;                 // ISO date: "2026-10-24"
  currency: string;             // "EGP" or "USD"
  buy: number;                  // Buy price
  sell: number;                 // Sell price
  updatedAt: string;            // ISO-8601 timestamp
  intraday: IntradayPoint[];    // Array of intra-session ticks
}

export interface GoldLatestIntradayAllResponse {
  date: string;                 // ISO date: "2026-10-24"
  currency: string;             // "EGP"
  updatedAt: string;            // ISO-8601 timestamp
  karats: Record<string, {
    buy: number;
    sell: number;
    intraday: IntradayPoint[];
  }>;
  pound?: {
    buy: number;
    sell: number;
    intraday: IntradayPoint[];
  };
  international?: {
    buy: number;
    sell: number;
    currency: "USD";
    intraday: IntradayPoint[];
  };
}
```

#### D. Bank Rates Latest Intraday Response
```typescript
export interface RateLatestIntradayResponse {
  currency: string;             // e.g. "USD"
  bank?: string;                // e.g. "cib" (if filtered)
  date: string;                 // ISO date: "2026-10-24"
  buy: number;                  // e.g. 48.65
  sell: number;                 // e.g. 48.75
  updatedAt: string;            // ISO-8601 timestamp
  intraday: IntradayPoint[];    // Array of intra-session ticks
}
```

---

## 2. Validation & Boundary Rules

1. **Symbol Validation**:
   - Stocks: Validated against `stockCatalog` for EGX, and regional catalogs (`saudiStockCatalog`, `uaeStockCatalog`, `qatarStockCatalog`, `kuwaitStockCatalog`) for GCC. Unknown symbol rejects with 400 `INVALID_VALIDATION`.
   - Indices: Validated against `indexCatalog` (`isValidIndexCode`). Unknown index code rejects with 400 `INVALID_VALIDATION`.
   - Gold: Allowed values: `24`, `21`, `18`, `pound`, `international`, `all` (or empty for all). Invalid values reject with 400 `INVALID_VALIDATION`.
   - Bank Rates: Currency must be a 3-character ISO code (e.g. `USD`, `EUR`). Optional bank must be in `SUPPORTED_BANKS`.
2. **Empty Intraday Array for Inactive Assets**:
   - If a valid catalog security exists but recorded zero ticks in the resolved session, `intraday` MUST be `[]` (never `null` or `undefined`).
3. **Session Lookback Ceiling**:
   - The resolved session date MUST NOT be older than 7 calendar days relative to `todayCairo`. If older than 7 days, the session is treated as empty and falls back to daily closing summaries with `intraday: []`.
