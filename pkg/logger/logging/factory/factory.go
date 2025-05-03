// Package factory предоставляет функции для создания различных типов журнала.
package factory

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/core"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
)

// Константы для ошибок создания журналов.
const (
	errBuildDevLogger  = "failed to build development logger"
	errBuildProdLogger = "failed to build production logger"
)

// New создает новый журнал с указанными настройками ядра.
func New(core zapcore.Core) *logging.Logger {
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

	return logging.NewLogger(zapLogger, atomicLevel)
}

// Console создает журнал с выводом в консоль.
func Console(lvl level.LogLevel, json bool) *logging.Logger {
	encoder := core.CreateEncoder(json)
	atomicLevel := zap.NewAtomicLevelAt(lvl.ToZapLevel())

	zapCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	zapLogger := zap.New(zapCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logging.NewLogger(zapLogger, atomicLevel)
}

// Development создает журнал для разработки.
func Development() (*logging.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errBuildDevLogger, err)
	}

	return logging.NewLogger(zapLogger, zap.NewAtomicLevelAt(zapcore.DebugLevel)), nil
}

// Production создает журнал для релиза продукта.
func Production() (*logging.Logger, error) {
	cfg := zap.NewProductionConfig()
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errBuildProdLogger, err)
	}

	return logging.NewLogger(zapLogger, zap.NewAtomicLevelAt(zapcore.InfoLevel)), nil
}
