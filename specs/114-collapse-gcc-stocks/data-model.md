# Phase 1: Data Model & Schema Specifications

**Feature**: `specs/114-collapse-gcc-stocks`

---

## 1. Domain Entities

### MarketIdentifier
Normalized market identifier representing supported equity exchanges.

```typescript
export type CanonicalMarketId = "egx" | "saudi" | "uae" | "qatar" | "kuwait";
export type MarketCountryCode = "eg" | "sa" | "ae" | "qa" | "kw";
export type MarketQueryInput = CanonicalMarketId | MarketCountryCode | string;
```

**Normalization Mapping**:
| Input | Canonical Market | Country Code | Currency | Default Collection / Partition |
| :--- | :--- | :--- | :--- | :--- |
| `egx`, `eg` | `egx` | `EG` | `EGP` | `stock_prices_snapshot` (`egx`) |
| `saudi`, `sa` | `saudi` | `SA` | `SAR` | `gcc_stock_prices_sa` (`sa`) |
| `uae`, `ae` | `uae` | `AE` | `AED` | `gcc_stock_prices_ae` (`ae`) |
| `qatar`, `qa` | `qatar` | `QA` | `QAR` | `gcc_stock_prices_qa` (`qa`) |
| `kuwait`, `kw` | `kuwait` | `KW` | `KWD` | `gcc_stock_prices_kw` (`kw`) |

---

### StockPriceSnapshot
The root collection envelope returned by real-time endpoints.

```typescript
export interface StockPriceSnapshot {
  // ISO calendar date of the snapshot (YYYY-MM-DD)
  date: string;
  // Canonical market identifier
  market?: string;
  // Two-letter ISO country code
  country?: string;
  // Trading currency code (EGP, SAR, AED, QAR, KWD)
  currency: string;
  // ISO 8601 timestamp of last quote update
  updatedAt: string;
  // ISO 8601 timestamp when data was scraped/retrieved
  scrapedAt: string;
  // Freshness classification
  freshness: "current" | "stale" | "degraded";
  // Unique telemetry trace identifier
  sourceTraceId: string;
  // Data payload holding company quote mappings
  data: {
    country?: string;
    currency: string;
    updatedAt: string;
    // Map of ticker symbols to individual quote items
    stocks: Record<string, StockQuoteItem>;
  };
}
```

---

### StockQuoteItem
Individual company equity quote within a market snapshot.

```typescript
export interface StockQuoteItem {
  // Official exchange ticker symbol (e.g. "COMI", "1120", "EMAAR")
  symbol: string;
  // Company display name in Arabic
  nameAr?: string;
  // Company display name in English
  nameEn?: string;
  // Current trading price
  price: number;
  // Absolute price change since previous close
  change: number;
  // Percentage price change
  changePercent: number;
  // Trading direction trend indicator
  trend: "up" | "down" | "flat";
  // Total traded shares volume
  volume?: number;
  // Total traded monetary turnover
  turnover?: number;
  // Industry sector categorization
  sector?: string;
  // Associated benchmark indices (e.g. ["EGX30"], ["TASI"])
  benchmarkIndices?: string[];
  // ISO 8601 update timestamp
  updatedAt: string;
}
```

---

### IntradayTimeline
Tick series representing price movements throughout an active or recently concluded trading session.

```typescript
export interface IntradaySession {
  // Ticker symbol
  symbol: string;
  // Trading session calendar date (YYYY-MM-DD)
  date: string;
  // Trading currency
  currency: string;
  // Concluding or latest trade price
  price: number;
  // Absolute change
  change: number;
  // Percentage change
  changePercent: number;
  // ISO 8601 timestamp
  updatedAt: string;
  // Flag indicating whether session represents active today or backtracked previous trading session
  isToday?: boolean;
  // Chronological array of minute-by-minute tick points
  intraday: Array<{
    time: string;
    price: number;
  }>;
}
```

---

## 2. Validation Rules (Zod Schemas)

1. **Market Query Schema**:
   - `market`: optional string, validated against case-insensitive regex `^(egx|eg|saudi|sa|uae|ae|qatar|qa|kuwait|kw)$`.
   - Defaults to `'egx'` if omitted.
2. **Historical Date Schema**:
   - `date`: required string, format `^\d{4}-\d{2}-\d{2}$`.
   - Date must not be in the future (relative to Africa/Cairo time).
3. **Range Query Schema**:
   - `start`: required date `^\d{4}-\d{2}-\d{2}$`.
   - `end`: required date `^\d{4}-\d{2}-\d{2}$`.
   - `end` must be greater than or equal to `start`.
4. **Symbol Validation**:
   - Non-empty alphanumeric string trimmed of whitespace.
   - Evaluated against `MarketCatalog` for existence and metadata enrichment.
