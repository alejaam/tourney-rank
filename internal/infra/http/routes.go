package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/melisource/tourney-rank/internal/infra/http/handlers"
	"github.com/melisource/tourney-rank/internal/infra/http/middleware"
)

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
	Uptime    string `json:"uptime,omitempty"`
}

// ReadyStatus represents the readiness check response.
type ReadyStatus struct {
	Status   string            `json:"status"`
	Checks   map[string]string `json:"checks,omitempty"`
	Database string            `json:"database,omitempty"`
	Redis    string            `json:"redis,omitempty"`
}

// SystemInfo represents system information for debug endpoints.
type SystemInfo struct {
	Version    string `json:"version"`
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	NumCPU     int    `json:"num_cpu"`
	Goroutines int    `json:"goroutines"`
}

// Router sets up HTTP routes for the application.
type Router struct {
	mux       *http.ServeMux
	logger    *slog.Logger
	startTime time.Time
	version   string

	// Dependencies for readiness checks (optional)
	mongoChecker func() error
	redisChecker func() error

	// API handlers
	gameHandler        *handlers.GameHandler
	leaderboardHandler *handlers.LeaderboardHandler
	authHandler        *handlers.AuthHandler
	adminHandler       *handlers.AdminHandler
	playerHandler      *handlers.PlayerHandler

	// JWT secret for auth middleware
	jwtSecret string
}

// RouterOption configures the router.
type RouterOption func(*Router)

// WithVersion sets the application version.
func WithVersion(version string) RouterOption {
	return func(r *Router) {
		r.version = version
	}
}

// WithMongoDBChecker sets the MongoDB health checker.
func WithMongoDBChecker(checker func(ctx context.Context) error) RouterOption {
	return func(r *Router) {
		r.mongoChecker = func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return checker(ctx)
		}
	}
}

// WithRedisChecker sets the Redis health checker.
func WithRedisChecker(checker func() error) RouterOption {
	return func(r *Router) {
		r.redisChecker = checker
	}
}

// WithGameHandler sets the game handler.
func WithGameHandler(h *handlers.GameHandler) RouterOption {
	return func(r *Router) {
		r.gameHandler = h
	}
}

// WithLeaderboardHandler sets the leaderboard handler.
func WithLeaderboardHandler(h *handlers.LeaderboardHandler) RouterOption {
	return func(r *Router) {
		r.leaderboardHandler = h
	}
}

// WithAuthHandler sets the auth handler.
func WithAuthHandler(h *handlers.AuthHandler) RouterOption {
	return func(r *Router) {
		r.authHandler = h
	}
}

// WithAdminHandler sets the admin handler.
func WithAdminHandler(h *handlers.AdminHandler) RouterOption {
	return func(r *Router) {
		r.adminHandler = h
	}
}

// WithJWTSecret sets the JWT secret for authentication.
func WithJWTSecret(secret string) RouterOption {
	return func(r *Router) {
		r.jwtSecret = secret
	}
}

// WithPlayerHandler sets the player handler.
func WithPlayerHandler(h *handlers.PlayerHandler) RouterOption {
	return func(r *Router) {
		r.playerHandler = h
	}
}

