// Package pgxx содержит конфигурацию для pgx.
package pgxx

import "time"

// Config содержит конфигурацию для pgx.
type Config struct {
	PoolMaxConns    int           `yaml:"pool_max_conns" env:"AUTH_PGX_POOL_MAX_CONNS" env-default:"10"`
	PoolMinConns    int           `yaml:"pool_min_conns" env:"AUTH_PGX_POOL_MIN_CONNS" env-default:"1"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout" env:"AUTH_PGX_CONNECT_TIMEOUT" env-default:"10s"`
	AcquireTimeout  time.Duration `yaml:"acquire_timeout" env:"AUTH_PGX_POOL_ACQUIRE_TIMEOUT" env-default:"60s"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env:"AUTH_PGX_POOL_MAX_CONN_LIFETIME" env-default:"3600s"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env:"AUTH_PGX_POOL_MAX_CONN_IDLE_TIME" env-default:"600s"`
	PoolLifetime    time.Duration `yaml:"pool_lifetime" env:"AUTH_PGX_POOL_LIFETIME" env-default:"3600s"`
	MigratePath     string        `yaml:"migrate_path" env:"AUTH_MIGRATIONS_DIR" env-default:"./migrations/auth"`
}
