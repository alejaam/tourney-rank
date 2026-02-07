package player

import (
	"context"

	"github.com/google/uuid"
)

// LeaderboardEntry represents a single entry in a leaderboard.
type LeaderboardEntry struct {
	Rank          int                    `json:"rank"`
	PlayerID      uuid.UUID              `json:"player_id"`
	DisplayName   string                 `json:"display_name"`
	AvatarURL     string                 `json:"avatar_url"`
	RankingScore  float64                `json:"ranking_score"`
	Tier          Tier                   `json:"tier"`
	MatchesPlayed int                    `json:"matches_played"`
	Stats         map[string]interface{} `json:"stats"`
}

// PlayerRankInfo contains rank information for a player.
type PlayerRankInfo struct {
	Rank         int64
	RankingScore float64
	Tier         Tier
}

// StatsRepository defines the contract for PlayerStats persistence.
type StatsRepository interface {
	Create(ctx context.Context, stats *PlayerStats) error
	GetByID(ctx context.Context, id uuid.UUID) (*PlayerStats, error)
	GetByPlayerAndGame(ctx context.Context, playerID, gameID uuid.UUID) (*PlayerStats, error)
	GetByPlayer(ctx context.Context, playerID uuid.UUID) ([]*PlayerStats, error)
	GetOrCreate(ctx context.Context, playerID, gameID uuid.UUID) (*PlayerStats, error)
	Update(ctx context.Context, stats *PlayerStats) error
	UpdateRanking(ctx context.Context, id uuid.UUID, score float64, tier Tier) error
	IncrementStats(ctx context.Context, id uuid.UUID, statsToAdd map[string]interface{}) error
	GetLeaderboard(ctx context.Context, gameID uuid.UUID, limit, offset int64) ([]LeaderboardEntry, error)
	GetLeaderboardByTier(ctx context.Context, gameID uuid.UUID, tier Tier, limit int64) ([]LeaderboardEntry, error)
	GetPlayerRank(ctx context.Context, playerID, gameID uuid.UUID) (*PlayerRankInfo, error)
	CountByGame(ctx context.Context, gameID uuid.UUID) (int64, error)
	GetTierDistribution(ctx context.Context, gameID uuid.UUID) (map[Tier]int64, error)
}
