package factory_test

import (
	"bytes"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/factory"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MockCore struct {
	level zapcore.Level
}

func (m MockCore) Enabled(lvl zapcore.Level) bool {
	return lvl >= m.level
}

func (m MockCore) With(fields []zapcore.Field) zapcore.Core {
	return m
}

func (m MockCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if m.Enabled(ent.Level) {
		return ce.AddCore(ent, m)
	}
	return ce
}

func (m MockCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	return nil
}

func (m MockCore) Sync() error {
	return nil
}

func (m MockCore) Level() zapcore.Level {
	return m.level
}

func TestNewLogger(t *testing.T) {
	zapLogger := zap.NewExample()
	atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	logger := factory.NewLogger(zapLogger, atomicLevel)

	assert.Equal(t, zapLogger, logger.GetZapLogger(), "GetZapLogger should return the provided zap logger")

	assert.Equal(t, atomicLevel, logger.GetAtomicLevel(), "GetAtomicLevel should return the provided atomic level")
}

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		coreLevel     zapcore.Level
		expectedLevel zapcore.Level
		hasLevelFunc  bool
	}{
		{
			name:          "core with level method",
			coreLevel:     zapcore.DebugLevel,
			expectedLevel: zapcore.DebugLevel,
			hasLevelFunc:  true,
		},
		{
			name:          "core without level method",
			coreLevel:     zapcore.WarnLevel,
			expectedLevel: zapcore.InfoLevel,
			hasLevelFunc:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var core zapcore.Core
			if tc.hasLevelFunc {
				core = MockCore{level: tc.coreLevel}
			} else {
				core = noLevelMethodCore{}
			}

			logger := factory.New(core)
			require.NotNil(t, logger, "New should return a non-nil logger")

			assert.Equal(t, tc.expectedLevel, logger.GetAtomicLevel().Level(),
				"New should set the atomic level based on core's Level method if available, otherwise default to Info")

			assert.NotNil(t, logger.GetZapLogger(), "New should create a non-nil zap logger")
		})
	}
}

type noLevelMethodCore struct{}

func (n noLevelMethodCore) Enabled(level zapcore.Level) bool         { return true }
func (n noLevelMethodCore) With(fields []zapcore.Field) zapcore.Core { return n }
func (n noLevelMethodCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, n)
}
func (n noLevelMethodCore) Write(ent zapcore.Entry, fields []zapcore.Field) error { return nil }
func (n noLevelMethodCore) Sync() error                                           { return nil }

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
		{
			name:        "warn level with console encoder",
			lvl:         level.WarnLevel,
			json:        false,
			expectedLvl: zapcore.WarnLevel,
		},
		{
			name:        "error level with json encoder",
			lvl:         level.ErrorLevel,
			json:        true,
			expectedLvl: zapcore.ErrorLevel,
		},
		{
			name:        "fatal level with console encoder",
			lvl:         level.FatalLevel,
			json:        false,
			expectedLvl: zapcore.FatalLevel,
		},
		{
			name:        "unknown level defaults to info",
			lvl:         level.LogLevel(99),
			json:        true,
			expectedLvl: zapcore.InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := factory.Console(tc.lvl, tc.json)
			require.NotNil(t, logger, "Console should return a non-nil logger")

			assert.Equal(t, tc.expectedLvl, logger.GetAtomicLevel().Level(),
				"Console should set the atomic level based on the provided level")

			assert.NotNil(t, logger.GetZapLogger(), "Console should create a non-nil zap logger")
		})
	}
}

func TestDevelopment(t *testing.T) {
	logger, err := factory.Development()
	require.NoError(t, err, "Development should not return an error")
	require.NotNil(t, logger, "Development should return a non-nil logger")

	assert.Equal(t, zapcore.DebugLevel, logger.GetAtomicLevel().Level(),
		"Development should set the level to Debug")

	assert.NotNil(t, logger.GetZapLogger(), "Development should create a non-nil zap logger")
}

func TestProduction(t *testing.T) {
	logger, err := factory.Production()
	require.NoError(t, err, "Production should not return an error")
	require.NotNil(t, logger, "Production should return a non-nil logger")

	assert.Equal(t, zapcore.InfoLevel, logger.GetAtomicLevel().Level(),
		"Production should set the level to Info")

	assert.NotNil(t, logger.GetZapLogger(), "Production should create a non-nil zap logger")
}

func TestGetZapLogger(t *testing.T) {
	zapLogger := zap.NewExample()
	logger := factory.NewLogger(zapLogger, zap.NewAtomicLevel())

	assert.Equal(t, zapLogger, logger.GetZapLogger(),
		"GetZapLogger should return the logger that was provided to NewLogger")
}

func TestGetAtomicLevel(t *testing.T) {
	atomicLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger := factory.NewLogger(zap.NewExample(), atomicLevel)

	assert.Equal(t, atomicLevel, logger.GetAtomicLevel(),
		"GetAtomicLevel should return the level that was provided to NewLogger")
}

func TestLoggerLogging(t *testing.T) {
	var buf bytes.Buffer
	writeSyncer := zapcore.AddSync(&buf)

	encoder := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())

	zapCore := zapcore.NewCore(
		encoder,
		writeSyncer,
		zap.NewAtomicLevelAt(zapcore.InfoLevel),
	)

	logger := factory.New(zapCore)

	zapLogger := logger.GetZapLogger()
	zapLogger.Info("test message")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "test message", "The logger should write the message to the buffer")
}
