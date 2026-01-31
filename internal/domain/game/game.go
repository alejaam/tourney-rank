// Package game provides domain entities and logic for supported competitive games.
package game

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidGameName is returned when game name is empty or invalid.
	ErrInvalidGameName = errors.New("game name cannot be empty")

	// ErrInvalidSlug is returned when slug is empty or invalid.
	ErrInvalidSlug = errors.New("game slug cannot be empty")

	// ErrInvalidStatSchema is returned when stat schema is malformed.
	ErrInvalidStatSchema = errors.New("stat schema must be valid")

	// ErrInvalidRankingWeights is returned when ranking weights don't sum to 1.0.
	ErrInvalidRankingWeights = errors.New("ranking weights must sum to 1.0")
)

// Game represents a competitive game supported by the platform.
// Each game has its own stat schema and ranking weights.
type Game struct {
	ID               uuid.UUID
	Name             string
	Slug             string
	Description      string
	StatSchema       StatSchema
	RankingWeights   RankingWeights
	PlatformIDFormat string
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// StatSchema defines the available statistics for a game.
// It's a flexible structure that allows different games to have different metrics.
type StatSchema map[string]StatField

// StatField defines a single statistic field with validation rules.
type StatField struct {
	Type  string      `json:"type"`  // integer, float, string
	Min   interface{} `json:"min"`   // minimum value (optional)
	Max   interface{} `json:"max"`   // maximum value (optional)
	Label string      `json:"label"` // human-readable label
}

// RankingWeights defines how different metrics are weighted for ranking calculation.
// The sum of all weights must equal 1.0.
type RankingWeights map[string]float64

// NewGame creates a new Game instance with validation.
func NewGame(name, slug, description, platformIDFormat string, schema StatSchema, weights RankingWeights) (*Game, error) {
	if name == "" {
		return nil, ErrInvalidGameName
	}

	if slug == "" {
		return nil, ErrInvalidSlug
	}

	if err := validateRankingWeights(weights); err != nil {
		return nil, err
	}

	return &Game{
		ID:               uuid.New(),
		Name:             name,
		Slug:             slug,
		Description:      description,
		StatSchema:       schema,
		RankingWeights:   weights,
		PlatformIDFormat: platformIDFormat,
		IsActive:         true,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, nil
}

// Activate activates the game.
func (g *Game) Activate() {
	g.IsActive = true
	g.UpdatedAt = time.Now()
}

// Deactivate deactivates the game.
func (g *Game) Deactivate() {
	g.IsActive = false
	g.UpdatedAt = time.Now()
}

// UpdateWeights updates the ranking weights after validation.
func (g *Game) UpdateWeights(weights RankingWeights) error {
	if err := validateRankingWeights(weights); err != nil {
		return err
	}

	g.RankingWeights = weights
	g.UpdatedAt = time.Now()
	return nil
}

// ValidateStat checks if a stat value is valid according to the schema.
func (g *Game) ValidateStat(statName string, value interface{}) error {
	field, exists := g.StatSchema[statName]
	if !exists {
		return nil // Unknown stats are allowed for flexibility
	}

	// Type validation would go here
	// For now, we trust the input and rely on DB constraints
	_ = field

	return nil
}

// validateRankingWeights ensures weights sum to 1.0 with tolerance for floating point.
func validateRankingWeights(weights RankingWeights) error {
	if len(weights) == 0 {
		return ErrInvalidRankingWeights
	}

	var sum float64
	for _, w := range weights {
		sum += w
	}

	const tolerance = 0.001
	if sum < 1.0-tolerance || sum > 1.0+tolerance {
		return ErrInvalidRankingWeights
	}

	return nil
}
