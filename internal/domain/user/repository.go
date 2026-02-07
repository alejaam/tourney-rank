package user

import "context"

// Repository defines the contract for User persistence.
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	// Admin operations
	GetAll(ctx context.Context) ([]*User, error)
	Delete(ctx context.Context, id string) error
	UpdateRole(ctx context.Context, id string, role Role) error
}
