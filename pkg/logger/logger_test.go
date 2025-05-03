package logger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type mockCtxLogger struct {
	debugCalled bool
	infoCalled  bool
	warnCalled  bool
	errorCalled bool
	fatalCalled bool
	syncCalled  bool
	lastMsg     string
	lastFields  []logger.Field
	level       logger.LogLevel
	withFields  []logger.Field
	syncErr     error
}

func newMockCtxLogger() *mockCtxLogger {
	return &mockCtxLogger{
		level: logger.InfoLevel,
	}
}

func (m *mockCtxLogger) Debug(msg string, fields ...logger.Field) {
	m.debugCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockCtxLogger) Info(msg string, fields ...logger.Field) {
	m.infoCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockCtxLogger) Warn(msg string, fields ...logger.Field) {
	m.warnCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockCtxLogger) Error(msg string, fields ...logger.Field) {
	m.errorCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockCtxLogger) Fatal(msg string, fields ...logger.Field) {
	m.fatalCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockCtxLogger) With(fields ...logger.Field) logger.Logger {
	newLogger := newMockCtxLogger()
	newLogger.withFields = make([]logger.Field, len(m.withFields)+len(fields))
	copy(newLogger.withFields, m.withFields)
	copy(newLogger.withFields[len(m.withFields):], fields)
	newLogger.level = m.level
	return newLogger
}

func (m *mockCtxLogger) SetLevel(level logger.LogLevel) {
	m.level = level
}

func (m *mockCtxLogger) GetLevel() logger.LogLevel {
	return m.level
}

func (m *mockCtxLogger) Sync() error {
	m.syncCalled = true
	return m.syncErr
}

type mockZapLogger struct {
	*mockCtxLogger
	rawLogger *zap.Logger
}

func newMockZapLogger() *mockZapLogger {
	return &mockZapLogger{
		mockCtxLogger: newMockCtxLogger(),
		rawLogger:     zap.NewNop(),
	}
}

func (m *mockZapLogger) RawLogger() *zap.Logger {
	return m.rawLogger
}

func (m *mockZapLogger) With(fields ...logger.Field) logger.Logger {
	newLogger := newMockZapLogger()
	newLogger.withFields = make([]logger.Field, len(m.withFields)+len(fields))
	copy(newLogger.withFields, m.withFields)
	copy(newLogger.withFields[len(m.withFields):], fields)
	newLogger.level = m.level
	return newLogger
}

var errTest = errors.New("test error")

func TestZapAdapter(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(l logger.ZapLogger, msg string, fields ...logger.Field)
		checkLog func(logs []observer.LoggedEntry) bool
	}{
		{
			name: "Debug",
			logFunc: func(l logger.ZapLogger, msg string, fields ...logger.Field) {
				l.Debug(msg, fields...)
			},
			checkLog: func(logs []observer.LoggedEntry) bool {
				return len(logs) > 0 && logs[0].Level == zapcore.DebugLevel
			},
		},
		{
			name: "Info",
			logFunc: func(l logger.ZapLogger, msg string, fields ...logger.Field) {
				l.Info(msg, fields...)
			},
			checkLog: func(logs []observer.LoggedEntry) bool {
				return len(logs) > 0 && logs[0].Level == zapcore.InfoLevel
			},
		},
		{
			name: "Warn",
			logFunc: func(l logger.ZapLogger, msg string, fields ...logger.Field) {
				l.Warn(msg, fields...)
			},
			checkLog: func(logs []observer.LoggedEntry) bool {
				return len(logs) > 0 && logs[0].Level == zapcore.WarnLevel
			},
		},
		{
			name: "Error",
			logFunc: func(l logger.ZapLogger, msg string, fields ...logger.Field) {
				l.Error(msg, fields...)
			},
			checkLog: func(logs []observer.LoggedEntry) bool {
				return len(logs) > 0 && logs[0].Level == zapcore.ErrorLevel
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			core, observedLogs := observer.New(zapcore.DebugLevel)
			zapLogger := logger.New(core)

			tc.logFunc(zapLogger, "test message", "field1")

			logs := observedLogs.All()
			assert.True(t, tc.checkLog(logs), "log level should match expected")
			if len(logs) > 0 {
				assert.Equal(t, "test message", logs[0].Message, "log message should match")
			}
		})
	}
}

