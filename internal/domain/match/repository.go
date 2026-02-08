package match

import (
	"context"
)

// Repository defines the interface for match persistence
type Repository interface {
	// Create stores a new match
	Create(ctx context.Context, match *Match) error

	// GetByID retrieves a match by ID
	GetByID(ctx context.Context, id string) (*Match, error)

	// GetByTournament retrieves all matches in a tournament with pagination
	GetByTournament(ctx context.Context, tournamentID string, limit int, offset int) ([]Match, error)

	// GetByTeam retrieves all matches for a specific team
	GetByTeam(ctx context.Context, teamID string, limit int, offset int) ([]Match, error)

	// GetByPlayer retrieves all matches involving a specific player
	GetByPlayer(ctx context.Context, playerID string, limit int, offset int) ([]Match, error)

	// GetUnverified retrieves all unverified (draft) matches for admin review
	GetUnverified(ctx context.Context, limit int, offset int) ([]Match, error)

	// GetTournamentUnverified retrieves unverified matches in a specific tournament
	GetTournamentUnverified(ctx context.Context, tournamentID string, limit int, offset int) ([]Match, error)

	// Update updates an existing match
	Update(ctx context.Context, match *Match) error

	// CountByTournament returns the total number of matches in a tournament
	CountByTournament(ctx context.Context, tournamentID string) (int, error)

	// CountUnverified returns total unverified matches
	CountUnverified(ctx context.Context) (int, error)

	// DeleteByID deletes a match (for testing purposes)
	DeleteByID(ctx context.Context, id string) error
}
