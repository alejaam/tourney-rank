// Package game provides domain entities and logic for supported competitive games.
package game

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		gameName      string
		slug          string
		description   string
		platformID    string
		schema        StatSchema
		weights       RankingWeights
		expectedError error
	}{
		{
			name:        "valid game creation",
			gameName:    "Call of Duty: Warzone",
			slug:        "warzone",
			description: "Battle Royale",
			platformID:  "activision_id",
			schema: StatSchema{
				"kills":  StatField{Type: "integer", Min: 0, Label: "Kills"},
				"deaths": StatField{Type: "integer", Min: 0, Label: "Deaths"},
			},
			weights: RankingWeights{
				"kd_ratio":  0.5,
				"avg_kills": 0.5,
			},
			expectedError: nil,
		},
		{
			name:          "empty name",
			gameName:      "",
			slug:          "warzone",
			schema:        StatSchema{},
			weights:       RankingWeights{"kd": 1.0},
			expectedError: ErrInvalidGameName,
		},
		{
			name:          "empty slug",
			gameName:      "Warzone",
			slug:          "",
			schema:        StatSchema{},
			weights:       RankingWeights{"kd": 1.0},
			expectedError: ErrInvalidSlug,
		},
		{
			name:     "invalid weights sum",
			gameName: "Warzone",
			slug:     "warzone",
			schema:   StatSchema{},
			weights: RankingWeights{
				"kd":  0.5,
				"avg": 0.3,
			},
			expectedError: ErrInvalidRankingWeights,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game, err := NewGame(
				tc.gameName,
				tc.slug,
				tc.description,
				tc.platformID,
				tc.schema,
				tc.weights,
			)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tc.expectedError))
				require.Nil(t, game)
			} else {
				require.NoError(t, err)
				require.NotNil(t, game)
				require.Equal(t, tc.gameName, game.Name)
				require.Equal(t, tc.slug, game.Slug)
				require.True(t, game.IsActive)
				require.NotEqual(t, time.Time{}, game.CreatedAt)
			}
		})
	}
}

func TestGame_ActivateDeactivate(t *testing.T) {
	t.Parallel()

	game, err := NewGame(
		"Test Game",
		"test",
		"",
		"",
		StatSchema{},
		RankingWeights{"kd": 1.0},
	)
	require.NoError(t, err)
	require.True(t, game.IsActive)

	game.Deactivate()
	require.False(t, game.IsActive)

	game.Activate()
	require.True(t, game.IsActive)
}

func TestGame_UpdateWeights(t *testing.T) {
	t.Parallel()

	game, err := NewGame(
		"Test Game",
		"test",
		"",
		"",
		StatSchema{},
		RankingWeights{"kd": 1.0},
	)
	require.NoError(t, err)

	newWeights := RankingWeights{
		"kd":  0.6,
		"avg": 0.4,
	}

	err = game.UpdateWeights(newWeights)
	require.NoError(t, err)
	require.Equal(t, newWeights, game.RankingWeights)

	invalidWeights := RankingWeights{"kd": 0.5}
	err = game.UpdateWeights(invalidWeights)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidRankingWeights))
}

func TestValidateRankingWeights(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		weights       RankingWeights
		expectedError bool
	}{
		{
			name:          "valid weights sum to 1.0",
			weights:       RankingWeights{"a": 0.5, "b": 0.5},
			expectedError: false,
		},
		{
			name:          "valid weights with tolerance",
			weights:       RankingWeights{"a": 0.333, "b": 0.333, "c": 0.334},
			expectedError: false,
		},
		{
			name:          "invalid weights sum less than 1.0",
			weights:       RankingWeights{"a": 0.3, "b": 0.3},
			expectedError: true,
		},
		{
			name:          "invalid weights sum more than 1.0",
			weights:       RankingWeights{"a": 0.7, "b": 0.7},
			expectedError: true,
		},
		{
			name:          "empty weights",
			weights:       RankingWeights{},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRankingWeights(tc.weights)

			if tc.expectedError {
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrInvalidRankingWeights))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
