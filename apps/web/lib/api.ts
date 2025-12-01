import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
    withCredentials: true, // Important for cookies if used
});

// Request interceptor to add auth token
api.interceptors.request.use(
    (config) => {
        // Check if running in browser
        if (typeof window !== 'undefined') {
            const userStr = localStorage.getItem('user');
            if (userStr) {
                try {
                    const user = JSON.parse(userStr);
                    if (user.token) {
                        config.headers.Authorization = `Bearer ${user.token}`;
                    }
                } catch (e) {
                    console.error('Error parsing user from local storage', e);
                }
            }
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Response interceptor to handle 401s
api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response && error.response.status === 401) {
            // If 401, clear local storage and redirect to login
            if (typeof window !== 'undefined') {
                localStorage.removeItem('user');
                // Only redirect if not already on login page to avoid loops
                if (!window.location.pathname.includes('/login')) {
                    window.location.href = '/login';
                }
            }
        }
        return Promise.reject(error);
    }
);
