package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	tournamentdomain "github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/alejaam/tourney-rank/internal/infra/http/middleware"
	tournamentusecase "github.com/alejaam/tourney-rank/internal/usecase/tournament"
	"github.com/google/uuid"
)

// TournamentHandler handles HTTP requests for tournament operations.
type TournamentHandler struct {
	service *tournamentusecase.Service
	logger  *slog.Logger
}

// NewTournamentHandler creates a new tournament handler.
func NewTournamentHandler(service *tournamentusecase.Service, logger *slog.Logger) *TournamentHandler {
	return &TournamentHandler{
		service: service,
		logger:  logger,
	}
}

// CreateTournament handles POST /api/v1/tournaments
func (h *TournamentHandler) CreateTournament(w http.ResponseWriter, r *http.Request) {
	var req tournamentusecase.CreateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user ID from context (set by auth middleware)
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	tournament, err := h.service.CreateTournament(r.Context(), req, userID)
	if err != nil {
		h.logger.Error("Failed to create tournament", "error", err)
		status := http.StatusInternalServerError
		message := "Failed to create tournament"

		if errors.Is(err, tournamentdomain.ErrInvalidName) ||
			errors.Is(err, tournamentdomain.ErrInvalidTeamSize) ||
			errors.Is(err, tournamentdomain.ErrInvalidDates) {
			status = http.StatusBadRequest
			message = err.Error()
		}

		h.errorResponse(w, status, message)
		return
	}

	h.jsonResponse(w, http.StatusCreated, tournament)
}

// GetTournament handles GET /api/v1/tournaments/{id}
func (h *TournamentHandler) GetTournament(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	tournament, err := h.service.GetTournament(r.Context(), id)
	if err != nil {
		if errors.Is(err, tournamentdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Tournament not found")
			return
		}
		h.logger.Error("Failed to get tournament", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get tournament")
		return
	}

	h.jsonResponse(w, http.StatusOK, tournament)
}

// ListTournaments handles GET /api/v1/tournaments
func (h *TournamentHandler) ListTournaments(w http.ResponseWriter, r *http.Request) {
	var req tournamentusecase.ListTournamentsRequest

	// Parse query parameters
	if gameIDStr := r.URL.Query().Get("game_id"); gameIDStr != "" {
		gameID, err := uuid.Parse(gameIDStr)
		if err == nil {
			req.GameID = &gameID
		}
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := tournamentdomain.Status(statusStr)
		req.Status = &status
	}

	if createdByStr := r.URL.Query().Get("created_by"); createdByStr != "" {
		createdBy, err := uuid.Parse(createdByStr)
		if err == nil {
			req.CreatedBy = &createdBy
		}
	}

	// Pagination
	req.Limit = parseIntQueryParam(r, "limit", 20)
	req.Offset = parseIntQueryParam(r, "offset", 0)

	response, err := h.service.ListTournaments(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list tournaments", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to list tournaments")
		return
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// UpdateTournament handles PATCH /api/v1/tournaments/{id}
func (h *TournamentHandler) UpdateTournament(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	var req tournamentusecase.UpdateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tournament, err := h.service.UpdateTournament(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, tournamentdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Tournament not found")
			return
		}
		h.logger.Error("Failed to update tournament", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, tournament)
}

// UpdateTournamentStatus handles PATCH /api/v1/tournaments/{id}/status
func (h *TournamentHandler) UpdateTournamentStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	var req tournamentusecase.UpdateTournamentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tournament, err := h.service.UpdateTournamentStatus(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, tournamentdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Tournament not found")
			return
		}
		if errors.Is(err, tournamentdomain.ErrInvalidStatus) {
			h.errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to update tournament status", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to update tournament status")
		return
	}

	h.jsonResponse(w, http.StatusOK, tournament)
}

// DeleteTournament handles DELETE /api/v1/tournaments/{id}
func (h *TournamentHandler) DeleteTournament(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	if err := h.service.DeleteTournament(r.Context(), id); err != nil {
		if errors.Is(err, tournamentdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Tournament not found")
			return
		}
		h.logger.Error("Failed to delete tournament", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetActiveTournaments handles GET /api/v1/tournaments/active
func (h *TournamentHandler) GetActiveTournaments(w http.ResponseWriter, r *http.Request) {
	tournaments, err := h.service.GetActiveTournaments(r.Context())
	if err != nil {
		h.logger.Error("Failed to get active tournaments", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get active tournaments")
		return
	}

	h.jsonResponse(w, http.StatusOK, tournaments)
}

// GetPlayerActiveTournament handles GET /api/v1/players/me/active-tournament
func (h *TournamentHandler) GetPlayerActiveTournament(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context (set by auth middleware)
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	playerID, err := uuid.Parse(userInfo.ID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	tournament, err := h.service.GetActiveTournamentForPlayer(r.Context(), playerID)
	if err != nil {
		h.logger.Error("Failed to get player active tournament", "error", err, "player_id", playerID)
		h.errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, tournament)
}

// GetTournamentStats handles GET /api/v1/tournaments/{id}/stats
func (h *TournamentHandler) GetTournamentStats(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	stats, err := h.service.GetTournamentStats(r.Context(), id)
	if err != nil {
		if errors.Is(err, tournamentdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Tournament not found")
			return
		}
		h.logger.Error("Failed to get tournament stats", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get tournament stats")
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// jsonResponse writes a JSON response.
func (h *TournamentHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// errorResponse writes an error response.
func (h *TournamentHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}

// parseIntQueryParam parses an integer query parameter with a default value.
func parseIntQueryParam(r *http.Request, param string, defaultValue int) int {
	val := r.URL.Query().Get(param)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}
