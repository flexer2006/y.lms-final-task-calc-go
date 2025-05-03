package request_test

import (
	"context"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogger struct {
	fields []mockField
}

type mockField struct {
	key   string
	value any
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		fields: make([]mockField, 0),
	}
}

func (m *mockLogger) With(fields ...mockField) *mockLogger {
	newLogger := &mockLogger{}

	newLogger.fields = make([]mockField, len(m.fields)+len(fields))
	copy(newLogger.fields, m.fields)
	copy(newLogger.fields[len(m.fields):], fields)
	return newLogger
}

func fieldConstructor(key string, value interface{}) mockField {
	return mockField{key: key, value: value}
}

func TestWithRequestID(t *testing.T) {
	t.Run("with valid context", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"

		newCtx := request.WithRequestID(ctx, id)
		require.NotEqual(t, ctx, newCtx, "WithRequestID should return a new context")

		result, ok := request.ID(newCtx)
		assert.True(t, ok, "ID should exist in the context")
		assert.Equal(t, id, result, "ID should match the one set")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context
		id := "test-id"

		newCtx := request.WithRequestID(ctx, id)
		assert.Nil(t, newCtx, "WithRequestID with nil context should return nil")
	})
}

func TestID(t *testing.T) {
	t.Run("with request_id in context", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = request.WithRequestID(ctx, id)

		result, ok := request.ID(ctx)
		assert.True(t, ok, "ID should exist in the context")
		assert.Equal(t, id, result, "ID should match the one set")
	})

	t.Run("without request_id in context", func(t *testing.T) {
		ctx := context.Background()

		result, ok := request.ID(ctx)
		assert.False(t, ok, "ID should not exist in the context")
		assert.Empty(t, result, "result should be empty")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context

		result, ok := request.ID(ctx)
		assert.False(t, ok, "ID should not exist in nil context")
		assert.Empty(t, result, "result should be empty")
	})
}

func TestGenerateRequestID(t *testing.T) {
	id := request.GenerateRequestID()
	assert.NotEmpty(t, id, "generated ID should not be empty")
	assert.True(t, request.IsValidUUID(id), "generated ID should be a valid UUID")

	anotherID := request.GenerateRequestID()
	assert.NotEqual(t, id, anotherID, "generated IDs should be different")
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "valid UUID",
			id:       "550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "empty string",
			id:       "",
			expected: false,
		},
		{
			name:     "invalid format",
			id:       "not-a-uuid",
			expected: false,
		},
		{
			name:     "invalid characters",
			id:       "550e8400-e29b-41d4-a716-44665544000z",
			expected: false,
		},
		{
			name:     "missing hyphens",
			id:       "550e8400e29b41d4a716446655440000",
			expected: false,
		},
		{
			name:     "too short",
			id:       "550e8400-e29b-41d4-a716-4466554400",
			expected: false,
		},
		{
			name:     "too long",
			id:       "550e8400-e29b-41d4-a716-4466554400000",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := request.IsValidUUID(tc.id)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWithRequestIDField(t *testing.T) {
	t.Run("with request_id in context", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = request.WithRequestID(ctx, id)

		logger := newMockLogger()
		result := request.WithRequestIDField(ctx, logger, fieldConstructor)

		require.NotNil(t, result, "WithRequestIDField should return a logger")
		require.Len(t, result.fields, 1, "logger should contain one field")
		assert.Equal(t, request.RequestIDFieldName, result.fields[0].key, "field name should be request_id")
		assert.Equal(t, id, result.fields[0].value, "field value should be the ID")
	})

	t.Run("without request_id in context", func(t *testing.T) {
		ctx := context.Background()
		logger := newMockLogger()

		result := request.WithRequestIDField(ctx, logger, fieldConstructor)
		assert.Equal(t, logger, result, "WithRequestIDField should return the original logger")
		assert.Empty(t, result.fields, "logger should not contain any fields")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context
		logger := newMockLogger()

		result := request.WithRequestIDField(ctx, logger, fieldConstructor)
		assert.Equal(t, logger, result, "WithRequestIDField should return the original logger")
	})

	t.Run("with empty request_id", func(t *testing.T) {
		ctx := context.Background()
		ctx = request.WithRequestID(ctx, "")
		logger := newMockLogger()

		result := request.WithRequestIDField(ctx, logger, fieldConstructor)
		assert.Equal(t, logger, result, "WithRequestIDField should return the original logger")
	})
}

