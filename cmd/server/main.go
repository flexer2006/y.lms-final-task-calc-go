// Package main реализует точку входа HTTP API шлюза.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	httpserver "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http"

	authclient "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/grpc/clients/auth"
	orchclient "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/grpc/clients/orchestrator"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/config"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/shutdown"
	"go.uber.org/zap"
)

const (
	ErrInitLogger     = "failed to initialize logger"
	ErrSyncLogger     = "failed to sync logger"
	ErrLoadConfig     = "failed to load configuration"
	ErrConnectAuth    = "failed to connect to auth service"
	ErrConnectOrch    = "failed to connect to orchestrator service"
	ErrInitHTTPServer = "failed to initialize HTTP server"
	ErrStartHTTP      = "failed to start HTTP server"
)

const (
	ErrSyncStderr = "sync /dev/stderr: invalid argument"
	ErrSyncStdout = "sync /dev/stdout: invalid argument"
)

const (
	LogServiceStarted      = "API gateway service started"
	LogServiceShutdownDone = "API gateway service shutdown complete"
	LogLoadingConfig       = "loading configuration"
	LogConfigLoaded        = "configuration loaded"
	LogInitHTTPServer      = "initializing HTTP server"
	LogHTTPListening       = "HTTP server listening"
	LogHTTPShutdown        = "shutting down HTTP server"
	LogConnectingToAuth    = "connecting to auth service"
	LogConnectingToOrch    = "connecting to orchestrator service"
	LogServicesConnected   = "connected to all services"
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
	cfg, err := config.Load[setup.ServerConfig](ctx)
	if err != nil {
		logger.Error(ctx, log, ErrLoadConfig, zap.Error(err))
		exitCode = 1
		return
	}

	serverConfig := cfg.GetServerConfig()
	logger.Info(ctx, log, "HTTP server configuration loaded",
		zap.String("host", serverConfig.Host),
		zap.Int("port", serverConfig.Port))

	authConfig := cfg.GetAuthGRPCConfig()
	logger.Info(ctx, log, "Auth service configuration loaded",
		zap.String("host", authConfig.Host),
		zap.Int("port", authConfig.Port))

	orchConfig := cfg.GetOrchestratorGRPCConfig()
	logger.Info(ctx, log, "Orchestrator service configuration loaded",
		zap.String("host", orchConfig.Host),
		zap.Int("port", orchConfig.Port))

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

	logger.Info(ctx, log, LogConnectingToAuth)
	authAddress := fmt.Sprintf("%s:%d", authConfig.Host, authConfig.Port)

	authUseCase, err := authclient.NewAuthUseCase(ctx, authAddress)
	if err != nil {
		logger.Error(ctx, log, ErrConnectAuth, zap.Error(err))
		exitCode = 1
		return
	}

	// Properly handle Close error
	defer func() {
		if err := authUseCase.Close(); err != nil {
			logger.Error(ctx, log, "Failed to close auth use case", zap.Error(err))
		}
	}()
	logger.Info(ctx, log, "Connected to auth service")

	logger.Info(ctx, log, LogConnectingToOrch)
	orchAddress := fmt.Sprintf("%s:%d", orchConfig.Host, orchConfig.Port)

	orchUseCase, err := orchclient.NewCalculationUseCase(ctx, orchAddress)
	if err != nil {
		logger.Error(ctx, log, ErrConnectOrch, zap.Error(err))
		exitCode = 1
		return
	}

	// Properly handle Close error
	defer func() {
		if err := orchUseCase.Close(); err != nil {
			logger.Error(ctx, log, "Failed to close orchestrator use case", zap.Error(err))
		}
	}()
	logger.Info(ctx, log, LogServicesConnected)

	logger.Info(ctx, log, LogInitHTTPServer)
	server := httpserver.NewServer(serverConfig, authUseCase, orchUseCase)

	if err := server.Start(ctx); err != nil {
		logger.Error(ctx, log, ErrStartHTTP, zap.Error(err))
		exitCode = 1
		return
	}

	serverAddress := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	logger.Info(ctx, log, LogHTTPListening, zap.String("address", serverAddress))

	shutdown.Wait(ctx, cfg.GetShutdownTimeout(),
		func(ctx context.Context) error {
			logger.Info(ctx, log, LogHTTPShutdown)
			return server.Stop(ctx)
		},
	)

	logger.Info(ctx, log, LogServiceShutdownDone)
}
