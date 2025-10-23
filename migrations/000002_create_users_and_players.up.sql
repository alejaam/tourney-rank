-- Create users table: represents all system users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'team_leader', 'player')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create players table: extended profile for player role
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    platform_ids JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Create player_stats table: stores per-game statistics and rankings
CREATE TABLE IF NOT EXISTS player_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    matches_played INTEGER DEFAULT 0,
    ranking_score DECIMAL(10, 2) DEFAULT 0.00,
    tier VARCHAR(20) CHECK (tier IN ('elite', 'advanced', 'intermediate', 'beginner')),
    last_match_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(player_id, game_id)
);

-- Create indexes for common queries
CREATE INDEX idx_users_email ON users(email) WHERE is_active = true;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_players_user_id ON players(user_id);
CREATE INDEX idx_player_stats_game_ranking ON player_stats(game_id, ranking_score DESC);
CREATE INDEX idx_player_stats_tier ON player_stats(tier);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_player_stats_updated_at BEFORE UPDATE ON player_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user (password: admin123 - CHANGE IN PRODUCTION)
-- Password hash generated with bcrypt cost 10
INSERT INTO users (username, email, password_hash, role) VALUES
('admin', 'admin@tourneyrank.com', '$2a$10$YourHashedPasswordHere', 'admin');

COMMENT ON COLUMN players.platform_ids IS 'JSON object with platform-specific IDs: {"activision_id": "...", "epic_id": "...", "steam_id": "..."}';
COMMENT ON COLUMN player_stats.stats IS 'Aggregated statistics for this player in this game';
COMMENT ON COLUMN player_stats.ranking_score IS 'Calculated ranking score based on game-specific weights';
