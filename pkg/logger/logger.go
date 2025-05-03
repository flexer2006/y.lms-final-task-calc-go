// Package logger Пакет logger предоставляет единую точку входа для всей функциональности логирования.
package logger

import (
	"context"
	"fmt"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/ctxlog"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging"
	levelPkg "github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/request"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Константы, реэкспортированные из внутренних компонентов.
const (
	// RequestIDField имя поля для идентификатора запроса.
	RequestIDField = request.RequestIDFieldName

	// ErrSyncLogger сообщения об ошибках.
	ErrSyncLogger        = logging.ErrSyncLogger
	ErrSyncContextLogger = ctxlog.ErrSyncContextLogger

	errMsgSyncLogger        = "failed to sync logger"
	errMsgSyncContextLogger = "failed to sync context logger"
	errMsgCreateDevLogger   = "failed to create development logger"
	errMsgCreateProdLogger  = "failed to create production logger"
)

// LogLevel представляет уровни логирования, реэкспортированные из пакета ctxlog.
type LogLevel = ctxlog.LogLevel

// Уровни логирования, реэкспортированные из ctxlog.
const (
	DebugLevel = ctxlog.DebugLevel
	InfoLevel  = ctxlog.InfoLevel
	WarnLevel  = ctxlog.WarnLevel
	ErrorLevel = ctxlog.ErrorLevel
	FatalLevel = ctxlog.FatalLevel
)

// Field представляет общий тип для полей лога.
type Field = ctxlog.Field

// Logger использует непосредственно ctxlog Logger для совместимости интерфейсов.
type Logger = ctxlog.Logger

// ZapLogger расширяет интерфейс Logger функциональностью, специфичной для zap.
type ZapLogger interface {
	Logger
	RawLogger() *zap.Logger
}

// zapAdapter адаптирует logging.Logger для реализации интерфейса Logger.
type zapAdapter struct {
	logger *logging.Logger
}

// Проверяем, что zapAdapter реализует интерфейс Logger.
var _ Logger = (*zapAdapter)(nil)

// Debug фиксирует сообщение с уровнем Debug.
func (z *zapAdapter) Debug(msg string, fields ...Field) {
	zapFields := convertToZapFields(fields)
	z.logger.Debug(msg, zapFields...)
}

// Info фиксирует сообщение с уровнем Info.
func (z *zapAdapter) Info(msg string, fields ...Field) {
	zapFields := convertToZapFields(fields)
	z.logger.Info(msg, zapFields...)
}

// Warn фиксирует сообщение с уровнем Warn.
func (z *zapAdapter) Warn(msg string, fields ...Field) {
	zapFields := convertToZapFields(fields)
	z.logger.Warn(msg, zapFields...)
}

// Error фиксирует сообщение с уровнем Error.
func (z *zapAdapter) Error(msg string, fields ...Field) {
	zapFields := convertToZapFields(fields)
	z.logger.Error(msg, zapFields...)
}

// Fatal фиксирует сообщение с уровнем Fatal.
func (z *zapAdapter) Fatal(msg string, fields ...Field) {
	zapFields := convertToZapFields(fields)
	z.logger.Fatal(msg, zapFields...)
}

// With создает новый журнал с дополнительными полями.
func (z *zapAdapter) With(fields ...Field) Logger {
	zapFields := convertToZapFields(fields)
	if innerLogger, ok := z.logger.With(zapFields...).(*logging.Logger); ok {
		return &zapAdapter{logger: innerLogger}
	}
	return &zapAdapter{logger: z.logger}
}

// SetLevel устанавливает уровень логирования.
func (z *zapAdapter) SetLevel(lvl LogLevel) {
	z.logger.SetLevel(convertToLoggingLevel(lvl))
}

// GetLevel возвращает текущий уровень логирования.
func (z *zapAdapter) GetLevel() LogLevel {
	return convertFromLoggingLevel(z.logger.GetLevel())
}

// Sync сбрасывает буферизованные записи лога.
func (z *zapAdapter) Sync() error {
	if err := z.logger.Sync(); err != nil {
		return fmt.Errorf("%s: %w", errMsgSyncLogger, err)
	}
	return nil
}

// RawLogger возвращает нижележащий zap журнал.
func (z *zapAdapter) RawLogger() *zap.Logger {
	return z.logger.RawLogger()
}

// convertToZapFields конвертирует ctxlog.Field в zap.Field.
func convertToZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		switch f := field.(type) {
		case zap.Field:
			zapFields = append(zapFields, f)
		case string:
			zapFields = append(zapFields, zap.String("field", f))
		default:
			zapFields = append(zapFields, zap.Any("field", f))
		}
	}
	return zapFields
}

// convertToLoggingLevel конвертирует ctxlog.LogLevel в logging.level.LogLevel.
func convertToLoggingLevel(lvl LogLevel) levelPkg.LogLevel {
	switch lvl {
	case DebugLevel:
		return levelPkg.DebugLevel
	case InfoLevel:
		return levelPkg.InfoLevel
	case WarnLevel:
		return levelPkg.WarnLevel
	case ErrorLevel:
		return levelPkg.ErrorLevel
	case FatalLevel:
		return levelPkg.FatalLevel
	default:
		return levelPkg.InfoLevel
	}
}