func TestEnsureRequestID(t *testing.T) {
	t.Run("with existing request_id", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = request.WithRequestID(ctx, id)

		newCtx, resultID := request.EnsureRequestID(ctx)
		assert.Equal(t, ctx, newCtx, "EnsureRequestID should return the original context")
		assert.Equal(t, id, resultID, "EnsureRequestID should return the existing ID")
	})

	t.Run("without request_id", func(t *testing.T) {
		ctx := context.Background()

		newCtx, resultID := request.EnsureRequestID(ctx)
		require.NotEqual(t, ctx, newCtx, "EnsureRequestID should return a new context")
		assert.True(t, request.IsValidUUID(resultID), "EnsureRequestID should generate a valid UUID")

		extractedID, ok := request.ID(newCtx)
		assert.True(t, ok, "ID should exist in the context")
		assert.Equal(t, resultID, extractedID, "ID in context should match the returned one")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context

		newCtx, resultID := request.EnsureRequestID(ctx)
		require.NotNil(t, newCtx, "EnsureRequestID should create a context")
		assert.True(t, request.IsValidUUID(resultID), "EnsureRequestID should generate a valid UUID")
	})
}

func TestExtractOrGenerateRequestID(t *testing.T) {
	t.Run("with existing request_id", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = request.WithRequestID(ctx, id)

		resultID := request.ExtractOrGenerateRequestID(ctx)
		assert.Equal(t, id, resultID, "ExtractOrGenerateRequestID should return the existing ID")
	})

	t.Run("without request_id", func(t *testing.T) {
		ctx := context.Background()

		resultID := request.ExtractOrGenerateRequestID(ctx)
		assert.True(t, request.IsValidUUID(resultID), "ExtractOrGenerateRequestID should generate a valid UUID")
	})

	t.Run("with nil context", func(t *testing.T) {
		var ctx context.Context

		resultID := request.ExtractOrGenerateRequestID(ctx)
		assert.True(t, request.IsValidUUID(resultID), "ExtractOrGenerateRequestID should generate a valid UUID")
	})
}

func TestLogger(t *testing.T) {
	t.Run("with request_id in context", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = request.WithRequestID(ctx, id)

		baseLogger := newMockLogger()
		getLoggerFunc := func(ctx context.Context) *mockLogger {
			return baseLogger
		}

		result := request.Logger(ctx, getLoggerFunc, fieldConstructor)
		require.NotNil(t, result, "Logger should return a logger")
		require.Len(t, result.fields, 1, "logger should contain one field")
		assert.Equal(t, request.RequestIDFieldName, result.fields[0].key, "field name should be request_id")
		assert.Equal(t, id, result.fields[0].value, "field value should be the ID")
	})

	t.Run("without request_id in context", func(t *testing.T) {
		ctx := context.Background()
		baseLogger := newMockLogger()

		getLoggerFunc := func(ctx context.Context) *mockLogger {
			return baseLogger
		}

		result := request.Logger(ctx, getLoggerFunc, fieldConstructor)
		assert.Equal(t, baseLogger, result, "Logger should return the original logger")
		assert.Empty(t, result.fields, "logger should not contain any fields")
	})

	t.Run("custom field constructor", func(t *testing.T) {
		ctx := context.Background()
		id := "test-id"
		ctx = request.WithRequestID(ctx, id)

		baseLogger := newMockLogger()
		getLoggerFunc := func(ctx context.Context) *mockLogger {
			return baseLogger
		}

		customFieldConstructor := func(key string, value any) mockField {
			return mockField{key: "prefix_" + key, value: value}
		}

		result := request.Logger(ctx, getLoggerFunc, customFieldConstructor)
		require.Len(t, result.fields, 1)
		assert.Equal(t, "prefix_"+request.RequestIDFieldName, result.fields[0].key)
	})
}
