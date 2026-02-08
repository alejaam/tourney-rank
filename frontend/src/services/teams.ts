import api from "../lib/axios";
import type { Team, TeamWithMembers } from "../types/api";

export interface CreateTeamRequest {
  tournament_id: string;
  name: string;
  tag?: string;
  logo_url?: string;
}

export interface JoinTeamRequest {
  invite_code: string;
}

export const teamApi = {
  /**
   * Create a new team for a tournament
   */
  createTeam: async (data: CreateTeamRequest): Promise<Team> => {
    const response = await api.post<Team>("/teams", data);
    return response.data;
  },

  /**
   * Join a team via invite code
   */
  joinTeam: async (data: JoinTeamRequest): Promise<Team> => {
    const response = await api.post<Team>("/teams/join", data);
    return response.data;
  },

  /**
   * Get team by ID with member details
   */
  getTeamWithMembers: async (teamId: string): Promise<TeamWithMembers> => {
    const response = await api.get<TeamWithMembers>(`/teams/${teamId}`);
    return response.data;
  },

  /**
   * Get all teams for a tournament
   */
  getTournamentTeams: async (tournamentId: string): Promise<Team[]> => {
    const response = await api.get<{ teams: Team[] }>(
      `/tournaments/${tournamentId}/teams`,
    );
    return response.data.teams;
  },

  /**
   * Get my team in a specific tournament
   */
  getMyTeamInTournament: async (
    tournamentId: string,
  ): Promise<TeamWithMembers | null> => {
    try {
      const response = await api.get<TeamWithMembers>(
        `/tournaments/${tournamentId}/my-team`,
      );
      return response.data;
    } catch (error) {
      return null;
    }
  },
};
