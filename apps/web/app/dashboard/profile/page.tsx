'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { PasswordStrength } from '@/components/ui/PasswordStrength';
import { useAuth } from '@/lib/auth_context';
import { api } from '@/lib/api';

// Profile Schema
const profileSchema = z.object({
    name: z.string().min(2, 'Name must be at least 2 characters'),
    email: z.string().email('Invalid email address'),
    phone: z.string().optional(),
    organization: z.string().optional(),
});

// Password Schema
const passwordSchema = z.object({
    currentPassword: z.string().min(1, 'Current password is required'),
    newPassword: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
}).refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
});

type ProfileFormData = z.infer<typeof profileSchema>;
type PasswordFormData = z.infer<typeof passwordSchema>;

export default function ProfilePage() {
    const { user, setUser } = useAuth();
    const [loading, setLoading] = useState(false);
    const [passwordLoading, setPasswordLoading] = useState(false);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);

    const {
        register: registerProfile,
        handleSubmit: handleProfileSubmit,
        formState: { errors: profileErrors },
        reset: resetProfile,
    } = useForm<ProfileFormData>({
        resolver: zodResolver(profileSchema),
    });

    useEffect(() => {
        if (user) {
            resetProfile({
                name: user.name,
                email: user.email,
                organization: user.organization || '',
            });
        }
    }, [user, resetProfile]);

    const {
        register: registerPassword,
        handleSubmit: handlePasswordSubmit,
        watch,
        reset: resetPassword,
        formState: { errors: passwordErrors },
    } = useForm<PasswordFormData>({
        resolver: zodResolver(passwordSchema),
    });

    const onProfileSubmit = async (data: ProfileFormData) => {
        setLoading(true);
        setSuccessMessage(null);
        try {
            const response = await api.put('/api/v1/fiduciary/profile', data);
            setUser({ ...user!, ...response.data });
            setSuccessMessage('Profile updated successfully');
            setTimeout(() => setSuccessMessage(null), 3000);
        } catch (error) {
            console.error('Profile update error:', error);
            alert('Failed to update profile.');
        } finally {
            setLoading(false);
        }
    };

    const onPasswordSubmit = async (data: PasswordFormData) => {
        setPasswordLoading(true);
        setSuccessMessage(null);
        try {
            await api.put('/api/v1/auth/password', {
                current_password: data.currentPassword,
                new_password: data.newPassword
            });
            setSuccessMessage('Password changed successfully');
            resetPassword();
            setTimeout(() => setSuccessMessage(null), 3000);
        } catch (error) {
            console.error('Password change error:', error);
            alert('Failed to change password.');
        } finally {
            setPasswordLoading(false);
        }
    };

    return (
        <DashboardLayout>
            <div className="space-y-6 max-w-4xl mx-auto">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Profile Settings</h1>
                    <p className="text-gray-600">Manage your account information and security</p>
                </div>

                {successMessage && (
                    <div className="bg-success-light text-success-dark px-4 py-3 rounded-lg flex items-center">
                        <svg
                            className="w-5 h-5 mr-2"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2}
                            stroke="currentColor"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
                        </svg>
                        {successMessage}
                    </div>
                )}

                {/* Personal Information */}
                <Card>
                    <CardHeader>
                        <CardTitle>Personal Information</CardTitle>
                        <CardDescription>Update your personal details and contact info.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleProfileSubmit(onProfileSubmit)} className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <Input
                                    label="Full Name"
                                    {...registerProfile('name')}
                                    error={profileErrors.name?.message}
                                />
                                <Input
                                    label="Email Address"
                                    type="email"
                                    {...registerProfile('email')}
                                    error={profileErrors.email?.message}
                                    disabled // Email usually requires verification to change
                                    helperText="Contact support to change email"
                                />
                                <Input
                                    label="Phone Number"
                                    {...registerProfile('phone')}
                                    error={profileErrors.phone?.message}
                                />
                                {user?.type === 'fiduciary' && (
                                    <Input
                                        label="Organization"
                                        {...registerProfile('organization')}
                                        error={profileErrors.organization?.message}
                                        disabled
                                    />
                                )}
                            </div>
                            <div className="flex justify-end">
                                <Button type="submit" variant="primary" loading={loading}>
                                    Save Changes
                                </Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>

                {/* Security */}
                <Card>
                    <CardHeader>
                        <CardTitle>Security</CardTitle>
                        <CardDescription>Manage your password and security preferences.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handlePasswordSubmit(onPasswordSubmit)} className="space-y-4 max-w-md">
                            <Input
                                label="Current Password"
                                type="password"
                                {...registerPassword('currentPassword')}
                                error={passwordErrors.currentPassword?.message}
                            />

                            <div className="space-y-2">
                                <Input
                                    label="New Password"
                                    type="password"
                                    {...registerPassword('newPassword')}
                                    error={passwordErrors.newPassword?.message}
                                />
                                <PasswordStrength password={watch('newPassword') || ''} />
                            </div>

                            <Input
                                label="Confirm New Password"
                                type="password"
                                {...registerPassword('confirmPassword')}
                                error={passwordErrors.confirmPassword?.message}
                            />

                            <div className="flex justify-end">
                                <Button type="submit" variant="secondary" loading={passwordLoading}>
                                    Update Password
                                </Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            </div>
        </DashboardLayout>
    );
}
