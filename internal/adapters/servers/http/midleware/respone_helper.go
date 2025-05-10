package midleware

import "net/http"

// responseWriterWrapper wraps http.ResponseWriter to capture status code.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}
