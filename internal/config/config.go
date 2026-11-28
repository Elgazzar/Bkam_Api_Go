// Package config provides strongly-typed environment configuration management for the Bkam API Gateway.
package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
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

	// DatabaseURL provides a complete PostgreSQL connection string (takes precedence).
	DatabaseURL string
	// DBUser specifies the PostgreSQL user account.
	DBUser string
	// DBPassword specifies the PostgreSQL password.
	DBPassword string
	// DBHost specifies the PostgreSQL host.
	DBHost string
	// DBPort specifies the PostgreSQL port.
	DBPort int
	// DBName specifies the PostgreSQL database name.
	DBName string
	// DBSSLMode specifies SSL mode (e.g. "disable", "require").
	DBSSLMode string
}

// GetDatabaseConnString resolves the PostgreSQL connection URL from environment variables.
func (c *Config) GetDatabaseConnString() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	user := url.QueryEscape(c.DBUser)
	pass := url.QueryEscape(c.DBPassword)

	auth := user
	if c.DBPassword != "" {
		auth = fmt.Sprintf("%s:%s", user, pass)
	}

	return fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=%s",
		auth, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

// loadDotEnv parses a .env file and sets environment variables if not already set.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}

// Load reads and parses all runtime configuration from environment variables with safe defaults.
func Load() *Config {
	loadDotEnv(".env")
	cfg := &Config{
		Port:        3001,
		Host:        "0.0.0.0",
		Environment: "development",
		SnapshotDir: "./var/snapshots",
		APIKey:      "local-dev-key",
		ServerID:    "server-go-01",

		DatabaseURL: "",
		DBUser:      "m7moudelgazzar",
		DBPassword:  "",
		DBHost:      "bkam_db",
		DBPort:      5432,
		DBName:      "bkam_db",
		DBSSLMode:   "disable",
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.Port = p
		}
	}

	if hostStr := os.Getenv("HOST"); hostStr != "" {
		cfg.Host = strings.TrimSpace(hostStr)
	}

	if envStr := os.Getenv("NODE_ENV"); envStr != "" {
		cfg.Environment = strings.TrimSpace(envStr)
	} else if envStr := os.Getenv("ENV"); envStr != "" {
		cfg.Environment = strings.TrimSpace(envStr)
	}

	if snapDir := os.Getenv("SNAPSHOT_DIR"); snapDir != "" {
		cfg.SnapshotDir = strings.TrimSpace(snapDir)
	}

	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		cfg.APIKey = strings.TrimSpace(apiKey)
	}

	if serverID := os.Getenv("SERVER_ID"); serverID != "" {
		cfg.ServerID = strings.TrimSpace(serverID)
	}

	// PostgreSQL Database configuration
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = strings.TrimSpace(dbURL)
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.DBUser = strings.TrimSpace(dbUser)
	}

	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.DBPassword = strings.TrimSpace(dbPass)
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.DBHost = strings.TrimSpace(dbHost)
	}

	if dbPortStr := os.Getenv("DB_PORT"); dbPortStr != "" {
		if p, err := strconv.Atoi(dbPortStr); err == nil && p > 0 {
			cfg.DBPort = p
		}
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.DBName = strings.TrimSpace(dbName)
	}

	if sslMode := os.Getenv("DB_SSLMODE"); sslMode != "" {
		cfg.DBSSLMode = strings.TrimSpace(sslMode)
	}

	// Ensure the snapshot directory exists on filesystem.
	if err := os.MkdirAll(cfg.SnapshotDir, 0755); err != nil {
		_ = err
	}

	return cfg
}
