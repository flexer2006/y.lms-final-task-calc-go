// Package core предоставляет функции для создания и настройки ядра логирования.
package core

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger/logging/level"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New создает новое ядро логирования.
func New(encoder zapcore.Encoder, writer zapcore.WriteSyncer, lvl level.LogLevel) zapcore.Core {
	return zapcore.NewCore(
		encoder,
		writer,
		zap.NewAtomicLevelAt(lvl.ToZapLevel()),
	)
}

// CreateEncoder создает кодировщик с заданными настройками.
func CreateEncoder(json bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if json {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}
