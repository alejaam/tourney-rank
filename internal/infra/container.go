// Package infra provides infrastructure setup and dependency injection.
package infra

import (
	"context"
	"log/slog"
	"time"

	"github.com/alejaam/tourney-rank/internal/config"
	"github.com/alejaam/tourney-rank/internal/domain/bracket"
	"github.com/alejaam/tourney-rank/internal/domain/game"
	"github.com/alejaam/tourney-rank/internal/domain/match"
	"github.com/alejaam/tourney-rank/internal/domain/player"
	"github.com/alejaam/tourney-rank/internal/domain/team"
	"github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/alejaam/tourney-rank/internal/domain/user"
	"github.com/alejaam/tourney-rank/internal/infra/http/handlers"
	"github.com/alejaam/tourney-rank/internal/infra/mongodb"
	"github.com/alejaam/tourney-rank/internal/usecase/admin"
	"github.com/alejaam/tourney-rank/internal/usecase/auth"
	bracketusecase "github.com/alejaam/tourney-rank/internal/usecase/bracket"
	leaderboardusecase "github.com/alejaam/tourney-rank/internal/usecase/leaderboard"
	matchusecase "github.com/alejaam/tourney-rank/internal/usecase/match"
	playerusecase "github.com/alejaam/tourney-rank/internal/usecase/player"
	teamusecase "github.com/alejaam/tourney-rank/internal/usecase/team"
	tournamentusecase "github.com/alejaam/tourney-rank/internal/usecase/tournament"
	userusecase "github.com/alejaam/tourney-rank/internal/usecase/user"
)

// Container holds all application dependencies.
type Container struct {
	// Config
	Config *config.Config

	// Logger
	Logger *slog.Logger

	// MongoDB Client
	MongoClient *mongodb.Client

	// Repositories (concrete implementations)
	GameRepoConcrete        *mongodb.GameRepository
	PlayerRepoConcrete      *mongodb.PlayerRepository
	PlayerStatsRepoConcrete *mongodb.PlayerStatsRepository
	UserRepoConcrete        *mongodb.UserRepository
	TournamentRepoConcrete  *mongodb.TournamentRepository
	TeamRepoConcrete        *mongodb.TeamRepository
	MatchRepoConcrete       *mongodb.MatchRepository
	BracketRepoConcrete     *mongodb.BracketRepository

	// Repositories (interfaces)
	GameRepo        game.Repository
	PlayerRepo      player.Repository
	PlayerStatsRepo player.StatsRepository
	UserRepo        user.Repository
	TournamentRepo  tournament.Repository
	TeamRepo        team.Repository
	MatchRepo       match.Repository
	BracketRepo     bracket.Repository

	// Services - Auth
	AuthService *auth.Service

	// Services - User
	UserService *userusecase.Service

	// Services - Player
	PlayerService *playerusecase.Service

	// Services - Leaderboard
	LeaderboardService *leaderboardusecase.Service

	// Services - Tournament
	TournamentService *tournamentusecase.Service

	// Services - Team
	TeamService *teamusecase.Service

	// Services - Match
	MatchService *matchusecase.Service

	// Services - Bracket
	BracketService *bracketusecase.Service

	// Services - Admin
	AdminUserService   *admin.UserService
	AdminGameService   *admin.GameService
	AdminPlayerService *admin.PlayerService

	// Handlers
	AuthHandler        *handlers.AuthHandler
	AdminHandler       *handlers.AdminHandler
	PlayerHandler      *handlers.PlayerHandler
	GameHandler        *handlers.GameHandler
	LeaderboardHandler *handlers.LeaderboardHandler
	TournamentHandler  *handlers.TournamentHandler
	TeamHandler        *handlers.TeamHandler
	MatchHandler       *handlers.MatchHandler
	BracketHandler     *handlers.BracketHandler
}

// NewContainer creates and initializes all dependencies.
func NewContainer(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Container, error) {
	c := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Initialize MongoDB
	if err := c.initMongoDB(ctx); err != nil {
		return nil, err
	}

	// Initialize repositories
	if err := c.initRepositories(ctx); err != nil {
		return nil, err
	}

	// Initialize services
	if err := c.initServices(); err != nil {
		return nil, err
	}

	// Initialize handlers
	if err := c.initHandlers(); err != nil {
		return nil, err
	}

	return c, nil
}

// initMongoDB initializes the MongoDB client and connection.
func (c *Container) initMongoDB(ctx context.Context) error {
	client, err := mongodb.NewClient(ctx, mongodb.Config{
		URI:          c.Config.MongoDBURI,
		DatabaseName: c.Config.MongoDBDatabase,
	}, c.Logger)
	if err != nil {
		return err
	}

	c.MongoClient = client
	return nil
}

