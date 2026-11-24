// Package firestore implements the inbound real-time push streaming adapter connecting Google Cloud Firestore snapshot streams to domain engines.
package firestore

// Import standard library packages for string manipulation, time handling, and formatting.
import (
	// fmt provides formatted string printing and error wrapping.
	"fmt"
	// strings provides uppercase and trimming normalization.
	"strings"
	// time handles timestamps and freshness evaluations.
	"time"

	// domain defines core rate models, snapshots, and the bank catalog registry.
	"bkam-api/internal/domain"
)

// RawBankRateInfo reflects the raw rate fields stored inside a Firestore bank document's rates map.
type RawBankRateInfo struct {
	// Buy is the bank's purchase rate.
	Buy float64 `json:"buy"`
	// Sell is the bank's selling rate.
	Sell float64 `json:"sell"`
	// OldBuy is the previous purchase rate if known.
	OldBuy *float64 `json:"oldBuy,omitempty"`
	// OldSell is the previous selling rate if known.
	OldSell *float64 `json:"oldSell,omitempty"`
	// Trend indicates direction ('up', 'down', 'stable').
	Trend string `json:"trend,omitempty"`
	// BuyTrend is alternative directional indicator.
	BuyTrend string `json:"buyTrend,omitempty"`
	// DailyHighBuy records the day's peak buy price.
	DailyHighBuy *float64 `json:"dailyHighBuy,omitempty"`
	// DailyHighSell records the day's peak sell price.
	DailyHighSell *float64 `json:"dailyHighSell,omitempty"`
	// DailyLowBuy records the day's lowest buy price.
	DailyLowBuy *float64 `json:"dailyLowBuy,omitempty"`
	// DailyLowSell records the day's lowest sell price.
	DailyLowSell *float64 `json:"dailyLowSell,omitempty"`
	// YesterdayCloseBuy records prior session close buy rate.
	YesterdayCloseBuy *float64 `json:"yesterdayCloseBuy,omitempty"`
	// YesterdayCloseSell records prior session close sell rate.
	YesterdayCloseSell *float64 `json:"yesterdayCloseSell,omitempty"`
	// LastCloseSnapshotDate records the ISO date string of prior session close.
	LastCloseSnapshotDate *string `json:"lastCloseSnapshotDate,omitempty"`
}

// RawBankDocument represents the root structure of a bank document in the Firestore "rates" collection.
type RawBankDocument struct {
	// BankID identifies the bank code (e.g. "nbe", "cib").
	BankID string `json:"bankId"`
	// Timestamp records document creation or scraping time.
	Timestamp any `json:"timestamp"`
	// BankUpdatedAt records when bank published rates.
	BankUpdatedAt any `json:"bankUpdatedAt,omitempty"`
	// Status indicates scraping status ("SUCCESS", "FAILED").
	Status string `json:"status"`
	// Rates maps currency codes (e.g. "USD", "EUR") to rate details.
	Rates map[string]RawBankRateInfo `json:"rates"`
}

// RatesProcessor transforms raw Firestore bank documents into enriched, validated RateSnapshot models.
type RatesProcessor struct {
	// catalog provides bank metadata (Arabic/English names, brand logos).
	catalog *domain.CatalogRegistry
}

// NewRatesProcessor constructs a new RatesProcessor instance.
func NewRatesProcessor() *RatesProcessor {
	// Return processor backed by singleton catalog registry.
	return &RatesProcessor{
		// Assign catalog registry.
		catalog: domain.GetCatalogRegistry(),
	}
}

