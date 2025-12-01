'use client';

import { ReactNode, useState } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { Sidebar } from '@/components/layout/Sidebar';
import { Header } from '@/components/layout/Header';
import { cn } from '@/lib/utils';

interface DashboardLayoutProps {
    children: ReactNode;
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

    return (
        <ProtectedRoute>
            <div className="min-h-screen bg-gray-50">
                <Sidebar collapsed={sidebarCollapsed} onCollapse={setSidebarCollapsed} />
                <Header sidebarCollapsed={sidebarCollapsed} />

                <main
                    className={cn(
                        'pt-16 transition-all duration-300',
                        sidebarCollapsed ? 'ml-16' : 'ml-64'
                    )}
                >
                    <div className="p-6">{children}</div>
                </main>
            </div>
        </ProtectedRoute>
    );
}
