import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { OnboardingBanner } from '../components/player';
import { Button, Card, CardContent, CardHeader, CardTitle } from '../components/ui';
import { useLogout } from '../features/auth/hooks';
import { playerApi } from '../services/player';
import type { AuthState } from '../store/authStore';
import { useAuthStore } from '../store/authStore';

export const DashboardPage = () => {
    const user = useAuthStore((state: AuthState) => state.user);
    const logout = useLogout();

    // Fetch player profile
    const { data: player, isLoading, error } = useQuery({
        queryKey: ['player', 'me'],
        queryFn: playerApi.getMyProfile,
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
                            <Link to="/admin">
                                <Button variant="secondary">Admin Panel</Button>
                            </Link>
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
                            <p className="text-4xl font-bold text-green-500">0</p>
                            <p className="text-gray-400 text-sm">Active tournaments</p>
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
            </div>
        </div>
    );
};
