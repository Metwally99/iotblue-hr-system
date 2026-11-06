package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mahmoudmetwally/hr-management-system/internal/model"
	"github.com/mahmoudmetwally/hr-management-system/internal/platform/database"
	"github.com/mahmoudmetwally/hr-management-system/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	databaseURL := required(logger, "DATABASE_URL")
	name := required(logger, "ADMIN_NAME")
	email := strings.ToLower(strings.TrimSpace(required(logger, "ADMIN_EMAIL")))
	password := required(logger, "ADMIN_PASSWORD")
	department := value("ADMIN_DEPARTMENT", "Administration")
	leaveDays := intValue(logger, "ADMIN_LEAVE_DAYS", 20)
	if len(password) < 8 || len(password) > 72 {
		logger.Error("ADMIN_PASSWORD must contain 8 to 72 bytes")
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("hash admin password", "error", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	created, err := repository.NewEmployeeRepository(pool).BootstrapAdmin(ctx, model.BootstrapAdminParams{
		DepartmentName: department, Name: strings.TrimSpace(name), Email: email, PasswordHash: string(hash),
		LeaveYear: time.Now().Year(), TotalLeaveDays: leaveDays,
	})
	if err != nil {
		logger.Error("bootstrap administrator", "error", err)
		os.Exit(1)
	}
	if created {
		logger.Info("administrator created", "email", email)
	} else {
		logger.Info("bootstrap skipped because employees already exist")
	}
}

func required(logger *slog.Logger, key string) string {
	result := strings.TrimSpace(os.Getenv(key))
	if result == "" {
		logger.Error(key + " is required")
		os.Exit(1)
	}
	return result
}

func value(key, fallback string) string {
	if result := strings.TrimSpace(os.Getenv(key)); result != "" {
		return result
	}
	return fallback
}

func intValue(logger *slog.Logger, key string, fallback int) int {
	raw := value(key, strconv.Itoa(fallback))
	result, err := strconv.Atoi(raw)
	if err != nil || result < 1 || result > 365 {
		logger.Error(key + " must be an integer between 1 and 365")
		os.Exit(1)
	}
	return result
}
