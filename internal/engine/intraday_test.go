// Package engine provides core domain business logic engines following Deep Module design principles.
package engine

// Import standard library testing and time packages.
import (
	// testing provides Go testing primitives.
	"testing"
	// time provides temporal fixtures.
	"time"

	// domain defines intraday data structures.
	"bkam-api/internal/domain"
)

// TestIntradayTimelineStoreStorage verifies caching and point normalization in IntradayTimelineStore.
func TestIntradayTimelineStoreStorage(t *testing.T) {
	// Construct MarketScheduleEngine.
	sched := NewMarketScheduleEngine()
	// Construct IntradayTimelineStore.
	store := NewIntradayTimelineStore(sched, nil, nil)

	// Fixture raw points for various assets.
	goldPoints := map[string]any{
		"21": []any{
			map[string]any{"time": "2026-10-01T10:00:00Z", "price": 3500.0, "buy": 3480.0, "sell": 3500.0},
			map[string]any{"time": "2026-10-01T11:00:00Z", "price": 3510.0, "buy": 3490.0, "sell": 3510.0},
		},
		"international": map[string]any{
			"elements": []any{
				map[string]any{"time": "2026-10-01T10:00:00Z", "price": 2650.5},
			},
		},
	}

	// Update intraday store for gold.
	store.UpdateIntraday("gold", "default", "2026-10-01", goldPoints)

	// Retrieve session.
	date, data, exists := store.GetIntraday("gold", "default")
	// Assert session existence.
	if !exists {
		// Report error.
		t.Fatalf("expected gold intraday session to exist")
	}
	// Assert session date.
	if date != "2026-10-01" {
		// Report error.
		t.Errorf("expected date 2026-10-01, got %s", date)
	}

	// Test symbol lookup for "21k" (alias resolution).
	points21 := store.LookupSymbolPoints(data, "21k")
	// Assert points length.
	if len(points21) != 2 {
		// Report error.
		t.Fatalf("expected 2 points for 21k, got %d", len(points21))
	}
	// Assert price on first point.
	if points21[0].Price != 3500.0 {
		// Report error.
		t.Errorf("expected price 3500.0, got %f", points21[0].Price)
	}
	// Assert buy pointer on first point.
	if points21[0].Buy == nil || *points21[0].Buy != 3480.0 {
		// Report error.
		t.Errorf("expected buy price 3480.0, got %v", points21[0].Buy)
	}

	// Test symbol lookup for "intl" alias.
	pointsIntl := store.LookupSymbolPoints(data, "intl")
	// Assert points length.
	if len(pointsIntl) != 1 {
		// Report error.
		t.Fatalf("expected 1 point for intl, got %d", len(pointsIntl))
	}
	// Assert price on intl point.
	if pointsIntl[0].Price != 2650.5 {
		// Report error.
		t.Errorf("expected price 2650.5, got %f", pointsIntl[0].Price)
	}
}

// TestIntradayTimelineStoreNormalization verifies robust parsing of varied point structures.
func TestIntradayTimelineStoreNormalization(t *testing.T) {
	// Construct IntradayTimelineStore.
	store := NewIntradayTimelineStore(NewMarketScheduleEngine(), nil, nil)

	// Test nil input.
	nilPoints := store.NormalizePoints(nil)
	// Assert empty slice.
	if len(nilPoints) != 0 {
		// Report error.
		t.Errorf("expected 0 points for nil, got %d", len(nilPoints))
	}

	// Test typed domain slice input.
	typedPoints := []domain.IntradayPoint{
		{Time: "2026-10-01T10:00:00Z", Price: 100.0},
	}
	// Normalize.
	res := store.NormalizePoints(typedPoints)
	// Assert length.
	if len(res) != 1 || res[0].Price != 100.0 {
		// Report error.
		t.Errorf("expected 1 point with price 100.0, got %v", res)
	}
}

// TestIntradaySessionDateResolution verifies zero-read trading session date calculation.
func TestIntradaySessionDateResolution(t *testing.T) {
	// Construct MarketScheduleEngine.
	sched := NewMarketScheduleEngine()
	// Construct IntradayTimelineStore.
	store := NewIntradayTimelineStore(sched, nil, nil)

	// Pre-seed an active session for today.
	todayCairo := sched.NowCairo().Format("2006-01-02")
	// Update store.
	store.UpdateIntraday("stocks", "egx", todayCairo, map[string]any{
		"COMI": []any{
			map[string]any{"time": time.Now().UTC().Format(time.RFC3339), "price": 85.5},
		},
	})

	// Resolve session date.
	resolvedDate, isToday := store.ResolveSessionDate("stocks", "egx")
	// Assert resolved date matches today.
	if resolvedDate != todayCairo {
		// Report error.
		t.Errorf("expected session date %s, got %s", todayCairo, resolvedDate)
	}
	// Assert isToday is true.
	if !isToday {
		// Report error.
		t.Errorf("expected isToday to be true")
	}
}
