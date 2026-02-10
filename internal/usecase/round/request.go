package round

import (
	"time"

	rounddomain "github.com/alejaam/tourney-rank/internal/domain/round"
	"github.com/google/uuid"
)

// CreateRoundRequest represents the request to create a new round.
type CreateRoundRequest struct {
	TournamentID uuid.UUID  `json:"tournament_id"`
	Number       int        `json:"number"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	MatchCount   int        `json:"match_count,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// UpdateRoundRequest represents the request to update a round.
type UpdateRoundRequest struct {
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	MatchCount   *int       `json:"match_count,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}

// UpdateRoundStatusRequest represents the request to update round status.
type UpdateRoundStatusRequest struct {
	Status rounddomain.Status `json:"status"`
}

// GetRoundsRequest represents the request to list rounds.
type GetRoundsRequest struct {
	TournamentID uuid.UUID
	Status       *rounddomain.Status
	Limit        int
	Offset       int
}