// convertFromLoggingLevel конвертирует logging.level.LogLevel в ctxlog.LogLevel.
func convertFromLoggingLevel(lvl levelPkg.LogLevel) LogLevel {
	switch lvl {
	case levelPkg.DebugLevel:
		return DebugLevel
	case levelPkg.InfoLevel:
		return InfoLevel
	case levelPkg.WarnLevel:
		return WarnLevel
	case levelPkg.ErrorLevel:
		return ErrorLevel
	case levelPkg.FatalLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}

// New создает новый журнал с заданным ядром.
func New(core zapcore.Core) ZapLogger {
	return &zapAdapter{
		logger: logging.New(core),
	}
}

// Console создает консольный журнал.
func Console(lvl LogLevel, json bool) ZapLogger {
	loggingLevel := convertToLoggingLevel(lvl)
	return &zapAdapter{
		logger: logging.Console(loggingLevel, json),
	}
}

// Development создает журнал для разработки.
func Development() (ZapLogger, error) {
	logger, err := logging.Development()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMsgCreateDevLogger, err)
	}
	return &zapAdapter{logger: logger}, nil
}

// Production создает журнал для продуктовой версии.
func Production() (ZapLogger, error) {
	logger, err := logging.Production()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMsgCreateDevLogger, err)
	}
	return &zapAdapter{logger: logger}, nil
}

// WithLogger добавляет журнал в контекст.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return ctxlog.WithLogger(ctx, logger)
}

// FromContext извлекает журнал из контекста.
func FromContext(ctx context.Context) (Logger, bool) {
	return ctxlog.FromContext(ctx)
}

// GetLogger получает журнал из контекста или возвращает журнал по умолчанию.
func GetLogger(ctx context.Context, defaultLogger Logger) Logger {
	return ctxlog.GetLogger(ctx, defaultLogger)
}

// WithFields добавляет поля к журналу в контексте.
func WithFields(ctx context.Context, defaultLogger Logger, fields ...Field) context.Context {
	return ctxlog.WithFields(ctx, defaultLogger, fields...)
}

// WithRequestID добавляет идентификатор запроса в контекст.
func WithRequestID(ctx context.Context, id string) context.Context {
	return request.WithRequestID(ctx, id)
}

// RequestID извлекает идентификатор запроса из контекста.
func RequestID(ctx context.Context) (string, bool) {
	return request.ID(ctx)
}

// GenerateRequestID генерирует новый идентификатор запроса.
func GenerateRequestID() string {
	return request.GenerateRequestID()
}

// EnsureRequestID обеспечивает наличие идентификатора запроса в контексте.
func EnsureRequestID(ctx context.Context) (context.Context, string) {
	return request.EnsureRequestID(ctx)
}

// WithRequestIDField добавляет поле идентификатора запроса к журналу.
func WithRequestIDField(ctx context.Context, logger ZapLogger) ZapLogger {
	id, ok := request.ID(ctx)
	if !ok || id == "" {
		return logger
	}

	zapLogger := logger.(*zapAdapter)
	zapField := zap.String(request.RequestIDFieldName, id)

	if innerLogger, ok := zapLogger.logger.With(zapField).(*logging.Logger); ok {
		return &zapAdapter{
			logger: innerLogger,
		}
	}
	return logger
}

// ContextLogger получает или создает журнал с идентификатором запроса из контекста.
func ContextLogger(ctx context.Context, defaultLogger ZapLogger) ZapLogger {
	if ctxLogger, ok := FromContext(ctx); ok {
		if zapLogger, ok := ctxLogger.(ZapLogger); ok {
			return WithRequestIDField(ctx, zapLogger)
		}
	}

	return WithRequestIDField(ctx, defaultLogger)
}

// Log фиксирует сообщение с указанным уровнем, используя журнал из контекста.
func Log(ctx context.Context, defaultLogger Logger, level LogLevel, msg string, fields ...Field) {
	ctxlog.Logg(ctx, defaultLogger, level, msg, fields...)
}

// Debug фиксирует сообщение уровня Debug, используя журнал из контекста.
func Debug(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	ctxlog.Debug(ctx, defaultLogger, msg, fields...)
}

// Info фиксирует сообщение уровня Info, используя журнал из контекста.
func Info(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	ctxlog.Info(ctx, defaultLogger, msg, fields...)
}

// Warn фиксирует сообщение уровня Warn, используя журнал из контекста.
func Warn(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	ctxlog.Warn(ctx, defaultLogger, msg, fields...)
}

// Error фиксирует сообщение уровня Error, используя журнал из контекста.
func Error(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	ctxlog.Error(ctx, defaultLogger, msg, fields...)
}

// Fatal фиксирует сообщение уровня Fatal, используя журнал из контекста.
func Fatal(ctx context.Context, defaultLogger Logger, msg string, fields ...Field) {
	ctxlog.Fatal(ctx, defaultLogger, msg, fields...)
}

// Sync сбрасывает буферизованные записи лога из контекста.
func Sync(ctx context.Context, defaultLogger Logger) error {
	if err := ctxlog.Sync(ctx, defaultLogger); err != nil {
		return fmt.Errorf("%s: %w", errMsgSyncContextLogger, err)
	}
	return nil
}
