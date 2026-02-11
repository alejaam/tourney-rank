package leaderboard

import (
	"context"
	"fmt"

	"github.com/alejaam/tourney-rank/internal/domain/game"
	"github.com/alejaam/tourney-rank/internal/domain/player"
	"github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/google/uuid"
)

// LeaderboardEntry represents a single entry in the leaderboard response.
type LeaderboardEntry struct {
	Rank          int                    `json:"rank"`
	PlayerID      uuid.UUID              `json:"player_id"`
	DisplayName   string                 `json:"display_name"`
	AvatarURL     string                 `json:"avatar_url"`
	RankingScore  float64                `json:"ranking_score"`
	Tier          string                 `json:"tier"`
	MatchesPlayed int                    `json:"matches_played"`
	Stats         map[string]interface{} `json:"stats"`
}

// PlayerRankResponse represents a player's rank information.
type PlayerRankResponse struct {
	PlayerID     uuid.UUID `json:"player_id"`
	GameID       uuid.UUID `json:"game_id"`
	Rank         int64     `json:"rank"`
	RankingScore float64   `json:"ranking_score"`
	Tier         string    `json:"tier"`
	Percentile   float64   `json:"percentile"`
}

// TierDistribution represents the distribution of players across tiers.
type TierDistribution map[string]int64

// Service provides leaderboard operations.
type Service struct {
	statsRepo      player.StatsRepository
	gameRepo       game.Repository
	tournamentRepo tournament.Repository
}

// NewService creates a new leaderboard service.
func NewService(statsRepo player.StatsRepository, gameRepo game.Repository, tournamentRepo tournament.Repository) *Service {
	return &Service{
		statsRepo:      statsRepo,
		gameRepo:       gameRepo,
		tournamentRepo: tournamentRepo,
	}
}

// GetLeaderboard retrieves the leaderboard for a game.
func (s *Service) GetLeaderboard(ctx context.Context, gameID uuid.UUID, limit, offset int64) ([]LeaderboardEntry, string, int64, error) {
	// Validate game exists
	g, err := s.gameRepo.GetByID(ctx, gameID.String())
	if err != nil {
		if err == game.ErrNotFound {
			return nil, "", 0, fmt.Errorf("game not found")
		}
		return nil, "", 0, err
	}

	// Get leaderboard entries
	entries, err := s.statsRepo.GetLeaderboard(ctx, gameID, limit, offset)
	if err != nil {
		return nil, "", 0, err
	}

	// Convert domain entries to response DTOs
	response := make([]LeaderboardEntry, 0, len(entries))
	for _, entry := range entries {
		response = append(response, LeaderboardEntry{
			Rank:          entry.Rank,
			PlayerID:      entry.PlayerID,
			DisplayName:   entry.DisplayName,
			AvatarURL:     entry.AvatarURL,
			RankingScore:  entry.RankingScore,
			Tier:          string(entry.Tier),
			MatchesPlayed: entry.MatchesPlayed,
			Stats:         entry.Stats,
		})
	}

	// Get total count
	total, err := s.statsRepo.CountByGame(ctx, gameID)
	if err != nil {
		total = 0
	}

	return response, g.Name, total, nil
}

