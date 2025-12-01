'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

const schema = z.object({
    email: z.string().email('Please enter a valid email address'),
});

type FormData = z.infer<typeof schema>;

export default function ForgotPasswordPage() {
    const [loading, setLoading] = useState(false);
    const [submitted, setSubmitted] = useState(false);
    const [email, setEmail] = useState('');

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<FormData>({
        resolver: zodResolver(schema),
    });

    const onSubmit = async (data: FormData) => {
        setLoading(true);
        try {
            // TODO: Implement forgot password API call
            console.log('Forgot password for:', data.email);
            await new Promise((resolve) => setTimeout(resolve, 1000));
            setEmail(data.email);
            setSubmitted(true);
        } catch (error) {
            console.error('Forgot password error:', error);
        } finally {
            setLoading(false);
        }
    };

    if (submitted) {
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
                                    d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75"
                                />
                            </svg>
                        </div>
                        <h1 className="text-2xl font-semibold text-gray-900">Check your email</h1>
                        <p className="mt-2 text-sm text-gray-600">
                            We&apos;ve sent password reset instructions to <strong>{email}</strong>
                        </p>
                    </div>

                    <div className="bg-gray-50 rounded-lg p-4">
                        <h3 className="text-sm font-medium text-gray-900 mb-2">What to do next:</h3>
                        <ol className="text-sm text-gray-600 space-y-2 list-decimal list-inside">
                            <li>Check your email inbox</li>
                            <li>Click the reset password link</li>
                            <li>Enter your new password</li>
                            <li>Sign in with your new credentials</li>
                        </ol>
                    </div>

                    <div className="border-t border-gray-200 pt-6 space-y-3">
                        <p className="text-sm text-gray-600 text-center">
                            Didn&apos;t receive the email?{' '}
                            <button
                                onClick={() => setSubmitted(false)}
                                className="text-purple hover:text-purple-hover font-medium"
                            >
                                Try again
                            </button>
                        </p>
                        <a href="/login" className="btn-secondary w-full flex justify-center">
                            Back to Login
                        </a>
                    </div>
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
                                d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1 1 21.75 8.25Z"
                            />
                        </svg>
                    </div>
                    <h1 className="text-3xl font-semibold text-gray-900">Forgot password?</h1>
                    <p className="mt-2 text-sm text-gray-600">
                        No worries, we&apos;ll send you reset instructions.
                    </p>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                    <Input
                        label="Email address"
                        type="email"
                        autoComplete="email"
                        required
                        {...register('email')}
                        error={errors.email?.message}
                        helperText="Enter the email associated with your account"
                    />

                    <Button
                        type="submit"
                        variant="primary"
                        fullWidth
                        loading={loading}
                    >
                        Send reset link
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
