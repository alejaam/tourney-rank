package admin

import (
	"context"
	"fmt"

	"github.com/alejaam/tourney-rank/internal/domain/user"
)

// UserService provides admin operations for user management.
type UserService struct {
	userRepo user.Repository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo user.Repository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// ListUsersResponse contains the list of users.
type ListUsersResponse struct {
	Users []*user.User `json:"users"`
	Total int          `json:"total"`
}

// UpdateRoleRequest represents the data needed to update a user's role.
type UpdateRoleRequest struct {
	Role user.Role `json:"role"`
}

// ListUsers retrieves all users.
func (s *UserService) ListUsers(ctx context.Context) (*ListUsersResponse, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	return &ListUsersResponse{
		Users: users,
		Total: len(users),
	}, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, id string) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return u, nil
}

// DeleteUser removes a user by ID.
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

// UpdateRole changes a user's role.
func (s *UserService) UpdateRole(ctx context.Context, id string, req UpdateRoleRequest) error {
	// Validate role
	if req.Role != user.RoleAdmin && req.Role != user.RoleUser {
		return fmt.Errorf("invalid role: %s", req.Role)
	}

	if err := s.userRepo.UpdateRole(ctx, id, req.Role); err != nil {
		return fmt.Errorf("updating user role: %w", err)
	}
	return nil
}
