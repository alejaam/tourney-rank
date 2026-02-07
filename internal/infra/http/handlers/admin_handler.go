package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/melisource/tourney-rank/internal/domain/user"
	"github.com/melisource/tourney-rank/internal/usecase/admin"
)

// AdminHandler handles HTTP requests for admin operations.
type AdminHandler struct {
	userService   *admin.UserService
	gameService   *admin.GameService
	playerService *admin.PlayerService
	logger        *slog.Logger
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(
	userService *admin.UserService,
	gameService *admin.GameService,
	playerService *admin.PlayerService,
	logger *slog.Logger,
) *AdminHandler {
	return &AdminHandler{
		userService:   userService,
		gameService:   gameService,
		playerService: playerService,
		logger:        logger,
	}
}

// ============= USER MANAGEMENT =============

// ListUsers handles GET /api/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	res, err := h.userService.ListUsers(r.Context())
	if err != nil {
		h.logger.Error("failed to list users", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	h.jsonResponse(w, http.StatusOK, res)
}

// GetUser handles GET /api/admin/users/:id
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "user id is required")
		return
	}

	u, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("failed to get user", "id", id, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	h.jsonResponse(w, http.StatusOK, u)
}

// DeleteUser handles DELETE /api/admin/users/:id
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "user id is required")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("failed to delete user", "id", id, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateUserRole handles PATCH /api/admin/users/:id/role
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "user id is required")
		return
	}

	var req admin.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.userService.UpdateRole(r.Context(), id, req); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("failed to update user role", "id", id, "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============= GAME MANAGEMENT =============

// ListGames handles GET /api/admin/games
func (h *AdminHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	res, err := h.gameService.ListGames(r.Context())
	if err != nil {
		h.logger.Error("failed to list games", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to list games")
		return
	}

	h.jsonResponse(w, http.StatusOK, res)
}

// GetGame handles GET /api/admin/games/:id
func (h *AdminHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	g, err := h.gameService.GetGame(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get game", "id", id, "error", err)
		h.errorResponse(w, http.StatusNotFound, "game not found")
		return
	}

	h.jsonResponse(w, http.StatusOK, g)
}

// CreateGame handles POST /api/admin/games
func (h *AdminHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req admin.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	g, err := h.gameService.CreateGame(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create game", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusCreated, g)
}

// UpdateGame handles PUT /api/admin/games/:id
func (h *AdminHandler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	var req admin.UpdateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	g, err := h.gameService.UpdateGame(r.Context(), id, req)
	if err != nil {
		h.logger.Error("failed to update game", "id", id, "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, g)
}

// DeleteGame handles DELETE /api/admin/games/:id
func (h *AdminHandler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	if err := h.gameService.DeleteGame(r.Context(), id); err != nil {
		h.logger.Error("failed to delete game", "id", id, "error", err)
		h.errorResponse(w, http.StatusNotFound, "game not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============= PLAYER MANAGEMENT =============

// ListPlayers handles GET /api/admin/players
func (h *AdminHandler) ListPlayers(w http.ResponseWriter, r *http.Request) {
	res, err := h.playerService.ListPlayers(r.Context())
	if err != nil {
		h.logger.Error("failed to list players", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to list players")
		return
	}

	h.jsonResponse(w, http.StatusOK, res)
}

// GetPlayer handles GET /api/admin/players/:id
func (h *AdminHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "player id is required")
		return
	}

	p, err := h.playerService.GetPlayer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get player", "id", id, "error", err)
		h.errorResponse(w, http.StatusNotFound, "player not found")
		return
	}

	h.jsonResponse(w, http.StatusOK, p)
}

// CreatePlayer handles POST /api/admin/players
func (h *AdminHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req admin.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.playerService.CreatePlayer(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create player", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusCreated, p)
}

// BanPlayer handles PATCH /api/v1/admin/players/:id/ban
func (h *AdminHandler) BanPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "player id is required")
		return
	}

	p, err := h.playerService.BanPlayer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to ban player", "id", id, "error", err)
		h.errorResponse(w, http.StatusNotFound, "player not found")
		return
	}

	h.jsonResponse(w, http.StatusOK, p)
}

// UnbanPlayer handles PATCH /api/v1/admin/players/:id/unban
func (h *AdminHandler) UnbanPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "player id is required")
		return
	}

	p, err := h.playerService.UnbanPlayer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to unban player", "id", id, "error", err)
		h.errorResponse(w, http.StatusNotFound, "player not found")
		return
	}

	h.jsonResponse(w, http.StatusOK, p)
}

// UpdatePlayer handles PUT /api/admin/players/:id
func (h *AdminHandler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "player id is required")
		return
	}

	var req admin.UpdatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.playerService.UpdatePlayer(r.Context(), id, req)
	if err != nil {
		h.logger.Error("failed to update player", "id", id, "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, p)
}

// DeletePlayer handles DELETE /api/admin/players/:id
func (h *AdminHandler) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.errorResponse(w, http.StatusBadRequest, "player id is required")
		return
	}

	if err := h.playerService.DeletePlayer(r.Context(), id); err != nil {
		h.logger.Error("failed to delete player", "id", id, "error", err)
		h.errorResponse(w, http.StatusNotFound, "player not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============= HELPER METHODS =============

// jsonResponse writes a JSON response.
func (h *AdminHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", "error", err)
	}
}

// errorResponse writes an error response.
func (h *AdminHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
