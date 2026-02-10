package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alejaam/tourney-rank/internal/domain/round"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// RoundsCollection is the MongoDB collection name for rounds.
	RoundsCollection = "rounds"
)

// roundDocument represents the MongoDB document structure for a round.
type roundDocument struct {
	ID           uuid.UUID  `bson:"_id"`
	TournamentID uuid.UUID  `bson:"tournament_id"`
	Number       int        `bson:"number"`
	Status       string     `bson:"status"`
	ScheduledFor time.Time  `bson:"scheduled_for"`
	StartedAt    *time.Time `bson:"started_at,omitempty"`
	CompletedAt  *time.Time `bson:"completed_at,omitempty"`
	MatchCount   int        `bson:"match_count"`
	Notes        string     `bson:"notes,omitempty"`
	CreatedAt    time.Time  `bson:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"`
}

// RoundRepository implements round persistence using MongoDB.
type RoundRepository struct {
	collection *mongo.Collection
}

// NewRoundRepository creates a new RoundRepository.
func NewRoundRepository(db *mongo.Database) *RoundRepository {
	return &RoundRepository{
		collection: db.Collection(RoundsCollection),
	}
}

// Create inserts a new round into the database.
func (r *RoundRepository) Create(ctx context.Context, rnd *round.Round) error {
	doc := toRoundDocument(rnd)
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert round: %w", err)
	}
	return nil
}

// GetByID retrieves a round by its ID.
func (r *RoundRepository) GetByID(ctx context.Context, id uuid.UUID) (*round.Round, error) {
	var doc roundDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, round.ErrNotFound
		}
		return nil, fmt.Errorf("find round by id: %w", err)
	}
	return toRoundEntity(&doc), nil
}

// GetByTournamentAndNumber retrieves a round by tournament ID and round number.
func (r *RoundRepository) GetByTournamentAndNumber(ctx context.Context, tournamentID uuid.UUID, number int) (*round.Round, error) {
	var doc roundDocument
	err := r.collection.FindOne(ctx, bson.M{
		"tournament_id": tournamentID,
		"number":        number,
	}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, round.ErrNotFound
		}
		return nil, fmt.Errorf("find round by tournament and number: %w", err)
	}
	return toRoundEntity(&doc), nil
}

// GetByTournament retrieves all rounds for a tournament.
func (r *RoundRepository) GetByTournament(ctx context.Context, tournamentID uuid.UUID) ([]*round.Round, error) {
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"tournament_id": tournamentID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find rounds by tournament: %w", err)
	}
	defer cursor.Close(ctx)

	var rounds []*round.Round
	for cursor.Next(ctx) {
		var doc roundDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode round: %w", err)
		}
		rounds = append(rounds, toRoundEntity(&doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return rounds, nil
}

// Update updates an existing round.
func (r *RoundRepository) Update(ctx context.Context, rnd *round.Round) error {
	doc := toRoundDocument(rnd)
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": rnd.ID}, doc)
	if err != nil {
		return fmt.Errorf("update round: %w", err)
	}
	if result.MatchedCount == 0 {
		return round.ErrNotFound
	}
	return nil
}

// Delete removes a round from the database.
func (r *RoundRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete round: %w", err)
	}
	if result.DeletedCount == 0 {
		return round.ErrNotFound
	}
	return nil
}

// EnsureIndexes creates necessary database indexes.
func (r *RoundRepository) EnsureIndexes(ctx context.Context) error {
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tournament_id", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "tournament_id", Value: 1},
				{Key: "number", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(0)
	if _, err := r.collection.Indexes().CreateMany(ctx, indexModels, opts); err != nil {
		return fmt.Errorf("create indexes: %w", err)
	}

	return nil
}

// Helper functions

func toRoundDocument(rnd *round.Round) *roundDocument {
	return &roundDocument{
		ID:           rnd.ID,
		TournamentID: rnd.TournamentID,
		Number:       rnd.Number,
		Status:       string(rnd.Status),
		ScheduledFor: rnd.ScheduledFor,
		StartedAt:    rnd.StartedAt,
		CompletedAt:  rnd.CompletedAt,
		MatchCount:   rnd.MatchCount,
		Notes:        rnd.Notes,
		CreatedAt:    rnd.CreatedAt,
		UpdatedAt:    rnd.UpdatedAt,
	}
}

func toRoundEntity(doc *roundDocument) *round.Round {
	return &round.Round{
		ID:           doc.ID,
		TournamentID: doc.TournamentID,
		Number:       doc.Number,
		Status:       round.Status(doc.Status),
		ScheduledFor: doc.ScheduledFor,
		StartedAt:    doc.StartedAt,
		CompletedAt:  doc.CompletedAt,
		MatchCount:   doc.MatchCount,
		Notes:        doc.Notes,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}
