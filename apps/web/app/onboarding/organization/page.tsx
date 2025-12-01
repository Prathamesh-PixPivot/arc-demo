'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';

const organizationSchema = z.object({
    organizationName: z.string().min(2, 'Organization name is required'),
    industry: z.string().min(1, 'Industry is required'),
    companySize: z.string().min(1, 'Company size is required'),
    taxId: z.string().optional(),
    website: z.string().url('Invalid URL').optional().or(z.literal('')),
    address: z.string().min(5, 'Address is required'),
    country: z.string().min(1, 'Country is required'),
});

type OrganizationFormData = z.infer<typeof organizationSchema>;

const industries = [
    { value: 'technology', label: 'Technology' },
    { value: 'healthcare', label: 'Healthcare' },
    { value: 'finance', label: 'Finance' },
    { value: 'retail', label: 'Retail' },
    { value: 'education', label: 'Education' },
    { value: 'other', label: 'Other' },
];

const companySizes = [
    { value: '1-10', label: '1-10 employees' },
    { value: '11-50', label: '11-50 employees' },
    { value: '51-200', label: '51-200 employees' },
    { value: '201-500', label: '201-500 employees' },
    { value: '500+', label: '500+ employees' },
];

const countries = [
    { value: 'US', label: 'United States' },
    { value: 'UK', label: 'United Kingdom' },
    { value: 'CA', label: 'Canada' },
    { value: 'IN', label: 'India' },
    // Add more as needed
];

export default function OnboardingOrganizationPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<OrganizationFormData>({
        resolver: zodResolver(organizationSchema),
    });

    const onSubmit = async (data: OrganizationFormData) => {
        setLoading(true);
        setError(null);
        try {
            // Assuming there's an endpoint to update organization details or create one linked to the user
            // Since the user is already created via SSO, we might need to update their profile or create a tenant.
            // For now, let's assume we post to /auth/fiduciary/onboarding or similar.
            // Or maybe /api/v1/tenants if they are creating a new tenant.

            // Based on previous signup logic, we might need to create a tenant.
            await api.post('/auth/fiduciary/onboarding', data);

            router.push('/dashboard');
        } catch (err: any) {
            console.error('Onboarding error:', err);
            setError(err.response?.data?.message || 'Failed to save organization details');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8 bg-white p-8 rounded-lg shadow-md">
                <div>
                    <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
                        Setup your Organization
                    </h2>
                    <p className="mt-2 text-center text-sm text-gray-600">
                        Please provide details about your organization to continue.
                    </p>
                </div>

                {error && (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative">
                        {error}
                    </div>
                )}

                <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
                    <div className="rounded-md shadow-sm -space-y-px">
                        <Input
                            label="Organization Name"
                            type="text"
                            required
                            {...register('organizationName')}
                            error={errors.organizationName?.message}
                        />

                        <div className="grid grid-cols-2 gap-4 mt-4">
                            <Select
                                label="Industry"
                                required
                                options={industries}
                                {...register('industry')}
                                error={errors.industry?.message}
                            />
                            <Select
                                label="Company Size"
                                required
                                options={companySizes}
                                {...register('companySize')}
                                error={errors.companySize?.message}
                            />
                        </div>

                        <div className="mt-4">
                            <Input
                                label="Address"
                                type="text"
                                required
                                {...register('address')}
                                error={errors.address?.message}
                            />
                        </div>

                        <div className="mt-4">
                            <Select
                                label="Country"
                                required
                                options={countries}
                                {...register('country')}
                                error={errors.country?.message}
                            />
                        </div>
                    </div>

                    <div>
                        <Button
                            type="submit"
                            variant="primary"
                            fullWidth
                            loading={loading}
                        >
                            Complete Setup
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
