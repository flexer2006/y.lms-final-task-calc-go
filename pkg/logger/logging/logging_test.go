package logging_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/factory"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var (
	errTest = errors.New("test error")
)

func TestNew(t *testing.T) {
	core, observedLogs := observer.New(zapcore.DebugLevel)

	logger := logging.New(core)
	require.NotNil(t, logger, "New should return a non-nil logger")

	logger.Info("test message", zap.String("key", "value"))

	logs := observedLogs.All()
	require.Len(t, logs, 1, "should have 1 log entry")
	assert.Equal(t, "test message", logs[0].Message, "log message should match")
	assert.Equal(t, zapcore.InfoLevel, logs[0].Level, "log level should be Info")

	field, exists := logs[0].ContextMap()["key"]
	assert.True(t, exists, "field should exist in context map")
	assert.Equal(t, "value", field, "field value should match")
}

func TestConsole(t *testing.T) {
	tests := []struct {
		name        string
		lvl         level.LogLevel
		json        bool
		expectedLvl zapcore.Level
	}{
		{
			name:        "debug level with console encoder",
			lvl:         level.DebugLevel,
			json:        false,
			expectedLvl: zapcore.DebugLevel,
		},
		{
			name:        "info level with json encoder",
			lvl:         level.InfoLevel,
			json:        true,
			expectedLvl: zapcore.InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := logging.Console(tc.lvl, tc.json)
			require.NotNil(t, logger, "console should return a non-nil logger")

			assert.Equal(t, tc.lvl, logger.GetLevel(),
				"console should set the atomic level based on the provided level")
		})
	}
}

func TestDevelopment(t *testing.T) {
	logger, err := logging.Development()
	require.NoError(t, err, "development should not return an error")
	require.NotNil(t, logger, "development should return a non-nil logger")

	assert.Equal(t, level.DebugLevel, logger.GetLevel(),
		"development should set the level to Debug")

	wrappedErr := fmt.Errorf("%s: %w", factory.ErrBuildDevLogger, errTest)
	assert.Contains(t, wrappedErr.Error(), factory.ErrBuildDevLogger)
}

func TestProduction(t *testing.T) {
	logger, err := logging.Production()
	require.NoError(t, err, "production should not return an error")
	require.NotNil(t, logger, "production should return a non-nil logger")

	assert.Equal(t, level.InfoLevel, logger.GetLevel(),
		"production should set the level to Info")

	wrappedErr := fmt.Errorf("%s: %w", factory.ErrBuildProdLogger, errTest)
	assert.Contains(t, wrappedErr.Error(), factory.ErrBuildProdLogger)
}

func TestWith(t *testing.T) {
	core, observedLogs := observer.New(zapcore.DebugLevel)

	logger := logging.New(core)
	withLogger := logger.With(zap.String("common_key", "common_value"))
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
	logger := logging.New(core)

	levels := []level.LogLevel{
		level.DebugLevel,
		level.InfoLevel,
		level.WarnLevel,
		level.ErrorLevel,
		level.FatalLevel,
	}

	for _, lvl := range levels {
		logger.SetLevel(lvl)
		assert.Equal(t, lvl, logger.GetLevel(), "GetLevel should return the level set by SetLevel")
	}
}

func TestLogMethods(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(l *logging.Logger, msg string, fields ...zap.Field)
		expected zapcore.Level
	}{
		{
			name: "Debug",
			logFunc: func(l *logging.Logger, msg string, fields ...zap.Field) {
				l.Debug(msg, fields...)
			},
			expected: zapcore.DebugLevel,
		},
		{
			name: "Info",
			logFunc: func(l *logging.Logger, msg string, fields ...zap.Field) {
				l.Info(msg, fields...)
			},
			expected: zapcore.InfoLevel,
		},
		{
			name: "Warn",
			logFunc: func(l *logging.Logger, msg string, fields ...zap.Field) {
				l.Warn(msg, fields...)
			},
			expected: zapcore.WarnLevel,
		},
		{
			name: "Error",
			logFunc: func(l *logging.Logger, msg string, fields ...zap.Field) {
				l.Error(msg, fields...)
			},
			expected: zapcore.ErrorLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			core, observedLogs := observer.New(zapcore.DebugLevel)
			logger := logging.New(core)

			tc.logFunc(logger, "test message", zap.String("key", "value"))

			logs := observedLogs.All()
			require.Len(t, logs, 1, "should have 1 log entry")
			assert.Equal(t, "test message", logs[0].Message, "log message should match")
			assert.Equal(t, tc.expected, logs[0].Level, "log level should match expected level")

			field, exists := logs[0].ContextMap()["key"]
			assert.True(t, exists, "field should exist in context map")
			assert.Equal(t, "value", field, "field value should match")
		})
	}
}

func TestSync(t *testing.T) {
	var buf bytes.Buffer
	encoderConfig := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	ws := zapcore.AddSync(&buf)
	zapCore := zapcore.NewCore(encoder, ws, zapcore.DebugLevel)
	logger := logging.New(zapCore)

	err := logger.Sync()
	require.NoError(t, err, "sync should not return an error")

	var nilLogger *logging.Logger
	err = nilLogger.Sync()
	require.NoError(t, err, "sync should handle nil logger gracefully")
}

func TestRawLogger(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := logging.New(core)

	rawLogger := logger.RawLogger()
	require.NotNil(t, rawLogger, "RawLogger should return a non-nil zap.Logger")

	rawLogger.Info("test from raw logger", zap.String("source", "raw"))
}

func TestLoggNilLogger(t *testing.T) {
	var nilLogger *logging.Logger

	nilLogger.Debug("test")
	nilLogger.Info("test")
	nilLogger.Warn("test")
	nilLogger.Error("test")
}

func TestLoggDefaultLevel(t *testing.T) {
	core, observedLogs := observer.New(zapcore.DebugLevel)
	logger := logging.New(core)

	logger.Info("default level test")

	logs := observedLogs.All()
	require.Len(t, logs, 1, "should have 1 log entry")
	assert.Equal(t, zapcore.InfoLevel, logs[0].Level, "log level should be Info")
}

func TestLoggerImplementsInterface(t *testing.T) {
	var logger logging.LoggerInterface
	core, _ := observer.New(zapcore.DebugLevel)
	logger = logging.New(core)

	require.NotNil(t, logger, "logger should not be nil")
}

func TestFatalLevel(t *testing.T) {
	core, _ := observer.New(zapcore.FatalLevel)

	assert.True(t, core.Enabled(zapcore.FatalLevel), "Fatal level should be enabled")

	entry := zapcore.Entry{
		Level:   zapcore.FatalLevel,
		Message: "fatal message that won't execute",
	}

	checkedEntry := core.Check(entry, nil)
	assert.NotNil(t, checkedEntry, "should create a checked entry for Fatal level")
}

func TestCustomLevel(t *testing.T) {
	specialCore, logs := observer.New(zapcore.DebugLevel)
	specialLogger := logging.New(specialCore)

	specialLogger.Debug("test debug")
	specialLogger.Info("test info")
	specialLogger.Warn("test warn")
	specialLogger.Error("test error")

	specialLogger.Info("custom level message")

	allLogs := logs.All()
	lastLog := allLogs[len(allLogs)-1]

	assert.Equal(t, zapcore.InfoLevel, lastLog.Level, "unknown level should default to Info")
	assert.Equal(t, "custom level message", lastLog.Message)
}
