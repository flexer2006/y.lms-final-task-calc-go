package level_test

import (
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		name     string
		level    level.LogLevel
		expected string
	}{
		{
			name:     "debug level",
			level:    level.DebugLevel,
			expected: "debug",
		},
		{
			name:     "info level",
			level:    level.InfoLevel,
			expected: "info",
		},
		{
			name:     "warn level",
			level:    level.WarnLevel,
			expected: "warn",
		},
		{
			name:     "error level",
			level:    level.ErrorLevel,
			expected: "error",
		},
		{
			name:     "fatal level",
			level:    level.FatalLevel,
			expected: "fatal",
		},
		{
			name:     "unknown level",
			level:    level.LogLevel(99),
			expected: "unknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.level.String()
			assert.Equal(t, tc.expected, actual, "string should return the correct value")
		})
	}
}

func TestLogLevelToZapLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    level.LogLevel
		expected zapcore.Level
	}{
		{
			name:     "debug level",
			level:    level.DebugLevel,
			expected: zapcore.DebugLevel,
		},
		{
			name:     "info level",
			level:    level.InfoLevel,
			expected: zapcore.InfoLevel,
		},
		{
			name:     "warn level",
			level:    level.WarnLevel,
			expected: zapcore.WarnLevel,
		},
		{
			name:     "error level",
			level:    level.ErrorLevel,
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "fatal level",
			level:    level.FatalLevel,
			expected: zapcore.FatalLevel,
		},
		{
			name:     "unknown level defaults to Info",
			level:    level.LogLevel(99),
			expected: zapcore.InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.level.ToZapLevel()
			assert.Equal(t, tc.expected, actual, "toZapLevel should return the correct zap level")
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		expected level.LogLevel
	}{
		{
			name:     "debug string",
			str:      "debug",
			expected: level.DebugLevel,
		},
		{
			name:     "info string",
			str:      "info",
			expected: level.InfoLevel,
		},
		{
			name:     "warn string",
			str:      "warn",
			expected: level.WarnLevel,
		},
		{
			name:     "error string",
			str:      "error",
			expected: level.ErrorLevel,
		},
		{
			name:     "fatal string",
			str:      "fatal",
			expected: level.FatalLevel,
		},
		{
			name:     "unknown string defaults to Info",
			str:      "invalid",
			expected: level.InfoLevel,
		},
		{
			name:     "empty string defaults to Info",
			str:      "",
			expected: level.InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := level.Parse(tc.str)
			assert.Equal(t, tc.expected, actual, "parse should return the correct LogLevel")
		})
	}
}

func TestFromZapLevel(t *testing.T) {
	tests := []struct {
		name     string
		zapLevel zapcore.Level
		expected level.LogLevel
	}{
		{
			name:     "debug zap level",
			zapLevel: zapcore.DebugLevel,
			expected: level.DebugLevel,
		},
		{
			name:     "info zap level",
			zapLevel: zapcore.InfoLevel,
			expected: level.InfoLevel,
		},
		{
			name:     "warn zap level",
			zapLevel: zapcore.WarnLevel,
			expected: level.WarnLevel,
		},
		{
			name:     "error zap level",
			zapLevel: zapcore.ErrorLevel,
			expected: level.ErrorLevel,
		},
		{
			name:     "fatal zap level",
			zapLevel: zapcore.FatalLevel,
			expected: level.FatalLevel,
		},
		{
			name:     "d panic zap level (not directly mapped) defaults to Info",
			zapLevel: zapcore.DPanicLevel,
			expected: level.InfoLevel,
		},
		{
			name:     "panic zap level (not directly mapped) defaults to Info",
			zapLevel: zapcore.PanicLevel,
			expected: level.InfoLevel,
		},
		{
			name:     "custom zap level defaults to Info",
			zapLevel: zapcore.Level(99),
			expected: level.InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := level.FromZapLevel(tc.zapLevel)
			assert.Equal(t, tc.expected, actual, "fromZapLevel should return the correct LogLevel")
		})
	}
}

func TestLogLevelComparisons(t *testing.T) {
	t.Run("Level ordering is correct", func(t *testing.T) {
		assert.Less(t, level.DebugLevel, level.InfoLevel, "debug should be lower than Info")
		assert.Less(t, level.InfoLevel, level.WarnLevel, "info should be lower than Warn")
		assert.Less(t, level.WarnLevel, level.ErrorLevel, "warn should be lower than Error")
		assert.Less(t, level.ErrorLevel, level.FatalLevel, "error should be lower than Fatal")
	})
}
