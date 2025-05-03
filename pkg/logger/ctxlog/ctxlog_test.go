package ctxlog_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/ctxlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogger struct {
	debugCalled bool
	infoCalled  bool
	warnCalled  bool
	errorCalled bool
	fatalCalled bool
	syncCalled  bool
	lastMsg     string
	lastFields  []ctxlog.Field
	level       ctxlog.LogLevel
	withFields  []ctxlog.Field
	syncErr     error
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		level: ctxlog.InfoLevel,
	}
}

func (m *mockLogger) Debug(msg string, fields ...ctxlog.Field) {
	m.debugCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Info(msg string, fields ...ctxlog.Field) {
	m.infoCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Warn(msg string, fields ...ctxlog.Field) {
	m.warnCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Error(msg string, fields ...ctxlog.Field) {
	m.errorCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Fatal(msg string, fields ...ctxlog.Field) {
	m.fatalCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) With(fields ...ctxlog.Field) ctxlog.Logger {
	newLogger := newMockLogger()
	// Fix for appendAssign - copying fields explicitly instead of using append
	newLogger.withFields = make([]ctxlog.Field, len(m.withFields)+len(fields))
	copy(newLogger.withFields, m.withFields)
	copy(newLogger.withFields[len(m.withFields):], fields)
	newLogger.level = m.level
	return newLogger
}

func (m *mockLogger) SetLevel(level ctxlog.LogLevel) {
	m.level = level
}

func (m *mockLogger) GetLevel() ctxlog.LogLevel {
	return m.level
}

func (m *mockLogger) Sync() error {
	m.syncCalled = true
	return m.syncErr
}

var (
	errTest = errors.New("test error")
)

func TestWithLogger(t *testing.T) {
	t.Run("with valid logger and context", func(t *testing.T) {
		ctx := context.Background()
		logger := newMockLogger()

		newCtx := ctxlog.WithLogger(ctx, logger)
		require.NotEqual(t, ctx, newCtx, "WithLogger should return a new context")

		result, ok := ctxlog.FromContext(newCtx)
		assert.True(t, ok, "should find logger in context")
		assert.Equal(t, logger, result, "should return the same logger")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context
		logger := newMockLogger()

		newCtx := ctxlog.WithLogger(ctx, logger)
		assert.Equal(t, ctx, newCtx, "WithLogger should return the same context when ctx is nil")
	})

	t.Run("with nil logger", func(t *testing.T) {
		ctx := context.Background()
		var logger ctxlog.Logger

		newCtx := ctxlog.WithLogger(ctx, logger)
		assert.Equal(t, ctx, newCtx, "WithLogger should return the same context when logger is nil")
	})
}

func TestFromContext(t *testing.T) {
	t.Run("with logger in context", func(t *testing.T) {
		ctx := context.Background()
		logger := newMockLogger()
		ctx = ctxlog.WithLogger(ctx, logger)

		result, ok := ctxlog.FromContext(ctx)
		assert.True(t, ok, "should find logger in context")
		assert.Equal(t, logger, result, "should return the same logger")
	})

	t.Run("without logger in context", func(t *testing.T) {
		ctx := context.Background()

		result, ok := ctxlog.FromContext(ctx)
		assert.False(t, ok, "should not find logger in context")
		assert.Nil(t, result, "should return nil logger")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context

		result, ok := ctxlog.FromContext(ctx)
		assert.False(t, ok, "should not find logger in nil context")
		assert.Nil(t, result, "should return nil logger")
	})
}

func TestGetLogger(t *testing.T) {
	t.Run("with logger in context", func(t *testing.T) {
		ctx := context.Background()
		logger := newMockLogger()
		defaultLogger := newMockLogger()
		ctx = ctxlog.WithLogger(ctx, logger)

		result := ctxlog.GetLogger(ctx, defaultLogger)
		assert.Equal(t, logger, result, "should return logger from context")
	})

	t.Run("without logger in context", func(t *testing.T) {
		ctx := context.Background()
		defaultLogger := newMockLogger()

		result := ctxlog.GetLogger(ctx, defaultLogger)
		assert.Equal(t, defaultLogger, result, "should return default logger")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context
		defaultLogger := newMockLogger()

		result := ctxlog.GetLogger(ctx, defaultLogger)
		assert.Equal(t, defaultLogger, result, "should return default logger")
	})
}

func TestWithFields(t *testing.T) {
	t.Run("add fields to existing logger", func(t *testing.T) {
		ctx := context.Background()
		logger := newMockLogger()
		ctx = ctxlog.WithLogger(ctx, logger)

		field1 := "field1"
		field2 := "field2"
		newCtx := ctxlog.WithFields(ctx, logger, field1, field2)

		result, ok := ctxlog.FromContext(newCtx)
		require.True(t, ok, "should find logger in context")

		mockResult := result.(*mockLogger)

		require.Len(t, mockResult.withFields, 2, "should have 2 fields")
		assert.Equal(t, field1, mockResult.withFields[0], "should contain first field")
		assert.Equal(t, field2, mockResult.withFields[1], "should contain second field")
	})

	t.Run("with nil logger", func(t *testing.T) {
		ctx := context.Background()
		var defaultLogger ctxlog.Logger

		newCtx := ctxlog.WithFields(ctx, defaultLogger, "field")
		assert.Equal(t, ctx, newCtx, "WithFields should return original context when logger is nil")
	})
}

func TestLogg(t *testing.T) {
	tests := []struct {
		name    string
		level   ctxlog.LogLevel
		checkFn func(m *mockLogger) bool
	}{
		{
			name:    "debug level",
			level:   ctxlog.DebugLevel,
			checkFn: func(m *mockLogger) bool { return m.debugCalled },
		},
		{
			name:    "info level",
			level:   ctxlog.InfoLevel,
			checkFn: func(m *mockLogger) bool { return m.infoCalled },
		},
		{
			name:    "warn level",
			level:   ctxlog.WarnLevel,
			checkFn: func(m *mockLogger) bool { return m.warnCalled },
		},
		{
			name:    "error level",
			level:   ctxlog.ErrorLevel,
			checkFn: func(m *mockLogger) bool { return m.errorCalled },
		},
		{
			name:    "fatal level",
			level:   ctxlog.FatalLevel,
			checkFn: func(m *mockLogger) bool { return m.fatalCalled },
		},
		{
			name:    "unknown level defaults to info",
			level:   ctxlog.LogLevel(99),
			checkFn: func(m *mockLogger) bool { return m.infoCalled },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			logger := newMockLogger()

			ctxlog.Logg(ctx, logger, tc.level, "test message", "field1")

			assert.True(t, tc.checkFn(logger), "Logg should call the correct log method")
			assert.Equal(t, "test message", logger.lastMsg, "message should be set correctly")

			require.Len(t, logger.lastFields, 1, "should have one field")
			assert.Equal(t, "field1", logger.lastFields[0], "field should be set correctly")
		})
	}

	t.Run("with nil logger", func(t *testing.T) {
		ctx := context.Background()
		var logger ctxlog.Logger

		ctxlog.Logg(ctx, logger, ctxlog.InfoLevel, "test message")
	})
}

func TestLogMethods(t *testing.T) {
	tests := []struct {
		name    string
		logFn   func(ctx context.Context, logger ctxlog.Logger, msg string, fields ...ctxlog.Field)
		checkFn func(m *mockLogger) bool
	}{
		{
			name:    "Debug",
			logFn:   ctxlog.Debug,
			checkFn: func(m *mockLogger) bool { return m.debugCalled },
		},
		{
			name:    "Info",
			logFn:   ctxlog.Info,
			checkFn: func(m *mockLogger) bool { return m.infoCalled },
		},
		{
			name:    "Warn",
			logFn:   ctxlog.Warn,
			checkFn: func(m *mockLogger) bool { return m.warnCalled },
		},
		{
			name:    "Error",
			logFn:   ctxlog.Error,
			checkFn: func(m *mockLogger) bool { return m.errorCalled },
		},
		{
			name:    "Fatal",
			logFn:   ctxlog.Fatal,
			checkFn: func(m *mockLogger) bool { return m.fatalCalled },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			logger := newMockLogger()

			tc.logFn(ctx, logger, "test message", "field1")

			assert.True(t, tc.checkFn(logger), "should call the correct log method")
			assert.Equal(t, "test message", logger.lastMsg, "message should be set correctly")

			require.Len(t, logger.lastFields, 1, "should have one field")
			assert.Equal(t, "field1", logger.lastFields[0], "field should be set correctly")
		})
	}

	t.Run("with nil logger", func(t *testing.T) {
		ctx := context.Background()
		var logger ctxlog.Logger

		ctxlog.Debug(ctx, logger, "test message")
		ctxlog.Info(ctx, logger, "test message")
		ctxlog.Warn(ctx, logger, "test message")
		ctxlog.Error(ctx, logger, "test message")
		ctxlog.Fatal(ctx, logger, "test message")
	})
}

func TestSync(t *testing.T) {
	t.Run("successful sync", func(t *testing.T) {
		ctx := context.Background()
		logger := newMockLogger()

		err := ctxlog.Sync(ctx, logger)
		require.NoError(t, err)
		assert.True(t, logger.syncCalled, "Sync should call the logger's Sync method")
	})

	t.Run("sync with error", func(t *testing.T) {
		ctx := context.Background()
		logger := newMockLogger()
		logger.syncErr = errTest

		err := ctxlog.Sync(ctx, logger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), ctxlog.ErrSyncContextLogger, "error should contain context message")
		assert.ErrorIs(t, err, errTest, "error should wrap the original error")
	})

	t.Run("with nil logger", func(t *testing.T) {
		ctx := context.Background()
		var logger ctxlog.Logger

		err := ctxlog.Sync(ctx, logger)
		require.NoError(t, err, "Sync should handle nil logger gracefully")
	})
}

func TestLoggerInterface(t *testing.T) {
	t.Run("mock logger implements interface", func(t *testing.T) {
		var _ ctxlog.Logger = &mockLogger{}
	})
}

func TestErrSyncContextLogger(t *testing.T) {
	err := fmt.Errorf("%s: %w", ctxlog.ErrSyncContextLogger, errTest)
	assert.Contains(t, err.Error(), ctxlog.ErrSyncContextLogger)
	assert.ErrorIs(t, err, errTest)
}
