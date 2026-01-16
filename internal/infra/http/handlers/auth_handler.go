package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/melisource/tourney-rank/internal/usecase/auth"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	service *auth.Service
	logger  *slog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(service *auth.Service, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to register user", "error", err)
		h.errorResponse(w, http.StatusConflict, err.Error()) // Assumption: error is duplicate logic
		return
	}

	h.jsonResponse(w, http.StatusCreated, res)
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			h.errorResponse(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.logger.Error("failed to login", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.jsonResponse(w, http.StatusOK, res)
}

// jsonResponse writes a JSON response.
func (h *AuthHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", "error", err)
	}
}

// errorResponse writes an error response.
func (h *AuthHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
