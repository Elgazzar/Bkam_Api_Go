package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	content := `
# Comment line
TEST_BKAM_PORT=3999
TEST_BKAM_USER="custom_user"
TEST_BKAM_HOST='custom_host'
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test .env: %v", err)
	}

	loadDotEnv(envPath)

	if os.Getenv("TEST_BKAM_PORT") != "3999" {
		t.Errorf("expected 3999, got %s", os.Getenv("TEST_BKAM_PORT"))
	}
	if os.Getenv("TEST_BKAM_USER") != "custom_user" {
		t.Errorf("expected custom_user, got %s", os.Getenv("TEST_BKAM_USER"))
	}
	if os.Getenv("TEST_BKAM_HOST") != "custom_host" {
		t.Errorf("expected custom_host, got %s", os.Getenv("TEST_BKAM_HOST"))
	}
}

func TestGetDatabaseConnString(t *testing.T) {
	cfg := &Config{
		DBUser:     "user1",
		DBPassword: "secret@password",
		DBHost:     "localhost",
		DBPort:     5432,
		DBName:     "mydb",
		DBSSLMode:  "disable",
	}

	connStr := cfg.GetDatabaseConnString()
	expected := "postgres://user1:secret%40password@localhost:5432/mydb?sslmode=disable"
	if connStr != expected {
		t.Errorf("expected %s, got %s", expected, connStr)
	}

	// URL precedence test
	cfg.DatabaseURL = "postgres://full_url:5432/mydb"
	if cfg.GetDatabaseConnString() != "postgres://full_url:5432/mydb" {
		t.Errorf("expected DatabaseURL precedence")
	}
}
