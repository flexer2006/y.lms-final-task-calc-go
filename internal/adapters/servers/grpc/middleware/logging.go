package middleware

import (
	"context"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	fieldMethod       = "grpc.method"
	fieldRequestID    = "request.id"
	fieldCode         = "grpc.code"
	fieldDuration     = "duration"
	fieldClientStream = "grpc.stream.is_client_stream"
	fieldServerStream = "grpc.stream.is_server_stream"
)

func UnaryServerLogging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		ctx, requestID := logger.EnsureRequestID(ctx)

		// Создаем дефолтный логгер, если его нет в контексте
		defaultLogger, err := logger.Development()
		if err != nil {
			defaultLogger = logger.Console(logger.InfoLevel, false)
		}

		// Используем defaultLogger как запасной вариант
		log := logger.ContextLogger(ctx, defaultLogger)

		// Добавляем поля и сохраняем как ZapLogger
		logWithFields := log.With(
			zap.String(fieldMethod, info.FullMethod),
			zap.String(fieldRequestID, requestID),
		)

		// Сохраняем логгер в контекст
		ctx = logger.WithLogger(ctx, logWithFields)

		// Используем логгер с полями для последующих записей
		logWithFields.Info("gRPC request started")

		// Вызываем обработчик
		resp, err := handler(ctx, req)

		// Определяем код статуса
		code := extractStatusCode(err)

		// Логируем завершение запроса
		if err != nil {
			logWithFields.Info("gRPC request completed",
				zap.String(fieldCode, code.String()),
				zap.Duration(fieldDuration, time.Since(start)),
				zap.Error(err),
			)
		} else {
			logWithFields.Info("gRPC request completed",
				zap.String(fieldCode, code.String()),
				zap.Duration(fieldDuration, time.Since(start)),
			)
		}

		return resp, err
	}
}

func StreamServerLogging() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		ctx, requestID := logger.EnsureRequestID(ss.Context())

		// Создаем дефолтный логгер, если его нет в контексте
		defaultLogger, err := logger.Development()
		if err != nil {
			defaultLogger = logger.Console(logger.InfoLevel, false)
		}

		// Используем defaultLogger как запасной вариант
		log := logger.ContextLogger(ctx, defaultLogger)

		// Добавляем поля и сохраняем
		logWithFields := log.With(
			zap.String(fieldMethod, info.FullMethod),
			zap.String(fieldRequestID, requestID),
			zap.Bool(fieldClientStream, info.IsClientStream),
			zap.Bool(fieldServerStream, info.IsServerStream),
		)

		// Создаем обертку для ServerStream с обновленным контекстом
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          logger.WithLogger(ctx, logWithFields),
		}

		logWithFields.Info("gRPC stream started")

		// Вызываем обработчик с оберткой
		err = handler(srv, wrappedStream)

		// Определяем код статуса
		code := extractStatusCode(err)

		// Логируем завершение потока
		if err != nil {
			logWithFields.Info("gRPC stream completed",
				zap.String(fieldCode, code.String()),
				zap.Duration(fieldDuration, time.Since(start)),
				zap.Error(err),
			)
		} else {
			logWithFields.Info("gRPC stream completed",
				zap.String(fieldCode, code.String()),
				zap.Duration(fieldDuration, time.Since(start)),
			)
		}

		return err
	}
}

func extractStatusCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	if st, ok := status.FromError(err); ok {
		return st.Code()
	}

	return codes.Internal
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
