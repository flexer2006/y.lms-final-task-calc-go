// Package ctxlog предоставляет утилиты для работы с логгерами в контексте.
package ctxlog

import (
	"context"
	"fmt"
)

// ErrSyncContextLogger Константа для сообщения об ошибке.
const (
	ErrSyncContextLogger = "failed to sync context logger"
)

// Field представляет поле логирования.
type Field any

// LogLevel представляет уровень логирования.
type LogLevel uint8

// Уровни логирования с использованием iota.
const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Logger определяет интерфейс для логгера, совместимый с различными реализациями.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	Sync() error
}

// loggerKey для хранения журнала в контексте.
type loggerKey struct{}

// WithLogger добавляет логгер в контекст.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	if ctx == nil || logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext возвращает логгер из контекста. Если в контексте нет логгера,
// возвращается (nil, false).
func FromContext(ctx context.Context) (Logger, bool) {
	if ctx == nil {
		return nil, false
	}

	logger, ok := ctx.Value(loggerKey{}).(Logger)
	return logger, ok
}

// GetLogger возвращает логгер из контекста или defaultLogger, если логгер не найден.
func GetLogger(ctx context.Context, defaultLogger Logger) Logger {
	if logger, ok := FromContext(ctx); ok && logger != nil {
		return logger
	}
	return defaultLogger
}

// WithFields добавляет поля к логгеру в контексте.
func WithFields(ctx context.Context, defaultLogger Logger, fields ...Field) context.Context {
	logger := GetLogger(ctx, defaultLogger)
	if logger == nil {
		return ctx
	}
	return WithLogger(ctx, logger.With(fields...))
}

// Logg унифицированный метод логирования, использующий логгер из контекста.
func Logg(ctx context.Context, defaultLogger Logger, level LogLevel, msg string, fields ...Field) {
	logger := GetLogger(ctx, defaultLogger)
	if logger == nil {
		return
	}

	switch level {
	case DebugLevel:
		logger.Debug(msg, fields...)
	case InfoLevel:
		logger.Info(msg, fields...)
	case WarnLevel:
		logger.Warn(msg, fields...)
	case ErrorLevel:
		logger.Error(msg, fields...)
	case FatalLevel:
		logger.Fatal(msg, fields...)
	default:
		logger.Info(msg, fields...)
	}
}

// Debug логирует сообщение с уровнем Debug, используя логгер из контекста.
func Debug(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	logger := GetLogger(ctx, defaultLogger)
	if logger != nil {
		logger.Debug(msg, fields...)
	}
}

// Info логирует сообщение с уровнем Info, используя логгер из контекста.
func Info(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	logger := GetLogger(ctx, defaultLogger)
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Warn логирует сообщение с уровнем Warn, используя логгер из контекста.
func Warn(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	logger := GetLogger(ctx, defaultLogger)
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// Error логирует сообщение с уровнем Error, используя логгер из контекста.
func Error(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	logger := GetLogger(ctx, defaultLogger)
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// Fatal логирует сообщение с уровнем Fatal, используя логгер из контекста.
func Fatal(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	logger := GetLogger(ctx, defaultLogger)
	if logger != nil {
		logger.Fatal(msg, fields...)
	}
}

// Sync сбрасывает буферизованные записи логгера из контекста.
func Sync(ctx context.Context, defaultLogger Logger) error {
	logger := GetLogger(ctx, defaultLogger)
	if logger != nil {
		if err := logger.Sync(); err != nil {
			return fmt.Errorf("%s: %w", ErrSyncContextLogger, err)
		}
	}
	return nil
}
