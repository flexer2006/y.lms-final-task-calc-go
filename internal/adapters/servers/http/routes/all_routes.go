package routes

import (
	"net/http"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/handlers/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/handlers/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/midleware"
	authAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	orchAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

const (
	apiVersion = "/api/v1"

	authPrefix   = apiVersion + "/auth"
	pathRegister = "/register"
	pathLogin    = "/login"
	pathRefresh  = "/refresh"
	pathLogout   = "/logout"

	calcPrefix = apiVersion + "/calculations"
	pathRoot   = "/"
	pathByID   = "/{id}"

	pathHealth    = "/health"
	apiHealthMsg  = "API Gateway is healthy"
	authHealthMsg = "Auth service is healthy"
	calcHealthMsg = "Orchestrator service is healthy"
)

func NewRouter(authUseCase authAPI.UseCaseUser, calcUseCase orchAPI.UseCaseCalculation) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Root health check
	r.Get(pathHealth, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(apiHealthMsg)); err != nil {
			// Log the error
			logger.ContextLogger(r.Context(), nil).Error("Failed to write health check response", zap.Error(err))
		}
	})

	// Auth routes
	registerAuthRoutes(r, authUseCase)

	// Calculation routes
	registerCalculationRoutes(r, calcUseCase, authUseCase)

	return r
}

func registerAuthRoutes(r chi.Router, authUseCase authAPI.UseCaseUser) {
	authHandler := auth.NewHandler(authUseCase)

	r.Route(authPrefix, func(r chi.Router) {
		r.Use(chiMiddleware.RequestID)
		r.Use(midleware.Logger)
		r.Use(midleware.Recovery)
		r.Use(midleware.ErrorHandler)

		r.Post(pathRegister, authHandler.Register)
		r.Post(pathLogin, authHandler.Login)
		r.Post(pathRefresh, authHandler.RefreshToken)
		r.Get(pathHealth, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(authHealthMsg)); err != nil {
				logger.ContextLogger(r.Context(), nil).Error("Failed to write health check response", zap.Error(err))
			}
		})

		r.Group(func(r chi.Router) {
			r.Use(midleware.AuthMiddleware(authUseCase))
			r.Post(pathLogout, authHandler.Logout)
		})
	})
}

func registerCalculationRoutes(r chi.Router, calcUseCase orchAPI.UseCaseCalculation, authUseCase authAPI.UseCaseUser) {
	calcHandler := orchestrator.NewHandler(calcUseCase)

	r.Route(calcPrefix, func(r chi.Router) {
		r.Use(chiMiddleware.RequestID)
		r.Use(midleware.Logger)
		r.Use(midleware.Recovery)
		r.Use(midleware.ErrorHandler)
		r.Use(midleware.AuthMiddleware(authUseCase))

		r.Post(pathRoot, calcHandler.CalculateExpression)
		r.Get(pathRoot, calcHandler.ListCalculations)
		r.Get(pathByID, calcHandler.GetCalculation)
		r.Get(pathHealth, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(calcHealthMsg)); err != nil {
				logger.ContextLogger(r.Context(), nil).Error("Failed to write health check response", zap.Error(err))
			}
		})
	})
}
