'use client';

import { useState, useEffect } from 'react';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { Badge } from '@/components/ui/Badge';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/Table';
import { api } from '@/lib/api';

interface AuditLog {
    id: string;
    action: string;
    userId: string;
    resourceId: string;
    timestamp: string;
    details: string;
}

export default function AuditLogsPage() {
    const [logs, setLogs] = useState<AuditLog[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [filterAction, setFilterAction] = useState('all');

    const fetchLogs = async () => {
        try {
            const response = await api.get('/api/v1/fiduciary/audit-logs');
            setLogs(response.data);
        } catch (error) {
            console.error('Failed to fetch audit logs:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchLogs();
    }, []);

    const filteredLogs = logs.filter((log) => {
        const matchesSearch =
            log.userId.toLowerCase().includes(searchTerm.toLowerCase()) ||
            log.action.toLowerCase().includes(searchTerm.toLowerCase()) ||
            log.details.toLowerCase().includes(searchTerm.toLowerCase());

        const matchesFilter = filterAction === 'all' || log.action === filterAction;

        return matchesSearch && matchesFilter;
    });

    const actionOptions = [
        { value: 'all', label: 'All Actions' },
        { value: 'LOGIN', label: 'Login' },
        { value: 'CREATE_PURPOSE', label: 'Create Purpose' },
        { value: 'DELETE_USER', label: 'Delete User' },
        { value: 'CONSENT_GRANT', label: 'Consent Grant' },
        { value: 'DSR_REQUEST', label: 'DSR Request' },
    ];

    return (
        <DashboardLayout>
            <div className="space-y-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Audit Logs</h1>
                    <p className="text-gray-600">Track system activity and security events</p>
                </div>

                {/* Filters */}
                <div className="flex flex-col sm:flex-row gap-4 bg-white p-4 rounded-lg border border-gray-200">
                    <div className="flex-1">
                        <Input
                            placeholder="Search logs..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="w-full"
                        />
                    </div>
                    <div className="w-full sm:w-64">
                        <Select
                            options={actionOptions}
                            value={filterAction}
                            onChange={(e) => setFilterAction(e.target.value)}
                        />
                    </div>
                </div>

                {/* Table */}
                <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Timestamp</TableHead>
                                <TableHead>Action</TableHead>
                                <TableHead>User ID</TableHead>
                                <TableHead>Resource ID</TableHead>
                                <TableHead>Details</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                                        Loading logs...
                                    </TableCell>
                                </TableRow>
                            ) : filteredLogs.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center py-8 text-gray-500">
                                        No logs found matching your criteria.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredLogs.map((log) => (
                                    <TableRow key={log.id}>
                                        <TableCell className="whitespace-nowrap text-gray-500 text-xs">
                                            {new Date(log.timestamp).toLocaleString()}
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="gray" className="font-mono text-xs">
                                                {log.action}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-sm">{log.userId}</TableCell>
                                        <TableCell className="text-sm">{log.resourceId}</TableCell>
                                        <TableCell className="text-sm text-gray-600 max-w-xs truncate" title={log.details}>
                                            {log.details}
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>
        </DashboardLayout>
    );
}
