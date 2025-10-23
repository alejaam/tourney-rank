-- Create games table: represents supported competitive games
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    stat_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    ranking_weights JSONB NOT NULL DEFAULT '{}'::jsonb,
    platform_id_format VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for active games lookup
CREATE INDEX idx_games_slug ON games(slug) WHERE is_active = true;

-- Insert default game: Call of Duty Warzone
INSERT INTO games (name, slug, description, stat_schema, ranking_weights, platform_id_format) VALUES
(
    'Call of Duty: Warzone',
    'warzone',
    'Battle Royale mode for Call of Duty',
    '{
        "kills": {"type": "integer", "min": 0, "label": "Kills"},
        "deaths": {"type": "integer", "min": 0, "label": "Deaths"},
        "damage": {"type": "integer", "min": 0, "label": "Damage"},
        "contracts": {"type": "integer", "min": 0, "label": "Contracts"},
        "cash": {"type": "integer", "min": 0, "label": "Cash"}
    }'::jsonb,
    '{
        "kd_ratio": 0.40,
        "avg_kills": 0.30,
        "avg_damage": 0.20,
        "consistency": 0.10
    }'::jsonb,
    'activision_id'
);

-- Create comment on stat_schema column
COMMENT ON COLUMN games.stat_schema IS 'Flexible JSON schema defining available statistics for this game';
COMMENT ON COLUMN games.ranking_weights IS 'Weights for ranking calculation, must sum to 1.0';
