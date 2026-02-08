package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	bracketdomain "github.com/alejaam/tourney-rank/internal/domain/bracket"
	bracketusecase "github.com/alejaam/tourney-rank/internal/usecase/bracket"
	"github.com/google/uuid"
)

// BracketHandler handles HTTP requests for bracket operations.
type BracketHandler struct {
	service *bracketusecase.Service
	logger  *slog.Logger
}

// NewBracketHandler creates a new bracket handler.
func NewBracketHandler(service *bracketusecase.Service, logger *slog.Logger) *BracketHandler {
	return &BracketHandler{
		service: service,
		logger:  logger,
	}
}

// GenerateBracket handles POST /api/v1/brackets/generate
func (h *BracketHandler) GenerateBracket(w http.ResponseWriter, r *http.Request) {
	var req bracketusecase.GenerateBracketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	bracket, err := h.service.GenerateBracket(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to generate bracket", "error", err)
		status := http.StatusInternalServerError
		message := "Failed to generate bracket"

		if errors.Is(err, bracketdomain.ErrInsufficientTeams) {
			status = http.StatusBadRequest
			message = "Insufficient teams to generate bracket (minimum 2 teams required)"
		} else if errors.Is(err, bracketdomain.ErrBracketAlreadyExists) {
			status = http.StatusConflict
			message = "Bracket already exists for this tournament"
		} else if errors.Is(err, bracketdomain.ErrInvalidFormat) {
			status = http.StatusBadRequest
			message = "Invalid bracket format"
		}

		h.errorResponse(w, status, message)
		return
	}

	h.jsonResponse(w, http.StatusCreated, bracket)
}

// GetBracket handles GET /api/v1/brackets/{id}
func (h *BracketHandler) GetBracket(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid bracket ID")
		return
	}

	bracket, err := h.service.GetBracket(r.Context(), id)
	if err != nil {
		if errors.Is(err, bracketdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Bracket not found")
			return
		}
		h.logger.Error("Failed to get bracket", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get bracket")
		return
	}

	h.jsonResponse(w, http.StatusOK, bracket)
}

// GetTournamentBracket handles GET /api/v1/tournaments/{id}/bracket
func (h *BracketHandler) GetTournamentBracket(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	tournamentID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	bracket, err := h.service.GetTournamentBracket(r.Context(), tournamentID)
	if err != nil {
		if errors.Is(err, bracketdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Bracket not found for this tournament")
			return
		}
		h.logger.Error("Failed to get tournament bracket", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to get tournament bracket")
		return
	}

	h.jsonResponse(w, http.StatusOK, bracket)
}

// SetMatchupWinner handles POST /api/v1/matchups/{id}/winner
func (h *BracketHandler) SetMatchupWinner(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	matchupID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid matchup ID")
		return
	}

	var req bracketusecase.SetMatchupWinnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	matchup, err := h.service.SetMatchupWinner(r.Context(), matchupID, req)
	if err != nil {
		h.logger.Error("Failed to set matchup winner", "error", err)
		status := http.StatusInternalServerError
		message := "Failed to set matchup winner"

		if errors.Is(err, bracketdomain.ErrMatchupNotFound) {
			status = http.StatusNotFound
			message = "Matchup not found"
		} else if errors.Is(err, bracketdomain.ErrMatchupAlreadyPlayed) {
			status = http.StatusConflict
			message = "Matchup already completed"
		}

		h.errorResponse(w, status, message)
		return
	}

	h.jsonResponse(w, http.StatusOK, matchup)
}

// DeleteBracket handles DELETE /api/v1/brackets/{id}
func (h *BracketHandler) DeleteBracket(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	bracketID, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid bracket ID")
		return
	}

	if err := h.service.DeleteBracket(r.Context(), bracketID); err != nil {
		if errors.Is(err, bracketdomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Bracket not found")
			return
		}
		h.logger.Error("Failed to delete bracket", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to delete bracket")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "Bracket deleted successfully"})
}

func (h *BracketHandler) errorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *BracketHandler) jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