// NewRouter creates a new HTTP router with all routes configured.
func NewRouter(logger *slog.Logger, opts ...RouterOption) *Router {
	r := &Router{
		mux:       http.NewServeMux(),
		logger:    logger,
		startTime: time.Now(),
		version:   "dev",
	}

	for _, opt := range opts {
		opt(r)
	}

	r.setupRoutes()
	return r
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// setupRoutes configures all HTTP routes.
func (r *Router) setupRoutes() {
	// Health and readiness endpoints (no middleware)
	r.mux.HandleFunc("GET /healthz", r.handleHealth)
	r.mux.HandleFunc("GET /readyz", r.handleReady)
	r.mux.HandleFunc("GET /livez", r.handleHealth) // Alias for Kubernetes

	// System info (development only in production)
	r.mux.HandleFunc("GET /debug/info", r.handleSystemInfo)

	// API routes with middleware
	r.mux.HandleFunc("GET /api/ping", r.withMiddleware(r.handlePing))
	r.mux.HandleFunc("GET /api/v1/ping", r.withMiddleware(r.handlePing))

	// Auth API routes
	if r.authHandler != nil {
		r.mux.HandleFunc("POST /api/v1/auth/register", r.withMiddleware(r.authHandler.Register))
		r.mux.HandleFunc("POST /api/v1/auth/login", r.withMiddleware(r.authHandler.Login))
	}

	// Game API routes
	if r.gameHandler != nil {
		r.mux.HandleFunc("GET /api/v1/games", r.withMiddleware(r.gameHandler.List))
		r.mux.HandleFunc("POST /api/v1/games", r.withMiddleware(r.gameHandler.Create))
		r.mux.HandleFunc("GET /api/v1/games/{id}", r.withMiddleware(r.gameHandler.GetByID))
		r.mux.HandleFunc("PATCH /api/v1/games/{id}/status", r.withMiddleware(r.gameHandler.UpdateStatus))
		r.mux.HandleFunc("DELETE /api/v1/games/{id}", r.withMiddleware(r.gameHandler.Delete))
	}

	// Leaderboard API routes
	if r.leaderboardHandler != nil {
		r.mux.HandleFunc("GET /api/v1/leaderboard/{gameId}", r.withMiddleware(r.leaderboardHandler.GetLeaderboard))
		r.mux.HandleFunc("GET /api/v1/leaderboard/{gameId}/tier/{tier}", r.withMiddleware(r.leaderboardHandler.GetLeaderboardByTier))
		r.mux.HandleFunc("GET /api/v1/leaderboard/{gameId}/player/{playerId}", r.withMiddleware(r.leaderboardHandler.GetPlayerRank))
		r.mux.HandleFunc("GET /api/v1/leaderboard/{gameId}/tiers", r.withMiddleware(r.leaderboardHandler.GetTierDistribution))
	}

	// Player API routes (protected by auth middleware only)
	if r.playerHandler != nil && r.jwtSecret != "" {
		r.setupPlayerRoutes()
	}

	// Admin API routes (protected by auth + admin middleware)
	if r.adminHandler != nil && r.jwtSecret != "" {
		r.setupAdminRoutes()
	}

	// Root handler
	r.mux.HandleFunc("GET /", r.handleRoot)
}

// setupPlayerRoutes configures player routes with authentication (no admin check).
func (r *Router) setupPlayerRoutes() {
	authMw := r.createAuthMiddleware()

	// Player profile endpoints
	r.mux.Handle("GET /api/v1/players/me", r.withMiddlewareHandler(authMw(http.HandlerFunc(r.playerHandler.GetMyProfile))))
	r.mux.Handle("POST /api/v1/players/me", r.withMiddlewareHandler(authMw(http.HandlerFunc(r.playerHandler.CreateMyProfile))))
	r.mux.Handle("PUT /api/v1/players/me", r.withMiddlewareHandler(authMw(http.HandlerFunc(r.playerHandler.UpdateMyProfile))))

	// Player stats endpoints
	r.mux.Handle("GET /api/v1/players/me/stats", r.withMiddlewareHandler(authMw(http.HandlerFunc(r.playerHandler.GetMyStats))))
	r.mux.Handle("GET /api/v1/players/me/stats/{gameId}", r.withMiddlewareHandler(authMw(http.HandlerFunc(r.playerHandler.GetMyGameStats))))
}

// setupAdminRoutes configures admin-only routes with authentication.
func (r *Router) setupAdminRoutes() {
	// Import middleware package
	mw := r.getMiddleware()

	// User management
	r.mux.Handle("GET /api/v1/admin/users", mw(http.HandlerFunc(r.adminHandler.ListUsers)))
	r.mux.Handle("GET /api/v1/admin/users/{id}", mw(http.HandlerFunc(r.adminHandler.GetUser)))
	r.mux.Handle("DELETE /api/v1/admin/users/{id}", mw(http.HandlerFunc(r.adminHandler.DeleteUser)))
	r.mux.Handle("PATCH /api/v1/admin/users/{id}/role", mw(http.HandlerFunc(r.adminHandler.UpdateUserRole)))

	// Game management
	r.mux.Handle("GET /api/v1/admin/games", mw(http.HandlerFunc(r.adminHandler.ListGames)))
	r.mux.Handle("GET /api/v1/admin/games/{id}", mw(http.HandlerFunc(r.adminHandler.GetGame)))
	r.mux.Handle("POST /api/v1/admin/games", mw(http.HandlerFunc(r.adminHandler.CreateGame)))
	r.mux.Handle("PUT /api/v1/admin/games/{id}", mw(http.HandlerFunc(r.adminHandler.UpdateGame)))
	r.mux.Handle("DELETE /api/v1/admin/games/{id}", mw(http.HandlerFunc(r.adminHandler.DeleteGame)))

	// Player management
	r.mux.Handle("GET /api/v1/admin/players", mw(http.HandlerFunc(r.adminHandler.ListPlayers)))
	r.mux.Handle("GET /api/v1/admin/players/{id}", mw(http.HandlerFunc(r.adminHandler.GetPlayer)))
	r.mux.Handle("POST /api/v1/admin/players", mw(http.HandlerFunc(r.adminHandler.CreatePlayer)))
	r.mux.Handle("PATCH /api/v1/admin/players/{id}/ban", mw(http.HandlerFunc(r.adminHandler.BanPlayer)))
	r.mux.Handle("PATCH /api/v1/admin/players/{id}/unban", mw(http.HandlerFunc(r.adminHandler.UnbanPlayer)))
	r.mux.Handle("PUT /api/v1/admin/players/{id}", mw(http.HandlerFunc(r.adminHandler.UpdatePlayer)))
	r.mux.Handle("DELETE /api/v1/admin/players/{id}", mw(http.HandlerFunc(r.adminHandler.DeletePlayer)))
}

// getMiddleware returns a middleware chain that applies auth + admin + logging.
func (r *Router) getMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Import the middleware package here to avoid circular dependency
		authMw := r.createAuthMiddleware()
		adminMw := r.createAdminMiddleware()

		// Chain: logging -> recovery -> auth -> admin -> handler
		return r.withMiddlewareHandler(authMw(adminMw(next)))
	}
}

