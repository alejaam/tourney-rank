// Package handlers provides HTTP handlers for the API endpoints.
package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/alejaam/tourney-rank/internal/domain/game"
	"github.com/alejaam/tourney-rank/internal/infra/mongodb"
)

// GameHandler handles HTTP requests for game resources.
type GameHandler struct {
	repo   *mongodb.GameRepository
	logger *slog.Logger
}

// NewGameHandler creates a new GameHandler.
func NewGameHandler(repo *mongodb.GameRepository, logger *slog.Logger) *GameHandler {
	return &GameHandler{
		repo:   repo,
		logger: logger,
	}
}

// CreateGameRequest represents the request body for creating a game.
type CreateGameRequest struct {
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Description      string                 `json:"description"`
	StatSchema       map[string]interface{} `json:"stat_schema"`
	RankingWeights   map[string]float64     `json:"ranking_weights"`
	PlatformIDFormat string                 `json:"platform_id_format"`
}

// GameResponse represents a game in API responses.
type GameResponse struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Description      string                 `json:"description"`
	StatSchema       map[string]interface{} `json:"stat_schema"`
	RankingWeights   map[string]float64     `json:"ranking_weights"`
	PlatformIDFormat string                 `json:"platform_id_format"`
	IsActive         bool                   `json:"is_active"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

// ListGamesResponse represents the response for listing games.
type ListGamesResponse struct {
	Games []GameResponse `json:"games"`
	Total int            `json:"total"`
}

// List handles GET /api/v1/games
func (h *GameHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query params
	activeOnly := r.URL.Query().Get("active") == "true"

	games, err := h.repo.List(ctx, activeOnly)
	if err != nil {
		h.logger.Error("failed to list games", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to list games")
		return
	}

	response := ListGamesResponse{
		Games: make([]GameResponse, 0, len(games)),
		Total: len(games),
	}

	for _, g := range games {
		response.Games = append(response.Games, toGameResponse(g))
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// GetByID handles GET /api/v1/games/{id}
func (h *GameHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	// Try to parse as UUID first
	id, err := uuid.Parse(idStr)
	if err != nil {
		// If not UUID, try to find by slug
		g, err := h.repo.GetBySlug(ctx, idStr)
		if err != nil {
			if errors.Is(err, game.ErrNotFound) {
				h.errorResponse(w, http.StatusNotFound, "game not found")
				return
			}
			h.logger.Error("failed to get game by slug", "slug", idStr, "error", err)
			h.errorResponse(w, http.StatusInternalServerError, "failed to get game")
			return
		}
		h.jsonResponse(w, http.StatusOK, toGameResponse(g))
		return
	}

	g, err := h.repo.GetByID(ctx, id.String())
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "game not found")
			return
		}
		h.logger.Error("failed to get game", "id", id, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get game")
		return
	}

	h.jsonResponse(w, http.StatusOK, toGameResponse(g))
}

// Create handles POST /api/v1/games
func (h *GameHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if err := h.validateCreateRequest(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Convert stat schema to domain type
	statSchema := make(game.StatSchema)
	for key, val := range req.StatSchema {
		if field, ok := val.(map[string]interface{}); ok {
			statSchema[key] = game.StatField{
				Type:  getString(field, "type"),
				Min:   field["min"],
				Max:   field["max"],
				Label: getString(field, "label"),
			}
		}
	}

	// Create domain entity
	g, err := game.NewGame(
		req.Name,
		req.Slug,
		req.Description,
		req.PlatformIDFormat,
		statSchema,
		req.RankingWeights,
	)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Save to database
	if err := h.repo.Create(ctx, g); err != nil {
		if errors.Is(err, mongodb.ErrGameAlreadyExists) {
			h.errorResponse(w, http.StatusConflict, "game with this slug already exists")
			return
		}
		h.logger.Error("failed to create game", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to create game")
		return
	}

	h.logger.Info("game created", "id", g.ID, "slug", g.Slug)
	h.jsonResponse(w, http.StatusCreated, toGameResponse(g))
}

// UpdateStatus handles PATCH /api/v1/games/{id}/status
func (h *GameHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid game id")
		return
	}

	var req struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.repo.SetActive(ctx, id, req.Active); err != nil {
		if errors.Is(err, game.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "game not found")
			return
		}
		h.logger.Error("failed to update game status", "id", id, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to update game status")
		return
	}

	h.logger.Info("game status updated", "id", id, "active", req.Active)
	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"id":     id.String(),
		"active": req.Active,
	})
}

// Delete handles DELETE /api/v1/games/{id}
func (h *GameHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid game id")
		return
	}

	if err := h.repo.Delete(ctx, id.String()); err != nil {
		if errors.Is(err, game.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "game not found")
			return
		}
		h.logger.Error("failed to delete game", "id", id, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to delete game")
		return
	}

	h.logger.Info("game deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

// validateCreateRequest validates the create game request.
func (h *GameHandler) validateCreateRequest(req *CreateGameRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.Slug) == "" {
		return errors.New("slug is required")
	}
	if len(req.RankingWeights) == 0 {
		return errors.New("ranking_weights is required")
	}
	return nil
}

// jsonResponse writes a JSON response.
func (h *GameHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// errorResponse writes an error response.
func (h *GameHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}

// toGameResponse converts a domain game to an API response.
func toGameResponse(g *game.Game) GameResponse {
	statSchema := make(map[string]interface{})
	for key, field := range g.StatSchema {
		statSchema[key] = map[string]interface{}{
			"type":  field.Type,
			"min":   field.Min,
			"max":   field.Max,
			"label": field.Label,
		}
	}

	return GameResponse{
		ID:               g.ID.String(),
		Name:             g.Name,
		Slug:             g.Slug,
		Description:      g.Description,
		StatSchema:       statSchema,
		RankingWeights:   g.RankingWeights,
		PlatformIDFormat: g.PlatformIDFormat,
		IsActive:         g.IsActive,
		CreatedAt:        g.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        g.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// getString safely extracts a string from a map.
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
