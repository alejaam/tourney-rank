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
	// PlayerStatsCollection is the MongoDB collection name for player stats.
	PlayerStatsCollection = "player_stats"
)

var (
	// ErrPlayerStatsNotFound is returned when player stats are not found.
	ErrPlayerStatsNotFound = errors.New("player stats not found")
)

// playerStatsDocument represents the MongoDB document structure for player stats.
type playerStatsDocument struct {
	ID            string                 `bson:"_id"`
	PlayerID      string                 `bson:"player_id"`
	GameID        string                 `bson:"game_id"`
	Stats         map[string]interface{} `bson:"stats"`
	MatchesPlayed int                    `bson:"matches_played"`
	RankingScore  float64                `bson:"ranking_score"`
	Tier          string                 `bson:"tier"`
	LastMatchAt   *time.Time             `bson:"last_match_at"`
	CreatedAt     time.Time              `bson:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at"`
}

// LeaderboardEntry represents a single entry in the leaderboard.
type LeaderboardEntry struct {
	Rank          int                    `json:"rank"`
	PlayerID      uuid.UUID              `json:"player_id"`
	DisplayName   string                 `json:"display_name"`
	AvatarURL     string                 `json:"avatar_url"`
	RankingScore  float64                `json:"ranking_score"`
	Tier          player.Tier            `json:"tier"`
	MatchesPlayed int                    `json:"matches_played"`
	Stats         map[string]interface{} `json:"stats"`
}

// PlayerStatsRepository implements player stats persistence using MongoDB.
type PlayerStatsRepository struct {
	collection       *mongo.Collection
	playerCollection *mongo.Collection
}

// NewPlayerStatsRepository creates a new PlayerStatsRepository.
func NewPlayerStatsRepository(client *Client) *PlayerStatsRepository {
	return &PlayerStatsRepository{
		collection:       client.Collection(PlayerStatsCollection),
		playerCollection: client.Collection(PlayersCollection),
	}
}

// Create inserts new player stats into the database.
func (r *PlayerStatsRepository) Create(ctx context.Context, ps *player.PlayerStats) error {
	doc := toPlayerStatsDocument(ps)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert player stats: %w", err)
	}

	return nil
}

// GetByID retrieves player stats by ID.
func (r *PlayerStatsRepository) GetByID(ctx context.Context, id uuid.UUID) (*player.PlayerStats, error) {
	var doc playerStatsDocument

	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPlayerStatsNotFound
		}
		return nil, fmt.Errorf("find player stats by id: %w", err)
	}

	return toPlayerStatsEntity(&doc)
}

// GetByPlayerAndGame retrieves player stats for a specific player and game.
func (r *PlayerStatsRepository) GetByPlayerAndGame(ctx context.Context, playerID, gameID uuid.UUID) (*player.PlayerStats, error) {
	var doc playerStatsDocument

	filter := bson.M{
		"player_id": playerID.String(),
		"game_id":   gameID.String(),
	}

	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPlayerStatsNotFound
		}
		return nil, fmt.Errorf("find player stats by player and game: %w", err)
	}

	return toPlayerStatsEntity(&doc)
}

// GetOrCreate retrieves player stats or creates them if they don't exist.
func (r *PlayerStatsRepository) GetOrCreate(ctx context.Context, playerID, gameID uuid.UUID) (*player.PlayerStats, error) {
	ps, err := r.GetByPlayerAndGame(ctx, playerID, gameID)
	if err == nil {
		return ps, nil
	}

	if !errors.Is(err, ErrPlayerStatsNotFound) {
		return nil, err
	}

	// Create new player stats
	ps = player.NewPlayerStats(playerID, gameID)
	if err := r.Create(ctx, ps); err != nil {
		return nil, err
	}

	return ps, nil
}

// Update updates existing player stats.
func (r *PlayerStatsRepository) Update(ctx context.Context, ps *player.PlayerStats) error {
	doc := toPlayerStatsDocument(ps)

	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": ps.ID.String()},
		doc,
	)
	if err != nil {
		return fmt.Errorf("update player stats: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrPlayerStatsNotFound
	}

	return nil
}