// withMiddleware wraps a handler with logging and recovery middleware.
func (r *Router) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Recovery
		defer func() {
			if err := recover(); err != nil {
				r.logger.Error("panic recovered", "error", err, "path", req.URL.Path)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Logging
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(wrapped, req)
		r.logger.Info("request",
			"method", req.Method,
			"path", req.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start),
			"remote_addr", req.RemoteAddr,
		)
	}
}

// withMiddlewareHandler wraps an http.Handler with logging and recovery middleware.
func (r *Router) withMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Recovery
		defer func() {
			if err := recover(); err != nil {
				r.logger.Error("panic recovered", "error", err, "path", req.URL.Path)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Logging
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, req)
		r.logger.Info("request",
			"method", req.Method,
			"path", req.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start),
			"remote_addr", req.RemoteAddr,
		)
	})
}

// createAuthMiddleware creates the auth middleware.
func (r *Router) createAuthMiddleware() func(http.Handler) http.Handler {
	return middleware.Auth(r.jwtSecret, r.logger)
}

// createAdminMiddleware creates the admin-only middleware.
func (r *Router) createAdminMiddleware() func(http.Handler) http.Handler {
	return middleware.AdminOnly(r.logger)
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleRoot handles the root endpoint.
func (r *Router) handleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	r.jsonResponse(w, http.StatusOK, map[string]string{
		"service": "TourneyRank API",
		"version": r.version,
		"docs":    "/api/v1",
	})
}

// handleHealth handles the health check endpoint.
// This is a liveness probe - returns 200 if the process is alive.
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   r.version,
		Uptime:    time.Since(r.startTime).Round(time.Second).String(),
	}

	r.jsonResponse(w, http.StatusOK, status)
}

// handleReady handles the readiness check endpoint.
// This is a readiness probe - returns 200 if the service can accept traffic.
func (r *Router) handleReady(w http.ResponseWriter, req *http.Request) {
	status := ReadyStatus{
		Status: "ok",
		Checks: make(map[string]string),
	}
	allHealthy := true

	// Check MongoDB if checker is configured
	if r.mongoChecker != nil {
		if err := r.mongoChecker(); err != nil {
			status.Database = "unhealthy: " + err.Error()
			status.Checks["mongodb"] = "fail"
			allHealthy = false
		} else {
			status.Database = "healthy"
			status.Checks["mongodb"] = "pass"
		}
	} else {
		status.Database = "not configured"
		status.Checks["mongodb"] = "skip"
	}

	// Check Redis if checker is configured
	if r.redisChecker != nil {
		if err := r.redisChecker(); err != nil {
			status.Redis = "unhealthy: " + err.Error()
			status.Checks["redis"] = "fail"
			allHealthy = false
		} else {
			status.Redis = "healthy"
			status.Checks["redis"] = "pass"
		}
	} else {
		status.Redis = "not configured"
		status.Checks["redis"] = "skip"
	}

	if !allHealthy {
		status.Status = "degraded"
		r.jsonResponse(w, http.StatusServiceUnavailable, status)
		return
	}

	r.jsonResponse(w, http.StatusOK, status)
}

// handleSystemInfo returns system information.
func (r *Router) handleSystemInfo(w http.ResponseWriter, req *http.Request) {
	info := SystemInfo{
		Version:    r.version,
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		NumCPU:     runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
	}

	r.jsonResponse(w, http.StatusOK, info)
}

// handlePing is a simple ping endpoint.
func (r *Router) handlePing(w http.ResponseWriter, req *http.Request) {
	r.jsonResponse(w, http.StatusOK, map[string]string{
		"message": "pong",
	})
}

// jsonResponse writes a JSON response with the given status code.
func (r *Router) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		r.logger.Error("failed to encode JSON response", "error", err)
	}
}
