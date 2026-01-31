// Package player provides domain entities and logic for players and their statistics.
package player

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidUsername is returned when username is empty or invalid.
	ErrInvalidUsername = errors.New("username cannot be empty")

	// ErrInvalidEmail is returned when email is empty or invalid.
	ErrInvalidEmail = errors.New("email cannot be empty")

	// ErrInvalidTier is returned when tier value is not recognized.
	ErrInvalidTier = errors.New("invalid tier value")
)

// Tier represents player skill level.
type Tier string

const (
	// TierElite represents top 5% players.
	TierElite Tier = "elite"

	// TierAdvanced represents top 20% players.
	TierAdvanced Tier = "advanced"

	// TierIntermediate represents top 50% players.
	TierIntermediate Tier = "intermediate"

	// TierBeginner represents remaining players.
	TierBeginner Tier = "beginner"
)

// Player represents a player in the system.
type Player struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	DisplayName string
	AvatarURL   string
	Bio         string
	PlatformIDs map[string]string // e.g., {"activision_id": "...", "epic_id": "..."}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PlayerStats represents a player's statistics for a specific game.
type PlayerStats struct {
	ID            uuid.UUID
	PlayerID      uuid.UUID
	GameID        uuid.UUID
	Stats         map[string]interface{} // Flexible stats storage
	MatchesPlayed int
	RankingScore  float64
	Tier          Tier
	LastMatchAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewPlayer creates a new Player instance.
func NewPlayer(userID uuid.UUID, displayName string) (*Player, error) {
	if displayName == "" {
		return nil, ErrInvalidUsername
	}

	return &Player{
		ID:          uuid.New(),
		UserID:      userID,
		DisplayName: displayName,
		PlatformIDs: make(map[string]string),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

// UpdateProfile updates player profile information.
func (p *Player) UpdateProfile(displayName, avatarURL, bio string) {
	if displayName != "" {
		p.DisplayName = displayName
	}
	if avatarURL != "" {
		p.AvatarURL = avatarURL
	}
	p.Bio = bio
	p.UpdatedAt = time.Now()
}

// SetPlatformID sets a platform-specific ID for the player.
func (p *Player) SetPlatformID(platform, id string) {
	if p.PlatformIDs == nil {
		p.PlatformIDs = make(map[string]string)
	}
	p.PlatformIDs[platform] = id
	p.UpdatedAt = time.Now()
}

// GetPlatformID retrieves a platform-specific ID.
func (p *Player) GetPlatformID(platform string) (string, bool) {
	id, exists := p.PlatformIDs[platform]
	return id, exists
}

// NewPlayerStats creates a new PlayerStats instance.
func NewPlayerStats(playerID, gameID uuid.UUID) *PlayerStats {
	return &PlayerStats{
		ID:            uuid.New(),
		PlayerID:      playerID,
		GameID:        gameID,
		Stats:         make(map[string]interface{}),
		MatchesPlayed: 0,
		RankingScore:  0.0,
		Tier:          TierBeginner,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// UpdateStats updates player statistics after a match.
func (ps *PlayerStats) UpdateStats(newStats map[string]interface{}) {
	for key, value := range newStats {
		ps.Stats[key] = value
	}
	ps.MatchesPlayed++
	ps.UpdatedAt = time.Now()
	now := time.Now()
	ps.LastMatchAt = &now
}

// UpdateRankingScore updates the calculated ranking score and tier.
func (ps *PlayerStats) UpdateRankingScore(score float64, tier Tier) error {
	if !isValidTier(tier) {
		return ErrInvalidTier
	}

	ps.RankingScore = score
	ps.Tier = tier
	ps.UpdatedAt = time.Now()
	return nil
}

// GetStat retrieves a specific stat value.
func (ps *PlayerStats) GetStat(key string) (interface{}, bool) {
	val, exists := ps.Stats[key]
	return val, exists
}

// GetStatAsFloat retrieves a stat value as float64.
func (ps *PlayerStats) GetStatAsFloat(key string) float64 {
	val, exists := ps.Stats[key]
	if !exists {
		return 0.0
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0.0
	}
}

// GetStatAsInt retrieves a stat value as int.
func (ps *PlayerStats) GetStatAsInt(key string) int {
	val, exists := ps.Stats[key]
	if !exists {
		return 0
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// CalculateKDRatio calculates kill/death ratio if applicable.
func (ps *PlayerStats) CalculateKDRatio() float64 {
	kills := ps.GetStatAsFloat("kills")
	deaths := ps.GetStatAsFloat("deaths")

	if deaths == 0 {
		return kills
	}

	return kills / deaths
}

// isValidTier checks if a tier value is valid.
func isValidTier(tier Tier) bool {
	switch tier {
	case TierElite, TierAdvanced, TierIntermediate, TierBeginner:
		return true
	default:
		return false
	}
}

// DetermineTierByPercentile determines tier based on percentile ranking.
func DetermineTierByPercentile(percentile float64) Tier {
	switch {
	case percentile >= 95.0:
		return TierElite
	case percentile >= 80.0:
		return TierAdvanced
	case percentile >= 50.0:
		return TierIntermediate
	default:
		return TierBeginner
	}
}
