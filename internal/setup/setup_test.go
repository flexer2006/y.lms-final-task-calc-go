package setup

import (
	"testing"
	"time"

	authpgx "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/db/pgxx"
	authpg "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/db/postgres"
	authgrpc "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/grpc"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/jwt"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/logger"
	orchagent "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/agent"
	orchpgx "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db/pgxx"
	orchpg "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db/postgres"
	orchgrpc "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/grpc"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/server"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/shutdown"
	"github.com/stretchr/testify/assert"
)

func createBaseConfig() BaseConfig {
	return BaseConfig{
		Logger: logger.Config{
			Level:        "info",
			Format:       "json",
			Output:       "stdout",
			TimeEncoding: "iso8601",
			Caller:       true,
			Stacktrace:   true,
			Model:        "development",
		},
		GracefulShutdown: shutdown.Config{
			ShutdownTimeout: 5 * time.Second,
		},
		JWT: jwt.Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 24 * time.Hour,
			BCryptCost:      10,
		},
	}
}

func createAuthConfig() AuthConfig {
	base := createBaseConfig()
	return AuthConfig{
		Logger:           base.Logger,
		GracefulShutdown: base.GracefulShutdown,
		JWT:              base.JWT,
		AuthGrpc: authgrpc.Config{
			Host: "0.0.0.0",
			Port: 50052,
		},
		AuthDbPostgres: authpg.Config{
			Host:              "auth-db",
			Port:              5432,
			User:              "auth",
			Password:          "auth",
			Database:          "auth",
			SSLMode:           "disable",
			ConnRetry:         3,
			ConnRetryInterval: 5 * time.Second,
			StatementTimeout:  60 * time.Second,
			ApplicationName:   "auth-service",
		},
		AuthDbPgx: authpgx.Config{
			PoolMaxConns:    10,
			PoolMinConns:    1,
			ConnectTimeout:  10 * time.Second,
			AcquireTimeout:  60 * time.Second,
			MaxConnLifetime: 3600 * time.Second,
			MaxConnIdleTime: 600 * time.Second,
			PoolLifetime:    3600 * time.Second,
			MigratePath:     "./migrations/auth",
		},
	}
}

func createOrchestratorConfig() OrchestratorConfig {
	base := createBaseConfig()
	return OrchestratorConfig{
		Logger:           base.Logger,
		GracefulShutdown: base.GracefulShutdown,
		JWT:              base.JWT,
		OrchGrpc: orchgrpc.Config{
			Host: "0.0.0.0",
			Port: 50053,
		},
		OrchAgent: orchagent.Config{
			ComputerPower:       4,
			TimeAddition:        1 * time.Second,
			TimeSubtraction:     1 * time.Second,
			TimeMultiplications: 2 * time.Second,
			TimeDivisions:       2 * time.Second,
			MaxOperations:       100,
		},
		OrchDbPostgres: orchpg.Config{
			Host:              "orchestrator-db",
			Port:              5433,
			User:              "orchestrator",
			Password:          "orchestrator",
			Database:          "orchestrator",
			SSLMode:           "disable",
			ConnRetry:         3,
			ConnRetryInterval: 5 * time.Second,
			StatementTimeout:  60 * time.Second,
			ApplicationName:   "orchestrator-service",
		},
		OrchDbPgx: orchpgx.Config{
			PoolMaxConns:    10,
			PoolMinConns:    1,
			ConnectTimeout:  10 * time.Second,
			AcquireTimeout:  60 * time.Second,
			MaxConnLifetime: 3600 * time.Second,
			MaxConnIdleTime: 600 * time.Second,
			PoolLifetime:    3600 * time.Second,
			MigratePath:     "./migrations/orchestrator",
		},
	}
}

