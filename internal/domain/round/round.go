// Package round provides domain entities and logic for tournament rounds.
package round

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound              = errors.New("round not found")
	ErrInvalidRoundNum       = errors.New("invalid round number")
	ErrInvalidStatus         = errors.New("invalid round status")
	ErrCannotRetrocedeStatus = errors.New("cannot move round to a previous status")
	ErrNoMatches             = errors.New("no matches in round")
	ErrMatchesIncomplete     = errors.New("not all matches are completed")
)

// Status enumeration for round states.
type Status string

const (
	StatusPending   Status = "pending"
	StatusOngoing   Status = "ongoing"
	StatusCompleted Status = "completed"
	StatusCanceled  Status = "canceled"
)

func ValidStatuses() []Status {
	return []Status{StatusPending, StatusOngoing, StatusCompleted, StatusCanceled}
}

func (s Status) IsValid() bool {
	for _, valid := range ValidStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// Round represents a tournament round containing multiple matches.
type Round struct {
	ID           uuid.UUID  `bson:"_id" json:"id"`
	TournamentID uuid.UUID  `bson:"tournament_id" json:"tournament_id"`
	Number       int        `bson:"number" json:"number"`
	Status       Status     `bson:"status" json:"status"`
	ScheduledFor time.Time  `bson:"scheduled_for" json:"scheduled_for"`
	StartedAt    *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt  *time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	MatchCount   int        `bson:"match_count" json:"match_count"`
	Notes        string     `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt    time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at" json:"updated_at"`
}

// NewRound creates a new round for a tournament.
func NewRound(tournamentID uuid.UUID, number int, scheduledFor time.Time) (*Round, error) {
	if number < 1 {
		return nil, ErrInvalidRoundNum
	}

	now := time.Now().UTC()
	return &Round{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		Number:       number,
		Status:       StatusPending,
		ScheduledFor: scheduledFor,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Start transitions the round from Pending to Ongoing.
func (r *Round) Start() error {
	if r.Status != StatusPending {
		return ErrInvalidStatus
	}
	now := time.Now().UTC()
	r.Status = StatusOngoing
	r.StartedAt = &now
	r.UpdatedAt = now
	return nil
}

// Complete transitions the round from Ongoing to Completed.
func (r *Round) Complete() error {
	if r.Status != StatusOngoing {
		return ErrInvalidStatus
	}
	now := time.Now().UTC()
	r.Status = StatusCompleted
	r.CompletedAt = &now
	r.UpdatedAt = now
	return nil
}

// Cancel transitions the round to Canceled (can be done from any active state).
func (r *Round) Cancel() error {
	if r.Status == StatusCanceled || r.Status == StatusCompleted {
		return ErrCannotRetrocedeStatus
	}
	now := time.Now().UTC()
	r.Status = StatusCanceled
	r.UpdatedAt = now
	return nil
}

// IsActive returns true if the round is currently happening.
func (r *Round) IsActive() bool {
	return r.Status == StatusOngoing
}

// IsCompleted returns true if all matches have been played.
func (r *Round) IsCompleted() bool {
	return r.Status == StatusCompleted
}

// SetMatchCount updates the total number of matches in this round.
func (r *Round) SetMatchCount(count int) error {
	if count < 1 {
		return ErrNoMatches
	}
	r.MatchCount = count
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// SetNotes sets optional notes for the round (e.g., announcements, updates).
func (r *Round) SetNotes(notes string) {
	r.Notes = notes
	r.UpdatedAt = time.Now().UTC()
}
