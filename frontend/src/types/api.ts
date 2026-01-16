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
  username: string;
  platform_ids: Record<string, string>;
  tier: string;
  created_at: string;
}

export interface LeaderboardEntry {
  rank: number;
  player_id: string;
  username: string;
  score: number;
  tier: string;
  stats: Record<string, number>;
}

// API Error
export interface ApiError {
  error: string;
}
