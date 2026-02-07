package player

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/melisource/tourney-rank/internal/domain/player"
)

// Service provides player operations for regular users.
type Service struct {
	playerRepo player.Repository
}

// NewService creates a new player service.
func NewService(playerRepo player.Repository) *Service {
	return &Service{
		playerRepo: playerRepo,
	}
}

// UpdateProfileRequest represents the data needed to update a player profile.
type UpdateProfileRequest struct {
	DisplayName       string            `json:"display_name"`
	AvatarURL         string            `json:"avatar_url,omitempty"`
	Bio               string            `json:"bio,omitempty"`
	PlatformIDs       map[string]string `json:"platform_ids,omitempty"`
	BirthYear         int               `json:"birth_year,omitempty"`
	Region            string            `json:"region,omitempty"`
	PreferredPlatform string            `json:"preferred_platform,omitempty"`
	Language          string            `json:"language,omitempty"`
}

// CreateProfileRequest represents the data needed to create a player profile.
type CreateProfileRequest struct {
	DisplayName       string            `json:"display_name"`
	PreferredPlatform string            `json:"preferred_platform"`
	AvatarURL         string            `json:"avatar_url,omitempty"`
	Bio               string            `json:"bio,omitempty"`
	PlatformIDs       map[string]string `json:"platform_ids,omitempty"`
	BirthYear         int               `json:"birth_year,omitempty"`
	Region            string            `json:"region,omitempty"`
	Language          string            `json:"language,omitempty"`
}

// GetOrCreateByUserID gets a player by user ID, creating one if it doesn't exist.
func (s *Service) GetOrCreateByUserID(ctx context.Context, userID uuid.UUID, defaultDisplayName string) (*player.Player, error) {
	// Try to get existing player
	p, err := s.playerRepo.GetByUserID(ctx, userID.String())
	if err == nil {
		return p, nil
	}

	// If not found, create new player
	p, err = player.NewPlayer(userID, defaultDisplayName)
	if err != nil {
		return nil, err
	}

	if err := s.playerRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetMyProfile gets the player profile for the authenticated user.
func (s *Service) GetMyProfile(ctx context.Context, userID uuid.UUID) (*player.Player, error) {
	p, err := s.playerRepo.GetByUserID(ctx, userID.String())
	if err != nil {
		return nil, err
	}
	return p, nil
}

// UpdateMyProfile updates the player profile for the authenticated user.
func (s *Service) UpdateMyProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*player.Player, error) {
	// Get existing player
	p, err := s.playerRepo.GetByUserID(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	// Update basic profile fields
	p.UpdateProfile(req.DisplayName, req.AvatarURL, req.Bio)

	// Update extended profile fields
	if err := p.UpdateExtendedProfile(req.BirthYear, req.Region, req.PreferredPlatform, req.Language); err != nil {
		return nil, err
	}

	// Update platform IDs if provided
	if req.PlatformIDs != nil {
		for platform, platformID := range req.PlatformIDs {
			p.SetPlatformID(platform, platformID)
		}
	}

	// Save to repository
	if err := s.playerRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// CreateProfile creates a player profile for the authenticated user.
func (s *Service) CreateProfile(ctx context.Context, userID uuid.UUID, req CreateProfileRequest) (*player.Player, error) {
	// Validate required fields
	if req.DisplayName == "" {
		return nil, player.ErrInvalidUsername
	}
	if req.PreferredPlatform == "" {
		return nil, player.ErrInvalidPlatform
	}

	// Check if player already exists
	existing, err := s.playerRepo.GetByUserID(ctx, userID.String())
	if err == nil && existing != nil {
		return nil, errors.New("player profile already exists")
	}

	// Create new player
	p, err := player.NewPlayer(userID, req.DisplayName)
	if err != nil {
		return nil, err
	}

	// Set basic optional fields
	p.AvatarURL = req.AvatarURL
	p.Bio = req.Bio

	// Set extended profile fields
	if err := p.UpdateExtendedProfile(req.BirthYear, req.Region, req.PreferredPlatform, req.Language); err != nil {
		return nil, err
	}

	// Set platform IDs if provided
	if req.PlatformIDs != nil {
		for platform, platformID := range req.PlatformIDs {
			p.SetPlatformID(platform, platformID)
		}
	}

	if err := s.playerRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
