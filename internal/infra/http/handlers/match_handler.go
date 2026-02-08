// Package handlers provides HTTP handlers for the API endpoints.
package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/alejaam/tourney-rank/internal/domain/match"
	"github.com/alejaam/tourney-rank/internal/infra/http/middleware"
	usecasematch "github.com/alejaam/tourney-rank/internal/usecase/match"
)

// MatchHandler handles HTTP requests for match resources.
type MatchHandler struct {
	logger  *slog.Logger
	service *usecasematch.Service
}

// NewMatchHandler creates a new MatchHandler.
func NewMatchHandler(logger *slog.Logger, service *usecasematch.Service) *MatchHandler {
	return &MatchHandler{
		logger:  logger,
		service: service,
	}
}

// HandleSubmitMatch handles POST /api/v1/matches
// Requires authentication. Team captain submits match report.
func (h *MatchHandler) HandleSubmitMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userInfo, ok := middleware.GetUserInfo(ctx)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	captainID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req usecasematch.SubmitMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.SubmitMatch(ctx, req, captainID)
	if err != nil {
		h.handleMatchError(w, err)
		return
	}

	h.logger.Info("match submitted", "id", resp.ID, "tournament_id", resp.TournamentID)
	h.jsonResponse(w, http.StatusCreated, resp)
}

// HandleGetTournamentMatches handles GET /api/v1/tournaments/{tournament_id}/matches
// Public endpoint. Returns verified matches for a tournament.
func (h *MatchHandler) HandleGetTournamentMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tournamentIDStr := r.PathValue("tournament_id")
	if tournamentIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "tournament_id is required")
		return
	}

	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid tournament_id")
		return
	}

	limit := h.parseIntQueryParam(r, "limit", 20)
	offset := h.parseIntQueryParam(r, "offset", 0)

	resp, err := h.service.GetTournamentMatches(ctx, tournamentID, usecasematch.MatchHistoryRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("failed to get tournament matches", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get matches")
		return
	}

	h.jsonResponse(w, http.StatusOK, resp)
}

// HandleGetPlayerMatches handles GET /api/v1/players/me/matches
// Requires authentication. Returns match history for the authenticated player.
func (h *MatchHandler) HandleGetPlayerMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userInfo, ok := middleware.GetUserInfo(ctx)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	playerID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	limit := h.parseIntQueryParam(r, "limit", 10)
	offset := h.parseIntQueryParam(r, "offset", 0)

	resp, err := h.service.GetMatchHistory(ctx, playerID, usecasematch.MatchHistoryRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("failed to get player matches", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get match history")
		return
	}

	h.jsonResponse(w, http.StatusOK, resp)
}

// HandleGetMatch handles GET /api/v1/matches/{id}
// Public endpoint. Returns a single match by ID.
func (h *MatchHandler) HandleGetMatch(w http.ResponseWriter, r *http.Request) {
	// This would be implemented if you need direct match lookup
	// For now, this is a placeholder
	h.errorResponse(w, http.StatusNotImplemented, "not implemented")
}

// HandleGetUnverifiedMatches handles GET /api/v1/admin/matches/unverified
// Requires admin authentication. Returns unverified matches for review.
func (h *MatchHandler) HandleGetUnverifiedMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := h.parseIntQueryParam(r, "limit", 20)
	offset := h.parseIntQueryParam(r, "offset", 0)

	resp, err := h.service.GetUnverifiedMatches(ctx, usecasematch.MatchHistoryRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("failed to get unverified matches", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get unverified matches")
		return
	}

	h.jsonResponse(w, http.StatusOK, resp)
}

// HandleVerifyMatch handles PATCH /api/v1/admin/matches/{id}/verify
// Requires admin authentication. Admin approves or rejects a match.
func (h *MatchHandler) HandleVerifyMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userInfo, ok := middleware.GetUserInfo(ctx)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	adminID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	matchIDStr := r.PathValue("id")
	if matchIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "match id is required")
		return
	}

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid match id")
		return
	}

	var req usecasematch.VerifyMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.AdminVerifyMatch(ctx, matchID, req, adminID)
	if err != nil {
		h.handleMatchError(w, err)
		return
	}

	status := "verified"
	if !req.Approved {
		status = "rejected"
	}

	h.logger.Info("match status updated", "id", resp.ID, "status", status)
	h.jsonResponse(w, http.StatusOK, resp)
}

// Helper functions

// jsonResponse marshals data to JSON and writes the response.
func (h *MatchHandler) jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// errorResponse writes an error response with the given status code and message.
func (h *MatchHandler) errorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode error response", "error", err)
	}
}

// parseIntQueryParam parses an integer query parameter with a default value.
func (h *MatchHandler) parseIntQueryParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// handleMatchError converts domain errors to appropriate HTTP status codes.
func (h *MatchHandler) handleMatchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, match.ErrNotFound):
		h.errorResponse(w, http.StatusNotFound, "match not found")

	case errors.Is(err, match.ErrTournamentNotActive):
		h.errorResponse(w, http.StatusBadRequest, "tournament is not active")

	case errors.Is(err, match.ErrNotCaptain):
		h.errorResponse(w, http.StatusForbidden, "only team captain can submit matches")

	case errors.Is(err, match.ErrPlayerNotInTeam):
		h.errorResponse(w, http.StatusBadRequest, "player is not in the team")

	case errors.Is(err, match.ErrTeamSizeMismatch):
		h.errorResponse(w, http.StatusBadRequest, "player stats count does not match team size")

	case errors.Is(err, match.ErrInvalidPlacement):
		h.errorResponse(w, http.StatusBadRequest, "placement must be between 1 and 100")

	case errors.Is(err, match.ErrInvalidKills):
		h.errorResponse(w, http.StatusBadRequest, "kills cannot be negative")

	case errors.Is(err, match.ErrInvalidPlayerStats):
		h.errorResponse(w, http.StatusBadRequest, "invalid player statistics")

	case errors.Is(err, match.ErrMatchNotDraft):
		h.errorResponse(w, http.StatusBadRequest, "only draft matches can be verified")

	default:
		h.logger.Error("failed to process match", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "internal server error")
	}
}
