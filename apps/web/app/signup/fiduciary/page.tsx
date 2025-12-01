'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { StepIndicator } from '@/components/ui/StepIndicator';
import { PasswordStrength } from '@/components/ui/PasswordStrength';
import { api } from '@/lib/api';

const steps = [
    { id: 1, name: 'Account', description: 'Personal details' },
    { id: 2, name: 'Organization', description: 'Company info' },
    { id: 3, name: 'Verify', description: 'Email verification' },
];

const step1Schema = z.object({
    name: z.string().min(1, 'Name is required'),
    email: z.string().email('Please enter a valid email address'),
    phone: z.string().min(10, 'Please enter a valid phone number'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
});

const step2Schema = z.object({
    organizationName: z.string().min(1, 'Organization name is required'),
    taxId: z.string().optional(),
    website: z.string().url('Please enter a valid URL').optional().or(z.literal('')),
    industry: z.string().min(1, 'Industry is required'),
    companySize: z.string().min(1, 'Company size is required'),
    address: z.string().min(1, 'Address is required'),
    country: z.string().min(1, 'Country is required'),
});

type Step1Data = z.infer<typeof step1Schema>;
type Step2Data = z.infer<typeof step2Schema>;

const industries = [
    { value: 'technology', label: 'Technology' },
    { value: 'healthcare', label: 'Healthcare' },
    { value: 'finance', label: 'Finance' },
    { value: 'retail', label: 'Retail' },
    { value: 'manufacturing', label: 'Manufacturing' },
    { value: 'education', label: 'Education' },
    { value: 'government', label: 'Government' },
    { value: 'other', label: 'Other' },
];

const companySizes = [
    { value: '1-10', label: '1-10 employees' },
    { value: '11-50', label: '11-50 employees' },
    { value: '51-200', label: '51-200 employees' },
    { value: '201-500', label: '201-500 employees' },
    { value: '501-1000', label: '501-1000 employees' },
    { value: '1000+', label: '1000+ employees' },
];

const countries = [
    { value: 'IN', label: 'India' },
    { value: 'US', label: 'United States' },
    { value: 'UK', label: 'United Kingdom' },
    { value: 'DE', label: 'Germany' },
    { value: 'FR', label: 'France' },
    { value: 'SG', label: 'Singapore' },
    { value: 'AU', label: 'Australia' },
    { value: 'CA', label: 'Canada' },
];

export default function FiduciarySignupPage() {
    const [currentStep, setCurrentStep] = useState(1);
    const [loading, setLoading] = useState(false);
    const [step1Data, setStep1Data] = useState<Step1Data | null>(null);
    const [error, setError] = useState<string | null>(null);

    const step1Form = useForm<Step1Data>({
        resolver: zodResolver(step1Schema),
    });

    const step2Form = useForm<Step2Data>({
        resolver: zodResolver(step2Schema),
    });

    const handleStep1 = async (data: Step1Data) => {
        setStep1Data(data);
        setCurrentStep(2);
    };

    const handleStep2 = async (data: Step2Data) => {
        if (!step1Data) return;

        setLoading(true);
        try {
            const nameParts = step1Data.name.split(' ');
            const firstName = nameParts[0];
            const lastName = nameParts.length > 1 ? nameParts.slice(1).join(' ') : '';

            const payload = {
                email: step1Data.email,
                firstName: firstName,
                lastName: lastName,
                phone: step1Data.phone,
                password: step1Data.password,
                confirmPassword: step1Data.confirmPassword,
                role: "admin", // Default role, backend assigns Super Admin
                organization: {
                    name: data.organizationName,
                    industry: data.industry,
                    companySize: data.companySize,
                    taxId: data.taxId,
                    website: data.website,
                    email: step1Data.email, // Using admin email for org email default
                    phone: step1Data.phone, // Using admin phone for org phone default
                    address: data.address,
                    country: data.country
                }
            };

            await api.post('/auth/fiduciary/signup', payload);

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
                                d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Z"
                            />
                        </svg>
                    </div>
                    <h1 className="text-3xl font-semibold text-gray-900">Create fiduciary account</h1>
                    <p className="mt-2 text-sm text-gray-600">
                        Register your organization on Arc Privacy Platform
                    </p>
                </div>

                {error && (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
                        <span className="block sm:inline">{error}</span>
                    </div>
                )}

                {/* Step Indicator */}
                <StepIndicator steps={steps} currentStep={currentStep} />

                {/* Step 1: Personal Information */}
                {currentStep === 1 && (
                    <>
                        <form onSubmit={step1Form.handleSubmit(handleStep1)} className="space-y-4">
                            <Input
                                label="Full Name"
                                type="text"
                                required
                                {...step1Form.register('name')}
                                error={step1Form.formState.errors.name?.message}
                            />

                            <Input
                                label="Email Address"
                                type="email"
                                autoComplete="email"
                                required
                                {...step1Form.register('email')}
                                error={step1Form.formState.errors.email?.message}
                                helperText="This will be your login email"
                            />

                            <Input
                                label="Phone Number"
                                type="tel"
                                autoComplete="tel"
                                required
                                {...step1Form.register('phone')}
                                error={step1Form.formState.errors.phone?.message}
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
                    </>
                )}

                {/* Step 2: Organization Information */}
                {currentStep === 2 && (
                    <form onSubmit={step2Form.handleSubmit(handleStep2)} className="space-y-4">
                        <Input
                            label="Organization Name"
                            type="text"
                            required
                            {...step2Form.register('organizationName')}
                            error={step2Form.formState.errors.organizationName?.message}
                        />

                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <Input
                                label="Tax ID / Registration Number"
                                type="text"
                                helperText="Optional"
                                {...step2Form.register('taxId')}
                                error={step2Form.formState.errors.taxId?.message}
                            />
                            <Input
                                label="Website"
                                type="url"
                                placeholder="https://example.com"
                                helperText="Optional"
                                {...step2Form.register('website')}
                                error={step2Form.formState.errors.website?.message}
                            />
                        </div>

                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <Select
                                label="Industry"
                                required
                                options={industries}
                                {...step2Form.register('industry')}
                                error={step2Form.formState.errors.industry?.message}
                            />
                            <Select
                                label="Company Size"
                                required
                                options={companySizes}
                                {...step2Form.register('companySize')}
                                error={step2Form.formState.errors.companySize?.message}
                            />
                        </div>

                        <Input
                            label="Address"
                            type="text"
                            required
                            {...step2Form.register('address')}
                            error={step2Form.formState.errors.address?.message}
                        />

                        <Select
                            label="Country"
                            required
                            options={countries}
                            {...step2Form.register('country')}
                            error={step2Form.formState.errors.country?.message}
                        />

                        <div className="flex justify-between pt-4">
                            <button
                                type="button"
                                onClick={() => setCurrentStep(1)}
                                className="btn-tertiary"
                            >
                                Back
                            </button>
                            <Button type="submit" variant="primary" loading={loading}>
                                Create Account
                            </Button>
                        </div>
                    </form>
                )}

                {/* Step 3: Verification */}
                {currentStep === 3 && (
                    <div className="text-center space-y-6">
                        <div className="bg-purple-50 rounded-lg p-6">
                            <div className="flex flex-col items-center">
                                <div className="h-12 w-12 bg-purple-100 rounded-full flex items-center justify-center mb-4">
                                    <svg className="h-6 w-6 text-purple" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                                    </svg>
                                </div>
                                <h3 className="text-lg font-medium text-gray-900">Check your email</h3>
                                <p className="mt-2 text-sm text-gray-600">
                                    We&apos;ve sent a verification link to <strong>{step1Data?.email}</strong>
                                </p>
                                <div className="mt-4 text-left bg-white p-4 rounded-md shadow-sm w-full">
                                    <p className="text-sm font-medium text-gray-900 mb-2">Next steps:</p>
                                    <ul className="mt-2 text-sm text-purple-dark space-y-1">
                                        <li>• Verify your email address</li>
                                        <li>• Our team will review your organization</li>
                                        <li>• You&apos;ll receive approval within 24-48 hours</li>
                                        <li>• Once approved, you can start managing consents</li>
                                    </ul>
                                </div>
                            </div>
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

                {/* SSO Buttons */}
                {currentStep === 1 && (
                    <>
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
                                    window.location.href = `${apiUrl}/auth/sso/google?mode=signup&userType=fiduciary`;
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
                                    window.location.href = `${apiUrl}/auth/sso/microsoft?mode=signup&userType=fiduciary`;
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
