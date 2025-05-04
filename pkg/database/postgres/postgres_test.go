package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/postgres"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidConfig() postgres.Config {
	return postgres.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		Database:        "testdb",
		SSLMode:         "disable",
		MinConns:        2,
		MaxConns:        10,
		ConnTimeout:     5 * time.Second,
		HealthPeriod:    1 * time.Minute,
		ApplicationName: "test-app",
	}
}

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		name         string
		modifyConfig func(*postgres.Config)
		expectedErr  error
	}{
		{
			name:         "Valid config",
			modifyConfig: func(c *postgres.Config) {},
			expectedErr:  nil,
		},
		{
			name:         "Missing host",
			modifyConfig: func(c *postgres.Config) { c.Host = "" },
			expectedErr:  postgres.ErrHostRequired,
		},
		{
			name:         "Invalid port - too low",
			modifyConfig: func(c *postgres.Config) { c.Port = 0 },
			expectedErr:  postgres.ErrInvalidPort,
		},
		{
			name:         "Invalid port - too high",
			modifyConfig: func(c *postgres.Config) { c.Port = 65536 },
			expectedErr:  postgres.ErrInvalidPort,
		},
		{
			name:         "Missing user",
			modifyConfig: func(c *postgres.Config) { c.User = "" },
			expectedErr:  postgres.ErrUserRequired,
		},
		{
			name:         "Missing database",
			modifyConfig: func(c *postgres.Config) { c.Database = "" },
			expectedErr:  postgres.ErrDatabaseRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createValidConfig()
			tc.modifyConfig(&cfg)
			err := cfg.Validate()

			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigDSN(t *testing.T) {
	testCases := []struct {
		name     string
		config   postgres.Config
		expected string
	}{
		{
			name: "Basic DSN",
			config: postgres.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Database: "testdb",
			},
			expected: "postgres://postgres:password@localhost:5432/testdb",
		},
		{
			name: "DSN with SSL mode",
			config: postgres.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "postgres://postgres:password@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "DSN with application name",
			config: postgres.Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "testdb",
				ApplicationName: "test-app",
			},
			expected: "postgres://postgres:password@localhost:5432/testdb?application_name=test-app",
		},
		{
			name: "DSN with multiple params",
			config: postgres.Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "testdb",
				SSLMode:         "disable",
				ApplicationName: "test-app",
			},
			expected: "postgres://postgres:password@localhost:5432/testdb?sslmode=disable&application_name=test-app",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.config.DSN())
		})
	}

	t.Run("DSN with special characters", func(t *testing.T) {
		cfg := postgres.Config{
			Host:     "localhost",
			Port:     5432,
			User:     "post:gres",
			Password: "p@ss:word",
			Database: "test@db",
		}
		dsn := cfg.DSN()

		assert.Contains(t, dsn, "post:gres")
		assert.Contains(t, dsn, "p@ss:word")
		assert.Contains(t, dsn, "test@db")
	})
}

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...logger.Field)  {}
func (m *mockLogger) Info(msg string, fields ...logger.Field)   {}
func (m *mockLogger) Warn(msg string, fields ...logger.Field)   {}
func (m *mockLogger) Error(msg string, fields ...logger.Field)  {}
func (m *mockLogger) Fatal(msg string, fields ...logger.Field)  {}
func (m *mockLogger) With(fields ...logger.Field) logger.Logger { return m }
func (m *mockLogger) SetLevel(level logger.LogLevel)            {}
func (m *mockLogger) GetLevel() logger.LogLevel                 { return logger.InfoLevel }
func (m *mockLogger) Sync() error                               { return nil }

func setupLoggerContext() context.Context {
	ctx := context.Background()
	return logger.WithLogger(ctx, &mockLogger{})
}

func TestIntegration_Full_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupLoggerContext()

	cfg := postgres.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "auth",
		Password:        "auth",
		Database:        "auth",
		SSLMode:         "disable",
		MinConns:        1,
		MaxConns:        5,
		ConnTimeout:     5 * time.Second,
		HealthPeriod:    1 * time.Minute,
		ApplicationName: "postgres-test",
	}

	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(ctx)

	returnedCfg := db.Config()
	assert.Equal(t, cfg, returnedCfg)

	dsn := db.GetDSN()
	assert.Equal(t, cfg.DSN(), dsn)

	pool := db.Pool()
	assert.NotNil(t, pool)

	err = db.Ping(ctx)
	require.NoError(t, err)

	conn, err := db.AcquireConn(ctx)
	require.NoError(t, err)
	assert.NotNil(t, conn)
	conn.Release()
}

func TestIntegration_NewWithDSN(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupLoggerContext()

	dsn := "postgres://auth:auth@localhost:5432/auth?sslmode=disable" //nolint:gosec

	db, err := postgres.NewWithDSN(ctx, dsn, 1, 5)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(ctx)

	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestNew_ValidationError(t *testing.T) {
	ctx := setupLoggerContext()

	invalidConfigs := []struct {
		name   string
		config postgres.Config
	}{
		{
			name: "Empty host",
			config: postgres.Config{
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Database: "testdb",
			},
		},
		{
			name: "Invalid port",
			config: postgres.Config{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Password: "password",
				Database: "testdb",
			},
		},
		{
			name: "Empty user",
			config: postgres.Config{
				Host:     "localhost",
				Port:     5432,
				Password: "password",
				Database: "testdb",
			},
		},
		{
			name: "Empty database",
			config: postgres.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
			},
		},
	}

	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			db, err := postgres.New(ctx, tc.config)
			require.Error(t, err)
			assert.Nil(t, db)
			assert.Contains(t, err.Error(), "invalid database configuration")
		})
	}
}

func TestNew_ConnectionError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection error test in short mode")
	}

	ctx := setupLoggerContext()

	cfg := postgres.Config{
		Host:        "nonexistenthost",
		Port:        5432,
		User:        "postgres",
		Password:    "postgres",
		Database:    "postgres",
		SSLMode:     "disable",
		ConnTimeout: 1 * time.Second,
	}

	db, err := postgres.New(ctx, cfg)
	require.Error(t, err)
	assert.Nil(t, db)

	assert.Contains(t, err.Error(), "failed to ping database")
}

func TestErrConnectionPoolNil(t *testing.T) {
	db := &postgres.Database{}
	ctx := context.Background()

	err := db.Ping(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, errors.Unwrap(err), postgres.ErrConnectionPoolNil)
}

func TestDatabase_AcquireConn_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := setupLoggerContext()

	cfg := postgres.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "auth",
		Password: "auth",
		Database: "auth",
		SSLMode:  "disable",
	}

	db, err := postgres.New(ctx, cfg)
	require.NoError(t, err)
	defer db.Close(ctx)

	conn, err := db.AcquireConn(ctx)
	require.NoError(t, err)
	assert.NotNil(t, conn)

	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)

	conn.Release()
}
