package auth

import (
	"net/http"

	authHandlers "github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/handlers/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/midleware"
	authAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const (
	apiPrefix = "/api/v1/auth"

	pathRegister = "/register"
	pathLogin    = "/login"
	pathRefresh  = "/refresh"
	pathLogout   = "/logout"
	pathHealth   = "/health"

	healthMessage = "Auth service is healthy"
)

func RegisterRoutes(r chi.Router, authUseCase authAPI.UseCaseUser) {
	handler := authHandlers.NewHandler(authUseCase)

	r.Route(apiPrefix, func(r chi.Router) {
		r.Use(chiMiddleware.RequestID)
		r.Use(midleware.Logger)
		r.Use(midleware.Recovery)
		r.Use(midleware.ErrorHandler)

		r.Post(pathRegister, handler.Register)
		r.Post(pathLogin, handler.Login)
		r.Post(pathRefresh, handler.RefreshToken)
		r.Get(pathHealth, healthCheckHandler)

		r.Group(func(r chi.Router) {
			r.Use(midleware.AuthMiddleware(authUseCase))
			r.Post(pathLogout, handler.Logout)
		})
	})
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(healthMessage)); err != nil {
		logger.ContextLogger(r.Context(), nil).Error("Failed to write health check response", zap.Error(err))
	}
}
