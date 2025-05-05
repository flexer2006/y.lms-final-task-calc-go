package postgres

import "time"

type Config struct {
	Host              string        `yaml:"host" env:"ORCHESTRATOR_POSTGRES_DB_HOST" env-default:"orchestrator-db"`
	Port              int           `yaml:"port" env:"ORCHESTRATOR_POSTGRES_DB_PORT" env-default:"5433"`
	User              string        `yaml:"user" env:"ORCHESTRATOR_POSTGRES_DB_USER" env-default:"orchestrator"`
	Password          string        `yaml:"password" env:"ORCHESTRATOR_POSTGRES_DB_PASSWORD" env-default:"orchestrator"`
	Database          string        `yaml:"database" env:"ORCHESTRATOR_POSTGRES_DB_NAME" env-default:"orchestrator"`
	SSLMode           string        `yaml:"sslmode" env:"ORCHESTRATOR_POSTGRES_DB_SSL_MODE" env-default:"disable"`
	ConnRetry         int           `yaml:"timeout" env:"ORCHESTRATOR_POSTGRES_DB_CONNECT_RETRY" env-default:"3"`
	ConnRetryInterval time.Duration `yaml:"timeout_interval" env:"ORCHESTRATOR_POSTGRES_DB_CONNECT_RETRY_INTERVAL" env-default:"5s"`
	StatementTimeout  time.Duration `yaml:"statement_timeout" env:"ORCHESTRATOR_POSTGRES_DB_STATEMENT_TIMEOUT" env-default:"60s"`
	ApplicationName   string        `yaml:"application_name" env:"ORCHESTRATOR_POSTGRES_DB_APPLICATION_NAME" env-default:"orchestrator-service"`
}
