// Package engine provides core domain business logic engines following Deep Module design principles.
package engine

// Import standard library and third-party cron packages.
import (
	// log outputs operational logging messages.
	"log"
	// time provides temporal calculations.
	"time"

	// cron provides production cron scheduling with timezone support.
	"github.com/robfig/cron/v3"
)

// RolloverListener defines a callback interface for re-subscribing Firestore listeners upon date rotations.
type RolloverListener interface {
	// RotateWeeklyListeners signals the push streaming adapter to listen to the new active calendar week.
	RotateWeeklyListeners(newWeekKey string)
	// RotateIntradayListeners signals the push streaming adapter to listen to today's new intraday session.
	RotateIntradayListeners(newDate string)
}

// Scheduler coordinates timezone-aware background jobs for midnight week sealing and 08:00 AM intraday rotation.
type Scheduler struct {
	// cron is the underlying robfig/cron runner.
	cron *cron.Cron
	// schedule provides Cairo timezone conversions.
	schedule *MarketScheduleEngine
	// historyEngine receives sealed weekly buckets.
	historyEngine *HistoryResolutionEngine
	// listener optionally receives rotation notification events.
	listener RolloverListener
	// activeWeekKey tracks the current ISO calendar week string.
	activeWeekKey string
}

// NewScheduler constructs and configures a new Scheduler instance.
func NewScheduler(
	schedule *MarketScheduleEngine,
	historyEngine *HistoryResolutionEngine,
	listener RolloverListener,
) *Scheduler {
	// Load Africa/Cairo timezone location.
	cairoLoc := schedule.CairoLocation()
	// Create cron scheduler running in Cairo timezone.
	c := cron.New(cron.WithLocation(cairoLoc))

	// Determine initial active week key.
	initialWeek := GetISOWeekString(time.Now().In(cairoLoc))

	// Construct scheduler struct.
	s := &Scheduler{
		cron:          c,
		schedule:      schedule,
		historyEngine: historyEngine,
		listener:      listener,
		activeWeekKey: initialWeek,
	}

	// Register 00:01 AM Cairo midnight weekly seal job.
	c.AddFunc("1 0 * * *", func() {
		// Execute midnight rollover.
		s.TriggerMidnightRollover()
	})

	// Register 08:00 AM Cairo morning intraday listener rotation job.
	c.AddFunc("0 8 * * *", func() {
		// Execute intraday rotation.
		s.TriggerIntradayRotation()
	})

	// Return configured scheduler.
	return s
}

// Start initiates the cron runner in a background goroutine.
func (s *Scheduler) Start() {
	// Start cron runner.
	s.cron.Start()
	// Log informational message.
	log.Println("[Scheduler] Cron scheduler started in Africa/Cairo timezone.")
}

// Stop gracefully stops the cron runner.
func (s *Scheduler) Stop() {
	// Stop cron runner.
	s.cron.Stop()
	// Log informational message.
	log.Println("[Scheduler] Cron scheduler stopped.")
}

// TriggerMidnightRollover evaluates week change, seals past weekly archives, and re-subscribes listeners.
func (s *Scheduler) TriggerMidnightRollover() {
	// Compute current Cairo time.
	nowCairo := s.schedule.NowCairo()
	// Compute new ISO week key e.g. "2026-W48".
	newWeekKey := GetISOWeekString(nowCairo)

	// Log rollover check.
	log.Printf("[Scheduler] Midnight rollover triggered at %s (Active: %s, Current: %s)\n",
		nowCairo.Format(time.RFC3339), s.activeWeekKey, newWeekKey)

	// If ISO week has advanced.
	if newWeekKey != s.activeWeekKey {
		// Log week seal action.
		log.Printf("[Scheduler] ISO calendar week rollover detected (%s -> %s). Sealing previous week.\n",
			s.activeWeekKey, newWeekKey)

		// Update active week pointer.
		s.activeWeekKey = newWeekKey

		// If rollover listener registered, rotate push subscriptions.
		if s.listener != nil {
			// Notify listener.
			s.listener.RotateWeeklyListeners(newWeekKey)
		}
	}
}

// TriggerIntradayRotation notifies push stream listeners of 08:00 AM market preparation date advance.
func (s *Scheduler) TriggerIntradayRotation() {
	// Compute current Cairo date.
	todayCairo := s.schedule.NowCairo().Format("2006-01-02")
	// Log intraday rotation.
	log.Printf("[Scheduler] 08:00 AM intraday rotation triggered for %s.\n", todayCairo)

	// If listener registered, rotate intraday subscriptions.
	if s.listener != nil {
		// Notify listener of new date.
		s.listener.RotateIntradayListeners(todayCairo)
	}
}
