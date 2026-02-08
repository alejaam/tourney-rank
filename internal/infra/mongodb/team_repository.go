package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/melisource/tourney-rank/internal/domain/team"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TeamRepository implements team.Repository using MongoDB.
type TeamRepository struct {
	collection *mongo.Collection
}

// NewTeamRepository creates a new MongoDB team repository.
func NewTeamRepository(db *mongo.Database) *TeamRepository {
	return &TeamRepository{
		collection: db.Collection("teams"),
	}
}

// EnsureIndexes creates necessary indexes for the teams collection.
func (r *TeamRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tournament_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "captain_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "member_ids", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "invite_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tournament_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tournament_id", Value: 1},
				{Key: "name", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("creating team indexes: %w", err)
	}

	return nil
}

// Create stores a new team.
func (r *TeamRepository) Create(ctx context.Context, t *team.Team) error {
	_, err := r.collection.InsertOne(ctx, t)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("team with this invite code already exists")
		}
		return fmt.Errorf("inserting team: %w", err)
	}
	return nil
}

// GetByID retrieves a team by its ID.
func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*team.Team, error) {
	var t team.Team
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, team.ErrNotFound
		}
		return nil, fmt.Errorf("finding team: %w", err)
	}
	return &t, nil
}

// GetByInviteCode retrieves a team by its invite code.
func (r *TeamRepository) GetByInviteCode(ctx context.Context, inviteCode string) (*team.Team, error) {
	var t team.Team
	err := r.collection.FindOne(ctx, bson.M{"invite_code": inviteCode}).Decode(&t)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, team.ErrNotFound
		}
		return nil, fmt.Errorf("finding team by invite code: %w", err)
	}
	return &t, nil
}

// Update updates an existing team.
func (r *TeamRepository) Update(ctx context.Context, t *team.Team) error {
	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": t.ID},
		t,
	)
	if err != nil {
		return fmt.Errorf("updating team: %w", err)
	}
	if result.MatchedCount == 0 {
		return team.ErrNotFound
	}
	return nil
}

// Delete removes a team by its ID.
func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting team: %w", err)
	}
	if result.DeletedCount == 0 {
		return team.ErrNotFound
	}
	return nil
}

// GetByTournamentID retrieves all teams for a tournament.
func (r *TeamRepository) GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]*team.Team, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"tournament_id": tournamentID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("finding teams by tournament: %w", err)
	}
	defer cursor.Close(ctx)

	var teams []*team.Team
	if err := cursor.All(ctx, &teams); err != nil {
		return nil, fmt.Errorf("decoding teams: %w", err)
	}

	return teams, nil
}

// GetByPlayerID retrieves teams where player is a member.
func (r *TeamRepository) GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*team.Team, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"member_ids": playerID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("finding teams by player: %w", err)
	}
	defer cursor.Close(ctx)

	var teams []*team.Team
	if err := cursor.All(ctx, &teams); err != nil {
		return nil, fmt.Errorf("decoding teams: %w", err)
	}

	return teams, nil
}

// GetPlayerTeamInTournament retrieves a player's team in a specific tournament.
func (r *TeamRepository) GetPlayerTeamInTournament(ctx context.Context, playerID, tournamentID uuid.UUID) (*team.Team, error) {
	var t team.Team
	err := r.collection.FindOne(
		ctx,
		bson.M{
			"tournament_id": tournamentID,
			"member_ids":    playerID,
		},
	).Decode(&t)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, team.ErrNotFound
		}
		return nil, fmt.Errorf("finding player team in tournament: %w", err)
	}
	return &t, nil
}

// CountByTournamentID returns the number of teams in a tournament.
func (r *TeamRepository) CountByTournamentID(ctx context.Context, tournamentID uuid.UUID) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"tournament_id": tournamentID})
	if err != nil {
		return 0, fmt.Errorf("counting teams: %w", err)
	}
	return count, nil
}

// List retrieves teams with optional filtering.
func (r *TeamRepository) List(ctx context.Context, filter team.ListFilter) ([]*team.Team, error) {
	// Build query filter
	query := bson.M{}

	if filter.TournamentID != nil {
		query["tournament_id"] = *filter.TournamentID
	}

	if filter.PlayerID != nil {
		query["member_ids"] = *filter.PlayerID
	}

	if filter.Status != nil {
		query["status"] = *filter.Status
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
		return nil, fmt.Errorf("listing teams: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var teams []*team.Team
	if err := cursor.All(ctx, &teams); err != nil {
		return nil, fmt.Errorf("decoding teams: %w", err)
	}

	return teams, nil
}
