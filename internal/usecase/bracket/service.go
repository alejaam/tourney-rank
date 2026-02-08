// Package bracket provides use cases for bracket management.
package bracket

import (
	"context"
	"errors"

	"github.com/alejaam/tourney-rank/internal/domain/bracket"
	"github.com/alejaam/tourney-rank/internal/domain/team"
	"github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/google/uuid"
)

// Service handles bracket use cases.
type Service struct {
	bracketRepo    bracket.Repository
	teamRepo       team.Repository
	tournamentRepo tournament.Repository
	generator      *bracket.Generator
}

// NewService creates a new bracket service.
func NewService(
	bracketRepo bracket.Repository,
	teamRepo team.Repository,
	tournamentRepo tournament.Repository,
) *Service {
	return &Service{
		bracketRepo:    bracketRepo,
		teamRepo:       teamRepo,
		tournamentRepo: tournamentRepo,
		generator:      bracket.NewGenerator(),
	}
}

// GenerateBracketRequest represents the request to generate a bracket.
type GenerateBracketRequest struct {
	TournamentID uuid.UUID      `json:"tournament_id"`
	Format       bracket.Format `json:"format"`
	IsSeeded     bool           `json:"is_seeded"`
}

// MatchupResponse represents a matchup with team details.
type MatchupResponse struct {
	*bracket.Matchup
	Team1Name *string `json:"team1_name,omitempty"`
	Team2Name *string `json:"team2_name,omitempty"`
}

// BracketWithMatchups represents a bracket with all its matchups.
type BracketWithMatchups struct {
	*bracket.Bracket
	Matchups []*MatchupResponse `json:"matchups"`
}

// GenerateBracket generates a bracket for a tournament.
func (s *Service) GenerateBracket(ctx context.Context, req GenerateBracketRequest) (*BracketWithMatchups, error) {
	// Verify tournament exists
	t, err := s.tournamentRepo.GetByID(ctx, req.TournamentID)
	if err != nil {
		return nil, err
	}

	// Check if bracket already exists
	existingBracket, err := s.bracketRepo.GetByTournamentID(ctx, req.TournamentID)
	if err == nil && existingBracket != nil {
		return nil, bracket.ErrBracketAlreadyExists
	}

	// Get all teams for the tournament
	teams, err := s.teamRepo.GetByTournamentID(ctx, req.TournamentID)
	if err != nil {
		return nil, err
	}

	if len(teams) < 2 {
		return nil, bracket.ErrInsufficientTeams
	}

	// Convert teams to seeds
	teamSeeds := make([]bracket.TeamSeed, len(teams))
	for i, tm := range teams {
		teamSeeds[i] = bracket.TeamSeed{
			TeamID: tm.ID,
			Seed:   i + 1,
		}
	}

	// Calculate total rounds
	totalRounds := bracket.CalculateTotalRounds(req.Format, len(teams))

	// Create bracket
	b, err := bracket.NewBracket(req.TournamentID, req.Format, totalRounds, req.IsSeeded)
	if err != nil {
		return nil, err
	}

	if err := s.bracketRepo.Create(ctx, b); err != nil {
		return nil, err
	}

	// Generate matchups based on format
	var matchups []*bracket.Matchup
	switch req.Format {
	case bracket.FormatSingleElimination:
		matchups, err = s.generator.GenerateSingleEliminationMatchups(b, teamSeeds)
	case bracket.FormatRoundRobin:
		matchups, err = s.generator.GenerateRoundRobinMatchups(b, teamSeeds)
	default:
		return nil, errors.New("unsupported bracket format")
	}

	if err != nil {
		return nil, err
	}

	// Save all matchups
	for _, matchup := range matchups {
		if err := s.bracketRepo.CreateMatchup(ctx, matchup); err != nil {
			return nil, err
		}
	}

	// Update tournament status to active
	if err := t.UpdateStatus(tournament.StatusActive); err != nil {
		return nil, err
	}
	if err := s.tournamentRepo.Update(ctx, t); err != nil {
		return nil, err
	}

	// Build response with team names
	matchupResponses := make([]*MatchupResponse, len(matchups))
	for i, m := range matchups {
		mr := &MatchupResponse{Matchup: m}

		if m.Team1ID != nil {
			if team1, err := s.teamRepo.GetByID(ctx, *m.Team1ID); err == nil {
				mr.Team1Name = &team1.Name
			}
		}

		if m.Team2ID != nil {
			if team2, err := s.teamRepo.GetByID(ctx, *m.Team2ID); err == nil {
				mr.Team2Name = &team2.Name
			}
		}

		matchupResponses[i] = mr
	}

	return &BracketWithMatchups{
		Bracket:  b,
		Matchups: matchupResponses,
	}, nil
}

