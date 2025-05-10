package middleware

import (
	"context"
	"errors"
	"fmt"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrorMapping = map[error]codes.Code{
	domainerrors.ErrUserAlreadyExists:   codes.AlreadyExists,
	domainerrors.ErrInvalidCredentials:  codes.Unauthenticated,
	domainerrors.ErrUserNotFound:        codes.NotFound,
	domainerrors.ErrInvalidToken:        codes.Unauthenticated,
	domainerrors.ErrTokenExpired:        codes.Unauthenticated,
	domainerrors.ErrTokenNotFound:       codes.NotFound,
	domainerrors.ErrTokenRevoked:        codes.Unauthenticated,
	domainerrors.ErrInternalServerError: codes.Internal,
}

func UnaryServerError() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			err = mapError(ctx, err)
		}
		return resp, err
	}
}

func StreamServerError() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err != nil {
			err = mapError(ss.Context(), err)
		}
		return err
	}
}

func mapError(ctx context.Context, err error) error {
	if st, ok := status.FromError(err); ok {
		return fmt.Errorf("gRPC status error: %w", st.Err())
	}

	log := logger.ContextLogger(ctx, nil)

	statusCode := codes.Unknown
	for domainErr, code := range ErrorMapping {
		if errors.Is(err, domainErr) {
			statusCode = code
			break
		}
	}

	if statusCode == codes.Unknown {
		log.Error("unhandled error", zap.Error(err))
		statusCode = codes.Internal
	}

	return fmt.Errorf("mapped gRPC error: %w", status.Error(statusCode, err.Error()))
}
