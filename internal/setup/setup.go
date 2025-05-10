package setup

import (
	"fmt"
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
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database"
)

// BaseConfig содержит общие поля для всех конфигураций.
type BaseConfig struct {
	Logger           logger.Config
	GracefulShutdown shutdown.Config
	JWT              jwt.Config
}

// AuthConfig содержит конфигурацию для сервиса аутентификации.
type AuthConfig struct {
	Logger           logger.Config
	GracefulShutdown shutdown.Config
	JWT              jwt.Config
	AuthGrpc         authgrpc.Config
	AuthDbPostgres   authpg.Config
	AuthDbPgx        authpgx.Config
}

// OrchestratorConfig содержит конфигурацию для сервиса оркестрации.
type OrchestratorConfig struct {
	Logger           logger.Config
	GracefulShutdown shutdown.Config
	JWT              jwt.Config
	OrchGrpc         orchgrpc.Config
	OrchAgent        orchagent.Config
	OrchDbPostgres   orchpg.Config
	OrchDbPgx        orchpgx.Config
}

// ServerConfig содержит конфигурацию для API сервера.
type ServerConfig struct {
	Logger           logger.Config
	GracefulShutdown shutdown.Config
	JWT              jwt.Config
	Server           server.Config
	AuthGrpc         authgrpc.Config
	OrchGrpc         orchgrpc.Config
	OrchAgent        orchagent.Config
}

// GetLoggerConfig возвращает конфигурацию журнала.
func (c *BaseConfig) GetLoggerConfig() logger.Config {
	return c.Logger
}

// GetJWTConfig возвращает конфигурацию JWT.
func (c *BaseConfig) GetJWTConfig() jwt.Config {
	return c.JWT
}

// GetShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *BaseConfig) GetShutdownConfig() shutdown.Config {
	return c.GracefulShutdown
}

// GetAccessTokenTTL возвращает время жизни access token.
func (c *BaseConfig) GetAccessTokenTTL() time.Duration {
	return c.JWT.AccessTokenTTL
}

// GetRefreshTokenTTL возвращает время жизни refresh token.
func (c *BaseConfig) GetRefreshTokenTTL() time.Duration {
	return c.JWT.RefreshTokenTTL
}

// GetShutdownTimeout возвращает timeout для graceful shutdown.
func (c *BaseConfig) GetShutdownTimeout() time.Duration {
	return c.GracefulShutdown.ShutdownTimeout
}

// GetLoggerConfig возвращает конфигурацию журнала.
func (c *AuthConfig) GetLoggerConfig() logger.Config {
	return c.Logger
}

// GetJWTConfig возвращает конфигурацию JWT.
func (c *AuthConfig) GetJWTConfig() jwt.Config {
	return c.JWT
}

// GetAuthGRPCConfig возвращает конфигурацию gRPC для сервиса авторизации.
func (c *AuthConfig) GetAuthGRPCConfig() authgrpc.Config {
	return c.AuthGrpc
}

// GetAuthPostgresConfig возвращает конфигурацию Postgres для сервиса авторизации.
func (c *AuthConfig) GetAuthPostgresConfig() authpg.Config {
	return c.AuthDbPostgres
}

// GetAuthPgxConfig возвращает конфигурацию pgx для сервиса авторизации.
func (c *AuthConfig) GetAuthPgxConfig() authpgx.Config {
	return c.AuthDbPgx
}

// GetShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *AuthConfig) GetShutdownConfig() shutdown.Config {
	return c.GracefulShutdown
}

// GetAuthGRPCAddress возвращает адрес gRPC сервера авторизации.
func (c *AuthConfig) GetAuthGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.AuthGrpc.Host, c.AuthGrpc.Port)
}

// GetConnectionURL возвращает URL-строку подключения для миграций.
func (c *AuthConfig) GetConnectionURL() string {
	pg := c.AuthDbPostgres
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pg.User, pg.Password, pg.Host, pg.Port, pg.Database)
}

// GetDSN возвращает DSN-строку подключения для Postgres.
func (c *AuthConfig) GetDSN() string {
	pg := c.AuthDbPostgres
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pg.Host, pg.Port, pg.User, pg.Password, pg.Database)
}

// GetAccessTokenTTL возвращает длительность жизни JWT access token.
func (c *AuthConfig) GetAccessTokenTTL() time.Duration {
	return c.JWT.AccessTokenTTL
}

// GetRefreshTokenTTL возвращает длительность жизни JWT refresh token.
func (c *AuthConfig) GetRefreshTokenTTL() time.Duration {
	return c.JWT.RefreshTokenTTL
}

// GetShutdownTimeout возвращает timeout для graceful shutdown.
func (c *AuthConfig) GetShutdownTimeout() time.Duration {
	return c.GracefulShutdown.ShutdownTimeout
}

// GetLoggerConfig возвращает конфигурацию журнала.
func (c *OrchestratorConfig) GetLoggerConfig() logger.Config {
	return c.Logger
}

// GetJWTConfig возвращает конфигурацию JWT.
func (c *OrchestratorConfig) GetJWTConfig() jwt.Config {
	return c.JWT
}

// GetOrchestratorGRPCConfig возвращает конфигурацию gRPC для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorGRPCConfig() orchgrpc.Config {
	return c.OrchGrpc
}

// GetOrchestratorAgentConfig возвращает конфигурацию агентов для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorAgentConfig() orchagent.Config {
	return c.OrchAgent
}

// GetOrchestratorPostgresConfig возвращает конфигурацию Postgres для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorPostgresConfig() orchpg.Config {
	return c.OrchDbPostgres
}

// GetOrchestratorPgxConfig возвращает конфигурацию pgx для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorPgxConfig() orchpgx.Config {
	return c.OrchDbPgx
}

// GetShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *OrchestratorConfig) GetShutdownConfig() shutdown.Config {
	return c.GracefulShutdown
}

// GetOrchestratorGRPCAddress возвращает адрес gRPC сервера оркестрации.
func (c *OrchestratorConfig) GetOrchestratorGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.OrchGrpc.Host, c.OrchGrpc.Port)
}

// GetConnectionURL возвращает URL-строку подключения для миграций.
func (c *OrchestratorConfig) GetConnectionURL() string {
	pg := c.OrchDbPostgres
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pg.User, pg.Password, pg.Host, pg.Port, pg.Database)
}

// GetDSN возвращает DSN-строку подключения для Postgres.
func (c *OrchestratorConfig) GetDSN() string {
	pg := c.OrchDbPostgres
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pg.Host, pg.Port, pg.User, pg.Password, pg.Database)
}

// GetAccessTokenTTL возвращает длительность жизни JWT access token.
func (c *OrchestratorConfig) GetAccessTokenTTL() time.Duration {
	return c.JWT.AccessTokenTTL
}

// GetRefreshTokenTTL возвращает длительность жизни JWT refresh token.
func (c *OrchestratorConfig) GetRefreshTokenTTL() time.Duration {
	return c.JWT.RefreshTokenTTL
}

// GetShutdownTimeout возвращает timeout для graceful shutdown.
func (c *OrchestratorConfig) GetShutdownTimeout() time.Duration {
	return c.GracefulShutdown.ShutdownTimeout
}

// GetAgentComputerPower возвращает количество агентов для вычислений.
func (c *OrchestratorConfig) GetAgentComputerPower() int {
	return c.OrchAgent.ComputerPower
}

// GetAgentOperationTimes возвращает времена выполнения операций для агентов.
func (c *OrchestratorConfig) GetAgentOperationTimes() map[string]time.Duration {
	return map[string]time.Duration{
		"addition":       c.OrchAgent.TimeAddition,
		"subtraction":    c.OrchAgent.TimeSubtraction,
		"multiplication": c.OrchAgent.TimeMultiplications,
		"division":       c.OrchAgent.TimeDivisions,
	}
}

// GetLoggerConfig возвращает конфигурацию журнала.
func (c *ServerConfig) GetLoggerConfig() logger.Config {
	return c.Logger
}

// GetServerConfig возвращает конфигурацию HTTP сервера.
func (c *ServerConfig) GetServerConfig() server.Config {
	return c.Server
}

// GetShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *ServerConfig) GetShutdownConfig() shutdown.Config {
	return c.GracefulShutdown
}

// GetServerAddress возвращает адрес HTTP сервера.
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetShutdownTimeout возвращает timeout для graceful shutdown.
func (c *ServerConfig) GetShutdownTimeout() time.Duration {
	return c.GracefulShutdown.ShutdownTimeout
}

// GetAuthGRPCConfig возвращает конфигурацию gRPC для сервиса авторизации.
func (c *ServerConfig) GetAuthGRPCConfig() struct {
	Host string
	Port int
} {
	return struct {
		Host string
		Port int
	}{
		Host: c.AuthGrpc.Host,
		Port: c.AuthGrpc.Port,
	}
}

// GetOrchestratorGRPCConfig возвращает конфигурацию gRPC для сервиса оркестрации.
func (c *ServerConfig) GetOrchestratorGRPCConfig() struct {
	Host string
	Port int
} {
	return struct {
		Host string
		Port int
	}{
		Host: c.OrchGrpc.Host,
		Port: c.OrchGrpc.Port,
	}
}

// GetMaxOperations возвращает максимальное количество операций в одном выражении.
func (c *OrchestratorConfig) GetMaxOperations() int {
	return c.OrchAgent.MaxOperations
}

// ToPostgresConfig converts AuthConfig's postgres config to database.PostgresConfig.
func (c *AuthConfig) ToPostgresConfig() database.PostgresConfig {
	return database.PostgresConfig{
		Host:            c.AuthDbPostgres.Host,
		Port:            c.AuthDbPostgres.Port,
		User:            c.AuthDbPostgres.User,
		Password:        c.AuthDbPostgres.Password,
		Database:        c.AuthDbPostgres.Database,
		SSLMode:         c.AuthDbPostgres.SSLMode,
		ApplicationName: c.AuthDbPostgres.ApplicationName,
		ConnTimeout:     c.AuthDbPostgres.ConnRetryInterval,
		MinConns:        c.AuthDbPgx.PoolMinConns,
		MaxConns:        c.AuthDbPgx.PoolMaxConns,
	}
}

// ToPostgresConfig converts OrchestratorConfig's postgres config to database.PostgresConfig.
func (c *OrchestratorConfig) ToPostgresConfig() database.PostgresConfig {
	return database.PostgresConfig{
		Host:            c.OrchDbPostgres.Host,
		Port:            c.OrchDbPostgres.Port,
		User:            c.OrchDbPostgres.User,
		Password:        c.OrchDbPostgres.Password,
		Database:        c.OrchDbPostgres.Database,
		SSLMode:         c.OrchDbPostgres.SSLMode,
		ApplicationName: c.OrchDbPostgres.ApplicationName,
		ConnTimeout:     c.OrchDbPostgres.ConnRetryInterval,
		MinConns:        c.OrchDbPgx.PoolMinConns,
		MaxConns:        c.OrchDbPgx.PoolMaxConns,
		MaxConnLifetime: c.OrchDbPgx.MaxConnLifetime,
		MaxConnIdleTime: c.OrchDbPgx.MaxConnIdleTime,
		HealthPeriod:    30 * time.Second,
	}
}
