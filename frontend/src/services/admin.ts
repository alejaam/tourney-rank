import api from "../lib/axios";
import type {
  CreateGameRequest,
  CreatePlayerRequest,
  Game,
  ListGamesResponse,
  ListPlayersResponse,
  ListUsersResponse,
  Player,
  UpdateGameRequest,
  UpdatePlayerRequest,
  UpdateRoleRequest,
  User,
} from "../types/api";

export const adminApi = {
  // User Management
  users: {
    list: async (): Promise<ListUsersResponse> => {
      const response = await api.get<ListUsersResponse>("/admin/users");
      return response.data;
    },

    get: async (id: string): Promise<User> => {
      const response = await api.get<User>(`/admin/users/${id}`);
      return response.data;
    },

    delete: async (id: string): Promise<void> => {
      await api.delete(`/admin/users/${id}`);
    },

    updateRole: async (id: string, data: UpdateRoleRequest): Promise<void> => {
      await api.patch(`/admin/users/${id}/role`, data);
    },
  },

  // Game Management
  games: {
    list: async (): Promise<ListGamesResponse> => {
      const response = await api.get<ListGamesResponse>("/admin/games");
      return response.data;
    },

    get: async (id: string): Promise<Game> => {
      const response = await api.get<Game>(`/admin/games/${id}`);
      return response.data;
    },

    create: async (data: CreateGameRequest): Promise<Game> => {
      const response = await api.post<Game>("/admin/games", data);
      return response.data;
    },

    update: async (id: string, data: UpdateGameRequest): Promise<Game> => {
      const response = await api.put<Game>(`/admin/games/${id}`, data);
      return response.data;
    },

    delete: async (id: string): Promise<void> => {
      await api.delete(`/admin/games/${id}`);
    },
  },

  // Player Management
  players: {
    list: async (): Promise<ListPlayersResponse> => {
      const response = await api.get<ListPlayersResponse>("/admin/players");
      return response.data;
    },

    get: async (id: string): Promise<Player> => {
      const response = await api.get<Player>(`/admin/players/${id}`);
      return response.data;
    },

    create: async (data: CreatePlayerRequest): Promise<Player> => {
      const response = await api.post<Player>("/admin/players", data);
      return response.data;
    },

    update: async (id: string, data: UpdatePlayerRequest): Promise<Player> => {
      const response = await api.put<Player>(`/admin/players/${id}`, data);
      return response.data;
    },

    delete: async (id: string): Promise<void> => {
      await api.delete(`/admin/players/${id}`);
    },

    ban: async (id: string): Promise<Player> => {
      const response = await api.patch<Player>(`/admin/players/${id}/ban`);
      return response.data;
    },

    unban: async (id: string): Promise<Player> => {
      const response = await api.patch<Player>(`/admin/players/${id}/unban`);
      return response.data;
    },
  },
};
