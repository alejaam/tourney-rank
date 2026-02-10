package mongodb

import (
	"context"
	"errors"

	"github.com/alejaam/tourney-rank/internal/domain/bracket"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BracketRepository implements bracket.Repository using MongoDB.
type BracketRepository struct {
	brackets *mongo.Collection
	matchups *mongo.Collection
}

// NewBracketRepository creates a new MongoDB bracket repository.
func NewBracketRepository(db *mongo.Database) *BracketRepository {
	return &BracketRepository{
		brackets: db.Collection("brackets"),
		matchups: db.Collection("matchups"),
	}
}

// Create creates a new bracket.
func (r *BracketRepository) Create(ctx context.Context, b *bracket.Bracket) error {
	_, err := r.brackets.InsertOne(ctx, b)
	if err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a bracket by ID.
func (r *BracketRepository) GetByID(ctx context.Context, id uuid.UUID) (*bracket.Bracket, error) {
	var b bracket.Bracket
	err := r.brackets.FindOne(ctx, bson.M{"_id": id}).Decode(&b)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, bracket.ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

// GetByTournamentID retrieves a bracket by tournament ID.
func (r *BracketRepository) GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*bracket.Bracket, error) {
	var b bracket.Bracket
	err := r.brackets.FindOne(ctx, bson.M{"tournament_id": tournamentID}).Decode(&b)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, bracket.ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

// Update updates a bracket.
func (r *BracketRepository) Update(ctx context.Context, b *bracket.Bracket) error {
	result, err := r.brackets.ReplaceOne(ctx, bson.M{"_id": b.ID}, b)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return bracket.ErrNotFound
	}
	return nil
}

// Delete deletes a bracket.
func (r *BracketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.brackets.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return bracket.ErrNotFound
	}
	return nil
}

// CreateMatchup creates a new matchup.
func (r *BracketRepository) CreateMatchup(ctx context.Context, m *bracket.Matchup) error {
	_, err := r.matchups.InsertOne(ctx, m)
	return err
}

// GetMatchup retrieves a matchup by ID.
func (r *BracketRepository) GetMatchup(ctx context.Context, id uuid.UUID) (*bracket.Matchup, error) {
	var m bracket.Matchup
	err := r.matchups.FindOne(ctx, bson.M{"_id": id}).Decode(&m)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, bracket.ErrMatchupNotFound
		}
		return nil, err
	}
	return &m, nil
}

// GetMatchupsByBracket retrieves all matchups for a bracket.
func (r *BracketRepository) GetMatchupsByBracket(ctx context.Context, bracketID uuid.UUID) ([]*bracket.Matchup, error) {
	cursor, err := r.matchups.Find(ctx, bson.M{"bracket_id": bracketID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matchups []*bracket.Matchup
	if err := cursor.All(ctx, &matchups); err != nil {
		return nil, err
	}
	return matchups, nil
}

// GetMatchupsByRound retrieves all matchups for a specific round in a bracket.
func (r *BracketRepository) GetMatchupsByRound(ctx context.Context, bracketID uuid.UUID, round int) ([]*bracket.Matchup, error) {
	cursor, err := r.matchups.Find(ctx, bson.M{
		"bracket_id": bracketID,
		"round":      round,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matchups []*bracket.Matchup
	if err := cursor.All(ctx, &matchups); err != nil {
		return nil, err
	}
	return matchups, nil
}

// UpdateMatchup updates a matchup.
func (r *BracketRepository) UpdateMatchup(ctx context.Context, m *bracket.Matchup) error {
	result, err := r.matchups.ReplaceOne(ctx, bson.M{"_id": m.ID}, m)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return bracket.ErrMatchupNotFound
	}
	return nil
}

// DeleteMatchup deletes a matchup.
func (r *BracketRepository) DeleteMatchup(ctx context.Context, id uuid.UUID) error {
	result, err := r.matchups.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return bracket.ErrMatchupNotFound
	}
	return nil
}

// EnsureIndexes creates necessary database indexes.
func (r *BracketRepository) EnsureIndexes(ctx context.Context) error {
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tournament_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "bracket_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "round", Value: 1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(0)

	// Create indexes for matchups collection
	if _, err := r.matchups.Indexes().CreateMany(ctx, indexModels, opts); err != nil {
		return err
	}

	return nil
}