// initRepositories initializes all repository implementations.
func (c *Container) initRepositories(ctx context.Context) error {
	// Create repositories
	gameRepo := mongodb.NewGameRepository(c.MongoClient)
	playerRepo := mongodb.NewPlayerRepository(c.MongoClient)
	playerStatsRepo := mongodb.NewPlayerStatsRepository(c.MongoClient)
	userRepo := mongodb.NewUserRepository(c.MongoClient)
	tournamentRepo := mongodb.NewTournamentRepository(c.MongoClient.Database())
	teamRepo := mongodb.NewTeamRepository(c.MongoClient.Database())
	matchRepo := mongodb.NewMatchRepository(c.MongoClient.Database())
	bracketRepo := mongodb.NewBracketRepository(c.MongoClient.Database())

	// Store concrete implementations
	c.GameRepoConcrete = gameRepo
	c.PlayerRepoConcrete = playerRepo
	c.PlayerStatsRepoConcrete = playerStatsRepo
	c.UserRepoConcrete = userRepo
	c.TournamentRepoConcrete = tournamentRepo
	c.TeamRepoConcrete = teamRepo
	c.MatchRepoConcrete = matchRepo
	c.BracketRepoConcrete = bracketRepo

	// Store as interfaces
	c.GameRepo = gameRepo
	c.PlayerRepo = playerRepo
	c.PlayerStatsRepo = playerStatsRepo
	c.UserRepo = userRepo
	c.TournamentRepo = tournamentRepo
	c.TeamRepo = teamRepo
	c.MatchRepo = matchRepo
	c.BracketRepo = bracketRepo

	// Ensure indexes
	repos := []struct {
		name string
		repo interface {
			EnsureIndexes(ctx context.Context) error
		}
	}{
		{"game", gameRepo},
		{"player", playerRepo},
		{"user", userRepo},
		{"player_stats", playerStatsRepo},
		{"tournament", tournamentRepo},
		{"team", teamRepo},
		{"match", matchRepo},
		{"bracket", bracketRepo},
	}

	for _, r := range repos {
		if err := r.repo.EnsureIndexes(ctx); err != nil {
			c.Logger.Warn("failed to ensure indexes", "repo", r.name, "error", err)
		}
	}

	return nil
}

// initServices initializes all service implementations.
func (c *Container) initServices() error {
	// Auth service
	c.AuthService = auth.NewService(c.UserRepo, c.Config.JWTSecret, 24*time.Hour)

	// User service
	c.UserService = userusecase.NewService(c.UserRepo)

	// Player service
	c.PlayerService = playerusecase.NewService(c.PlayerRepo)

	// Leaderboard service
	c.LeaderboardService = leaderboardusecase.NewService(c.PlayerStatsRepo, c.GameRepo)

	// Tournament service
	c.TournamentService = tournamentusecase.NewService(c.TournamentRepo, c.TeamRepo, c.GameRepo)

	// Team service
	c.TeamService = teamusecase.NewService(c.TeamRepo, c.TournamentRepo, c.PlayerRepo)

	// Match service
	c.MatchService = matchusecase.NewService(
		c.MatchRepo,
		c.TeamRepo,
		c.TournamentRepo,
		c.PlayerRepo,
		c.PlayerStatsRepo,
		c.PlayerService,
		nil, // ranking service - TODO: initialize when created
	)

	// Bracket service
	c.BracketService = bracketusecase.NewService(c.BracketRepo, c.TeamRepo, c.TournamentRepo)

	// Admin services
	c.AdminUserService = admin.NewUserService(c.UserRepo)
	c.AdminGameService = admin.NewGameService(c.GameRepo)
	c.AdminPlayerService = admin.NewPlayerService(c.PlayerRepo)

	return nil
}

// initHandlers initializes all HTTP handlers.
func (c *Container) initHandlers() error {
	c.AuthHandler = handlers.NewAuthHandler(c.AuthService, c.UserService, c.Logger)
	c.AdminHandler = handlers.NewAdminHandler(c.AdminUserService, c.AdminGameService, c.AdminPlayerService, c.Logger)
	c.PlayerHandler = handlers.NewPlayerHandler(c.PlayerService, c.PlayerStatsRepo, c.GameRepo, c.Logger)
	c.GameHandler = handlers.NewGameHandler(c.GameRepo, c.Logger)
	c.LeaderboardHandler = handlers.NewLeaderboardHandler(c.LeaderboardService, c.Logger)
	c.TournamentHandler = handlers.NewTournamentHandler(c.TournamentService, c.PlayerService, c.Logger)
	c.TeamHandler = handlers.NewTeamHandler(c.TeamService, c.PlayerService, c.Logger)
	c.MatchHandler = handlers.NewMatchHandler(c.Logger, c.MatchService, c.PlayerService)
	c.BracketHandler = handlers.NewBracketHandler(c.BracketService, c.Logger)

	return nil
}

// Close closes all resources held by the container.
func (c *Container) Close(ctx context.Context) error {
	if c.MongoClient != nil {
		return c.MongoClient.Close(ctx)
	}
	return nil
}
