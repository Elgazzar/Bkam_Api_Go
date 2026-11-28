// Package main is the entrypoint for the Bkam Financial API Gateway service.
package main

// Import standard library packages for signal trapping, context cancellation, HTTP serving, and logging.
import (
	// context provides root context cancellation for background workers.
	"context"
	// errors checks error conditions.
	"errors"
	// fmt formats console output strings.
	"fmt"
	// log outputs structured log entries.
	"log"
	// net/http provides HTTP server implementation.
	"net/http"
	// os provides access to operating system signals.
	"os"
	// os/signal traps SIGINT and SIGTERM for graceful shutdown.
	"os/signal"
	// syscall defines standard POSIX shutdown signals.
	"syscall"
	// time provides shutdown timeouts and temporal calculations.
	"time"

	// adapterhttp provides inbound HTTP handlers, middleware, and routing.
	adapterhttp "bkam-api/internal/adapter/http"
	// adapterstorage provides filesystem persistence for snapshots and historical weekly buckets.
	adapterstorage "bkam-api/internal/adapter/storage"
	// config loads validated runtime configuration from environment variables.
	"bkam-api/internal/config"
	// engine provides domain business logic engines.
	"bkam-api/internal/engine"
	// postgres provides PostgreSQL 17 persistence and LISTEN/NOTIFY invalidation.
	"bkam-api/internal/storage/postgres"
)

// main is the application entry point.
func main() {
	// Load configuration from environment variables.
	cfg := config.Load()
	// Log startup greeting.
	log.Printf("[Main] Starting Bkam Financial API Gateway (Go) in %s mode on %s:%d\n",
		cfg.Environment, cfg.Host, cfg.Port)

	// Create root application context with cancellation.
	rootCtx, rootCancel := context.WithCancel(context.Background())
	// Defer cancellation on exit.
	defer rootCancel()

	// 1. Initialize persistent disk storage adapter.
	diskAdapter, err := adapterstorage.NewDiskAdapter(cfg.SnapshotDir)
	// Check for storage initialization errors.
	if err != nil {
		// Log fatal error.
		log.Fatalf("[Main] Failed to initialize disk adapter: %v\n", err)
	}

	// 2. Initialize Cairo market schedule and trading calendar engine.
	marketSchedule := engine.NewMarketScheduleEngine()

	// 3. Initialize LiveSnapshotCoordinator with disk persistence adapter.
	coordinator := engine.NewLiveSnapshotCoordinator(diskAdapter)
	// Hydrate in-memory cache from cold boot persistent disk snapshots.
	coordinator.HydrateFromDisk()

	// 4. Initialize HistoryResolutionEngine for multi-tier historical queries.
	historyEngine := engine.NewHistoryResolutionEngine(diskAdapter, coordinator, marketSchedule)

	// 5. Initialize IntradayTimelineStore for sparkline caching and off-hours session backtracking.
	intradayStore := engine.NewIntradayTimelineStore(marketSchedule, coordinator, historyEngine)

	// 6. Connect to PostgreSQL (or run in cached offline mode if database unconfigured/offline).
	var pgStorage *postgres.Storage
	storage, err := postgres.NewStorage(rootCtx, cfg, coordinator)
	if err == nil {
		pgStorage = storage
		defer storage.Stop()
		log.Printf("[Main] Connected to PostgreSQL 17 at %s:%d/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)

		// Hydrate in-memory cache from PostgreSQL live_snapshots table.
		if err := pgStorage.Hydrate(rootCtx); err != nil {
			log.Printf("[Main] Warning: PostgreSQL hydration error: %v\n", err)
		}

		// Start PostgreSQL LISTEN snapshot_update background invalidation listener.
		pgStorage.StartListener(rootCtx)
	} else {
		log.Printf("[Main] Warning: Could not connect to PostgreSQL (%v). Running in cached offline mode.\n", err)
	}

	// 7. Initialize and start Cairo timezone cron scheduler.
	scheduler := engine.NewScheduler(marketSchedule, historyEngine, nil)
	// Start scheduler.
	scheduler.Start()

	// 8. Construct HTTP REST domain handlers.
	ratesHandler := adapterhttp.NewRatesHandler(coordinator, historyEngine)
	ratesHandler.SetIntradayStore(intradayStore)
	commoditiesHandler := adapterhttp.NewCommoditiesHandler(coordinator, historyEngine)
	commoditiesHandler.SetIntradayStore(intradayStore)
	stockHandler := adapterhttp.NewStockHandler(coordinator, historyEngine)
	stockHandler.SetIntradayStore(intradayStore)
	indicesHandler := adapterhttp.NewIndicesHandler(coordinator, historyEngine)
	quotesHandler := adapterhttp.NewQuotesHandler(coordinator)
	healthHandler := adapterhttp.NewHealthHandler(coordinator)
	versionHandler := adapterhttp.NewVersionHandler(cfg.Environment)
	metricsHandler := adapterhttp.NewMetricsHandler(marketSchedule)

	// 9. Assemble Chi router with platform routes and security middleware.
	router := adapterhttp.NewRouter(&adapterhttp.RouterDependencies{
		APIKey:             cfg.APIKey,
		Coordinator:        coordinator,
		MarketSchedule:     marketSchedule,
		QuotesHandler:      quotesHandler,
		RatesHandler:       ratesHandler,
		CommoditiesHandler: commoditiesHandler,
		StockHandler:       stockHandler,
		IndicesHandler:     indicesHandler,
		HealthHandler:      healthHandler,
		VersionHandler:     versionHandler,
		MetricsHandler:     metricsHandler,
	})

	// 10. Configure standard HTTP server.
	serverAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	// Allocate HTTP server.
	server := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 11. Start HTTP server in background goroutine.
	go func() {
		// Log listening status.
		log.Printf("[Main] HTTP server listening on http://%s\n", serverAddr)
		// Begin accepting incoming connections.
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Fatal error if unexpected exit.
			log.Fatalf("[Main] HTTP server failed: %v\n", err)
		}
	}()

	// 12. Trap OS signals for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	// Notify on SIGINT and SIGTERM.
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	// Block until signal received.
	sig := <-sigChan
	// Log shutdown trigger.
	log.Printf("[Main] Received signal %v. Initiating graceful shutdown...\n", sig)

	// 13. Stop cron scheduler.
	scheduler.Stop()

	// 14. Stop background PostgreSQL listener and close connection pool.
	if pgStorage != nil {
		pgStorage.Stop()
	}

	// 15. Stop HTTP server with 5-second graceful drain timeout.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	// Defer cancel.
	defer shutdownCancel()
	// Shutdown server.
	if err := server.Shutdown(shutdownCtx); err != nil {
		// Log shutdown error.
		log.Printf("[Main] HTTP server shutdown error: %v\n", err)
	}

	// Log clean shutdown completion.
	log.Println("[Main] Bkam Financial API Gateway shutdown completed cleanly.")
}
