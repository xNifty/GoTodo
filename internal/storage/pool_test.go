package storage

import (
	"strings"
	"testing"
)

func TestDatabaseDSNRequiresEnvVars(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	_, err := databaseDSN()
	if err == nil {
		t.Fatal("expected error when required DB env vars are missing")
	}
}

func TestDatabaseDSNDefaultsSSLModeToDisable(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "gotodo")
	t.Setenv("DB_SSLMODE", "")

	dsn, err := databaseDSN()
	if err != nil {
		t.Fatalf("databaseDSN failed: %v", err)
	}

	expected := "postgres://postgres:secret@localhost:5432/gotodo?sslmode=disable"
	if dsn != expected {
		t.Fatalf("expected %q, got %q", expected, dsn)
	}
}

func TestDatabaseDSNCustomSSLMode(t *testing.T) {
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "appuser")
	t.Setenv("DB_PASSWORD", "secretpass")
	t.Setenv("DB_NAME", "proddb")

	testCases := []struct {
		envVal   string
		expected string
	}{
		{"require", "sslmode=require"},
		{"true", "sslmode=require"},
		{"1", "sslmode=require"},
		{"disable", "sslmode=disable"},
		{"false", "sslmode=disable"},
		{"0", "sslmode=disable"},
		{"verify-full", "sslmode=verify-full"},
	}

	for _, tc := range testCases {
		t.Setenv("DB_SSLMODE", tc.envVal)
		dsn, err := databaseDSN()
		if err != nil {
			t.Fatalf("databaseDSN failed for %q: %v", tc.envVal, err)
		}
		if !strings.Contains(dsn, tc.expected) {
			t.Fatalf("for DB_SSLMODE=%q, expected %q in %q", tc.envVal, tc.expected, dsn)
		}
	}
}
