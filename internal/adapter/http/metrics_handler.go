// Package http provides inbound HTTP routing, handlers, middleware, and error responders.
package http

// Import standard library packages for HTTP interfaces, synchronization, and strconv.
import (
	// net/http provides response and request primitives.
	"net/http"
	// strconv parses integer query parameters.
	"strconv"
	// strings provides string trimming.
	"strings"
	// sync provides mutex synchronization for concurrent metric counters.
	"sync"

	// domain defines metrics payload models.
	"bkam-api/internal/domain"
	// engine provides market schedule calculations.
	"bkam-api/internal/engine"
)

// MetricsHandler coordinates request analytics tracking and serves GET /v1/metrics.
type MetricsHandler struct {
	// mu synchronizes access to the in-memory metric counters.
	mu sync.RWMutex
	// schedule provides Cairo timezone conversions.
	schedule *engine.MarketScheduleEngine
	// todayCounts stores request counts for the current Cairo calendar date.
	todayCounts domain.CategoryCounts
	// todayTotal stores total requests for the current Cairo date.
	todayTotal int
	// currentDate stores the active Cairo date string (YYYY-MM-DD).
	currentDate string
}

// NewMetricsHandler constructs a new MetricsHandler instance.
func NewMetricsHandler(schedule *engine.MarketScheduleEngine) *MetricsHandler {
	// Calculate current Cairo date.
	nowCairo := schedule.NowCairo().Format("2006-01-02")
	// Return initialized handler.
	return &MetricsHandler{
		// Store schedule engine.
		schedule: schedule,
		// Initialize date.
		currentDate: nowCairo,
	}
}

// RecordRequest increments the request counter for a specified financial domain category.
func (h *MetricsHandler) RecordRequest(category string) {
	// Acquire write lock.
	h.mu.Lock()
	// Defer release of write lock.
	defer h.mu.Unlock()

	// Check if Cairo date has rolled over.
	todayCairo := h.schedule.NowCairo().Format("2006-01-02")
	if todayCairo != h.currentDate {
		// Reset counters for new day.
		h.currentDate = todayCairo
		h.todayTotal = 0
		h.todayCounts = domain.CategoryCounts{}
	}

	// Increment total counter.
	h.todayTotal++

	// Increment specific domain category counter.
	switch strings.ToLower(category) {
	case "banks", "rates":
		h.todayCounts.Banks++
	case "gold":
		h.todayCounts.Gold++
	case "stocks":
		h.todayCounts.Stocks++
	case "indices":
		h.todayCounts.Indices++
	case "gas":
		h.todayCounts.Gas++
	default:
		h.todayCounts.Other++
	}
}

// GetMetrics handles GET /v1/metrics with optional ?days= lookback parameter.
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// Extract days query parameter.
	daysStr := strings.TrimSpace(r.URL.Query().Get("days"))
	// Default lookback duration is 1 day (today).
	days := 1

	// If days parameter provided, parse and validate.
	if daysStr != "" {
		// Parse integer.
		parsedDays, err := strconv.Atoi(daysStr)
		// Check for parse error or range bounds violation (1 to 30).
		if err != nil || parsedDays < 1 || parsedDays > 30 {
			// Reject invalid lookback duration matching Zod validation contract.
			RespondError(w, http.StatusBadRequest, "INVALID_QUERY", "Days parameter cannot exceed the 30-day retention window.", []any{
				map[string]any{
					"field":   "days",
					"message": "Days parameter cannot exceed the 30-day retention window.",
				},
			})
			// Return early.
			return
		}
		// Assign parsed value.
		days = parsedDays
	}

	// Acquire read lock to snapshot current metrics.
	h.mu.RLock()
	// Copy date.
	dateStr := h.currentDate
	// Copy total.
	totalReqs := h.todayTotal
	// Copy categories.
	counts := h.todayCounts
	// Release read lock.
	h.mu.RUnlock()

	// Construct today's metrics record.
	todayMetrics := domain.DailyMetrics{
		Date:       dateStr,
		Total:      totalReqs,
		Categories: counts,
	}

	// Construct metrics response payload.
	responsePayload := domain.MetricsResponse{
		ServerID: "bkam-api-go",
		Date:     dateStr,
		Today:    todayMetrics,
	}

	// If history requested (days > 1), provide history array.
	if days > 1 {
		// Pre-populate empty history array matching contract.
		responsePayload.History = make([]domain.DailyMetrics, 0)
	}

	// Transmit JSON metrics response.
	RespondJSON(w, http.StatusOK, responsePayload)
}
