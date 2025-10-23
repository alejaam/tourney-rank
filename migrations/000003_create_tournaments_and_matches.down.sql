DROP TRIGGER IF EXISTS update_matches_updated_at ON matches;
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;
DROP TRIGGER IF EXISTS update_tournaments_updated_at ON tournaments;

DROP INDEX IF EXISTS idx_registrations_tournament_status;
DROP INDEX IF EXISTS idx_match_stats_player;
DROP INDEX IF EXISTS idx_match_stats_match;
DROP INDEX IF EXISTS idx_matches_tournament;
DROP INDEX IF EXISTS idx_team_members_player;
DROP INDEX IF EXISTS idx_teams_tournament;
DROP INDEX IF EXISTS idx_tournaments_dates;
DROP INDEX IF EXISTS idx_tournaments_game_status;

DROP TABLE IF EXISTS registrations;
DROP TABLE IF EXISTS match_stats;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS tournaments;
