package core_test

import (
	"testing"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/core"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type mockWriteSyncer struct {
	written [][]byte
}

func (m *mockWriteSyncer) Write(p []byte) (n int, err error) {
	m.written = append(m.written, p)
	return len(p), nil
}

func (m *mockWriteSyncer) Sync() error {
	return nil
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		lvl      level.LogLevel
		expected zapcore.Level
	}{
		{
			name:     "debug level",
			lvl:      level.DebugLevel,
			expected: zapcore.DebugLevel,
		},
		{
			name:     "info level",
			lvl:      level.InfoLevel,
			expected: zapcore.InfoLevel,
		},
		{
			name:     "warn level",
			lvl:      level.WarnLevel,
			expected: zapcore.WarnLevel,
		},
		{
			name:     "error level",
			lvl:      level.ErrorLevel,
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "fatal level",
			lvl:      level.FatalLevel,
			expected: zapcore.FatalLevel,
		},
		{
			name:     "unknown level defaults to info",
			lvl:      level.LogLevel(99),
			expected: zapcore.InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{})
			writer := &mockWriteSyncer{}

			c := core.New(encoder, writer, tc.lvl)
			require.NotNil(t, c, "core should not be nil")

			assert.True(t, c.Enabled(tc.expected), "core should enable messages at the expected level")

			if tc.expected < zapcore.FatalLevel {
				assert.True(t, c.Enabled(tc.expected+1), "core should enable messages at higher levels")
			}

			if tc.expected > zapcore.DebugLevel {
				assert.False(t, c.Enabled(tc.expected-1), "core should not enable messages at lower levels")
			}
		})
	}
}

func TestCreateEncoder(t *testing.T) {
	t.Run("json encoder", func(t *testing.T) {
		encoder := core.CreateEncoder(true)
		require.NotNil(t, encoder, "encoder should not be nil")

		entry := zapcore.Entry{
			Message: "test message",
			Level:   zapcore.InfoLevel,
		}
		fields := []zapcore.Field{
			zap.String("key", "value"),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		require.NoError(t, err, "encoding should not fail")

		assert.Contains(t, buf.String(), `"key":"value"`, "JSON encoder should format fields as JSON")
		assert.Contains(t, buf.String(), `"msg":"test message"`, "JSON encoder should include message")
	})

	t.Run("console encoder", func(t *testing.T) {
		encoder := core.CreateEncoder(false)
		require.NotNil(t, encoder, "encoder should not be nil")

		entry := zapcore.Entry{
			Message: "test message",
			Level:   zapcore.InfoLevel,
		}
		fields := []zapcore.Field{
			zap.String("key", "value"),
		}

		buf, err := encoder.EncodeEntry(entry, fields)
		require.NoError(t, err, "encoding should not fail")

		assert.Contains(t, buf.String(), "test message", "Console encoder should include message")
		assert.Contains(t, buf.String(), "key", "Console encoder should include field key")
		assert.Contains(t, buf.String(), "value", "Console encoder should include field value")
	})

	t.Run("time encoding format", func(t *testing.T) {
		testTime := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)

		expectedFormat := "2023-01-02T15:04:05"

		jsonEncoder := core.CreateEncoder(true)
		consoleEncoder := core.CreateEncoder(false)

		entry := zapcore.Entry{
			Time:    testTime,
			Message: "test",
			Level:   zapcore.InfoLevel,
		}

		jsonBuf, err := jsonEncoder.EncodeEntry(entry, nil)
		require.NoError(t, err)

		consoleBuf, err := consoleEncoder.EncodeEntry(entry, nil)
		require.NoError(t, err)

		jsonOutput := jsonBuf.String()
		consoleOutput := consoleBuf.String()

		assert.Contains(t, jsonOutput, expectedFormat,
			"JSON encoder should use ISO8601 time format, output: %s", jsonOutput)
		assert.Contains(t, consoleOutput, expectedFormat,
			"Console encoder should use ISO8601 time format, output: %s", consoleOutput)
	})
}
