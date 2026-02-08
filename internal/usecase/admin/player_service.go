package admin

import (
	"context"
	"fmt"

	"github.com/alejaam/tourney-rank/internal/domain/player"
	"github.com/google/uuid"
)

// PlayerService provides admin operations for player management.
type PlayerService struct {
	playerRepo player.Repository
}

// NewPlayerService creates a new PlayerService.
func NewPlayerService(playerRepo player.Repository) *PlayerService {
	return &PlayerService{
		playerRepo: playerRepo,
	}
}

// CreatePlayerRequest represents the data needed to create a player.
type CreatePlayerRequest struct {
	UserID      uuid.UUID         `json:"user_id"`
	DisplayName string            `json:"display_name"`
	AvatarURL   string            `json:"avatar_url"`
	Bio         string            `json:"bio"`
	PlatformIDs map[string]string `json:"platform_ids"`
}

// UpdatePlayerRequest represents the data needed to update a player.
type UpdatePlayerRequest struct {
	DisplayName string            `json:"display_name"`
	AvatarURL   string            `json:"avatar_url"`
	Bio         string            `json:"bio"`
	PlatformIDs map[string]string `json:"platform_ids"`
}

// ListPlayersResponse contains the list of players.
type ListPlayersResponse struct {
	Players []*player.Player `json:"players"`
	Total   int              `json:"total"`
}

// CreatePlayer creates a new player.
func (s *PlayerService) CreatePlayer(ctx context.Context, req CreatePlayerRequest) (*player.Player, error) {
	p, err := player.NewPlayer(req.UserID, req.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("creating player entity: %w", err)
	}

	// Set optional fields
	p.AvatarURL = req.AvatarURL
	p.Bio = req.Bio
	if req.PlatformIDs != nil {
		p.PlatformIDs = req.PlatformIDs
	}

	if err := s.playerRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("saving player: %w", err)
	}

	return p, nil
}

// ListPlayers retrieves all players.
func (s *PlayerService) ListPlayers(ctx context.Context) (*ListPlayersResponse, error) {
	players, err := s.playerRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing players: %w", err)
	}

	return &ListPlayersResponse{
		Players: players,
		Total:   len(players),
	}, nil
}

// GetPlayer retrieves a player by ID.
func (s *PlayerService) GetPlayer(ctx context.Context, id string) (*player.Player, error) {
	p, err := s.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting player: %w", err)
	}
	return p, nil
}

// UpdatePlayer updates an existing player.
func (s *PlayerService) UpdatePlayer(ctx context.Context, id string, req UpdatePlayerRequest) (*player.Player, error) {
	// Get existing player
	p, err := s.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting player: %w", err)
	}

	// Update fields
	p.DisplayName = req.DisplayName
	p.AvatarURL = req.AvatarURL
	p.Bio = req.Bio
	if req.PlatformIDs != nil {
		p.PlatformIDs = req.PlatformIDs
	}

	if err := s.playerRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("updating player: %w", err)
	}

	return p, nil
}

// DeletePlayer removes a player by ID.
func (s *PlayerService) DeletePlayer(ctx context.Context, id string) error {
	if err := s.playerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting player: %w", err)
	}
	return nil
}

// BanPlayer marks a player as banned.
func (s *PlayerService) BanPlayer(ctx context.Context, id string) (*player.Player, error) {
	p, err := s.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting player: %w", err)
	}

	p.Ban()

	if err := s.playerRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("updating player: %w", err)
	}

	return p, nil
}

// UnbanPlayer removes the banned status from a player.
func (s *PlayerService) UnbanPlayer(ctx context.Context, id string) (*player.Player, error) {
	p, err := s.playerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting player: %w", err)
	}

	p.Unban()

	if err := s.playerRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("updating player: %w", err)
	}

	return p, nil
}
