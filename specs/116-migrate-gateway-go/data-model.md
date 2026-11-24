# Data Model: Migrate Bkam Financial API Gateway to Idiomatic Go

**Feature**: `116-migrate-gateway-go`  
**Date**: 2026-11-16  
**Status**: Completed  

---

## 1. Domain Entities & Memory Models

### 1.1 Currency Rates Domain

```go
package domain

import "time"

// Trend represents directional price movement ('up' | 'down')
type Trend string

const (
    TrendUp   Trend = "up"
    TrendDown Trend = "down"
)

// CurrencyRate represents an enriched exchange rate quote from a specific bank
type CurrencyRate struct {
    NameAr            string    `json:"name_ar"`
    NameEn            string    `json:"name_en"`
    Logo              string    `json:"logo"`
    BankCode          string    `json:"bankCode"`
    CurrencyCode      string    `json:"currencyCode"`
    BuyRate           float64   `json:"buyRate"`
    SellRate          float64   `json:"sellRate"`
    Trend             Trend     `json:"trend"`
    OldBuyRate        *float64  `json:"oldBuyRate"`
    OldSellRate       *float64  `json:"oldSellRate"`
    RetrievedAt       time.Time `json:"retrievedAt"`
    BankUpdatedAt     *time.Time`json:"bankUpdatedAt"`
    Freshness         string    `json:"freshness"`
    SourceTraceID     string    `json:"sourceTraceId"`
    DailyHighBuy      *float64  `json:"dailyHighBuy"`
    DailyHighSell     *float64  `json:"dailyHighSell"`
    DailyLowBuy       *float64  `json:"dailyLowBuy"`
    DailyLowSell      *float64  `json:"dailyLowSell"`
    YesterdayCloseBuy *float64  `json:"yesterdayCloseBuy"`
    YesterdayCloseSell*float64  `json:"yesterdayCloseSell"`
    LastCloseDate     *string   `json:"lastCloseSnapshotDate,omitempty"`
}

// AvailabilityRecord represents operational scraping status for a bank
type AvailabilityRecord struct {
    BankCode      string    `json:"bankCode"`
    Status        string    `json:"status"` // "current" | "stale"
    LastSuccessAt time.Time `json:"lastSuccessAt"`
}

// RateSnapshot represents the public rates snapshot envelope
type RateSnapshot struct {
    RequestedAt    time.Time            `json:"requestedAt"`
    Freshness      string               `json:"freshness"` // "current" | "partial" | "stale"
    RequestFilters map[string]string    `json:"requestFilters"`
    Rates          []CurrencyRate       `json:"rates"`
    Availability   []AvailabilityRecord `json:"availability"`
    Warnings       []string             `json:"warnings"`
    Currency       string               `json:"currency,omitempty"`
}
```

---

### 1.2 Commodities Domain (Gold & Petroleum)

```go
package domain

import "time"

// GoldKaratPrice represents buy/sell rates for a specific gold karat
type GoldKaratPrice struct {
    Buy            float64  `json:"buy"`
    Sell           float64  `json:"sell"`
    OldBuy         *float64 `json:"oldBuy,omitempty"`
    OldSell        *float64 `json:"oldSell,omitempty"`
    Trend          Trend    `json:"trend"`
    YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
}

// GoldSnapshot represents the precious metals quote envelope
type GoldSnapshot struct {
    Source        string                    `json:"source"`
    Currency      string                    `json:"currency"`
    UpdatedAt     time.Time                 `json:"updatedAt"`
    ScrapedAt     time.Time                 `json:"scrapedAt"`
    Freshness     string                    `json:"freshness"`
    SourceTraceID string                    `json:"sourceTraceId"`
    Rates         map[string]GoldKaratPrice `json:"rates"` // "24k", "21k", "18k", "ounce", "pound"
}

// GasPriceItem represents fuel pump price details
type GasPriceItem struct {
    Price          float64  `json:"price"`
    Unit           string   `json:"unit"`
    OldPrice       *float64 `json:"oldPrice,omitempty"`
    Trend          Trend    `json:"trend"`
    YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
}

// GasSnapshot represents petroleum fuel rates
type GasSnapshot struct {
    Source        string                  `json:"source"`
    Currency      string                  `json:"currency"`
    UpdatedAt     time.Time               `json:"updatedAt"`
    ScrapedAt     time.Time               `json:"scrapedAt"`
    Freshness     string                  `json:"freshness"`
    SourceTraceID string                  `json:"sourceTraceId"`
    Rates         map[string]GasPriceItem `json:"rates"` // "benzene95", "benzene92", "benzene80", "diesel", "cng"
}
```

---

### 1.3 Equities & Benchmark Indices Domain