func TestWithMethod(t *testing.T) {
	core, observedLogs := observer.New(zapcore.DebugLevel)
	zapLogger := logger.New(core)

	withLogger := zapLogger.With(zap.String("common_key", "common_value"))
	require.NotNil(t, withLogger, "with should return a non-nil logger")

	withLogger.Info("test message", zap.String("specific_key", "specific_value"))

	logs := observedLogs.All()
	require.Len(t, logs, 1, "should have 1 log entry")

	contextMap := logs[0].ContextMap()
	assert.Equal(t, "common_value", contextMap["common_key"], "common field should be present")
	assert.Equal(t, "specific_value", contextMap["specific_key"], "specific field should be present")
}

func TestSetAndGetLevel(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	zapLogger := logger.New(core)

	levels := []logger.LogLevel{
		logger.DebugLevel,
		logger.InfoLevel,
		logger.WarnLevel,
		logger.ErrorLevel,
		logger.FatalLevel,
	}

	for _, lvl := range levels {
		zapLogger.SetLevel(lvl)
		assert.Equal(t, lvl, zapLogger.GetLevel(), "GetLevel should return the level set by SetLevel")
	}
}

func TestSync(t *testing.T) {
	t.Run("successful sync", func(t *testing.T) {
		core, _ := observer.New(zapcore.DebugLevel)
		zapLogger := logger.New(core)

		err := zapLogger.Sync()
		assert.NoError(t, err, "Sync should not return an error")
	})
}

func TestRawLogger(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	zapLogger := logger.New(core)

	rawLogger := zapLogger.RawLogger()
	assert.NotNil(t, rawLogger, "RawLogger should return a non-nil zap.Logger")
}

func TestFactoryFunctions(t *testing.T) {
	t.Run("New creates a valid logger", func(t *testing.T) {
		core, _ := observer.New(zapcore.DebugLevel)
		zapLogger := logger.New(core)
		assert.NotNil(t, zapLogger, "New should return a non-nil logger")

		rawLogger := zapLogger.RawLogger()
		assert.NotNil(t, rawLogger, "should provide a raw zap.Logger")
	})

	t.Run("Console creates a valid logger", func(t *testing.T) {
		loggerInstance := logger.Console(logger.DebugLevel, false)
		assert.NotNil(t, loggerInstance, "Console should return a non-nil logger")
		assert.Equal(t, logger.DebugLevel, loggerInstance.GetLevel(), "level should be set correctly")
	})

	t.Run("Development creates a valid logger", func(t *testing.T) {
		loggerInstance, err := logger.Development()
		require.NoError(t, err, "Development should not return an error")
		require.NotNil(t, loggerInstance, "Development should return a non-nil logger")
		assert.Equal(t, logger.DebugLevel, loggerInstance.GetLevel(), "level should be Debug")
	})

	t.Run("Production creates a valid logger", func(t *testing.T) {
		loggerInstance, err := logger.Production()
		require.NoError(t, err, "Production should not return an error")
		require.NotNil(t, loggerInstance, "Production should return a non-nil logger")
		assert.Equal(t, logger.InfoLevel, loggerInstance.GetLevel(), "level should be Info")
	})
}

