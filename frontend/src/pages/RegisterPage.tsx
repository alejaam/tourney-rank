import { AxiosError } from 'axios';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Button, Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle, Input } from '../components/ui';
import { useRegister } from '../features/auth/hooks';
import type { ApiError } from '../types/api';

export const RegisterPage = () => {
    const [username, setUsername] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [validationError, setValidationError] = useState('');
    const register = useRegister();

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        setValidationError('');

        if (password !== confirmPassword) {
            setValidationError('Passwords do not match');
            return;
        }

        if (password.length < 8) {
            setValidationError('Password must be at least 8 characters');
            return;
        }

        register.mutate({ username, email, password });
    };

    const errorMessage = register.error instanceof AxiosError
        ? (register.error.response?.data as ApiError)?.error || 'Registration failed'
        : validationError || null;

    return (
        <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
            <Card className="w-full max-w-md">
                <CardHeader>
                    <CardTitle>Create Account</CardTitle>
                    <CardDescription>Join TourneyRank and start competing</CardDescription>
                </CardHeader>

                <form onSubmit={handleSubmit}>
                    <CardContent className="space-y-4">
                        {errorMessage && (
                            <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-2 rounded-lg text-sm">
                                {errorMessage}
                            </div>
                        )}

                        <Input
                            label="Username"
                            type="text"
                            name="username"
                            placeholder="Choose a username"
                            value={username}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)}
                            required
                        />

                        <Input
                            label="Email"
                            type="email"
                            name="email"
                            placeholder="you@example.com"
                            value={email}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
                            required
                        />

                        <Input
                            label="Password"
                            type="password"
                            name="password"
                            placeholder="At least 8 characters"
                            value={password}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
                            required
                        />

                        <Input
                            label="Confirm Password"
                            type="password"
                            name="confirmPassword"
                            placeholder="Repeat your password"
                            value={confirmPassword}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setConfirmPassword(e.target.value)}
                            required
                        />
                    </CardContent>

                    <CardFooter className="flex flex-col gap-4">
                        <Button
                            type="submit"
                            className="w-full"
                            isLoading={register.isPending}
                        >
                            Create Account
                        </Button>

                        <p className="text-center text-gray-400 text-sm">
                            Already have an account?{' '}
                            <Link to="/login" className="text-blue-500 hover:text-blue-400 font-medium">
                                Sign in
                            </Link>
                        </p>
                    </CardFooter>
                </form>
            </Card>
        </div>
    );
};
