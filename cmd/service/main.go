// Package main is the entry point for the TourneyRank application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alejaam/tourney-rank/internal/config"
	"github.com/alejaam/tourney-rank/internal/infra"
	httpserver "github.com/alejaam/tourney-rank/internal/infra/http"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Setup structured logger
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger.Info("TourneyRank starting",
		"version", Version,
		"environment", cfg.Environment,
		"http_port", cfg.HTTPPort,
	)

	// Create context that listens for shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize dependency container
	container, err := infra.NewContainer(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("initialize container: %w", err)
	}
	defer container.Close(ctx)

	// Setup HTTP router with dependency injection
	routerOpts := []httpserver.RouterOption{
		httpserver.WithAuthHandler(container.AuthHandler),
		httpserver.WithAdminHandler(container.AdminHandler),
		httpserver.WithPlayerHandler(container.PlayerHandler),
		httpserver.WithJWTSecret(cfg.JWTSecret),
		httpserver.WithVersion(Version),
		httpserver.WithMongoDBChecker(container.MongoClient.Ping),
		httpserver.WithGameHandler(container.GameHandler),
		httpserver.WithLeaderboardHandler(container.LeaderboardHandler),
		httpserver.WithTournamentHandler(container.TournamentHandler),
		httpserver.WithTeamHandler(container.TeamHandler),
		httpserver.WithMatchHandler(container.MatchHandler),
		httpserver.WithBracketHandler(container.BracketHandler),
	}

	router := httpserver.NewRouter(logger, routerOpts...)

	// Create and start HTTP server
	server := httpserver.NewServer(cfg.HTTPAddr(), router, logger)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return err
	case sig := <-sigChan:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
		return err
	}

	logger.Info("application stopped gracefully")
	return nil
}
