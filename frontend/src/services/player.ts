import api from "../lib/axios";
import type {
  Player,
  PlayerGamesSummary,
  PlayerGameStatsDetail,
} from "../types/api";

export interface UpdatePlayerProfileRequest {
  display_name?: string;
  avatar_url?: string;
  bio?: string;
  platform_ids?: Record<string, string>;
}

export interface CreatePlayerProfileRequest {
  display_name: string;
}

export const playerApi = {
  /**
   * Get my player profile (auto-creates if doesn't exist)
   */
  getMyProfile: async (): Promise<Player> => {
    const response = await api.get<Player>("/players/me");
    return response.data;
  },

  /**
   * Create my player profile
   */
  createMyProfile: async (
    data: CreatePlayerProfileRequest,
  ): Promise<Player> => {
    const response = await api.post<Player>("/players/me", data);
    return response.data;
  },

  /**
   * Update my player profile
   */
  updateMyProfile: async (
    data: UpdatePlayerProfileRequest,
  ): Promise<Player> => {
    const response = await api.put<Player>("/players/me", data);
    return response.data;
  },

  /**
   * Get my player profile with all game stats
   */
  getMyProfileWithStats: async (): Promise<PlayerGamesSummary> => {
    const response = await api.get<PlayerGamesSummary>("/players/me/stats");
    return response.data;
  },

  /**
   * Get my stats for a specific game including rank and percentile
   */
  getMyGameStats: async (gameId: string): Promise<PlayerGameStatsDetail> => {
    const response = await api.get<PlayerGameStatsDetail>(
      `/players/me/stats/${gameId}`,
    );
    return response.data;
  },
};
