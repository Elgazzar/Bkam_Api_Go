// Package config provides strongly-typed environment configuration management for the Bkam API Gateway.
package config

// Import standard library packages for environment variables and string handling.
import (
	// os provides access to operating system environment variables.
	"os"
	// strconv allows converting string configuration values into integers or booleans.
	"strconv"
	// strings provides string inspection and manipulation utilities.
	"strings"
)

// Config represents the validated runtime configuration for the API Gateway service.
type Config struct {
	// Port specifies the TCP port on which the HTTP server will listen (e.g. 3001).
	Port int
	// Host specifies the network interface address to bind to (e.g. "0.0.0.0").
	Host string
	// Environment indicates runtime tier ("development", "test", "production").
	Environment string
	// SnapshotDir points to the local directory used for disk snapshot and historical cache persistence.
	SnapshotDir string
	// APIKey defines the shared secret expected in x-api-key headers or ?apiKey= query params.
	APIKey string
	// ServerID identifies this server instance within distributed logs and response headers.
	ServerID string
	// ProjectID specifies the Google Cloud Firestore project identifier.
	ProjectID string
	// EmulatorHost optionally points to the Firestore local emulator (e.g. "127.0.0.1:8080").
	EmulatorHost string
}

// Load reads and parses all runtime configuration from environment variables with safe defaults.
func Load() *Config {
	// Initialize default configuration values matching platform operational expectations.
	cfg := &Config{
		// Default HTTP listening port is 3001 to maintain compatibility with existing tests.
		Port: 3001,
		// Default host binds to all available network interfaces.
		Host: "0.0.0.0",
		// Default environment is development.
		Environment: "development",
		// Default snapshot persistence directory is ./var/snapshots.
		SnapshotDir: "./var/snapshots",
		// Default local development API key matching test fixtures.
		APIKey: "local-dev-key",
		// Default server identifier for tracing and telemetry.
		ServerID: "server-go-01",
		// Default Google Cloud Firestore project identifier.
		ProjectID: "bkam-api",
		// Default emulator host is empty unless specified by the environment.
		EmulatorHost: "",
	}

	// Read PORT environment variable override if present.
	if portStr := os.Getenv("PORT"); portStr != "" {
		// Attempt parsing integer value from the port string.
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			// Assign valid parsed port number.
			cfg.Port = p
		}
	}

	// Read HOST environment variable override if present.
	if hostStr := os.Getenv("HOST"); hostStr != "" {
		// Assign customized host address.
		cfg.Host = strings.TrimSpace(hostStr)
	}

	// Read NODE_ENV or ENV environment variable override.
	if envStr := os.Getenv("NODE_ENV"); envStr != "" {
		// Assign specified runtime environment.
		cfg.Environment = strings.TrimSpace(envStr)
	} else if envStr := os.Getenv("ENV"); envStr != "" {
		// Fallback check for standard ENV variable.
		cfg.Environment = strings.TrimSpace(envStr)
	}

	// Read SNAPSHOT_DIR environment variable override if present.
	if snapDir := os.Getenv("SNAPSHOT_DIR"); snapDir != "" {
		// Assign customized persistent snapshot directory path.
		cfg.SnapshotDir = strings.TrimSpace(snapDir)
	}

	// Read API_KEY environment variable override if present.
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		// Assign custom authorized API key.
		cfg.APIKey = strings.TrimSpace(apiKey)
	}

	// Read SERVER_ID environment variable override if present.
	if serverID := os.Getenv("SERVER_ID"); serverID != "" {
		// Assign custom server instance identifier.
		cfg.ServerID = strings.TrimSpace(serverID)
	}

	// Read FIRESTORE_PROJECT_ID or GOOGLE_CLOUD_PROJECT override if present.
	if projID := os.Getenv("FIRESTORE_PROJECT_ID"); projID != "" {
		// Assign explicit Firestore project identifier.
		cfg.ProjectID = strings.TrimSpace(projID)
	} else if gcpProj := os.Getenv("GOOGLE_CLOUD_PROJECT"); gcpProj != "" {
		// Assign standard Google Cloud project identifier.
		cfg.ProjectID = strings.TrimSpace(gcpProj)
	}

	// Read FIRESTORE_EMULATOR_HOST environment variable if running against emulator.
	if emuHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); emuHost != "" {
		// Assign active local emulator host address.
		cfg.EmulatorHost = strings.TrimSpace(emuHost)
	}

	// Ensure the snapshot directory exists on filesystem.
	if err := os.MkdirAll(cfg.SnapshotDir, 0755); err != nil {
		// If creation fails, ignore error for now as runtime will handle missing paths.
		_ = err
	}

	// Return the populated and validated configuration pointer.
	return cfg
}
