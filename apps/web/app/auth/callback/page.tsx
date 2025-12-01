'use client';

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/lib/auth_context';

function CallbackContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const { setUser } = useAuth();

    useEffect(() => {
        const token = searchParams.get('token');
        const userType = searchParams.get('userType');
        const next = searchParams.get('next');
        const error = searchParams.get('error');

        if (error) {
            const errorMsg = searchParams.get('hint') || 'Authentication failed';
            router.push(`/login?error=${error}&message=${encodeURIComponent(errorMsg)}`);
            return;
        }

        if (!token || !userType) {
            router.push('/login?error=invalid_callback');
            return;
        }

        // Decode JWT to get user info
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));

            const user = {
                id: payload.fiduciaryId || payload.id,
                email: payload.email,
                name: payload.name || payload.email,
                type: userType as 'user' | 'fiduciary',
                token: token,
                organization: payload.organization,
                age: payload.age,
            };

            // Store token and type
            localStorage.setItem('token', token);
            localStorage.setItem('userType', userType);

            // Update auth context
            setUser(user);

            // Smart redirection based on `next` param or onboarding status
            if (next) {
                router.push(next);
            } else if (userType === 'fiduciary' && !payload.organization) {
                router.push('/onboarding/organization');
            } else if (userType === 'user' && !payload.age) {
                router.push('/onboarding/profile');
            } else {
                router.push('/dashboard');
            }
        } catch (error) {
            console.error('Token decode failed:', error);
            router.push('/login?error=invalid_token');
        }
    }, [searchParams, router, setUser]);

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="text-center">
                <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-purple"></div>
                <p className="mt-4 text-gray-600">Completing authentication...</p>
            </div>
        </div>
    );
}

export default function AuthCallbackPage() {
    return (
        <Suspense fallback={<div>Loading...</div>}>
            <CallbackContent />
        </Suspense>
    );
}
