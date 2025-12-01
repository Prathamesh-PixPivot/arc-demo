'use client';

import { useAuth } from '@/lib/auth_context';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
} from 'recharts';

const data = [
    { name: 'Mon', consents: 40 },
    { name: 'Tue', consents: 30 },
    { name: 'Wed', consents: 20 },
    { name: 'Thu', consents: 27 },
    { name: 'Fri', consents: 18 },
    { name: 'Sat', consents: 23 },
    { name: 'Sun', consents: 34 },
];

const activityFeed = [
    { id: 1, action: 'Login', user: 'John Doe', time: '2 mins ago' },
    { id: 2, action: 'Consent Updated', user: 'Jane Smith', time: '1 hour ago' },
    { id: 3, action: 'New DSR Request', user: 'Mike Johnson', time: '3 hours ago' },
    { id: 4, action: 'Profile Update', user: 'Sarah Williams', time: '5 hours ago' },
    { id: 5, action: 'Login', user: 'David Brown', time: '1 day ago' },
];

export default function DashboardPage() {
    const { user, loading, authError, checkAuth } = useAuth();

    if (loading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-purple"></div>
                    <p className="mt-4 text-gray-600">Loading dashboard...</p>
                </div>
            </div>
        );
    }

    if (authError) {
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
                                d="M12 9v2.25m0 4.5h.008v.008H12v-.008zM21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                            />
                        </svg>
                    </div>
                    <h2 className="text-xl font-semibold text-gray-900 mb-2">Authentication Error</h2>
                    <p className="text-gray-600 mb-6">{authError}</p>
                    <div className="flex gap-3 justify-center">
                        <button onClick={() => checkAuth()} className="btn-secondary">
                            Retry
                        </button>
                        <a href="/login" className="btn-primary">
                            Login Again
                        </a>
                    </div>
                </div>
            </div>
        );
    }

    if (!user) {
        return null; // ProtectedRoute will handle redirect
    }

    return (
        <ProtectedRoute requiredType="fiduciary">
            <div className="min-h-screen bg-gray-50">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                    <div className="mb-8">
                        <h1 className="text-3xl font-bold text-gray-900">
                            Welcome back, {user?.name}
                        </h1>
                        <p className="mt-1 text-sm text-gray-500">
                            Here's what's happening with your organization today.
                        </p>
                    </div>

                    {/* Stats Cards */}
                    <div className="grid grid-cols-1 gap-5 sm:grid-cols-3 mb-8">
                        <div className="bg-white overflow-hidden shadow rounded-lg">
                            <div className="px-4 py-5 sm:p-6">
                                <dt className="text-sm font-medium text-gray-500 truncate">
                                    Total Consents Managed
                                </dt>
                                <dd className="mt-1 text-3xl font-semibold text-purple-600">
                                    12,345
                                </dd>
                            </div>
                        </div>
                        <div className="bg-white overflow-hidden shadow rounded-lg">
                            <div className="px-4 py-5 sm:p-6">
                                <dt className="text-sm font-medium text-gray-500 truncate">
                                    Pending DSRs
                                </dt>
                                <dd className="mt-1 text-3xl font-semibold text-yellow-600">
                                    23
                                </dd>
                            </div>
                        </div>
                        <div className="bg-white overflow-hidden shadow rounded-lg">
                            <div className="px-4 py-5 sm:p-6">
                                <dt className="text-sm font-medium text-gray-500 truncate">
                                    Trust Score
                                </dt>
                                <dd className="mt-1 text-3xl font-semibold text-green-600">
                                    98%
                                </dd>
                            </div>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                        {/* Chart */}
                        <div className="bg-white shadow rounded-lg p-6">
                            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                                Consent Activity
                            </h3>
                            <div className="h-72">
                                <ResponsiveContainer width="100%" height="100%">
                                    <LineChart
                                        data={data}
                                        margin={{
                                            top: 5,
                                            right: 30,
                                            left: 20,
                                            bottom: 5,
                                        }}
                                    >
                                        <CartesianGrid strokeDasharray="3 3" />
                                        <XAxis dataKey="name" />
                                        <YAxis />
                                        <Tooltip />
                                        <Line
                                            type="monotone"
                                            dataKey="consents"
                                            stroke="#7c3aed" // Purple-600
                                            activeDot={{ r: 8 }}
                                        />
                                    </LineChart>
                                </ResponsiveContainer>
                            </div>
                        </div>

                        {/* Activity Feed */}
                        <div className="bg-white shadow rounded-lg p-6">
                            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                                Recent Activity
                            </h3>
                            <div className="flow-root">
                                <ul role="list" className="-my-5 divide-y divide-gray-200">
                                    {activityFeed.map((activity) => (
                                        <li key={activity.id} className="py-4">
                                            <div className="flex items-center space-x-4">
                                                <div className="flex-shrink-0">
                                                    <span className="inline-block h-8 w-8 rounded-full bg-purple-100 text-purple-600 flex items-center justify-center font-bold text-xs">
                                                        {activity.user.charAt(0)}
                                                    </span>
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <p className="text-sm font-medium text-gray-900 truncate">
                                                        {activity.action}
                                                    </p>
                                                    <p className="text-sm text-gray-500 truncate">
                                                        by {activity.user}
                                                    </p>
                                                </div>
                                                <div>
                                                    <span className="inline-flex items-center shadow-sm px-2.5 py-0.5 border border-gray-300 text-sm leading-5 font-medium rounded-full text-gray-700 bg-white hover:bg-gray-50">
                                                        {activity.time}
                                                    </span>
                                                </div>
                                            </div>
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </ProtectedRoute>
    );
}
