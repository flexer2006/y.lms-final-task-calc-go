// Package postgres предоставляет функциональность для работы с PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Константы для сообщений об ошибках и логирования.
const (
	errInvalidConfig     = "invalid database configuration"
	errParseConfig       = "failed to parse database configuration"
	errCreateConnPool    = "failed to create connection pool"
	errPingDatabase      = "failed to ping database"
	errAcquireConn       = "failed to acquire connection from pool"
	errNilConnectionPool = "connection pool is nil"

	logConnecting         = "connecting to postgres database"
	logConnectingDSN      = "connecting to postgres database using DSN"
	logConnected          = "connected to postgres database"
	logClosing            = "closing postgres database connection"
	logMinConnsExceedsMax = "MinConns value exceeds maximum allowed value, setting to max int32"
	logMaxConnsExceedsMax = "MaxConns value exceeds maximum allowed value, setting to max int32"
)

// Статические ошибки для проверки конфигурации.
var (
	ErrHostRequired     = errors.New("database host is required")
	ErrInvalidPort      = errors.New("invalid database port")
	ErrUserRequired     = errors.New("database user is required")
	ErrDatabaseRequired = errors.New("database name is required")
)

// Config хранит параметры для подключения к базе данных PostgreSQL.
type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	ApplicationName string
	ConnTimeout     time.Duration
	MinConns        int
	MaxConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthPeriod    time.Duration
}

// Validate проверяет правильность конфигурации базы данных.
func (c PostgresConfig) Validate() error {
	if c.Host == "" {
		return ErrHostRequired
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("%w: %d", ErrInvalidPort, c.Port)
	}

	if c.User == "" {
		return ErrUserRequired
	}

	if c.Database == "" {
		return ErrDatabaseRequired
	}

	return nil
}

// DSN возвращает строку подключения к базе данных.
func (c PostgresConfig) DSN() string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.Database)

	params := make([]string, 0)

	if c.SSLMode != "" {
		params = append(params, fmt.Sprintf("sslmode=%s", c.SSLMode))
	}

	if c.ApplicationName != "" {
		params = append(params, fmt.Sprintf("application_name=%s", c.ApplicationName))
	}

	if len(params) > 0 {
		dsn += "?" + params[0]
		for i := 1; i < len(params); i++ {
			dsn += "&" + params[i]
		}
	}

	return dsn
}

// Database представляет соединение с PostgreSQL.
type Database struct {
	pool   *pgxpool.Pool
	config PostgresConfig
}

// New создает новое соединение с базой данных PostgreSQL.
func New(ctx context.Context, config PostgresConfig) (*Database, error) {
	if err := config.Validate(); err != nil {
		logger.Error(ctx, nil, errInvalidConfig, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errInvalidConfig, err)
	}

	dsn := config.DSN()

	logger.Info(ctx, nil, logConnecting,
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database),
		zap.String("user", config.User))

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error(ctx, nil, errParseConfig, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errParseConfig, err)
	}

	if config.MinConns > 0 {
		if config.MinConns > math.MaxInt32 {
			logger.Warn(ctx, nil, logMinConnsExceedsMax)
			poolCfg.MinConns = math.MaxInt32
		} else {
			poolCfg.MinConns = int32(config.MinConns)
		}
	}

	if config.MaxConns > 0 {
		if config.MaxConns > math.MaxInt32 {
			logger.Warn(ctx, nil, logMaxConnsExceedsMax)
			poolCfg.MaxConns = math.MaxInt32
		} else {
			poolCfg.MaxConns = int32(config.MaxConns)
		}
	}

	if config.ConnTimeout > 0 {
		poolCfg.ConnConfig.ConnectTimeout = config.ConnTimeout
	} else {
		poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second
	}

	if config.HealthPeriod > 0 {
		poolCfg.HealthCheckPeriod = config.HealthPeriod
	} else {
		poolCfg.HealthCheckPeriod = 1 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error(ctx, nil, errCreateConnPool, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errCreateConnPool, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error(ctx, nil, errPingDatabase, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errPingDatabase, err)
	}

	logger.Info(ctx, nil, logConnected,
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database))

	return &Database{
		pool:   pool,
		config: config,
	}, nil
}

// NewWithDSN создает новое соединение с базой данных по DSN.
func NewWithDSN(ctx context.Context, dsn string, minConn, maxConn int) (*Database, error) {
	logger.Info(ctx, nil, logConnectingDSN)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error(ctx, nil, errParseConfig, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errParseConfig, err)
	}

	if minConn > 0 {
		if minConn > math.MaxInt32 {
			logger.Warn(ctx, nil, logMinConnsExceedsMax)
			poolCfg.MinConns = math.MaxInt32
		} else {
			poolCfg.MinConns = int32(minConn)
		}
	}

	if maxConn > 0 {
		if maxConn > math.MaxInt32 {
			logger.Warn(ctx, nil, logMaxConnsExceedsMax)
			poolCfg.MaxConns = math.MaxInt32
		} else {
			poolCfg.MaxConns = int32(maxConn)
		}
	}

	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error(ctx, nil, errCreateConnPool, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errCreateConnPool, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error(ctx, nil, errPingDatabase, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errPingDatabase, err)
	}

	logger.Info(ctx, nil, logConnected)

	config := PostgresConfig{
		MinConns: minConn,
		MaxConns: maxConn,
	}

	return &Database{
		pool:   pool,
		config: config,
	}, nil
}

// Pool возвращает пул соединений с базой данных.
func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

// Close закрывает соединение с базой данных.
func (db *Database) Close(ctx context.Context) {
	logger.Info(ctx, nil, logClosing)
	db.pool.Close()
}

// ErrConnectionPoolNil - ошибка отсутствия пула соединений.
var ErrConnectionPoolNil = errors.New("connection pool is nil")

// Ping проверяет доступность базы данных.
func (db *Database) Ping(ctx context.Context) error {
	if db.pool == nil {
		err := ErrConnectionPoolNil
		return fmt.Errorf("%s: %w", errPingDatabase, err)
	}

	if err := db.pool.Ping(ctx); err != nil {
		logger.Error(ctx, nil, errPingDatabase, zap.Error(err))
		return fmt.Errorf("%s: %w", errPingDatabase, err)
	}

	return nil
}

// Config возвращает конфигурацию базы данных.
func (db *Database) Config() PostgresConfig {
	return db.config
}

// GetDSN возвращает строку подключения к базе данных.
func (db *Database) GetDSN() string {
	return db.config.DSN()
}

// AcquireConn получает подключение из пула.
func (db *Database) AcquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		logger.Error(ctx, nil, errAcquireConn, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errAcquireConn, err)
	}
	return conn, nil
}
