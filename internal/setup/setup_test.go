package setup_test

import (
	"testing"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth"
	authdb "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/db"
	authpgx "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/db/pgxx"
	authgrpc "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/grpc"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/jwt"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/logger"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator"
	orchagent "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/agent"
	orchdb "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db"
	orchpgx "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db/pgxx"
	orchpg "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db/postgres"
	orchgrpc "github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/grpc"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/server"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/shutdown"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/postgres"
	"github.com/stretchr/testify/assert"
)

func TestBaseConfig(t *testing.T) {
	baseConfig := setup.BaseConfig{
		Logger: logger.Config{
			Level:        "debug",
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

	authConfig := setup.AuthConfig{
		BaseConfig: baseConfig,
	}

	t.Run("BaseConfig methods through AuthConfig", func(t *testing.T) {
		assert.Equal(t, baseConfig.Logger, authConfig.GetLoggerConfig())
		assert.Equal(t, baseConfig.JWT, authConfig.GetJWTConfig())
		assert.Equal(t, baseConfig.GracefulShutdown, authConfig.GetShutdownConfig())
		assert.Equal(t, baseConfig.JWT.AccessTokenTTL, authConfig.GetAccessTokenTTL())
		assert.Equal(t, baseConfig.JWT.RefreshTokenTTL, authConfig.GetRefreshTokenTTL())
		assert.Equal(t, baseConfig.GracefulShutdown.ShutdownTimeout, authConfig.GetShutdownTimeout())
	})
}

func TestAuthConfig(t *testing.T) {
	authGrpcConfig := authgrpc.Config{
		Host: "auth-host",
		Port: 50052,
	}

	authPostgresConfig := postgres.Config{
		Host:            "postgres-host",
		Port:            5432,
		User:            "postgres-user",
		Password:        "postgres-password",
		Database:        "postgres-db",
		SSLMode:         "disable",
		MinConns:        2,
		MaxConns:        10,
		ConnTimeout:     10 * time.Second,
		HealthPeriod:    30 * time.Second,
		ApplicationName: "auth-service",
	}

	authPgxConfig := authpgx.Config{
		PoolMaxConns:    10,
		PoolMinConns:    1,
		ConnectTimeout:  10 * time.Second,
		AcquireTimeout:  60 * time.Second,
		MaxConnLifetime: 3600 * time.Second,
		MaxConnIdleTime: 600 * time.Second,
		PoolLifetime:    3600 * time.Second,
		MigratePath:     "./migrations/auth",
	}

	authDBConfig := authdb.Config{
		Postgres: authPostgresConfig,
		Pgx:      authPgxConfig,
	}

	authConfig := setup.AuthConfig{
		BaseConfig: setup.BaseConfig{
			Logger: logger.Config{
				Level:        "debug",
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
		},
		Auth: auth.Config{
			Db:   authDBConfig,
			Grpc: authGrpcConfig,
		},
	}

	tests := []struct {
		name     string
		method   func() any
		expected any
	}{
		{
			name:     "GetAuthGRPCConfig",
			method:   func() any { return authConfig.GetAuthGRPCConfig() },
			expected: authGrpcConfig,
		},
		{
			name:     "GetAuthPostgresConfig",
			method:   func() any { return authConfig.GetAuthPostgresConfig() },
			expected: authPostgresConfig,
		},
		{
			name:     "GetAuthPgxConfig",
			method:   func() any { return authConfig.GetAuthPgxConfig() },
			expected: authPgxConfig,
		},
		{
			name:     "GetAuthDBConfig",
			method:   func() any { return authConfig.GetAuthDBConfig() },
			expected: authDBConfig,
		},
		{
			name:     "GetAuthGRPCAddress",
			method:   func() any { return authConfig.GetAuthGRPCAddress() },
			expected: "auth-host:50052",
		},
		{
			name:     "GetConnectionURL",
			method:   func() any { return authConfig.GetConnectionURL() },
			expected: "postgres://postgres-user:postgres-password@postgres-host:5432/postgres-db?sslmode=disable",
		},
		{
			name:     "GetDSN",
			method:   func() any { return authConfig.GetDSN() },
			expected: "host=postgres-host port=5432 user=postgres-user password=postgres-password dbname=postgres-db sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrchestratorConfig(t *testing.T) {
	orchGrpcConfig := orchgrpc.Config{
		Host: "orch-host",
		Port: 50053,
	}

	orchPostgresConfig := orchpg.Config{
		Host:              "orch-postgres-host",
		Port:              5433,
		User:              "orch-postgres-user",
		Password:          "orch-postgres-password",
		Database:          "orch-postgres-db",
		SSLMode:           "disable",
		ConnRetry:         3,
		ConnRetryInterval: 5 * time.Second,
		StatementTimeout:  60 * time.Second,
		ApplicationName:   "orchestrator-service",
	}

	orchPgxConfig := orchpgx.Config{
		PoolMaxConns:    10,
		PoolMinConns:    1,
		ConnectTimeout:  10 * time.Second,
		AcquireTimeout:  60 * time.Second,
		MaxConnLifetime: 3600 * time.Second,
		MaxConnIdleTime: 600 * time.Second,
		PoolLifetime:    3600 * time.Second,
		MigratePath:     "./migrations/orchestrator",
	}

	orchDBConfig := orchdb.Config{
		Postgres: orchPostgresConfig,
		Pgx:      orchPgxConfig,
	}

	// Добавляем конфигурацию агентов
	orchAgentConfig := orchagent.Config{
		ComputerPower:       4,
		TimeAddition:        1 * time.Second,
		TimeSubtraction:     1 * time.Second,
		TimeMultiplications: 2 * time.Second,
		TimeDivisions:       2 * time.Second,
	}

	orchConfig := setup.OrchestratorConfig{
		BaseConfig: setup.BaseConfig{
			Logger: logger.Config{
				Level:        "debug",
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
		},
		Orchestrator: orchestrator.Config{
			Db:    orchDBConfig,
			Grpc:  orchGrpcConfig,
			Agent: orchAgentConfig, // Добавляем конфигурацию агентов
		},
	}

	tests := []struct {
		name     string
		method   func() any
		expected any
	}{
		{
			name:     "GetLoggerConfig",
			method:   func() any { return orchConfig.GetLoggerConfig() },
			expected: orchConfig.Logger,
		},
		{
			name:     "GetJWTConfig",
			method:   func() any { return orchConfig.GetJWTConfig() },
			expected: orchConfig.JWT,
		},
		{
			name:     "GetShutdownConfig",
			method:   func() any { return orchConfig.GetShutdownConfig() },
			expected: orchConfig.GracefulShutdown,
		},
		{
			name:     "GetOrchestratorGRPCConfig",
			method:   func() any { return orchConfig.GetOrchestratorGRPCConfig() },
			expected: orchGrpcConfig,
		},
		{
			name:     "GetOrchestratorAgentConfig",
			method:   func() any { return orchConfig.GetOrchestratorAgentConfig() },
			expected: orchAgentConfig,
		},
		{
			name:     "GetOrchestratorPostgresConfig",
			method:   func() any { return orchConfig.GetOrchestratorPostgresConfig() },
			expected: orchPostgresConfig,
		},
		{
			name:     "GetOrchestratorPgxConfig",
			method:   func() any { return orchConfig.GetOrchestratorPgxConfig() },
			expected: orchPgxConfig,
		},
		{
			name:     "GetOrchestratorDBConfig",
			method:   func() any { return orchConfig.GetOrchestratorDBConfig() },
			expected: orchDBConfig,
		},
		{
			name:     "GetOrchestratorGRPCAddress",
			method:   func() any { return orchConfig.GetOrchestratorGRPCAddress() },
			expected: "orch-host:50053",
		},
		{
			name:     "GetConnectionURL",
			method:   func() any { return orchConfig.GetConnectionURL() },
			expected: "postgres://orch-postgres-user:orch-postgres-password@orch-postgres-host:5433/orch-postgres-db?sslmode=disable",
		},
		{
			name:     "GetDSN",
			method:   func() any { return orchConfig.GetDSN() },
			expected: "host=orch-postgres-host port=5433 user=orch-postgres-user password=orch-postgres-password dbname=orch-postgres-db sslmode=disable",
		},
		{
			name:     "GetAgentComputerPower",
			method:   func() any { return orchConfig.GetAgentComputerPower() },
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			assert.Equal(t, tt.expected, result)
		})
	}

	// Отдельный тест для метода GetAgentOperationTimes, так как результат - это map
	t.Run("GetAgentOperationTimes", func(t *testing.T) {
		expected := map[string]time.Duration{
			"addition":       1 * time.Second,
			"subtraction":    1 * time.Second,
			"multiplication": 2 * time.Second,
			"division":       2 * time.Second,
		}
		result := orchConfig.GetAgentOperationTimes()

		assert.Equal(t, expected["addition"], result["addition"])
		assert.Equal(t, expected["subtraction"], result["subtraction"])
		assert.Equal(t, expected["multiplication"], result["multiplication"])
		assert.Equal(t, expected["division"], result["division"])
		assert.Equal(t, len(expected), len(result))
	})
}

func TestServerConfig(t *testing.T) {
	serverConfig := setup.ServerConfig{
		BaseConfig: setup.BaseConfig{
			Logger: logger.Config{
				Level:        "debug",
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
		},
		Server: server.Config{
			Host:         "server-host",
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	tests := []struct {
		name     string
		method   func() any
		expected any
	}{
		{
			name:     "GetLoggerConfig",
			method:   func() any { return serverConfig.GetLoggerConfig() },
			expected: serverConfig.Logger,
		},
		{
			name:     "GetServerConfig",
			method:   func() any { return serverConfig.GetServerConfig() },
			expected: serverConfig.Server,
		},
		{
			name:     "GetShutdownConfig",
			method:   func() any { return serverConfig.GetShutdownConfig() },
			expected: serverConfig.GracefulShutdown,
		},
		{
			name:     "GetServerAddress",
			method:   func() any { return serverConfig.GetServerAddress() },
			expected: "server-host:8080",
		},
		{
			name:     "GetShutdownTimeout",
			method:   func() any { return serverConfig.GetShutdownTimeout() },
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDurationMethods(t *testing.T) {
	baseConfig := setup.BaseConfig{
		JWT: jwt.Config{
			AccessTokenTTL:  30 * time.Minute,
			RefreshTokenTTL: 48 * time.Hour,
		},
		GracefulShutdown: shutdown.Config{
			ShutdownTimeout: 10 * time.Second,
		},
	}

	authConfig := setup.AuthConfig{
		BaseConfig: baseConfig,
	}

	orchConfig := setup.OrchestratorConfig{
		BaseConfig: baseConfig,
	}

	serverConfig := setup.ServerConfig{
		BaseConfig: baseConfig,
	}

	tests := []struct {
		name     string
		method   func() time.Duration
		expected time.Duration
	}{
		{
			name:     "AuthConfig_GetAccessTokenTTL",
			method:   authConfig.GetAccessTokenTTL,
			expected: 30 * time.Minute,
		},
		{
			name:     "AuthConfig_GetRefreshTokenTTL",
			method:   authConfig.GetRefreshTokenTTL,
			expected: 48 * time.Hour,
		},
		{
			name:     "AuthConfig_GetShutdownTimeout",
			method:   authConfig.GetShutdownTimeout,
			expected: 10 * time.Second,
		},
		{
			name:     "OrchestratorConfig_GetAccessTokenTTL",
			method:   orchConfig.GetAccessTokenTTL,
			expected: 30 * time.Minute,
		},
		{
			name:     "OrchestratorConfig_GetRefreshTokenTTL",
			method:   orchConfig.GetRefreshTokenTTL,
			expected: 48 * time.Hour,
		},
		{
			name:     "OrchestratorConfig_GetShutdownTimeout",
			method:   orchConfig.GetShutdownTimeout,
			expected: 10 * time.Second,
		},
		{
			name:     "ServerConfig_GetShutdownTimeout",
			method:   serverConfig.GetShutdownTimeout,
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			assert.Equal(t, tt.expected, result)
		})
	}
}
