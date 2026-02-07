import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, CardContent, CardHeader, CardTitle, Input } from '../components/ui';
import { adminApi } from '../services/admin';
import { useAuthStore } from '../store/authStore';
import type { Game, Player, User } from '../types/api';

export const AdminPage = () => {
    const user = useAuthStore((state) => state.user);
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = useState<'users' | 'games' | 'players'>('users');

    // Redirect if not admin
    useEffect(() => {
        if (!user || user.role !== 'admin') {
            navigate('/dashboard');
        }
    }, [user, navigate]);

    if (!user || user.role !== 'admin') {
        return null;
    }

    return (
        <div className="min-h-screen bg-gray-900 p-6">
            <div className="max-w-7xl mx-auto">
                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-3xl font-bold text-white mb-2">Admin Panel</h1>
                    <p className="text-gray-400">Manage users, games, and players</p>
                </div>

                {/* Tabs */}
                <div className="flex gap-2 mb-6 border-b border-gray-700">
                    <button
                        onClick={() => setActiveTab('users')}
                        className={`px-4 py-2 font-medium transition-colors ${activeTab === 'users'
                            ? 'text-blue-500 border-b-2 border-blue-500'
                            : 'text-gray-400 hover:text-gray-300'
                            }`}
                    >
                        Users
                    </button>
                    <button
                        onClick={() => setActiveTab('games')}
                        className={`px-4 py-2 font-medium transition-colors ${activeTab === 'games'
                            ? 'text-blue-500 border-b-2 border-blue-500'
                            : 'text-gray-400 hover:text-gray-300'
                            }`}
                    >
                        Games
                    </button>
                    <button
                        onClick={() => setActiveTab('players')}
                        className={`px-4 py-2 font-medium transition-colors ${activeTab === 'players'
                            ? 'text-blue-500 border-b-2 border-blue-500'
                            : 'text-gray-400 hover:text-gray-300'
                            }`}
                    >
                        Players
                    </button>
                </div>

                {/* Content */}
                {activeTab === 'users' && <UserManagement />}
                {activeTab === 'games' && <GameManagement />}
                {activeTab === 'players' && <PlayerManagement />}
            </div>
        </div>
    );
};

