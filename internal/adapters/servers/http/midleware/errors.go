package midleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("response writer error: %w", err)
	}
	return n, nil
}

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		if ww.statusCode < http.StatusBadRequest {
			return
		}

		logger.ContextLogger(r.Context(), nil).Error("HTTP error response",
			zap.Int("status_code", ww.statusCode),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method))
	})
}

func HandleError(ctx context.Context, w http.ResponseWriter, err error, statusCode int) {
	response := ErrorResponse{}

	// Check if the error is our custom APIError type
	var apiErr APIError
	if errors.As(err, &apiErr) {
		response.Error.Message = apiErr.Message
		response.Error.Code = apiErr.Code
	} else {
		// For standard errors
		response.Error.Message = err.Error()
		response.Error.Code = "INTERNAL_ERROR"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		// If we failed to encode the error response, log it
		log := logger.ContextLogger(ctx, nil)
		log.Error("failed to encode error response",
			zap.Error(encodeErr),
			zap.Error(err),
			zap.Int("status_code", statusCode))
	}
}
