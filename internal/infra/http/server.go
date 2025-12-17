// Package http provides HTTP server and handlers for the API.
package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Server wraps the HTTP server with graceful shutdown support.
type Server struct {
	server *http.Server
	logger *slog.Logger
}

// NewServer creates a new HTTP server with the provided configuration.
func NewServer(addr string, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

// Start begins listening for HTTP requests.
// This method blocks until the server is shut down.
func (s *Server) Start() error {
	s.logger.Info("HTTP server starting", "addr", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the server with the given timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("HTTP server shutting down")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.server.Addr
}
