'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';

const profileSchema = z.object({
    age: z.number().min(1, 'Age is required').max(120, 'Invalid age'),
    location: z.string().min(2, 'Location is required'),
    guardianEmail: z.string().email('Invalid email').optional().or(z.literal('')),
});

type ProfileFormData = z.infer<typeof profileSchema>;

export default function OnboardingProfilePage() {
    const router = useRouter();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const {
        register,
        handleSubmit,
        watch,
        formState: { errors },
    } = useForm<ProfileFormData>({
        resolver: zodResolver(profileSchema),
    });

    const age = watch('age');
    const isMinor = age && age < 18;

    const onSubmit = async (data: ProfileFormData) => {
        setLoading(true);
        setError(null);

        if (isMinor && !data.guardianEmail) {
            setError('Guardian email is required for users under 18');
            return;
        }

        try {
            await api.post('/auth/user/onboarding', data);
            router.push('/dashboard');
        } catch (err: any) {
            console.error('Onboarding error:', err);
            setError(err.response?.data?.message || 'Failed to save profile');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8 bg-white p-8 rounded-lg shadow-md">
                <div>
                    <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
                        Complete your Profile
                    </h2>
                    <p className="mt-2 text-center text-sm text-gray-600">
                        We need a few more details to protect your privacy.
                    </p>
                </div>

                {error && (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative">
                        {error}
                    </div>
                )}

                <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
                    <div className="rounded-md shadow-sm space-y-4">
                        <Input
                            label="Age"
                            type="number"
                            required
                            {...register('age', { valueAsNumber: true })}
                            error={errors.age?.message}
                        />

                        <Input
                            label="Location (City, Country)"
                            type="text"
                            required
                            {...register('location')}
                            error={errors.location?.message}
                        />

                        {isMinor && (
                            <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-md">
                                <p className="text-sm text-yellow-800 mb-2">
                                    Since you are under 18, we need your guardian's consent.
                                </p>
                                <Input
                                    label="Guardian Email"
                                    type="email"
                                    required={isMinor}
                                    {...register('guardianEmail')}
                                    error={errors.guardianEmail?.message}
                                />
                            </div>
                        )}
                    </div>

                    <div>
                        <Button
                            type="submit"
                            variant="primary"
                            fullWidth
                            loading={loading}
                        >
                            Complete Profile
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