// GetLeaderboardByTier retrieves the leaderboard filtered by tier.
func (s *Service) GetLeaderboardByTier(ctx context.Context, gameID uuid.UUID, tierStr string, limit int64) ([]LeaderboardEntry, error) {
	// Validate tier
	tier := player.Tier(tierStr)
	if !isValidTier(tier) {
		return nil, fmt.Errorf("invalid tier: %s", tierStr)
	}

	// Get leaderboard entries by tier
	entries, err := s.statsRepo.GetLeaderboardByTier(ctx, gameID, tier, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	response := make([]LeaderboardEntry, 0, len(entries))
	for _, entry := range entries {
		response = append(response, LeaderboardEntry{
			Rank:          entry.Rank,
			PlayerID:      entry.PlayerID,
			DisplayName:   entry.DisplayName,
			AvatarURL:     entry.AvatarURL,
			RankingScore:  entry.RankingScore,
			Tier:          string(entry.Tier),
			MatchesPlayed: entry.MatchesPlayed,
			Stats:         entry.Stats,
		})
	}

	return response, nil
}

// GetPlayerRank retrieves a player's rank in a specific game.
func (s *Service) GetPlayerRank(ctx context.Context, playerID, gameID uuid.UUID) (*PlayerRankResponse, error) {
	// Get player rank info
	rankInfo, err := s.statsRepo.GetPlayerRank(ctx, playerID, gameID)
	if err != nil {
		if err == player.ErrStatsNotFound {
			return nil, fmt.Errorf("player has no stats for this game")
		}
		return nil, err
	}

	// Get total count for percentile
	total, err := s.statsRepo.CountByGame(ctx, gameID)
	if err != nil {
		total = 1
	}

	// Calculate percentile
	percentile := 0.0
	if total > 0 {
		percentile = float64(total-rankInfo.Rank+1) / float64(total) * 100
		if percentile < 0 {
			percentile = 0
		}
	}

	return &PlayerRankResponse{
		PlayerID:     playerID,
		GameID:       gameID,
		Rank:         rankInfo.Rank,
		RankingScore: rankInfo.RankingScore,
		Tier:         string(rankInfo.Tier),
		Percentile:   percentile,
	}, nil
}

// GetTierDistribution retrieves the distribution of players across tiers.
func (s *Service) GetTierDistribution(ctx context.Context, gameID uuid.UUID) (TierDistribution, int64, error) {
	distribution, err := s.statsRepo.GetTierDistribution(ctx, gameID)
	if err != nil {
		return nil, 0, err
	}

	// Convert to string keys
	response := make(TierDistribution)
	var total int64
	for tier, count := range distribution {
		response[string(tier)] = count
		total += count
	}

	return response, total, nil
}

// TournamentLeaderboardEntry represents a team's ranking in a tournament.
type TournamentLeaderboardEntry struct {
	Rank             int       `json:"rank"`
	TeamID           uuid.UUID `json:"team_id"`
	TeamName         string    `json:"team_name"`
	CaptainID        uuid.UUID `json:"captain_id"`
	CaptainName      string    `json:"captain_name"`
	TotalScore       float64   `json:"total_score"`
	MatchesPlayed    int       `json:"matches_played"`
	MatchesWon       int       `json:"matches_won"`
	TotalKills       int       `json:"total_kills"`
	AverageKills     float64   `json:"average_kills"`
	BestPlacement    int       `json:"best_placement"`
	WorstPlacement   int       `json:"worst_placement"`
	AveragePlacement float64   `json:"average_placement"`
}

// GetTournamentLeaderboard retrieves the cumulative leaderboard for an entire tournament.
// For now, this returns a placeholder since match-to-tournament linkage isn't fully implemented.
// In production, this should aggregate verified matches scoped to the tournament.
func (s *Service) GetTournamentLeaderboard(ctx context.Context, tournamentID uuid.UUID, limit, offset int) ([]TournamentLeaderboardEntry, int64, error) {
	// Verify tournament exists
	tourn, err := s.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, 0, fmt.Errorf("tournament not found: %w", err)
	}

	if tourn == nil {
		return nil, 0, fmt.Errorf("tournament not found")
	}

	// TODO: Implement proper tournament leaderboard aggregation
	// This requires matches to have a tournament_id field and be associated with rounds
	// For now, return empty leaderboard to avoid compile errors
	entries := make([]TournamentLeaderboardEntry, 0)

	return entries, 0, nil
}

// isValidTier checks if a tier str represents a valid Tier.
func isValidTier(tier player.Tier) bool {
	switch tier {
	case player.TierElite, player.TierAdvanced, player.TierIntermediate, player.TierBeginner:
		return true
	default:
		return false
	}
}
