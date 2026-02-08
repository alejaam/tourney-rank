package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TournamentRepository implements tournament.Repository using MongoDB.
type TournamentRepository struct {
	collection *mongo.Collection
}

// NewTournamentRepository creates a new MongoDB tournament repository.
func NewTournamentRepository(db *mongo.Database) *TournamentRepository {
	return &TournamentRepository{
		collection: db.Collection("tournaments"),
	}
}

// EnsureIndexes creates necessary indexes for the tournaments collection.
func (r *TournamentRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "game_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_by", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "start_date", Value: 1},
				{Key: "end_date", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("creating tournament indexes: %w", err)
	}

	return nil
}

// Create stores a new tournament.
func (r *TournamentRepository) Create(ctx context.Context, t *tournament.Tournament) error {
	_, err := r.collection.InsertOne(ctx, t)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("tournament already exists")
		}
		return fmt.Errorf("inserting tournament: %w", err)
	}
	return nil
}

// GetByID retrieves a tournament by its ID.
func (r *TournamentRepository) GetByID(ctx context.Context, id uuid.UUID) (*tournament.Tournament, error) {
	var t tournament.Tournament
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, tournament.ErrNotFound
		}
		return nil, fmt.Errorf("finding tournament: %w", err)
	}
	return &t, nil
}

// Update updates an existing tournament.
func (r *TournamentRepository) Update(ctx context.Context, t *tournament.Tournament) error {
	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": t.ID},
		t,
	)
	if err != nil {
		return fmt.Errorf("updating tournament: %w", err)
	}
	if result.MatchedCount == 0 {
		return tournament.ErrNotFound
	}
	return nil
}

// Delete removes a tournament by its ID.
func (r *TournamentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting tournament: %w", err)
	}
	if result.DeletedCount == 0 {
		return tournament.ErrNotFound
	}
	return nil
}

// List retrieves tournaments with optional filtering.
func (r *TournamentRepository) List(ctx context.Context, filter tournament.ListFilter) ([]*tournament.Tournament, error) {
	// Build query filter
	query := bson.M{}

	if filter.GameID != nil {
		query["game_id"] = *filter.GameID
	}

	if filter.Status != nil {
		query["status"] = *filter.Status
	}

	if filter.CreatedBy != nil {
		query["created_by"] = *filter.CreatedBy
	}

	// Set options
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}

	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	// Execute query
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("listing tournaments: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var tournaments []*tournament.Tournament
	if err := cursor.All(ctx, &tournaments); err != nil {
		return nil, fmt.Errorf("decoding tournaments: %w", err)
	}

	return tournaments, nil
}

// GetByGameID retrieves all tournaments for a specific game.
func (r *TournamentRepository) GetByGameID(ctx context.Context, gameID uuid.UUID) ([]*tournament.Tournament, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"game_id": gameID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("finding tournaments by game: %w", err)
	}
	defer cursor.Close(ctx)

	var tournaments []*tournament.Tournament
	if err := cursor.All(ctx, &tournaments); err != nil {
		return nil, fmt.Errorf("decoding tournaments: %w", err)
	}

	return tournaments, nil
}

// GetByStatus retrieves tournaments by status.
func (r *TournamentRepository) GetByStatus(ctx context.Context, status tournament.Status) ([]*tournament.Tournament, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"status": status},
		options.Find().SetSort(bson.D{{Key: "start_date", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("finding tournaments by status: %w", err)
	}
	defer cursor.Close(ctx)

	var tournaments []*tournament.Tournament
	if err := cursor.All(ctx, &tournaments); err != nil {
		return nil, fmt.Errorf("decoding tournaments: %w", err)
	}

	return tournaments, nil
}

// GetActiveTournaments retrieves all currently active tournaments.
func (r *TournamentRepository) GetActiveTournaments(ctx context.Context) ([]*tournament.Tournament, error) {
	return r.GetByStatus(ctx, tournament.StatusActive)
}

// CountByGameID returns the number of tournaments for a game.
func (r *TournamentRepository) CountByGameID(ctx context.Context, gameID uuid.UUID) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"game_id": gameID})
	if err != nil {
		return 0, fmt.Errorf("counting tournaments: %w", err)
	}
	return count, nil
}
