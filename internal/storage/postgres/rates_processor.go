// Package postgres provides PostgreSQL 17 persistence and real-time LISTEN/NOTIFY in-memory cache invalidation.
package postgres

import (
	"fmt"
	"strings"
	"time"

	"bkam-api/internal/domain"
)

// RawBankRateInfo reflects the raw rate fields stored inside a bank's rates map.
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

// RawBankDocument represents the root structure of a bank document.
type RawBankDocument struct {
	// BankID identifies the bank code (e.g. "nbe", "cib").
	BankID string `json:"bankId"`
	// BankCode is alternative bank identifier field.
	BankCode string `json:"bankCode,omitempty"`
	// BankNameAr is optional Arabic name from scraper.
	BankNameAr string `json:"bankNameAr,omitempty"`
	// BankNameEn is optional English name from scraper.
	BankNameEn string `json:"bankNameEn,omitempty"`
	// Timestamp records document creation or scraping time.
	Timestamp any `json:"timestamp,omitempty"`
	// BankUpdatedAt records when bank published rates.
	BankUpdatedAt any `json:"bankUpdatedAt,omitempty"`
	// Status indicates scraping status ("SUCCESS", "FAILED").
	Status string `json:"status,omitempty"`
	// Rates maps currency codes (e.g. "USD", "EUR") to rate details.
	Rates map[string]RawBankRateInfo `json:"rates"`
}

// RatesProcessor transforms raw bank documents into enriched, validated RateSnapshot models.
type RatesProcessor struct {
	catalog *domain.CatalogRegistry
}

// NewRatesProcessor constructs a new RatesProcessor instance.
func NewRatesProcessor() *RatesProcessor {
	return &RatesProcessor{
		catalog: domain.GetCatalogRegistry(),
	}
}

// ProcessRawDocuments processes a collection of raw bank documents and returns an enriched RateSnapshot.
func (p *RatesProcessor) ProcessRawDocuments(docs []RawBankDocument, traceID string) *domain.RateSnapshot {
	rates := make([]domain.CurrencyRate, 0)
	availability := make([]domain.AvailabilityRecord, 0)
	warnings := make([]string, 0)
	now := time.Now().UTC()

	for _, doc := range docs {
		rawCode := doc.BankID
		if rawCode == "" {
			rawCode = doc.BankCode
		}
		bankCode := strings.ToUpper(strings.TrimSpace(rawCode))
		if bankCode == "" {
			continue
		}

		bankMeta, found := p.catalog.FindBank(bankCode)
		nameAr := bankCode
		nameEn := bankCode
		logo := ""

		if doc.BankNameAr != "" {
			nameAr = doc.BankNameAr
		}
		if doc.BankNameEn != "" {
			nameEn = doc.BankNameEn
		}

		if found {
			if doc.BankNameAr == "" {
				nameAr = bankMeta.NameAr
			}
			if doc.BankNameEn == "" {
				nameEn = bankMeta.NameEn
			}
			logo = bankMeta.Logo
		}

		retrievedAt := now
		availStatus := "stale"
		if doc.Status == "" || strings.EqualFold(doc.Status, "SUCCESS") {
			availStatus = "current"
		} else {
			warnings = append(warnings, fmt.Sprintf("Bank %s reported failure in upstream feed.", bankCode))
		}

		availability = append(availability, domain.AvailabilityRecord{
			BankCode:      bankCode,
			Status:        availStatus,
			LastSuccessAt: retrievedAt,
		})

		for currCode, info := range doc.Rates {
			currencyCode := strings.ToUpper(strings.TrimSpace(currCode))
			if currencyCode == "" {
				continue
			}

			var oldBuy *float64
			if info.OldBuy != nil {
				oldBuy = info.OldBuy
			} else if info.YesterdayCloseBuy != nil && *info.YesterdayCloseBuy != info.Buy {
				oldBuy = info.YesterdayCloseBuy
			}

			var oldSell *float64
			if info.OldSell != nil {
				oldSell = info.OldSell
			} else if info.YesterdayCloseSell != nil && *info.YesterdayCloseSell != info.Sell {
				oldSell = info.YesterdayCloseSell
			}

			trend := domain.TrendUp
			if oldBuy != nil {
				if info.Buy < *oldBuy {
					trend = domain.TrendDown
				} else if info.Buy > *oldBuy {
					trend = domain.TrendUp
				}
			} else if strings.EqualFold(info.Trend, "down") || strings.EqualFold(info.BuyTrend, "down") {
				trend = domain.TrendDown
			}

			rate := domain.CurrencyRate{
				NameAr:             nameAr,
				NameEn:             nameEn,
				Logo:               logo,
				BankCode:           bankCode,
				CurrencyCode:       currencyCode,
				BuyRate:            info.Buy,
				SellRate:           info.Sell,
				Trend:              trend,
				OldBuyRate:         oldBuy,
				OldSellRate:        oldSell,
				RetrievedAt:        retrievedAt,
				Freshness:          availStatus,
				SourceTraceID:      traceID,
				DailyHighBuy:       info.DailyHighBuy,
				DailyHighSell:      info.DailyHighSell,
				DailyLowBuy:        info.DailyLowBuy,
				DailyLowSell:       info.DailyLowSell,
				YesterdayCloseBuy:  info.YesterdayCloseBuy,
				YesterdayCloseSell: info.YesterdayCloseSell,
				LastCloseDate:      info.LastCloseSnapshotDate,
			}

			rates = append(rates, rate)
		}
	}

	freshness := "stale"
	currentCount := 0
	for _, a := range availability {
		if a.Status == "current" {
			currentCount++
		}
	}
	if currentCount == len(availability) && len(availability) > 0 {
		freshness = "current"
	} else if currentCount > 0 {
		freshness = "partial"
	}

	return &domain.RateSnapshot{
		RequestedAt:    now,
		Freshness:      freshness,
		RequestFilters: map[string]string{},
		Rates:          rates,
		Availability:   availability,
		Warnings:       warnings,
	}
}
