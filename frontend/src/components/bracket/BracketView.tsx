import { useEffect, useState } from "react";
import { bracketApi } from "../../services/brackets";
import type { BracketWithMatchups, MatchupResponse } from "../../types/api";
import { Button } from "../ui/Button";

interface BracketViewProps {
    tournamentId: string;
    isAdmin?: boolean;
}

export function BracketView({ tournamentId, isAdmin = false }: BracketViewProps) {
    const [bracket, setBracket] = useState<BracketWithMatchups | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedMatchup, setSelectedMatchup] = useState<MatchupResponse | null>(null);
    const [winnerSelection, setWinnerSelection] = useState<string>("");

    useEffect(() => {
        loadBracket();
    }, [tournamentId]);

    const loadBracket = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await bracketApi.getTournamentBracket(tournamentId);
            setBracket(data);
        } catch (err: any) {
            console.error("Failed to load bracket:", err);
            setError(
                err?.response?.data?.error ||
                err?.message ||
                "Bracket not found for this tournament"
            );
        } finally {
            setLoading(false);
        }
    };

    const handleSetWinner = async () => {
        if (!selectedMatchup || !winnerSelection) return;

        try {
            await bracketApi.setMatchupWinner(selectedMatchup.id, {
                winner_id: winnerSelection,
            });
            setSelectedMatchup(null);
            setWinnerSelection("");
            loadBracket();
        } catch (err: any) {
            console.error("Failed to set winner:", err);
            alert("Failed to set winner: " + (err?.response?.data?.error || err?.message));
        }
    };

    if (loading) {
        return <div className="text-center py-8">Loading bracket...</div>;
    }

    if (error) {
        return (
            <div className="bg-yellow-50 border border-yellow-200 rounded p-4">
                <p className="text-yellow-700">{error}</p>
            </div>
        );
    }

    if (!bracket) {
        return (
            <div className="bg-gray-50 border border-gray-200 rounded p-4">
                <p className="text-gray-700">No bracket generated yet.</p>
            </div>
        );
    }

    // Group matchups by round
    const matchupsByRound: Record<number, MatchupResponse[]> = {};
    bracket.matchups.forEach((matchup) => {
        if (!matchupsByRound[matchup.round]) {
            matchupsByRound[matchup.round] = [];
        }
        matchupsByRound[matchup.round].push(matchup);
    });

    return (
        <div className="space-y-6">
            {/* Bracket Header */}
            <div className="bg-white p-4 rounded-lg shadow">
                <h2 className="text-xl font-bold mb-2">
                    {bracket.format.replace("_", " ").toUpperCase()} Bracket
                </h2>
                <div className="text-sm text-gray-600 space-y-1">
                    <p>
                        <span className="font-medium">Current Round:</span> {bracket.current_round}{" "}
                        / {bracket.total_rounds}
                    </p>
                    <p>
                        <span className="font-medium">Seeded:</span> {bracket.is_seeded ? "Yes" : "No"}
                    </p>
                    <p>
                        <span className="font-medium">Total Matches:</span> {bracket.matchups.length}
                    </p>
                </div>
            </div>

            {/* Rounds */}
            <div className="space-y-4">
                {Object.keys(matchupsByRound)
                    .map(Number)
                    .sort((a, b) => a - b)
                    .map((round) => (
                        <div key={round} className="bg-white p-4 rounded-lg shadow">
                            <h3 className="text-lg font-bold mb-3">
                                Round {round}
                                {round === bracket.current_round && (
                                    <span className="ml-2 text-sm font-normal text-blue-600">
                                        (Current)
                                    </span>
                                )}
                            </h3>
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                                {matchupsByRound[round].map((matchup) => (
                                    <div
                                        key={matchup.id}
                                        className={`border rounded p-3 ${matchup.status === "completed"
                                            ? "bg-green-50 border-green-200"
                                            : "bg-gray-50 border-gray-200"
                                            }`}
                                    >
                                        <div className="text-xs text-gray-500 mb-2">
                                            Match #{matchup.match_number}
                                        </div>

                                        {/* Team 1 */}
                                        <div
                                            className={`p-2 rounded mb-1 ${matchup.winner_id === matchup.team1_id
                                                ? "bg-green-200 font-bold"
                                                : "bg-white"
                                                }`}
                                        >
                                            {matchup.team1_name || "TBD"}
                                        </div>

                                        <div className="text-center text-xs text-gray-500 my-1">vs</div>

                                        {/* Team 2 */}
                                        <div
                                            className={`p-2 rounded mb-2 ${matchup.winner_id === matchup.team2_id
                                                ? "bg-green-200 font-bold"
                                                : "bg-white"
                                                }`}
                                        >
                                            {matchup.team2_name || "TBD"}
                                        </div>

                                        {/* Status */}
                                        <div className="text-xs text-gray-600 mt-2">
                                            Status: <span className="font-medium">{matchup.status}</span>
                                        </div>

                                        {/* Admin Controls */}
                                        {isAdmin &&
                                            matchup.status !== "completed" &&
                                            matchup.team1_id &&
                                            matchup.team2_id && (
                                                <Button
                                                    variant="secondary"
                                                    size="sm"
                                                    onClick={() => {
                                                        setSelectedMatchup(matchup);
                                                        setWinnerSelection("");
                                                    }}
                                                    className="mt-2 w-full"
                                                >
                                                    Set Winner
                                                </Button>
                                            )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    ))}
            </div>

            {/* Winner Selection Modal */}
            {selectedMatchup && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                    <div className="bg-white rounded-lg p-6 max-w-md w-full">
                        <h3 className="text-lg font-bold mb-4">Select Winner</h3>

                        <div className="space-y-2 mb-4">
                            {selectedMatchup.team1_id && (
                                <label className="flex items-center p-3 border rounded cursor-pointer hover:bg-gray-50">
                                    <input
                                        type="radio"
                                        name="winner"
                                        value={selectedMatchup.team1_id}
                                        checked={winnerSelection === selectedMatchup.team1_id}
                                        onChange={(e) => setWinnerSelection(e.target.value)}
                                        className="mr-3"
                                    />
                                    <span className="font-medium">
                                        {selectedMatchup.team1_name}
                                    </span>
                                </label>
                            )}

                            {selectedMatchup.team2_id && (
                                <label className="flex items-center p-3 border rounded cursor-pointer hover:bg-gray-50">
                                    <input
                                        type="radio"
                                        name="winner"
                                        value={selectedMatchup.team2_id}
                                        checked={winnerSelection === selectedMatchup.team2_id}
                                        onChange={(e) => setWinnerSelection(e.target.value)}
                                        className="mr-3"
                                    />
                                    <span className="font-medium">
                                        {selectedMatchup.team2_name}
                                    </span>
                                </label>
                            )}
                        </div>

                        <div className="flex gap-2">
                            <Button
                                variant="primary"
                                onClick={handleSetWinner}
                                disabled={!winnerSelection}
                                className="flex-1"
                            >
                                Confirm Winner
                            </Button>
                            <Button
                                variant="secondary"
                                onClick={() => {
                                    setSelectedMatchup(null);
                                    setWinnerSelection("");
                                }}
                                className="flex-1"
                            >
                                Cancel
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
