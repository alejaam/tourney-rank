import { TournamentManager } from "../components/tournament";

export function TournamentsPage() {
    return (
        <div className="min-h-screen bg-gray-50 py-8">
            <div className="max-w-6xl mx-auto px-4">
                <h1 className="text-4xl font-bold text-gray-900 mb-2">Tournament Management</h1>
                <p className="text-gray-600 mb-8">
                    Create and manage tournaments for your competitive events.
                </p>

                <TournamentManager />
            </div>
        </div>
    );
}
