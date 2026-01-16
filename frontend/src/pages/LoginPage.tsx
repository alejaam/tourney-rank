import { AxiosError } from 'axios';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Button, Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle, Input } from '../components/ui';
import { useLogin } from '../features/auth/hooks';
import type { ApiError } from '../types/api';

export const LoginPage = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const login = useLogin();

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        login.mutate({ email, password });
    };

    const errorMessage = login.error instanceof AxiosError
        ? (login.error.response?.data as ApiError)?.error || 'Login failed'
        : null;

    return (
        <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
            <Card className="w-full max-w-md">
                <CardHeader>
                    <CardTitle>Welcome Back</CardTitle>
                    <CardDescription>Sign in to your TourneyRank account</CardDescription>
                </CardHeader>

                <form onSubmit={handleSubmit}>
                    <CardContent className="space-y-4">
                        {errorMessage && (
                            <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-2 rounded-lg text-sm">
                                {errorMessage}
                            </div>
                        )}

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
                            placeholder="••••••••"
                            value={password}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
                            required
                        />
                    </CardContent>

                    <CardFooter className="flex flex-col gap-4">
                        <Button
                            type="submit"
                            className="w-full"
                            isLoading={login.isPending}
                        >
                            Sign In
                        </Button>

                        <p className="text-center text-gray-400 text-sm">
                            Don't have an account?{' '}
                            <Link to="/register" className="text-blue-500 hover:text-blue-400 font-medium">
                                Sign up
                            </Link>
                        </p>
                    </CardFooter>
                </form>
            </Card>
        </div>
    );
};
