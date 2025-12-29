// Package mongodb provides MongoDB connection management for the application.
package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	// DefaultConnectTimeout is the default timeout for establishing a connection.
	DefaultConnectTimeout = 10 * time.Second

	// DefaultPingTimeout is the default timeout for health checks.
	DefaultPingTimeout = 5 * time.Second

	// DefaultMaxRetries is the default number of connection retry attempts.
	DefaultMaxRetries = 3

	// DefaultRetryDelay is the default delay between retry attempts.
	DefaultRetryDelay = 2 * time.Second
)

// Client wraps the MongoDB client with additional functionality.
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	logger   *slog.Logger
	uri      string
	dbName   string
}

// Config holds the MongoDB connection configuration.
type Config struct {
	URI            string
	DatabaseName   string
	ConnectTimeout time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
}

// NewClient creates a new MongoDB client with the provided configuration.
func NewClient(ctx context.Context, cfg Config, logger *slog.Logger) (*Client, error) {
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = DefaultConnectTimeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = DefaultMaxRetries
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = DefaultRetryDelay
	}

	logger.Info("connecting to MongoDB",
		"database", cfg.DatabaseName,
		"timeout", cfg.ConnectTimeout,
	)

	// Configure client options
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetServerSelectionTimeout(cfg.ConnectTimeout).
		SetConnectTimeout(cfg.ConnectTimeout)

	// Connect with retry logic
	var client *mongo.Client
	var err error

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)

		client, err = mongo.Connect(connectCtx, clientOpts)
		cancel()

		if err == nil {
			// Verify connection with ping
			pingCtx, pingCancel := context.WithTimeout(ctx, DefaultPingTimeout)
			err = client.Ping(pingCtx, readpref.Primary())
			pingCancel()

			if err == nil {
				logger.Info("MongoDB connected successfully",
					"database", cfg.DatabaseName,
					"attempt", attempt,
				)
				break
			}
		}

		logger.Warn("MongoDB connection attempt failed",
			"attempt", attempt,
			"max_retries", cfg.MaxRetries,
			"error", err,
		)

		if attempt < cfg.MaxRetries {
			time.Sleep(cfg.RetryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB after %d attempts: %w", cfg.MaxRetries, err)
	}

	return &Client{
		client:   client,
		database: client.Database(cfg.DatabaseName),
		logger:   logger,
		uri:      cfg.URI,
		dbName:   cfg.DatabaseName,
	}, nil
}

// Database returns the configured database instance.
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection returns a collection from the configured database.
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Ping checks if the MongoDB connection is healthy.
func (c *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, DefaultPingTimeout)
	defer cancel()

	if err := c.client.Ping(pingCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("mongodb ping failed: %w", err)
	}

	return nil
}

// HealthCheck returns detailed health information about the MongoDB connection.
func (c *Client) HealthCheck(ctx context.Context) HealthStatus {
	start := time.Now()
	err := c.Ping(ctx)
	latency := time.Since(start)

	status := HealthStatus{
		Status:   "healthy",
		Latency:  latency.String(),
		Database: c.dbName,
	}

	if err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
	}

	return status
}

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status   string `json:"status"`
	Latency  string `json:"latency"`
	Database string `json:"database"`
	Error    string `json:"error,omitempty"`
}

// Close gracefully disconnects from MongoDB.
func (c *Client) Close(ctx context.Context) error {
	c.logger.Info("disconnecting from MongoDB")

	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("mongodb disconnect failed: %w", err)
	}

	c.logger.Info("MongoDB disconnected successfully")
	return nil
}

// RunInTransaction executes the given function within a MongoDB transaction.
// If the function returns an error, the transaction is aborted; otherwise, it's committed.
func (c *Client) RunInTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := c.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}
