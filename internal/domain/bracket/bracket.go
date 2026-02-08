// Package bracket provides domain entities and logic for tournament bracket management.
package bracket

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound             = errors.New("bracket not found")
	ErrInvalidFormat        = errors.New("invalid bracket format")
	ErrInvalidRound         = errors.New("invalid round number")
	ErrMatchupNotFound      = errors.New("matchup not found")
	ErrMatchupAlreadyPlayed = errors.New("matchup already played")
	ErrInsufficientTeams    = errors.New("insufficient teams to generate bracket")
	ErrBracketAlreadyExists = errors.New("bracket already exists for tournament")
)

// Format represents the bracket format type.
type Format string

const (
	FormatSingleElimination Format = "single_elimination"
	FormatDoubleElimination Format = "double_elimination"
	FormatRoundRobin        Format = "round_robin"
	FormatSwiss             Format = "swiss"
)

// ValidFormats returns all valid bracket formats.
func ValidFormats() []Format {
	return []Format{FormatSingleElimination, FormatDoubleElimination, FormatRoundRobin, FormatSwiss}
}

// IsValid checks if the format is valid.
func (f Format) IsValid() bool {
	for _, valid := range ValidFormats() {
		if f == valid {
			return true
		}
	}
	return false
}

// MatchupStatus represents the status of a matchup.
type MatchupStatus string

const (
	MatchupStatusPending    MatchupStatus = "pending"
	MatchupStatusInProgress MatchupStatus = "in_progress"
	MatchupStatusCompleted  MatchupStatus = "completed"
	MatchupStatusCanceled   MatchupStatus = "canceled"
)

// Matchup represents a single match between two teams in a bracket.
type Matchup struct {
	ID          uuid.UUID     `bson:"_id" json:"id"`
	BracketID   uuid.UUID     `bson:"bracket_id" json:"bracket_id"`
	Round       int           `bson:"round" json:"round"`
	MatchNumber int           `bson:"match_number" json:"match_number"`
	Team1ID     *uuid.UUID    `bson:"team1_id,omitempty" json:"team1_id,omitempty"`
	Team2ID     *uuid.UUID    `bson:"team2_id,omitempty" json:"team2_id,omitempty"`
	WinnerID    *uuid.UUID    `bson:"winner_id,omitempty" json:"winner_id,omitempty"`
	Status      MatchupStatus `bson:"status" json:"status"`
	ScheduledAt *time.Time    `bson:"scheduled_at,omitempty" json:"scheduled_at,omitempty"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

// NewMatchup creates a new matchup.
func NewMatchup(bracketID uuid.UUID, round, matchNumber int, team1ID, team2ID *uuid.UUID) *Matchup {
	now := time.Now().UTC()
	return &Matchup{
		ID:          uuid.New(),
		BracketID:   bracketID,
		Round:       round,
		MatchNumber: matchNumber,
		Team1ID:     team1ID,
		Team2ID:     team2ID,
		Status:      MatchupStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// SetWinner sets the winner of the matchup.
func (m *Matchup) SetWinner(winnerID uuid.UUID) error {
	if m.Status == MatchupStatusCompleted {
		return ErrMatchupAlreadyPlayed
	}

	// Validate winner is one of the participants
	if m.Team1ID != nil && *m.Team1ID == winnerID {
		m.WinnerID = &winnerID
	} else if m.Team2ID != nil && *m.Team2ID == winnerID {
		m.WinnerID = &winnerID
	} else {
		return errors.New("winner must be one of the participating teams")
	}

	m.Status = MatchupStatusCompleted
	m.UpdatedAt = time.Now().UTC()
	return nil
}

// IsBye returns true if this is a bye match (only one team).
func (m *Matchup) IsBye() bool {
	return (m.Team1ID == nil && m.Team2ID != nil) || (m.Team1ID != nil && m.Team2ID == nil)
}

// Bracket represents a tournament bracket structure.
type Bracket struct {
	ID           uuid.UUID `bson:"_id" json:"id"`
	TournamentID uuid.UUID `bson:"tournament_id" json:"tournament_id"`
	Format       Format    `bson:"format" json:"format"`
	TotalRounds  int       `bson:"total_rounds" json:"total_rounds"`
	CurrentRound int       `bson:"current_round" json:"current_round"`
	IsSeeded     bool      `bson:"is_seeded" json:"is_seeded"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}

// NewBracket creates a new bracket.
func NewBracket(tournamentID uuid.UUID, format Format, totalRounds int, isSeeded bool) (*Bracket, error) {
	if !format.IsValid() {
		return nil, ErrInvalidFormat
	}
	if totalRounds < 1 {
		return nil, ErrInvalidRound
	}

	now := time.Now().UTC()
	return &Bracket{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		Format:       format,
		TotalRounds:  totalRounds,
		CurrentRound: 1,
		IsSeeded:     isSeeded,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// AdvanceRound advances the bracket to the next round.
func (b *Bracket) AdvanceRound() error {
	if b.CurrentRound >= b.TotalRounds {
		return errors.New("bracket already at final round")
	}
	b.CurrentRound++
	b.UpdatedAt = time.Now().UTC()
	return nil
}

// IsComplete returns true if all rounds are completed.
func (b *Bracket) IsComplete() bool {
	return b.CurrentRound >= b.TotalRounds
}
