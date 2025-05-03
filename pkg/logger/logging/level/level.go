// Package level предоставляет типы и функции для работы с уровнями логирования.
package level

import (
	"go.uber.org/zap/zapcore"
)

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

const (
	infoLVL    = "infoLVL"
	debugLVL   = "debugLVL"
	warnLVL    = "warnLVL"
	errorLVL   = "error"
	fatalLVL   = "fatal"
	unknownLVL = "unknown"
)
const (
	zapInfoLevel  = zapcore.InfoLevel
	zapDebugLevel = zapcore.DebugLevel
	zapWarnLevel  = zapcore.WarnLevel
	zapErrorLevel = zapcore.ErrorLevel
	zapFatalLevel = zapcore.FatalLevel
)

// String возвращает строковое представление уровня логирования.
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return debugLVL
	case InfoLevel:
		return infoLVL
	case WarnLevel:
		return warnLVL
	case ErrorLevel:
		return errorLVL
	case FatalLevel:
		return fatalLVL
	default:
		return unknownLVL
	}
}

// ToZapLevel преобразует LogLevel в zap core Level.
func (l LogLevel) ToZapLevel() zapcore.Level {
	switch l {
	case DebugLevel:
		return zapDebugLevel
	case InfoLevel:
		return zapInfoLevel
	case WarnLevel:
		return zapWarnLevel
	case ErrorLevel:
		return zapErrorLevel
	case FatalLevel:
		return zapFatalLevel
	default:
		return zapInfoLevel
	}
}

// Parse конвертирует строку в LogLevel.
func Parse(level string) LogLevel {
	switch level {
	case debugLVL:
		return DebugLevel
	case infoLVL:
		return InfoLevel
	case warnLVL:
		return WarnLevel
	case errorLVL:
		return ErrorLevel
	case fatalLVL:
		return FatalLevel
	default:
		return InfoLevel
	}
}

// FromZapLevel преобразует zap core Level в LogLevel.
func FromZapLevel(level zapcore.Level) LogLevel {
	switch level {
	case zapDebugLevel:
		return DebugLevel
	case zapInfoLevel:
		return InfoLevel
	case zapWarnLevel:
		return WarnLevel
	case zapErrorLevel:
		return ErrorLevel
	case zapFatalLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}
