import { useState } from 'react';
import type { Player } from '../../types/api';
import { Button, Card, CardContent } from '../ui';
import { PlayerProfileForm } from './PlayerProfileForm';

interface OnboardingBannerProps {
    player: Player;
}

export const OnboardingBanner = ({ player }: OnboardingBannerProps) => {
    const [showForm, setShowForm] = useState(false);

    // Only show banner if player has default name
    if (player.display_name !== 'Player') {
        return null;
    }

    if (showForm) {
        return (
            <Card className="mb-6 border-blue-500/50 bg-blue-500/10">
                <CardContent className="py-6">
                    <h3 className="text-xl font-bold text-white mb-2">Complete Your Profile</h3>
                    <p className="text-gray-300 mb-4">
                        Set up your gamer profile to get started with TourneyRank
                    </p>
                    <PlayerProfileForm
                        player={player}
                        onSuccess={() => setShowForm(false)}
                        onCancel={() => setShowForm(false)}
                    />
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="mb-6 border-blue-500/50 bg-blue-500/10">
            <CardContent className="py-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h3 className="text-lg font-bold text-white mb-1">
                            Welcome to TourneyRank! 🎮
                        </h3>
                        <p className="text-gray-300">
                            Complete your profile to personalize your gaming experience
                        </p>
                    </div>
                    <Button onClick={() => setShowForm(true)}>
                        Complete Profile
                    </Button>
                </div>
            </CardContent>
        </Card>
    );
};
