package admin

import (
	"context"
	"fmt"

	"github.com/alejaam/tourney-rank/internal/domain/game"
)

// GameService provides admin operations for game management.
type GameService struct {
	gameRepo game.Repository
}

// NewGameService creates a new GameService.
func NewGameService(gameRepo game.Repository) *GameService {
	return &GameService{
		gameRepo: gameRepo,
	}
}

// CreateGameRequest represents the data needed to create a game.
type CreateGameRequest struct {
	Name             string              `json:"name"`
	Slug             string              `json:"slug"`
	Description      string              `json:"description"`
	PlatformIDFormat string              `json:"platform_id_format"`
	StatSchema       game.StatSchema     `json:"stat_schema"`
	RankingWeights   game.RankingWeights `json:"ranking_weights"`
}

// UpdateGameRequest represents the data needed to update a game.
type UpdateGameRequest struct {
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	PlatformIDFormat string              `json:"platform_id_format"`
	StatSchema       game.StatSchema     `json:"stat_schema"`
	RankingWeights   game.RankingWeights `json:"ranking_weights"`
	IsActive         bool                `json:"is_active"`
}

// ListGamesResponse contains the list of games.
type ListGamesResponse struct {
	Games []*game.Game `json:"games"`
	Total int          `json:"total"`
}

// CreateGame creates a new game.
func (s *GameService) CreateGame(ctx context.Context, req CreateGameRequest) (*game.Game, error) {
	g, err := game.NewGame(
		req.Name,
		req.Slug,
		req.Description,
		req.PlatformIDFormat,
		req.StatSchema,
		req.RankingWeights,
	)
	if err != nil {
		return nil, fmt.Errorf("creating game entity: %w", err)
	}

	if err := s.gameRepo.Create(ctx, g); err != nil {
		return nil, fmt.Errorf("saving game: %w", err)
	}

	return g, nil
}

// ListGames retrieves all games.
func (s *GameService) ListGames(ctx context.Context) (*ListGamesResponse, error) {
	games, err := s.gameRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing games: %w", err)
	}

	return &ListGamesResponse{
		Games: games,
		Total: len(games),
	}, nil
}

// GetGame retrieves a game by ID.
func (s *GameService) GetGame(ctx context.Context, id string) (*game.Game, error) {
	g, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting game: %w", err)
	}
	return g, nil
}

// UpdateGame updates an existing game.
func (s *GameService) UpdateGame(ctx context.Context, id string, req UpdateGameRequest) (*game.Game, error) {
	// Get existing game
	g, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting game: %w", err)
	}

	// Update fields
	g.Name = req.Name
	g.Description = req.Description
	g.PlatformIDFormat = req.PlatformIDFormat
	g.StatSchema = req.StatSchema
	g.RankingWeights = req.RankingWeights
	g.IsActive = req.IsActive

	if err := s.gameRepo.Update(ctx, g); err != nil {
		return nil, fmt.Errorf("updating game: %w", err)
	}

	return g, nil
}

// DeleteGame removes a game by ID.
func (s *GameService) DeleteGame(ctx context.Context, id string) error {
	if err := s.gameRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting game: %w", err)
	}
	return nil
}
