package leaderboard

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/alejaam/tourney-rank/internal/domain/match"
	"github.com/alejaam/tourney-rank/internal/domain/player"
	"github.com/alejaam/tourney-rank/internal/domain/team"
	"github.com/alejaam/tourney-rank/internal/domain/tournament"
)

type leaderboardTournamentRepo struct{ tournament *tournament.Tournament }

func (r leaderboardTournamentRepo) Create(context.Context, *tournament.Tournament) error { return nil }
func (r leaderboardTournamentRepo) GetByID(_ context.Context, id uuid.UUID) (*tournament.Tournament, error) {
	if r.tournament != nil && r.tournament.ID == id {
		return r.tournament, nil
	}
	return nil, tournament.ErrNotFound
}
func (r leaderboardTournamentRepo) Update(context.Context, *tournament.Tournament) error { return nil }
func (r leaderboardTournamentRepo) Delete(context.Context, uuid.UUID) error              { return nil }
func (r leaderboardTournamentRepo) List(context.Context, tournament.ListFilter) ([]*tournament.Tournament, error) {
	return nil, nil
}
func (r leaderboardTournamentRepo) GetByGameID(context.Context, uuid.UUID) ([]*tournament.Tournament, error) {
	return nil, nil
}
func (r leaderboardTournamentRepo) GetByStatus(context.Context, tournament.Status) ([]*tournament.Tournament, error) {
	return nil, nil
}
func (r leaderboardTournamentRepo) GetActiveTournaments(context.Context) ([]*tournament.Tournament, error) {
	return nil, nil
}
func (r leaderboardTournamentRepo) CountByGameID(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

type leaderboardTeamRepo struct{ teams []*team.Team }

func (r leaderboardTeamRepo) Create(context.Context, *team.Team) error { return nil }
func (r leaderboardTeamRepo) GetByID(context.Context, uuid.UUID) (*team.Team, error) {
	return nil, team.ErrNotFound
}
func (r leaderboardTeamRepo) GetByInviteCode(context.Context, string) (*team.Team, error) {
	return nil, team.ErrNotFound
}
func (r leaderboardTeamRepo) Update(context.Context, *team.Team) error { return nil }
func (r leaderboardTeamRepo) Delete(context.Context, uuid.UUID) error  { return nil }
func (r leaderboardTeamRepo) GetByTournamentID(context.Context, uuid.UUID) ([]*team.Team, error) {
	return r.teams, nil
}
func (r leaderboardTeamRepo) GetByPlayerID(context.Context, uuid.UUID) ([]*team.Team, error) {
	return nil, nil
}
func (r leaderboardTeamRepo) GetPlayerTeamInTournament(context.Context, uuid.UUID, uuid.UUID) (*team.Team, error) {
	return nil, team.ErrNotFound
}
func (r leaderboardTeamRepo) CountByTournamentID(context.Context, uuid.UUID) (int64, error) {
	return int64(len(r.teams)), nil
}
func (r leaderboardTeamRepo) List(context.Context, team.ListFilter) ([]*team.Team, error) {
	return nil, nil
}

type leaderboardMatchRepo struct{ matches []match.Match }

func (r leaderboardMatchRepo) Create(context.Context, *match.Match) error { return nil }
func (r leaderboardMatchRepo) GetByID(context.Context, string) (*match.Match, error) {
	return nil, match.ErrNotFound
}
func (r leaderboardMatchRepo) GetByTournament(_ context.Context, _ string, limit, offset int) ([]match.Match, error) {
	if offset >= len(r.matches) {
		return nil, nil
	}
	end := offset + limit
	if end > len(r.matches) {
		end = len(r.matches)
	}
	return r.matches[offset:end], nil
}
func (r leaderboardMatchRepo) GetByTeam(context.Context, string, int, int) ([]match.Match, error) {
	return nil, nil
}
func (r leaderboardMatchRepo) GetByPlayer(context.Context, string, int, int) ([]match.Match, error) {
	return nil, nil
}
func (r leaderboardMatchRepo) GetUnverified(context.Context, int, int) ([]match.Match, error) {
	return nil, nil
}
func (r leaderboardMatchRepo) GetTournamentUnverified(context.Context, string, int, int) ([]match.Match, error) {
	return nil, nil
}
func (r leaderboardMatchRepo) Update(context.Context, *match.Match) error { return nil }
func (r leaderboardMatchRepo) CountByTournament(context.Context, string) (int, error) {
	return len(r.matches), nil
}
func (r leaderboardMatchRepo) CountUnverified(context.Context) (int, error) { return 0, nil }
func (r leaderboardMatchRepo) DeleteByID(context.Context, string) error     { return nil }

type leaderboardPlayerRepo struct{ players map[uuid.UUID]*player.Player }

func (r leaderboardPlayerRepo) Create(context.Context, *player.Player) error { return nil }
func (r leaderboardPlayerRepo) GetByID(_ context.Context, id string) (*player.Player, error) {
	playerID, err := uuid.Parse(id)
	if err != nil {
		return nil, player.ErrNotFound
	}
	if p := r.players[playerID]; p != nil {
		return p, nil
	}
	return nil, player.ErrNotFound
}
func (r leaderboardPlayerRepo) GetByUserID(context.Context, string) (*player.Player, error) {
	return nil, player.ErrNotFound
}
func (r leaderboardPlayerRepo) GetAll(context.Context) ([]*player.Player, error) { return nil, nil }
func (r leaderboardPlayerRepo) Update(context.Context, *player.Player) error     { return nil }
func (r leaderboardPlayerRepo) Delete(context.Context, string) error             { return nil }

func TestGetTournamentLeaderboardAggregatesVerifiedMatches(t *testing.T) {
	tournamentID := uuid.New()
	captainA, captainB := uuid.New(), uuid.New()
	teamA := &team.Team{ID: uuid.New(), Name: "Alpha", CaptainID: captainA}
	teamB := &team.Team{ID: uuid.New(), Name: "Bravo", CaptainID: captainB}
	tournamentRepo := leaderboardTournamentRepo{tournament: &tournament.Tournament{
		ID:            tournamentID,
		ScoringSchema: tournament.ScoringSchema{Weights: map[string]float64{"placement": 2, "kills": 3}},
	}}
	matches := []match.Match{
		{TeamID: teamA.ID, Status: match.StatusVerified, TeamPlacement: 1, TeamKills: 5},
		{TeamID: teamA.ID, Status: match.StatusVerified, TeamPlacement: 3, TeamKills: 2},
		{TeamID: teamB.ID, Status: match.StatusVerified, TeamPlacement: 2, TeamKills: 10},
		{TeamID: teamB.ID, Status: match.StatusDraft, TeamPlacement: 1, TeamKills: 99},
	}
	service := NewService(nil, nil, leaderboardMatchRepo{matches: matches}, leaderboardTeamRepo{teams: []*team.Team{teamA, teamB}}, leaderboardPlayerRepo{players: map[uuid.UUID]*player.Player{
		captainA: {ID: captainA, DisplayName: "Captain A"},
		captainB: {ID: captainB, DisplayName: "Captain B"},
	}}, tournamentRepo)

	entries, total, err := service.GetTournamentLeaderboard(context.Background(), tournamentID, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(entries) != 2 {
		t.Fatalf("got %d entries of %d", len(entries), total)
	}
	if entries[0].TeamID != teamA.ID || entries[0].TotalScore != 417 || entries[0].MatchesPlayed != 2 || entries[0].AveragePlacement != 2 {
		t.Fatalf("unexpected winner: %+v", entries[0])
	}
	if entries[1].TeamID != teamB.ID || entries[1].TotalScore != 228 || entries[1].MatchesPlayed != 1 {
		t.Fatalf("unexpected runner-up: %+v", entries[1])
	}
	if entries[0].Rank != 1 || entries[1].Rank != 2 {
		t.Fatalf("unexpected ranks: %+v", entries)
	}
}
