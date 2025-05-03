// Package request предоставляет функциональность для работы с идентификаторами запросов в логах.
package request

import (
	"context"
	"regexp"

	"github.com/google/uuid"
)

// Константы для работы с идентификаторами запросов.
const (
	// RequestIDFieldName имя поля для идентификатора запроса в логах.
	RequestIDFieldName = "request_id"
)

// ctxKey тип для ключей контекста, использующий строгую типизацию.
type ctxKey struct{ name string }

// RequestIDKey ключ для идентификатора запроса в контексте.
var RequestIDKey = &ctxKey{RequestIDFieldName}

// Регулярное выражение для проверки UUID.
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// WithRequestID добавляет идентификатор запроса в контекст.
func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, RequestIDKey, id)
}

// ID извлекает идентификатор запроса из контекста.
func ID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(RequestIDKey).(string)
	return id, ok
}

// GenerateRequestID генерирует новый идентификатор запроса.
func GenerateRequestID() string {
	return uuid.New().String()
}

// IsValidUUID проверяет, соответствует ли строка формату UUID.
func IsValidUUID(id string) bool {
	return uuidRegex.MatchString(id)
}

// WithRequestIDField добавляет поле request_id к логгеру, если оно присутствует в контексте.
// Logger представляет любой объект, который поддерживает метод With для добавления полей.
func WithRequestIDField[T any, L interface {
	With(fields ...T) L
}](ctx context.Context, logger L, fieldConstructor func(key string, value any) T) L {
	if ctx == nil {
		return logger
	}

	id, ok := ID(ctx)
	if !ok || id == "" {
		return logger
	}

	field := fieldConstructor(RequestIDFieldName, id)
	return logger.With(field)
}

// EnsureRequestID проверяет наличие идентификатора запроса в контексте и добавляет новый, если отсутствует.
func EnsureRequestID(ctx context.Context) (context.Context, string) {
	if ctx == nil {
		ctx = context.Background()
	}

	if id, ok := ID(ctx); ok && id != "" {
		return ctx, id
	}

	id := GenerateRequestID()
	return WithRequestID(ctx, id), id
}

// ExtractOrGenerateRequestID получает ID из контекста или генерирует новый.
func ExtractOrGenerateRequestID(ctx context.Context) string {
	id, ok := ID(ctx)
	if !ok || id == "" {
		return GenerateRequestID()
	}
	return id
}

// Logger получает логгер из контекста и добавляет к нему request_id.
func Logger[T any, L interface {
	With(fields ...T) L
}](ctx context.Context, getLoggerFunc func(context.Context) L, fieldConstructor func(key string, value interface{}) T) L {
	logger := getLoggerFunc(ctx)
	return WithRequestIDField(ctx, logger, fieldConstructor)
}
