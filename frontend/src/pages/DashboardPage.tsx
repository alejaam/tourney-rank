import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { OnboardingBanner } from '../components/player';
import { JoinTournamentModal } from '../components/tournament';
import { Button, Card, CardContent, CardHeader, CardTitle } from '../components/ui';
import { useLogout } from '../features/auth/hooks';
import { playerApi } from '../services/player';
import { tournamentApi } from '../services/tournaments';
import type { AuthState } from '../store/authStore';
import { useAuthStore } from '../store/authStore';

export const DashboardPage = () => {
    const user = useAuthStore((state: AuthState) => state.user);
    const logout = useLogout();
    const [showJoinModal, setShowJoinModal] = useState(false);

    // Fetch player profile
    const { data: player, isLoading, error } = useQuery({
        queryKey: ['player', 'me'],
        queryFn: playerApi.getMyProfile,
        enabled: !!user,
    });

    // Fetch player's active tournament
    const { data: activeTournament } = useQuery({
        queryKey: ['player', 'active-tournament'],
        queryFn: tournamentApi.getPlayerActiveTournament,
        enabled: !!user,
    });

    return (
        <div className="min-h-screen bg-gray-900 p-6">
            <div className="max-w-4xl mx-auto">
                {/* Header */}
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h1 className="text-3xl font-bold text-white">Dashboard</h1>
                        <p className="text-gray-400">
                            Welcome back, {player?.display_name || user?.username}!
                        </p>
                    </div>
                    <div className="flex gap-3">
                        {user?.role === 'admin' && (
                            <>
                                <Link to="/admin">
                                    <Button variant="secondary">Admin Panel</Button>
                                </Link>
                                <Link to="/tournaments">
                                    <Button variant="secondary">Create Tournament</Button>
                                </Link>
                                <Link to="/tournament-admin">
                                    <Button variant="secondary">Manage Tournaments</Button>
                                </Link>
                            </>
                        )}
                        <Button variant="secondary" onClick={logout}>
                            Logout
                        </Button>
                    </div>
                </div>

                {/* Onboarding Banner - Show if player has default name */}
                {player && <OnboardingBanner player={player} />}

                {/* Player Profile Card */}
                {isLoading ? (
                    <Card className="mb-8">
                        <CardContent className="py-8">
                            <div className="text-center text-gray-400">
                                Loading profile...
                            </div>
                        </CardContent>
                    </Card>
                ) : error ? (
                    <Card className="mb-8">
                        <CardContent className="py-8">
                            <div className="text-center text-red-500">
                                Failed to load profile
                            </div>
                        </CardContent>
                    </Card>
                ) : player ? (
                    <Card className="mb-8">
                        <CardHeader className="flex flex-row items-center justify-between">
                            <CardTitle>Player Profile</CardTitle>
                            <Link to="/profile">
                                <Button size="sm" variant="secondary">
                                    View Full Profile
                                </Button>
                            </Link>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-start gap-4">
                                {player.avatar_url && (
                                    <img
                                        src={player.avatar_url}
                                        alt={player.display_name}
                                        className="w-16 h-16 rounded-full"
                                    />
                                )}
                                <div className="flex-1">
                                    <h3 className="text-xl font-bold text-white">
                                        {player.display_name}
                                    </h3>
                                    {player.bio && (
                                        <p className="text-gray-400 mt-2">{player.bio}</p>
                                    )}
                                    {player.platform_ids && Object.keys(player.platform_ids).length > 0 && (
                                        <div className="mt-3">
                                            <p className="text-sm text-gray-500 mb-1">Platform IDs:</p>
                                            <div className="flex gap-2 flex-wrap">
                                                {Object.entries(player.platform_ids).map(([platform, id]) => (
                                                    <span
                                                        key={platform}
                                                        className="px-2 py-1 bg-gray-800 rounded text-xs text-gray-300"
                                                    >
                                                        {platform}: {id}
                                                    </span>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ) : null}

                {/* Quick Stats */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg">Games</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-4xl font-bold text-blue-500">0</p>
                            <p className="text-gray-400 text-sm">Available games</p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg">Tournaments</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-4xl font-bold text-green-500">
                                {activeTournament ? 1 : 0}
                            </p>
                            <p className="text-gray-400 text-sm">Active tournaments</p>
                            {activeTournament && (
                                <p className="text-sm text-green-400 mt-2 truncate">
                                    {activeTournament.name}
                                </p>
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg">Your Rank</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-4xl font-bold text-yellow-500">-</p>
                            <p className="text-gray-400 text-sm">Global position</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Quick Actions */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    <Card className="hover:border-blue-500 transition-colors cursor-pointer">
                        <div onClick={() => setShowJoinModal(true)} className="block p-6">
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-blue-500/10 rounded-lg">
                                    <svg
                                        className="w-8 h-8 text-blue-500"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={2}
                                            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                                        />
                                    </svg>
                                </div>
                                <div>
                                    <h3 className="font-bold text-white">Join Tournament</h3>
                                    <p className="text-sm text-gray-400">
                                        Browse and join active tournaments
                                    </p>
                                </div>
                            </div>
                        </div>
                    </Card>

                    <Card className="hover:border-green-500 transition-colors cursor-pointer">
                        <Link to="/match-report" className="block p-6">
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-green-500/10 rounded-lg">
                                    <svg
                                        className="w-8 h-8 text-green-500"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={2}
                                            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                                        />
                                    </svg>
                                </div>
                                <div>
                                    <h3 className="font-bold text-white">Report Match</h3>
                                    <p className="text-sm text-gray-400">
                                        Submit your match results
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>

                    <Card className="hover:border-purple-500 transition-colors cursor-pointer">
                        <Link to="/profile" className="block p-6">
                            <div className="flex items-center gap-4">
                                <div className="p-3 bg-purple-500/10 rounded-lg">
                                    <svg
                                        className="w-8 h-8 text-purple-500"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={2}
                                            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                                        />
                                    </svg>
                                </div>
                                <div>
                                    <h3 className="font-bold text-white">View Stats</h3>
                                    <p className="text-sm text-gray-400">
                                        Check your rankings and stats
                                    </p>
                                </div>
                            </div>
                        </Link>
                    </Card>
                </div>
                <Card>
                    <CardHeader>
                        <CardTitle>Tournaments</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-gray-400 mb-4">
                            Browse and manage tournaments
                        </p>
                        <Link to="/tournaments">
                            <Button>View Tournaments</Button>
                        </Link>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Report Match</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-gray-400 mb-4">
                            Submit your match results and statistics
                        </p>
                        <Link to="/match-report">
                            <Button>Go to Match Report</Button>
                        </Link>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>View Leaderboard</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-gray-400 mb-4">
                            Check rankings across all tournaments
                        </p>
                        <Button disabled>Coming Soon</Button>
                    </CardContent>
                </Card>
            </div>

            {/* Recent Activity */}
            <Card>
                <CardHeader>
                    <CardTitle>Recent Activity</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="text-center py-8 text-gray-500">
                        <p>No recent activity</p>
                        <p className="text-sm mt-2">Start by joining a tournament!</p>
                    </div>
                </CardContent>
            </Card>

            {/* Join Tournament Modal */}
            <JoinTournamentModal
                isOpen={showJoinModal}
                onClose={() => setShowJoinModal(false)}
            />
        </div>
    );
};
