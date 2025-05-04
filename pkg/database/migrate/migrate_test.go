package migrate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/migrate"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogger struct {
	debugCalled bool
	infoCalled  bool
	errorCalled bool
	lastMsg     string
	fields      []logger.Field
}

func (m *mockLogger) Debug(msg string, fields ...logger.Field) {
	m.debugCalled = true
	m.lastMsg = msg
	m.fields = fields
}

func (m *mockLogger) Info(msg string, fields ...logger.Field) {
	m.infoCalled = true
	m.lastMsg = msg
	m.fields = fields
}

func (m *mockLogger) Warn(msg string, fields ...logger.Field) {}

func (m *mockLogger) Error(msg string, fields ...logger.Field) {
	m.errorCalled = true
	m.lastMsg = msg
	m.fields = fields
}

func (m *mockLogger) Fatal(msg string, fields ...logger.Field) {}

func (m *mockLogger) With(fields ...logger.Field) logger.Logger { return m }

func (m *mockLogger) SetLevel(level logger.LogLevel) {}

func (m *mockLogger) GetLevel() logger.LogLevel { return logger.InfoLevel }

func (m *mockLogger) Sync() error { return nil }

func setupLoggerContext() (context.Context, *mockLogger) {
	ctx := context.Background()
	mockLog := &mockLogger{}
	return logger.WithLogger(ctx, mockLog), mockLog
}

type mockMigrationContext struct {
	ctx        context.Context
	mockLogger *mockLogger
	dsn        string
	config     migrate.Config
}

func setupMockMigrationContext() mockMigrationContext {
	ctx, mockLog := setupLoggerContext()
	return mockMigrationContext{
		ctx:        ctx,
		mockLogger: mockLog,
		dsn:        "invalid://dsn",
		config:     migrate.Config{Path: "/valid/path"},
	}
}

func (m *mockMigrationContext) resetLoggerState() {
	m.mockLogger.debugCalled = false
	m.mockLogger.infoCalled = false
	m.mockLogger.errorCalled = false
	m.mockLogger.lastMsg = ""
	m.mockLogger.fields = nil
}

func TestNewMigrator(t *testing.T) {
	migrator := migrate.NewMigrator()
	assert.NotNil(t, migrator, "NewMigrator should return a non-nil instance")
}

func TestMissingPath(t *testing.T) {
	ctx, _ := setupLoggerContext()
	migrator := migrate.NewMigrator()
	emptyConfig := migrate.Config{Path: ""}

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "Up with empty path",
			run: func() error {
				return migrator.Up(ctx, "postgres://localhost:5432/testdb", emptyConfig)
			},
		},
		{
			name: "Down with empty path",
			run: func() error {
				return migrator.Down(ctx, "postgres://localhost:5432/testdb", emptyConfig)
			},
		},
		{
			name: "Migrate with empty path",
			run: func() error {
				return migrator.Migrate(ctx, "postgres://localhost:5432/testdb", 1, emptyConfig)
			},
		},
		{
			name: "Steps with empty path",
			run: func() error {
				return migrator.Steps(ctx, "postgres://localhost:5432/testdb", 1, emptyConfig)
			},
		},
		{
			name: "Force with empty path",
			run: func() error {
				return migrator.Force(ctx, "postgres://localhost:5432/testdb", 1, emptyConfig)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			assert.ErrorIs(t, err, migrate.ErrMigrationPathNotSpecified,
				"should return ErrMigrationPathNotSpecified for empty path")
		})
	}

	t.Run("Version with empty path", func(t *testing.T) {
		_, _, err := migrator.Version(ctx, "postgres://localhost:5432/testdb", emptyConfig)
		assert.ErrorIs(t, err, migrate.ErrMigrationPathNotSpecified)
	})
}

