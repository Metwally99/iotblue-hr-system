package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mahmoudmetwally/hr-management-system/internal/config"
	"github.com/mahmoudmetwally/hr-management-system/internal/handler"
	"github.com/mahmoudmetwally/hr-management-system/internal/platform/database"
	"github.com/mahmoudmetwally/hr-management-system/internal/repository"
	"github.com/mahmoudmetwally/hr-management-system/internal/security"
	"github.com/mahmoudmetwally/hr-management-system/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tokenManager, err := security.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTokenTTL)
	if err != nil {
		logger.Error("configure token manager", "error", err)
		os.Exit(1)
	}
	employeeRepository := repository.NewEmployeeRepository(pool)
	departmentRepository := repository.NewDepartmentRepository(pool)
	authService := service.NewAuthService(employeeRepository, tokenManager)
	departmentService := service.NewDepartmentService(departmentRepository)
	employeeService := service.NewEmployeeService(employeeRepository)
	leaveRepository := repository.NewLeaveRepository(pool)
	leaveService := service.NewLeaveService(leaveRepository)
	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		logger.Error("load company timezone", "error", err)
		os.Exit(1)
	}
	attendanceRepository := repository.NewAttendanceRepository(pool)
	attendanceService := service.NewAttendanceService(attendanceRepository, location)

	router := handler.NewRouter(handler.RouterDependencies{
		Logger: logger, Database: pool, Auth: authService, Departments: departmentService,
		Employees: employeeService, Leave: leaveService, Attendance: attendanceService,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
	})
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("api listening", "address", cfg.HTTPAddress, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err = <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped", "error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("api stopped")
}
