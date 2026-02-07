import type { InternalAxiosRequestConfig } from "axios";
import axios, { AxiosError } from "axios";
import { useAuthStore } from "../store/authStore";
import type { ApiError } from "../types/api";
import { showError } from "./toast";

const api = axios.create({
  baseURL: "/api/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token;
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    if (error.response?.status === 401) {
      // Logout and dispatch custom event for navigation
      try {
        await axios.post("/api/v1/auth/logout");
      } catch {
        // Best-effort cleanup; ignore logout failures.
      }
      useAuthStore.getState().logout();
      showError("Sesión expirada. Por favor inicia sesión nuevamente.");
      window.dispatchEvent(new CustomEvent("auth:logout"));
    }
    return Promise.reject(error);
  },
);

export default api;
