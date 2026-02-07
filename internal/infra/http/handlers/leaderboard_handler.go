package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/melisource/tourney-rank/internal/domain/player"
	"github.com/melisource/tourney-rank/internal/usecase/leaderboard"
)

// LeaderboardHandler handles HTTP requests for leaderboard resources.
type LeaderboardHandler struct {
	service *leaderboard.Service
	logger  *slog.Logger
}

// NewLeaderboardHandler creates a new LeaderboardHandler.
func NewLeaderboardHandler(service *leaderboard.Service, logger *slog.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{
		service: service,
		logger:  logger,
	}
}

// GetLeaderboard handles GET /api/v1/leaderboard/{gameId}
func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract game ID from path
	gameIDStr := r.PathValue("gameId")
	if gameIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	// Try to parse as UUID first, if not try slug lookup via service
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid game id format")
		return
	}

	// Parse pagination params
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	// Clamp limit
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Get leaderboard
	entries, gameName, total, err := h.service.GetLeaderboard(ctx, gameID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get leaderboard", "game_id", gameID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get leaderboard")
		return
	}

	response := map[string]interface{}{
		"game_id":   gameID.String(),
		"game_name": gameName,
		"entries":   entries,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// GetLeaderboardByTier handles GET /api/v1/leaderboard/{gameId}/tier/{tier}
func (h *LeaderboardHandler) GetLeaderboardByTier(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gameIDStr := r.PathValue("gameId")
	tierStr := r.PathValue("tier")

	if gameIDStr == "" || tierStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id and tier are required")
		return
	}

	// Parse game ID
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid game id format")
		return
	}

	// Parse pagination
	limit := parseIntParam(r, "limit", 50)

	// Get leaderboard by tier
	entries, err := h.service.GetLeaderboardByTier(ctx, gameID, tierStr, limit)
	if err != nil {
		h.logger.Error("failed to get leaderboard by tier", "game_id", gameID, "tier", tierStr, "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]interface{}{
		"game_id": gameID.String(),
		"tier":    tierStr,
		"entries": entries,
		"limit":   limit,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// GetPlayerRank handles GET /api/v1/leaderboard/{gameId}/player/{playerId}
func (h *LeaderboardHandler) GetPlayerRank(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gameIDStr := r.PathValue("gameId")
	playerIDStr := r.PathValue("playerId")

	if gameIDStr == "" || playerIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id and player id are required")
		return
	}

	// Parse IDs
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid game id format")
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid player id format")
		return
	}

	// Get player rank
	rankResp, err := h.service.GetPlayerRank(ctx, playerID, gameID)
	if err != nil {
		h.logger.Error("failed to get player rank", "game_id", gameID, "player_id", playerID, "error", err)
		if errors.Is(err, player.ErrStatsNotFound) {
			h.errorResponse(w, http.StatusNotFound, "player has no stats for this game")
		} else {
			h.errorResponse(w, http.StatusInternalServerError, "failed to get player rank")
		}
		return
	}

	h.jsonResponse(w, http.StatusOK, rankResp)
}

// GetTierDistribution handles GET /api/v1/leaderboard/{gameId}/tiers
func (h *LeaderboardHandler) GetTierDistribution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gameIDStr := r.PathValue("gameId")
	if gameIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	// Parse game ID
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid game id format")
		return
	}

	// Get tier distribution
	distribution, total, err := h.service.GetTierDistribution(ctx, gameID)
	if err != nil {
		h.logger.Error("failed to get tier distribution", "game_id", gameID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get tier distribution")
		return
	}

	response := map[string]interface{}{
		"game_id":       gameID.String(),
		"distribution":  distribution,
		"total_players": total,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// jsonResponse writes a JSON response.
func (h *LeaderboardHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// errorResponse writes an error response.
func (h *LeaderboardHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}

// parseIntParam parses an integer query parameter with a default value.
func parseIntParam(r *http.Request, name string, defaultVal int64) int64 {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}

	return parsed
}