func createServerConfig() ServerConfig {
	base := createBaseConfig()
	return ServerConfig{
		Logger:           base.Logger,
		GracefulShutdown: base.GracefulShutdown,
		JWT:              base.JWT,
		Server: server.Config{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		AuthGrpc: authgrpc.Config{
			Host: "auth",
			Port: 50052,
		},
		OrchGrpc: orchgrpc.Config{
			Host: "orchestrator",
			Port: 50053,
		},
		OrchAgent: orchagent.Config{
			ComputerPower:       4,
			TimeAddition:        1 * time.Second,
			TimeSubtraction:     1 * time.Second,
			TimeMultiplications: 2 * time.Second,
			TimeDivisions:       2 * time.Second,
			MaxOperations:       100,
		},
	}
}

// Тесты для BaseConfig
func TestBaseConfig(t *testing.T) {
	config := createBaseConfig()

	t.Run("GetLoggerConfig", func(t *testing.T) {
		result := config.GetLoggerConfig()
		assert.Equal(t, config.Logger, result)
	})

	t.Run("GetJWTConfig", func(t *testing.T) {
		result := config.GetJWTConfig()
		assert.Equal(t, config.JWT, result)
	})

	t.Run("GetShutdownConfig", func(t *testing.T) {
		result := config.GetShutdownConfig()
		assert.Equal(t, config.GracefulShutdown, result)
	})

	t.Run("GetAccessTokenTTL", func(t *testing.T) {
		result := config.GetAccessTokenTTL()
		assert.Equal(t, config.JWT.AccessTokenTTL, result)
	})

	t.Run("GetRefreshTokenTTL", func(t *testing.T) {
		result := config.GetRefreshTokenTTL()
		assert.Equal(t, config.JWT.RefreshTokenTTL, result)
	})

	t.Run("GetShutdownTimeout", func(t *testing.T) {
		result := config.GetShutdownTimeout()
		assert.Equal(t, config.GracefulShutdown.ShutdownTimeout, result)
	})
}

// Тесты для AuthConfig
func TestAuthConfig(t *testing.T) {
	config := createAuthConfig()

	t.Run("GetLoggerConfig", func(t *testing.T) {
		result := config.GetLoggerConfig()
		assert.Equal(t, config.Logger, result)
	})

	t.Run("GetJWTConfig", func(t *testing.T) {
		result := config.GetJWTConfig()
		assert.Equal(t, config.JWT, result)
	})

	t.Run("GetAuthGRPCConfig", func(t *testing.T) {
		result := config.GetAuthGRPCConfig()
		assert.Equal(t, config.AuthGrpc, result)
	})

	t.Run("GetAuthPostgresConfig", func(t *testing.T) {
		result := config.GetAuthPostgresConfig()
		assert.Equal(t, config.AuthDbPostgres, result)
	})

	t.Run("GetAuthPgxConfig", func(t *testing.T) {
		result := config.GetAuthPgxConfig()
		assert.Equal(t, config.AuthDbPgx, result)
	})

	t.Run("GetShutdownConfig", func(t *testing.T) {
		result := config.GetShutdownConfig()
		assert.Equal(t, config.GracefulShutdown, result)
	})

	t.Run("GetAuthGRPCAddress", func(t *testing.T) {
		result := config.GetAuthGRPCAddress()
		expected := "0.0.0.0:50052"
		assert.Equal(t, expected, result)
	})

	t.Run("GetConnectionURL", func(t *testing.T) {
		result := config.GetConnectionURL()
		expected := "postgres://auth:auth@auth-db:5432/auth?sslmode=disable"
		assert.Equal(t, expected, result)
	})

	t.Run("GetDSN", func(t *testing.T) {
		result := config.GetDSN()
		expected := "host=auth-db port=5432 user=auth password=auth dbname=auth sslmode=disable"
		assert.Equal(t, expected, result)
	})

	t.Run("GetAccessTokenTTL", func(t *testing.T) {
		result := config.GetAccessTokenTTL()
		assert.Equal(t, config.JWT.AccessTokenTTL, result)
	})

	t.Run("GetRefreshTokenTTL", func(t *testing.T) {
		result := config.GetRefreshTokenTTL()
		assert.Equal(t, config.JWT.RefreshTokenTTL, result)
	})

	t.Run("GetShutdownTimeout", func(t *testing.T) {
		result := config.GetShutdownTimeout()
		assert.Equal(t, config.GracefulShutdown.ShutdownTimeout, result)
	})

	t.Run("ToPostgresConfig", func(t *testing.T) {
		result := config.ToPostgresConfig()
		assert.Equal(t, config.AuthDbPostgres.Host, result.Host)
		assert.Equal(t, config.AuthDbPostgres.Port, result.Port)
		assert.Equal(t, config.AuthDbPostgres.User, result.User)
		assert.Equal(t, config.AuthDbPostgres.Password, result.Password)
		assert.Equal(t, config.AuthDbPostgres.Database, result.Database)
		assert.Equal(t, config.AuthDbPostgres.SSLMode, result.SSLMode)
		assert.Equal(t, config.AuthDbPostgres.ApplicationName, result.ApplicationName)
		assert.Equal(t, config.AuthDbPostgres.ConnRetryInterval, result.ConnTimeout)
		assert.Equal(t, config.AuthDbPgx.PoolMinConns, result.MinConns)
		assert.Equal(t, config.AuthDbPgx.PoolMaxConns, result.MaxConns)
	})
}

// Тесты для OrchestratorConfig
func TestOrchestratorConfig(t *testing.T) {
	config := createOrchestratorConfig()

	t.Run("GetLoggerConfig", func(t *testing.T) {
		result := config.GetLoggerConfig()
		assert.Equal(t, config.Logger, result)
	})

	t.Run("GetJWTConfig", func(t *testing.T) {
		result := config.GetJWTConfig()
		assert.Equal(t, config.JWT, result)
	})

	t.Run("GetOrchestratorGRPCConfig", func(t *testing.T) {
		result := config.GetOrchestratorGRPCConfig()
		assert.Equal(t, config.OrchGrpc, result)
	})

	t.Run("GetOrchestratorAgentConfig", func(t *testing.T) {
		result := config.GetOrchestratorAgentConfig()
		assert.Equal(t, config.OrchAgent, result)
	})

	t.Run("GetOrchestratorPostgresConfig", func(t *testing.T) {
		result := config.GetOrchestratorPostgresConfig()
		assert.Equal(t, config.OrchDbPostgres, result)
	})

	t.Run("GetOrchestratorPgxConfig", func(t *testing.T) {
		result := config.GetOrchestratorPgxConfig()
		assert.Equal(t, config.OrchDbPgx, result)
	})

	t.Run("GetShutdownConfig", func(t *testing.T) {
		result := config.GetShutdownConfig()
		assert.Equal(t, config.GracefulShutdown, result)
	})

	t.Run("GetOrchestratorGRPCAddress", func(t *testing.T) {
		result := config.GetOrchestratorGRPCAddress()
		expected := "0.0.0.0:50053"
		assert.Equal(t, expected, result)
	})

	t.Run("GetConnectionURL", func(t *testing.T) {
		result := config.GetConnectionURL()
		expected := "postgres://orchestrator:orchestrator@orchestrator-db:5433/orchestrator?sslmode=disable"
		assert.Equal(t, expected, result)
	})

	t.Run("GetDSN", func(t *testing.T) {
		result := config.GetDSN()
		expected := "host=orchestrator-db port=5433 user=orchestrator password=orchestrator dbname=orchestrator sslmode=disable"
		assert.Equal(t, expected, result)
	})

	t.Run("GetAccessTokenTTL", func(t *testing.T) {
		result := config.GetAccessTokenTTL()
		assert.Equal(t, config.JWT.AccessTokenTTL, result)
	})

	t.Run("GetRefreshTokenTTL", func(t *testing.T) {
		result := config.GetRefreshTokenTTL()
		assert.Equal(t, config.JWT.RefreshTokenTTL, result)
	})

	t.Run("GetShutdownTimeout", func(t *testing.T) {
		result := config.GetShutdownTimeout()
		assert.Equal(t, config.GracefulShutdown.ShutdownTimeout, result)
	})

	t.Run("GetAgentComputerPower", func(t *testing.T) {
		result := config.GetAgentComputerPower()
		assert.Equal(t, config.OrchAgent.ComputerPower, result)
	})

	t.Run("GetAgentOperationTimes", func(t *testing.T) {
		result := config.GetAgentOperationTimes()
		assert.Equal(t, config.OrchAgent.TimeAddition, result["addition"])
		assert.Equal(t, config.OrchAgent.TimeSubtraction, result["subtraction"])
		assert.Equal(t, config.OrchAgent.TimeMultiplications, result["multiplication"])
		assert.Equal(t, config.OrchAgent.TimeDivisions, result["division"])
	})

	t.Run("GetMaxOperations", func(t *testing.T) {
		result := config.GetMaxOperations()
		assert.Equal(t, config.OrchAgent.MaxOperations, result)
	})

	t.Run("ToPostgresConfig", func(t *testing.T) {
		result := config.ToPostgresConfig()
		assert.Equal(t, config.OrchDbPostgres.Host, result.Host)
		assert.Equal(t, config.OrchDbPostgres.Port, result.Port)
		assert.Equal(t, config.OrchDbPostgres.User, result.User)
		assert.Equal(t, config.OrchDbPostgres.Password, result.Password)
		assert.Equal(t, config.OrchDbPostgres.Database, result.Database)
		assert.Equal(t, config.OrchDbPostgres.SSLMode, result.SSLMode)
		assert.Equal(t, config.OrchDbPostgres.ApplicationName, result.ApplicationName)
		assert.Equal(t, config.OrchDbPostgres.ConnRetryInterval, result.ConnTimeout)
		assert.Equal(t, config.OrchDbPgx.PoolMinConns, result.MinConns)
		assert.Equal(t, config.OrchDbPgx.PoolMaxConns, result.MaxConns)
		assert.Equal(t, config.OrchDbPgx.MaxConnLifetime, result.MaxConnLifetime)
		assert.Equal(t, config.OrchDbPgx.MaxConnIdleTime, result.MaxConnIdleTime)
		assert.Equal(t, 30*time.Second, result.HealthPeriod)
	})
}

// Тесты для ServerConfig
func TestServerConfig(t *testing.T) {
	config := createServerConfig()

	t.Run("GetLoggerConfig", func(t *testing.T) {
		result := config.GetLoggerConfig()
		assert.Equal(t, config.Logger, result)
	})

	t.Run("GetServerConfig", func(t *testing.T) {
		result := config.GetServerConfig()
		assert.Equal(t, config.Server, result)
	})

	t.Run("GetShutdownConfig", func(t *testing.T) {
		result := config.GetShutdownConfig()
		assert.Equal(t, config.GracefulShutdown, result)
	})

	t.Run("GetServerAddress", func(t *testing.T) {
		result := config.GetServerAddress()
		expected := "0.0.0.0:8080"
		assert.Equal(t, expected, result)
	})

	t.Run("GetShutdownTimeout", func(t *testing.T) {
		result := config.GetShutdownTimeout()
		assert.Equal(t, config.GracefulShutdown.ShutdownTimeout, result)
	})

	t.Run("GetAuthGRPCConfig", func(t *testing.T) {
		result := config.GetAuthGRPCConfig()
		assert.Equal(t, config.AuthGrpc.Host, result.Host)
		assert.Equal(t, config.AuthGrpc.Port, result.Port)
	})

	t.Run("GetOrchestratorGRPCConfig", func(t *testing.T) {
		result := config.GetOrchestratorGRPCConfig()
		assert.Equal(t, config.OrchGrpc.Host, result.Host)
		assert.Equal(t, config.OrchGrpc.Port, result.Port)
	})
}