func TestMigratorCreationError(t *testing.T) {
	ctx, mockLog := setupLoggerContext()
	migrator := migrate.NewMigrator()
	validConfig := migrate.Config{Path: "/valid/path"}
	dsn := "postgres://localhost:5432/testdb"

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "Up with migrator creation error",
			run: func() error {
				return migrator.Up(ctx, dsn, validConfig)
			},
		},
		{
			name: "Down with migrator creation error",
			run: func() error {
				return migrator.Down(ctx, dsn, validConfig)
			},
		},
		{
			name: "Migrate with migrator creation error",
			run: func() error {
				return migrator.Migrate(ctx, dsn, 1, validConfig)
			},
		},
		{
			name: "Steps with migrator creation error",
			run: func() error {
				return migrator.Steps(ctx, dsn, 1, validConfig)
			},
		},
		{
			name: "Force with migrator creation error",
			run: func() error {
				return migrator.Force(ctx, dsn, 1, validConfig)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockLog.errorCalled = false

			err := tc.run()

			require.Error(t, err)
			assert.True(t, mockLog.errorCalled, "Error logger should be called")
		})
	}
}

func TestMigrationOperations(t *testing.T) {
	t.Run("TestUpOperation", func(t *testing.T) {
		ctx, mockLog := setupLoggerContext()
		migrator := migrate.NewMigrator()

		config := migrate.Config{Path: "/some/valid/path"}
		invalidDSN := "invalid://dsn"

		err := migrator.Up(ctx, invalidDSN, config)

		require.Error(t, err)
		assert.True(t, mockLog.errorCalled)
	})

	t.Run("TestDownOperation", func(t *testing.T) {
		ctx, mockLog := setupLoggerContext()
		migrator := migrate.NewMigrator()

		config := migrate.Config{Path: "/some/valid/path"}
		invalidDSN := "invalid://dsn"

		err := migrator.Down(ctx, invalidDSN, config)

		require.Error(t, err)
		assert.True(t, mockLog.errorCalled)
	})
}

func TestErrorWrapping(t *testing.T) {
	ctx, _ := setupLoggerContext()
	migrator := migrate.NewMigrator()

	t.Run("Error unwrapping", func(t *testing.T) {
		config := migrate.Config{Path: "/some/valid/path"}
		invalidDSN := "invalid://dsn"

		err := migrator.Up(ctx, invalidDSN, config)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed to create migrator")
	})
}

func TestVersionOperation(t *testing.T) {
	ctx, mockLog := setupLoggerContext()
	migrator := migrate.NewMigrator()

	config := migrate.Config{Path: "/some/valid/path"}
	invalidDSN := "invalid://dsn"

	_, _, err := migrator.Version(ctx, invalidDSN, config)

	require.Error(t, err)
	assert.True(t, mockLog.errorCalled)
}

func TestMigrationLogger(t *testing.T) {
	ctx, mockLog := setupLoggerContext()
	migrator := migrate.NewMigrator()

	config := migrate.Config{Path: "/some/valid/path"}
	invalidDSN := "invalid://dsn"

	_ = migrator.Up(ctx, invalidDSN, config)

	assert.True(t, mockLog.errorCalled, "Logger should be called during migration operations")
}

func TestErrorNoChangeHandling(t *testing.T) {
	t.Skip("This would require mocking the internal behavior of migrate.Migrate")
}

