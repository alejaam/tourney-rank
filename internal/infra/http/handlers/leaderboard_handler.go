// Package handlers provides HTTP handlers for the API endpoints.
package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/melisource/tourney-rank/internal/domain/player"
	"github.com/melisource/tourney-rank/internal/infra/mongodb"
)

// LeaderboardHandler handles HTTP requests for leaderboard resources.
type LeaderboardHandler struct {
	statsRepo *mongodb.PlayerStatsRepository
	gameRepo  *mongodb.GameRepository
	logger    *slog.Logger
}

// NewLeaderboardHandler creates a new LeaderboardHandler.
func NewLeaderboardHandler(
	statsRepo *mongodb.PlayerStatsRepository,
	gameRepo *mongodb.GameRepository,
	logger *slog.Logger,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		statsRepo: statsRepo,
		gameRepo:  gameRepo,
		logger:    logger,
	}
}

// LeaderboardResponse represents the leaderboard API response.
type LeaderboardResponse struct {
	GameID   string                     `json:"game_id"`
	GameName string                     `json:"game_name"`
	Entries  []mongodb.LeaderboardEntry `json:"entries"`
	Total    int64                      `json:"total"`
	Limit    int64                      `json:"limit"`
	Offset   int64                      `json:"offset"`
}

// TierDistributionResponse represents the tier distribution response.
type TierDistributionResponse struct {
	GameID       string           `json:"game_id"`
	Distribution map[string]int64 `json:"distribution"`
	TotalPlayers int64            `json:"total_players"`
}

// PlayerRankResponse represents the player rank response.
type PlayerRankResponse struct {
	PlayerID     string  `json:"player_id"`
	GameID       string  `json:"game_id"`
	Rank         int64   `json:"rank"`
	RankingScore float64 `json:"ranking_score"`
	Tier         string  `json:"tier"`
	Percentile   float64 `json:"percentile"`
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

	// Parse game ID or get by slug
	var gameID uuid.UUID
	var gameName string

	id, err := uuid.Parse(gameIDStr)
	if err != nil {
		// Try to find by slug
		g, err := h.gameRepo.GetBySlug(ctx, gameIDStr)
		if err != nil {
			if errors.Is(err, mongodb.ErrGameNotFound) {
				h.errorResponse(w, http.StatusNotFound, "game not found")
				return
			}
			h.logger.Error("failed to get game by slug", "slug", gameIDStr, "error", err)
			h.errorResponse(w, http.StatusInternalServerError, "failed to get game")
			return
		}
		gameID = g.ID
		gameName = g.Name
	} else {
		// Get game by ID for name
		g, err := h.gameRepo.GetByID(ctx, id.String())
		if err != nil {
			if errors.Is(err, mongodb.ErrGameNotFound) {
				h.errorResponse(w, http.StatusNotFound, "game not found")
				return
			}
			h.logger.Error("failed to get game", "id", id, "error", err)
			h.errorResponse(w, http.StatusInternalServerError, "failed to get game")
			return
		}
		gameID = id
		gameName = g.Name
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
	entries, err := h.statsRepo.GetLeaderboard(ctx, gameID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get leaderboard", "game_id", gameID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get leaderboard")
		return
	}

	// Get total count
	total, err := h.statsRepo.CountByGame(ctx, gameID)
	if err != nil {
		h.logger.Error("failed to count players", "game_id", gameID, "error", err)
		// Continue with 0 total
		total = 0
	}

	response := LeaderboardResponse{
		GameID:   gameID.String(),
		GameName: gameName,
		Entries:  entries,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
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

	// Validate tier
	validTiers := map[string]bool{
		"bronze": true, "silver": true, "gold": true,
		"platinum": true, "diamond": true, "master": true,
	}
	if !validTiers[tierStr] {
		h.errorResponse(w, http.StatusBadRequest, "invalid tier")
		return
	}

	// Get game ID
	var gameID uuid.UUID
	id, err := uuid.Parse(gameIDStr)
	if err != nil {
		g, err := h.gameRepo.GetBySlug(ctx, gameIDStr)
		if err != nil {
			if errors.Is(err, mongodb.ErrGameNotFound) {
				h.errorResponse(w, http.StatusNotFound, "game not found")
				return
			}
			h.errorResponse(w, http.StatusInternalServerError, "failed to get game")
			return
		}
		gameID = g.ID
	} else {
		gameID = id
	}

	// Parse pagination
	limit := parseIntParam(r, "limit", 50)

	// Get leaderboard by tier
	tier := player.Tier(tierStr)
	entries, err := h.statsRepo.GetLeaderboardByTier(ctx, gameID, tier, limit)
	if err != nil {
		h.logger.Error("failed to get leaderboard by tier", "game_id", gameID, "tier", tierStr, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get leaderboard")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"game_id": gameID.String(),
		"tier":    tierStr,
		"entries": entries,
		"limit":   limit,
	})
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

	// Get game ID
	var gameID uuid.UUID
	id, err := uuid.Parse(gameIDStr)
	if err != nil {
		g, err := h.gameRepo.GetBySlug(ctx, gameIDStr)
		if err != nil {
			if errors.Is(err, mongodb.ErrGameNotFound) {
				h.errorResponse(w, http.StatusNotFound, "game not found")
				return
			}
			h.errorResponse(w, http.StatusInternalServerError, "failed to get game")
			return
		}
		gameID = g.ID
	} else {
		gameID = id
	}

	// Parse player ID
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid player id")
		return
	}

	// Get player rank
	rank, err := h.statsRepo.GetPlayerRank(ctx, gameID, playerID)
	if err != nil {
		if errors.Is(err, mongodb.ErrPlayerStatsNotFound) {
			h.errorResponse(w, http.StatusNotFound, "player stats not found")
			return
		}
		h.logger.Error("failed to get player rank", "game_id", gameID, "player_id", playerID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get player rank")
		return
	}

	// Get total players to calculate percentile
	total, _ := h.statsRepo.CountByGame(ctx, gameID)
	percentile := 0.0
	if total > 0 {
		percentile = float64(total-rank.Rank+1) / float64(total) * 100
	}

	response := PlayerRankResponse{
		PlayerID:     playerID.String(),
		GameID:       gameID.String(),
		Rank:         rank.Rank,
		RankingScore: rank.RankingScore,
		Tier:         string(rank.Tier),
		Percentile:   percentile,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// GetTierDistribution handles GET /api/v1/leaderboard/{gameId}/tiers
func (h *LeaderboardHandler) GetTierDistribution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gameIDStr := r.PathValue("gameId")
	if gameIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	// Get game ID
	var gameID uuid.UUID
	id, err := uuid.Parse(gameIDStr)
	if err != nil {
		g, err := h.gameRepo.GetBySlug(ctx, gameIDStr)
		if err != nil {
			if errors.Is(err, mongodb.ErrGameNotFound) {
				h.errorResponse(w, http.StatusNotFound, "game not found")
				return
			}
			h.errorResponse(w, http.StatusInternalServerError, "failed to get game")
			return
		}
		gameID = g.ID
	} else {
		gameID = id
	}

	// Get tier distribution
	distribution, err := h.statsRepo.GetTierDistribution(ctx, gameID)
	if err != nil {
		h.logger.Error("failed to get tier distribution", "game_id", gameID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get tier distribution")
		return
	}

	// Convert to string keys and calculate total
	stringDist := make(map[string]int64)
	var total int64
	for tier, count := range distribution {
		stringDist[string(tier)] = count
		total += count
	}

	response := TierDistributionResponse{
		GameID:       gameID.String(),
		Distribution: stringDist,
		TotalPlayers: total,
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