func TestContextFunctions(t *testing.T) {
	t.Run("WithLogger adds logger to context", func(t *testing.T) {
		ctx := context.Background()
		mockLogger := newMockCtxLogger()

		newCtx := logger.WithLogger(ctx, mockLogger)
		require.NotEqual(t, ctx, newCtx, "WithLogger should return a new context")

		result, ok := logger.FromContext(newCtx)
		assert.True(t, ok, "should find logger in context")
		assert.Equal(t, mockLogger, result, "should return the same logger")
	})

	t.Run("FromContext retrieves logger from context", func(t *testing.T) {
		ctx := context.Background()
		mockLogger := newMockCtxLogger()
		ctx = logger.WithLogger(ctx, mockLogger)

		result, ok := logger.FromContext(ctx)
		assert.True(t, ok, "should find logger in context")
		assert.Equal(t, mockLogger, result, "should return the same logger")

		result, ok = logger.FromContext(context.Background())
		assert.False(t, ok, "should not find logger in empty context")
		assert.Nil(t, result, "result should be nil")
	})

	t.Run("GetLogger retrieves logger from context or uses default", func(t *testing.T) {
		ctx := context.Background()
		mockLogger := newMockCtxLogger()
		defaultLogger := newMockCtxLogger()

		result := logger.GetLogger(ctx, mockLogger)
		assert.Equal(t, mockLogger, result, "should return default logger for empty context")

		ctx = logger.WithLogger(ctx, mockLogger)
		result = logger.GetLogger(ctx, defaultLogger)
		assert.Equal(t, mockLogger, result, "should return logger from context")
	})

	t.Run("WithFields adds fields to logger in context", func(t *testing.T) {
		ctx := context.Background()
		mockLogger := newMockCtxLogger()
		ctx = logger.WithLogger(ctx, mockLogger)

		field1 := "field1"
		field2 := "field2"
		newCtx := logger.WithFields(ctx, mockLogger, field1, field2)

		result, ok := logger.FromContext(newCtx)
		require.True(t, ok, "should find logger in context")

		mockResult := result.(*mockCtxLogger)
		require.Len(t, mockResult.withFields, 2, "should have 2 fields")
		assert.Equal(t, field1, mockResult.withFields[0], "should contain first field")
		assert.Equal(t, field2, mockResult.withFields[1], "should contain second field")
	})
}

func TestRequestIDFunctions(t *testing.T) {
	t.Run("WithRequestID adds ID to context", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"

		newCtx := logger.WithRequestID(ctx, id)
		require.NotEqual(t, ctx, newCtx, "WithRequestID should return a new context")

		result, ok := logger.RequestID(newCtx)
		assert.True(t, ok, "ID should exist in the context")
		assert.Equal(t, id, result, "ID should match the one set")
	})

	t.Run("RequestID retrieves ID from context", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = logger.WithRequestID(ctx, id)

		result, ok := logger.RequestID(ctx)
		assert.True(t, ok, "ID should exist in the context")
		assert.Equal(t, id, result, "ID should match the one set")

		result, ok = logger.RequestID(context.Background())
		assert.False(t, ok, "ID should not exist in empty context")
		assert.Empty(t, result, "result should be empty")
	})

	t.Run("GenerateRequestID creates valid UUID", func(t *testing.T) {
		id := logger.GenerateRequestID()
		assert.NotEmpty(t, id, "generated ID should not be empty")

		assert.Len(t, id, 36, "UUID should be 36 characters")
		assert.Contains(t, id, "-", "UUID should contain hyphens")
	})

	t.Run("EnsureRequestID creates or returns ID", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = logger.WithRequestID(ctx, id)

		newCtx, resultID := logger.EnsureRequestID(ctx)
		assert.Equal(t, ctx, newCtx, "EnsureRequestID should return original context")
		assert.Equal(t, id, resultID, "should return existing ID")

		ctx = context.Background()
		newCtx, resultID = logger.EnsureRequestID(ctx)
		require.NotEqual(t, ctx, newCtx, "EnsureRequestID should create new context")
		assert.NotEmpty(t, resultID, "should generate an ID")

		extractedID, ok := logger.RequestID(newCtx)
		assert.True(t, ok, "ID should be in context")
		assert.Equal(t, resultID, extractedID, "extracted ID should match returned ID")
	})

	t.Run("WithRequestIDField adds request ID to logger", func(t *testing.T) {
		core, observedLogs := observer.New(zapcore.DebugLevel)
		zapLogger := logger.New(core)

		ctx := context.Background()
		id := "test-request-id"
		ctx = logger.WithRequestID(ctx, id)

		loggerWithID := logger.WithRequestIDField(ctx, zapLogger)
		loggerWithID.Info("test message")

		logs := observedLogs.All()
		require.Len(t, logs, 1, "should have 1 log entry")
		assert.Equal(t, id, logs[0].ContextMap()[logger.RequestIDField], "request_id should be in the log")
	})

	t.Run("ContextLogger gets logger with request ID", func(t *testing.T) {
		core, observedLogs := observer.New(zapcore.DebugLevel)
		zapLogger := logger.New(core)

		ctx := context.Background()
		id := "test-request-id"
		ctx = logger.WithRequestID(ctx, id)

		contextLogger := logger.ContextLogger(ctx, zapLogger)
		contextLogger.Info("test message")

		logs := observedLogs.All()
		require.Len(t, logs, 1, "should have 1 log entry")
		assert.Equal(t, id, logs[0].ContextMap()[logger.RequestIDField], "request_id should be in the log")
	})
}

