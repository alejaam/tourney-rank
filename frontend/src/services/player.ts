import api from "../lib/axios";
import type { Player } from "../types/api";

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
};
