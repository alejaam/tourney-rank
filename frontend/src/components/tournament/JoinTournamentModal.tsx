import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "../../lib/toast";
import { teamSizeToLabel } from "../../lib/utils";
import { playerApi } from "../../services/player";
import { teamApi, type CreateTeamRequest } from "../../services/teams";
import { tournamentApi } from "../../services/tournaments";
import { useAuthStore } from "../../store/authStore";
import type { PlayerGameStatsDetail, Tournament } from "../../types/api";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";

interface JoinTournamentModalProps {
    isOpen: boolean;
    onClose: () => void;
}

type JoinMethod = "create" | "join" | null;

export function JoinTournamentModal({ isOpen, onClose }: JoinTournamentModalProps) {
    const navigate = useNavigate();
    const user = useAuthStore((state) => state.user);
    const [tournaments, setTournaments] = useState<Tournament[]>([]);
    const [selectedTournament, setSelectedTournament] = useState<Tournament | null>(null);
    const [warzoneStats, setWarzoneStats] = useState<PlayerGameStatsDetail | null>(null);
    const [loading, setLoading] = useState(true);
    const [submitting, setSubmitting] = useState(false);
    const [joinMethod, setJoinMethod] = useState<JoinMethod>(null);

    // For creating a team
    const [teamName, setTeamName] = useState("");
    const [teamTag, setTeamTag] = useState("");

    // For joining a team
    const [inviteCode, setInviteCode] = useState("");

    useEffect(() => {
        if (isOpen) {
            loadData();
        }
    }, [isOpen]);

    const loadData = async () => {
        setLoading(true);
        try {
            // Load open tournaments for Warzone
            const response = await tournamentApi.listTournaments({
                status: "open",
                limit: 50,
            });

            // Filter Warzone tournaments only (handle null/undefined tournaments)
            const allTournaments = response.tournaments || [];

            console.log("🏆 All open tournaments:", allTournaments.length);
            console.log("📋 Tournaments data:", allTournaments);

            const warzoneTournaments = allTournaments.filter(
                (t) => t.game_name?.toLowerCase().includes("warzone") ||
                    t.name.toLowerCase().includes("warzone")
            );

            console.log("🎮 Warzone tournaments found:", warzoneTournaments.length);
            setTournaments(warzoneTournaments);

            // Load player's Warzone stats
            try {
                // Try to find Warzone game ID - you may need to adjust this
                const firstWarzoneTournament = warzoneTournaments[0];
                if (firstWarzoneTournament) {
                    const stats = await playerApi.getMyGameStats(firstWarzoneTournament.game_id);
                    setWarzoneStats(stats);
                }
            } catch (error) {
                console.warn("No Warzone stats found for player");
            }
        } catch (error: any) {
            console.error("Failed to load tournaments:", error);
            toast.error("Failed to load tournaments");
        } finally {
            setLoading(false);
        }
    };

    const checkEligibility = (tournament: Tournament): { eligible: boolean; reasons: string[] } => {
        const reasons: string[] = [];

        if (!warzoneStats) {
            reasons.push("No Warzone stats found. Play some matches first!");
            return { eligible: false, reasons };
        }

        const rules = tournament.rules;

        // Check max K/D
        if (rules.max_kd !== undefined) {
            const playerKD = Number(warzoneStats.stats.kd || 0);
            if (playerKD > rules.max_kd) {
                reasons.push(`K/D too high (${playerKD.toFixed(2)} > ${rules.max_kd})`);
            }
        }

        // Check min K/D
        if (rules.min_kd !== undefined) {
            const playerKD = Number(warzoneStats.stats.kd || 0);
            if (playerKD < rules.min_kd) {
                reasons.push(`K/D too low (${playerKD.toFixed(2)} < ${rules.min_kd})`);
            }
        }

        // Check min matches played
        if (rules.min_matches_played !== undefined) {
            if (warzoneStats.matches_played < rules.min_matches_played) {
                reasons.push(
                    `Not enough matches played (${warzoneStats.matches_played} < ${rules.min_matches_played})`
                );
            }
        }

        // Check if tournament is full
        if (rules.max_teams > 0 && tournament.current_teams !== undefined) {
            if (tournament.current_teams >= rules.max_teams) {
                reasons.push("Tournament is full");
            }
        }

        return { eligible: reasons.length === 0, reasons };
    };

    const handleSelectTournament = (tournament: Tournament) => {
        setSelectedTournament(tournament);
        setJoinMethod(null);
        setTeamName("");
        setTeamTag("");
        setInviteCode("");
    };

    const handleCreateTeam = async () => {
        if (!selectedTournament || !teamName.trim()) {
            toast.error("Please enter a team name");
            return;
        }

        setSubmitting(true);
        try {
            const createData: CreateTeamRequest = {
                tournament_id: selectedTournament.id,
                name: teamName.trim(),
                tag: teamTag.trim() || undefined,
            };

            const team = await teamApi.createTeam(createData);
            toast.success(`Team "${team.name}" created successfully!`);

            // Navigate to tournament page or team management
            onClose();
            navigate(`/tournaments/${selectedTournament.id}`);
        } catch (error: any) {
            console.error("Failed to create team:", error);
            toast.error(error?.response?.data?.error || "Failed to create team");
        } finally {
            setSubmitting(false);
        }
    };

    const handleJoinTeam = async () => {
        if (!inviteCode.trim()) {
            toast.error("Please enter an invite code");
            return;
        }

        setSubmitting(true);
        try {
            const team = await teamApi.joinTeam({ invite_code: inviteCode.trim() });
            toast.success(`Successfully joined team!`);

            onClose();
            navigate(`/tournaments/${team.tournament_id}`);
        } catch (error: any) {
            console.error("Failed to join team:", error);
            toast.error(error?.response?.data?.error || "Failed to join team");
        } finally {
            setSubmitting(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center p-4 z-50">
            <div className="bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
                {/* Header */}
                <div className="sticky top-0 bg-gray-800 border-b border-gray-700 p-6 flex justify-between items-center">
                    <h2 className="text-2xl font-bold text-white">Join Warzone Tournament</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-white text-2xl"
                    >
                        ×
                    </button>
                </div>

                <div className="p-6">
                    {loading ? (
                        <div className="text-center py-12 text-gray-400">
                            Loading tournaments...
                        </div>
                    ) : tournaments.length === 0 ? (
                        <div className="text-center py-12">
                            <div className="text-6xl mb-4">🏆</div>
                            <p className="text-xl text-gray-300 mb-2">
                                No open Warzone tournaments available
                            </p>
                            <p className="text-gray-400 mb-6">
                                Tournaments must be in "open" status to accept registrations
                            </p>
                            {user?.role === "admin" && (
                                <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4 mb-4 text-left max-w-md mx-auto">
                                    <p className="text-blue-300 text-sm mb-2">
                                        📋 <strong>Admin Instructions:</strong>
                                    </p>
                                    <ol className="text-blue-200 text-sm space-y-1 list-decimal list-inside">
                                        <li>Go to Tournament Admin</li>
                                        <li>Select a tournament</li>
                                        <li>Click "Open Registration"</li>
                                    </ol>
                                </div>
                            )}
                            <div className="flex gap-3 justify-center">
                                {user?.role === "admin" && (
                                    <Link to="/tournament-admin">
                                        <Button variant="primary">
                                            Go to Tournament Admin
                                        </Button>
                                    </Link>
                                )}
                                <Button onClick={onClose} variant="secondary">
                                    Close
                                </Button>
                            </div>
                        </div>
                    ) : !selectedTournament ? (
                        /* Tournament Selection */
                        <div>
                            <h3 className="text-lg font-bold text-white mb-4">
                                Select a Tournament
                            </h3>

                            {/* Player Stats Summary */}
                            {warzoneStats && (
                                <div className="bg-gray-700 rounded-lg p-4 mb-6">
                                    <h4 className="text-sm font-medium text-gray-300 mb-2">
                                        Your Warzone Stats
                                    </h4>
                                    <div className="grid grid-cols-3 gap-4 text-sm">
                                        <div>
                                            <span className="text-gray-400">K/D Ratio:</span>
                                            <p className="text-white font-bold">
                                                {Number(warzoneStats.stats.kd || 0).toFixed(2)}
                                            </p>
                                        </div>
                                        <div>
                                            <span className="text-gray-400">Matches Played:</span>
                                            <p className="text-white font-bold">
                                                {warzoneStats.matches_played}
                                            </p>
                                        </div>
                                        <div>
                                            <span className="text-gray-400">Ranking Score:</span>
                                            <p className="text-white font-bold">
                                                {warzoneStats.ranking_score.toFixed(0)}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            )}

                            <div className="space-y-4">
                                {tournaments.map((tournament) => {
                                    const eligibility = checkEligibility(tournament);

                                    return (
                                        <div
                                            key={tournament.id}
                                            className={`border rounded-lg p-4 ${eligibility.eligible
                                                ? "border-gray-600 bg-gray-750 hover:border-blue-500 cursor-pointer"
                                                : "border-red-900 bg-red-900/10 opacity-75"
                                                }`}
                                            onClick={() =>
                                                eligibility.eligible &&
                                                handleSelectTournament(tournament)
                                            }
                                        >
                                            <div className="flex justify-between items-start">
                                                <div className="flex-1">
                                                    <h4 className="text-lg font-bold text-white">
                                                        {tournament.name}
                                                    </h4>
                                                    {tournament.description && (
                                                        <p className="text-gray-400 text-sm mt-1">
                                                            {tournament.description}
                                                        </p>
                                                    )}
                                                    <div className="flex gap-4 mt-3 text-sm">
                                                        <span className="text-gray-400">
                                                            Team Size:{" "}
                                                            <span className="text-white font-medium">
                                                                {teamSizeToLabel(tournament.team_size)}
                                                            </span>
                                                        </span>
                                                        {tournament.rules.max_teams > 0 && (
                                                            <span className="text-gray-400">
                                                                Teams:{" "}
                                                                <span className="text-white font-medium">
                                                                    {tournament.current_teams || 0} /{" "}
                                                                    {tournament.rules.max_teams}
                                                                </span>
                                                            </span>
                                                        )}
                                                        {tournament.prize_pool && (
                                                            <span className="text-gray-400">
                                                                Prize:{" "}
                                                                <span className="text-green-400 font-medium">
                                                                    {tournament.prize_pool}
                                                                </span>
                                                            </span>
                                                        )}
                                                    </div>

                                                    {/* Requirements */}
                                                    {(tournament.rules.max_kd !== undefined ||
                                                        tournament.rules.min_kd !== undefined ||
                                                        tournament.rules.min_matches_played !==
                                                        undefined) && (
                                                            <div className="mt-3 text-xs text-gray-400">
                                                                <span className="font-medium">
                                                                    Requirements:
                                                                </span>{" "}
                                                                {tournament.rules.min_kd !== undefined && (
                                                                    <span>
                                                                        Min K/D: {tournament.rules.min_kd}
                                                                    </span>
                                                                )}
                                                                {tournament.rules.max_kd !== undefined && (
                                                                    <span className="ml-2">
                                                                        Max K/D: {tournament.rules.max_kd}
                                                                    </span>
                                                                )}
                                                                {tournament.rules.min_matches_played !==
                                                                    undefined && (
                                                                        <span className="ml-2">
                                                                            Min Matches:{" "}
                                                                            {tournament.rules.min_matches_played}
                                                                        </span>
                                                                    )}
                                                            </div>
                                                        )}
                                                </div>

                                                {eligibility.eligible ? (
                                                    <div className="ml-4">
                                                        <span className="inline-block px-3 py-1 bg-green-500/20 text-green-400 rounded text-sm font-medium">
                                                            Eligible
                                                        </span>
                                                    </div>
                                                ) : (
                                                    <div className="ml-4">
                                                        <span className="inline-block px-3 py-1 bg-red-500/20 text-red-400 rounded text-sm font-medium">
                                                            Not Eligible
                                                        </span>
                                                    </div>
                                                )}
                                            </div>

                                            {!eligibility.eligible && (
                                                <div className="mt-3 text-sm text-red-400">
                                                    {eligibility.reasons.map((reason, idx) => (
                                                        <div key={idx}>• {reason}</div>
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    ) : (
                        /* Team Creation/Join */
                        <div>
                            <button
                                onClick={() => setSelectedTournament(null)}
                                className="text-blue-400 hover:text-blue-300 mb-4 text-sm"
                            >
                                ← Back to tournaments
                            </button>

                            <div className="bg-gray-700 rounded-lg p-4 mb-6">
                                <h3 className="text-xl font-bold text-white">
                                    {selectedTournament.name}
                                </h3>
                                <p className="text-gray-400 text-sm mt-1">
                                    {selectedTournament.description}
                                </p>
                            </div>

                            {!joinMethod ? (
                                <div className="space-y-4">
                                    <h4 className="text-lg font-medium text-white">
                                        How would you like to participate?
                                    </h4>

                                    <button
                                        onClick={() => setJoinMethod("create")}
                                        className="w-full border-2 border-gray-600 hover:border-blue-500 rounded-lg p-6 text-left transition-colors"
                                    >
                                        <h5 className="text-lg font-bold text-white mb-2">
                                            Create a New Team
                                        </h5>
                                        <p className="text-gray-400 text-sm">
                                            Start your own team and invite other players to join
                                        </p>
                                    </button>

                                    <button
                                        onClick={() => setJoinMethod("join")}
                                        className="w-full border-2 border-gray-600 hover:border-green-500 rounded-lg p-6 text-left transition-colors"
                                    >
                                        <h5 className="text-lg font-bold text-white mb-2">
                                            Join Existing Team
                                        </h5>
                                        <p className="text-gray-400 text-sm">
                                            Have an invite code? Join a team that's already formed
                                        </p>
                                    </button>
                                </div>
                            ) : joinMethod === "create" ? (
                                <div className="space-y-4">
                                    <button
                                        onClick={() => setJoinMethod(null)}
                                        className="text-blue-400 hover:text-blue-300 text-sm"
                                    >
                                        ← Change method
                                    </button>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Team Name *
                                        </label>
                                        <Input
                                            type="text"
                                            value={teamName}
                                            onChange={(e) => setTeamName(e.target.value)}
                                            placeholder="Enter team name"
                                            maxLength={50}
                                        />
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Team Tag (optional)
                                        </label>
                                        <Input
                                            type="text"
                                            value={teamTag}
                                            onChange={(e) => setTeamTag(e.target.value)}
                                            placeholder="e.g., WZ, PRO, etc."
                                            maxLength={10}
                                        />
                                    </div>

                                    <div className="bg-blue-500/10 border border-blue-500/30 rounded p-4 text-sm text-blue-300">
                                        <p className="font-medium mb-1">📋 What happens next:</p>
                                        <ul className="list-disc list-inside space-y-1 text-blue-200">
                                            <li>You'll become the team captain</li>
                                            <li>You'll receive an invite code to share</li>
                                            <li>
                                                Invite {selectedTournament.team_size - 1} more player
                                                {selectedTournament.team_size - 1 !== 1 ? "s" : ""}
                                            </li>
                                        </ul>
                                    </div>

                                    <div className="flex gap-3">
                                        <Button
                                            onClick={handleCreateTeam}
                                            disabled={submitting || !teamName.trim()}
                                            className="flex-1"
                                        >
                                            {submitting ? "Creating..." : "Create Team"}
                                        </Button>
                                        <Button
                                            onClick={() => setJoinMethod(null)}
                                            variant="secondary"
                                        >
                                            Cancel
                                        </Button>
                                    </div>
                                </div>
                            ) : (
                                <div className="space-y-4">
                                    <button
                                        onClick={() => setJoinMethod(null)}
                                        className="text-blue-400 hover:text-blue-300 text-sm"
                                    >
                                        ← Change method
                                    </button>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">
                                            Invite Code *
                                        </label>
                                        <Input
                                            type="text"
                                            value={inviteCode}
                                            onChange={(e) => setInviteCode(e.target.value)}
                                            placeholder="Enter 6-character invite code"
                                            maxLength={6}
                                        />
                                        <p className="text-xs text-gray-400 mt-1">
                                            Ask your team captain for the invite code
                                        </p>
                                    </div>

                                    <div className="flex gap-3">
                                        <Button
                                            onClick={handleJoinTeam}
                                            disabled={submitting || !inviteCode.trim()}
                                            className="flex-1"
                                        >
                                            {submitting ? "Joining..." : "Join Team"}
                                        </Button>
                                        <Button
                                            onClick={() => setJoinMethod(null)}
                                            variant="secondary"
                                        >
                                            Cancel
                                        </Button>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
