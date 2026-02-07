import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Button, Card, CardContent, CardHeader, CardTitle } from '../components/ui';
import { playerApi } from '../services/player';
import { useAuthStore } from '../store/authStore';
import { useState } from 'react';

export const PlayerProfilePage = () => {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const [selectedGameId, setSelectedGameId] = useState<string | null>(null);

  // Fetch profile with stats
  const { data: profileData, isLoading: profileLoading } = useQuery({
    queryKey: ['player', 'profile-with-stats'],
    queryFn: playerApi.getMyProfileWithStats,
    enabled: !!user,
  });

  // Fetch specific game stats
  const { data: gameStats, isLoading: gameStatsLoading } = useQuery({
    queryKey: ['player', 'game-stats', selectedGameId],
    queryFn: () => (selectedGameId ? playerApi.getMyGameStats(selectedGameId) : Promise.resolve(null)),
    enabled: !!selectedGameId && !!user,
  });

  // Set first game as selected if not already set
  if (profileData?.games && profileData.games.length > 0 && !selectedGameId) {
    setSelectedGameId(profileData.games[0].game_id);
  }

  if (!user) {
    navigate('/dashboard');
    return null;
  }

  if (profileLoading) {
    return (
      <div className="min-h-screen bg-gray-900 p-6">
        <div className="max-w-6xl mx-auto">
          <div className="text-center text-gray-400">Loading your profile...</div>
        </div>
      </div>
    );
  }

  if (!profileData) {
    return (
      <div className="min-h-screen bg-gray-900 p-6">
        <div className="max-w-6xl mx-auto">
          <Card>
            <CardContent className="py-8">
              <div className="text-center">
                <p className="text-gray-400 mb-4">No profile data found</p>
                <Button onClick={() => navigate('/dashboard')}>
                  Back to Dashboard
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  const { player, games } = profileData;

  return (
    <div className="min-h-screen bg-gray-900 p-6">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <Button
            variant="secondary"
            onClick={() => navigate('/dashboard')}
            className="mb-4"
          >
            ← Back to Dashboard
          </Button>
          <div className="flex items-start gap-6">
            {player.avatar_url && (
              <img
                src={player.avatar_url}
                alt={player.display_name}
                className="w-24 h-24 rounded-lg object-cover border-2 border-blue-500"
              />
            )}
            <div className="flex-1">
              <h1 className="text-4xl font-bold text-white mb-2">
                {player.display_name}
              </h1>
              {player.bio && (
                <p className="text-gray-300 mb-4">{player.bio}</p>
              )}
              <p className="text-sm text-gray-500">
                Joined {new Date(player.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        </div>

        {/* Games Tabs */}
        {games && games.length > 0 ? (
          <>
            <div className="flex gap-2 mb-6 border-b border-gray-700 overflow-x-auto">
              {games.map((game) => (
                <button
                  key={game.game_id}
                  onClick={() => setSelectedGameId(game.game_id)}
                  className={`px-4 py-2 font-medium whitespace-nowrap transition-colors ${
                    selectedGameId === game.game_id
                      ? 'text-blue-500 border-b-2 border-blue-500'
                      : 'text-gray-400 hover:text-gray-300'
                  }`}
                >
                  {game.game_name}
                </button>
              ))}
            </div>

            {/* Game Stats Display */}
            {selectedGameId && gameStatsLoading ? (
              <div className="text-center text-gray-400">Loading stats...</div>
            ) : selectedGameId && gameStats ? (
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Main Stats Card */}
                <div className="lg:col-span-2">
                  <Card>
                    <CardHeader>
                      <CardTitle>
                        {gameStats.game_name || 'Game Stats'}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-6">
                        {/* Tier and Ranking Score */}
                        <div>
                          <div className="flex items-center justify-between mb-4">
                            <h3 className="text-sm font-medium text-gray-400">
                              Ranking Score
                            </h3>
                            <span
                              className={`px-3 py-1 rounded-full text-sm font-semibold ${
                                getTierColor(gameStats.tier)
                              }`}
                            >
                              {gameStats.tier.charAt(0).toUpperCase() +
                                gameStats.tier.slice(1)}
                            </span>
                          </div>
                          <div className="mb-2">
                            <div className="text-2xl font-bold text-blue-400">
                              {gameStats.ranking_score.toFixed(0)}
                            </div>
                            <p className="text-xs text-gray-500">/1000</p>
                          </div>
                          {/* Progress bar */}
                          <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
                            <div
                              className="bg-gradient-to-r from-blue-500 to-blue-400 h-full rounded-full transition-all"
                              style={{
                                width: `${Math.min(
                                  (gameStats.ranking_score / 1000) * 100,
                                  100,
                                )}%`,
                              }}
                            />
                          </div>
                        </div>

                        {/* Matches Played */}
                        <div className="border-t border-gray-700 pt-4">
                          <p className="text-sm text-gray-400 mb-2">
                            Matches Played
                          </p>
                          <p className="text-2xl font-bold text-white">
                            {gameStats.matches_played}
                          </p>
                        </div>

                        {/* Last Match */}
                        {gameStats.last_match_at && (
                          <div className="border-t border-gray-700 pt-4">
                            <p className="text-sm text-gray-400 mb-2">
                              Last Match
                            </p>
                            <p className="text-white">
                              {new Date(
                                gameStats.last_match_at,
                              ).toLocaleDateString()}{' '}
                              {new Date(
                                gameStats.last_match_at,
                              ).toLocaleTimeString()}
                            </p>
                          </div>
                        )}

                        {/* Custom Stats Grid */}
                        {Object.keys(gameStats.stats || {}).length > 0 && (
                          <div className="border-t border-gray-700 pt-4">
                            <p className="text-sm font-medium text-gray-400 mb-4">
                              Statistics
                            </p>
                            <div className="grid grid-cols-2 gap-4">
                              {Object.entries(gameStats.stats).map(
                                ([key, value]) => (
                                  <div
                                    key={key}
                                    className="bg-gray-800 rounded p-3"
                                  >
                                    <p className="text-xs text-gray-400 uppercase">
                                      {key.replace(/_/g, ' ')}
                                    </p>
                                    <p className="text-lg font-bold text-white">
                                      {typeof value === 'number'
                                        ? value.toFixed(2)
                                        : value}
                                    </p>
                                  </div>
                                ),
                              )}
                            </div>
                          </div>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Leaderboard Position Card */}
                <Card>
                  <CardHeader>
                    <CardTitle>Leaderboard Position</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div>
                        <p className="text-sm text-gray-400 mb-1">Rank</p>
                        <p className="text-3xl font-bold text-blue-400">
                          #{gameStats.rank}
                        </p>
                      </div>
                      <div className="border-t border-gray-700 pt-4">
                        <p className="text-sm text-gray-400 mb-1">Percentile</p>
                        <p className="text-2xl font-bold text-white">
                          {(gameStats.percentile * 100).toFixed(1)}%
                        </p>
                        <p className="text-xs text-gray-500">
                          Top{' '}
                          {(gameStats.percentile * 100).toFixed(1)}% of
                          players
                        </p>
                      </div>
                      <div className="border-t border-gray-700 pt-4">
                        <p className="text-sm text-gray-400">
                          Keep playing to climb the leaderboard! 📈
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            ) : null}
          </>
        ) : (
          <Card>
            <CardContent className="py-8">
              <div className="text-center">
                <p className="text-gray-400 mb-4">
                  You haven't played any games yet. Start playing to see your
                  stats!
                </p>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
};

// Helper function to get tier color classes
function getTierColor(tier: string): string {
  switch (tier.toLowerCase()) {
    case 'elite':
      return 'bg-yellow-500/20 text-yellow-400';
    case 'advanced':
      return 'bg-purple-500/20 text-purple-400';
    case 'intermediate':
      return 'bg-blue-500/20 text-blue-400';
    case 'beginner':
      return 'bg-gray-500/20 text-gray-400';
    default:
      return 'bg-gray-500/20 text-gray-400';
  }
}
