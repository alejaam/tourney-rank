import { clsx, type ClassValue } from "clsx";

export const cn = (...inputs: ClassValue[]) => clsx(inputs);

/**
 * Converts a numeric team size to a readable label
 */
export const teamSizeToLabel = (teamSize: number): string => {
  const labels: Record<number, string> = {
    1: "Solo",
    2: "Duos",
    3: "Trios",
    4: "Quads",
  };
  return labels[teamSize] || `${teamSize} Players`;
};
