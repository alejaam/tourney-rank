import { useState } from "react";
import type { Tournament } from "../../types/api";
import { TournamentForm } from "./TournamentForm";
import { TournamentList } from "./TournamentList";

export function TournamentManager() {
    const [view, setView] = useState<"create" | "list">("list");
    const [selectedTournament, setSelectedTournament] = useState<Tournament | null>(null);

    const handleTournamentCreated = (tournament: Tournament) => {
        setSelectedTournament(tournament);
        setView("list");
    };

    const handleTournamentSelect = (tournament: Tournament) => {
        setSelectedTournament(tournament);
    };

    return (
        <div className="space-y-6">
            <div className="flex gap-4 border-b border-gray-200 pb-4">
                <button
                    onClick={() => {
                        setView("list");
                        setSelectedTournament(null);
                    }}
                    className={`px-4 py-2 font-medium transition ${view === "list"
                        ? "text-blue-600 border-b-2 border-blue-600"
                        : "text-gray-600 hover:text-gray-900"
                        }`}
                >
                    All Tournaments
                </button>
                <button
                    onClick={() => {
                        setView("create");
                        setSelectedTournament(null);
                    }}
                    className={`px-4 py-2 font-medium transition ${view === "create"
                        ? "text-blue-600 border-b-2 border-blue-600"
                        : "text-gray-600 hover:text-gray-900"
                        }`}
                >
                    Create Tournament
                </button>
            </div>

            {view === "create" ? (
                <div className="max-w-2xl">
                    <h2 className="text-2xl font-bold text-gray-900 mb-6">
                        {selectedTournament ? "Edit Tournament" : "Create New Tournament"}
                    </h2>
                    <TournamentForm
                        onSuccess={handleTournamentCreated}
                        initialData={selectedTournament || undefined}
                    />
                </div>
            ) : (
                <TournamentList
                    onTournamentSelect={handleTournamentSelect}
                    onTournamentDelete={() => setSelectedTournament(null)}
                    isAdmin={true}
                />
            )}
        </div>
    );
}
