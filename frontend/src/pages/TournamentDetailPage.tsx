import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";
import { toast } from "../lib/toast";
import { teamApi, type CreateTeamRequest } from "../services/teams";
import { tournamentApi } from "../services/tournaments";
import { useAuthStore } from "../store/authStore";

type JoinMethod = "create" | "join" | null;

export function TournamentDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const user = useAuthStore((state) => state.user);
    const isAdmin = user?.role === "admin";

    const [joinMethod, setJoinMethod] = useState<JoinMethod>(null);
    const [teamName, setTeamName] = useState("");
    const [teamTag, setTeamTag] = useState("");
    const [inviteCode, setInviteCode] = useState("");
    const [submitting, setSubmitting] = useState(false);

    const { data: tournament, isLoading, error, refetch } = useQuery({
        queryKey: ["tournament", id],
        queryFn: async () => {
            if (!id) throw new Error("Tournament ID is required");
            return await tournamentApi.getTournament(id);
        },
        enabled: !!id,
    });

    // Determine if user can admin this tournament (creator or global admin)
    const isCreator = user?.id && tournament?.created_by === user.id;
    const canManageTournament = isAdmin || isCreator;

    const handleNavigateBack = () => {
        navigate("/tournaments");
    };

    const handleStatusChange = async (newStatus: "draft" | "open" | "active" | "finished" | "canceled") => {
        if (!tournament?.id) return;

        try {
            if (newStatus === "active") {
                if (!confirm("Are you sure you want to start this tournament?")) {
                    return;
                }
            } else if (newStatus === "finished") {
                if (!confirm("Are you sure you want to finish this tournament?")) {
                    return;
                }
            }

            await tournamentApi.updateTournamentStatus(tournament.id, newStatus);
            toast.success(`Tournament ${newStatus}`);
            // Refetch the tournament data instead of full page reload
            await refetch();
        } catch (err) {
            toast.error(err instanceof Error ? err.message : "Failed to update tournament");
        }
    };

    const handleCreateTeam = async () => {
        if (!tournament || !teamName.trim()) {
            toast.error("Please enter a team name");
            return;
        }

        setSubmitting(true);
        try {
            const createData: CreateTeamRequest = {
                tournament_id: tournament.id,
                name: teamName.trim(),
                tag: teamTag.trim() || undefined,
            };

            const team = await teamApi.createTeam(createData);
            toast.success(`Team "${team.name}" created successfully!`);
            setJoinMethod(null);
            setTeamName("");
            setTeamTag("");
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
            const result = await teamApi.joinTeam({ invite_code: inviteCode.trim() });
            toast.success(`Successfully joined ${result.name}!`);
            setJoinMethod(null);
            setInviteCode("");
            // Refresh after success
            setTimeout(() => window.location.reload(), 1500);
        } catch (error: any) {
            console.error("Failed to join team:", error);
            const errorMsg = error?.response?.data?.error || error?.message || "Failed to join team";
            toast.error(errorMsg);
        } finally {
            setSubmitting(false);
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p>Loading tournament...</p>
                </div>
            </div>
        );
    }

    if (error || !tournament) {
        return (
            <div className="max-w-4xl mx-auto px-4 py-8">
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                    <p className="font-bold">Error loading tournament</p>
                    <p>{error instanceof Error ? error.message : "Tournament not found"}</p>
                </div>
                <Button onClick={handleNavigateBack} className="mt-4">
                    Back to Tournaments
                </Button>
            </div>
        );
    }

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
        <div className="max-w-4xl mx-auto px-4 py-8">
            {/* Header with back button */}
            <div className="mb-8">
                <Button onClick={handleNavigateBack} variant="secondary" size="sm">
                    ← Back to Tournaments
                </Button>
            </div>

            {/* Tournament Info */}
            <div className="bg-white rounded-lg shadow-md p-6 mb-8">
                <div className="flex justify-between items-start mb-4">
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900">{tournament.name}</h1>
                        <p className="text-gray-600 mt-2">{tournament.description}</p>
                    </div>
                    <span
                        className={`px-4 py-2 rounded-full font-medium ${getStatusBadgeClass(
                            tournament.status
                        )}`}
                    >
                        {tournament.status.toUpperCase()}
                    </span>
                </div>

                <div className="grid grid-cols-2 gap-4 mt-6">
                    <div>
                        <p className="text-sm text-gray-500">Start Date</p>
                        <p className="text-lg font-semibold">
                            {new Date(tournament.start_date).toLocaleDateString()}
                        </p>
                    </div>
                    <div>
                        <p className="text-sm text-gray-500">End Date</p>
                        <p className="text-lg font-semibold">
                            {new Date(tournament.end_date).toLocaleDateString()}
                        </p>
                    </div>
                    <div>
                        <p className="text-sm text-gray-500">Team Size</p>
                        <p className="text-lg font-semibold">{tournament.team_size} per team</p>
                    </div>
                    {tournament.lobby_code && (
                        <div>
                            <p className="text-sm text-gray-500">Lobby Code</p>
                            <p className="text-lg font-semibold font-mono">{tournament.lobby_code}</p>
                        </div>
                    )}
                </div>

                {/* Admin Controls */}
                {canManageTournament && (
                    <div className="mt-6 pt-6 border-t border-gray-200">
                        <p className="text-sm font-semibold text-gray-600 mb-3">Tournament Management</p>
                        <div className="flex gap-2 flex-wrap">
                            {tournament.status === "draft" && (
                                <Button
                                    onClick={() => handleStatusChange("open")}
                                    variant="primary"
                                    size="sm"
                                >
                                    Open Registration
                                </Button>
                            )}
                            {tournament.status === "open" && (
                                <Button
                                    onClick={() => handleStatusChange("active")}
                                    variant="primary"
                                    size="sm"
                                >
                                    Start Tournament
                                </Button>
                            )}
                            {tournament.status === "active" && (
                                <Button
                                    onClick={() => handleStatusChange("finished")}
                                    variant="primary"
                                    size="sm"
                                >
                                    Finish Tournament
                                </Button>
                            )}
                            {tournament.status !== "canceled" && tournament.status !== "finished" && (
                                <Button
                                    onClick={() => handleStatusChange("canceled")}
                                    variant="danger"
                                    size="sm"
                                >
                                    Cancel Tournament
                                </Button>
                            )}
                        </div>
                    </div>
                )}
            </div>

            {/* Join/Create Team Section */}
            {tournament.status !== "finished" && tournament.status !== "canceled" && (
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-8">
                    <h2 className="text-xl font-bold text-gray-900 mb-4">Team Registration</h2>
                    <p className="text-sm text-gray-600 mb-4">Tournament Status: <span className="font-semibold uppercase">{tournament.status}</span></p>

                    {joinMethod === null && (
                        <div className="flex gap-4 flex-wrap">
                            <Button
                                onClick={() => setJoinMethod("create")}
                                variant="primary"
                            >
                                ✨ Create Team
                            </Button>
                            <Button
                                onClick={() => setJoinMethod("join")}
                                variant="secondary"
                            >
                                📋 Join with Code
                            </Button>
                        </div>
                    )}

                    {joinMethod === "create" && (
                        <div className="mt-4 space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-900 mb-2">
                                    Team Name
                                </label>
                                <Input
                                    placeholder="Enter team name"
                                    value={teamName}
                                    onChange={(e) => setTeamName(e.target.value)}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-900 mb-2">
                                    Team Tag (optional)
                                </label>
                                <Input
                                    placeholder="e.g., XXXX"
                                    value={teamTag}
                                    onChange={(e) => setTeamTag(e.target.value)}
                                    maxLength={4}
                                />
                            </div>
                            <div className="flex gap-2">
                                <Button
                                    onClick={handleCreateTeam}
                                    variant="primary"
                                    disabled={submitting}
                                >
                                    {submitting ? "Creating..." : "Create Team"}
                                </Button>
                                <Button
                                    onClick={() => {
                                        setJoinMethod(null);
                                        setTeamName("");
                                        setTeamTag("");
                                    }}
                                    variant="secondary"
                                >
                                    Cancel
                                </Button>
                            </div>
                        </div>
                    )}

                    {joinMethod === "join" && (
                        <div className="mt-4 space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-900 mb-2">
                                    Invite Code
                                </label>
                                <p className="text-sm text-gray-600 mb-3">
                                    Ask your team captain for the invite code (8 character code)
                                </p>
                                <Input
                                    placeholder="e.g., a1b2c3d4"
                                    value={inviteCode}
                                    onChange={(e) => setInviteCode(e.target.value.toUpperCase())}
                                    maxLength={8}
                                />
                            </div>
                            <div className="flex gap-2">
                                <Button
                                    onClick={handleJoinTeam}
                                    variant="primary"
                                    disabled={submitting || !inviteCode.trim()}
                                >
                                    {submitting ? "Joining..." : "Join Team"}
                                </Button>
                                <Button
                                    onClick={() => {
                                        setJoinMethod(null);
                                        setInviteCode("");
                                    }}
                                    variant="secondary"
                                >
                                    Cancel
                                </Button>
                            </div>
                        </div>
                    )}
                </div>
            )}

            {/* Tournament Rules */}
            {tournament.rules && (
                <div className="bg-white rounded-lg shadow-md p-6">
                    <h2 className="text-xl font-bold text-gray-900 mb-4">Tournament Rules</h2>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <p className="text-sm text-gray-500">Max Teams</p>
                            <p className="text-lg font-semibold">{tournament.rules.max_teams}</p>
                        </div>
                        <div>
                            <p className="text-sm text-gray-500">Min Matches</p>
                            <p className="text-lg font-semibold">{tournament.rules.min_matches}</p>
                        </div>
                        <div>
                            <p className="text-sm text-gray-500">Max Matches</p>
                            <p className="text-lg font-semibold">{tournament.rules.max_matches}</p>
                        </div>
                        <div>
                            <p className="text-sm text-gray-500">Late Registration</p>
                            <p className="text-lg font-semibold">
                                {tournament.rules.allow_late_registration ? "Allowed" : "Not Allowed"}
                            </p>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
