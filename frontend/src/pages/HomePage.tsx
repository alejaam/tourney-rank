import { Link } from 'react-router-dom';
import { Button, Card, CardContent } from '../components/ui';
import type { AuthState } from '../store/authStore';
import { useAuthStore } from '../store/authStore';

export const HomePage = () => {
    const isAuthenticated = useAuthStore((state: AuthState) => state.isAuthenticated);

    return (
        <div className="min-h-screen bg-gray-900">
            {/* Hero Section */}
            <div className="relative overflow-hidden">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
                    <div className="text-center">
                        <h1 className="text-5xl md:text-6xl font-extrabold text-white mb-6">
                            <span className="text-blue-500">Tourney</span>Rank
                        </h1>
                        <p className="text-xl md:text-2xl text-gray-400 mb-8 max-w-2xl mx-auto">
                            Multi-game tournament management platform with automated ranking system for competitive gaming communities.
                        </p>
                        <div className="flex gap-4 justify-center">
                            {isAuthenticated ? (
                                <Link to="/dashboard">
                                    <Button size="lg">Go to Dashboard</Button>
                                </Link>
                            ) : (
                                <>
                                    <Link to="/register">
                                        <Button size="lg">Get Started</Button>
                                    </Link>
                                    <Link to="/login">
                                        <Button size="lg" variant="secondary">Sign In</Button>
                                    </Link>
                                </>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Features Section */}
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
                <h2 className="text-3xl font-bold text-white text-center mb-12">Features</h2>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                    <Card>
                        <CardContent className="text-center py-8">
                            <div className="text-4xl mb-4">🎮</div>
                            <h3 className="text-xl font-semibold text-white mb-2">Multi-Game Support</h3>
                            <p className="text-gray-400">
                                Warzone, Fortnite, Apex, Valorant, and more. Flexible schema supports any game.
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="text-center py-8">
                            <div className="text-4xl mb-4">📊</div>
                            <h3 className="text-xl font-semibold text-white mb-2">Automated Ranking</h3>
                            <p className="text-gray-400">
                                Smart calculation based on K/D, damage, consistency, and game-specific metrics.
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="text-center py-8">
                            <div className="text-4xl mb-4">🏆</div>
                            <h3 className="text-xl font-semibold text-white mb-2">Live Leaderboards</h3>
                            <p className="text-gray-400">
                                Real-time updates via WebSockets. Watch rankings change as matches complete.
                            </p>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* Footer */}
            <footer className="border-t border-gray-800 py-8">
                <div className="max-w-7xl mx-auto px-4 text-center text-gray-500">
                    <p>© 2026 TourneyRank. Built for competitive gamers.</p>
                </div>
            </footer>
        </div>
    );
};
