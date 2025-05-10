// Package main реализует точку входа службы оркестрации.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	pgorch "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/db/postgres/orchestrator"
	grpcserver "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/grpc"
	grpcorch "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/grpc/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/services/parser"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/orchestrator/calculation"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/orchestrator/processor"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup"
	orchv1 "github.com/flexer2006/y.lms-final-task-calc-go/pkg/api/proto/v1/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/config"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/migrate"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/shutdown"
	"go.uber.org/zap"

	memAgent "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/db/memory/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/agent/executor"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/agent/pool"
	"github.com/google/uuid"
)

const (
	ErrInitLogger     = "failed to initialize logger"
	ErrSyncLogger     = "failed to sync logger"
	ErrLoadConfig     = "failed to load configuration"
	ErrInitDB         = "failed to initialize database"
	ErrRunMigrations  = "failed to run migrations"
	ErrInitGRPCServer = "failed to initialize gRPC server"
	ErrStartGRPC      = "failed to start gRPC server"
)

const (
	ErrSyncStderr = "sync /dev/stderr: invalid argument"
	ErrSyncStdout = "sync /dev/stdout: invalid argument"
)

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
	LogInitGRPCServer      = "initializing gRPC server"
	LogGRPCListening       = "gRPC server listening"
	LogGRPCShutdown        = "shutting down gRPC server"
	LogRegisteringService  = "registering orchestrator gRPC service"
	LogInitServices        = "initializing services"
	LogServicesInitialized = "services initialized"
	LogInitProcessor       = "initializing operation processor"
	LogProcessorStarted    = "operation processor started"
	LogProcessorShutdown   = "shutting down operation processor"
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

	logger.Info(ctx, log, LogInitDB)

	// Get base config from environment
	dbConfig := cfg.ToPostgresConfig()

	// Set parameters that might be missing from environment config
	dbConfig.MaxConnLifetime = cfg.OrchDbPgx.MaxConnLifetime
	dbConfig.MaxConnIdleTime = cfg.OrchDbPgx.MaxConnIdleTime
	dbConfig.ConnTimeout = cfg.OrchDbPgx.ConnectTimeout
	dbConfig.HealthPeriod = 30 * time.Second

	// Log the connection pool settings
	logger.Info(ctx, log, "Configuring database connection pool",
		zap.Int("min_connections", dbConfig.MinConns),
		zap.Int("max_connections", dbConfig.MaxConns),
		zap.Duration("conn_timeout", dbConfig.ConnTimeout),
		zap.Duration("max_lifetime", dbConfig.MaxConnLifetime),
		zap.Duration("idle_timeout", dbConfig.MaxConnIdleTime))

	db, err := database.NewPostgres(ctx, dbConfig)
	if err != nil {
		logger.Error(ctx, log, ErrInitDB, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogDBInitialized)

	logger.Info(ctx, log, LogRunMigrations)

	dbHandler := &database.Handler{
		DB:       db,
		Migrator: database.NewMigrator(),
	}

	migrateConfig := migrate.Config{
		Path: cfg.GetOrchestratorPgxConfig().MigratePath,
	}
	if err := dbHandler.MigrateUp(ctx, migrateConfig); err != nil {
		logger.Error(ctx, log, ErrRunMigrations, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogMigrationsCompleted)

	logger.Info(ctx, log, "Initializing repositories")
	calculationRepo := pgorch.NewCalculationRepository(dbHandler)
	operationRepo := pgorch.NewOperationRepository(dbHandler)
	logger.Info(ctx, log, "Repositories initialized")

	logger.Info(ctx, log, LogInitServices)
	parserService := parser.NewService(cfg.GetMaxOperations())
	logger.Info(ctx, log, LogServicesInitialized)

	logger.Info(ctx, log, "Initializing use cases")
	calculationUseCase := calculation.NewUseCase(calculationRepo, operationRepo, parserService)
	logger.Info(ctx, log, "Use cases initialized")

	logger.Info(ctx, log, "Initializing agent components")

	agentStorage := memAgent.NewAgentStorage()

	operationTimes := map[string]time.Duration{
		"addition":       agentConfig.TimeAddition,
		"subtraction":    agentConfig.TimeSubtraction,
		"multiplication": agentConfig.TimeMultiplications,
		"division":       agentConfig.TimeDivisions,
	}

	agentPool, err := pool.NewAgentPool(agentStorage, operationRepo, operationTimes, agentConfig.ComputerPower)
	if err != nil {
		logger.Error(ctx, log, "Failed to create agent pool", zap.Error(err))
		exitCode = 1
		return
	}
	agentPool.Start(ctx)

	operationExecutor := executor.NewOperationExecutor(agentPool, 3, 500*time.Millisecond)

	logger.Info(ctx, log, "Agent components initialized")

	logger.Info(ctx, log, LogInitProcessor)
	processorConfig := processor.AgentConfig{
		AgentID:             uuid.New().String()[:8],
		ComputerPower:       agentConfig.ComputerPower,
		TimeAddition:        agentConfig.TimeAddition,
		TimeSubtraction:     agentConfig.TimeSubtraction,
		TimeMultiplications: agentConfig.TimeMultiplications,
		TimeDivisions:       agentConfig.TimeDivisions,
	}

	operationProcessor := processor.NewProcessor(
		operationRepo,
		calculationRepo,
		calculationUseCase,
		processorConfig,
		operationExecutor,
		agentPool,
	)

	if err := operationProcessor.Start(ctx); err != nil {
		logger.Error(ctx, log, "Failed to start operation processor", zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogProcessorStarted)

	logger.Info(ctx, log, LogInitGRPCServer)

	grpcServer := grpcserver.NewServerOrchestrator()

	orchestratorServer := grpcorch.NewServer(calculationUseCase)
	logger.Info(ctx, log, LogRegisteringService)
	orchv1.RegisterOrchestratorServiceServer(grpcServer, orchestratorServer)

	grpcAddress := fmt.Sprintf("%s:%d", grpcConfig.Host, grpcConfig.Port)
	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		logger.Error(ctx, log, ErrInitGRPCServer, zap.Error(err))
		exitCode = 1
		return
	}

	go func() {
		logger.Info(ctx, log, LogGRPCListening, zap.String("address", grpcAddress))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error(ctx, log, ErrStartGRPC, zap.Error(err))
		}
	}()

	shutdown.Wait(ctx, cfg.GetShutdownTimeout(),
		func(ctx context.Context) error {
			logger.Info(ctx, log, LogGRPCShutdown)
			grpcServer.GracefulStop()

			logger.Info(ctx, log, LogProcessorShutdown)
			operationProcessor.Stop()

			logger.Info(ctx, log, "Shutting down agent pool")
			agentPool.Stop(ctx) // Pass context here

			logger.Info(ctx, log, LogClosingDB)
			db.Close(ctx)
			return nil
		},
	)

	logger.Info(ctx, log, LogServiceShutdownDone)
}
