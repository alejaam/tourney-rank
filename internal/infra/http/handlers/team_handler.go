package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	teamdomain "github.com/melisource/tourney-rank/internal/domain/team"
	teamusecase "github.com/melisource/tourney-rank/internal/usecase/team"
)

// TeamHandler handles HTTP requests for team operations.
type TeamHandler struct {
	service *teamusecase.Service
	logger  *slog.Logger
}

// NewTeamHandler creates a new team handler.
func NewTeamHandler(service *teamusecase.Service, logger *slog.Logger) *TeamHandler {
	return &TeamHandler{
		service: service,
		logger:  logger,
	}
}

// CreateTeam handles POST /api/v1/teams
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req teamusecase.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get player ID from context (set by auth middleware)
	playerID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.service.CreateTeam(r.Context(), req, playerID)
	if err != nil {
		h.logger.Error("Failed to create team", "error", err)
		status := http.StatusInternalServerError
		message := "Failed to create team"

		if errors.Is(err, teamdomain.ErrInvalidName) {
			status = http.StatusBadRequest
			message = err.Error()
		} else if err.Error() == "tournament not found" || err.Error() == "player not found" {
			status = http.StatusBadRequest
			message = err.Error()
		} else if err.Error() == "tournament is full" ||
			err.Error() == "player is already in a team for this tournament" {
			status = http.StatusConflict
			message = err.Error()
		}

		h.errorResponse(w, status, message)
		return
	}

	h.jsonResponse(w, http.StatusCreated, team)
}

// GetTeam handles GET /api/v1/teams/{id}
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	team, err := h.service.GetTeam(r.Context(), id)
	if err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Team not found")
			return
		}
		h.logger.Error("Failed to get team", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get team")
		return
	}

	h.jsonResponse(w, http.StatusOK, team)
}

// GetTeamWithMembers handles GET /api/v1/teams/{id}/members
func (h *TeamHandler) GetTeamWithMembers(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	teamWithMembers, err := h.service.GetTeamWithMembers(r.Context(), id)
	if err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Team not found")
			return
		}
		h.logger.Error("Failed to get team with members", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get team details")
		return
	}

	h.jsonResponse(w, http.StatusOK, teamWithMembers)
}

// JoinTeam handles POST /api/v1/teams/join
func (h *TeamHandler) JoinTeam(w http.ResponseWriter, r *http.Request) {
	var req teamusecase.JoinTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get player ID from context (set by auth middleware)
	playerID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.service.JoinTeam(r.Context(), req, playerID)
	if err != nil {
		h.logger.Error("Failed to join team", "error", err, "player_id", playerID)
		status := http.StatusInternalServerError
		message := "Failed to join team"

		if errors.Is(err, teamdomain.ErrInvalidInviteCode) || errors.Is(err, teamdomain.ErrNotFound) {
			status = http.StatusNotFound
			message = "Invalid invite code"
		} else if errors.Is(err, teamdomain.ErrPlayerAlreadyInTeam) || errors.Is(err, teamdomain.ErrTeamFull) {
			status = http.StatusConflict
			message = err.Error()
		}

		h.errorResponse(w, status, message)
		return
	}

	h.jsonResponse(w, http.StatusOK, team)
}

// RemoveMember handles DELETE /api/v1/teams/{id}/members
func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req teamusecase.RemoveMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get player ID from context (set by auth middleware)
	requesterID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.service.RemoveMember(r.Context(), teamID, req, requesterID)
	if err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Team not found")
			return
		}
		if errors.Is(err, teamdomain.ErrNotCaptain) {
			h.errorResponse(w, http.StatusForbidden, "Only captain can remove members")
			return
		}
		if errors.Is(err, teamdomain.ErrCannotRemoveCaptain) {
			h.errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to remove member", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	h.jsonResponse(w, http.StatusOK, team)
}

// LeaveTeam handles POST /api/v1/teams/{id}/leave
func (h *TeamHandler) LeaveTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Get player ID from context (set by auth middleware)
	playerID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.LeaveTeam(r.Context(), teamID, playerID); err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Team not found")
			return
		}
		h.logger.Error("Failed to leave team", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TransferCaptaincy handles POST /api/v1/teams/{id}/transfer-captain
func (h *TeamHandler) TransferCaptaincy(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req teamusecase.TransferCaptaincyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get player ID from context (set by auth middleware)
	currentCaptainID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.service.TransferCaptaincy(r.Context(), teamID, req, currentCaptainID)
	if err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Team not found")
			return
		}
		if errors.Is(err, teamdomain.ErrNotCaptain) {
			h.errorResponse(w, http.StatusForbidden, "Only captain can transfer captaincy")
			return
		}
		h.logger.Error("Failed to transfer captaincy", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, team)
}

// ListTeamsByTournament handles GET /api/v1/tournaments/{tournamentId}/teams
func (h *TeamHandler) ListTeamsByTournament(w http.ResponseWriter, r *http.Request) {
	tournamentIDStr := r.PathValue("tournamentId")
	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	teams, err := h.service.ListTeamsByTournament(r.Context(), tournamentID)
	if err != nil {
		h.logger.Error("Failed to list teams", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to list teams")
		return
	}

	h.jsonResponse(w, http.StatusOK, teams)
}

// GetPlayerTeamInTournament handles GET /api/v1/tournaments/{tournamentId}/my-team
func (h *TeamHandler) GetPlayerTeamInTournament(w http.ResponseWriter, r *http.Request) {
	tournamentIDStr := r.PathValue("tournamentId")
	tournamentID, err := uuid.Parse(tournamentIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	// Get player ID from context (set by auth middleware)
	playerID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.service.GetPlayerTeamInTournament(r.Context(), playerID, tournamentID)
	if err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "No team found for this tournament")
			return
		}
		h.logger.Error("Failed to get player team", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get team")
		return
	}

	h.jsonResponse(w, http.StatusOK, team)
}

// GetPlayerTeams handles GET /api/v1/players/me/teams
func (h *TeamHandler) GetPlayerTeams(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context (set by auth middleware)
	playerID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	teams, err := h.service.GetPlayerTeams(r.Context(), playerID)
	if err != nil {
		h.logger.Error("Failed to get player teams", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get teams")
		return
	}

	h.jsonResponse(w, http.StatusOK, teams)
}

// UpdateTeam handles PATCH /api/v1/teams/{id}
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req teamusecase.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get player ID from context (set by auth middleware)
	playerID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.service.UpdateTeam(r.Context(), teamID, req, playerID)
	if err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Team not found")
			return
		}
		if errors.Is(err, teamdomain.ErrNotCaptain) {
			h.errorResponse(w, http.StatusForbidden, "Only captain can update team")
			return
		}
		h.logger.Error("Failed to update team", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to update team")
		return
	}

	h.jsonResponse(w, http.StatusOK, team)
}

// DisbandTeam handles DELETE /api/v1/teams/{id}
func (h *TeamHandler) DisbandTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Get player ID from context (set by auth middleware)
	playerID, ok := r.Context().Value("player_id").(uuid.UUID)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.DisbandTeam(r.Context(), teamID, playerID); err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Team not found")
			return
		}
		if errors.Is(err, teamdomain.ErrNotCaptain) {
			h.errorResponse(w, http.StatusForbidden, "Only captain can disband team")
			return
		}
		h.logger.Error("Failed to disband team", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// jsonResponse writes a JSON response.
func (h *TeamHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// errorResponse writes an error response.
func (h *TeamHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
