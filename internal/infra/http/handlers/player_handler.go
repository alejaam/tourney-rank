package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/melisource/tourney-rank/internal/infra/http/middleware"
	playerusecase "github.com/melisource/tourney-rank/internal/usecase/player"
)

// PlayerHandler handles HTTP requests for player operations.
type PlayerHandler struct {
	service *playerusecase.Service
	logger  *slog.Logger
}

// NewPlayerHandler creates a new PlayerHandler.
func NewPlayerHandler(service *playerusecase.Service, logger *slog.Logger) *PlayerHandler {
	return &PlayerHandler{
		service: service,
		logger:  logger,
	}
}

// GetMyProfile returns the player profile for the authenticated user.
// GET /api/v1/players/me
func (h *PlayerHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.logger.Error("invalid user id", "error", err, "user_id", userInfo.ID)
		h.errorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	player, err := h.service.GetMyProfile(r.Context(), userID)
	if err != nil {
		h.logger.Debug("player not found, attempting auto-create", "user_id", userID)
		// Auto-create player if not found
		player, err = h.service.GetOrCreateByUserID(r.Context(), userID, "Player")
		if err != nil {
			h.logger.Error("failed to get or create player", "user_id", userID, "error", err)
			h.errorResponse(w, http.StatusInternalServerError, "failed to get player profile")
			return
		}
	}

	h.jsonResponse(w, http.StatusOK, player)
}

// UpdateMyProfile updates the player profile for the authenticated user.
// PUT /api/v1/players/me
func (h *PlayerHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.logger.Error("invalid user id", "error", err, "user_id", userInfo.ID)
		h.errorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req playerusecase.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	player, err := h.service.UpdateMyProfile(r.Context(), userID, req)
	if err != nil {
		h.logger.Error("failed to update player profile", "user_id", userID, "error", err)

		if errors.Is(err, errors.New("player not found")) {
			h.errorResponse(w, http.StatusNotFound, "player profile not found")
			return
		}

		h.errorResponse(w, http.StatusInternalServerError, "failed to update player profile")
		return
	}

	h.jsonResponse(w, http.StatusOK, player)
}

// CreateMyProfile creates a player profile for the authenticated user.
// POST /api/v1/players/me
func (h *PlayerHandler) CreateMyProfile(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.logger.Error("invalid user id", "error", err, "user_id", userInfo.ID)
		h.errorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		h.errorResponse(w, http.StatusBadRequest, "display_name is required")
		return
	}

	player, err := h.service.CreateProfile(r.Context(), userID, req.DisplayName)
	if err != nil {
		h.logger.Error("failed to create player profile", "user_id", userID, "error", err)

		if err.Error() == "player profile already exists" {
			h.errorResponse(w, http.StatusConflict, "player profile already exists")
			return
		}

		h.errorResponse(w, http.StatusInternalServerError, "failed to create player profile")
		return
	}

	h.jsonResponse(w, http.StatusCreated, player)
}

// jsonResponse writes a JSON response.
func (h *PlayerHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", "error", err)
	}
}

// errorResponse writes an error response.
func (h *PlayerHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
