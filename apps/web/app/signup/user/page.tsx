'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { StepIndicator } from '@/components/ui/StepIndicator';
import { PasswordStrength } from '@/components/ui/PasswordStrength';
import { api } from '@/lib/api';

const steps = [
    { id: 1, name: 'Account', description: 'Basic information' },
    { id: 2, name: 'Guardian', description: 'Parent/Guardian' },
    { id: 3, name: 'Verify', description: 'Email verification' },
];

const step1Schema = z.object({
    firstName: z.string().min(1, 'First name is required'),
    lastName: z.string().min(1, 'Last name is required'),
    email: z.string().email('Please enter a valid email address'),
    phone: z.string().optional(),
    age: z.coerce.number().min(1, 'Age is required').max(150, 'Invalid age'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
});

const step2Schema = z.object({
    guardianEmail: z.string().email('Please enter a valid email address'),
    guardianName: z.string().min(1, 'Guardian name is required'),
    relationship: z.string().min(1, 'Relationship is required'),
});

type Step1Data = z.infer<typeof step1Schema>;
type Step2Data = z.infer<typeof step2Schema>;

export default function UserSignupPage() {
    const [currentStep, setCurrentStep] = useState(1);
    const [loading, setLoading] = useState(false);
    const [step1Data, setStep1Data] = useState<Step1Data | null>(null);
    const [isMinor, setIsMinor] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const step1Form = useForm<Step1Data>({
        resolver: zodResolver(step1Schema),
    });

    const step2Form = useForm<Step2Data>({
        resolver: zodResolver(step2Schema),
    });

    const handleStep1 = async (data: Step1Data) => {
        setStep1Data(data);
        const minor = data.age < 18;
        setIsMinor(minor);

        if (minor) {
            setCurrentStep(2); // Go to guardian step
        } else {
            // For adults, submit directly
            await submitSignup(data);
        }
    };

    const handleStep2 = async (data: Step2Data) => {
        if (!step1Data) return;
        await submitSignup({ ...step1Data, ...data });
    };

    const submitSignup = async (data: any) => {
        setLoading(true);
        setError(null);
        try {
            const payload = {
                email: data.email,
                password: data.password,
                firstName: data.firstName,
                lastName: data.lastName,
                age: Number(data.age),
                phone: data.phone || "",
                guardianEmail: data.guardianEmail,
                // Location is optional and not in the form currently, can be added later
            };

            await api.post('/auth/user/signup', payload);

            // Only advance to success step if API call succeeds
            setCurrentStep(3);
        } catch (error: any) {
            console.error('Signup error:', error);
            setError(error.response?.data?.message || 'Signup failed. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-purple-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="card-elevated max-w-2xl w-full p-8 space-y-8">
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
                                d="M18 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0ZM3 19.235v-.11a6.375 6.375 0 0 1 12.75 0v.109A12.318 12.318 0 0 1 9.374 21c-2.331 0-4.512-.645-6.374-1.766Z"
                            />
                        </svg>
                    </div>
                    <h1 className="text-3xl font-semibold text-gray-900">Create your account</h1>
                    <p className="mt-2 text-sm text-gray-600">
                        Join Arc Privacy Platform as a Data Principal
                    </p>
                </div>

                {error && (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
                        <span className="block sm:inline">{error}</span>
                    </div>
                )}

                {/* Step Indicator */}
                <StepIndicator
                    steps={isMinor ? steps : steps.filter(s => s.id !== 2)}
                    currentStep={currentStep}
                />

                {/* Step 1: Account Information */}
                {currentStep === 1 && (
                    <>
                        <form onSubmit={step1Form.handleSubmit(handleStep1)} className="space-y-4">
                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                <Input
                                    label="First Name"
                                    type="text"
                                    required
                                    {...step1Form.register('firstName')}
                                    error={step1Form.formState.errors.firstName?.message}
                                />
                                <Input
                                    label="Last Name"
                                    type="text"
                                    required
                                    {...step1Form.register('lastName')}
                                    error={step1Form.formState.errors.lastName?.message}
                                />
                            </div>

                            <Input
                                label="Email Address"
                                type="email"
                                autoComplete="email"
                                required
                                {...step1Form.register('email')}
                                error={step1Form.formState.errors.email?.message}
                            />

                            <Input
                                label="Phone Number"
                                type="tel"
                                autoComplete="tel"
                                helperText="Optional"
                                {...step1Form.register('phone')}
                                error={step1Form.formState.errors.phone?.message}
                            />

                            <Input
                                label="Age"
                                type="number"
                                required
                                {...step1Form.register('age')}
                                error={step1Form.formState.errors.age?.message}
                                helperText="Required for age verification and consent compliance"
                            />

                            <Input
                                label="Password"
                                type="password"
                                autoComplete="new-password"
                                required
                                {...step1Form.register('password')}
                                error={step1Form.formState.errors.password?.message}
                            />

                            <PasswordStrength password={step1Form.watch('password') || ''} />

                            <Input
                                label="Confirm Password"
                                type="password"
                                autoComplete="new-password"
                                required
                                {...step1Form.register('confirmPassword')}
                                error={step1Form.formState.errors.confirmPassword?.message}
                            />

                            <div className="flex justify-between pt-4">
                                <a href="/login" className="btn-tertiary">
                                    Back to Login
                                </a>
                                <Button type="submit" variant="primary">
                                    Continue
                                </Button>
                            </div>
                        </form>

                        <div className="relative">
                            <div className="absolute inset-0 flex items-center">
                                <div className="w-full border-t border-gray-200" />
                            </div>
                            <div className="relative flex justify-center text-sm">
                                <span className="px-2 bg-white text-gray-500">Or sign up with</span>
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                            <Button
                                type="button"
                                variant="secondary"
                                onClick={() => {
                                    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
                                    window.location.href = `${apiUrl}/auth/sso/google?mode=signup&userType=user`;
                                }}
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
                                onClick={() => {
                                    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
                                    window.location.href = `${apiUrl}/auth/sso/microsoft?mode=signup&userType=user`;
                                }}
                                className="flex items-center justify-center gap-2"
                            >
                                <svg className="w-5 h-5" viewBox="0 0 23 23">
                                    <rect width="10.763" height="10.763" fill="#f25022" />
                                    <rect x="12.237" width="10.763" height="10.763" fill="#7fba00" />
                                    <rect y="12.237" width="10.763" height="10.763" fill="#00a4ef" />
                                    <rect x="12.237" y="12.237" width="10.763" height="10.763" fill="#ffb900" />
                                </svg>
                                Microsoft
                            </Button>
                        </div>
                    </>
                )}

                {/* Step 2: Guardian Information (for minors) */}
                {currentStep === 2 && isMinor && (
                    <form onSubmit={step2Form.handleSubmit(handleStep2)} className="space-y-4">
                        <div className="bg-info-light border border-info rounded-lg p-4 mb-6">
                            <div className="flex">
                                <svg
                                    className="h-5 w-5 text-info mr-3"
                                    xmlns="http://www.w3.org/2000/svg"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                >
                                    <path
                                        fillRule="evenodd"
                                        d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                                        clipRule="evenodd"
                                    />
                                </svg>
                                <div>
                                    <h3 className="text-sm font-medium text-info">Guardian Consent Required</h3>
                                    <p className="mt-1 text-sm text-info-dark">
                                        As you are under 18, we need your parent or guardian&apos;s consent to create your account.
                                        We&apos;ll send them a verification email.
                                    </p>
                                </div>
                            </div>
                        </div>

                        <Input
                            label="Guardian Email"
                            type="email"
                            required
                            {...step2Form.register('guardianEmail')}
                            error={step2Form.formState.errors.guardianEmail?.message}
                            helperText="We'll send a verification link to this email"
                        />

                        <Input
                            label="Guardian Name"
                            type="text"
                            required
                            {...step2Form.register('guardianName')}
                            error={step2Form.formState.errors.guardianName?.message}
                        />

                        <Input
                            label="Relationship"
                            type="text"
                            required
                            placeholder="e.g., Parent, Guardian"
                            {...step2Form.register('relationship')}
                            error={step2Form.formState.errors.relationship?.message}
                        />

                        <div className="flex justify-between pt-4">
                            <Button
                                type="button"
                                variant="secondary"
                                onClick={() => setCurrentStep(1)}
                            >
                                Back
                            </Button>
                            <Button type="submit" variant="primary" loading={loading}>
                                Submit
                            </Button>
                        </div>
                    </form>
                )}

                {/* Step 3: Verification */}
                {currentStep === 3 && (
                    <div className="text-center space-y-6">
                        <div className="inline-flex items-center justify-center w-20 h-20 bg-success-light rounded-full">
                            <svg
                                className="w-12 h-12 text-success"
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

                        <div>
                            <h2 className="text-2xl font-semibold text-gray-900">
                                {isMinor ? 'Guardian verification sent!' : 'Check your email'}
                            </h2>
                            <p className="mt-2 text-gray-600">
                                {isMinor
                                    ? `We&apos;ve sent a verification email to ${step2Form.watch('guardianEmail')}. Your account will be activated once your guardian approves.`
                                    : `We&apos;ve sent a verification link to ${step1Data?.email}. Please check your inbox and click the link to activate your account.`}
                            </p>
                        </div>

                        <div className="border-t border-gray-200 pt-6">
                            <p className="text-sm text-gray-600">
                                Didn&apos;t receive the email?{' '}
                                <button className="text-purple hover:text-purple-hover font-medium">
                                    Resend verification email
                                </button>
                            </p>
                        </div>

                        <a href="/login" className="btn-primary">
                            Go to Login
                        </a>
                    </div>
                )}

                {/* Sign in link */}
                {currentStep < 3 && (
                    <div className="text-center text-sm border-t border-gray-200 pt-6">
                        <span className="text-gray-600">Already have an account? </span>
                        <a
                            href="/login"
                            className="font-medium text-purple hover:text-purple-hover focus-ring rounded"
                        >
                            Sign in
                        </a>
                    </div>
                )}
            </div>
        </div>
    );
}
