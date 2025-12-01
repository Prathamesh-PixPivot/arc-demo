'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { useAuth } from '@/lib/auth_context';

const loginSchema = z.object({
    email: z.string().email('Please enter a valid email address'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
    const [userType, setUserType] = useState<'user' | 'fiduciary'>('user');
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();
    const [error, setError] = useState<string | null>(null);

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<LoginFormData>({
        resolver: zodResolver(loginSchema),
    });

    const onSubmit = async (data: LoginFormData) => {
        setLoading(true);
        setError(null);
        try {
            await login(data.email, data.password, userType);
        } catch (err: any) {
            console.error('Login error:', err);
            setError(err.response?.data?.error || 'Invalid email or password');
        } finally {
            setLoading(false);
        }
    };

    const handleSSO = (provider: 'google' | 'microsoft') => {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        window.location.href = `${apiUrl}/auth/sso/${provider}?mode=login`;
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-purple-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="card-elevated max-w-md w-full p-8 space-y-6">
                {/* Header */}
                <div className="text-center">
                    <div className="inline-flex items-center justify-center w-16 h-16 bg-purple rounded-xl shadow-subtle mb-4">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2}
                            stroke="white"
                            className="w-10 h-10"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z"
                            />
                        </svg>
                    </div>
                    <h1 className="text-3xl font-semibold text-gray-900">Welcome back</h1>
                    <p className="mt-2 text-sm text-gray-600">
                        Sign in to your Arc Privacy account
                    </p>
                </div>

                {/* User Type Toggle */}
                <div className="flex rounded-lg bg-gray-100 p-1" role="tablist">
                    <button
                        type="button"
                        role="tab"
                        aria-selected={userType === 'user'}
                        onClick={() => setUserType('user')}
                        className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors focus-ring ${userType === 'user'
                            ? 'bg-white text-purple shadow-sm'
                            : 'text-gray-600 hover:text-gray-900'
                            }`}
                    >
                        Data Principal
                    </button>
                    <button
                        type="button"
                        role="tab"
                        aria-selected={userType === 'fiduciary'}
                        onClick={() => setUserType('fiduciary')}
                        className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors focus-ring ${userType === 'fiduciary'
                            ? 'bg-white text-purple shadow-sm'
                            : 'text-gray-600 hover:text-gray-900'
                            }`}
                    >
                        Fiduciary
                    </button>
                </div>

                {/* Error Message */}
                {error && (
                    <div className="bg-error-light text-error-dark px-4 py-3 rounded-lg text-sm">
                        {error}
                    </div>
                )}

                {/* Login Form */}
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                    <Input
                        label="Email address"
                        type="email"
                        autoComplete="email"
                        required
                        {...register('email')}
                        error={errors.email?.message}
                    />

                    <Input
                        label="Password"
                        type="password"
                        autoComplete="current-password"
                        required
                        {...register('password')}
                        error={errors.password?.message}
                    />

                    <div className="flex items-center justify-between">
                        <div className="flex items-center">
                            <input
                                id="remember-me"
                                name="remember-me"
                                type="checkbox"
                                className="h-4 w-4 text-purple focus:ring-purple border-gray-300 rounded"
                            />
                            <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-700">
                                Remember me
                            </label>
                        </div>

                        <div className="text-sm">
                            <a
                                href="/forgot-password"
                                className="font-medium text-purple hover:text-purple-hover focus-ring rounded"
                            >
                                Forgot password?
                            </a>
                        </div>
                    </div>

                    <Button
                        type="submit"
                        variant="primary"
                        fullWidth
                        loading={loading}
                    >
                        Sign in
                    </Button>
                </form>

                {/* Divider */}
                <div className="relative">
                    <div className="absolute inset-0 flex items-center">
                        <div className="w-full border-t border-gray-200" />
                    </div>
                    <div className="relative flex justify-center text-sm">
                        <span className="px-2 bg-white text-gray-500">Or continue with</span>
                    </div>
                </div>

                {/* SSO Buttons */}
                <div className="grid grid-cols-2 gap-3">
                    <Button
                        type="button"
                        variant="secondary"
                        onClick={() => handleSSO('google')}
                        className="flex items-center justify-center gap-2"
                    >
                        <svg className="w-5 h-5" viewBox="0 0 24 24">
                            <path
                                fill="#4285F4"
                                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                            />
                            <path
                                fill="#34A853"
                                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                            />
                            <path
                                fill="#FBBC05"
                                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                            />
                            <path
                                fill="#EA4335"
                                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                            />
                        </svg>
                        Google
                    </Button>
                    <Button
                        type="button"
                        variant="secondary"
                        onClick={() => handleSSO('microsoft')}
                        className="flex items-center justify-center gap-2"
                    >
                        <svg className="w-5 h-5" viewBox="0 0 23 23">
                            <path fill="#f3f3f3" d="M0 0h23v23H0z" />
                            <path fill="#f35325" d="M1 1h10v10H1z" />
                            <path fill="#81bc06" d="M12 1h10v10H12z" />
                            <path fill="#05a6f0" d="M1 12h10v10H1z" />
                            <path fill="#ffba08" d="M12 12h10v10H12z" />
                        </svg>
                        Microsoft
                    </Button>
                </div>

                {/* Sign up link */}
                <div className="text-center text-sm">
                    <span className="text-gray-600">Don&apos;t have an account? </span>
                    <a
                        href="/signup"
                        className="font-medium text-purple hover:text-purple-hover focus-ring rounded"
                    >
                        Sign up
                    </a>
                </div>
            </div>
        </div>
    );
}
