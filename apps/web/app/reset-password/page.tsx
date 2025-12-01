'use client';

import { useState, useEffect, Suspense } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { PasswordStrength } from '@/components/ui/PasswordStrength';
import { useSearchParams } from 'next/navigation';

const schema = z.object({
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
});

type FormData = z.infer<typeof schema>;

function ResetPasswordContent() {
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);
    const [tokenValid, setTokenValid] = useState<boolean | null>(null);
    const searchParams = useSearchParams();
    const token = searchParams?.get('token');

    const {
        register,
        handleSubmit,
        watch,
        formState: { errors },
    } = useForm<FormData>({
        resolver: zodResolver(schema),
    });

    useEffect(() => {
        // Validate token
        if (token) {
            // TODO: Implement token validation API call
            setTokenValid(true);
        } else {
            setTokenValid(false);
        }
    }, [token]);

    const onSubmit = async (data: FormData) => {
        setLoading(true);
        try {
            // TODO: Implement reset password API call
            console.log('Reset password with token:', token, data.password);
            await new Promise((resolve) => setTimeout(resolve, 1000));
            setSuccess(true);
        } catch (error) {
            console.error('Reset password error:', error);
        } finally {
            setLoading(false);
        }
    };

    if (tokenValid === null) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-purple-50">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-purple"></div>
                    <p className="mt-4 text-gray-600">Validating reset link...</p>
                </div>
            </div>
        );
    }

    if (tokenValid === false) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-purple-50 py-12 px-4 sm:px-6 lg:px-8">
                <div className="card-elevated max-w-md w-full p-8 space-y-6">
                    <div className="text-center">
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-error-light rounded-full mb-4">
                            <svg
                                className="w-10 h-10 text-error"
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                strokeWidth={2}
                                stroke="currentColor"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
                                />
                            </svg>
                        </div>
                        <h1 className="text-2xl font-semibold text-gray-900">Invalid or expired link</h1>
                        <p className="mt-2 text-sm text-gray-600">
                            This password reset link is invalid or has expired.
                        </p>
                    </div>

                    <a href="/forgot-password" className="btn-primary w-full flex justify-center">
                        Request new reset link
                    </a>
                </div>
            </div>
        );
    }

    if (success) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-purple-50 py-12 px-4 sm:px-6 lg:px-8">
                <div className="card-elevated max-w-md w-full p-8 space-y-6">
                    <div className="text-center">
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-success-light rounded-full mb-4">
                            <svg
                                className="w-10 h-10 text-success"
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                strokeWidth={2}
                                stroke="currentColor"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
                                />
                            </svg>
                        </div>
                        <h1 className="text-2xl font-semibold text-gray-900">Password reset successful!</h1>
                        <p className="mt-2 text-sm text-gray-600">
                            Your password has been successfully reset. You can now sign in with your new password.
                        </p>
                    </div>

                    <a href="/login" className="btn-primary w-full flex justify-center">
                        Go to Login
                    </a>
                </div>
            </div>
        );
    }

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
                                d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z"
                            />
                        </svg>
                    </div>
                    <h1 className="text-3xl font-semibold text-gray-900">Set new password</h1>
                    <p className="mt-2 text-sm text-gray-600">
                        Choose a strong password for your account.
                    </p>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                    <Input
                        label="New password"
                        type="password"
                        autoComplete="new-password"
                        required
                        {...register('password')}
                        error={errors.password?.message}
                    />

                    <PasswordStrength password={watch('password') || ''} />

                    <Input
                        label="Confirm new password"
                        type="password"
                        autoComplete="new-password"
                        required
                        {...register('confirmPassword')}
                        error={errors.confirmPassword?.message}
                    />

                    <Button
                        type="submit"
                        variant="primary"
                        fullWidth
                        loading={loading}
                    >
                        Reset password
                    </Button>
                </form>

                {/* Back to login */}
                <div className="text-center text-sm border-t border-gray-200 pt-6">
                    <a
                        href="/login"
                        className="inline-flex items-center text-purple hover:text-purple-hover font-medium focus-ring rounded"
                    >
                        <svg
                            className="w-4 h-4 mr-2"
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2}
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"
                            />
                        </svg>
                        Back to login
                    </a>
                </div>
            </div>
        </div>
    );
}

export default function ResetPasswordPage() {
    return (
        <Suspense fallback={
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-purple-50">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-purple"></div>
                    <p className="mt-4 text-gray-600">Loading...</p>
                </div>
            </div>
        }>
            <ResetPasswordContent />
        </Suspense>
    );
}
