// Package postgres содержит конфигурацию для PostgreSQL.
package postgres

import "time"

// Config содержит конфигурацию для PostgreSQL.
type Config struct {
	Host              string        `yaml:"host" env:"AUTH_POSTGRES_DB_HOST" env-default:"auth-db"`
	Port              int           `yaml:"port" env:"AUTH_POSTGRES_DB_PORT" env-default:"5432"`
	User              string        `yaml:"user" env:"AUTH_POSTGRES_DB_USER" env-default:"auth"`
	Password          string        `yaml:"password" env:"AUTH_POSTGRES_DB_PASSWORD" env-default:"auth"`
	Database          string        `yaml:"database" env:"AUTH_POSTGRES_DB_NAME" env-default:"auth"`
	SSLMode           string        `yaml:"sslmode" env:"AUTH_POSTGRES_DB_SSL_MODE" env-default:"disable"`
	ConnRetry         int           `yaml:"timeout" env:"AUTH_POSTGRES_DB_CONNECT_RETRY" env-default:"3"`
	ConnRetryInterval time.Duration `yaml:"timeout_interval" env:"AUTH_POSTGRES_DB_CONNECT_RETRY_INTERVAL" env-default:"5s"`
	StatementTimeout  time.Duration `yaml:"statement_timeout" env:"AUTH_POSTGRES_DB_STATEMENT_TIMEOUT" env-default:"60s"`
	ApplicationName   string        `yaml:"application_name" env:"AUTH_POSTGRES_DB_APPLICATION_NAME" env-default:"auth-service"`
}
