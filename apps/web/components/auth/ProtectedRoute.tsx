'use client';

import { ReactNode } from 'react';
import { useAuth } from '@/lib/auth_context';
import { useRouter } from 'next/navigation';

interface ProtectedRouteProps {
    children: ReactNode;
    requiredType?: 'user' | 'fiduciary' | 'superadmin';
}

export function ProtectedRoute({ children, requiredType }: ProtectedRouteProps) {
    const { user, loading } = useAuth();
    const router = useRouter();

    if (loading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-purple"></div>
                    <p className="mt-4 text-gray-600">Loading...</p>
                </div>
            </div>
        );
    }

    if (!user) {
        router.push('/login');
        return null;
    }

    // Onboarding Checks
    const pathname = typeof window !== 'undefined' ? window.location.pathname : '';
    const isOnboardingPage = pathname.startsWith('/onboarding');

    if (user.type === 'fiduciary' && (!user.organization || !user.organization.id) && !isOnboardingPage) {
        router.push('/onboarding/organization');
        return null;
    }

    if (user.type === 'user' && (!user.age || user.age === 0) && !isOnboardingPage) {
        router.push('/onboarding/profile');
        return null;
    }

    // If on onboarding page but profile is complete, redirect to dashboard
    if (isOnboardingPage) {
        if (user.type === 'fiduciary' && user.organization && user.organization.id) {
            router.push('/dashboard');
            return null;
        }
        if (user.type === 'user' && user.age && user.age > 0) {
            router.push('/dashboard');
            return null;
        }
    }

    if (requiredType && user.type !== requiredType) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="card-elevated max-w-md p-8 text-center">
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
                    <h2 className="text-xl font-semibold text-gray-900 mb-2">Access Denied</h2>
                    <p className="text-gray-600 mb-6">
                        You don&apos;t have permission to access this page.
                    </p>
                    <a href="/dashboard" className="btn-primary">
                        Go to Dashboard
                    </a>
                </div>
            </div>
        );
    }

    return <>{children}</>;
}
