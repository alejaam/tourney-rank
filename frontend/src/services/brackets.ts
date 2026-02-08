import api from "../lib/axios";
import type {
  Bracket,
  BracketWithMatchups,
  GenerateBracketRequest,
  Matchup,
  SetMatchupWinnerRequest,
} from "../types/api";

export const bracketApi = {
  // Generate a new bracket for a tournament
  generateBracket: async (
    data: GenerateBracketRequest,
  ): Promise<BracketWithMatchups> => {
    const response = await api.post("/brackets/generate", data);
    return response.data;
  },

  // Get bracket by ID
  getBracket: async (bracketId: string): Promise<Bracket> => {
    const response = await api.get(`/brackets/${bracketId}`);
    return response.data;
  },

  // Get bracket for a specific tournament
  getTournamentBracket: async (
    tournamentId: string,
  ): Promise<BracketWithMatchups> => {
    const response = await api.get(`/tournaments/${tournamentId}/bracket`);
    return response.data;
  },

  // Set winner for a matchup
  setMatchupWinner: async (
    matchupId: string,
    data: SetMatchupWinnerRequest,
  ): Promise<Matchup> => {
    const response = await api.post(`/matchups/${matchupId}/winner`, data);
    return response.data;
  },

  // Delete a bracket
  deleteBracket: async (bracketId: string): Promise<void> => {
    await api.delete(`/brackets/${bracketId}`);
  },
};
