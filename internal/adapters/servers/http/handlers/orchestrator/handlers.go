package orchestrator

import (
	"encoding/json"
	"net/http"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/midleware"
	orchAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const contentTypeJSON = "application/json"

type Handler struct {
	calcUseCase orchAPI.UseCaseCalculation
}

func NewHandler(calcUseCase orchAPI.UseCaseCalculation) *Handler {
	return &Handler{calcUseCase: calcUseCase}
}

type CalculateRequest struct {
	Expression string `json:"expression"`
}

func (h *Handler) CalculateExpression(w http.ResponseWriter, r *http.Request) {
	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		midleware.HandleError(r.Context(), w, err, http.StatusBadRequest)
		return
	}

	userID, err := midleware.GetUserIDFromContext(r.Context())
	if err != nil {
		midleware.HandleError(r.Context(), w, err, http.StatusUnauthorized)
		return
	}

	calculation, err := h.calcUseCase.CalculateExpression(r.Context(), userID, req.Expression)
	if err != nil {
		logger.ContextLogger(r.Context(), nil).Error("failed to create calculation", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusInternalServerError)
		return
	}

	respondJSON(w, calculation, http.StatusAccepted, logger.ContextLogger(r.Context(), nil))
}

func (h *Handler) GetCalculation(w http.ResponseWriter, r *http.Request) {
	calculationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		midleware.HandleError(r.Context(), w, err, http.StatusBadRequest)
		return
	}

	userID, err := midleware.GetUserIDFromContext(r.Context())
	if err != nil {
		midleware.HandleError(r.Context(), w, err, http.StatusUnauthorized)
		return
	}

	calculation, err := h.calcUseCase.GetCalculation(r.Context(), calculationID, userID)
	if err != nil {
		logger.ContextLogger(r.Context(), nil).Error("failed to get calculation",
			zap.String("calculation_id", calculationID.String()),
			zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusNotFound)
		return
	}

	respondJSON(w, calculation, http.StatusOK, logger.ContextLogger(r.Context(), nil))
}

func (h *Handler) ListCalculations(w http.ResponseWriter, r *http.Request) {
	userID, err := midleware.GetUserIDFromContext(r.Context())
	if err != nil {
		midleware.HandleError(r.Context(), w, err, http.StatusUnauthorized)
		return
	}

	calculations, err := h.calcUseCase.ListCalculations(r.Context(), userID)
	if err != nil {
		logger.ContextLogger(r.Context(), nil).Error("failed to list calculations", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusInternalServerError)
		return
	}

	respondJSON(w, calculations, http.StatusOK, logger.ContextLogger(r.Context(), nil))
}

func respondJSON(w http.ResponseWriter, data any, statusCode int, log logger.Logger) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// We've already written headers, so we can't change the status code
		log.Error("failed to encode response to JSON",
			zap.Error(err),
			zap.Int("status_code", statusCode))
	}
}
