'use client';

import { useState, useEffect, createContext, useContext } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { api } from './api';

interface User {
    id: string;
    name: string;
    email: string;
    type: 'user' | 'fiduciary';
    token?: string;
    organization?: any;
    avatar?: string;
    age?: number;
    guardianEmail?: string;
}

interface AuthContextType {
    user: User | null;
    setUser: (user: User | null) => void;
    loading: boolean;
    authError: string | null;
    login: (email: string, password: string, type: 'user' | 'fiduciary') => Promise<void>;
    logout: () => Promise<void>;
    checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const [authError, setAuthError] = useState<string | null>(null);
    const router = useRouter();
    const pathname = usePathname();

    const checkAuth = async () => {
        setAuthError(null);
        const token = localStorage.getItem('token');
        const userType = localStorage.getItem('userType');

        if (!token || !userType) {
            setLoading(false);
            return;
        }

        try {
            const endpoint = userType === 'fiduciary' ? '/api/v1/auth/fiduciary/me' : '/api/v1/auth/user/me';
            const response = await api.get(endpoint, {
                headers: { Authorization: `Bearer ${token}` },
            });

            setUser({
                ...response.data,
                type: userType as 'user' | 'fiduciary',
                token,
            });
        } catch (error: any) {
            // Don't crash - handle gracefully
            console.error('Auth check failed:', error);
            setAuthError(error.response?.data?.error || 'Failed to verify authentication');
            localStorage.removeItem('token');
            localStorage.removeItem('userType');
            setUser(null);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        checkAuth();
    }, []);

    const login = async (email: string, password: string, type: 'user' | 'fiduciary') => {
        setAuthError(null);
        try {
            const endpoint = type === 'fiduciary' ? '/api/v1/auth/fiduciary/login' : '/api/v1/auth/user/login';
            const response = await api.post(endpoint, { email, password });

            const data = response.data;

            // Store token separately for security
            localStorage.setItem('token', data.token);
            localStorage.setItem('userType', type);

            const userWithToken: User = {
                id: data.userId || data.fiduciaryId,
                email: data.email,
                name: data.name || email,
                type: type,
                token: data.token,
                organization: data.organization,
                age: data.age,
            };

            setUser(userWithToken);

            // Check onboarding status from backend response
            if (type === 'fiduciary' && (!data.organization || !data.organization.id)) {
                router.push('/onboarding/organization');
            } else if (type === 'user' && (!data.age || data.age === 0)) {
                router.push('/onboarding/profile');
            } else {
                router.push('/dashboard');
            }
        } catch (error: any) {
            const errorMessage = error.response?.data?.error || 'Login failed. Please check your credentials.';
            setAuthError(errorMessage);
            throw new Error(errorMessage);
        }
    };

    const logout = async () => {
        try {
            if (user) {
                const endpoint = user.type === 'fiduciary' ? '/api/v1/auth/fiduciary/logout' : '/api/v1/auth/user/logout';
                await api.post(endpoint, {}, {
                    headers: { Authorization: `Bearer ${user.token}` },
                });
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            localStorage.removeItem('token');
            localStorage.removeItem('userType');
            setUser(null);
            setAuthError(null);
            router.push('/login');
        }
    };

    return (
        <AuthContext.Provider value={{ user, setUser, loading, authError, login, logout, checkAuth }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
