import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { BracketGenerator, BracketView } from "../components/bracket";
import { Button } from "../components/ui/Button";
import { toast } from "../lib/toast";
import { teamApi } from "../services/teams";
import { tournamentApi } from "../services/tournaments";
import { useAuthStore } from "../store/authStore";
import type { Team, Tournament } from "../types/api";

export function TournamentAdminPage() {
    const navigate = useNavigate();
    const user = useAuthStore((state) => state.user);
    const [searchParams, setSearchParams] = useSearchParams();
    const [tournaments, setTournaments] = useState<Tournament[]>([]);
    const [selectedTournament, setSelectedTournament] = useState<Tournament | null>(null);
    const [teams, setTeams] = useState<Team[]>([]);
    const [teamsLoading, setTeamsLoading] = useState(false);
    const [activeTab, setActiveTab] = useState<"overview" | "teams" | "bracket" | "matches">(
        "overview"
    );
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadTournaments();
    }, []);

    useEffect(() => {
        const tournamentId = searchParams.get("tournament");
        if (tournamentId && tournaments.length > 0) {
            const tournament = tournaments.find((t) => t.id === tournamentId);
            if (tournament) {
                setSelectedTournament(tournament);
            }
        }
    }, [searchParams, tournaments]);

    useEffect(() => {
        if (selectedTournament) {
            loadTeamsForTournament(selectedTournament.id);
        }
    }, [selectedTournament]);

    const loadTeamsForTournament = async (tournamentId: string) => {
        setTeamsLoading(true);
        try {
            const teamsData = await teamApi.getTournamentTeams(tournamentId);
            setTeams(teamsData || []);
        } catch (err) {
            console.error("Failed to load teams:", err);
            setTeams([]);
        } finally {
            setTeamsLoading(false);
        }
    };

    const loadTournaments = async () => {
        try {
            const response = await tournamentApi.listTournaments({ limit: 100 });
            setTournaments(response.tournaments || []);
        } catch (err) {
            console.error("Failed to load tournaments:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleTournamentSelect = (tournament: Tournament) => {
        setSelectedTournament(tournament);
        setSearchParams({ tournament: tournament.id });
        setActiveTab("overview");
    };

    const handleStatusChange = async (newStatus: Tournament["status"]) => {
        if (!selectedTournament) return;

        // Validation for starting tournament
        if (newStatus === "active") {
            if (teams.length < 2) {
                toast.error(`Need at least 2 teams to start tournament (currently ${teams.length})`);
                return;
            }
            if (!confirm(`Start tournament with ${teams.length} teams?`)) {
                return;
            }
        }

        try {
            await tournamentApi.updateTournamentStatus(selectedTournament.id, newStatus);
            toast.success(`Tournament ${newStatus}`);
            // Reload tournaments to get updated data
            loadTournaments();
            // Update selected tournament
            const updated = await tournamentApi.getTournament(selectedTournament.id);
            setSelectedTournament(updated);
        } catch (err: any) {
            console.error("Failed to update status:", err);
            toast.error("Failed to update status: " + (err?.response?.data?.error || err?.message));
        }
    };

    if (loading) {
        return (
            <div className="container mx-auto p-6">
                <div className="text-center py-12">Loading tournaments...</div>
            </div>
        );
    }

    // Only admins can access tournament admin page
    if (user?.role !== "admin") {
        return (
            <div className="container mx-auto p-6">
                <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                    <p className="font-bold">Access Denied</p>
                    <p>Only administrators can access the tournament administration panel.</p>
                </div>
                <Button onClick={() => navigate("/tournaments")}>
                    Back to Tournaments
                </Button>
            </div>
        );
    }

    return (
        <div className="container mx-auto p-6">
            <div className="flex items-center justify-between mb-6">
                <h1 className="text-3xl font-bold">Tournament Administration</h1>
                <Button variant="primary" onClick={() => navigate("/tournaments")}>Create Tournament</Button>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                {/* Tournament List Sidebar */}
                <div className="lg:col-span-1">
                    <div className="bg-white rounded-lg shadow p-4">
                        <h2 className="text-lg font-bold mb-4">Your Tournaments</h2>
                        <div className="space-y-2">
                            {tournaments.length === 0 ? (
                                <p className="text-gray-500 text-sm">No tournaments yet</p>
                            ) : (
                                tournaments.map((tournament) => (
                                    <button
                                        key={tournament.id}
                                        onClick={() => handleTournamentSelect(tournament)}
                                        className={`w-full text-left p-3 rounded border ${selectedTournament?.id === tournament.id
                                            ? "border-blue-500 bg-blue-50"
                                            : "border-gray-200 hover:bg-gray-50"
                                            }`}
                                    >
                                        <div className="font-medium truncate">{tournament.name}</div>
                                        <div className="text-xs text-gray-500 mt-1">
                                            <span
                                                className={`inline-block px-2 py-0.5 rounded ${tournament.status === "active"
                                                    ? "bg-green-100 text-green-800"
                                                    : tournament.status === "open"
                                                        ? "bg-blue-100 text-blue-800"
                                                        : "bg-gray-100 text-gray-800"
                                                    }`}
                                            >
                                                {tournament.status}
                                            </span>
                                        </div>
                                    </button>
                                ))
                            )}
                        </div>
                    </div>
                </div>

                {/* Main Content */}
                <div className="lg:col-span-3">
                    {!selectedTournament ? (
                        <div className="bg-gray-50 border border-gray-200 rounded-lg p-12 text-center">
                            <p className="text-gray-600 mb-4">
                                Select a tournament from the list to manage it
                            </p>
                            <Button variant="primary" onClick={() => navigate("/tournaments")}>
                                Create Tournament
                            </Button>
                        </div>
                    ) : (
                        <div className="space-y-6">
                            {/* Tournament Header */}
                            <div className="bg-white rounded-lg shadow p-6">
                                <div className="flex justify-between items-start">
                                    <div>
                                        <h2 className="text-2xl font-bold">{selectedTournament.name}</h2>
                                        <p className="text-gray-600 mt-1">
                                            {selectedTournament.description}
                                        </p>
                                        <div className="flex gap-4 mt-3 text-sm text-gray-600">
                                            <span>
                                                Status:{" "}
                                                <span className="font-medium">
                                                    {selectedTournament.status}
                                                </span>
                                            </span>
                                            <span>
                                                Team Size:{" "}
                                                <span className="font-medium">
                                                    {selectedTournament.team_size}
                                                </span>
                                            </span>
                                        </div>
                                    </div>

                                    {/* Status Controls */}
                                    <div className="flex flex-col gap-2">
                                        {selectedTournament.status === "draft" && (
                                            <Button
                                                variant="primary"
                                                size="sm"
                                                onClick={() => handleStatusChange("open")}
                                            >
                                                Open Registration
                                            </Button>
                                        )}
                                        {selectedTournament.status === "open" && (
                                            <Button
                                                variant="primary"
                                                size="sm"
                                                onClick={() => handleStatusChange("active")}
                                            >
                                                Start Tournament
                                            </Button>
                                        )}
                                        {selectedTournament.status === "active" && (
                                            <Button
                                                variant="secondary"
                                                size="sm"
                                                onClick={() => handleStatusChange("finished")}
                                            >
                                                Finish Tournament
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            </div>

                            {/* Tabs */}
                            <div className="bg-white rounded-lg shadow">
                                <div className="border-b border-gray-200">
                                    <nav className="flex">
                                        {[
                                            { id: "overview", label: "Overview" },
                                            { id: "teams", label: "Teams" },
                                            { id: "bracket", label: "Bracket" },
                                            { id: "matches", label: "Matches" },
                                        ].map((tab) => (
                                            <button
                                                key={tab.id}
                                                onClick={() =>
                                                    setActiveTab(
                                                        tab.id as typeof activeTab
                                                    )
                                                }
                                                className={`px-6 py-3 font-medium border-b-2 ${activeTab === tab.id
                                                    ? "border-blue-500 text-blue-600"
                                                    : "border-transparent text-gray-500 hover:text-gray-700"
                                                    }`}
                                            >
                                                {tab.label}
                                            </button>
                                        ))}
                                    </nav>
                                </div>

                                <div className="p-6">
                                    {activeTab === "overview" && (
                                        <div className="space-y-4">
                                            <h3 className="text-lg font-bold">Tournament Details</h3>
                                            <div className="grid grid-cols-2 gap-4 text-sm">
                                                <div>
                                                    <span className="text-gray-600">Start Date:</span>
                                                    <p className="font-medium">
                                                        {new Date(
                                                            selectedTournament.start_date
                                                        ).toLocaleDateString()}
                                                    </p>
                                                </div>
                                                <div>
                                                    <span className="text-gray-600">End Date:</span>
                                                    <p className="font-medium">
                                                        {new Date(
                                                            selectedTournament.end_date
                                                        ).toLocaleDateString()}
                                                    </p>
                                                </div>
                                                {selectedTournament.prize_pool && (
                                                    <div>
                                                        <span className="text-gray-600">
                                                            Prize Pool:
                                                        </span>
                                                        <p className="font-medium">
                                                            {selectedTournament.prize_pool}
                                                        </p>
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    )}

                                    {activeTab === "teams" && (
                                        <div>
                                            <h3 className="text-lg font-bold mb-4">Registered Teams ({teams.length})</h3>
                                            {teamsLoading ? (
                                                <p className="text-gray-600">Loading teams...</p>
                                            ) : teams.length === 0 ? (
                                                <p className="text-gray-600">No teams registered yet</p>
                                            ) : (
                                                <div className="space-y-3">
                                                    {teams.map((t) => (
                                                        <div key={t.id} className="border border-gray-200 rounded-lg p-4">
                                                            <div className="flex justify-between items-start">
                                                                <div>
                                                                    <h4 className="font-semibold text-lg">{t.name}</h4>
                                                                    {t.tag && <p className="text-sm text-gray-600">Tag: {t.tag}</p>}
                                                                    <p className="text-sm text-gray-600">Captain: {t.captain_name || 'Unknown'}</p>
                                                                    <p className="text-sm text-gray-600">Members: {t.member_ids?.length || 1}</p>
                                                                </div>
                                                                <div className="text-right">
                                                                    <span className={`inline-block px-3 py-1 rounded text-sm font-medium ${t.status === "ready" ? "bg-green-100 text-green-800" :
                                                                            t.status === "pending" ? "bg-yellow-100 text-yellow-800" :
                                                                                "bg-gray-100 text-gray-800"
                                                                        }`}>
                                                                        {t.status}
                                                                    </span>
                                                                </div>
                                                            </div>
                                                            <div className="mt-3 pt-3 border-t border-gray-200">
                                                                <p className="text-xs text-gray-500">Invite Code: <code className="bg-gray-100 px-2 py-1 rounded">{t.invite_code}</code></p>
                                                            </div>
                                                        </div>
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    )}

                                    {activeTab === "bracket" && (
                                        <div className="space-y-6">
                                            {(selectedTournament.status === "open" ||
                                                selectedTournament.status === "draft") && (
                                                    <BracketGenerator
                                                        tournamentId={selectedTournament.id}
                                                        onSuccess={() => {
                                                            loadTournaments();
                                                        }}
                                                    />
                                                )}
                                            <BracketView
                                                tournamentId={selectedTournament.id}
                                                isAdmin={true}
                                            />
                                        </div>
                                    )}

                                    {activeTab === "matches" && (
                                        <div>
                                            <h3 className="text-lg font-bold mb-4">Match Results</h3>
                                            <p className="text-gray-600">
                                                Match verification coming soon...
                                            </p>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
