DROP TRIGGER IF EXISTS update_player_stats_updated_at ON player_stats;
DROP TRIGGER IF EXISTS update_players_updated_at ON players;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_player_stats_tier;
DROP INDEX IF EXISTS idx_player_stats_game_ranking;
DROP INDEX IF EXISTS idx_players_user_id;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS player_stats;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS users;
