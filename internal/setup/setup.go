package setup

import (
	"fmt"
	"time"

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
)

// BaseConfig содержит общие поля для всех конфигураций.
type BaseConfig struct {
	Logger           logger.Config
	GracefulShutdown shutdown.Config
	JWT              jwt.Config
}

// getLoggerConfig возвращает конфигурацию журнала.
func (c *BaseConfig) getLoggerConfig() logger.Config {
	return c.Logger
}

// getJWTConfig возвращает конфигурацию JWT.
func (c *BaseConfig) getJWTConfig() jwt.Config {
	return c.JWT
}

// getShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *BaseConfig) getShutdownConfig() shutdown.Config {
	return c.GracefulShutdown
}

// getAccessTokenTTL возвращает время жизни access token.
func (c *BaseConfig) getAccessTokenTTL() time.Duration {
	return c.JWT.AccessTokenTTL
}

// getRefreshTokenTTL возвращает время жизни refresh token.
func (c *BaseConfig) getRefreshTokenTTL() time.Duration {
	return c.JWT.RefreshTokenTTL
}

// getShutdownTimeout возвращает timeout для graceful shutdown.
func (c *BaseConfig) getShutdownTimeout() time.Duration {
	return c.GracefulShutdown.ShutdownTimeout
}

type AuthConfig struct {
	BaseConfig
	Auth auth.Config
}

type OrchestratorConfig struct {
	BaseConfig
	Orchestrator orchestrator.Config
}

type ServerConfig struct {
	BaseConfig
	Server server.Config
}

// GetLoggerConfig возвращает конфигурацию журнала.
func (c *AuthConfig) GetLoggerConfig() logger.Config {
	return c.getLoggerConfig()
}

// GetJWTConfig возвращает конфигурацию JWT.
func (c *AuthConfig) GetJWTConfig() jwt.Config {
	return c.getJWTConfig()
}

// GetAuthGRPCConfig возвращает конфигурацию gRPC для сервиса авторизации.
func (c *AuthConfig) GetAuthGRPCConfig() authgrpc.Config {
	return c.Auth.Grpc
}

// GetAuthPostgresConfig возвращает конфигурацию Postgres для сервиса авторизации.
func (c *AuthConfig) GetAuthPostgresConfig() postgres.Config {
	return c.Auth.Db.Postgres
}

// GetAuthPgxConfig возвращает конфигурацию pgx для сервиса авторизации.
func (c *AuthConfig) GetAuthPgxConfig() authpgx.Config {
	return c.Auth.Db.Pgx
}

// GetAuthDBConfig возвращает конфигурацию базы данных для сервиса авторизации.
func (c *AuthConfig) GetAuthDBConfig() authdb.Config {
	return c.Auth.Db
}

// GetShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *AuthConfig) GetShutdownConfig() shutdown.Config {
	return c.getShutdownConfig()
}

// GetAuthGRPCAddress возвращает адрес gRPC сервера авторизации.
func (c *AuthConfig) GetAuthGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.Auth.Grpc.Host, c.Auth.Grpc.Port)
}

// GetConnectionURL возвращает URL-строку подключения для миграций.
func (c *AuthConfig) GetConnectionURL() string {
	pg := c.Auth.Db.Postgres
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pg.User, pg.Password, pg.Host, pg.Port, pg.Database)
}

// GetDSN возвращает DSN-строку подключения для Postgres.
func (c *AuthConfig) GetDSN() string {
	pg := c.Auth.Db.Postgres
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pg.Host, pg.Port, pg.User, pg.Password, pg.Database)
}

// GetAccessTokenTTL возвращает длительность жизни JWT access token.
func (c *AuthConfig) GetAccessTokenTTL() time.Duration {
	return c.getAccessTokenTTL()
}

// GetRefreshTokenTTL возвращает длительность жизни JWT refresh token.
func (c *AuthConfig) GetRefreshTokenTTL() time.Duration {
	return c.getRefreshTokenTTL()
}

// GetShutdownTimeout возвращает timeout для graceful shutdown.
func (c *AuthConfig) GetShutdownTimeout() time.Duration {
	return c.getShutdownTimeout()
}

