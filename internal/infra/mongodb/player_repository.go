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

	"github.com/melisource/tourney-rank/internal/domain/player"
)

const (
	// PlayersCollection is the MongoDB collection name for players.
	PlayersCollection = "players"
)

var (
	// ErrPlayerAlreadyExists is returned when trying to create a player that already exists.
	ErrPlayerAlreadyExists = errors.New("player already exists")
)

// playerDocument represents the MongoDB document structure for a player.
type playerDocument struct {
	ID                string            `bson:"_id"`
	UserID            string            `bson:"user_id"`
	DisplayName       string            `bson:"display_name"`
	AvatarURL         string            `bson:"avatar_url,omitempty"`
	Bio               string            `bson:"bio,omitempty"`
	PlatformIDs       map[string]string `bson:"platform_ids,omitempty"`
	BirthYear         int               `bson:"birth_year,omitempty"`
	Region            string            `bson:"region,omitempty"`
	PreferredPlatform string            `bson:"preferred_platform,omitempty"`
	Language          string            `bson:"language,omitempty"`
	IsBanned          bool              `bson:"is_banned"`
	BannedAt          *time.Time        `bson:"banned_at,omitempty"`
	CreatedAt         time.Time         `bson:"created_at"`
	UpdatedAt         time.Time         `bson:"updated_at"`
}

// PlayerRepository implements player persistence using MongoDB.
type PlayerRepository struct {
	collection *mongo.Collection
}

// NewPlayerRepository creates a new PlayerRepository.
func NewPlayerRepository(client *Client) *PlayerRepository {
	return &PlayerRepository{
		collection: client.Collection(PlayersCollection),
	}
}

// Create inserts a new player into the database.
func (r *PlayerRepository) Create(ctx context.Context, p *player.Player) error {
	doc := toPlayerDocument(p)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrPlayerAlreadyExists
		}
		return fmt.Errorf("insert player: %w", err)
	}

	return nil
}

// GetByID retrieves a player by their ID.
func (r *PlayerRepository) GetByID(ctx context.Context, id string) (*player.Player, error) {
	var doc playerDocument

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, player.ErrNotFound
		}
		return nil, fmt.Errorf("find player by id: %w", err)
	}

	return toPlayerEntity(&doc)
}

// GetByUserID retrieves a player by their user ID.
func (r *PlayerRepository) GetByUserID(ctx context.Context, userID string) (*player.Player, error) {
	var doc playerDocument

	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, player.ErrNotFound
		}
		return nil, fmt.Errorf("find player by user id: %w", err)
	}

	return toPlayerEntity(&doc)
}

// GetByPlatformID retrieves a player by a platform-specific ID.
func (r *PlayerRepository) GetByPlatformID(ctx context.Context, platform, platformID string) (*player.Player, error) {
	var doc playerDocument

	filter := bson.M{
		fmt.Sprintf("platform_ids.%s", platform): platformID,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, player.ErrNotFound
		}
		return nil, fmt.Errorf("find player by platform id: %w", err)
	}

	return toPlayerEntity(&doc)
}

// List retrieves players with pagination.
func (r *PlayerRepository) List(ctx context.Context, limit, offset int64) ([]*player.Player, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "display_name", Value: 1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find players: %w", err)
	}
	defer cursor.Close(ctx)

	var players []*player.Player
	for cursor.Next(ctx) {
		var doc playerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode player: %w", err)
		}

		p, err := toPlayerEntity(&doc)
		if err != nil {
			return nil, fmt.Errorf("convert player entity: %w", err)
		}
		players = append(players, p)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return players, nil
}

// GetAll retrieves all players without pagination.
func (r *PlayerRepository) GetAll(ctx context.Context) ([]*player.Player, error) {
	opts := options.Find().SetSort(bson.D{{Key: "display_name", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find all players: %w", err)
	}
	defer cursor.Close(ctx)

	var players []*player.Player
	for cursor.Next(ctx) {
		var doc playerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode player: %w", err)
		}

		p, err := toPlayerEntity(&doc)
		if err != nil {
			return nil, fmt.Errorf("convert player entity: %w", err)
		}
		players = append(players, p)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return players, nil
}

// Search searches players by display name.
func (r *PlayerRepository) Search(ctx context.Context, query string, limit int64) ([]*player.Player, error) {
	filter := bson.M{
		"display_name": bson.M{
			"$regex":   query,
			"$options": "i", // case-insensitive
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "display_name", Value: 1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("search players: %w", err)
	}
	defer cursor.Close(ctx)

	var players []*player.Player
	for cursor.Next(ctx) {
		var doc playerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode player: %w", err)
		}

		p, err := toPlayerEntity(&doc)
		if err != nil {
			return nil, fmt.Errorf("convert player entity: %w", err)
		}
		players = append(players, p)
	}

	return players, nil
}

// Update updates an existing player.
func (r *PlayerRepository) Update(ctx context.Context, p *player.Player) error {
	doc := toPlayerDocument(p)

	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": p.ID.String()},
		doc,
	)
	if err != nil {
		return fmt.Errorf("update player: %w", err)
	}

	if result.MatchedCount == 0 {
		return player.ErrNotFound
	}

	return nil
}

// Delete removes a player from the database.
func (r *PlayerRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete player: %w", err)
	}

	if result.DeletedCount == 0 {
		return player.ErrNotFound
	}

	return nil
}

// Count returns the total number of players.
func (r *PlayerRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("count players: %w", err)
	}
	return count, nil
}

// EnsureIndexes creates necessary indexes for the players collection.
func (r *PlayerRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "display_name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "display_name", Value: "text"}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("create player indexes: %w", err)
	}

	return nil
}

// toPlayerDocument converts a domain Player to a MongoDB document.
func toPlayerDocument(p *player.Player) *playerDocument {
	return &playerDocument{
		ID:                p.ID.String(),
		UserID:            p.UserID.String(),
		DisplayName:       p.DisplayName,
		AvatarURL:         p.AvatarURL,
		Bio:               p.Bio,
		PlatformIDs:       p.PlatformIDs,
		BirthYear:         p.BirthYear,
		Region:            p.Region,
		PreferredPlatform: p.PreferredPlatform,
		Language:          p.Language,
		IsBanned:          p.IsBanned,
		BannedAt:          p.BannedAt,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// toPlayerEntity converts a MongoDB document to a domain Player.
func toPlayerEntity(doc *playerDocument) (*player.Player, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("parse player id: %w", err)
	}

	userID, err := uuid.Parse(doc.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	platformIDs := doc.PlatformIDs
	if platformIDs == nil {
		platformIDs = make(map[string]string)
	}

	return &player.Player{
		ID:                id,
		UserID:            userID,
		DisplayName:       doc.DisplayName,
		AvatarURL:         doc.AvatarURL,
		Bio:               doc.Bio,
		PlatformIDs:       platformIDs,
		BirthYear:         doc.BirthYear,
		Region:            doc.Region,
		PreferredPlatform: doc.PreferredPlatform,
		Language:          doc.Language,
		IsBanned:          doc.IsBanned,
		BannedAt:          doc.BannedAt,
		CreatedAt:         doc.CreatedAt,
		UpdatedAt:         doc.UpdatedAt,
	}, nil
}
