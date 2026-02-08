package tournament

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for tournament persistence operations.
type Repository interface {
	// Create stores a new tournament.
	Create(ctx context.Context, tournament *Tournament) error

	// GetByID retrieves a tournament by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Tournament, error)

	// Update updates an existing tournament.
	Update(ctx context.Context, tournament *Tournament) error

	// Delete removes a tournament by its ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves tournaments with optional filtering.
	List(ctx context.Context, filter ListFilter) ([]*Tournament, error)

	// GetByGameID retrieves all tournaments for a specific game.
	GetByGameID(ctx context.Context, gameID uuid.UUID) ([]*Tournament, error)

	// GetByStatus retrieves tournaments by status.
	GetByStatus(ctx context.Context, status Status) ([]*Tournament, error)

	// GetActiveTournaments retrieves all currently active tournaments.
	GetActiveTournaments(ctx context.Context) ([]*Tournament, error)

	// CountByGameID returns the number of tournaments for a game.
	CountByGameID(ctx context.Context, gameID uuid.UUID) (int64, error)
}

// ListFilter defines filtering options for listing tournaments.
type ListFilter struct {
	// GameID filters by game (optional).
	GameID *uuid.UUID

	// Status filters by tournament status (optional).
	Status *Status

	// CreatedBy filters by creator user ID (optional).
	CreatedBy *uuid.UUID

	// Limit is the maximum number of results to return.
	Limit int

	// Offset is the number of results to skip.
	Offset int
}