func TestLogFunctions(t *testing.T) {
	testCases := []struct {
		name    string
		logFunc func(ctx context.Context, defaultLogger logger.Logger, msg string, fields ...logger.Field)
		checkFn func(m *mockCtxLogger) bool
	}{
		{
			name:    "Debug logs debug level messages",
			logFunc: logger.Debug,
			checkFn: func(m *mockCtxLogger) bool { return m.debugCalled },
		},
		{
			name:    "Info logs info level messages",
			logFunc: logger.Info,
			checkFn: func(m *mockCtxLogger) bool { return m.infoCalled },
		},
		{
			name:    "Warn logs warn level messages",
			logFunc: logger.Warn,
			checkFn: func(m *mockCtxLogger) bool { return m.warnCalled },
		},
		{
			name:    "Error logs error level messages",
			logFunc: logger.Error,
			checkFn: func(m *mockCtxLogger) bool { return m.errorCalled },
		},
		{
			name:    "Fatal logs fatal level messages",
			logFunc: logger.Fatal,
			checkFn: func(m *mockCtxLogger) bool { return m.fatalCalled },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockLogger := newMockCtxLogger()

			tc.logFunc(ctx, mockLogger, "test message", "field1")
			assert.True(t, tc.checkFn(mockLogger), "should call correct log method on default logger")

			ctxLogger := newMockCtxLogger()
			ctx = logger.WithLogger(ctx, ctxLogger)
			tc.logFunc(ctx, mockLogger, "context message", "field2")
			assert.True(t, tc.checkFn(ctxLogger), "should call correct log method on context logger")
		})
	}

	t.Run("Log function with level parameter", func(t *testing.T) {
		ctx := context.Background()
		mockLogger := newMockCtxLogger()

		testCases := []struct {
			level   logger.LogLevel
			checkFn func(m *mockCtxLogger) bool
		}{
			{logger.DebugLevel, func(m *mockCtxLogger) bool { return m.debugCalled }},
			{logger.InfoLevel, func(m *mockCtxLogger) bool { return m.infoCalled }},
			{logger.WarnLevel, func(m *mockCtxLogger) bool { return m.warnCalled }},
			{logger.ErrorLevel, func(m *mockCtxLogger) bool { return m.errorCalled }},
			{logger.FatalLevel, func(m *mockCtxLogger) bool { return m.fatalCalled }},
		}

		for _, tc := range testCases {
			logger.Log(ctx, mockLogger, tc.level, "test message", "field")
			assert.True(t, tc.checkFn(mockLogger), "Log should call correct method based on level")
		}
	})

	t.Run("Sync flushes logs", func(t *testing.T) {
		ctx := context.Background()
		mockLogger := newMockCtxLogger()

		err := logger.Sync(ctx, mockLogger)
		require.NoError(t, err, "Sync should not return error")
		assert.True(t, mockLogger.syncCalled, "Sync should call underlying Sync method")

		mockLoggerWithErr := newMockCtxLogger()
		mockLoggerWithErr.syncErr = errTest

		err = logger.Sync(ctx, mockLoggerWithErr)
		require.Error(t, err, "Sync should return error when underlying Sync fails")
		require.ErrorIs(t, err, errTest, "error should wrap the original error")
	})
}

func TestZapLoggerInterface(t *testing.T) {
	var _ logger.ZapLogger = (*mockZapLogger)(nil)

	core, _ := observer.New(zapcore.DebugLevel)
	zapLogger := logger.New(core)
	_ = logger.ZapLogger(zapLogger)
}