// User Management Component
const UserManagement = () => {
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadUsers();
    }, []);

    const loadUsers = async () => {
        try {
            setLoading(true);
            const data = await adminApi.users.list();
            setUsers(data.users || []);
        } catch (error) {
            console.error('Failed to load users:', error);
            setUsers([]);
        } finally {
            setLoading(false);
        }
    };

    const handleCopyUserId = (id: string) => {
        navigator.clipboard.writeText(id);
        alert('User ID copied to clipboard!');
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this user?')) return;

        try {
            await adminApi.users.delete(id);
            setUsers(users.filter((u) => u.id !== id));
        } catch (error) {
            console.error('Failed to delete user:', error);
            alert('Failed to delete user');
        }
    };

    const handleToggleRole = async (id: string, currentRole: string) => {
        const newRole = currentRole === 'admin' ? 'user' : 'admin';

        try {
            await adminApi.users.updateRole(id, { role: newRole });
            setUsers(users.map((u) => (u.id === id ? { ...u, role: newRole } : u)));
        } catch (error) {
            console.error('Failed to update role:', error);
            alert('Failed to update role');
        }
    };

    if (loading) {
        return <div className="text-gray-400">Loading users...</div>;
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>User Management</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead>
                            <tr className="border-b border-gray-700">
                                <th className="text-left py-3 px-4 text-gray-400 font-medium">Username</th>
                                <th className="text-left py-3 px-4 text-gray-400 font-medium">Email</th>
                                <th className="text-left py-3 px-4 text-gray-400 font-medium">User ID</th>
                                <th className="text-left py-3 px-4 text-gray-400 font-medium">Role</th>
                                <th className="text-left py-3 px-4 text-gray-400 font-medium">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {users.map((user) => (
                                <tr key={user.id} className="border-b border-gray-700">
                                    <td className="py-3 px-4 text-white">{user.username}</td>
                                    <td className="py-3 px-4 text-gray-400">{user.email}</td>
                                    <td className="py-3 px-4">
                                        <button
                                            onClick={() => handleCopyUserId(user.id)}
                                            className="text-xs text-gray-500 bg-gray-800 px-2 py-1 rounded hover:bg-gray-700 hover:text-gray-400 transition-colors cursor-pointer"
                                            title="Click to copy full ID"
                                        >
                                            {user.id.slice(0, 8)}...
                                        </button>
                                    </td>
                                    <td className="py-3 px-4">
                                        <span
                                            className={`px-2 py-1 rounded text-xs font-medium ${user.role === 'admin'
                                                ? 'bg-blue-500/20 text-blue-400'
                                                : 'bg-gray-500/20 text-gray-400'
                                                }`}
                                        >
                                            {user.role}
                                        </span>
                                    </td>
                                    <td className="py-3 px-4">
                                        <div className="flex gap-2">
                                            <Button
                                                size="sm"
                                                variant="secondary"
                                                onClick={() => handleToggleRole(user.id, user.role)}
                                            >
                                                Toggle Role
                                            </Button>
                                            <Button
                                                size="sm"
                                                variant="danger"
                                                onClick={() => handleDelete(user.id)}
                                            >
                                                Delete
                                            </Button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </CardContent>
        </Card>
    );
};

// Game Management Component
const GameManagement = () => {
    const [games, setGames] = useState<Game[]>([]);
    const [loading, setLoading] = useState(true);
    const [showForm, setShowForm] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        slug: '',
        description: '',
        is_active: true,
        platform_id_format: 'username',
        stat_schema: {
            kills: { type: 'integer', min: 0, label: 'Kills' },
            deaths: { type: 'integer', min: 0, label: 'Deaths' },
            wins: { type: 'integer', min: 0, label: 'Wins' },
        },
        ranking_weights: {
            kills: 0.4,
            deaths: 0.3,
            wins: 0.3,
        },
    });

    useEffect(() => {
        loadGames();
    }, []);

    const loadGames = async () => {
        try {
            setLoading(true);
            const data = await adminApi.games.list();
            setGames(data.games || []);
        } catch (error) {
            console.error('Failed to load games:', error);
            setGames([]);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this game?')) return;

        try {
            await adminApi.games.delete(id);
            setGames(games.filter((g) => g.id !== id));
        } catch (error) {
            console.error('Failed to delete game:', error);
            alert('Failed to delete game');
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        try {
            await adminApi.games.create(formData);
            setFormData({
                name: '',
                slug: '',
                description: '',
                is_active: true,
                platform_id_format: 'username',
                stat_schema: {
                    kills: { type: 'integer', min: 0, label: 'Kills' },
                    deaths: { type: 'integer', min: 0, label: 'Deaths' },
                    wins: { type: 'integer', min: 0, label: 'Wins' },
                },
                ranking_weights: {
                    kills: 0.4,
                    deaths: 0.3,
                    wins: 0.3,
                },
            });
            setShowForm(false);
            loadGames();
        } catch (error) {
            console.error('Failed to create game:', error);
            alert('Failed to create game');
        }
    };

    if (loading) {
        return <div className="text-gray-400">Loading games...</div>;
    }

    return (
        <Card>
            <CardHeader>
                <div className="flex justify-between items-center">
                    <CardTitle>Game Management</CardTitle>
                    <Button onClick={() => setShowForm(!showForm)}>
                        {showForm ? 'Cancel' : '+ Add Game'}
                    </Button>
                </div>
            </CardHeader>
            <CardContent>
                {showForm && (
                    <form onSubmit={handleSubmit} className="mb-6 p-4 border border-gray-700 rounded-lg">
                        <div className="grid gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-400 mb-1">
                                    Name
                                </label>
                                <Input
                                    required
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                    placeholder="League of Legends"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-400 mb-1">
                                    Slug
                                </label>
                                <Input
                                    required
                                    value={formData.slug}
                                    onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                                    placeholder="league-of-legends"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-400 mb-1">
                                    Description
                                </label>
                                <textarea
                                    required
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    rows={3}
                                    placeholder="Multiplayer online battle arena game"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-400 mb-1">
                                    Platform ID Format
                                </label>
                                <Input
                                    required
                                    value={formData.platform_id_format}
                                    onChange={(e) => setFormData({ ...formData, platform_id_format: e.target.value })}
                                    placeholder="username"
                                />
                                <p className="text-xs text-gray-500 mt-1">Format for player IDs (e.g., username, riot_id)</p>
                            </div>
                            <div className="flex items-center gap-2">
                                <input
                                    type="checkbox"
                                    checked={formData.is_active}
                                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                                    className="w-4 h-4"
                                />
                                <label className="text-sm text-gray-400">Active</label>
                            </div>
                            <div className="bg-gray-800/50 p-3 rounded border border-gray-700">
                                <p className="text-xs text-gray-400 mb-1">
                                    <strong>Default Stats:</strong> kills (40%), deaths (30%), wins (30%)
                                </p>
                                <p className="text-xs text-gray-500">
                                    You can customize stat schema and ranking weights later via API
                                </p>
                            </div>
                        </div>
                        <Button type="submit" className="mt-4">Create Game</Button>
                    </form>
                )}

                <div className="grid gap-4">
                    {games && games.length > 0 ? games.map((game) => (
                        <div key={game.id} className="border border-gray-700 rounded-lg p-4">
                            <div className="flex justify-between items-start">
                                <div>
                                    <h3 className="text-white font-medium text-lg">{game.name}</h3>
                                    <p className="text-gray-400 text-sm">{game.slug}</p>
                                    <p className="text-gray-500 text-sm mt-2">{game.description}</p>
                                    <span
                                        className={`inline-block mt-2 px-2 py-1 rounded text-xs ${game.is_active
                                            ? 'bg-green-500/20 text-green-400'
                                            : 'bg-red-500/20 text-red-400'
                                            }`}
                                    >
                                        {game.is_active ? 'Active' : 'Inactive'}
                                    </span>
                                </div>
                                <Button size="sm" variant="danger" onClick={() => handleDelete(game.id)}>
                                    Delete
                                </Button>
                            </div>
                        </div>
                    )) : (
                        <div className="text-center py-8 text-gray-500">
                            <p>No games yet</p>
                            <p className="text-sm mt-2">Click "Add Game" to create one</p>
                        </div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
};

// Player Management Component
const PlayerManagement = () => {
    const [players, setPlayers] = useState<Player[]>([]);
    const [loading, setLoading] = useState(true);
    const [actionPlayerId, setActionPlayerId] = useState<string | null>(null);

    useEffect(() => {
        const loadPlayers = async () => {
            try {
                setLoading(true);
                const data = await adminApi.players.list();
                setPlayers(data.players || []);
            } catch (error) {
                console.error('Failed to load players:', error);
                setPlayers([]);
            } finally {
                setLoading(false);
            }
        };

        loadPlayers();
    }, []);

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this player?')) return;

        try {
            await adminApi.players.delete(id);
            setPlayers(players.filter((p) => p.id !== id));
        } catch (error) {
            console.error('Failed to delete player:', error);
            alert('Failed to delete player');
        }
    };

    const handleToggleBan = async (player: Player) => {
        const nextAction = player.is_banned ? 'unban' : 'ban';
        const confirmMsg = player.is_banned
            ? 'Unban this player?'
            : 'Ban this player?';

        if (!confirm(confirmMsg)) return;

        try {
            setActionPlayerId(player.id);
            const updated = nextAction === 'ban'
                ? await adminApi.players.ban(player.id)
                : await adminApi.players.unban(player.id);

            setPlayers(players.map((p) => (p.id === player.id ? updated : p)));
        } catch (error) {
            console.error(`Failed to ${nextAction} player:`, error);
            alert(`Failed to ${nextAction} player`);
        } finally {
            setActionPlayerId(null);
        }
    };

    if (loading) {
        return <div className="text-gray-400">Loading players...</div>;
    }

    return (
        <Card>
            <CardHeader>
                <div className="flex justify-between items-center">
                    <CardTitle>Player Management</CardTitle>
                </div>
            </CardHeader>
            <CardContent>
                {players.length === 0 ? (
                    <div className="text-center py-8 text-gray-500">
                        <p>No players yet</p>
                        <p className="text-sm mt-2">Player profiles are created automatically</p>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full">
                            <thead>
                                <tr className="border-b border-gray-700">
                                    <th className="text-left py-3 px-4 text-gray-400 font-medium">Display Name</th>
                                    <th className="text-left py-3 px-4 text-gray-400 font-medium">Status</th>
                                    <th className="text-left py-3 px-4 text-gray-400 font-medium">Bio</th>
                                    <th className="text-left py-3 px-4 text-gray-400 font-medium">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {players.map((player) => (
                                    <tr key={player.id} className="border-b border-gray-700">
                                        <td className="py-3 px-4 text-white">{player.display_name}</td>
                                        <td className="py-3 px-4">
                                            <span
                                                className={`px-2 py-1 rounded text-xs font-medium ${player.is_banned
                                                    ? 'bg-red-500/20 text-red-400'
                                                    : 'bg-green-500/20 text-green-400'
                                                    }`}
                                            >
                                                {player.is_banned ? 'Banned' : 'Active'}
                                            </span>
                                        </td>
                                        <td className="py-3 px-4 text-gray-400">{player.bio || '-'}</td>
                                        <td className="py-3 px-4">
                                            <div className="flex gap-2">
                                                <Button
                                                    size="sm"
                                                    variant={player.is_banned ? 'secondary' : 'danger'}
                                                    isLoading={actionPlayerId === player.id}
                                                    onClick={() => handleToggleBan(player)}
                                                >
                                                    {player.is_banned ? 'Unban' : 'Ban'}
                                                </Button>
                                                <Button
                                                    size="sm"
                                                    variant="danger"
                                                    onClick={() => handleDelete(player.id)}
                                                    disabled={actionPlayerId === player.id}
                                                >
                                                    Delete
                                                </Button>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};
