import { useMutation, useQueryClient } from '@tanstack/react-query';
import { isAxiosError } from 'axios';
import { useState } from 'react';
import { showError, showSuccess } from '../../lib/toast';
import { playerApi } from '../../services/player';
import type { ApiError, CreateProfileRequest, Player, UpdateProfileRequest } from '../../types/api';
import { Button, Input } from '../ui';

interface PlayerProfileFormProps {
    player?: Player;
    onSuccess?: () => void;
    onCancel?: () => void;
}

const PLATFORMS = ['PC', 'PlayStation', 'Xbox', 'Nintendo', 'Mobile', 'Crossplay'];
const LANGUAGES = ['English', 'Spanish', 'Portuguese', 'French', 'German', 'Italian', 'Japanese', 'Korean', 'Chinese'];
const REGIONS = ['NA', 'LATAM', 'EU', 'ASIA', 'OCE', 'MENA', 'Africa'];

const getErrorMessage = (error: unknown, fallback: string) => {
    if (isAxiosError<ApiError>(error)) {
        return error.response?.data?.error || error.message || fallback;
    }
    return fallback;
};

export const PlayerProfileForm = ({ player, onSuccess, onCancel }: PlayerProfileFormProps) => {
    const queryClient = useQueryClient();
    const isEditing = !!player;

    const [formData, setFormData] = useState({
        display_name: player?.display_name || '',
        preferred_platform: player?.preferred_platform || '',
        birth_year: player?.birth_year?.toString() || '',
        region: player?.region || '',
        language: player?.language || '',
        avatar_url: player?.avatar_url || '',
        bio: player?.bio || '',
    });

    const [platformIds, setPlatformIds] = useState<Record<string, string>>(
        player?.platform_ids || {}
    );
    const [newPlatformKey, setNewPlatformKey] = useState('');
    const [newPlatformValue, setNewPlatformValue] = useState('');

    const createMutation = useMutation({
        mutationFn: (data: CreateProfileRequest) => playerApi.createMyProfile(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['player'] });
            showSuccess('Profile created successfully!');
            onSuccess?.();
        },
        onError: (error: unknown) => {
            showError(getErrorMessage(error, 'Failed to create profile'));
        },
    });

    const updateMutation = useMutation({
        mutationFn: (data: UpdateProfileRequest) => playerApi.updateMyProfile(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['player'] });
            showSuccess('Profile updated successfully!');
            onSuccess?.();
        },
        onError: (error: unknown) => {
            showError(getErrorMessage(error, 'Failed to update profile'));
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        const requestData = {
            display_name: formData.display_name,
            preferred_platform: formData.preferred_platform,
            birth_year: formData.birth_year ? parseInt(formData.birth_year) : undefined,
            region: formData.region || undefined,
            language: formData.language || undefined,
            avatar_url: formData.avatar_url || undefined,
            bio: formData.bio || undefined,
            platform_ids: Object.keys(platformIds).length > 0 ? platformIds : undefined,
        };

        if (isEditing) {
            updateMutation.mutate(requestData);
        } else {
            createMutation.mutate(requestData as CreateProfileRequest);
        }
    };

    const handleAddPlatformId = () => {
        if (newPlatformKey && newPlatformValue) {
            setPlatformIds({ ...platformIds, [newPlatformKey]: newPlatformValue });
            setNewPlatformKey('');
            setNewPlatformValue('');
        }
    };

    const handleRemovePlatformId = (key: string) => {
        const updated = { ...platformIds };
        delete updated[key];
        setPlatformIds(updated);
    };

    const isLoading = createMutation.isPending || updateMutation.isPending;

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            {/* Required Fields */}
            <div className="space-y-4">
                <h3 className="text-sm font-medium text-gray-300">Required Information</h3>

                <Input
                    label="Display Name *"
                    value={formData.display_name}
                    onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                    placeholder="Your gamer name"
                    required
                    minLength={3}
                />

                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                        Preferred Platform *
                    </label>
                    <select
                        value={formData.preferred_platform}
                        onChange={(e) => setFormData({ ...formData, preferred_platform: e.target.value })}
                        className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                    >
                        <option value="">Select a platform</option>
                        {PLATFORMS.map((platform) => (
                            <option key={platform} value={platform}>
                                {platform}
                            </option>
                        ))}
                    </select>
                </div>
            </div>

            {/* Optional Fields */}
            <div className="space-y-4 pt-4 border-t border-gray-700">
                <h3 className="text-sm font-medium text-gray-300">Optional Information</h3>

                <Input
                    label="Birth Year"
                    type="number"
                    value={formData.birth_year}
                    onChange={(e) => setFormData({ ...formData, birth_year: e.target.value })}
                    placeholder="1990"
                    min={1900}
                    max={new Date().getFullYear()}
                />

                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">Region</label>
                    <select
                        value={formData.region}
                        onChange={(e) => setFormData({ ...formData, region: e.target.value })}
                        className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="">Select your region</option>
                        {REGIONS.map((region) => (
                            <option key={region} value={region}>
                                {region}
                            </option>
                        ))}
                    </select>
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">Language</label>
                    <select
                        value={formData.language}
                        onChange={(e) => setFormData({ ...formData, language: e.target.value })}
                        className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="">Select your language</option>
                        {LANGUAGES.map((lang) => (
                            <option key={lang} value={lang}>
                                {lang}
                            </option>
                        ))}
                    </select>
                </div>

                <Input
                    label="Avatar URL"
                    type="url"
                    value={formData.avatar_url}
                    onChange={(e) => setFormData({ ...formData, avatar_url: e.target.value })}
                    placeholder="https://example.com/avatar.jpg"
                />

                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">Bio</label>
                    <textarea
                        value={formData.bio}
                        onChange={(e) => setFormData({ ...formData, bio: e.target.value })}
                        placeholder="Tell us about yourself..."
                        className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        rows={3}
                    />
                </div>
            </div>

            {/* Platform IDs */}
            <div className="space-y-4 pt-4 border-t border-gray-700">
                <h3 className="text-sm font-medium text-gray-300">Platform IDs</h3>

                {Object.entries(platformIds).map(([key, value]) => (
                    <div key={key} className="flex items-center gap-2">
                        <span className="px-3 py-2 bg-gray-800 rounded-lg text-gray-300 flex-1">
                            {key}: {value}
                        </span>
                        <Button
                            type="button"
                            variant="secondary"
                            size="sm"
                            onClick={() => handleRemovePlatformId(key)}
                        >
                            Remove
                        </Button>
                    </div>
                ))}

                <div className="flex gap-2">
                    <Input
                        value={newPlatformKey}
                        onChange={(e) => setNewPlatformKey(e.target.value)}
                        placeholder="Platform (e.g., activision_id)"
                        className="flex-1"
                    />
                    <Input
                        value={newPlatformValue}
                        onChange={(e) => setNewPlatformValue(e.target.value)}
                        placeholder="ID"
                        className="flex-1"
                    />
                    <Button
                        type="button"
                        variant="secondary"
                        onClick={handleAddPlatformId}
                        disabled={!newPlatformKey || !newPlatformValue}
                    >
                        Add
                    </Button>
                </div>
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-4">
                <Button type="submit" className="flex-1" isLoading={isLoading}>
                    {isEditing ? 'Update Profile' : 'Create Profile'}
                </Button>
                {onCancel && (
                    <Button type="button" variant="secondary" onClick={onCancel} disabled={isLoading}>
                        Cancel
                    </Button>
                )}
            </div>
        </form>
    );
};
