package factory_test

import (
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/factory"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type mockCore struct {
	level      zapcore.Level
	hasLevel   bool
	enabledFor zapcore.Level
	fields     []zapcore.Field
}

func (m *mockCore) Enabled(lvl zapcore.Level) bool {
	return lvl >= m.enabledFor
}

func (m *mockCore) With(fields []zapcore.Field) zapcore.Core {
	m.fields = append(m.fields, fields...)
	return m
}

func (m *mockCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if m.Enabled(ent.Level) {
		return ce.AddCore(ent, m)
	}
	return ce
}

func (m *mockCore) Write(_ zapcore.Entry, fields []zapcore.Field) error {
	m.fields = append(m.fields, fields...)
	return nil
}

func (m *mockCore) Sync() error {
	return nil
}

func (m *mockCore) Level() zapcore.Level {
	if !m.hasLevel {
		return m.enabledFor
	}
	return m.level
}

type basicCore struct{}

func (b *basicCore) Enabled(_ zapcore.Level) bool        { return true }
func (b *basicCore) With(_ []zapcore.Field) zapcore.Core { return b }
func (b *basicCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, b)
}
func (b *basicCore) Write(_ zapcore.Entry, _ []zapcore.Field) error { return nil }
func (b *basicCore) Sync() error                                    { return nil }

func TestNew(t *testing.T) {
	testCases := []struct {
		name           string
		core           zapcore.Core
		expectedLevel  level.LogLevel
		coreHasLevel   bool
		coreLevel      zapcore.Level
		coreEnabledFor zapcore.Level
	}{
		{
			name:          "creates logger with core that has no Level method",
			core:          &basicCore{},
			expectedLevel: level.InfoLevel,
			coreHasLevel:  false,
		},
		{
			name:           "extracts level from core with Level method",
			coreHasLevel:   true,
			coreLevel:      zapcore.DebugLevel,
			coreEnabledFor: zapcore.DebugLevel,
			expectedLevel:  level.DebugLevel,
		},
		{
			name:           "uses enabledFor as level when hasLevel is false",
			coreHasLevel:   false,
			coreEnabledFor: zapcore.InfoLevel,
			expectedLevel:  level.InfoLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var core zapcore.Core

			if tc.core != nil {
				core = tc.core
			} else {
				core = &mockCore{
					level:      tc.coreLevel,
					hasLevel:   tc.coreHasLevel,
					enabledFor: tc.coreEnabledFor,
				}
			}

			logger := factory.New(core)

			require.NotNil(t, logger, "logger should not be nil")
			assert.Equal(t, tc.expectedLevel, logger.GetLevel(), "logger should have expected level")
		})
	}
}

func TestConsole(t *testing.T) {
	testCases := []struct {
		name     string
		logLevel level.LogLevel
		json     bool
	}{
		{name: "console logger with debug level", logLevel: level.DebugLevel, json: false},
		{name: "JSON logger with info level", logLevel: level.InfoLevel, json: true},
		{name: "console logger with warn level", logLevel: level.WarnLevel, json: false},
		{name: "JSON logger with error level", logLevel: level.ErrorLevel, json: true},
		{name: "console logger with fatal level", logLevel: level.FatalLevel, json: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := factory.Console(tc.logLevel, tc.json)

			require.NotNil(t, logger, "logger should not be nil")
			assert.Equal(t, tc.logLevel, logger.GetLevel(), "logger should have expected level")

			logger.Info("test message")
			_ = logger.Sync()
		})
	}
}

func TestDevelopment(t *testing.T) {
	logger, err := factory.Development()

	require.NoError(t, err, "development logger creation should not error")
	require.NotNil(t, logger, "logger should not be nil")
	assert.Equal(t, level.DebugLevel, logger.GetLevel(), "development logger should have debug level")

	logger.Debug("debug message")
	_ = logger.Sync()
}

func TestProduction(t *testing.T) {
	logger, err := factory.Production()

	require.NoError(t, err, "production logger creation should not error")
	require.NotNil(t, logger, "logger should not be nil")
	assert.Equal(t, level.InfoLevel, logger.GetLevel(), "production logger should have info level")

	logger.Info("info message")
	_ = logger.Sync()
}

func TestNewWithCustomCore(t *testing.T) {
	zapLevels := []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.FatalLevel,
	}

	for _, zapLvl := range zapLevels {
		t.Run("level "+zapLvl.String(), func(t *testing.T) {
			encoder := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
			writer := zapcore.AddSync(zapcore.NewMultiWriteSyncer())
			atomicLevel := zap.NewAtomicLevelAt(zapLvl)
			core := zapcore.NewCore(encoder, writer, atomicLevel)

			logger := factory.New(core)

			require.NotNil(t, logger, "logger should not be nil")
			expectedLevel := level.FromZapLevel(zapLvl)
			assert.Equal(t, expectedLevel, logger.GetLevel(),
				"logger should have level matching the input zap core level")
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := factory.New(core)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	assert.Equal(t, 3, recorded.Len(), "should record exactly 3 messages (info, warn, error)")

	entries := recorded.All()
	assert.Equal(t, "info message", entries[0].Message, "first message should be info")
	assert.Equal(t, "warn message", entries[1].Message, "second message should be warn")
	assert.Equal(t, "error message", entries[2].Message, "third message should be error")

	assert.Equal(t, zapcore.InfoLevel, entries[0].Level, "first message should be at info level")
	assert.Equal(t, zapcore.WarnLevel, entries[1].Level, "second message should be at warn level")
	assert.Equal(t, zapcore.ErrorLevel, entries[2].Level, "third message should be at error level")
}