// GetBracket retrieves a bracket by ID.
func (s *Service) GetBracket(ctx context.Context, id uuid.UUID) (*bracket.Bracket, error) {
	return s.bracketRepo.GetByID(ctx, id)
}

// GetTournamentBracket retrieves the bracket for a tournament.
func (s *Service) GetTournamentBracket(ctx context.Context, tournamentID uuid.UUID) (*BracketWithMatchups, error) {
	b, err := s.bracketRepo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	matchups, err := s.bracketRepo.GetMatchupsByBracket(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	// Build response with team names
	matchupResponses := make([]*MatchupResponse, len(matchups))
	for i, m := range matchups {
		mr := &MatchupResponse{Matchup: m}

		if m.Team1ID != nil {
			if team1, err := s.teamRepo.GetByID(ctx, *m.Team1ID); err == nil {
				mr.Team1Name = &team1.Name
			}
		}

		if m.Team2ID != nil {
			if team2, err := s.teamRepo.GetByID(ctx, *m.Team2ID); err == nil {
				mr.Team2Name = &team2.Name
			}
		}

		matchupResponses[i] = mr
	}

	return &BracketWithMatchups{
		Bracket:  b,
		Matchups: matchupResponses,
	}, nil
}

// SetMatchupWinnerRequest represents the request to set a matchup winner.
type SetMatchupWinnerRequest struct {
	WinnerID uuid.UUID `json:"winner_id"`
}

// SetMatchupWinner sets the winner of a matchup and advances bracket if needed.
func (s *Service) SetMatchupWinner(ctx context.Context, matchupID uuid.UUID, req SetMatchupWinnerRequest) (*bracket.Matchup, error) {
	matchup, err := s.bracketRepo.GetMatchup(ctx, matchupID)
	if err != nil {
		return nil, err
	}

	if err := matchup.SetWinner(req.WinnerID); err != nil {
		return nil, err
	}

	if err := s.bracketRepo.UpdateMatchup(ctx, matchup); err != nil {
		return nil, err
	}

	// Check if all matchups in current round are complete
	b, err := s.bracketRepo.GetByID(ctx, matchup.BracketID)
	if err != nil {
		return nil, err
	}

	roundMatchups, err := s.bracketRepo.GetMatchupsByRound(ctx, b.ID, b.CurrentRound)
	if err != nil {
		return nil, err
	}

	allComplete := true
	for _, rm := range roundMatchups {
		if rm.Status != bracket.MatchupStatusCompleted {
			allComplete = false
			break
		}
	}

	// If all complete and not final round, generate next round
	if allComplete && b.CurrentRound < b.TotalRounds && b.Format == bracket.FormatSingleElimination {
		nextMatchups, err := s.generator.GenerateNextRoundMatchups(b, roundMatchups)
		if err != nil {
			return nil, err
		}

		for _, nm := range nextMatchups {
			if err := s.bracketRepo.CreateMatchup(ctx, nm); err != nil {
				return nil, err
			}
		}

		if err := b.AdvanceRound(); err != nil {
			return nil, err
		}

		if err := s.bracketRepo.Update(ctx, b); err != nil {
			return nil, err
		}
	}

	return matchup, nil
}

// DeleteBracket deletes a bracket and all its matchups.
func (s *Service) DeleteBracket(ctx context.Context, bracketID uuid.UUID) error {
	// Get all matchups first
	matchups, err := s.bracketRepo.GetMatchupsByBracket(ctx, bracketID)
	if err != nil {
		return err
	}

	// Delete all matchups
	for _, m := range matchups {
		if err := s.bracketRepo.DeleteMatchup(ctx, m.ID); err != nil {
			return err
		}
	}

	// Delete bracket
	return s.bracketRepo.Delete(ctx, bracketID)
}