// ProcessRawDocuments processes a collection of raw bank documents and returns an enriched RateSnapshot.
func (p *RatesProcessor) ProcessRawDocuments(docs []RawBankDocument, traceID string) *domain.RateSnapshot {
	// Initialize slice of enriched currency rates.
	rates := make([]domain.CurrencyRate, 0)
	// Initialize slice of bank availability records.
	availability := make([]domain.AvailabilityRecord, 0)
	// Initialize slice of operational warnings.
	warnings := make([]string, 0)

	// Now timestamp in UTC.
	now := time.Now().UTC()

	// Iterate over each bank document.
	for _, doc := range docs {
		// Normalize bank code to uppercase.
		bankCode := strings.ToUpper(strings.TrimSpace(doc.BankID))
		// If bank code is empty, skip document.
		if bankCode == "" {
			// Skip invalid doc.
			continue
		}

		// Look up bank metadata in catalog.
		bankMeta, found := p.catalog.FindBank(bankCode)
		// Default localized names and logo.
		nameAr := bankCode
		// Default English name.
		nameEn := bankCode
		// Default empty logo.
		logo := ""
		// If found in catalog, populate official metadata.
		if found {
			// Set official Arabic name.
			nameAr = bankMeta.NameAr
			// Set official English name.
			nameEn = bankMeta.NameEn
			// Set official logo URL.
			logo = bankMeta.Logo
		}

		// Determine document retrieval time.
		retrievedAt := now

		// Record bank operational availability.
		availStatus := "stale"
		// If status is SUCCESS, mark current.
		if strings.EqualFold(doc.Status, "SUCCESS") {
			// Set current.
			availStatus = "current"
		} else {
			// Add warning.
			warnings = append(warnings, fmt.Sprintf("Bank %s reported failure in upstream feed.", bankCode))
		}

		// Append availability record.
		availability = append(availability, domain.AvailabilityRecord{
			// Bank code.
			BankCode: bankCode,
			// Operational status.
			Status: availStatus,
			// Timestamp.
			LastSuccessAt: retrievedAt,
		})

		// Process each currency quote inside bank document.
		for currCode, info := range doc.Rates {
			// Normalize currency code to uppercase.
			currencyCode := strings.ToUpper(strings.TrimSpace(currCode))
			// Guard against empty currency codes.
			if currencyCode == "" {
				// Skip empty currency.
				continue
			}

			// Resolve old buy rate with fallback to yesterday's close if missing.
			var oldBuy *float64
			// Check if oldBuy explicitly provided.
			if info.OldBuy != nil {
				// Assign old buy.
				oldBuy = info.OldBuy
			} else if info.YesterdayCloseBuy != nil && *info.YesterdayCloseBuy != info.Buy {
				// Fallback to yesterday close buy.
				oldBuy = info.YesterdayCloseBuy
			}

			// Resolve old sell rate with fallback.
			var oldSell *float64
			// Check if oldSell explicitly provided.
			if info.OldSell != nil {
				// Assign old sell.
				oldSell = info.OldSell
			} else if info.YesterdayCloseSell != nil && *info.YesterdayCloseSell != info.Sell {
				// Fallback to yesterday close sell.
				oldSell = info.YesterdayCloseSell
			}

			// Determine directional trend ('up' or 'down', strictly never 'stable').
			trend := domain.TrendUp
			// Evaluate directional movement.
			if strings.EqualFold(info.Trend, "down") || strings.EqualFold(info.BuyTrend, "down") {
				// Explicit down trend.
				trend = domain.TrendDown
			} else if oldBuy != nil && info.Buy < *oldBuy {
				// Calculated price reduction.
				trend = domain.TrendDown
			} else if info.DailyHighBuy != nil && info.DailyLowBuy != nil && *info.DailyHighBuy != *info.DailyLowBuy && info.Buy == *info.DailyLowBuy {
				// Rate at session low implies down.
				trend = domain.TrendDown
			}

			// Construct enriched CurrencyRate item.
			rateItem := domain.CurrencyRate{
				// Localized Arabic bank name.
				NameAr: nameAr,
				// Localized English bank name.
				NameEn: nameEn,
				// Bank brand logo HTTPS URL.
				Logo: logo,
				// Normalized uppercase bank code.
				BankCode: bankCode,
				// Normalized uppercase ISO currency code.
				CurrencyCode: currencyCode,
				// Purchase exchange rate.
				BuyRate: info.Buy,
				// Selling exchange rate.
				SellRate: info.Sell,
				// Directional movement trend.
				Trend: trend,
				// Prior recorded purchase rate.
				OldBuyRate: oldBuy,
				// Prior recorded selling rate.
				OldSellRate: oldSell,
				// Ingestion timestamp.
				RetrievedAt: retrievedAt,
				// Temporal freshness marker.
				Freshness: "current",
				// Distributed request trace ID.
				SourceTraceID: traceID,
				// Daily session extremes.
				DailyHighBuy:       info.DailyHighBuy,
				DailyHighSell:      info.DailyHighSell,
				DailyLowBuy:        info.DailyLowBuy,
				DailyLowSell:       info.DailyLowSell,
				YesterdayCloseBuy:  info.YesterdayCloseBuy,
				YesterdayCloseSell: info.YesterdayCloseSell,
				// Prior close date string.
				LastCloseDate: info.LastCloseSnapshotDate,
			}

			// Append to rate list.
			rates = append(rates, rateItem)
		}
	}

	// Calculate consolidated snapshot freshness state.
	freshness := "current"
	// If warnings exist alongside valid rates, mark as partial.
	if len(warnings) > 0 && len(rates) > 0 {
		// Set partial freshness.
		freshness = "partial"
	} else if len(rates) == 0 {
		// Set stale freshness if empty.
		freshness = "stale"
	}

	// Construct and return final RateSnapshot envelope.
	return &domain.RateSnapshot{
		// Processed timestamp.
		RequestedAt: now,
		// Health status.
		Freshness: freshness,
		// Request filters.
		RequestFilters: make(map[string]string),
		// Enriched rates array.
		Rates: rates,
		// Availability summary.
		Availability: availability,
		// Warnings list.
		Warnings: warnings,
	}
}
