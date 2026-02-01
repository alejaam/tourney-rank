// API Types for TourneyRank

// Auth
export interface User {
  id: string;
  username: string;
  email: string;
  role: 'user' | 'admin';
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
  avatar_url: string;
  bio: string;
  platform_ids: Record<string, string>;
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
  role: 'user' | 'admin';
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

// API Error
export interface ApiError {
  error: string;
}
