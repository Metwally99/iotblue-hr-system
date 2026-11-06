package config

import (
	"testing"
	"time"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "a-secret-that-is-at-least-32-characters")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing DATABASE_URL to fail")
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "a-secret-that-is-at-least-32-characters")
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDRESS", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Environment != "development" {
		t.Fatalf("Environment = %q", cfg.Environment)
	}
	if cfg.HTTPAddress != ":8080" {
		t.Fatalf("HTTPAddress = %q", cfg.HTTPAddress)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %v", cfg.ShutdownTimeout)
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("AccessTokenTTL = %v", cfg.AccessTokenTTL)
	}
}

func TestLoadParsesCORSOrigins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "this-is-a-long-test-secret-with-32-chars")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://hr.example.com ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 CORS origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[1] != "https://hr.example.com" {
		t.Fatalf("expected whitespace to be trimmed, got %q", cfg.CORSAllowedOrigins[1])
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "a-secret-that-is-at-least-32-characters")
	t.Setenv("SHUTDOWN_TIMEOUT", "eventually")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid SHUTDOWN_TIMEOUT to fail")
	}
}

func TestLoadRequiresStrongJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "too-short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected short JWT_SECRET to fail")
	}
}

func TestLoadRejectsInvalidTimezone(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "a-secret-that-is-at-least-32-characters")
	t.Setenv("APP_TIMEZONE", "Moon/Sea-of-Tranquility")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid APP_TIMEZONE to fail")
	}
}
