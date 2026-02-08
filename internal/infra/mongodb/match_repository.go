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

	"github.com/alejaam/tourney-rank/internal/domain/match"
)

const (
	// MatchesCollection is the MongoDB collection name for matches.
	MatchesCollection = "matches"
)

// matchDocument represents the MongoDB document structure for a match.
type matchDocument struct {
	ID              string                     `bson:"_id"`
	TournamentID    string                     `bson:"tournament_id"`
	TeamID          string                     `bson:"team_id"`
	GameID          string                     `bson:"game_id"`
	Status          string                     `bson:"status"`
	TeamPlacement   int                        `bson:"team_placement"`
	TeamKills       int                        `bson:"team_kills"`
	PlayerStats     []playerMatchStatsDocument `bson:"player_stats"`
	ScreenshotURL   string                     `bson:"screenshot_url"`
	RejectionReason string                     `bson:"rejection_reason,omitempty"`
	SubmittedBy     string                     `bson:"submitted_by"`
	CreatedAt       time.Time                  `bson:"created_at"`
	UpdatedAt       time.Time                  `bson:"updated_at"`
	VerifiedAt      *time.Time                 `bson:"verified_at,omitempty"`
	VerifiedBy      *string                    `bson:"verified_by,omitempty"`
}

// playerMatchStatsDocument represents player stats for a match.
type playerMatchStatsDocument struct {
	PlayerID    string                 `bson:"player_id"`
	Kills       int                    `bson:"kills"`
	Damage      int                    `bson:"damage"`
	Assists     int                    `bson:"assists"`
	Deaths      int                    `bson:"deaths"`
	Downs       int                    `bson:"downs"`
	CustomStats map[string]interface{} `bson:"custom_stats"`
}

// MatchRepository implements match persistence using MongoDB.
type MatchRepository struct {
	collection *mongo.Collection
}

// NewMatchRepository creates a new MatchRepository.
func NewMatchRepository(db *mongo.Database) *MatchRepository {
	return &MatchRepository{
		collection: db.Collection(MatchesCollection),
	}
}

// EnsureIndexes creates the necessary MongoDB indexes for matches.
func (r *MatchRepository) EnsureIndexes(ctx context.Context) error {
	indexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tournament_id", Value: 1}, {Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "team_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "submitted_by", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "player_stats.player_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("create indexes: %w", err)
	}

	return nil
}

// Create inserts a new match into the database.
func (r *MatchRepository) Create(ctx context.Context, m *match.Match) error {
	doc := toMatchDocument(m)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert match: %w", err)
	}

	return nil
}

// GetByID retrieves a match by ID.
func (r *MatchRepository) GetByID(ctx context.Context, id string) (*match.Match, error) {
	var doc matchDocument

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, match.ErrNotFound
		}
		return nil, fmt.Errorf("find match by id: %w", err)
	}

	return toMatchEntity(&doc)
}

// GetByTournament retrieves all matches in a tournament with pagination.
func (r *MatchRepository) GetByTournament(ctx context.Context, tournamentID string, limit int, offset int) ([]match.Match, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{"tournament_id": tournamentID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find matches by tournament: %w", err)
	}
	defer cursor.Close(ctx)

	return decodeMatches(ctx, cursor)
}

// GetByTeam retrieves all matches for a specific team.
func (r *MatchRepository) GetByTeam(ctx context.Context, teamID string, limit int, offset int) ([]match.Match, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{"team_id": teamID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find matches by team: %w", err)
	}
	defer cursor.Close(ctx)

	return decodeMatches(ctx, cursor)
}

// GetByPlayer retrieves all matches involving a specific player.
func (r *MatchRepository) GetByPlayer(ctx context.Context, playerID string, limit int, offset int) ([]match.Match, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	filter := bson.M{
		"player_stats.player_id": playerID,
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find matches by player: %w", err)
	}
	defer cursor.Close(ctx)

	return decodeMatches(ctx, cursor)
}

// GetUnverified retrieves all unverified (draft) matches for admin review.
func (r *MatchRepository) GetUnverified(ctx context.Context, limit int, offset int) ([]match.Match, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{"status": string(match.StatusDraft)}, opts)
	if err != nil {
		return nil, fmt.Errorf("find unverified matches: %w", err)
	}
	defer cursor.Close(ctx)

	return decodeMatches(ctx, cursor)
}

// GetTournamentUnverified retrieves unverified matches in a specific tournament.
func (r *MatchRepository) GetTournamentUnverified(ctx context.Context, tournamentID string, limit int, offset int) ([]match.Match, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	filter := bson.M{
		"tournament_id": tournamentID,
		"status":        string(match.StatusDraft),
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find tournament unverified matches: %w", err)
	}
	defer cursor.Close(ctx)

	return decodeMatches(ctx, cursor)
}

