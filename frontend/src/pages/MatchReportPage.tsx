import { MatchReportForm } from "../components/match/MatchReportForm";

export function MatchReportPage() {
    return (
        <div className="min-h-screen bg-gray-50 py-8">
            <div className="max-w-2xl mx-auto px-4">
                <h1 className="text-3xl font-bold text-gray-900 mb-2">Report Match</h1>
                <p className="text-gray-600 mb-8">
                    Submit your match results and statistics. Only team captains can submit reports.
                </p>

                <MatchReportForm
                    onSuccess={() => {
                        // Could navigate elsewhere or refresh data
                        console.log("Match submitted successfully");
                    }}
                />
            </div>
        </div>
    );
}
