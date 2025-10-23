// Package ranking provides ranking calculation strategies for different games.
package ranking

import (
	"context"
	"math"

	"github.com/melisource/tourney-rank/internal/domain/game"
	"github.com/melisource/tourney-rank/internal/domain/player"
)

// WarzoneCalculator implements ranking calculation for Call of Duty: Warzone.
type WarzoneCalculator struct{}

// NewWarzoneCalculator creates a new Warzone ranking calculator.
func NewWarzoneCalculator() *WarzoneCalculator {
	return &WarzoneCalculator{}
}

// Calculate computes ranking score for Warzone based on game-specific weights.
// Formula: weighted sum of:
//   - K/D Ratio (default weight: 0.40)
//   - Average Kills per match (default weight: 0.30)
//   - Average Damage per match (default weight: 0.20)
//   - Consistency (1 - coefficient of variation) (default weight: 0.10)
func (wc *WarzoneCalculator) Calculate(ctx context.Context, stats *player.PlayerStats, g *game.Game) (float64, error) {
	if stats.MatchesPlayed == 0 {
		return 0, nil
	}

	// Extract stats
	kills := stats.GetStatAsFloat("total_kills")
	deaths := stats.GetStatAsFloat("total_deaths")
	damage := stats.GetStatAsFloat("total_damage")

	// Calculate K/D ratio
	var kdRatio float64
	if deaths > 0 {
		kdRatio = kills / deaths
	} else {
		kdRatio = kills
	}

	// Normalize K/D to 0-100 scale (cap at 5.0 K/D = 100 points)
	kdScore := math.Min(kdRatio*20, 100)

	// Calculate average kills per match
	avgKills := kills / float64(stats.MatchesPlayed)
	avgKillsScore := math.Min(avgKills*5, 100) // Cap at 20 kills = 100 points

	// Calculate average damage per match
	avgDamage := damage / float64(stats.MatchesPlayed)
	avgDamageScore := math.Min(avgDamage/30, 100) // Cap at 3000 damage = 100 points

	// Consistency: use coefficient of variation from kills
	// For now, simplified - would require match-by-match data
	// Assume 70 as baseline consistency
	consistencyScore := 70.0

	// Get weights from game configuration
	weights := g.RankingWeights

	kdWeight := getWeight(weights, "kd_ratio", 0.40)
	killsWeight := getWeight(weights, "avg_kills", 0.30)
	damageWeight := getWeight(weights, "avg_damage", 0.20)
	consistencyWeight := getWeight(weights, "consistency", 0.10)

	// Calculate weighted score
	score := (kdScore * kdWeight) +
		(avgKillsScore * killsWeight) +
		(avgDamageScore * damageWeight) +
		(consistencyScore * consistencyWeight)

	// Scale to 0-1000 range
	finalScore := score * 10

	return finalScore, nil
}

// SupportsGame returns true for Warzone.
func (wc *WarzoneCalculator) SupportsGame(gameSlug string) bool {
	return gameSlug == "warzone"
}

// getWeight retrieves weight from map with fallback to default.
func getWeight(weights game.RankingWeights, key string, defaultValue float64) float64 {
	if val, exists := weights[key]; exists {
		return val
	}
	return defaultValue
}

// DefaultCalculator is a generic calculator for games without specific strategy.
type DefaultCalculator struct{}

// NewDefaultCalculator creates a new default calculator.
func NewDefaultCalculator() *DefaultCalculator {
	return &DefaultCalculator{}
}

// Calculate uses a simple K/D-based ranking for generic games.
func (dc *DefaultCalculator) Calculate(ctx context.Context, stats *player.PlayerStats, g *game.Game) (float64, error) {
	if stats.MatchesPlayed == 0 {
		return 0, nil
	}

	kdRatio := stats.CalculateKDRatio()

	// Simple scoring: K/D * 100 + matches played as bonus
	score := (kdRatio * 100) + float64(stats.MatchesPlayed)

	return score, nil
}

// SupportsGame returns true for any game (fallback calculator).
func (dc *DefaultCalculator) SupportsGame(gameSlug string) bool {
	// Default calculator supports all games as fallback
	return true
}
