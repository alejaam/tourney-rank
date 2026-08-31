import { useState } from "react";
import { errorMessage } from "../../lib/error";
import type { BracketFormat, GenerateBracketRequest } from "../../types/api";
import { Button } from "../ui/Button";

interface BracketGeneratorProps {
    tournamentId: string;
    onSuccess?: () => void;
}

export function BracketGenerator({ tournamentId, onSuccess }: BracketGeneratorProps) {
    const [format, setFormat] = useState<BracketFormat>("single_elimination");
    const [isSeeded, setIsSeeded] = useState<boolean>(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleGenerate = async () => {
        setError(null);
        setLoading(true);

        try {
            const request: GenerateBracketRequest = {
                tournament_id: tournamentId,
                format,
                is_seeded: isSeeded,
            };

            const { bracketApi } = await import("../../services/brackets");
            await bracketApi.generateBracket(request);
            onSuccess?.();
        } catch (err: unknown) {
            console.error("Failed to generate bracket:", err);
            setError(errorMessage(err, "Failed to generate bracket"));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-white p-6 rounded-lg shadow-md">
            <h2 className="text-xl font-bold mb-4">Generate Tournament Bracket</h2>

            {error && (
                <div className="bg-red-50 border border-red-200 rounded p-4 mb-4">
                    <p className="text-red-700">⚠️ {error}</p>
                </div>
            )}

            <div className="space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Bracket Format
                    </label>
                    <select
                        value={format}
                        onChange={(e) => setFormat(e.target.value as BracketFormat)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="single_elimination">Single Elimination</option>
                        <option value="round_robin">Round Robin (All vs All)</option>
                        <option value="double_elimination" disabled>
                            Double Elimination (Coming Soon)
                        </option>
                        <option value="swiss" disabled>
                            Swiss System (Coming Soon)
                        </option>
                    </select>
                    <p className="text-sm text-gray-500 mt-1">
                        {format === "single_elimination" &&
                            "Teams are eliminated after one loss. Fast and exciting."}
                        {format === "round_robin" &&
                            "Every team plays every other team. Fair and comprehensive."}
                    </p>
                </div>

                <div className="flex items-center">
                    <input
                        type="checkbox"
                        id="is_seeded"
                        checked={isSeeded}
                        onChange={(e) => setIsSeeded(e.target.checked)}
                        className="mr-2"
                    />
                    <label htmlFor="is_seeded" className="text-sm text-gray-700">
                        Use seeded matchups (ranked teams get easier first matches)
                    </label>
                </div>

                <div className="flex gap-2 pt-4">
                    <Button variant="primary" onClick={handleGenerate} disabled={loading}>
                        {loading ? "Generating..." : "Generate Bracket"}
                    </Button>
                </div>

                <div className="text-sm text-gray-600 bg-blue-50 p-3 rounded">
                    <p className="font-medium mb-1">⚠️ Important</p>
                    <ul className="list-disc list-inside space-y-1">
                        <li>Make sure all teams have registered before generating</li>
                        <li>Once generated, the tournament status will change to "active"</li>
                        <li>You can regenerate if needed by deleting the current bracket</li>
                    </ul>
                </div>
            </div>
        </div>
    );
}
