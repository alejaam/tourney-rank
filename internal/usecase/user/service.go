package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/melisource/tourney-rank/internal/domain/user"
)

// Service provides user operations for regular users.
type Service struct {
	userRepo user.Repository
}

// NewService creates a new user service.
func NewService(userRepo user.Repository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// GetMe retrieves the user information for the authenticated user.
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID.String())
}