// Update updates an existing match.
func (r *MatchRepository) Update(ctx context.Context, m *match.Match) error {
	doc := toMatchDocument(m)

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": m.ID.String()}, doc)
	if err != nil {
		return fmt.Errorf("update match: %w", err)
	}

	if result.MatchedCount == 0 {
		return match.ErrNotFound
	}

	return nil
}

// CountByTournament returns the total number of matches in a tournament.
func (r *MatchRepository) CountByTournament(ctx context.Context, tournamentID string) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"tournament_id": tournamentID})
	if err != nil {
		return 0, fmt.Errorf("count matches by tournament: %w", err)
	}
	return int(count), nil
}

// CountUnverified returns total unverified matches.
func (r *MatchRepository) CountUnverified(ctx context.Context) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"status": string(match.StatusDraft)})
	if err != nil {
		return 0, fmt.Errorf("count unverified matches: %w", err)
	}
	return int(count), nil
}

// DeleteByID deletes a match (for testing purposes).
func (r *MatchRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete match: %w", err)
	}

	if result.DeletedCount == 0 {
		return match.ErrNotFound
	}

	return nil
}

// Helper functions

func toMatchDocument(m *match.Match) *matchDocument {
	playerStats := make([]playerMatchStatsDocument, len(m.PlayerStats))
	for i, ps := range m.PlayerStats {
		playerStats[i] = playerMatchStatsDocument{
			PlayerID:    ps.PlayerID.String(),
			Kills:       ps.Kills,
			Damage:      ps.Damage,
			Assists:     ps.Assists,
			Deaths:      ps.Deaths,
			Downs:       ps.Downs,
			CustomStats: ps.CustomStats,
		}
	}

	doc := &matchDocument{
		ID:              m.ID.String(),
		TournamentID:    m.TournamentID.String(),
		TeamID:          m.TeamID.String(),
		GameID:          m.GameID.String(),
		Status:          string(m.Status),
		TeamPlacement:   m.TeamPlacement,
		TeamKills:       m.TeamKills,
		PlayerStats:     playerStats,
		ScreenshotURL:   m.ScreenshotURL,
		RejectionReason: m.RejectionReason,
		SubmittedBy:     m.SubmittedBy.String(),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		VerifiedAt:      m.VerifiedAt,
	}

	if m.VerifiedBy != nil {
		verifiedByStr := m.VerifiedBy.String()
		doc.VerifiedBy = &verifiedByStr
	}

	return doc
}

func toMatchEntity(doc *matchDocument) (*match.Match, error) {
	playerStats := make([]match.PlayerMatchStats, len(doc.PlayerStats))
	for i, ps := range doc.PlayerStats {
		playerID, err := uuid.Parse(ps.PlayerID)
		if err != nil {
			return nil, fmt.Errorf("parse player id: %w", err)
		}
		playerStats[i] = match.PlayerMatchStats{
			PlayerID:    playerID,
			Kills:       ps.Kills,
			Damage:      ps.Damage,
			Assists:     ps.Assists,
			Deaths:      ps.Deaths,
			Downs:       ps.Downs,
			CustomStats: ps.CustomStats,
		}
	}

	tournamentID, err := uuid.Parse(doc.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("parse tournament id: %w", err)
	}

	teamID, err := uuid.Parse(doc.TeamID)
	if err != nil {
		return nil, fmt.Errorf("parse team id: %w", err)
	}

	gameID, err := uuid.Parse(doc.GameID)
	if err != nil {
		return nil, fmt.Errorf("parse game id: %w", err)
	}

	submittedBy, err := uuid.Parse(doc.SubmittedBy)
	if err != nil {
		return nil, fmt.Errorf("parse submitted by: %w", err)
	}

	m := &match.Match{
		ID:              uuid.MustParse(doc.ID),
		TournamentID:    tournamentID,
		TeamID:          teamID,
		GameID:          gameID,
		Status:          match.Status(doc.Status),
		TeamPlacement:   doc.TeamPlacement,
		TeamKills:       doc.TeamKills,
		PlayerStats:     playerStats,
		ScreenshotURL:   doc.ScreenshotURL,
		RejectionReason: doc.RejectionReason,
		SubmittedBy:     submittedBy,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       doc.UpdatedAt,
		VerifiedAt:      doc.VerifiedAt,
	}

	if doc.VerifiedBy != nil {
		verifiedBy, err := uuid.Parse(*doc.VerifiedBy)
		if err != nil {
			return nil, fmt.Errorf("parse verified by: %w", err)
		}
		m.VerifiedBy = &verifiedBy
	}

	return m, nil
}

func decodeMatches(ctx context.Context, cursor *mongo.Cursor) ([]match.Match, error) {
	var matches []match.Match
	for cursor.Next(ctx) {
		var doc matchDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode match: %w", err)
		}

		m, err := toMatchEntity(&doc)
		if err != nil {
			return nil, fmt.Errorf("convert match entity: %w", err)
		}
		matches = append(matches, *m)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return matches, nil
}
