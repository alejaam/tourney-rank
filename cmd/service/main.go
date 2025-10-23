// Package main is the entry point for the TourneyRank application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// DefaultHTTPPort is the default port for HTTP server.
	DefaultHTTPPort = "8080"

	// DefaultWSPort is the default port for WebSocket server.
	DefaultWSPort = "8081"

	// ShutdownTimeout is the graceful shutdown timeout.
	ShutdownTimeout = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// TODO: Load configuration
	// cfg, err := config.Load()
	// if err != nil {
	//     return fmt.Errorf("load config: %w", err)
	// }

	// TODO: Initialize database connection
	// db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	// if err != nil {
	//     return fmt.Errorf("connect to database: %w", err)
	// }
	// defer db.Close()

	// TODO: Initialize Redis cache
	// cache, err := redis.Connect(ctx, cfg.RedisURL)
	// if err != nil {
	//     return fmt.Errorf("connect to redis: %w", err)
	// }
	// defer cache.Close()

	// TODO: Initialize repositories
	// gameRepo := postgres.NewGameRepository(db)
	// playerRepo := postgres.NewPlayerRepository(db)
	// tournamentRepo := postgres.NewTournamentRepository(db)

	// TODO: Initialize domain services
	// rankingService := ranking.NewService(
	//     ranking.NewWarzoneCalculator(),
	//     ranking.NewDefaultCalculator(),
	// )

	// TODO: Initialize application services
	// tournamentService := app.NewTournamentService(tournamentRepo, gameRepo)
	// statsService := app.NewStatsService(playerRepo, rankingService, cache)

	// TODO: Initialize HTTP server
	// httpServer := http.NewServer(cfg.HTTPPort, tournamentService, statsService)

	// TODO: Initialize WebSocket server
	// wsServer := websocket.NewServer(cfg.WSPort)

	log.Println("TourneyRank starting...")
	log.Printf("HTTP Server would start on port %s", getEnv("HTTP_PORT", DefaultHTTPPort))
	log.Printf("WebSocket Server would start on port %s", getEnv("WS_PORT", DefaultWSPort))

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received, gracefully stopping...")

	shutdownCtx, cancel := context.WithTimeout(ctx, ShutdownTimeout)
	defer cancel()

	// TODO: Shutdown servers gracefully
	// if err := httpServer.Shutdown(shutdownCtx); err != nil {
	//     return fmt.Errorf("http server shutdown: %w", err)
	// }

	// if err := wsServer.Shutdown(shutdownCtx); err != nil {
	//     return fmt.Errorf("websocket server shutdown: %w", err)
	// }

	_ = shutdownCtx

	log.Println("Application stopped gracefully")
	return nil
}

// getEnv retrieves an environment variable with a fallback default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// mustGetEnv retrieves an environment variable or panics if not set.
func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return value
}
