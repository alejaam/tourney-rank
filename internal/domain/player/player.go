// Package player provides domain entities and logic for players and their statistics.
package player

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrNotFound is returned when a player is not found.
	ErrNotFound = errors.New("player not found")

	// ErrStatsNotFound is returned when player stats are not found.
	ErrStatsNotFound = errors.New("player stats not found")

	// ErrInvalidUsername is returned when username is empty or invalid.
	ErrInvalidUsername = errors.New("username cannot be empty")

	// ErrInvalidEmail is returned when email is empty or invalid.
	ErrInvalidEmail = errors.New("email cannot be empty")

	// ErrInvalidTier is returned when tier value is not recognized.
	ErrInvalidTier = errors.New("invalid tier value")

	// ErrInvalidBirthYear is returned when birth year is invalid.
	ErrInvalidBirthYear = errors.New("birth year must be between 1900 and current year")

	// ErrInvalidPlatform is returned when preferred platform is not recognized.
	ErrInvalidPlatform = errors.New("invalid preferred platform")
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

// Platform represents gaming platforms.
type Platform string

const (
	// PlatformPC represents PC/Desktop gaming.
	PlatformPC Platform = "PC"

	// PlatformPlayStation represents PlayStation consoles.
	PlatformPlayStation Platform = "PlayStation"

	// PlatformXbox represents Xbox consoles.
	PlatformXbox Platform = "Xbox"

	// PlatformNintendo represents Nintendo consoles.
	PlatformNintendo Platform = "Nintendo"

	// PlatformMobile represents mobile gaming.
	PlatformMobile Platform = "Mobile"

	// PlatformCrossplay represents cross-platform gaming.
	PlatformCrossplay Platform = "Crossplay"
)

// Player represents a player in the system.
type Player struct {
	ID                uuid.UUID         `bson:"_id" json:"id"`
	UserID            uuid.UUID         `bson:"user_id" json:"user_id"`
	DisplayName       string            `bson:"display_name" json:"display_name"`
	AvatarURL         string            `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	Bio               string            `bson:"bio,omitempty" json:"bio,omitempty"`
	PlatformIDs       map[string]string `bson:"platform_ids,omitempty" json:"platform_ids,omitempty"` // e.g., {"activision_id": "...", "epic_id": "..."}
	BirthYear         int               `bson:"birth_year,omitempty" json:"birth_year,omitempty"`
	Region            string            `bson:"region,omitempty" json:"region,omitempty"`
	PreferredPlatform string            `bson:"preferred_platform,omitempty" json:"preferred_platform,omitempty"`
	Language          string            `bson:"language,omitempty" json:"language,omitempty"`
	IsBanned          bool              `bson:"is_banned" json:"is_banned"`
	BannedAt          *time.Time        `bson:"banned_at,omitempty" json:"banned_at,omitempty"`
	CreatedAt         time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `bson:"updated_at" json:"updated_at"`
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

// UpdateExtendedProfile updates extended profile fields.
func (p *Player) UpdateExtendedProfile(birthYear int, region, preferredPlatform, language string) error {
	// Validate birth year if provided
	if birthYear > 0 {
		currentYear := time.Now().Year()
		if birthYear < 1900 || birthYear > currentYear {
			return ErrInvalidBirthYear
		}
		p.BirthYear = birthYear
	}

	// Validate preferred platform if provided
	if preferredPlatform != "" {
		if !isValidPlatform(preferredPlatform) {
			return ErrInvalidPlatform
		}
		p.PreferredPlatform = preferredPlatform
	}

	// Update other fields
	if region != "" {
		p.Region = region
	}
	if language != "" {
		p.Language = language
	}

	p.UpdatedAt = time.Now()
	return nil
}

// SetPlatformID sets a platform-specific ID for the player.
func (p *Player) SetPlatformID(platform, id string) {
	if p.PlatformIDs == nil {
		p.PlatformIDs = make(map[string]string)
	}
	p.PlatformIDs[platform] = id
	p.UpdatedAt = time.Now()
}

// Ban marks a player as banned.
func (p *Player) Ban() {
	now := time.Now().UTC()
	p.IsBanned = true
	p.BannedAt = &now
	p.UpdatedAt = now
}

// Unban removes the banned status from a player.
func (p *Player) Unban() {
	now := time.Now().UTC()
	p.IsBanned = false
	p.BannedAt = nil
	p.UpdatedAt = now
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

// isValidPlatform checks if a platform value is valid.
func isValidPlatform(platform string) bool {
	switch Platform(platform) {
	case PlatformPC, PlatformPlayStation, PlatformXbox, PlatformNintendo, PlatformMobile, PlatformCrossplay:
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
