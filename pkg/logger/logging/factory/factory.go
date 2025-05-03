// Package factory предоставляет функции для создания различных типов журнала.
package factory

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/core"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
)

// Logger представляет журнал.
type Logger struct {
	zapLogger *zap.Logger
	level     zap.AtomicLevel
}

// Константы для ошибок создания журналов.
const (
	ErrBuildDevLogger  = "failed to build development logger"
	ErrBuildProdLogger = "failed to build production logger"
)

// NewLogger создает новый журнал с заданным zap Logger и уровнем.
func NewLogger(zapLogger *zap.Logger, level zap.AtomicLevel) *Logger {
	return &Logger{
		zapLogger: zapLogger,
		level:     level,
	}
}

// New создает новый журнал с указанными настройками ядра.
func New(core zapcore.Core) *Logger {
	zapLogger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	var atomicLevel zap.AtomicLevel
	if le, ok := core.(interface{ Level() zapcore.Level }); ok {
		atomicLevel = zap.NewAtomicLevelAt(le.Level())
	} else {
		atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	return NewLogger(zapLogger, atomicLevel)
}

// Console создает журнал с выводом в консоль.
func Console(lvl level.LogLevel, json bool) *Logger {
	encoder := core.CreateEncoder(json)
	atomicLevel := zap.NewAtomicLevelAt(lvl.ToZapLevel())

	zapCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	zapLogger := zap.New(zapCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return NewLogger(zapLogger, atomicLevel)
}

// Development создает журнал для разработки.
func Development() (*Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrBuildDevLogger, err)
	}

	return NewLogger(zapLogger, zap.NewAtomicLevelAt(zapcore.DebugLevel)), nil
}

// Production создает журнал для релиза продукта.
func Production() (*Logger, error) {
	cfg := zap.NewProductionConfig()
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrBuildProdLogger, err)
	}

	return NewLogger(zapLogger, zap.NewAtomicLevelAt(zapcore.InfoLevel)), nil
}

// GetZapLogger возвращает нижележащий zap logger.
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}

// GetAtomicLevel возвращает атомарный уровень логирования.
func (l *Logger) GetAtomicLevel() zap.AtomicLevel {
	return l.level
}
