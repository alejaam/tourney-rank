// Package team provides domain entities and logic for team management.
package team

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound            = errors.New("team not found")
	ErrInvalidName         = errors.New("team name cannot be empty")
	ErrInvalidCaptain      = errors.New("captain ID cannot be empty")
	ErrPlayerAlreadyInTeam = errors.New("player is already in the team")
	ErrPlayerNotInTeam     = errors.New("player is not in the team")
	ErrTeamFull            = errors.New("team is full")
	ErrNotCaptain          = errors.New("only captain can perform this action")
	ErrCannotRemoveCaptain = errors.New("cannot remove captain from team")
	ErrInvalidInviteCode   = errors.New("invalid invite code")
	ErrTeamNotReady        = errors.New("team is not ready")
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusReady      Status = "ready"
	StatusActive     Status = "active"
	StatusEliminated Status = "eliminated"
	StatusDisbanded  Status = "disbanded"
)

type Team struct {
	ID           uuid.UUID   `bson:"_id" json:"id"`
	TournamentID uuid.UUID   `bson:"tournament_id" json:"tournament_id"`
	Name         string      `bson:"name" json:"name"`
	Tag          string      `bson:"tag,omitempty" json:"tag,omitempty"`
	CaptainID    uuid.UUID   `bson:"captain_id" json:"captain_id"`
	MemberIDs    []uuid.UUID `bson:"member_ids" json:"member_ids"`
	Status       Status      `bson:"status" json:"status"`
	InviteCode   string      `bson:"invite_code" json:"invite_code"`
	LogoURL      string      `bson:"logo_url,omitempty" json:"logo_url,omitempty"`
	CreatedAt    time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `bson:"updated_at" json:"updated_at"`
}

func NewTeam(tournamentID, captainID uuid.UUID, name string) (*Team, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if captainID == uuid.Nil {
		return nil, ErrInvalidCaptain
	}

	now := time.Now().UTC()
	return &Team{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		Name:         name,
		CaptainID:    captainID,
		MemberIDs:    []uuid.UUID{captainID},
		Status:       StatusPending,
		InviteCode:   generateInviteCode(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (t *Team) AddMember(playerID uuid.UUID) error {
	if t.HasMember(playerID) {
		return ErrPlayerAlreadyInTeam
	}

	t.MemberIDs = append(t.MemberIDs, playerID)
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (t *Team) RemoveMember(playerID uuid.UUID) error {
	if !t.HasMember(playerID) {
		return ErrPlayerNotInTeam
	}
	if playerID == t.CaptainID {
		return ErrCannotRemoveCaptain
	}

	newMembers := make([]uuid.UUID, 0, len(t.MemberIDs)-1)
	for _, id := range t.MemberIDs {
		if id != playerID {
			newMembers = append(newMembers, id)
		}
	}
	t.MemberIDs = newMembers
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (t *Team) HasMember(playerID uuid.UUID) bool {
	for _, id := range t.MemberIDs {
		if id == playerID {
			return true
		}
	}
	return false
}

func (t *Team) IsCaptain(playerID uuid.UUID) bool {
	return t.CaptainID == playerID
}

func (t *Team) TransferCaptaincy(newCaptainID uuid.UUID) error {
	if !t.HasMember(newCaptainID) {
		return ErrPlayerNotInTeam
	}

	t.CaptainID = newCaptainID
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (t *Team) UpdateStatus(newStatus Status) error {
	t.Status = newStatus
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (t *Team) SetTag(tag string) {
	t.Tag = tag
	t.UpdatedAt = time.Now().UTC()
}

func (t *Team) SetLogoURL(logoURL string) {
	t.LogoURL = logoURL
	t.UpdatedAt = time.Now().UTC()
}

func (t *Team) RegenerateInviteCode() {
	t.InviteCode = generateInviteCode()
	t.UpdatedAt = time.Now().UTC()
}

func (t *Team) IsReady() bool {
	return t.Status == StatusReady || t.Status == StatusActive
}

func (t *Team) MemberCount() int {
	return len(t.MemberIDs)
}

func generateInviteCode() string {
	return uuid.New().String()[:8]
}
