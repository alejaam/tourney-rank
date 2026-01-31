package game

import "context"

// Repository defines the contract for Game persistence.
// Ref: [GO-ARCH-02] - Interface defined where consumed (consumer-side pattern)
type Repository interface {
	Create(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id string) (*Game, error)
	GetBySlug(ctx context.Context, slug string) (*Game, error)
	GetAll(ctx context.Context) ([]*Game, error)
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, id string) error
}
