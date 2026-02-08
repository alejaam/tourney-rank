package bracket

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for bracket persistence.
type Repository interface {
	// Bracket operations
	Create(ctx context.Context, bracket *Bracket) error
	GetByID(ctx context.Context, id uuid.UUID) (*Bracket, error)
	GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*Bracket, error)
	Update(ctx context.Context, bracket *Bracket) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Matchup operations
	CreateMatchup(ctx context.Context, matchup *Matchup) error
	GetMatchup(ctx context.Context, id uuid.UUID) (*Matchup, error)
	GetMatchupsByBracket(ctx context.Context, bracketID uuid.UUID) ([]*Matchup, error)
	GetMatchupsByRound(ctx context.Context, bracketID uuid.UUID, round int) ([]*Matchup, error)
	UpdateMatchup(ctx context.Context, matchup *Matchup) error
	DeleteMatchup(ctx context.Context, id uuid.UUID) error
}
