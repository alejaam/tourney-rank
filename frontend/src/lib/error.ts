import { isAxiosError } from "axios";

export function errorMessage(error: unknown, fallback: string): string {
  if (isAxiosError<{ error?: string }>(error)) {
    return error.response?.data?.error || error.message || fallback;
  }
  return error instanceof Error ? error.message : fallback;
}