// GetLoggerConfig возвращает конфигурацию журнал.
func (c *OrchestratorConfig) GetLoggerConfig() logger.Config {
	return c.getLoggerConfig()
}

// GetJWTConfig возвращает конфигурацию JWT.
func (c *OrchestratorConfig) GetJWTConfig() jwt.Config {
	return c.getJWTConfig()
}

// GetOrchestratorGRPCConfig возвращает конфигурацию gRPC для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorGRPCConfig() orchgrpc.Config {
	return c.Orchestrator.Grpc
}

// GetOrchestratorAgentConfig возвращает конфигурацию агентов для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorAgentConfig() orchagent.Config {
	return c.Orchestrator.Agent
}

// GetOrchestratorPostgresConfig возвращает конфигурацию Postgres для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorPostgresConfig() orchpg.Config {
	return c.Orchestrator.Db.Postgres
}

// GetOrchestratorPgxConfig возвращает конфигурацию pgx для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorPgxConfig() orchpgx.Config {
	return c.Orchestrator.Db.Pgx
}

// GetOrchestratorDBConfig возвращает конфигурацию базы данных для сервиса оркестрации.
func (c *OrchestratorConfig) GetOrchestratorDBConfig() orchdb.Config {
	return c.Orchestrator.Db
}

// GetShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *OrchestratorConfig) GetShutdownConfig() shutdown.Config {
	return c.getShutdownConfig()
}

// GetOrchestratorGRPCAddress возвращает адрес gRPC сервера оркестрации.
func (c *OrchestratorConfig) GetOrchestratorGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.Orchestrator.Grpc.Host, c.Orchestrator.Grpc.Port)
}

// GetConnectionURL возвращает URL-строку подключения для миграций.
func (c *OrchestratorConfig) GetConnectionURL() string {
	pg := c.Orchestrator.Db.Postgres
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pg.User, pg.Password, pg.Host, pg.Port, pg.Database)
}

// GetDSN возвращает DSN-строку подключения для Postgres.
func (c *OrchestratorConfig) GetDSN() string {
	pg := c.Orchestrator.Db.Postgres
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pg.Host, pg.Port, pg.User, pg.Password, pg.Database)
}

// GetAccessTokenTTL возвращает длительность жизни JWT access token.
func (c *OrchestratorConfig) GetAccessTokenTTL() time.Duration {
	return c.getAccessTokenTTL()
}

// GetRefreshTokenTTL возвращает длительность жизни JWT refresh token.
func (c *OrchestratorConfig) GetRefreshTokenTTL() time.Duration {
	return c.getRefreshTokenTTL()
}

// GetShutdownTimeout возвращает timeout для graceful shutdown.
func (c *OrchestratorConfig) GetShutdownTimeout() time.Duration {
	return c.getShutdownTimeout()
}

// GetAgentComputerPower возвращает количество агентов для вычислений.
func (c *OrchestratorConfig) GetAgentComputerPower() int {
	return c.Orchestrator.Agent.ComputerPower
}

// GetAgentOperationTimes возвращает времена выполнения операций для агентов.
func (c *OrchestratorConfig) GetAgentOperationTimes() map[string]time.Duration {
	return map[string]time.Duration{
		"addition":       c.Orchestrator.Agent.TimeAddition,
		"subtraction":    c.Orchestrator.Agent.TimeSubtraction,
		"multiplication": c.Orchestrator.Agent.TimeMultiplications,
		"division":       c.Orchestrator.Agent.TimeDivisions,
	}
}

// GetLoggerConfig возвращает конфигурацию журнал.
func (c *ServerConfig) GetLoggerConfig() logger.Config {
	return c.getLoggerConfig()
}

// GetServerConfig возвращает конфигурацию HTTP сервера.
func (c *ServerConfig) GetServerConfig() server.Config {
	return c.Server
}

// GetShutdownConfig возвращает конфигурацию graceful shutdown.
func (c *ServerConfig) GetShutdownConfig() shutdown.Config {
	return c.getShutdownConfig()
}

// GetServerAddress возвращает адрес HTTP сервера.
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetShutdownTimeout возвращает timeout для graceful shutdown.
func (c *ServerConfig) GetShutdownTimeout() time.Duration {
	return c.getShutdownTimeout()
}
