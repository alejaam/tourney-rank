package team

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for team persistence operations.
type Repository interface {
	// Create stores a new team.
	Create(ctx context.Context, team *Team) error

	// GetByID retrieves a team by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Team, error)

	// GetByInviteCode retrieves a team by its invite code.
	GetByInviteCode(ctx context.Context, inviteCode string) (*Team, error)

	// Update updates an existing team.
	Update(ctx context.Context, team *Team) error

	// Delete removes a team by its ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByTournamentID retrieves all teams for a tournament.
	GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]*Team, error)

	// GetByPlayerID retrieves teams where player is a member.
	GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*Team, error)

	// GetPlayerTeamInTournament retrieves a player's team in a specific tournament.
	GetPlayerTeamInTournament(ctx context.Context, playerID, tournamentID uuid.UUID) (*Team, error)

	// CountByTournamentID returns the number of teams in a tournament.
	CountByTournamentID(ctx context.Context, tournamentID uuid.UUID) (int64, error)

	// List retrieves teams with optional filtering.
	List(ctx context.Context, filter ListFilter) ([]*Team, error)
}

// ListFilter defines filtering options for listing teams.
type ListFilter struct {
	// TournamentID filters by tournament (optional).
	TournamentID *uuid.UUID

	// PlayerID filters teams where player is a member (optional).
	PlayerID *uuid.UUID

	// Status filters by team status (optional).
	Status *Status

	// Limit is the maximum number of results to return.
	Limit int

	// Offset is the number of results to skip.
	Offset int
}
