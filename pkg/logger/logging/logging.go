// Package logging предоставляет базовую функциональность логирования.
package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/factory"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
)

// ErrSyncLogger Константа для сообщения об ошибке.
const (
	ErrSyncLogger = "failed to sync logger"
)

// LoggerInterface определяет интерфейс для журнала.
type LoggerInterface interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) LoggerInterface
	SetLevel(level level.LogLevel)
	GetLevel() level.LogLevel
	Sync() error
}

// Logger предоставляет интерфейс для логирования.
type Logger struct {
	zapLogger *zap.Logger
	level     zap.AtomicLevel
}

// Убедимся, что Logger реализует LoggerInterface.
var _ LoggerInterface = (*Logger)(nil)

// New создает новый журнал с указанными настройками ядра.
func New(core zapcore.Core) *Logger {
	factoryLogger := factory.New(core)
	return &Logger{
		zapLogger: factoryLogger.GetZapLogger(),
		level:     factoryLogger.GetAtomicLevel(),
	}
}

// Console создает журнал с выводом в консоль.
func Console(lvl level.LogLevel, json bool) *Logger {
	factoryLogger := factory.Console(lvl, json)
	return &Logger{
		zapLogger: factoryLogger.GetZapLogger(),
		level:     factoryLogger.GetAtomicLevel(),
	}
}

// Development создает журнал для разработки.
func Development() (*Logger, error) {
	factoryLogger, err := factory.Development()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", factory.ErrBuildDevLogger, err)
	}
	return &Logger{
		zapLogger: factoryLogger.GetZapLogger(),
		level:     factoryLogger.GetAtomicLevel(),
	}, nil
}

// Production создает журнал для релиза продукта.
func Production() (*Logger, error) {
	factoryLogger, err := factory.Production()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", factory.ErrBuildProdLogger, err)
	}
	return &Logger{
		zapLogger: factoryLogger.GetZapLogger(),
		level:     factoryLogger.GetAtomicLevel(),
	}, nil
}

// With создаёт новый журнал с дополнительными полями.
func (l *Logger) With(fields ...zap.Field) LoggerInterface {
	return &Logger{
		zapLogger: l.zapLogger.With(fields...),
		level:     l.level,
	}
}

// SetLevel изменяет уровень логирования.
func (l *Logger) SetLevel(lvl level.LogLevel) {
	l.level.SetLevel(lvl.ToZapLevel())
}

// GetLevel возвращает текущий уровень логирования.
func (l *Logger) GetLevel() level.LogLevel {
	return level.FromZapLevel(l.level.Level())
}

// Logg унифицированный метод логирования.
func (l *Logger) Logg(level zapcore.Level, msg string, fields ...zap.Field) {
	if l == nil || l.zapLogger == nil {
		return
	}

	switch level {
	case zapcore.DebugLevel:
		l.zapLogger.Debug(msg, fields...)
	case zapcore.InfoLevel:
		l.zapLogger.Info(msg, fields...)
	case zapcore.WarnLevel:
		l.zapLogger.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		l.zapLogger.Error(msg, fields...)
	case zapcore.FatalLevel:
		l.zapLogger.Fatal(msg, fields...)
	default:
		l.zapLogger.Info(msg, fields...)
	}
}

// Debug записывает сообщение с уровнем Debug.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logg(zapcore.DebugLevel, msg, fields...)
}

// Info записывает сообщение с уровнем Info.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logg(zapcore.InfoLevel, msg, fields...)
}

// Warn записывает сообщение с уровнем Warn.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logg(zapcore.WarnLevel, msg, fields...)
}

// Error записывает сообщение с уровнем Error.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logg(zapcore.ErrorLevel, msg, fields...)
}

// Fatal записывает сообщение с уровнем Fatal и завершает работу программы.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logg(zapcore.FatalLevel, msg, fields...)
}

// Sync сбрасывает буферизованные записи журнала.
func (l *Logger) Sync() error {
	if l == nil || l.zapLogger == nil {
		return nil
	}

	if err := l.zapLogger.Sync(); err != nil {
		return fmt.Errorf("%s: %w", ErrSyncLogger, err)
	}

	return nil
}

// RawLogger возвращает нижележащий zap Logger для расширенной функциональности.
func (l *Logger) RawLogger() *zap.Logger {
	return l.zapLogger
}