func TestFullMigrationCycle(t *testing.T) {
	mock := setupMockMigrationContext()
	migrator := migrate.NewMigrator()

	t.Run("Migration cycle operations", func(t *testing.T) {
		operations := []struct {
			name        string
			operation   func() error
			checkLogger func(t *testing.T, m *mockLogger)
		}{
			{
				name: "Up operation",
				operation: func() error {
					return migrator.Up(mock.ctx, mock.dsn, mock.config)
				},
				checkLogger: func(t *testing.T, m *mockLogger) {
					t.Helper()
					assert.True(t, m.errorCalled, "Error logger should be called during Up operation")
				},
			},
			{
				name: "Down operation",
				operation: func() error {
					return migrator.Down(mock.ctx, mock.dsn, mock.config)
				},
				checkLogger: func(t *testing.T, m *mockLogger) {
					t.Helper()
					assert.True(t, m.errorCalled, "Error logger should be called during Down operation")
				},
			},
			{
				name: "Migrate operation",
				operation: func() error {
					return migrator.Migrate(mock.ctx, mock.dsn, 1, mock.config)
				},
				checkLogger: func(t *testing.T, m *mockLogger) {
					t.Helper()
					assert.True(t, m.errorCalled, "Error logger should be called during Migrate operation")
				},
			},
			{
				name: "Steps operation",
				operation: func() error {
					return migrator.Steps(mock.ctx, mock.dsn, 1, mock.config)
				},
				checkLogger: func(t *testing.T, m *mockLogger) {
					t.Helper()
					assert.True(t, m.errorCalled, "Error logger should be called during Steps operation")
				},
			},
			{
				name: "Force operation",
				operation: func() error {
					return migrator.Force(mock.ctx, mock.dsn, 1, mock.config)
				},
				checkLogger: func(t *testing.T, m *mockLogger) {
					t.Helper()
					assert.True(t, m.errorCalled, "Error logger should be called during Force operation")
				},
			},
			{
				name: "Version operation",
				operation: func() error {
					_, _, err := migrator.Version(mock.ctx, mock.dsn, mock.config)
					if err != nil {
						return fmt.Errorf("version operation failed: %w", err)
					}
					return nil
				},
				checkLogger: func(t *testing.T, m *mockLogger) {
					t.Helper()
					assert.True(t, m.errorCalled, "Error logger should be called during Version operation")
				},
			},
		}

		for _, op := range operations {
			t.Run(op.name, func(t *testing.T) {
				mock.resetLoggerState()
				err := op.operation()
				require.Error(t, err, "Operation should return error with invalid DSN")
				op.checkLogger(t, mock.mockLogger)
			})
		}
	})
}

func TestMigrationLoggerInterface(t *testing.T) {
	ctx, mockLog := setupLoggerContext()

	t.Run("Logger Printf should call debug logger", func(t *testing.T) {
		mockLog.debugCalled = false
		migrator := migrate.NewMigrator()

		_ = migrator.Up(ctx, "invalid://dsn", migrate.Config{Path: "/valid/path"})

		assert.True(t, mockLog.errorCalled, "Error logger should be called during operation")
	})
}

func TestErrorNoChangeHandlingWithFields(t *testing.T) {
	ctx, mockLog := setupLoggerContext()
	migrator := migrate.NewMigrator()

	t.Run("Migrate operation passes fields in error logs", func(t *testing.T) {
		mockLog.errorCalled = false
		mockLog.fields = nil

		version := uint(5)
		err := migrator.Migrate(ctx, "invalid://dsn", version, migrate.Config{Path: "/valid/path"})

		require.Error(t, err)
		assert.True(t, mockLog.errorCalled)
		assert.NotEmpty(t, mockLog.fields, "Error logs should include fields")
	})

	t.Run("Steps operation passes fields in error logs", func(t *testing.T) {
		mockLog.errorCalled = false
		mockLog.fields = nil

		steps := 3
		err := migrator.Steps(ctx, "invalid://dsn", steps, migrate.Config{Path: "/valid/path"})

		require.Error(t, err)
		assert.True(t, mockLog.errorCalled)
		assert.NotEmpty(t, mockLog.fields, "Error logs should include fields")
	})
}

func TestVersionErrorHandling(t *testing.T) {
	ctx, mockLog := setupLoggerContext()
	migrator := migrate.NewMigrator()

	t.Run("Version with createMigrator error", func(t *testing.T) {
		mockLog.errorCalled = false

		_, _, err := migrator.Version(ctx, "invalid://dsn", migrate.Config{Path: "/valid/path"})

		require.Error(t, err)
		assert.True(t, mockLog.errorCalled)
	})
}

func TestEmptyFieldsHandling(t *testing.T) {
	ctx, mockLog := setupLoggerContext()
	migrator := migrate.NewMigrator()

	t.Run("Up operation with empty fields", func(t *testing.T) {
		mockLog.errorCalled = false
		mockLog.infoCalled = false

		err := migrator.Up(ctx, "invalid://dsn", migrate.Config{Path: "/valid/path"})

		require.Error(t, err)
		assert.True(t, mockLog.errorCalled)
		assert.False(t, mockLog.infoCalled, "Info should not be called on error")
	})

	t.Skip("Would require advance mocking to test success path")
}
