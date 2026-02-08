import api from "../lib/axios";
import type { CreateTournamentRequest, Tournament } from "../types/api";

export const tournamentApi = {
  /**
   * Create a new tournament
   */
  createTournament: async (
    data: CreateTournamentRequest,
  ): Promise<Tournament> => {
    const response = await api.post<Tournament>("/tournaments", data);
    return response.data;
  },

  /**
   * Get a tournament by ID
   */
  getTournament: async (id: string): Promise<Tournament> => {
    const response = await api.get<Tournament>(`/tournaments/${id}`);
    return response.data;
  },

  /**
   * List all tournaments with optional filters
   */
  listTournaments: async (params?: {
    game_id?: string;
    status?: string;
    created_by?: string;
    limit?: number;
    offset?: number;
  }): Promise<{
    tournaments: Tournament[];
    total: number;
    limit: number;
    offset: number;
  }> => {
    const response = await api.get("/tournaments", { params });
    return response.data;
  },

  /**
   * Get active tournaments
   */
  getActiveTournaments: async (): Promise<Tournament[]> => {
    const response = await api.get<Tournament[]>("/tournaments/active");
    return response.data;
  },

  /**
   * Update a tournament
   */
  updateTournament: async (
    id: string,
    data: {
      name?: string;
      description?: string;
      start_date?: string;
      end_date?: string;
      prize_pool?: string;
      banner_url?: string;
    },
  ): Promise<Tournament> => {
    const response = await api.patch<Tournament>(`/tournaments/${id}`, data);
    return response.data;
  },

  /**
   * Update tournament status
   */
  updateTournamentStatus: async (
    id: string,
    status: "draft" | "open" | "active" | "finished" | "canceled",
  ): Promise<Tournament> => {
    const response = await api.patch<Tournament>(`/tournaments/${id}/status`, {
      status,
    });
    return response.data;
  },

  /**
   * Delete a tournament
   */
  deleteTournament: async (id: string): Promise<void> => {
    await api.delete(`/tournaments/${id}`);
  },

  /**
   * Get tournament statistics
   */
  getTournamentStats: async (
    id: string,
  ): Promise<{
    tournament_id: string;
    total_teams: number;
    active_teams: number;
    total_matches: number;
    total_players: number;
  }> => {
    const response = await api.get(`/tournaments/${id}/stats`);
    return response.data;
  },
};
