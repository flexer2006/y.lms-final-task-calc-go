package orchestrator

import (
	"net/http"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/handlers/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/midleware"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	orchAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const (
	apiPrefix     = "/api/v1/calculations"
	pathRoot      = "/"
	pathByID      = "/{id}"
	pathHealth    = "/health"
	healthMessage = "Orchestrator service is healthy"
)

func RegisterRoutes(r chi.Router, calcUseCase orchAPI.UseCaseCalculation, authUseCase auth.UseCaseUser) {
	handler := orchestrator.NewHandler(calcUseCase)

	r.Route(apiPrefix, func(r chi.Router) {
		r.Use(chiMiddleware.RequestID)
		r.Use(midleware.Logger)
		r.Use(midleware.Recovery)
		r.Use(midleware.ErrorHandler)
		r.Use(midleware.AuthMiddleware(authUseCase))

		r.Post(pathRoot, handler.CalculateExpression)
		r.Get(pathRoot, handler.ListCalculations)
		r.Get(pathByID, handler.GetCalculation)
		r.Get(pathHealth, healthCheckHandler)
	})
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(healthMessage)); err != nil {
		logger.ContextLogger(r.Context(), nil).Error("Failed to write health check response", zap.Error(err))
	}
}
