package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"
	"time"
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

	// Root handler
	r.mux.HandleFunc("GET /", r.handleRoot)
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
