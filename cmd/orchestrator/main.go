// Package main реализует точку входа службы оркестрации.
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
	LogServiceStarted      = "orchestrator service started"
	LogServiceShutdownDone = "orchestrator service shutdown complete"
	LogLoadingConfig       = "loading configuration"
	LogConfigLoaded        = "configuration loaded"
	LogInitDB              = "initializing database connection"
	LogDBInitialized       = "database connection established"
	LogRunMigrations       = "running database migrations"
	LogMigrationsCompleted = "database migrations completed"
	LogClosingDB           = "closing database connections"
	LogAgentsInfo          = "agent configuration loaded"
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
	cfg, err := config.Load[setup.OrchestratorConfig](ctx)
	if err != nil {
		logger.Error(ctx, log, ErrLoadConfig, zap.Error(err))
		exitCode = 1
		return
	}

	pgConfig := cfg.GetOrchestratorPostgresConfig()
	logger.Info(ctx, log, "Database configuration loaded",
		zap.String("host", pgConfig.Host),
		zap.Int("port", pgConfig.Port),
		zap.String("database", pgConfig.Database),
		zap.String("user", pgConfig.User),
		zap.String("ssl_mode", pgConfig.SSLMode))

	agentConfig := cfg.GetOrchestratorAgentConfig()
	logger.Info(ctx, log, LogAgentsInfo,
		zap.Int("computer_power", agentConfig.ComputerPower),
		zap.Duration("time_addition", agentConfig.TimeAddition),
		zap.Duration("time_subtraction", agentConfig.TimeSubtraction),
		zap.Duration("time_multiplication", agentConfig.TimeMultiplications),
		zap.Duration("time_division", agentConfig.TimeDivisions))

	grpcConfig := cfg.GetOrchestratorGRPCConfig()
	logger.Info(ctx, log, "gRPC configuration loaded",
		zap.String("host", grpcConfig.Host),
		zap.Int("port", grpcConfig.Port))

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

	if pgConfig.Host == "orchestrator-db" {
		pgConfig.Host = "localhost"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting database host for local development", zap.String("host", pgConfig.Host))
	} else if pgConfig.Host == "" {
		pgConfig.Host = "localhost"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database host", zap.String("host", pgConfig.Host))
	}

	if pgConfig.Port == 0 {
		pgConfig.Port = 5433
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database port", zap.Int("port", pgConfig.Port))
	}

	if pgConfig.User == "" {
		pgConfig.User = "orchestrator"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database user", zap.String("user", pgConfig.User))
	}

	if pgConfig.Password == "" {
		pgConfig.Password = "orchestrator"
		isConfigChanged = true
		logger.Info(ctx, log, "Setting default database password", zap.String("password", "****"))
	}

	if pgConfig.Database == "" {
		pgConfig.Database = "orchestrator"
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

	dbConfig := database.PostgresConfig{
		Host:            pgConfig.Host,
		Port:            pgConfig.Port,
		User:            pgConfig.User,
		Password:        pgConfig.Password,
		Database:        pgConfig.Database,
		SSLMode:         pgConfig.SSLMode,
		ApplicationName: pgConfig.ApplicationName,
		ConnTimeout:     pgConfig.ConnRetryInterval,
	}

	db, err := database.NewPostgres(ctx, dbConfig)
	if err != nil {
		logger.Error(ctx, log, ErrInitDB, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogDBInitialized)

	logger.Info(ctx, log, LogRunMigrations)
	migrator := database.NewMigrator()
	migrateConfig := migrate.Config{
		Path: cfg.GetOrchestratorPgxConfig().MigratePath,
	}
	if err := migrator.Up(ctx, db.GetDSN(), migrateConfig); err != nil {
		logger.Error(ctx, log, ErrRunMigrations, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogMigrationsCompleted)

	// TODO: Добавить инициализацию и запуск сервера оркестрации

	shutdown.Wait(ctx, cfg.GetShutdownTimeout(),
		func(ctx context.Context) error {
			logger.Info(ctx, log, LogClosingDB)
			db.Close(ctx)
			return nil
		},
	)

	logger.Info(ctx, log, LogServiceShutdownDone)
}
