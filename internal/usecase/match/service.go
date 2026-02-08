package match

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	matchdomain "github.com/alejaam/tourney-rank/internal/domain/match"
	playerdomain "github.com/alejaam/tourney-rank/internal/domain/player"
	rankingdomain "github.com/alejaam/tourney-rank/internal/domain/ranking"
	teamdomain "github.com/alejaam/tourney-rank/internal/domain/team"
	tournamentdomain "github.com/alejaam/tourney-rank/internal/domain/tournament"
	usecaseplayer "github.com/alejaam/tourney-rank/internal/usecase/player"
)

// Service provides match operations.
type Service struct {
	matchRepo       matchdomain.Repository
	teamRepo        teamdomain.Repository
	tournamentRepo  tournamentdomain.Repository
	playerRepo      playerdomain.Repository
	playerStatsRepo playerdomain.StatsRepository
	playerService   *usecaseplayer.Service
	ranking         *rankingdomain.Service
}

// NewService creates a new match service.
func NewService(
	matchRepo matchdomain.Repository,
	teamRepo teamdomain.Repository,
	tournamentRepo tournamentdomain.Repository,
	playerRepo playerdomain.Repository,
	playerStatsRepo playerdomain.StatsRepository,
	playerService *usecaseplayer.Service,
	ranking *rankingdomain.Service,
) *Service {
	return &Service{
		matchRepo:       matchRepo,
		teamRepo:        teamRepo,
		tournamentRepo:  tournamentRepo,
		playerRepo:      playerRepo,
		playerStatsRepo: playerStatsRepo,
		playerService:   playerService,
		ranking:         ranking,
	}
}

// PlayerStatsInput represents player stats in a match submission.
type PlayerStatsInput struct {
	PlayerID    uuid.UUID              `json:"player_id"`
	Kills       int                    `json:"kills"`
	Damage      int                    `json:"damage"`
	Assists     int                    `json:"assists"`
	Deaths      int                    `json:"deaths"`
	Downs       int                    `json:"downs"`
	CustomStats map[string]interface{} `json:"custom_stats,omitempty"`
}

// SubmitMatchRequest represents a match submission request.
type SubmitMatchRequest struct {
	TournamentID  uuid.UUID          `json:"tournament_id"`
	TeamID        uuid.UUID          `json:"team_id"`
	GameID        uuid.UUID          `json:"game_id"`
	TeamPlacement int                `json:"team_placement"`
	TeamKills     int                `json:"team_kills"`
	PlayerStats   []PlayerStatsInput `json:"player_stats"`
	ScreenshotURL string             `json:"screenshot_url"`
}

// MatchResponse represents a match in API responses.
type MatchResponse struct {
	ID              uuid.UUID                      `json:"id"`
	TournamentID    uuid.UUID                      `json:"tournament_id"`
	TeamID          uuid.UUID                      `json:"team_id"`
	GameID          uuid.UUID                      `json:"game_id"`
	Status          string                         `json:"status"`
	TeamPlacement   int                            `json:"team_placement"`
	TeamKills       int                            `json:"team_kills"`
	PlayerStats     []matchdomain.PlayerMatchStats `json:"player_stats"`
	ScreenshotURL   string                         `json:"screenshot_url"`
	RejectionReason string                         `json:"rejection_reason,omitempty"`
	SubmittedBy     uuid.UUID                      `json:"submitted_by"`
	CreatedAt       string                         `json:"created_at"`
	UpdatedAt       string                         `json:"updated_at"`
	VerifiedAt      *string                        `json:"verified_at,omitempty"`
	VerifiedBy      *uuid.UUID                     `json:"verified_by,omitempty"`
}

