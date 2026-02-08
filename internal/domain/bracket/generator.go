package bracket

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Generator provides methods to generate tournament brackets.
type Generator struct {
	rng *rand.Rand
}

// NewGenerator creates a new bracket generator.
func NewGenerator() *Generator {
	return &Generator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// TeamSeed represents a team with its seed position.
type TeamSeed struct {
	TeamID uuid.UUID
	Seed   int
}

// GenerateSingleEliminationMatchups generates matchups for single elimination bracket.
func (g *Generator) GenerateSingleEliminationMatchups(bracket *Bracket, teams []TeamSeed) ([]*Matchup, error) {
	if len(teams) < 2 {
		return nil, ErrInsufficientTeams
	}

	// Shuffle teams if not seeded
	if !bracket.IsSeeded {
		g.shuffle(teams)
	}

	// Calculate number of first round matchups
	numTeams := len(teams)
	nextPowerOf2 := nextPowerOfTwo(numTeams)
	numByes := nextPowerOf2 - numTeams

	matchups := make([]*Matchup, 0)
	matchNumber := 1

	// First round matchups
	teamIdx := 0
	for i := 0; i < numTeams-numByes; i += 2 {
		team1ID := teams[teamIdx].TeamID
		team2ID := teams[teamIdx+1].TeamID
		matchup := NewMatchup(bracket.ID, 1, matchNumber, &team1ID, &team2ID)
		matchups = append(matchups, matchup)
		matchNumber++
		teamIdx += 2
	}

	// Add byes for remaining teams
	for i := 0; i < numByes; i++ {
		if teamIdx < len(teams) {
			teamID := teams[teamIdx].TeamID
			matchup := NewMatchup(bracket.ID, 1, matchNumber, &teamID, nil)
			matchup.WinnerID = &teamID // Bye matches auto-advance
			matchup.Status = MatchupStatusCompleted
			matchups = append(matchups, matchup)
			matchNumber++
			teamIdx++
		}
	}

	return matchups, nil
}

// GenerateRoundRobinMatchups generates matchups for round-robin format (all vs all).
func (g *Generator) GenerateRoundRobinMatchups(bracket *Bracket, teams []TeamSeed) ([]*Matchup, error) {
	if len(teams) < 2 {
		return nil, ErrInsufficientTeams
	}

	matchups := make([]*Matchup, 0)
	matchNumber := 1
	round := 1

	// Generate all possible matchups
	for i := 0; i < len(teams); i++ {
		for j := i + 1; j < len(teams); j++ {
			team1ID := teams[i].TeamID
			team2ID := teams[j].TeamID
			matchup := NewMatchup(bracket.ID, round, matchNumber, &team1ID, &team2ID)
			matchups = append(matchups, matchup)
			matchNumber++

			// Distribute matchups across rounds for better scheduling
			if matchNumber%((len(teams)-1)/2+1) == 0 {
				round++
			}
		}
	}

	return matchups, nil
}

// GenerateNextRoundMatchups generates matchups for the next round based on previous winners.
func (g *Generator) GenerateNextRoundMatchups(bracket *Bracket, previousRoundMatchups []*Matchup) ([]*Matchup, error) {
	// Collect winners from previous round
	winners := make([]uuid.UUID, 0)
	for _, matchup := range previousRoundMatchups {
		if matchup.Status == MatchupStatusCompleted && matchup.WinnerID != nil {
			winners = append(winners, *matchup.WinnerID)
		}
	}

	if len(winners) < 2 {
		return nil, errors.New("insufficient winners to generate next round")
	}

	// Generate matchups for next round
	matchups := make([]*Matchup, 0)
	matchNumber := 1
	nextRound := bracket.CurrentRound + 1

	for i := 0; i < len(winners); i += 2 {
		if i+1 < len(winners) {
			team1ID := winners[i]
			team2ID := winners[i+1]
			matchup := NewMatchup(bracket.ID, nextRound, matchNumber, &team1ID, &team2ID)
			matchups = append(matchups, matchup)
			matchNumber++
		} else {
			// Bye if odd number of winners
			team1ID := winners[i]
			matchup := NewMatchup(bracket.ID, nextRound, matchNumber, &team1ID, nil)
			matchup.WinnerID = &team1ID
			matchup.Status = MatchupStatusCompleted
			matchups = append(matchups, matchup)
			matchNumber++
		}
	}

	return matchups, nil
}

// shuffle randomizes the order of teams.
func (g *Generator) shuffle(teams []TeamSeed) {
	g.rng.Shuffle(len(teams), func(i, j int) {
		teams[i], teams[j] = teams[j], teams[i]
	})
}

// nextPowerOfTwo finds the next power of 2 greater than or equal to n.
func nextPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

// CalculateTotalRounds calculates the total number of rounds needed.
func CalculateTotalRounds(format Format, numTeams int) int {
	switch format {
	case FormatSingleElimination:
		return int(math.Ceil(math.Log2(float64(numTeams))))
	case FormatDoubleElimination:
		return int(math.Ceil(math.Log2(float64(numTeams)))) * 2
	case FormatRoundRobin:
		// Each team plays every other team once
		if numTeams%2 == 0 {
			return numTeams - 1
		}
		return numTeams
	case FormatSwiss:
		// Swiss system typically runs log2(n) + 1 rounds
		return int(math.Ceil(math.Log2(float64(numTeams)))) + 1
	default:
		return 1
	}
}
