// Package ranking provides ranking calculation strategies for different games.
package ranking

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/melisource/tourney-rank/internal/domain/game"
	"github.com/melisource/tourney-rank/internal/domain/player"
)

var (
	// ErrInvalidStats is returned when stats are invalid for ranking calculation.
	ErrInvalidStats = errors.New("invalid stats for ranking calculation")

	// ErrUnsupportedGame is returned when no calculator exists for a game.
	ErrUnsupportedGame = errors.New("unsupported game for ranking calculation")
)

// Calculator defines the interface for ranking calculation strategies.
// Each game can implement its own ranking algorithm.
type Calculator interface {
	// Calculate computes the ranking score for a player based on their stats.
	Calculate(ctx context.Context, stats *player.PlayerStats, game *game.Game) (float64, error)

	// SupportsGame returns true if this calculator can handle the given game.
	SupportsGame(gameSlug string) bool
}

// Service orchestrates ranking calculations using appropriate strategies.
type Service struct {
	calculators []Calculator
}

// NewService creates a new ranking service with registered calculators.
func NewService(calculators ...Calculator) *Service {
	return &Service{
		calculators: calculators,
	}
}

// CalculateRanking calculates ranking score and tier for a player in a specific game.
func (s *Service) CalculateRanking(ctx context.Context, stats *player.PlayerStats, game *game.Game) (float64, player.Tier, error) {
	calculator := s.findCalculator(game.Slug)
	if calculator == nil {
		return 0, player.TierBeginner, ErrUnsupportedGame
	}

	score, err := calculator.Calculate(ctx, stats, game)
	if err != nil {
		return 0, player.TierBeginner, err
	}

	// Tier determination would typically require comparing with other players
	// For now, use a simple score-based tier assignment
	tier := determineTierByScore(score)

	return score, tier, nil
}

// findCalculator finds the appropriate calculator for a game.
func (s *Service) findCalculator(gameSlug string) Calculator {
	for _, calc := range s.calculators {
		if calc.SupportsGame(gameSlug) {
			return calc
		}
	}
	return nil
}

// determineTierByScore determines tier based on absolute score.
// In a real scenario, this should use percentile ranking among all players.
func determineTierByScore(score float64) player.Tier {
	switch {
	case score >= 800:
		return player.TierElite
	case score >= 600:
		return player.TierAdvanced
	case score >= 400:
		return player.TierIntermediate
	default:
		return player.TierBeginner
	}
}

// UpdatePlayerRanking updates a player's ranking score and tier.
func UpdatePlayerRanking(ctx context.Context, stats *player.PlayerStats, score float64, tier player.Tier) error {
	return stats.UpdateRankingScore(score, tier)
}

// Repository defines the interface for ranking data access.
type Repository interface {
	// GetPlayerStats retrieves player stats for a specific game.
	GetPlayerStats(ctx context.Context, playerID, gameID uuid.UUID) (*player.PlayerStats, error)

	// UpdatePlayerStats updates player stats including ranking.
	UpdatePlayerStats(ctx context.Context, stats *player.PlayerStats) error

	// GetPlayersByGameRanking retrieves players sorted by ranking for a game.
	GetPlayersByGameRanking(ctx context.Context, gameID uuid.UUID, limit int) ([]*player.PlayerStats, error)

	// CalculatePercentile calculates the percentile rank of a player's score.
	CalculatePercentile(ctx context.Context, gameID uuid.UUID, score float64) (float64, error)
}
