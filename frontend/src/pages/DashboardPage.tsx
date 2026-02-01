import { Link } from 'react-router-dom';
import { Button, Card, CardContent, CardHeader, CardTitle } from '../components/ui';
import { useLogout } from '../features/auth/hooks';
import type { AuthState } from '../store/authStore';
import { useAuthStore } from '../store/authStore';

export const DashboardPage = () => {
    const user = useAuthStore((state: AuthState) => state.user);
    const logout = useLogout();

    return (
        <div className="min-h-screen bg-gray-900 p-6">
            <div className="max-w-4xl mx-auto">
                {/* Header */}
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h1 className="text-3xl font-bold text-white">Dashboard</h1>
                        <p className="text-gray-400">Welcome back, {user?.username}!</p>
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
