package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	fieldPanic      = "panic"
	fieldStack      = "stack"
	fieldMethodName = "grpc.method"
	msgPanicRecover = "Recovered from panic"
	errInternal     = "Internal server error"
	errUnknownPanic = "unknown panic"
)

func UnaryServerRecovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = handlePanic(ctx, r, info.FullMethod)
			}
		}()
		return handler(ctx, req)
	}
}

func StreamServerRecovery() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = handlePanic(ss.Context(), r, info.FullMethod)
			}
		}()
		return handler(srv, ss)
	}
}

func handlePanic(ctx context.Context, r any, method string) error {
	stack := debug.Stack()
	log := logger.ContextLogger(ctx, nil)

	log.Error(msgPanicRecover,
		zap.Any(fieldPanic, r),
		zap.String(fieldStack, string(stack)),
		zap.String(fieldMethodName, method),
	)

	var errorMsg string
	switch v := r.(type) {
	case string:
		errorMsg = v
	case error:
		errorMsg = v.Error()
	default:
		errorMsg = errUnknownPanic
	}

	return fmt.Errorf("gRPC recovery: %w", status.Error(codes.Internal, fmt.Sprintf("%s: %s", errInternal, errorMsg)))
}
