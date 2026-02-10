package game

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the contract for Game persistence.
type Repository interface {
	Create(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id string) (*Game, error)
	GetBySlug(ctx context.Context, slug string) (*Game, error)
	GetAll(ctx context.Context) ([]*Game, error)
	List(ctx context.Context, activeOnly bool) ([]*Game, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, id string) error
}