```go
package domain

import "time"

// StockPriceItem represents quote for an individual equity security
type StockPriceItem struct {
    Symbol         string   `json:"symbol"`
    NameAr         string   `json:"name_ar,omitempty"`
    NameEn         string   `json:"name_en,omitempty"`
    Price          float64  `json:"price"`
    Change         float64  `json:"change"`
    ChangePercent  float64  `json:"changePercent"`
    Volume         int64    `json:"volume,omitempty"`
    Value          float64  `json:"value,omitempty"`
    Open           *float64 `json:"open,omitempty"`
    High           *float64 `json:"high,omitempty"`
    Low            *float64 `json:"low,omitempty"`
    Close          *float64 `json:"close,omitempty"`
    YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
    Trend          Trend    `json:"trend"`
    TradeCount     int64    `json:"tradeCount,omitempty"`
}

// StockSnapshot represents equity quotes for a specific market exchange
type StockSnapshot struct {
    Source        string                    `json:"source"`
    Market        string                    `json:"market"` // "egx", "saudi", "uae", "qatar", "kuwait"
    Currency      string                    `json:"currency"` // "EGP", "SAR", "AED", "QAR", "KWD"
    UpdatedAt     time.Time                 `json:"updatedAt"`
    ScrapedAt     time.Time                 `json:"scrapedAt"`
    Freshness     string                    `json:"freshness"`
    SourceTraceID string                    `json:"sourceTraceId"`
    Stocks        map[string]StockPriceItem `json:"stocks"`
}

// IndexPriceItem represents market benchmark index figures
type IndexPriceItem struct {
    IndexCode      string   `json:"indexCode"`
    Name           string   `json:"name"`
    Price          float64  `json:"price"`
    Change         float64  `json:"change"`
    ChangePercent  float64  `json:"changePercent"`
    High           *float64 `json:"high,omitempty"`
    Low            *float64 `json:"low,omitempty"`
    YesterdayClose *float64 `json:"yesterdayClose,omitempty"`
    Trend          Trend    `json:"trend"`
}

// IndexSnapshot represents index benchmarks
type IndexSnapshot struct {
    Source        string                    `json:"source"`
    Market        string                    `json:"market"`
    Currency      string                    `json:"currency"`
    UpdatedAt     time.Time                 `json:"updatedAt"`
    ScrapedAt     time.Time                 `json:"scrapedAt"`
    Freshness     string                    `json:"freshness"`
    SourceTraceID string                    `json:"sourceTraceId"`
    Indices       map[string]IndexPriceItem `json:"indices"`
}
```

---

### 1.4 Centralized Live Snapshot Envelope (`LiveSnapshotCoordinator`)

```go
package engine

import "time"

// LiveSnapshotEnvelope encapsulates any domain payload with metadata
type LiveSnapshotEnvelope struct {
    Domain        string    `json:"domain"`        // "rates", "gold", "gas", "stocks", "indices", "gcc-stocks"
    Market        string    `json:"market,omitempty"` // "egx", "saudi", "uae", "qatar", "kuwait"
    Currency      string    `json:"currency"`      // "EGP", "SAR", "AED", "QAR", "KWD"
    UpdatedAt     time.Time `json:"updatedAt"`
    ScrapedAt     time.Time `json:"scrapedAt"`
    Freshness     string    `json:"freshness"`     // "current", "partial", "stale", "degraded"
    SourceTraceID string    `json:"sourceTraceId"`
    Data          any       `json:"data"`          // The concrete domain snapshot
}
```

---

### 1.5 Historical Multi-Tier Data Layout (`HistoryResolutionEngine`)

```go
package engine

import (
    "sync"
    "time"
)

// HistoryResolutionEngine coordinates historical archives across memory and disk
type HistoryResolutionEngine struct {
    mu                  sync.RWMutex
    // Tier 1: Canonical composite key -> daily date map -> snapshot payload
    // Example key: "rates_history_weekly/2026-W47" -> map["2026-11-16"]payload
    activeWeeklyBuckets map[string]map[string]any
    storageDir          string
}

// ResolvedDateSnapshot represents a single calendar date archive envelope
type ResolvedDateSnapshot struct {
    Data            any       `json:"data"`
    IsToday         bool      `json:"isToday"`
    Date            string    `json:"date"` // "YYYY-MM-DD"
    CacheTTLSeconds int       `json:"cacheTtlSeconds"`
}

// RangeOptions specifies interval and filters for multi-day ranges
type RangeOptions struct {
    StartDate string // "YYYY-MM-DD"
    EndDate   string // "YYYY-MM-DD"
    Symbol    string // Optional ticker (stocks/indices)
    Bank      string // Optional bank code (rates)
    Currency  string // Optional currency code
    Market    string // Optional regional market
}
```

---

### 1.6 Market Schedule Engine

```go
package engine

import "time"

// MarketScheduleEngine evaluates trading session status and TTLs in Africa/Cairo
type MarketScheduleEngine struct {
    cairoLoc *time.Location
    holidays map[string]string // "YYYY-MM-DD" -> Holiday Name
}

// MarketState reflects exchange state at a point in time
type MarketState struct {
    CurrentCairoDate           string `json:"currentCairoDate"`
    CairoDayOfWeek             int    `json:"cairoDayOfWeek"`
    IsTradingHours             bool   `json:"isTradingHours"`
    SecondsUntilNextMarketOpen int    `json:"secondsUntilNextMarketOpen"`
    NextMarketOpenTimestampMs  int64  `json:"nextMarketOpenTimestampMs"`
}
```

---

## 2. In-Memory Lifecycle States

```
[Cold Boot] 
      │
      ▼
Hydrate Live Snapshots & Active Week from Persistent Storage (Disk/JSON)
      │
      ▼
[Warmed Serving State] (Zero Firestore Reads)
      │
      ├─► Upstream Firestore Push Event ──► Update Memory In-Place + Async Disk Persist
      │
      ├─► Midnight Cairo Boundary (00:01 AM) ──► Seal Prior Week to Disk (30-day TTL)
      │                                         & Re-target Active Week Listeners
      │
      ├─► 08:00 AM Cairo Boundary ────────────► Rotate Intraday Listeners to New Day
      │
      └─► Shutdown Signal (SIGTERM) ──────────► Cancel Worker Goroutines & Flush Memory
```
