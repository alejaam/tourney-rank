// API Types for TourneyRank

// Auth
export interface User {
  id: string;
  username: string;
  email: string;
  role: "user" | "admin";
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

// Games
export interface Game {
  id: string;
  name: string;
  slug: string;
  description: string;
  stat_schema: Record<string, StatField>;
  ranking_weights: Record<string, number>;
  platform_id_format: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface StatField {
  type: string;
  min?: number;
  max?: number;
  label: string;
}

// Players & Leaderboard
export interface Player {
  id: string;
  user_id: string;
  display_name: string;
  avatar_url?: string;
  bio?: string;
  platform_ids?: Record<string, string>;
  birth_year?: number;
  region?: string;
  preferred_platform?: string;
  language?: string;
  is_banned: boolean;
  banned_at?: string;
  created_at: string;
  updated_at: string;
}

export interface LeaderboardEntry {
  rank: number;
  player_id: string;
  username: string;
  score: number;
  tier: string;
  stats: Record<string, number>;
}

// Player Stats
export interface PlayerStats {
  id: string;
  player_id: string;
  game_id: string;
  game_name?: string;
  ranking_score: number;
  tier: "elite" | "advanced" | "intermediate" | "beginner";
  matches_played: number;
  stats: Record<string, number | string>;
  last_match_at?: string;
  created_at: string;
  updated_at: string;
}

export interface PlayerGamesSummary {
  player: Player;
  games: Array<{
    game_id: string;
    game_name: string;
    stats: PlayerStats;
  }>;
}

export interface PlayerGameStatsDetail extends PlayerStats {
  rank: number;
  percentile: number;
}

// Admin API Types
export interface ListUsersResponse {
  users: User[];
  total: number;
}

export interface ListGamesResponse {
  games: Game[];
  total: number;
}

export interface ListPlayersResponse {
  players: Player[];
  total: number;
}

export interface UpdateRoleRequest {
  role: "user" | "admin";
}

export interface CreateGameRequest {
  name: string;
  slug: string;
  description: string;
  platform_id_format: string;
  stat_schema: Record<string, StatField>;
  ranking_weights: Record<string, number>;
}

export interface UpdateGameRequest {
  name: string;
  description: string;
  platform_id_format: string;
  stat_schema: Record<string, StatField>;
  ranking_weights: Record<string, number>;
  is_active: boolean;
}

export interface CreatePlayerRequest {
  user_id: string;
  display_name: string;
  avatar_url?: string;
  bio?: string;
  platform_ids?: Record<string, string>;
}

export interface UpdatePlayerRequest {
  display_name: string;
  avatar_url?: string;
  bio?: string;
  platform_ids?: Record<string, string>;
}

// Player Profile (for authenticated users creating/updating their own profile)
export interface CreateProfileRequest {
  display_name: string;
  preferred_platform: string;
  avatar_url?: string;
  bio?: string;
  platform_ids?: Record<string, string>;
  birth_year?: number;
  region?: string;
  language?: string;
}

export interface UpdateProfileRequest {
  display_name?: string;
  avatar_url?: string;
  bio?: string;
  platform_ids?: Record<string, string>;
  birth_year?: number;
  region?: string;
  preferred_platform?: string;
  language?: string;
}

// Tournaments
export interface Tournament {
  id: string;
  game_id: string;
  name: string;
  team_size: "solo" | "duos" | "trios" | "quads";
  status: "draft" | "open" | "active" | "finished" | "canceled";
  start_date: string;
  end_date: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTournamentRequest {
  game_id: string;
  name: string;
  team_size: "solo" | "duos" | "trios" | "quads";
  start_date: string;
  end_date: string;
}

// Teams
export interface Team {
  id: string;
  tournament_id: string;
  name: string;
  tag: string;
  captain_id: string;
  member_ids: string[];
  invite_code: string;
  logo_url?: string;
  created_at: string;
  updated_at: string;
}

export interface TeamWithMembers {
  id: string;
  tournament_id: string;
  name: string;
  tag: string;
  captain_id: string;
  members: Array<{
    id: string;
    display_name: string;
    avatar_url?: string;
  }>;
  invite_code: string;
  created_at: string;
  updated_at: string;
}

// Matches & Match Reports
export interface PlayerMatchStats {
  player_id: string;
  kills: number;
  damage: number;
  assists: number;
  deaths: number;
  downs: number;
  custom_stats?: Record<string, unknown>;
}

export interface Match {
  id: string;
  tournament_id: string;
  team_id: string;
  game_id: string;
  status: "draft" | "verified" | "rejected";
  team_placement: number;
  team_kills: number;
  player_stats: PlayerMatchStats[];
  screenshot_url: string;
  rejection_reason?: string;
  submitted_by: string;
  created_at: string;
  updated_at: string;
  verified_at?: string;
}

export interface SubmitMatchRequest {
  tournament_id: string;
  team_id: string;
  team_placement: number;
  team_kills: number;
  player_stats: PlayerMatchStats[];
  screenshot_url: string;
}

// API Error
export interface ApiError {
  error: string;
}
