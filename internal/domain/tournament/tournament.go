// Package tournament provides domain entities and logic for tournament management.
package tournament

import (
"errors"
"time"

"github.com/google/uuid"
)

var (
ErrNotFound = errors.New("tournament not found")
ErrInvalidName = errors.New("tournament name cannot be empty")
ErrInvalidTeamSize = errors.New("invalid team size")
ErrInvalidStatus = errors.New("invalid tournament status")
ErrInvalidDates = errors.New("start date must be before end date")
ErrTournamentNotActive = errors.New("tournament is not active")
ErrRegistrationClosed = errors.New("tournament registration is closed")
)

type Status string

const (
StatusDraft Status = "draft"
StatusOpen Status = "open"
StatusActive Status = "active"
StatusFinished Status = "finished"
StatusCanceled Status = "canceled"
)

func ValidStatuses() []Status {
	return []Status{StatusDraft, StatusOpen, StatusActive, StatusFinished, StatusCanceled}
}

func (s Status) IsValid() bool {
	for _, valid := range ValidStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

type TeamSize int

const (
TeamSizeSolo TeamSize = 1
TeamSizeDuos TeamSize = 2
TeamSizeTrios TeamSize = 3
TeamSizeQuads TeamSize = 4
)

func ValidTeamSizes() []TeamSize {
	return []TeamSize{TeamSizeSolo, TeamSizeDuos, TeamSizeTrios, TeamSizeQuads}
}

func (ts TeamSize) IsValid() bool {
	for _, valid := range ValidTeamSizes() {
		if ts == valid {
			return true
		}
	}
	return false
}

func (ts TeamSize) String() string {
	switch ts {
	case TeamSizeSolo:
		return "solo"
	case TeamSizeDuos:
		return "duos"
	case TeamSizeTrios:
		return "trios"
	case TeamSizeQuads:
		return "quads"
	default:
		return "unknown"
	}
}

type Rules struct {
	MaxTeams int `bson:"max_teams" json:"max_teams"`
	MinMatches int `bson:"min_matches" json:"min_matches"`
	MaxMatches int `bson:"max_matches" json:"max_matches"`
	RequireVerification bool `bson:"require_verification" json:"require_verification"`
	AllowLateRegistration bool `bson:"allow_late_registration" json:"allow_late_registration"`
	RegistrationDeadline *time.Time `bson:"registration_deadline,omitempty" json:"registration_deadline,omitempty"`
}

type Tournament struct {
	ID uuid.UUID `bson:"_id" json:"id"`
	GameID uuid.UUID `bson:"game_id" json:"game_id"`
	Name string `bson:"name" json:"name"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	TeamSize TeamSize `bson:"team_size" json:"team_size"`
	Status Status `bson:"status" json:"status"`
	Rules Rules `bson:"rules" json:"rules"`
	StartDate time.Time `bson:"start_date" json:"start_date"`
	EndDate time.Time `bson:"end_date" json:"end_date"`
	PrizePool string `bson:"prize_pool,omitempty" json:"prize_pool,omitempty"`
	BannerURL string `bson:"banner_url,omitempty" json:"banner_url,omitempty"`
	CreatedBy uuid.UUID `bson:"created_by" json:"created_by"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func NewTournament(gameID, createdBy uuid.UUID, name string, teamSize TeamSize, startDate, endDate time.Time) (*Tournament, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if !teamSize.IsValid() {
		return nil, ErrInvalidTeamSize
	}
	if !startDate.Before(endDate) {
		return nil, ErrInvalidDates
	}
	now := time.Now().UTC()
	return &Tournament{
		ID: uuid.New(),
		GameID: gameID,
		Name: name,
		TeamSize: teamSize,
		Status: StatusDraft,
		StartDate: startDate,
		EndDate: endDate,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
		Rules: Rules{
			MaxTeams: 0,
			MinMatches: 1,
			MaxMatches: 0,
			RequireVerification: false,
			AllowLateRegistration: true,
		},
	}, nil
}

func (t *Tournament) UpdateStatus(newStatus Status) error {
	if !newStatus.IsValid() {
		return ErrInvalidStatus
	}
	switch t.Status {
	case StatusDraft:
		if newStatus != StatusOpen && newStatus != StatusCanceled {
			return ErrInvalidStatus
		}
	case StatusOpen:
		if newStatus != StatusActive && newStatus != StatusCanceled {
			return ErrInvalidStatus
		}
	case StatusActive:
		if newStatus != StatusFinished && newStatus != StatusCanceled {
			return ErrInvalidStatus
		}
	case StatusFinished, StatusCanceled:
		return ErrInvalidStatus
	}
	t.Status = newStatus
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (t *Tournament) IsActive() bool {
	return t.Status == StatusActive
}

func (t *Tournament) IsAcceptingRegistrations() bool {
	if t.Status != StatusOpen {
		return false
	}
	if t.Rules.RegistrationDeadline != nil && time.Now().UTC().After(*t.Rules.RegistrationDeadline) {
		return false
	}
	return true
}

func (t *Tournament) CanAcceptLateRegistration() bool {
	return t.Status == StatusActive && t.Rules.AllowLateRegistration
}

func (t *Tournament) SetRules(rules Rules) {
	t.Rules = rules
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tournament) SetDescription(description string) {
	t.Description = description
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tournament) SetPrizePool(prizePool string) {
	t.PrizePool = prizePool
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tournament) SetBannerURL(bannerURL string) {
	t.BannerURL = bannerURL
	t.UpdatedAt = time.Now().UTC()
}
