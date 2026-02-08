// Package tournament provides use cases for tournament management.
package tournament

import (
	"context"
	"time"

	"github.com/alejaam/tourney-rank/internal/domain/game"
	"github.com/alejaam/tourney-rank/internal/domain/team"
	"github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/google/uuid"
)

// Service handles tournament use cases.
type Service struct {
	tournamentRepo tournament.Repository
	teamRepo       team.Repository
	gameRepo       game.Repository
}

// NewService creates a new tournament service.
func NewService(tournamentRepo tournament.Repository, teamRepo team.Repository, gameRepo game.Repository) *Service {
	return &Service{
		tournamentRepo: tournamentRepo,
		teamRepo:       teamRepo,
		gameRepo:       gameRepo,
	}
}

// CreateTournamentRequest represents the request to create a tournament.
type CreateTournamentRequest struct {
	GameID      uuid.UUID           `json:"game_id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	TeamSize    tournament.TeamSize `json:"team_size"`
	StartDate   time.Time           `json:"start_date"`
	EndDate     time.Time           `json:"end_date"`
	PrizePool   string              `json:"prize_pool,omitempty"`
	BannerURL   string              `json:"banner_url,omitempty"`
	Rules       tournament.Rules    `json:"rules"`
}

// UpdateTournamentRequest represents the request to update a tournament.
type UpdateTournamentRequest struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	StartDate   *time.Time        `json:"start_date,omitempty"`
	EndDate     *time.Time        `json:"end_date,omitempty"`
	PrizePool   *string           `json:"prize_pool,omitempty"`
	BannerURL   *string           `json:"banner_url,omitempty"`
	Rules       *tournament.Rules `json:"rules,omitempty"`
}

// UpdateTournamentStatusRequest represents the request to update tournament status.
type UpdateTournamentStatusRequest struct {
	Status tournament.Status `json:"status"`
}

// ListTournamentsRequest represents the request to list tournaments.
type ListTournamentsRequest struct {
	GameID    *uuid.UUID         `json:"game_id,omitempty"`
	Status    *tournament.Status `json:"status,omitempty"`
	CreatedBy *uuid.UUID         `json:"created_by,omitempty"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

// TournamentListResponse represents a paginated list of tournaments.
type TournamentListResponse struct {
	Tournaments []*tournament.Tournament `json:"tournaments"`
	Total       int64                    `json:"total"`
	Limit       int                      `json:"limit"`
	Offset      int                      `json:"offset"`
}

// TournamentStats represents statistics for a tournament.
type TournamentStats struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	TotalTeams   int64     `json:"total_teams"`
	ActiveTeams  int64     `json:"active_teams"`
	TotalMatches int64     `json:"total_matches"`
	TotalPlayers int64     `json:"total_players"`
}

// CreateTournament creates a new tournament.
func (s *Service) CreateTournament(ctx context.Context, req CreateTournamentRequest, createdBy uuid.UUID) (*tournament.Tournament, error) {
	// Validate game exists
	_, err := s.gameRepo.GetByID(ctx, req.GameID.String())
	if err != nil {
		return nil, err
	}

	t, err := tournament.NewTournament(req.GameID, createdBy, req.Name, req.TeamSize, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	t.Description = req.Description
	t.PrizePool = req.PrizePool
	t.BannerURL = req.BannerURL
	t.Rules = req.Rules

	if err := s.tournamentRepo.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

// UpdateTournament updates an existing tournament.
func (s *Service) UpdateTournament(ctx context.Context, id uuid.UUID, req UpdateTournamentRequest) (*tournament.Tournament, error) {
	t, err := s.tournamentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.StartDate != nil {
		t.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		t.EndDate = *req.EndDate
	}
	if req.PrizePool != nil {
		t.PrizePool = *req.PrizePool
	}
	if req.BannerURL != nil {
		t.BannerURL = *req.BannerURL
	}
	if req.Rules != nil {
		t.Rules = *req.Rules
	}

	t.UpdatedAt = time.Now().UTC()

	if err := s.tournamentRepo.Update(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

// UpdateTournamentStatus updates the status of a tournament.
func (s *Service) UpdateTournamentStatus(ctx context.Context, id uuid.UUID, req UpdateTournamentStatusRequest) (*tournament.Tournament, error) {
	t, err := s.tournamentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := t.UpdateStatus(req.Status); err != nil {
		return nil, err
	}

	if err := s.tournamentRepo.Update(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

// GetTournament retrieves a tournament by ID.
func (s *Service) GetTournament(ctx context.Context, id uuid.UUID) (*tournament.Tournament, error) {
	return s.tournamentRepo.GetByID(ctx, id)
}

// ListTournaments lists tournaments with optional filtering.
func (s *Service) ListTournaments(ctx context.Context, req ListTournamentsRequest) (*TournamentListResponse, error) {
	filter := tournament.ListFilter{
		GameID:    req.GameID,
		Status:    req.Status,
		CreatedBy: req.CreatedBy,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	tournaments, err := s.tournamentRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// For simplicity, not implementing total count here
	// In a real implementation, you'd add a Count method to the repository
	return &TournamentListResponse{
		Tournaments: tournaments,
		Total:       int64(len(tournaments)),
		Limit:       req.Limit,
		Offset:      req.Offset,
	}, nil
}

// GetActiveTournaments retrieves all active tournaments.
func (s *Service) GetActiveTournaments(ctx context.Context) ([]*tournament.Tournament, error) {
	return s.tournamentRepo.GetActiveTournaments(ctx)
}

// GetActiveTournamentForPlayer retrieves the active tournament for a player.
func (s *Service) GetActiveTournamentForPlayer(ctx context.Context, playerID uuid.UUID) (*tournament.Tournament, error) {
	// Get all active tournaments
	active, err := s.tournamentRepo.GetActiveTournaments(ctx)
	if err != nil {
		return nil, err
	}

	// For each active tournament, check if player has a team
	for _, t := range active {
		teams, err := s.teamRepo.GetByTournamentID(ctx, t.ID)
		if err != nil {
			continue
		}

		for _, tm := range teams {
			if tm.HasMember(playerID) {
				return t, nil
			}
		}
	}

	return nil, tournament.ErrNotFound
}

// DeleteTournament deletes a tournament.
func (s *Service) DeleteTournament(ctx context.Context, id uuid.UUID) error {
	return s.tournamentRepo.Delete(ctx, id)
}

// GetTournamentStats retrieves statistics for a tournament.
func (s *Service) GetTournamentStats(ctx context.Context, id uuid.UUID) (*TournamentStats, error) {
	// Verify tournament exists
	_, err := s.tournamentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	teams, err := s.teamRepo.GetByTournamentID(ctx, id)
	if err != nil {
		return nil, err
	}

	totalTeams := int64(len(teams))
	activeTeams := int64(0)
	totalPlayers := int64(0)

	for _, t := range teams {
		if t.Status == team.StatusActive || t.Status == team.StatusReady {
			activeTeams++
		}
		totalPlayers += int64(t.MemberCount())
	}

	return &TournamentStats{
		TournamentID: id,
		TotalTeams:   totalTeams,
		ActiveTeams:  activeTeams,
		TotalMatches: 0, // Would need match repository to calculate
		TotalPlayers: totalPlayers,
	}, nil
}
