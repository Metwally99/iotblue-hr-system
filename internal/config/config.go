package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	Environment        string
	HTTPAddress        string
	DatabaseURL        string
	JWTSecret          string
	JWTIssuer          string
	AccessTokenTTL     time.Duration
	Timezone           string
	CORSAllowedOrigins []string
	ShutdownTimeout    time.Duration
	ReadHeaderTimeout  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
}

func Load() (Config, error) {
	shutdownTimeout, err := duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	accessTokenTTL, err := duration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	timezone := value("APP_TIMEZONE", "Africa/Cairo")
	if _, err := time.LoadLocation(timezone); err != nil {
		return Config{}, errors.New("APP_TIMEZONE must be a valid IANA timezone")
	}

	return Config{
		Environment:        value("APP_ENV", "development"),
		HTTPAddress:        value("HTTP_ADDRESS", ":8080"),
		DatabaseURL:        databaseURL,
		JWTSecret:          jwtSecret,
		JWTIssuer:          value("JWT_ISSUER", "hr-management-api"),
		AccessTokenTTL:     accessTokenTTL,
		Timezone:           timezone,
		CORSAllowedOrigins: commaSeparated("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		ShutdownTimeout:    shutdownTimeout,
		ReadHeaderTimeout:  5 * time.Second,
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       30 * time.Second,
		IdleTimeout:        60 * time.Second,
	}, nil
}

func commaSeparated(key, fallback string) []string {
	raw := value(key, fallback)
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func value(key, fallback string) string {
	if result := os.Getenv(key); result != "" {
		return result
	}
	return fallback
}

func duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return 0, errors.New(key + " must be a valid duration")
	}
	return parsed, nil
}
