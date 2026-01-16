import api from '../lib/axios';
import type { Game } from '../types/api';

export const gamesApi = {
  list: async (): Promise<{ games: Game[]; total: number }> => {
    const response = await api.get('/games');
    return response.data;
  },

  getById: async (id: string): Promise<Game> => {
    const response = await api.get<Game>(`/games/${id}`);
    return response.data;
  },
};
