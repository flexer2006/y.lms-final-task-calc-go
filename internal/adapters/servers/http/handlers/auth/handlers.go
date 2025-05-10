package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/midleware"
	authAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	contentTypeJSON    = "application/json"
	authHeaderName     = "Authorization"
	bearerScheme       = "Bearer "
	tokenExpiryMinutes = 15
)

type Handler struct {
	authUseCase authAPI.UseCaseUser
	router      *chi.Mux
}

func NewHandler(authUseCase authAPI.UseCaseUser) *Handler {
	h := &Handler{
		authUseCase: authUseCase,
		router:      chi.NewRouter(),
	}

	h.router.Post("/register", h.Register)
	h.router.Post("/login", h.Login)
	h.router.Post("/refresh", h.RefreshToken)
	h.router.Post("/logout", h.Logout)

	return h
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	log := logger.ContextLogger(r.Context(), nil)

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode register request", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusBadRequest)
		return
	}

	userID, err := h.authUseCase.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Error("failed to register user", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusInternalServerError)
		return
	}

	tokens, err := h.authUseCase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Error("failed to login after registration",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusInternalServerError)
		return
	}

	respondJSON(w, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    time.Now().Add(tokenExpiryMinutes * time.Minute).Unix(),
	}, http.StatusCreated, log)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	log := logger.ContextLogger(r.Context(), nil)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode login request", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusBadRequest)
		return
	}

	tokens, err := h.authUseCase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Error("failed to login", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusUnauthorized)
		return
	}

	respondJSON(w, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    time.Now().Add(tokenExpiryMinutes * time.Minute).Unix(),
	}, http.StatusOK, log)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log := logger.ContextLogger(r.Context(), nil)

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode refresh token request", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusBadRequest)
		return
	}

	tokens, err := h.authUseCase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		log.Error("failed to refresh token", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusUnauthorized)
		return
	}

	respondJSON(w, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    time.Now().Add(tokenExpiryMinutes * time.Minute).Unix(),
	}, http.StatusOK, log)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	log := logger.ContextLogger(r.Context(), nil)

	authHeader := r.Header.Get(authHeaderName)
	if authHeader == "" {
		midleware.HandleError(r.Context(), w, midleware.ErrMissingToken, http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, bearerScheme) {
		midleware.HandleError(r.Context(), w, midleware.ErrInvalidAuthHeader, http.StatusUnauthorized)
		return
	}

	token := authHeader[len(bearerScheme):]
	if err := h.authUseCase.Logout(r.Context(), token); err != nil {
		log.Error("failed to logout", zap.Error(err))
		midleware.HandleError(r.Context(), w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Routes() *chi.Mux {
	return h.router
}

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int, log logger.Logger) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// We've already written headers, so we can't change the status code
		log.Error("failed to encode response to JSON",
			zap.Error(err),
			zap.Int("status_code", statusCode))
	}
}
