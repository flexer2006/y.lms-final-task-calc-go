// Package main реализует точку входа службы аутентификации.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	pgauth "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/db/postgres/auth"
	grpcserver "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/grpc"
	grpcauth "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/grpc/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/services/jwt"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/services/password"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/auth/usecase"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup"
	authv1 "github.com/flexer2006/y.lms-final-task-calc-go/pkg/api/proto/v1/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/config"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/migrate"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/shutdown"
	"go.uber.org/zap"
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
	LogServiceStarted      = "authentication service started"
	LogServiceShutdownDone = "authentication service shutdown complete"
	LogLoadingConfig       = "loading configuration"
	LogConfigLoaded        = "configuration loaded"
	LogInitDB              = "initializing database connection"
	LogDBInitialized       = "database connection established"
	LogRunMigrations       = "running database migrations"
	LogMigrationsCompleted = "database migrations completed"
	LogClosingDB           = "closing database connections"
	LogInitGRPCServer      = "initializing gRPC server"
	LogGRPCListening       = "gRPC server listening"
	LogGRPCShutdown        = "shutting down gRPC server"
	LogRegisteringService  = "registering auth gRPC service"
	LogInitServices        = "initializing services"
	LogServicesInitialized = "services initialized"
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

	logger.Info(ctx, log, LogInitDB)

	dbConfig := cfg.ToPostgresConfig()

	db, err := database.NewPostgres(ctx, dbConfig)
	if err != nil {
		logger.Error(ctx, log, ErrInitDB, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogDBInitialized)

	dbHandler := &database.Handler{
		DB:       db,
		Migrator: database.NewMigrator(),
	}

	logger.Info(ctx, log, LogRunMigrations)
	migrateConfig := migrate.Config{
		Path: cfg.GetAuthPgxConfig().MigratePath,
	}
	if err := dbHandler.MigrateUp(ctx, migrateConfig); err != nil {
		logger.Error(ctx, log, ErrRunMigrations, zap.Error(err))
		exitCode = 1
		return
	}
	logger.Info(ctx, log, LogMigrationsCompleted)

	logger.Info(ctx, log, "Initializing repositories")
	userRepo := pgauth.NewUserRepository(dbHandler)
	tokenRepo := pgauth.NewTokenRepository(dbHandler)
	logger.Info(ctx, log, "Repositories initialized")

	logger.Info(ctx, log, LogInitServices)
	jwtConfig := cfg.GetJWTConfig()
	passwordService := password.NewService(jwtConfig.BCryptCost)
	jwtService := jwt.NewService(
		jwtConfig.SecretKey,
		jwtConfig.AccessTokenTTL,
		jwtConfig.RefreshTokenTTL,
	)
	logger.Info(ctx, log, LogServicesInitialized)

	logger.Info(ctx, log, "Initializing use cases")
	authUseCase := usecase.NewAuthUseCase(userRepo, tokenRepo, passwordService, jwtService)
	logger.Info(ctx, log, "Use cases initialized")

	logger.Info(ctx, log, LogInitGRPCServer)
	grpcConfig := cfg.GetAuthGRPCConfig()

	grpcServer := grpcserver.NewServerAuth()

	authServer := grpcauth.NewServer(authUseCase)
	logger.Info(ctx, log, LogRegisteringService)
	authv1.RegisterAuthServiceServer(grpcServer, authServer)

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

			logger.Info(ctx, log, LogClosingDB)
			dbHandler.Close(ctx)
			return nil
		},
	)

	logger.Info(ctx, log, LogServiceShutdownDone)
}
