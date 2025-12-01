const fs = require('fs');
const path = 'd:\\arc-demo\\apps\\web\\lib\\hooks\\useAuth.ts';
const content = `'use client';

import { useState, useEffect, createContext, useContext } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { api } from '../api';

interface User {
    id: string;
    name: string;
    email: string;
    type: 'user' | 'fiduciary';
    token?: string;
    organization?: string;
}

interface AuthContextType {
    user: User | null;
    setUser: (user: User | null) => void;
    loading: boolean;
    login: (email: string, password: string, type: 'user' | 'fiduciary') => Promise<void>;
    logout: () => Promise<void>;
    checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const router = useRouter();
    const pathname = usePathname();

    const checkAuth = async () => {
        const userStr = localStorage.getItem('user');
        if (!userStr) {
            setLoading(false);
            return;
        }

        try {
            const storedUser = JSON.parse(userStr);
            // Validate token with backend
            const endpoint = storedUser.type === 'fiduciary'
                ? '/api/v1/auth/fiduciary/me'
                : '/api/v1/auth/user/me';

            const response = await api.get(endpoint);

            // Update user with fresh data from backend, keeping the token
            setUser({ ...response.data, token: storedUser.token, type: storedUser.type });
        } catch (error) {
            console.error('Auth check failed:', error);
            localStorage.removeItem('user');
            setUser(null);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        checkAuth();
    }, []);

    const login = async (email: string, password: string, type: 'user' | 'fiduciary') => {
        try {
            const endpoint = type === 'fiduciary'
                ? '/api/v1/auth/fiduciary/login'
                : '/api/v1/auth/user/login';

            const response = await api.post(endpoint, { email, password });

            const { token, user: userData } = response.data;

            const userWithToken = { ...userData, token, type };
            localStorage.setItem('user', JSON.stringify(userWithToken));
            setUser(userWithToken);

            router.push('/dashboard');
        } catch (error) {
            console.error('Login failed:', error);
            throw error;
        }
    };

    const logout = async () => {
        try {
            if (user) {
                const endpoint = user.type === 'fiduciary'
                    ? '/api/v1/auth/fiduciary/logout'
                    : '/api/v1/auth/user/logout';
                await api.post(endpoint);
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            localStorage.removeItem('user');
            setUser(null);
            router.push('/login');
        }
    };

    return (
        <AuthContext.Provider value={{ user, setUser, loading, login, logout, checkAuth }}>
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
`;

try {
    if (fs.existsSync(path)) {
        fs.unlinkSync(path);
    }
    fs.writeFileSync(path, content, 'utf8');
    console.log('Successfully wrote useAuth.ts');
} catch (error) {
    console.error('Error writing file:', error);
}
