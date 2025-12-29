// Package main is the entry point for the TourneyRank application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/melisource/tourney-rank/internal/config"
	httpserver "github.com/melisource/tourney-rank/internal/infra/http"
	"github.com/melisource/tourney-rank/internal/infra/mongodb"
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

	// Initialize MongoDB connection
	mongoClient, err := mongodb.NewClient(ctx, mongodb.Config{
		URI:          cfg.MongoDBURI,
		DatabaseName: cfg.MongoDBDatabase,
	}, logger)
	if err != nil {
		return fmt.Errorf("connect to mongodb: %w", err)
	}
	defer mongoClient.Close(ctx)

	// TODO: Initialize Redis cache when needed
	// cache, err := redis.Connect(ctx, cfg.RedisURL)
	// if err != nil {
	//     return fmt.Errorf("connect to redis: %w", err)
	// }
	// defer cache.Close()

	// Setup HTTP router with options
	routerOpts := []httpserver.RouterOption{
		httpserver.WithVersion(Version),
		httpserver.WithMongoDBChecker(mongoClient.Ping),
	}

	// Add health checkers if dependencies are configured
	// if cache != nil {
	//     routerOpts = append(routerOpts, httpserver.WithRedisChecker(cache.Ping))
	// }

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
