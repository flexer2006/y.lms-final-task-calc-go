// Package main реализует точку входа службы аутентификации.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/config"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/migrate"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/shutdown"
	"go.uber.org/zap"
)

// Константы для сообщений об ошибках.
const (
	ErrInitLogger     = "failed to initialize logger"
	ErrSyncLogger     = "failed to sync logger"
	ErrLoadConfig     = "failed to load configuration"
	ErrInitDB         = "failed to initialize database"
	ErrRunMigrations  = "failed to run migrations"
	ErrInitGRPCServer = "failed to initialize gRPC server"
	ErrStartGRPC      = "failed to start gRPC server"
)

// Константы для игнорируемых ошибок.
const (
	ErrSyncStderr = "sync /dev/stderr: invalid argument"
	ErrSyncStdout = "sync /dev/stdout: invalid argument"
)

// Константы для сообщений сервиса.
const (
	LogServiceStarted      = "authentication service started"
	LogServiceShutdownDone = "authentication service shutdown complete"
	LogLoadingConfig       = "loading configuration"
	LogConfigLoaded        = "configuration loaded"
	LogInitDB              = "initializing database connection"
	LogDBInitialized       = "database connection established"
	LogRunMigrations       = "running database migrations"
	LogMigrationsCompleted = "database migrations completed"
	LogClosingDB           = "closing database connections"
)

func main() {
	log, err := logger.Development()
	if err != nil {
		panic(fmt.Sprintf("%s: %v", ErrInitLogger, err))
	}

	ctx := context.Background()
	ctx, requestID := logger.EnsureRequestID(ctx)
	ctx = logger.WithLogger(ctx, log)

	var exitCode int
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	defer func() {
		if err := logger.Sync(ctx, log); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, ErrSyncStderr) || strings.Contains(errMsg, ErrSyncStdout) {
				return
			}
			fmt.Fprintf(os.Stderr, "%s: %v\n", ErrSyncLogger, err)
		}
	}()

	logger.Info(ctx, log, LogServiceStarted,
		zap.String("request_id", requestID),
		zap.String("startup_time", time.Now().Format(time.RFC3339)))

	logger.Info(ctx, log, LogLoadingConfig)
	cfg, err := config.Load[setup.AuthConfig](ctx)
	if err != nil {
		logger.Error(ctx, log, ErrLoadConfig, zap.Error(err))
		exitCode = 1
		return
	}

	pgConfig := cfg.GetAuthPostgresConfig()
	logger.Info(ctx, log, "Database configuration loaded",
		zap.String("host", pgConfig.Host),
		zap.Int("port", pgConfig.Port),
		zap.String("database", pgConfig.Database),
		zap.String("user", pgConfig.User),
		zap.String("ssl_mode", pgConfig.SSLMode))

	logger.Info(ctx, log, LogConfigLoaded)

	var logImpl logger.ZapLogger
	if cfg.Logger.Model == "production" {
		logImpl, err = logger.Production()
	} else {
		logImpl, err = logger.Development()
	}
	if err != nil {
		logger.Error(ctx, log, ErrInitLogger, zap.Error(err))
		exitCode = 1
		return
	}
	log = logImpl
	ctx = logger.WithLogger(ctx, log)

	isConfigChanged := false

	if pgConfig.Host == "" {
		pgConfig.Host = "localhost"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting database host for local development", zap.String("host", pgConfig.Host))
	}

	if pgConfig.Port == 0 {
		pgConfig.Port = 5432
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database port", zap.Int("port", pgConfig.Port))
	}

	if pgConfig.User == "" {
		pgConfig.User = "auth"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database user", zap.String("user", pgConfig.User))
	}

	if pgConfig.Password == "" {
		pgConfig.Password = "auth"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database password", zap.String("password", "****"))
	}

	if pgConfig.Database == "" {
		pgConfig.Database = "auth"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database name", zap.String("database", pgConfig.Database))
	}

	if pgConfig.SSLMode == "" {
		pgConfig.SSLMode = "disable"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database SSL mode", zap.String("ssl_mode", pgConfig.SSLMode))
	}

	if isConfigChanged {
		logger.Info(ctx, log, "Updated database configuration",
			zap.String("host", pgConfig.Host),
			zap.Int("port", pgConfig.Port),
			zap.String("database", pgConfig.Database),
			zap.String("user", pgConfig.User),
			zap.String("ssl_mode", pgConfig.SSLMode))
	}

	logger.Info(ctx, log, LogInitDB)
	db, err := database.NewPostgres(ctx, pgConfig)
	if err != nil {
		logger.Error(ctx, log, ErrInitDB, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogDBInitialized)

	logger.Info(ctx, log, LogRunMigrations)
	migrator := database.NewMigrator()
	migrateConfig := migrate.Config{
		Path: cfg.GetAuthPgxConfig().MigratePath,
	}
	if err := migrator.Up(ctx, db.GetDSN(), migrateConfig); err != nil {
		logger.Error(ctx, log, ErrRunMigrations, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogMigrationsCompleted)

	shutdown.Wait(ctx, cfg.GetShutdownTimeout(),
		func(ctx context.Context) error {
			logger.Info(ctx, log, LogClosingDB)
			db.Close(ctx)
			return nil
		},
	)

	logger.Info(ctx, log, LogServiceShutdownDone)
}
