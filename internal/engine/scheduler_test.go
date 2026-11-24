// Package engine provides core domain business logic engines following Deep Module design principles.
package engine

// Import standard library packages for testing and synchronization.
import (
	// testing provides Go unit test fixtures.
	"testing"
)

// mockRolloverListener implements RolloverListener for testing callback invocations.
type mockRolloverListener struct {
	// rotatedWeek records the week string received by RotateWeeklyListeners.
	rotatedWeek string
	// rotatedDate records the date string received by RotateIntradayListeners.
	rotatedDate string
}

// RotateWeeklyListeners records the new week key callback.
func (m *mockRolloverListener) RotateWeeklyListeners(newWeekKey string) {
	// Assign week.
	m.rotatedWeek = newWeekKey
}

// RotateIntradayListeners records the new date string callback.
func (m *mockRolloverListener) RotateIntradayListeners(newDate string) {
	// Assign date.
	m.rotatedDate = newDate
}

// TestSchedulerMidnightAndMorningTriggers verifies scheduler lifecycle and callback executions.
func TestSchedulerMidnightAndMorningTriggers(t *testing.T) {
	// Construct MarketScheduleEngine.
	sched := NewMarketScheduleEngine()
	// Mock listener.
	mock := &mockRolloverListener{}

	// Construct Scheduler.
	scheduler := NewScheduler(sched, nil, mock)

	// Verify initial active week is non-empty.
	if scheduler.activeWeekKey == "" {
		t.Fatalf("expected non-empty active week key")
	}

	// Start scheduler.
	scheduler.Start()
	// Stop scheduler immediately.
	scheduler.Stop()

	// Trigger intraday rotation.
	scheduler.TriggerIntradayRotation()
	// Assert rotated date matches today Cairo.
	expectedDate := sched.NowCairo().Format("2006-01-02")
	if mock.rotatedDate != expectedDate {
		t.Errorf("expected intraday rotation date %s, got %s", expectedDate, mock.rotatedDate)
	}

	// Simulate week advancement.
	scheduler.activeWeekKey = "2026-W01"
	// Trigger midnight rollover.
	scheduler.TriggerMidnightRollover()
	// Assert rotated week callback was invoked with new week.
	if mock.rotatedWeek == "" || mock.rotatedWeek == "2026-W01" {
		t.Errorf("expected rotated week to advance, got %s", mock.rotatedWeek)
	}
}
