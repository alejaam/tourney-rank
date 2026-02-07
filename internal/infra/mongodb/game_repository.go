// Package mongodb provides MongoDB repository implementations.
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/melisource/tourney-rank/internal/domain/game"
)

const (
	// GamesCollection is the MongoDB collection name for games.
	GamesCollection = "games"
)

var (
	// ErrGameAlreadyExists is returned when trying to create a game that already exists.
	ErrGameAlreadyExists = errors.New("game already exists")
)

// gameDocument represents the MongoDB document structure for a game.
type gameDocument struct {
	ID               string                 `bson:"_id"`
	Name             string                 `bson:"name"`
	Slug             string                 `bson:"slug"`
	Description      string                 `bson:"description"`
	StatSchema       map[string]interface{} `bson:"stat_schema"`
	RankingWeights   map[string]float64     `bson:"ranking_weights"`
	PlatformIDFormat string                 `bson:"platform_id_format"`
	IsActive         bool                   `bson:"is_active"`
	CreatedAt        time.Time              `bson:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at"`
}

// GameRepository implements game persistence using MongoDB.
type GameRepository struct {
	collection *mongo.Collection
}

// NewGameRepository creates a new GameRepository.
func NewGameRepository(client *Client) *GameRepository {
	return &GameRepository{
		collection: client.Collection(GamesCollection),
	}
}

// Create inserts a new game into the database.
func (r *GameRepository) Create(ctx context.Context, g *game.Game) error {
	doc := toGameDocument(g)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrGameAlreadyExists
		}
		return fmt.Errorf("insert game: %w", err)
	}

	return nil
}

// GetByID retrieves a game by its ID.
func (r *GameRepository) GetByID(ctx context.Context, id string) (*game.Game, error) {
	var doc gameDocument

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("find game by id: %w", err)
	}

	return toGameEntity(&doc)
}

// GetBySlug retrieves a game by its slug.
func (r *GameRepository) GetBySlug(ctx context.Context, slug string) (*game.Game, error) {
	var doc gameDocument

	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, game.ErrNotFound
		}
		return nil, fmt.Errorf("find game by slug: %w", err)
	}

	return toGameEntity(&doc)
}

// GetAll retrieves all games without filtering.
func (r *GameRepository) GetAll(ctx context.Context) ([]*game.Game, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find games: %w", err)
	}
	defer cursor.Close(ctx)

	var games []*game.Game
	for cursor.Next(ctx) {
		var doc gameDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode game: %w", err)
		}

		g, err := toGameEntity(&doc)
		if err != nil {
			return nil, fmt.Errorf("convert game entity: %w", err)
		}
		games = append(games, g)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return games, nil
}

// List retrieves all games with optional filtering.
func (r *GameRepository) List(ctx context.Context, activeOnly bool) ([]*game.Game, error) {
	filter := bson.M{}
	if activeOnly {
		filter["is_active"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find games: %w", err)
	}
	defer cursor.Close(ctx)

	var games []*game.Game
	for cursor.Next(ctx) {
		var doc gameDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode game: %w", err)
		}

		g, err := toGameEntity(&doc)
		if err != nil {
			return nil, fmt.Errorf("convert game entity: %w", err)
		}
		games = append(games, g)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return games, nil
}

// Update updates an existing game.
func (r *GameRepository) Update(ctx context.Context, g *game.Game) error {
	doc := toGameDocument(g)

	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": g.ID.String()},
		doc,
	)
	if err != nil {
		return fmt.Errorf("update game: %w", err)
	}

	if result.MatchedCount == 0 {
		return game.ErrNotFound
	}

	return nil
}

// Delete removes a game from the database.
func (r *GameRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete game: %w", err)
	}

	if result.DeletedCount == 0 {
		return game.ErrNotFound
	}

	return nil
}

// SetActive updates the active status of a game.
func (r *GameRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id.String()},
		bson.M{
			"$set": bson.M{
				"is_active":  active,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("set game active: %w", err)
	}

	if result.MatchedCount == 0 {
		return game.ErrNotFound
	}

	return nil
}

// EnsureIndexes creates necessary indexes for the games collection.
func (r *GameRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("create game indexes: %w", err)
	}

	return nil
}

// toGameDocument converts a domain Game to a MongoDB document.
func toGameDocument(g *game.Game) *gameDocument {
	statSchema := make(map[string]interface{})
	for k, v := range g.StatSchema {
		statSchema[k] = map[string]interface{}{
			"type":  v.Type,
			"min":   v.Min,
			"max":   v.Max,
			"label": v.Label,
		}
	}

	return &gameDocument{
		ID:               g.ID.String(),
		Name:             g.Name,
		Slug:             g.Slug,
		Description:      g.Description,
		StatSchema:       statSchema,
		RankingWeights:   g.RankingWeights,
		PlatformIDFormat: g.PlatformIDFormat,
		IsActive:         g.IsActive,
		CreatedAt:        g.CreatedAt,
		UpdatedAt:        g.UpdatedAt,
	}
}

// toGameEntity converts a MongoDB document to a domain Game.
func toGameEntity(doc *gameDocument) (*game.Game, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("parse game id: %w", err)
	}

	statSchema := make(game.StatSchema)
	for k, v := range doc.StatSchema {
		if m, ok := v.(map[string]interface{}); ok {
			field := game.StatField{}
			if t, ok := m["type"].(string); ok {
				field.Type = t
			}
			if l, ok := m["label"].(string); ok {
				field.Label = l
			}
			field.Min = m["min"]
			field.Max = m["max"]
			statSchema[k] = field
		}
	}

	return &game.Game{
		ID:               id,
		Name:             doc.Name,
		Slug:             doc.Slug,
		Description:      doc.Description,
		StatSchema:       statSchema,
		RankingWeights:   doc.RankingWeights,
		PlatformIDFormat: doc.PlatformIDFormat,
		IsActive:         doc.IsActive,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}, nil
}
