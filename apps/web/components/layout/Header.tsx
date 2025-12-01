'use client';

import { useState } from 'react';
import { useAuth } from '@/lib/auth_context';
import { cn } from '@/lib/utils';

interface HeaderProps {
    sidebarCollapsed: boolean;
}

export function Header({ sidebarCollapsed }: HeaderProps) {
    const { user, logout } = useAuth();
    const [dropdownOpen, setDropdownOpen] = useState(false);

    if (!user) return null;

    const initials = user.name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2);

    return (
        <header
            className={cn(
                'fixed top-0 right-0 z-40 h-16 bg-white border-b border-gray-200 transition-all duration-300',
                sidebarCollapsed ? 'left-16' : 'left-64'
            )}
        >
            <div className="h-full px-6 flex items-center justify-between">
                {/* Page Title - can be dynamic based on route */}
                <div>
                    <h1 className="text-xl font-semibold text-gray-900">Dashboard</h1>
                </div>

                {/* Right side - Notifications & User Menu */}
                <div className="flex items-center space-x-4">
                    {/* Notifications */}
                    <button
                        className="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg focus-ring"
                        aria-label="Notifications"
                    >
                        <svg
                            className="w-6 h-6"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={1.5}
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0"
                            />
                        </svg>
                        {/* Notification badge */}
                        <span className="absolute top-1.5 right-1.5 block h-2 w-2 rounded-full bg-error ring-2 ring-white" />
                    </button>

                    {/* User Menu */}
                    <div className="relative">
                        <button
                            onClick={() => setDropdownOpen(!dropdownOpen)}
                            className="flex items-center space-x-3 hover:bg-gray-100 rounded-lg px-3 py-2 focus-ring"
                            aria-expanded={dropdownOpen}
                            aria-haspopup="true"
                        >
                            <div className="w-8 h-8 rounded-full bg-purple text-white flex items-center justify-center text-sm font-medium">
                                {user.avatar ? (
                                    <img src={user.avatar} alt={user.name} className="w-8 h-8 rounded-full" />
                                ) : (
                                    initials
                                )}
                            </div>
                            <div className="hidden md:block text-left">
                                <p className="text-sm font-medium text-gray-900">{user.name}</p>
                                <p className="text-xs text-gray-500 capitalize">{user.type}</p>
                            </div>
                            <svg
                                className={cn(
                                    'w-4 h-4 text-gray-600 transition-transform',
                                    dropdownOpen && 'rotate-180'
                                )}
                                fill="none"
                                viewBox="0 0 24 24"
                                strokeWidth={2}
                                stroke="currentColor"
                            >
                                <path strokeLinecap="round" strokeLinejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" />
                            </svg>
                        </button>

                        {/* Dropdown Menu */}
                        {dropdownOpen && (
                            <>
                                <div
                                    className="fixed inset-0 z-10"
                                    onClick={() => setDropdownOpen(false)}
                                    aria-hidden="true"
                                />
                                <div className="absolute right-0 mt-2 w-56 bg-white rounded-lg shadow-elevated border border-gray-200 py-1 z-20">
                                    <div className="px-4 py-3 border-b border-gray-200">
                                        <p className="text-sm font-medium text-gray-900">{user.name}</p>
                                        <p className="text-sm text-gray-500 truncate">{user.email}</p>
                                    </div>

                                    <a
                                        href="/dashboard/profile"
                                        className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                                    >
                                        <svg
                                            className="w-4 h-4 mr-3 text-gray-600"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                            strokeWidth={1.5}
                                            stroke="currentColor"
                                        >
                                            <path
                                                strokeLinecap="round"
                                                strokeLinejoin="round"
                                                d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"
                                            />
                                        </svg>
                                        Profile Settings
                                    </a>

                                    <a
                                        href="/dashboard/preferences"
                                        className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                                    >
                                        <svg
                                            className="w-4 h-4 mr-3 text-gray-600"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                            strokeWidth={1.5}
                                            stroke="currentColor"
                                        >
                                            <path
                                                strokeLinecap="round"
                                                strokeLinejoin="round"
                                                d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z"
                                            />
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
                                        </svg>
                                        Preferences
                                    </a>

                                    <div className="border-t border-gray-200 my-1" />

                                    <button
                                        onClick={logout}
                                        className="flex items-center w-full px-4 py-2 text-sm text-error hover:bg-error-light"
                                    >
                                        <svg
                                            className="w-4 h-4 mr-3"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                            strokeWidth={1.5}
                                            stroke="currentColor"
                                        >
                                            <path
                                                strokeLinecap="round"
                                                strokeLinejoin="round"
                                                d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9"
                                            />
                                        </svg>
                                        Sign out
                                    </button>
                                </div>
                            </>
                        )}
                    </div>
                </div>
            </div>
        </header>
    );
}