// MatchHistoryRequest represents a request for match history with pagination.
type MatchHistoryRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// MatchListResponse represents a list of matches in API responses.
type MatchListResponse struct {
	Matches []MatchResponse `json:"matches"`
	Total   int             `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
}

// VerifyMatchRequest represents a request to verify or reject a match.
type VerifyMatchRequest struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

// SubmitMatch submits a new match report for verification.
func (s *Service) SubmitMatch(ctx context.Context, req SubmitMatchRequest, captainID uuid.UUID) (*MatchResponse, error) {
	// Verify tournament exists and is active
	tournament, err := s.tournamentRepo.GetByID(ctx, req.TournamentID)
	if err != nil {
		if errors.Is(err, tournamentdomain.ErrNotFound) {
			return nil, fmt.Errorf("tournament not found")
		}
		return nil, fmt.Errorf("get tournament: %w", err)
	}

	if tournament.Status != tournamentdomain.StatusActive {
		return nil, matchdomain.ErrTournamentNotActive
	}

	// Verify team exists
	team, err := s.teamRepo.GetByID(ctx, req.TeamID)
	if err != nil {
		if errors.Is(err, teamdomain.ErrNotFound) {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("get team: %w", err)
	}

	// Verify captain is the team captain
	if team.CaptainID != captainID {
		return nil, matchdomain.ErrNotCaptain
	}

	// Convert player stats
	playerStats := make([]matchdomain.PlayerMatchStats, len(req.PlayerStats))
	for i, ps := range req.PlayerStats {
		playerStats[i] = matchdomain.PlayerMatchStats{
			PlayerID:    ps.PlayerID,
			Kills:       ps.Kills,
			Damage:      ps.Damage,
			Assists:     ps.Assists,
			Deaths:      ps.Deaths,
			Downs:       ps.Downs,
			CustomStats: ps.CustomStats,
		}

		// Verify player is in team
		found := false
		for _, memberID := range team.MemberIDs {
			if memberID == ps.PlayerID {
				found = true
				break
			}
		}
		if !found {
			return nil, matchdomain.ErrPlayerNotInTeam
		}
	}

	// Verify all team members have stats (if team size is defined)
	if len(playerStats) != len(team.MemberIDs) {
		return nil, matchdomain.ErrTeamSizeMismatch
	}

	// Create match entity
	m, err := matchdomain.NewMatch(
		req.TournamentID,
		req.TeamID,
		req.GameID,
		req.TeamPlacement,
		req.TeamKills,
		playerStats,
		req.ScreenshotURL,
		captainID,
	)
	if err != nil {
		return nil, fmt.Errorf("create match: %w", err)
	}

	// Store match
	if err := s.matchRepo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("store match: %w", err)
	}

	return matchToResponse(m), nil
}

// GetMatchHistory retrieves a player's match history.
func (s *Service) GetMatchHistory(ctx context.Context, playerID uuid.UUID, req MatchHistoryRequest) (*MatchListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Get verified matches for the player
	matches, err := s.matchRepo.GetByPlayer(ctx, playerID.String(), req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("get player matches: %w", err)
	}

	// Filter only verified matches
	var verifiedMatches []matchdomain.Match
	for _, m := range matches {
		if m.Status == matchdomain.StatusVerified {
			verifiedMatches = append(verifiedMatches, m)
		}
	}

	responses := make([]MatchResponse, len(verifiedMatches))
	for i, m := range verifiedMatches {
		responses[i] = *matchToResponse(&m)
	}

	return &MatchListResponse{
		Matches: responses,
		Total:   len(verifiedMatches),
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// GetTournamentMatches retrieves all verified matches in a tournament.
func (s *Service) GetTournamentMatches(ctx context.Context, tournamentID uuid.UUID, req MatchHistoryRequest) (*MatchListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	matches, err := s.matchRepo.GetByTournament(ctx, tournamentID.String(), req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("get tournament matches: %w", err)
	}

	// Filter only verified matches
	var verifiedMatches []matchdomain.Match
	for _, m := range matches {
		if m.Status == matchdomain.StatusVerified {
			verifiedMatches = append(verifiedMatches, m)
		}
	}

	responses := make([]MatchResponse, len(verifiedMatches))
	for i, m := range verifiedMatches {
		responses[i] = *matchToResponse(&m)
	}

	return &MatchListResponse{
		Matches: responses,
		Total:   len(verifiedMatches),
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// AdminVerifyMatch approves or rejects a match report.
func (s *Service) AdminVerifyMatch(ctx context.Context, matchID uuid.UUID, req VerifyMatchRequest, adminID uuid.UUID) (*MatchResponse, error) {
	// Get match
	m, err := s.matchRepo.GetByID(ctx, matchID.String())
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}

	// Process verification/rejection
	if req.Approved {
		if err := m.VerifyMatch(adminID); err != nil {
			return nil, fmt.Errorf("verify match: %w", err)
		}

		// Update player stats after verification
		if err := s.updatePlayerStatsFromMatch(ctx, m); err != nil {
			return nil, fmt.Errorf("update player stats: %w", err)
		}
	} else {
		if err := m.RejectMatch(adminID, req.Reason); err != nil {
			return nil, fmt.Errorf("reject match: %w", err)
		}
	}

	// Update match in repository
	if err := s.matchRepo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("update match: %w", err)
	}

	return matchToResponse(m), nil
}

// GetUnverifiedMatches retrieves all unverified matches for admin review.
func (s *Service) GetUnverifiedMatches(ctx context.Context, req MatchHistoryRequest) (*MatchListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	matches, err := s.matchRepo.GetUnverified(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("get unverified matches: %w", err)
	}

	responses := make([]MatchResponse, len(matches))
	for i, m := range matches {
		responses[i] = *matchToResponse(&m)
	}

	return &MatchListResponse{
		Matches: responses,
		Total:   len(matches),
		Limit:   req.Limit,
		Offset:  req.Offset,
	}, nil
}

// updatePlayerStatsFromMatch updates player stats after match verification.
func (s *Service) updatePlayerStatsFromMatch(ctx context.Context, m *matchdomain.Match) error {
	for _, ps := range m.PlayerStats {
		// Get or create player stats for this game
		stats, err := s.playerStatsRepo.GetOrCreate(ctx, ps.PlayerID, m.GameID)
		if err != nil {
			return fmt.Errorf("get or create player stats: %w", err)
		}

		// Update stats from match
		statsToAdd := map[string]interface{}{
			"total_kills":   stats.GetStatAsInt("total_kills") + ps.Kills,
			"total_damage":  stats.GetStatAsInt("total_damage") + ps.Damage,
			"total_assists": stats.GetStatAsInt("total_assists") + ps.Assists,
			"total_deaths":  stats.GetStatAsInt("total_deaths") + ps.Deaths,
			"total_downs":   stats.GetStatAsInt("total_downs") + ps.Downs,
		}

		// Add custom stats if present
		for key, val := range ps.CustomStats {
			if key != "total_kills" && key != "total_damage" && key != "total_assists" && key != "total_deaths" && key != "total_downs" {
				statsToAdd[key] = val
			}
		}

		// Increment stats
		if err := s.playerStatsRepo.IncrementStats(ctx, stats.ID, statsToAdd); err != nil {
			return fmt.Errorf("increment player stats: %w", err)
		}

		// Recalculate KD ratio and ranking
		if err := recalculatePlayerRanking(ctx, ps.PlayerID, m.GameID, s.playerStatsRepo, s.ranking); err != nil {
			return fmt.Errorf("recalculate ranking: %w", err)
		}
	}

	return nil
}

// recalculatePlayerRanking updates player ranking after stats change.
func recalculatePlayerRanking(ctx context.Context, playerID, gameID uuid.UUID, statsRepo playerdomain.StatsRepository, ranking *rankingdomain.Service) error {
	stats, err := statsRepo.GetByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("get stats: %w", err)
	}

	// You would get the game here and recalculate
	// For now, this is a placeholder for the ranking recalculation logic
	_ = stats
	_ = ranking

	return nil
}

// Helper functions

func matchToResponse(m *matchdomain.Match) *MatchResponse {
	resp := &MatchResponse{
		ID:              m.ID,
		TournamentID:    m.TournamentID,
		TeamID:          m.TeamID,
		GameID:          m.GameID,
		Status:          string(m.Status),
		TeamPlacement:   m.TeamPlacement,
		TeamKills:       m.TeamKills,
		PlayerStats:     m.PlayerStats,
		ScreenshotURL:   m.ScreenshotURL,
		RejectionReason: m.RejectionReason,
		SubmittedBy:     m.SubmittedBy,
		CreatedAt:       m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if m.VerifiedAt != nil {
		verifiedAtStr := m.VerifiedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.VerifiedAt = &verifiedAtStr
	}

	resp.VerifiedBy = m.VerifiedBy

	return resp
}
