package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	playerdomain "github.com/melisource/tourney-rank/internal/domain/player"
	"github.com/melisource/tourney-rank/internal/infra/http/middleware"
	"github.com/melisource/tourney-rank/internal/infra/mongodb"
	playerusecase "github.com/melisource/tourney-rank/internal/usecase/player"
)

// PlayerHandler handles HTTP requests for player operations.
type PlayerHandler struct {
	service   *playerusecase.Service
	statsRepo *mongodb.PlayerStatsRepository
	gameRepo  *mongodb.GameRepository
	logger    *slog.Logger
}

// NewPlayerHandler creates a new PlayerHandler.
func NewPlayerHandler(
	service *playerusecase.Service,
	statsRepo *mongodb.PlayerStatsRepository,
	gameRepo *mongodb.GameRepository,
	logger *slog.Logger,
) *PlayerHandler {
	return &PlayerHandler{
		service:   service,
		statsRepo: statsRepo,
		gameRepo:  gameRepo,
		logger:    logger,
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

		if errors.Is(err, playerdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "player profile not found")
			return
		}
		if errors.Is(err, playerdomain.ErrInvalidBirthYear) {
			h.errorResponse(w, http.StatusBadRequest, "invalid birth_year")
			return
		}
		if errors.Is(err, playerdomain.ErrInvalidPlatform) {
			h.errorResponse(w, http.StatusBadRequest, "invalid preferred_platform")
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

	var req playerusecase.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		h.errorResponse(w, http.StatusBadRequest, "display_name is required")
		return
	}

	if req.PreferredPlatform == "" {
		h.errorResponse(w, http.StatusBadRequest, "preferred_platform is required")
		return
	}

	player, err := h.service.CreateProfile(r.Context(), userID, req)
	if err != nil {
		h.logger.Error("failed to create player profile", "user_id", userID, "error", err)

		if errors.Is(err, playerdomain.ErrInvalidUsername) {
			h.errorResponse(w, http.StatusBadRequest, "invalid display_name")
			return
		}
		if errors.Is(err, playerdomain.ErrInvalidPlatform) {
			h.errorResponse(w, http.StatusBadRequest, "invalid preferred_platform")
			return
		}
		if errors.Is(err, playerdomain.ErrInvalidBirthYear) {
			h.errorResponse(w, http.StatusBadRequest, "invalid birth_year")
			return
		}
		if err.Error() == "player profile already exists" {
			h.errorResponse(w, http.StatusConflict, "player profile already exists")
			return
		}

		h.errorResponse(w, http.StatusInternalServerError, "failed to create player profile")
		return
	}

	h.jsonResponse(w, http.StatusCreated, player)
}

// GetMyStats returns the player profile with game stats summary.
// GET /api/v1/players/me/stats
func (h *PlayerHandler) GetMyStats(w http.ResponseWriter, r *http.Request) {
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

	// Get player profile
	player, err := h.service.GetMyProfile(r.Context(), userID)
	if err != nil {
		h.logger.Debug("player not found, attempting auto-create", "user_id", userID)
		player, err = h.service.GetOrCreateByUserID(r.Context(), userID, "Player")
		if err != nil {
			h.logger.Error("failed to get or create player", "user_id", userID, "error", err)
			h.errorResponse(w, http.StatusInternalServerError, "failed to get player profile")
			return
		}
	}

	// Get all player stats
	allStats, err := h.statsRepo.GetByPlayer(r.Context(), player.ID)
	if err != nil {
		h.logger.Error("failed to get player stats", "user_id", userID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get player stats")
		return
	}

	// Enrich stats with game information
	type GameStats struct {
		GameID        string                 `json:"game_id"`
		GameName      string                 `json:"game_name"`
		Stats         map[string]interface{} `json:"stats"`
		RankingScore  float64                `json:"ranking_score"`
		Tier          string                 `json:"tier"`
		MatchesPlayed int                    `json:"matches_played"`
		LastMatchAt   *string                `json:"last_match_at"`
	}

	games := make([]GameStats, 0, len(allStats))
	for _, ps := range allStats {
		gameID := ps.GameID.String()
		gameName := gameID // fallback

		// Try to get game name from repo
		if game, err := h.gameRepo.GetByID(r.Context(), ps.GameID.String()); err == nil {
			gameName = game.Name
		}

		games = append(games, GameStats{
			GameID:        gameID,
			GameName:      gameName,
			Stats:         ps.Stats,
			RankingScore:  ps.RankingScore,
			Tier:          string(ps.Tier),
			MatchesPlayed: ps.MatchesPlayed,
			LastMatchAt:   lastMatchAtString(ps.LastMatchAt),
		})
	}

	response := map[string]interface{}{
		"player": player,
		"games":  games,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// GetMyGameStats returns detailed stats for the player in a specific game.
// GET /api/v1/players/me/stats/{gameId}
func (h *PlayerHandler) GetMyGameStats(w http.ResponseWriter, r *http.Request) {
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

	gameIDStr := r.PathValue("gameId")
	if gameIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "game id is required")
		return
	}

	// Get player
	player, err := h.service.GetMyProfile(r.Context(), userID)
	if err != nil {
		h.logger.Debug("player not found, attempting auto-create", "user_id", userID)
		player, err = h.service.GetOrCreateByUserID(r.Context(), userID, "Player")
		if err != nil {
			h.logger.Error("failed to get or create player", "user_id", userID, "error", err)
			h.errorResponse(w, http.StatusInternalServerError, "failed to get player profile")
			return
		}
	}

	// Parse game ID
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid game id format")
		return
	}

	// Get player stats for this game
	ps, err := h.statsRepo.GetByPlayerAndGame(r.Context(), player.ID, gameID)
	if err != nil {
		if errors.Is(err, playerdomain.ErrStatsNotFound) {
			h.errorResponse(w, http.StatusNotFound, "player has no stats for this game")
			return
		}
		h.logger.Error("failed to get player stats", "player_id", player.ID, "game_id", gameID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get player stats")
		return
	}

	// Get game name
	gameName := gameIDStr
	if game, err := h.gameRepo.GetByID(r.Context(), gameID.String()); err == nil {
		gameName = game.Name
	}

	// Get player rank and total count for percentile calculation
	rankInfo, err := h.statsRepo.GetPlayerRank(r.Context(), player.ID, gameID)
	if err != nil {
		h.logger.Error("failed to get player rank", "player_id", player.ID, "game_id", gameID, "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get player rank")
		return
	}

	totalCount, err := h.statsRepo.CountByGame(r.Context(), gameID)
	if err != nil {
		totalCount = 1 // fallback
	}

	// Calculate percentile
	var percentile float64
	if totalCount > 0 {
		percentile = float64(totalCount-rankInfo.Rank) / float64(totalCount)
		if percentile < 0 {
			percentile = 0
		}
	}

	response := map[string]interface{}{
		"id":             ps.ID.String(),
		"player_id":      ps.PlayerID.String(),
		"game_id":        ps.GameID.String(),
		"game_name":      gameName,
		"stats":          ps.Stats,
		"ranking_score":  ps.RankingScore,
		"tier":           string(ps.Tier),
		"matches_played": ps.MatchesPlayed,
		"last_match_at":  lastMatchAtString(ps.LastMatchAt),
		"rank":           rankInfo.Rank,
		"percentile":     percentile,
		"created_at":     ps.CreatedAt,
		"updated_at":     ps.UpdatedAt,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// lastMatchAtString converts a pointer to time to ISO string or nil
func lastMatchAtString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
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
