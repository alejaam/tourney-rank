import { useEffect, useState } from "react";
import { matchApi } from "../../services/matches";
import { playerApi } from "../../services/player";
import type {
    Player,
    PlayerMatchStats,
    SubmitMatchRequest,
    TeamWithMembers,
    Tournament,
} from "../../types/api";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";
import { N8nIntegrationGuide } from "./N8nIntegrationGuide";

interface MatchReportFormProps {
    onSuccess?: () => void;
}

export function MatchReportForm({ onSuccess }: MatchReportFormProps) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);

    const [tournament, setTournament] = useState<Tournament | null>(null);
    const [team, setTeam] = useState<TeamWithMembers | null>(null);
    const [myProfile, setMyProfile] = useState<Player | null>(null);

    const [teamPlacement, setTeamPlacement] = useState("");
    const [teamKills, setTeamKills] = useState("");
    const [screenshotUrl, setScreenshotUrl] = useState("");

    // Player stats by player ID
    const [playerStats, setPlayerStats] = useState<
        Record<string, Omit<PlayerMatchStats, "player_id">>
    >({});

    // Load initial data
    useEffect(() => {
        const loadData = async () => {
            setLoading(true);
            setError(null);
            try {
                // Get player profile
                const profile = await playerApi.getMyProfile();
                setMyProfile(profile);

                // Get active tournament
                const activeTourney = await matchApi.getActiveTournament();
                if (!activeTourney) {
                    setError("No active tournament found");
                    setLoading(false);
                    return;
                }
                setTournament(activeTourney);

                // Get player's team in this tournament
                const playerTeam = await matchApi.getPlayerTeamInTournament(
                    activeTourney.id
                );
                if (!playerTeam) {
                    setError("You are not in a team for this tournament");
                    setLoading(false);
                    return;
                }
                setTeam(playerTeam);

                // Initialize player stats
                const initialStats: Record<
                    string,
                    Omit<PlayerMatchStats, "player_id">
                > = {};
                playerTeam.members.forEach((member) => {
                    initialStats[member.id] = {
                        kills: 0,
                        damage: 0,
                        assists: 0,
                        deaths: 0,
                        downs: 0,
                    };
                });
                setPlayerStats(initialStats);
            } catch (err) {
                setError(
                    err instanceof Error ? err.message : "Failed to load tournament data"
                );
            } finally {
                setLoading(false);
            }
        };

        loadData();
    }, []);

    const handlePlayerStatChange = (
        playerId: string,
        field: keyof Omit<PlayerMatchStats, "player_id">,
        value: number
    ) => {
        setPlayerStats((prev) => ({
            ...prev,
            [playerId]: {
                ...prev[playerId],
                [field]: value,
            },
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setSuccess(false);

        if (!tournament || !team) {
            setError("Tournament or team data missing");
            return;
        }

        if (!teamPlacement || !teamKills || !screenshotUrl) {
            setError("Please fill in all required fields");
            return;
        }

        const placement = parseInt(teamPlacement, 10);
        if (placement < 1 || placement > 100) {
            setError("Team placement must be between 1 and 100");
            return;
        }

        // Check if user is captain
        if (myProfile?.id !== team.captain_id) {
            setError("Only the team captain can submit match reports");
            return;
        }

        try {
            setLoading(true);

            // Convert player stats to array format
            const playerStatsArray: PlayerMatchStats[] = Object.entries(
                playerStats
            ).map(([playerId, stats]) => ({
                player_id: playerId,
                ...stats,
            }));

            const submitData: SubmitMatchRequest = {
                tournament_id: tournament.id,
                team_id: team.id,
                team_placement: placement,
                team_kills: parseInt(teamKills, 10),
                player_stats: playerStatsArray,
                screenshot_url: screenshotUrl,
            };

            await matchApi.submitMatch(submitData);
            setSuccess(true);

            // Reset form
            setTeamPlacement("");
            setTeamKills("");
            setScreenshotUrl("");
            const initialStats: Record<
                string,
                Omit<PlayerMatchStats, "player_id">
            > = {};
            team.members.forEach((member) => {
                initialStats[member.id] = {
                    kills: 0,
                    damage: 0,
                    assists: 0,
                    deaths: 0,
                    downs: 0,
                };
            });
            setPlayerStats(initialStats);

            if (onSuccess) {
                setTimeout(() => onSuccess(), 1500);
            }
        } catch (err) {
            setError(
                err instanceof Error ? err.message : "Failed to submit match report"
            );
        } finally {
            setLoading(false);
        }
    };

    if (loading && !tournament) {
        return (
            <div className="flex items-center justify-center py-8">
                <p className="text-gray-500">Loading tournament data...</p>
            </div>
        );
    }

    if (error && !tournament) {
        return (
            <div className="bg-red-50 border border-red-200 rounded p-4">
                <p className="text-red-700">{error}</p>
            </div>
        );
    }

    if (!tournament || !team) {
        return (
            <div className="bg-yellow-50 border border-yellow-200 rounded p-4">
                <p className="text-yellow-700">
                    You need to be in a tournament team to report matches
                </p>
            </div>
        );
    }

    return (
        <div className="max-w-2xl mx-auto">
            <form onSubmit={handleSubmit} className="space-y-6">
                {/* N8N Integration Guide */}
                <N8nIntegrationGuide />

                {/* Tournament Info */}
                <div className="bg-blue-50 border border-blue-200 rounded p-4">
                    <h3 className="font-semibold text-blue-900 mb-2">
                        {tournament.name}
                    </h3>
                    <p className="text-sm text-blue-700">Team: {team.name}</p>
                    <p className="text-sm text-blue-700">
                        Team Size: {tournament.team_size}
                    </p>
                </div>

                {/* Success Message */}
                {success && (
                    <div className="bg-green-50 border border-green-200 rounded p-4">
                        <p className="text-green-700">✅ Match report submitted successfully!</p>
                    </div>
                )}

                {/* Error Message */}
                {error && (
                    <div className="bg-red-50 border border-red-200 rounded p-4">
                        <p className="text-red-700">{error}</p>
                    </div>
                )}

                {/* Team Placement */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Team Placement (1-100)
                    </label>
                    <Input
                        type="number"
                        min="1"
                        max="100"
                        value={teamPlacement}
                        onChange={(e) => setTeamPlacement(e.target.value)}
                        placeholder="e.g., 5"
                        required
                    />
                </div>

                {/* Team Kills */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Team Kills
                    </label>
                    <Input
                        type="number"
                        min="0"
                        value={teamKills}
                        onChange={(e) => setTeamKills(e.target.value)}
                        placeholder="e.g., 12"
                        required
                    />
                </div>

                {/* Screenshot URL */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Screenshot URL (from n8n)
                    </label>
                    <Input
                        type="url"
                        value={screenshotUrl}
                        onChange={(e) => setScreenshotUrl(e.target.value)}
                        placeholder="https://..."
                        required
                    />
                    <p className="text-xs text-gray-500 mt-1">
                        Paste the screenshot URL from n8n workflow
                    </p>
                </div>

                {/* Player Stats */}
                <div className="border rounded p-4">
                    <h4 className="font-semibold text-gray-900 mb-4">Player Stats</h4>
                    <div className="space-y-4">
                        {team.members.map((member) => (
                            <div
                                key={member.id}
                                className="border border-gray-200 rounded p-4 bg-gray-50"
                            >
                                <p className="font-medium text-gray-900 mb-3">
                                    {member.display_name}
                                </p>
                                <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
                                    <div>
                                        <label className="block text-xs font-medium text-gray-700 mb-1">
                                            Kills
                                        </label>
                                        <Input
                                            type="number"
                                            min="0"
                                            value={playerStats[member.id]?.kills || ""}
                                            onChange={(e) =>
                                                handlePlayerStatChange(
                                                    member.id,
                                                    "kills",
                                                    parseInt(e.target.value) || 0
                                                )
                                            }
                                            placeholder="0"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-xs font-medium text-gray-700 mb-1">
                                            Damage
                                        </label>
                                        <Input
                                            type="number"
                                            min="0"
                                            value={playerStats[member.id]?.damage || ""}
                                            onChange={(e) =>
                                                handlePlayerStatChange(
                                                    member.id,
                                                    "damage",
                                                    parseInt(e.target.value) || 0
                                                )
                                            }
                                            placeholder="0"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-xs font-medium text-gray-700 mb-1">
                                            Assists
                                        </label>
                                        <Input
                                            type="number"
                                            min="0"
                                            value={playerStats[member.id]?.assists || ""}
                                            onChange={(e) =>
                                                handlePlayerStatChange(
                                                    member.id,
                                                    "assists",
                                                    parseInt(e.target.value) || 0
                                                )
                                            }
                                            placeholder="0"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-xs font-medium text-gray-700 mb-1">
                                            Deaths
                                        </label>
                                        <Input
                                            type="number"
                                            min="0"
                                            value={playerStats[member.id]?.deaths || ""}
                                            onChange={(e) =>
                                                handlePlayerStatChange(
                                                    member.id,
                                                    "deaths",
                                                    parseInt(e.target.value) || 0
                                                )
                                            }
                                            placeholder="0"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-xs font-medium text-gray-700 mb-1">
                                            Downs
                                        </label>
                                        <Input
                                            type="number"
                                            min="0"
                                            value={playerStats[member.id]?.downs || ""}
                                            onChange={(e) =>
                                                handlePlayerStatChange(
                                                    member.id,
                                                    "downs",
                                                    parseInt(e.target.value) || 0
                                                )
                                            }
                                            placeholder="0"
                                        />
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Submit Button */}
                <Button
                    type="submit"
                    disabled={loading}
                    className="w-full bg-blue-600 hover:bg-blue-700 text-white"
                >
                    {loading ? "Submitting..." : "Submit Match Report"}
                </Button>
            </form>
        </div>
    );
}
