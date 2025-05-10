package midleware

import (
	"net/http"
	"runtime/debug"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
)

const (
	panicRecoveryMessage = "Internal server error occurred"
	panicErrorCode       = "SERVER_PANIC"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Explicitly capture the context at the beginning
		ctx := r.Context()

		defer func() {
			if recovered := recover(); recovered != nil {
				stack := debug.Stack()

				log := logger.ContextLogger(ctx, nil)
				log.Error("HTTP handler panic recovered",
					zap.Any("panic", recovered),
					zap.String("stack", string(stack)),
					zap.String("url", r.URL.String()),
					zap.String("method", r.Method),
					zap.String("remote_addr", r.RemoteAddr),
				)

				apiErr := NewAPIError(panicRecoveryMessage, panicErrorCode)

				HandleError(ctx, w, apiErr, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
