# Data Model: Stateless Batch Quotes API Endpoint

**Feature Branch**: `098-batch-quotes-api`
**Date**: 2026-10-31

## Overview
This document defines the data structures, Zod validation schemas, and entity relationships for the Stateless Batch Quotes API endpoint (`POST /v1/quotes` & `GET /v1/quotes`).

---

## Entities & Type Definitions

### 1. BatchQuoteRequest

Represents the inbound payload submitted by the mobile client.

```typescript
export interface BatchQuoteRequestBody {
  // Array of string asset identifiers (1 to 100 items)
  items: string[];
}

export interface BatchQuoteQueryParams {
  // Comma-separated list of string asset identifiers
  items: string;
}
```

#### Validation Rules (Zod Schema)
- `items` array must contain at least 1 item (`min(1)`).
- `items` array must not exceed 100 items (`max(100)`).
- Each item must be a non-empty trimmed string of length 1 to 50 characters.
- In query parameters, comma-separated tokens are split, trimmed, and empty tokens filtered out.

---

### 2. QuoteItem

Unified normalized representation of a market quote across all asset domains.

```typescript
export type AssetCategory = "currency" | "gold" | "index" | "stock";

export interface QuoteItem {
  // The exact identifier requested by the client (e.g. "CBE_USD", "21k", "EGX30", "COMI")
  id: string;

  // Domain category for classification and UI grouping
  category: AssetCategory;

  // Canonical human-readable display name (e.g. "Central Bank of Egypt - US Dollar", "21K Gold")
  name: string;

  // Primary numeric quote / last traded price
  price: number;

  // Optional buy rate (applicable to bank currencies and gold)
  buyPrice?: number | null;

  // Optional sell rate (applicable to bank currencies and gold)
  sellPrice?: number | null;

  // Absolute daily price change
  change?: number | null;

  // Percentage daily price change
  changePercent?: number | null;

  // Currency or pricing denomination (e.g. "EGP", "SAR", "AED", "QAR", "KWD", "USD")
  currency: string;

  // ISO 8601 string of the latest update timestamp
  updatedAt: string;
}
```

---

### 3. BatchQuoteResponse

Unified response envelope returned to the client.

```typescript
export interface BatchQuoteResponse {
  // Dictionary / Map keyed by the requested item identifier
  quotes: Record<string, QuoteItem>;

  // Array of identifier strings that could not be resolved or found
  unresolved: string[];

  // Server response ISO 8601 timestamp
  timestamp: string;
}
```

---

## State Transitions & Resolution Lifecycle

Since the endpoint is strictly stateless and read-only, there are no database state transitions. The request resolution follows a deterministic in-memory lookup pipeline:

```
[Inbound Request (POST/GET)]
         │
         ▼
[Validate & Deduplicate Identifiers]
         │
         ├── For each ID:
         │     │
         │     ├── 1. Test Currency Pattern (BANK_CURR)?
         │     │      ├── Matched: Lookup RatesSnapshotService -> Add to quotes
         │     │      └── Not matched / Not found -> Continue
         │     │
         │     ├── 2. Test Gold Pattern (24k, 21k, 18k, pound, ounce)?
         │     │      ├── Matched: Lookup GoldQueryService -> Add to quotes
         │     │      └── Not matched / Not found -> Continue
         │     │
         │     ├── 3. Test Index Pattern (EGX30, TASI, etc.)?
         │     │      ├── Matched: Lookup IndicesQueryService -> Add to quotes
         │     │      └── Not matched / Not found -> Continue
         │     │
         │     ├── 4. Test Stock Pattern (EGX or GCC)?
         │     │      ├── Matched: Lookup StockQueryService / GccStockQueryService -> Add to quotes
         │     │      └── Not matched / Not found -> Add to unresolved[]
         │
         ▼
[Assemble BatchQuoteResponse]
         │
         ▼
[Emit Response with Cache Headers & ETag]
```
