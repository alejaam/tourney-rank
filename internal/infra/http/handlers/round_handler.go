package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	rounddomain "github.com/alejaam/tourney-rank/internal/domain/round"
	tournamentdomain "github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/alejaam/tourney-rank/internal/domain/user"
	"github.com/alejaam/tourney-rank/internal/infra/http/middleware"
	roundusecase "github.com/alejaam/tourney-rank/internal/usecase/round"
	"github.com/google/uuid"
)

// RoundHandler handles HTTP requests for round operations.
type RoundHandler struct {
	service *roundusecase.Service
	logger  *slog.Logger
}

// NewRoundHandler creates a new round handler.
func NewRoundHandler(service *roundusecase.Service, logger *slog.Logger) *RoundHandler {
	return &RoundHandler{
		service: service,
		logger:  logger,
	}
}

// CreateRound handles POST /api/v1/tournaments/:id/rounds
func (h *RoundHandler) CreateRound(w http.ResponseWriter, r *http.Request) {
	tournamentID := r.PathValue("id")
	if tournamentID == "" {
		h.errorResponse(w, http.StatusBadRequest, "Tournament ID is required")
		return
	}

	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	var req roundusecase.CreateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Require admin authorization
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok || userInfo.Role != user.RoleAdmin {
		h.errorResponse(w, http.StatusForbidden, "Only tournament admins can create rounds")
		return
	}

	req.TournamentID = tID
	round, err := h.service.CreateRound(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create round", "error", err)
		switch {
		case errors.Is(err, tournamentdomain.ErrNotFound):
			h.errorResponse(w, http.StatusNotFound, "Tournament not found")
		case errors.Is(err, rounddomain.ErrInvalidRoundNum):
			h.errorResponse(w, http.StatusBadRequest, "Invalid round number")
		default:
			h.errorResponse(w, http.StatusInternalServerError, "Failed to create round")
		}
		return
	}

	h.successResponse(w, http.StatusCreated, round)
}

// GetRound handles GET /api/v1/tournaments/:id/rounds/:roundNum
func (h *RoundHandler) GetRound(w http.ResponseWriter, r *http.Request) {
	tournamentID := r.PathValue("id")
	roundNum := r.PathValue("roundNum")

	if tournamentID == "" || roundNum == "" {
		h.errorResponse(w, http.StatusBadRequest, "Tournament ID and round number are required")
		return
	}

	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	num, err := strconv.Atoi(roundNum)
	if err != nil || num <= 0 {
		h.errorResponse(w, http.StatusBadRequest, "Invalid round number")
		return
	}

	round, err := h.service.GetRoundByNumber(r.Context(), tID, num)
	if err != nil {
		h.logger.Error("Failed to get round", "error", err)
		if errors.Is(err, rounddomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Round not found")
		} else {
			h.errorResponse(w, http.StatusInternalServerError, "Failed to get round")
		}
		return
	}

	h.successResponse(w, http.StatusOK, round)
}

// ListRounds handles GET /api/v1/tournaments/:id/rounds
func (h *RoundHandler) ListRounds(w http.ResponseWriter, r *http.Request) {
	tournamentID := r.PathValue("id")
	if tournamentID == "" {
		h.errorResponse(w, http.StatusBadRequest, "Tournament ID is required")
		return
	}

	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	// Optional status filter
	status := r.URL.Query().Get("status")
	var statusFilter *rounddomain.Status
	if status != "" {
		s := rounddomain.Status(status)
		// Validate status
		switch s {
		case rounddomain.StatusPending, rounddomain.StatusOngoing, rounddomain.StatusCompleted, rounddomain.StatusCanceled:
			statusFilter = &s
		default:
			h.errorResponse(w, http.StatusBadRequest, "Invalid status filter")
			return
		}
	}

	rounds, err := h.service.GetTournamentRounds(r.Context(), tID, statusFilter)
	if err != nil {
		h.logger.Error("Failed to list rounds", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "Failed to list rounds")
		return
	}

	h.successResponse(w, http.StatusOK, rounds)
}

