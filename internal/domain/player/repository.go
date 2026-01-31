package player

import "context"

// Repository defines the contract for Player persistence.
// Ref: [GO-ARCH-02] - Interface defined where consumed (consumer-side pattern)
type Repository interface {
	Create(ctx context.Context, player *Player) error
	GetByID(ctx context.Context, id string) (*Player, error)
	GetByUserID(ctx context.Context, userID string) (*Player, error)
	GetAll(ctx context.Context) ([]*Player, error)
	Update(ctx context.Context, player *Player) error
	Delete(ctx context.Context, id string) error
}
