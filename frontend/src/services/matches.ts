import api from "../lib/axios";
import type {
  Match,
  SubmitMatchRequest,
  TeamWithMembers,
  Tournament,
} from "../types/api";

export const matchApi = {
  /**
   * Get the current player's active tournament
   */
  getActiveTournament: async (): Promise<Tournament | null> => {
    try {
      const response = await api.get<Tournament>(
        "/players/me/active-tournament",
      );
      return response.data;
    } catch (error) {
      // No active tournament
      return null;
    }
  },

  /**
   * Get the current player's team in a tournament
   */
  getPlayerTeamInTournament: async (
    tournamentId: string,
  ): Promise<TeamWithMembers | null> => {
    try {
      const response = await api.get<TeamWithMembers>(
        `/tournaments/${tournamentId}/my-team`,
      );
      return response.data;
    } catch (error) {
      // Player not in tournament
      return null;
    }
  },

  /**
   * Submit a match report (requires user to be captain)
   */
  submitMatch: async (data: SubmitMatchRequest): Promise<Match> => {
    const response = await api.post<Match>("/matches/report", data);
    return response.data;
  },

  /**
   * Get current player's match history
   */
  getMyMatches: async (
    limit = 10,
    offset = 0,
  ): Promise<{
    matches: Match[];
    total: number;
    limit: number;
    offset: number;
  }> => {
    const response = await api.get("/players/me/matches", {
      params: { limit, offset },
    });
    return response.data;
  },

  /**
   * Get all matches in a tournament (public)
   */
  getTournamentMatches: async (
    tournamentId: string,
    limit = 10,
    offset = 0,
  ): Promise<{
    matches: Match[];
    total: number;
    limit: number;
    offset: number;
  }> => {
    const response = await api.get(`/matches/tournament/${tournamentId}`, {
      params: { limit, offset },
    });
    return response.data;
  },

  /**
   * Get a single match (public)
   */
  getMatch: async (matchId: string): Promise<Match> => {
    const response = await api.get<Match>(`/matches/${matchId}`);
    return response.data;
  },

  /**
   * Admin: Get all unverified matches
   */
  getUnverifiedMatches: async (
    limit = 10,
    offset = 0,
  ): Promise<{
    matches: Match[];
    total: number;
    limit: number;
    offset: number;
  }> => {
    const response = await api.get("/admin/matches/unverified", {
      params: { limit, offset },
    });
    return response.data;
  },

  /**
   * Admin: Verify or reject a match
   */
  verifyMatch: async (
    matchId: string,
    approved: boolean,
    reason?: string,
  ): Promise<Match> => {
    const response = await api.patch<Match>(
      `/admin/matches/${matchId}/verify`,
      {
        approved,
        reason: reason || "",
      },
    );
    return response.data;
  },
};