// UpdateRound handles PATCH /api/v1/tournaments/:id/rounds/:roundNum
func (h *RoundHandler) UpdateRound(w http.ResponseWriter, r *http.Request) {
	tournamentID := r.PathValue("id")
	roundNum := r.PathValue("roundNum")

	if tournamentID == "" || roundNum == "" {
		h.errorResponse(w, http.StatusBadRequest, "Tournament ID and round number are required")
		return
	}

	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	num, err := strconv.Atoi(roundNum)
	if err != nil || num <= 0 {
		h.errorResponse(w, http.StatusBadRequest, "Invalid round number")
		return
	}

	// Require admin authorization
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok || userInfo.Role != user.RoleAdmin {
		h.errorResponse(w, http.StatusForbidden, "Only tournament admins can update rounds")
		return
	}

	var req roundusecase.UpdateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get round first
	round, err := h.service.GetRoundByNumber(r.Context(), tID, num)
	if err != nil {
		h.logger.Error("Failed to get round", "error", err)
		h.errorResponse(w, http.StatusNotFound, "Round not found")
		return
	}

	updatedRound, err := h.service.UpdateRound(r.Context(), round.ID, req)
	if err != nil {
		h.logger.Error("Failed to update round", "error", err)
		if errors.Is(err, rounddomain.ErrNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Round not found")
			return
		}
		h.errorResponse(w, http.StatusInternalServerError, "Failed to update round")
		return
	}

	h.successResponse(w, http.StatusOK, updatedRound)
}

// UpdateRoundStatus handles PATCH /api/v1/tournaments/:id/rounds/:roundNum/status
func (h *RoundHandler) UpdateRoundStatus(w http.ResponseWriter, r *http.Request) {
	tournamentID := r.PathValue("id")
	roundNum := r.PathValue("roundNum")

	if tournamentID == "" || roundNum == "" {
		h.errorResponse(w, http.StatusBadRequest, "Tournament ID and round number are required")
		return
	}

	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	num, err := strconv.Atoi(roundNum)
	if err != nil || num <= 0 {
		h.errorResponse(w, http.StatusBadRequest, "Invalid round number")
		return
	}

	// Require admin authorization
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok || userInfo.Role != user.RoleAdmin {
		h.errorResponse(w, http.StatusForbidden, "Only tournament admins can update round status")
		return
	}

	var req roundusecase.UpdateRoundStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get round first
	round, err := h.service.GetRoundByNumber(r.Context(), tID, num)
	if err != nil {
		h.logger.Error("Failed to get round", "error", err)
		h.errorResponse(w, http.StatusNotFound, "Round not found")
		return
	}

	updatedRound, err := h.service.UpdateRoundStatus(r.Context(), round.ID, req)
	if err != nil {
		h.logger.Error("Failed to update round status", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, http.StatusOK, updatedRound)
}

// DeleteRound handles DELETE /api/v1/tournaments/:id/rounds/:roundNum
func (h *RoundHandler) DeleteRound(w http.ResponseWriter, r *http.Request) {
	tournamentID := r.PathValue("id")
	roundNum := r.PathValue("roundNum")

	if tournamentID == "" || roundNum == "" {
		h.errorResponse(w, http.StatusBadRequest, "Tournament ID and round number are required")
		return
	}

	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid tournament ID")
		return
	}

	num, err := strconv.Atoi(roundNum)
	if err != nil || num <= 0 {
		h.errorResponse(w, http.StatusBadRequest, "Invalid round number")
		return
	}

	// Require admin authorization
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok || userInfo.Role != user.RoleAdmin {
		h.errorResponse(w, http.StatusForbidden, "Only tournament admins can delete rounds")
		return
	}

	// Get round first
	round, err := h.service.GetRoundByNumber(r.Context(), tID, num)
	if err != nil {
		h.logger.Error("Failed to get round", "error", err)
		h.errorResponse(w, http.StatusNotFound, "Round not found")
		return
	}

	if err := h.service.DeleteRound(r.Context(), round.ID); err != nil {
		h.logger.Error("Failed to delete round", "error", err)
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, http.StatusNoContent, nil)
}

// successResponse encodes a success response.
func (h *RoundHandler) successResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// errorResponse encodes an error response.
func (h *RoundHandler) errorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
