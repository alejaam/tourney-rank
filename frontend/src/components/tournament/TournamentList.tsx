import { useCallback, useEffect, useState } from "react";
import { teamSizeToLabel } from "../../lib/utils";
import { tournamentApi } from "../../services/tournaments";
import type { Tournament } from "../../types/api";
import { Button } from "../ui/Button";

interface TournamentListProps {
    onTournamentSelect?: (tournament: Tournament) => void;
    onTournamentDelete?: (id: string) => void;
    isAdmin?: boolean;
}

export function TournamentList({
    onTournamentSelect,
    onTournamentDelete,
    isAdmin = false,
}: TournamentListProps) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [tournaments, setTournaments] = useState<Tournament[]>([]);
    const [filter, setFilter] = useState<"all" | "draft" | "open" | "active" | "finished">("all");

    const loadTournaments = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const params = filter !== "all" ? { status: filter } : undefined;
            const response = await tournamentApi.listTournaments(params);
            setTournaments(response.tournaments || []);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load tournaments");
        } finally {
            setLoading(false);
        }
    }, [filter]);

    useEffect(() => {
        loadTournaments();
    }, [loadTournaments]);

    const handleDelete = async (id: string) => {
        if (window.confirm("Are you sure you want to delete this tournament?")) {
            try {
                await tournamentApi.deleteTournament(id);
                setTournaments((prev) => prev.filter((t) => t.id !== id));
                onTournamentDelete?.(id);
            } catch (err) {
                setError(err instanceof Error ? err.message : "Failed to delete tournament");
            }
        }
    };

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            year: "numeric",
        });
    };

    const getStatusBadgeClass = (status: string) => {
        switch (status) {
            case "draft":
                return "bg-gray-100 text-gray-800";
            case "open":
                return "bg-green-100 text-green-800";
            case "active":
                return "bg-blue-100 text-blue-800";
            case "finished":
                return "bg-purple-100 text-purple-800";
            case "canceled":
                return "bg-red-100 text-red-800";
            default:
                return "bg-gray-100 text-gray-800";
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex gap-2">
                {(["all", "draft", "open", "active", "finished"] as const).map(
                    (status) => (
                        <button
                            key={status}
                            onClick={() => setFilter(status)}
                            className={`px-4 py-2 rounded-md text-sm font-medium transition ${filter === status
                                ? "bg-blue-500 text-white"
                                : "bg-gray-200 text-gray-800 hover:bg-gray-300"
                                }`}
                        >
                            {status.charAt(0).toUpperCase() + status.slice(1)}
                        </button>
                    )
                )}
            </div>

            {error && (
                <div className="p-4 bg-red-100 border border-red-400 text-red-700 rounded">
                    {error}
                </div>
            )}

            {loading ? (
                <div className="text-center py-8">Loading tournaments...</div>
            ) : tournaments.length === 0 ? (
                <div className="text-center py-8 text-gray-500">No tournaments found</div>
            ) : (
                <div className="space-y-3">
                    {tournaments.map((tournament) => (
                        <div
                            key={tournament.id}
                            className="p-4 border border-gray-200 rounded-lg hover:shadow-md transition cursor-pointer"
                            onClick={() => onTournamentSelect?.(tournament)}
                        >
                            <div className="flex justify-between items-start">
                                <div className="flex-1">
                                    <h3 className="text-lg font-semibold text-gray-900">
                                        {tournament.name}
                                    </h3>
                                    <p className="text-sm text-gray-600 mt-1">
                                        {tournament.description}
                                    </p>
                                    <div className="flex gap-3 mt-3 text-sm text-gray-600">
                                        <span>📅 {formatDate(tournament.start_date)}</span>
                                        <span>👥 {teamSizeToLabel(tournament.team_size)}</span>
                                    </div>
                                </div>
                                <div className="flex flex-col items-end gap-2">
                                    <span
                                        className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusBadgeClass(
                                            tournament.status
                                        )}`}
                                    >
                                        {tournament.status}
                                    </span>
                                    {isAdmin && (
                                        <Button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDelete(tournament.id);
                                            }}
                                            variant="danger"
                                            size="sm"
                                        >
                                            Delete
                                        </Button>
                                    )}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
