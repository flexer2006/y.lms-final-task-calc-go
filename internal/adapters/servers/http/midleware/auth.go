package midleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	authHeaderName   = "Authorization"
	bearerScheme     = "Bearer"
	bearerPartsCount = 2
)

type userIDContextKey struct{}

type APIError struct {
	Message string
	Code    string
}

func (e APIError) Error() string {
	return e.Message
}

func NewAPIError(message, code string) APIError {
	return APIError{
		Message: message,
		Code:    code,
	}
}

var (
	ErrMissingToken      = NewAPIError("missing authentication token", "AUTH_MISSING_TOKEN")
	ErrInvalidAuthHeader = NewAPIError("invalid authorization header format", "AUTH_INVALID_HEADER")
	ErrInvalidToken      = NewAPIError("invalid or expired token", "AUTH_INVALID_TOKEN")
	ErrUserNotInContext  = NewAPIError("user ID not found in context", "AUTH_NO_USER_CONTEXT")
)

func AuthMiddleware(authUseCase auth.UseCaseUser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(authHeaderName)
			if authHeader == "" {
				HandleError(r.Context(), w, ErrMissingToken, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != bearerPartsCount || parts[0] != bearerScheme {
				HandleError(r.Context(), w, ErrInvalidAuthHeader, http.StatusUnauthorized)
				return
			}

			userID, err := authUseCase.ValidateToken(r.Context(), parts[1])
			if err != nil {
				logger.ContextLogger(r.Context(), nil).Error("token validation failed", zap.Error(err))
				HandleError(r.Context(), w, ErrInvalidToken, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDContextKey{}).(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrUserNotInContext
	}
	return userID, nil
}
