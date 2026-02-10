package round

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the contract for Round persistence.
type Repository interface {
	Create(ctx context.Context, round *Round) error
	GetByID(ctx context.Context, id uuid.UUID) (*Round, error)
	GetByTournamentAndNumber(ctx context.Context, tournamentID uuid.UUID, number int) (*Round, error)
	GetByTournament(ctx context.Context, tournamentID uuid.UUID) ([]*Round, error)
	Update(ctx context.Context, round *Round) error
	Delete(ctx context.Context, id uuid.UUID) error
}