// UpdateRanking updates only the ranking score and tier.
func (r *PlayerStatsRepository) UpdateRanking(ctx context.Context, id uuid.UUID, score float64, tier player.Tier) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id.String()},
		bson.M{
			"$set": bson.M{
				"ranking_score": score,
				"tier":          string(tier),
				"updated_at":    time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("update ranking: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrPlayerStatsNotFound
	}

	return nil
}

// IncrementStats increments stats after a match.
func (r *PlayerStatsRepository) IncrementStats(ctx context.Context, id uuid.UUID, statsToAdd map[string]interface{}) error {
	inc := bson.M{
		"matches_played": 1,
	}

	for k, v := range statsToAdd {
		inc["stats."+k] = v
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id.String()},
		bson.M{
			"$inc": inc,
			"$set": bson.M{
				"last_match_at": now,
				"updated_at":    now,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("increment stats: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrPlayerStatsNotFound
	}

	return nil
}

// GetLeaderboard retrieves the top players for a game.
func (r *PlayerStatsRepository) GetLeaderboard(ctx context.Context, gameID uuid.UUID, limit, offset int64) ([]LeaderboardEntry, error) {
	pipeline := mongo.Pipeline{
		// Match by game
		{{Key: "$match", Value: bson.M{"game_id": gameID.String()}}},
		// Sort by ranking score descending
		{{Key: "$sort", Value: bson.D{{Key: "ranking_score", Value: -1}}}},
		// Skip and limit for pagination
		{{Key: "$skip", Value: offset}},
		{{Key: "$limit", Value: limit}},
		// Lookup player info
		{{Key: "$lookup", Value: bson.M{
			"from":         PlayersCollection,
			"localField":   "player_id",
			"foreignField": "_id",
			"as":           "player_info",
		}}},
		// Unwind player info
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$player_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		// Project final fields
		{{Key: "$project", Value: bson.M{
			"player_id":      1,
			"ranking_score":  1,
			"tier":           1,
			"matches_played": 1,
			"stats":          1,
			"display_name":   "$player_info.display_name",
			"avatar_url":     "$player_info.avatar_url",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate leaderboard: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []LeaderboardEntry
	rank := int(offset) + 1

	for cursor.Next(ctx) {
		var result struct {
			PlayerID      string                 `bson:"player_id"`
			RankingScore  float64                `bson:"ranking_score"`
			Tier          string                 `bson:"tier"`
			MatchesPlayed int                    `bson:"matches_played"`
			Stats         map[string]interface{} `bson:"stats"`
			DisplayName   string                 `bson:"display_name"`
			AvatarURL     string                 `bson:"avatar_url"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode leaderboard entry: %w", err)
		}

		playerID, _ := uuid.Parse(result.PlayerID)

		entries = append(entries, LeaderboardEntry{
			Rank:          rank,
			PlayerID:      playerID,
			DisplayName:   result.DisplayName,
			AvatarURL:     result.AvatarURL,
			RankingScore:  result.RankingScore,
			Tier:          player.Tier(result.Tier),
			MatchesPlayed: result.MatchesPlayed,
			Stats:         result.Stats,
		})
		rank++
	}

	return entries, nil
}

// GetLeaderboardByTier retrieves top players filtered by tier.
func (r *PlayerStatsRepository) GetLeaderboardByTier(ctx context.Context, gameID uuid.UUID, tier player.Tier, limit int64) ([]LeaderboardEntry, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"game_id": gameID.String(),
			"tier":    string(tier),
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "ranking_score", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$lookup", Value: bson.M{
			"from":         PlayersCollection,
			"localField":   "player_id",
			"foreignField": "_id",
			"as":           "player_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$player_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$project", Value: bson.M{
			"player_id":      1,
			"ranking_score":  1,
			"tier":           1,
			"matches_played": 1,
			"stats":          1,
			"display_name":   "$player_info.display_name",
			"avatar_url":     "$player_info.avatar_url",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate leaderboard by tier: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []LeaderboardEntry
	rank := 1

	for cursor.Next(ctx) {
		var result struct {
			PlayerID      string                 `bson:"player_id"`
			RankingScore  float64                `bson:"ranking_score"`
			Tier          string                 `bson:"tier"`
			MatchesPlayed int                    `bson:"matches_played"`
			Stats         map[string]interface{} `bson:"stats"`
			DisplayName   string                 `bson:"display_name"`
			AvatarURL     string                 `bson:"avatar_url"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode leaderboard entry: %w", err)
		}

		playerID, _ := uuid.Parse(result.PlayerID)

		entries = append(entries, LeaderboardEntry{
			Rank:          rank,
			PlayerID:      playerID,
			DisplayName:   result.DisplayName,
			AvatarURL:     result.AvatarURL,
			RankingScore:  result.RankingScore,
			Tier:          player.Tier(result.Tier),
			MatchesPlayed: result.MatchesPlayed,
			Stats:         result.Stats,
		})
		rank++
	}

	return entries, nil
}

// GetPlayerRank retrieves a player's rank in a game.
func (r *PlayerStatsRepository) GetPlayerRank(ctx context.Context, playerID, gameID uuid.UUID) (int, error) {
	// Get player's score first
	ps, err := r.GetByPlayerAndGame(ctx, playerID, gameID)
	if err != nil {
		return 0, err
	}

	// Count players with higher score
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"game_id":       gameID.String(),
		"ranking_score": bson.M{"$gt": ps.RankingScore},
	})
	if err != nil {
		return 0, fmt.Errorf("count higher ranked players: %w", err)
	}

	return int(count) + 1, nil
}

// GetTierDistribution returns the count of players in each tier for a game.
func (r *PlayerStatsRepository) GetTierDistribution(ctx context.Context, gameID uuid.UUID) (map[player.Tier]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"game_id": gameID.String()}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$tier",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate tier distribution: %w", err)
	}
	defer cursor.Close(ctx)

	distribution := make(map[player.Tier]int64)

	for cursor.Next(ctx) {
		var result struct {
			Tier  string `bson:"_id"`
			Count int64  `bson:"count"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode tier distribution: %w", err)
		}

		distribution[player.Tier(result.Tier)] = result.Count
	}

	return distribution, nil
}

// GetTopStatsByGame returns top N players for a specific stat in a game.
func (r *PlayerStatsRepository) GetTopStatsByGame(ctx context.Context, gameID uuid.UUID, statName string, limit int64) ([]LeaderboardEntry, error) {
	statField := "stats." + statName

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"game_id": gameID.String()}}},
		{{Key: "$sort", Value: bson.D{{Key: statField, Value: -1}}}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$lookup", Value: bson.M{
			"from":         PlayersCollection,
			"localField":   "player_id",
			"foreignField": "_id",
			"as":           "player_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$player_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$project", Value: bson.M{
			"player_id":      1,
			"ranking_score":  1,
			"tier":           1,
			"matches_played": 1,
			"stats":          1,
			"display_name":   "$player_info.display_name",
			"avatar_url":     "$player_info.avatar_url",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate top stats: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []LeaderboardEntry
	rank := 1

	for cursor.Next(ctx) {
		var result struct {
			PlayerID      string                 `bson:"player_id"`
			RankingScore  float64                `bson:"ranking_score"`
			Tier          string                 `bson:"tier"`
			MatchesPlayed int                    `bson:"matches_played"`
			Stats         map[string]interface{} `bson:"stats"`
			DisplayName   string                 `bson:"display_name"`
			AvatarURL     string                 `bson:"avatar_url"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode top stats entry: %w", err)
		}

		playerID, _ := uuid.Parse(result.PlayerID)

		entries = append(entries, LeaderboardEntry{
			Rank:          rank,
			PlayerID:      playerID,
			DisplayName:   result.DisplayName,
			AvatarURL:     result.AvatarURL,
			RankingScore:  result.RankingScore,
			Tier:          player.Tier(result.Tier),
			MatchesPlayed: result.MatchesPlayed,
			Stats:         result.Stats,
		})
		rank++
	}

	return entries, nil
}

// EnsureIndexes creates necessary indexes for the player_stats collection.
func (r *PlayerStatsRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "player_id", Value: 1}, {Key: "game_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "game_id", Value: 1}, {Key: "ranking_score", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "game_id", Value: 1}, {Key: "tier", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "player_id", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("create player stats indexes: %w", err)
	}

	return nil
}

// toPlayerStatsDocument converts a domain PlayerStats to a MongoDB document.
func toPlayerStatsDocument(ps *player.PlayerStats) *playerStatsDocument {
	return &playerStatsDocument{
		ID:            ps.ID.String(),
		PlayerID:      ps.PlayerID.String(),
		GameID:        ps.GameID.String(),
		Stats:         ps.Stats,
		MatchesPlayed: ps.MatchesPlayed,
		RankingScore:  ps.RankingScore,
		Tier:          string(ps.Tier),
		LastMatchAt:   ps.LastMatchAt,
		CreatedAt:     ps.CreatedAt,
		UpdatedAt:     ps.UpdatedAt,
	}
}

// toPlayerStatsEntity converts a MongoDB document to a domain PlayerStats.
func toPlayerStatsEntity(doc *playerStatsDocument) (*player.PlayerStats, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("parse player stats id: %w", err)
	}

	playerID, err := uuid.Parse(doc.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("parse player id: %w", err)
	}

	gameID, err := uuid.Parse(doc.GameID)
	if err != nil {
		return nil, fmt.Errorf("parse game id: %w", err)
	}

	stats := doc.Stats
	if stats == nil {
		stats = make(map[string]interface{})
	}

	return &player.PlayerStats{
		ID:            id,
		PlayerID:      playerID,
		GameID:        gameID,
		Stats:         stats,
		MatchesPlayed: doc.MatchesPlayed,
		RankingScore:  doc.RankingScore,
		Tier:          player.Tier(doc.Tier),
		LastMatchAt:   doc.LastMatchAt,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}, nil
}
