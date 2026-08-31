import { useEffect, useState } from "react";
import { gamesApi } from "../../services/games";
import { errorMessage } from "../../lib/error";
import { tournamentApi } from "../../services/tournaments";
import type { CreateTournamentRequest, Tournament } from "../../types/api";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";

interface TournamentFormProps {
    onSuccess?: (tournament: Tournament) => void;
    initialData?: Tournament;
}

export function TournamentForm({ onSuccess, initialData }: TournamentFormProps) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [games, setGames] = useState<Array<{ id: string; name: string }>>([]);

    // Convert ISO date to datetime-local input format (YYYY-MM-DDTHH:mm)
    const formatDateForInput = (isoDate: string | undefined) => {
        if (!isoDate) return "";
        try {
            const date = new Date(isoDate);
            const year = date.getFullYear();
            const month = String(date.getMonth() + 1).padStart(2, "0");
            const day = String(date.getDate()).padStart(2, "0");
            const hours = String(date.getHours()).padStart(2, "0");
            const minutes = String(date.getMinutes()).padStart(2, "0");
            return `${year}-${month}-${day}T${hours}:${minutes}`;
        } catch {
            return "";
        }
    };

    const [formData, setFormData] = useState({
        game_id: initialData?.game_id || "",
        name: initialData?.name || "",
        description: initialData?.description || "",
        team_size: initialData?.team_size || 2,
        start_date: formatDateForInput(initialData?.start_date),
        end_date: formatDateForInput(initialData?.end_date),
        prize_pool: initialData?.prize_pool || "",
        banner_url: initialData?.banner_url || "",
    });

    useEffect(() => {
        const loadGames = async () => {
            try {
                const response = await gamesApi.list();
                setGames(response.games || []);
            } catch (err) {
                console.error("Failed to load games", err);
            }
        };
        loadGames();
    }, []);

    const handleChange = (
        e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>
    ) => {
        const { name, value } = e.target;
        setFormData((prev) => ({
            ...prev,
            [name]: name === "team_size" ? parseInt(value, 10) : value,
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setLoading(true);

        try {
            let tournament: Tournament;
            if (initialData) {
                const updateData = {
                    name: formData.name,
                    description: formData.description,
                    start_date: formData.start_date ? new Date(formData.start_date).toISOString() : "",
                    end_date: formData.end_date ? new Date(formData.end_date).toISOString() : "",
                    prize_pool: formData.prize_pool,
                    banner_url: formData.banner_url,
                };
                console.log("Updating tournament:", updateData);
                tournament = await tournamentApi.updateTournament(initialData.id, updateData);
            } else {
                // Ensure datetime-local values have seconds appended before ISO conversion
                const startDateStr = formData.start_date.includes(':00:')
                    ? formData.start_date
                    : `${formData.start_date}:00`;
                const endDateStr = formData.end_date.includes(':00:')
                    ? formData.end_date
                    : `${formData.end_date}:00`;

                const createReq: CreateTournamentRequest = {
                    game_id: formData.game_id,
                    name: formData.name,
                    description: formData.description,
                    team_size: Number(formData.team_size) as 1 | 2 | 3 | 4, // Ensure it's a number
                    start_date: new Date(startDateStr).toISOString(),
                    end_date: new Date(endDateStr).toISOString(),
                    prize_pool: formData.prize_pool,
                    banner_url: formData.banner_url,
                    rules: {
                        max_teams: 32,
                        min_matches: 1,
                        max_matches: 10,
                        require_verification: true,
                        allow_late_registration: false,
                    },
                };
                tournament = await tournamentApi.createTournament(createReq);
            }
            onSuccess?.(tournament);
        } catch (err: unknown) {
            setError(errorMessage(err, "Failed to save tournament"));
        } finally {
            setLoading(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
                <div className="p-4 bg-red-100 border border-red-400 text-red-700 rounded">
                    {error}
                </div>
            )}

            {!initialData && (
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Game *
                    </label>
                    <select
                        name="game_id"
                        value={formData.game_id}
                        onChange={handleChange}
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="">Select a game</option>
                        {games.map((game) => (
                            <option key={game.id} value={game.id}>
                                {game.name}
                            </option>
                        ))}
                    </select>
                </div>
            )}

            <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                    Tournament Name *
                </label>
                <Input
                    type="text"
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    placeholder="E.g., Spring Championship 2025"
                    required
                />
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                    Description
                </label>
                <textarea
                    name="description"
                    value={formData.description}
                    onChange={handleChange}
                    placeholder="Tournament details..."
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>

            {!initialData && (
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Team Size *
                    </label>
                    <select
                        name="team_size"
                        value={formData.team_size}
                        onChange={handleChange}
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="1">Solo</option>
                        <option value="2">Duos</option>
                        <option value="3">Trios</option>
                        <option value="4">Quads</option>
                    </select>
                </div>
            )}

            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Start Date *
                    </label>
                    <Input
                        type="datetime-local"
                        name="start_date"
                        value={formData.start_date}
                        onChange={handleChange}
                        required
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        End Date *
                    </label>
                    <Input
                        type="datetime-local"
                        name="end_date"
                        value={formData.end_date}
                        onChange={handleChange}
                        required
                    />
                </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Prize Pool
                    </label>
                    <Input
                        type="text"
                        name="prize_pool"
                        value={formData.prize_pool}
                        onChange={handleChange}
                        placeholder="E.g., $10,000"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Banner URL
                    </label>
                    <Input
                        type="url"
                        name="banner_url"
                        value={formData.banner_url}
                        onChange={handleChange}
                        placeholder="https://..."
                    />
                </div>
            </div>

            <Button type="submit" disabled={loading}>
                {loading ? "Saving..." : initialData ? "Update Tournament" : "Create Tournament"}
            </Button>
        </form>
    );
}
