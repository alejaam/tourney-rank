package match

import (
"errors"
"time"

"github.com/google/uuid"
)

// Status represents the verification status of a match report
type Status string

const (
StatusDraft    Status = "draft"    // Initial submission, pending verification
StatusVerified Status = "verified" // Admin approved the match report
StatusRejected Status = "rejected" // Admin rejected the match report
)

// PlayerMatchStats contains individual player performance in a match
type PlayerMatchStats struct {
	PlayerID    uuid.UUID              `bson:"player_id" json:"player_id"`
	Kills       int                    `bson:"kills" json:"kills"`
	Damage      int                    `bson:"damage" json:"damage"`
	Assists     int                    `bson:"assists" json:"assists"`
	Deaths      int                    `bson:"deaths" json:"deaths"`
	Downs       int                    `bson:"downs" json:"downs"`
	CustomStats map[string]interface{} `bson:"custom_stats" json:"custom_stats"`
}

// Match represents a tournament match result submission
type Match struct {
	ID              uuid.UUID           `bson:"_id" json:"id"`
	TournamentID    uuid.UUID           `bson:"tournament_id" json:"tournament_id"`
	TeamID          uuid.UUID           `bson:"team_id" json:"team_id"`
	GameID          uuid.UUID           `bson:"game_id" json:"game_id"`
	Status          Status              `bson:"status" json:"status"`
	TeamPlacement   int                 `bson:"team_placement" json:"team_placement"`
	TeamKills       int                 `bson:"team_kills" json:"team_kills"`
	PlayerStats     []PlayerMatchStats  `bson:"player_stats" json:"player_stats"`
	ScreenshotURL   string              `bson:"screenshot_url" json:"screenshot_url"`
	RejectionReason string              `bson:"rejection_reason,omitempty" json:"rejection_reason,omitempty"`
	SubmittedBy     uuid.UUID           `bson:"submitted_by" json:"submitted_by"` // Team captain who submitted
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updated_at"`
	VerifiedAt      *time.Time          `bson:"verified_at,omitempty" json:"verified_at,omitempty"`
	VerifiedBy      *uuid.UUID          `bson:"verified_by,omitempty" json:"verified_by,omitempty"`
}

// Error definitions
var (
ErrNotFound             = errors.New("match not found")
ErrInvalidPlacement     = errors.New("placement must be between 1 and 100")
ErrInvalidKills         = errors.New("kills cannot be negative")
ErrInvalidPlayerStats   = errors.New("invalid player stats in match")
ErrPlayerNotInTeam      = errors.New("player is not a member of the team")
ErrMissingPlayerStats   = errors.New("match must include stats for all team members")
ErrTeamSizeMismatch     = errors.New("number of players in stats does not match team size")
ErrInvalidStatus        = errors.New("invalid match status")
ErrAlreadyVerified      = errors.New("match has already been verified")
ErrMatchNotDraft        = errors.New("only draft matches can be verified")
ErrTournamentNotActive  = errors.New("tournament is not active")
ErrNotCaptain           = errors.New("player is not the team captain")
)

// NewMatch creates a new match with validation
func NewMatch(
tournamentID uuid.UUID,
teamID uuid.UUID,
gameID uuid.UUID,
teamPlacement int,
teamKills int,
playerStats []PlayerMatchStats,
screenshotURL string,
submittedBy uuid.UUID,
) (*Match, error) {
	if teamPlacement < 1 || teamPlacement > 100 {
		return nil, ErrInvalidPlacement
	}
	if teamKills < 0 {
		return nil, ErrInvalidKills
	}
	if len(playerStats) == 0 {
		return nil, ErrMissingPlayerStats
	}

	// Validate each player's stats
for _, ps := range playerStats {
if ps.Kills < 0 || ps.Damage < 0 || ps.Assists < 0 || ps.Deaths < 0 || ps.Downs < 0 {
return nil, ErrInvalidPlayerStats
}
}

now := time.Now()
return &Match{
ID:            uuid.New(),
TournamentID:  tournamentID,
TeamID:        teamID,
GameID:        gameID,
Status:        StatusDraft,
TeamPlacement: teamPlacement,
TeamKills:     teamKills,
PlayerStats:   playerStats,
ScreenshotURL: screenshotURL,
SubmittedBy:   submittedBy,
CreatedAt:     now,
UpdatedAt:     now,
}, nil
}

// VerifyMatch marks a match as verified by an admin
func (m *Match) VerifyMatch(adminID uuid.UUID) error {
if m.Status != StatusDraft {
return ErrMatchNotDraft
}
now := time.Now()
m.Status = StatusVerified
m.VerifiedAt = &now
m.VerifiedBy = &adminID
m.UpdatedAt = now
m.RejectionReason = ""
return nil
}

// RejectMatch marks a match as rejected with a reason
func (m *Match) RejectMatch(adminID uuid.UUID, reason string) error {
if m.Status != StatusDraft {
return ErrMatchNotDraft
}
now := time.Now()
m.Status = StatusRejected
m.VerifiedAt = &now
m.VerifiedBy = &adminID
m.UpdatedAt = now
m.RejectionReason = reason
return nil
}

// GetTotalTeamKills calculates team kills from player stats
func (m *Match) GetTotalTeamKills() int {
total := 0
for _, ps := range m.PlayerStats {
total += ps.Kills
}
return total
}

// GetTeamKDRatio calculates kill/death ratio for the team
func (m *Match) GetTeamKDRatio() float64 {
totalDeaths := 0
for _, ps := range m.PlayerStats {
totalDeaths += ps.Deaths
}
if totalDeaths == 0 {
return float64(m.TeamKills)
}
return float64(m.TeamKills) / float64(totalDeaths)
}

// IsVerified checks if the match has been approved
func (m *Match) IsVerified() bool {
return m.Status == StatusVerified
}
