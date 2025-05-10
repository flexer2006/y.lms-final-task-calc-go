package midleware

import (
	"net/http"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	headerRequestID = "X-Request-ID"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		requestID := r.Header.Get(headerRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set(headerRequestID, requestID)
		}

		w.Header().Set(headerRequestID, requestID)

		ctx := logger.WithRequestID(r.Context(), requestID)

		defaultLogger, err := logger.Development()
		if err != nil {
			defaultLogger = logger.Console(logger.InfoLevel, false)
		}

		log := logger.ContextLogger(ctx, defaultLogger)

		log = log.With(
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		).(logger.ZapLogger)

		ctx = logger.WithLogger(ctx, log)
		r = r.WithContext(ctx)

		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		log.Info("Request started")

		next.ServeHTTP(ww, r)

		duration := time.Since(startTime)
		log.Info("Request completed",
			zap.Int("status", ww.statusCode),
			zap.Duration("duration", duration),
		)
	})
}
