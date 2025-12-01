'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/Table';
import { ConsentForm } from '@/lib/types';
import { api } from '@/lib/api';

export default function ConsentFormsPage() {
    const router = useRouter();
    const [forms, setForms] = useState<ConsentForm[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');

    const fetchForms = async () => {
        try {
            const response = await api.get('/api/v1/fiduciary/consent-forms');
            setForms(response.data);
        } catch (error) {
            console.error('Failed to fetch consent forms:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchForms();
    }, []);

    const filteredForms = forms.filter(
        (f) =>
            f.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            f.id.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const handleDelete = async (id: string) => {
        if (confirm('Are you sure you want to delete this form?')) {
            try {
                await api.delete(`/api/v1/fiduciary/consent-forms/${id}`);
                await fetchForms();
            } catch (error) {
                console.error('Failed to delete form:', error);
                alert('Failed to delete form. Please try again.');
            }
        }
    };

    return (
        <DashboardLayout>
            <div className="space-y-6">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900">Consent Forms</h1>
                        <p className="text-gray-600">Manage and publish consent collection forms</p>
                    </div>
                    <Button onClick={() => router.push('/dashboard/consent-forms/builder')}>
                        <svg
                            className="w-5 h-5 mr-2"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2}
                            stroke="currentColor"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                        </svg>
                        Create New Form
                    </Button>
                </div>

                {/* Filters */}
                <div className="flex items-center gap-4 bg-white p-4 rounded-lg border border-gray-200">
                    <div className="relative flex-1 max-w-sm">
                        <svg
                            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2}
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607z"
                            />
                        </svg>
                        <input
                            type="text"
                            placeholder="Search forms..."
                            className="w-full pl-9 pr-4 py-2 text-sm border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-purple focus:border-transparent"
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                        />
                    </div>
                </div>

                {/* Table */}
                <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>Name</TableHead>
                                <TableHead>Purposes</TableHead>
                                <TableHead>Version</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Last Updated</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredForms.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                                        No consent forms found.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredForms.map((form) => (
                                    <TableRow key={form.id}>
                                        <TableCell className="font-medium">{form.id}</TableCell>
                                        <TableCell className="font-medium text-gray-900">{form.name}</TableCell>
                                        <TableCell>
                                            <Badge variant="gray">{form.purposeIds.length} purposes</Badge>
                                        </TableCell>
                                        <TableCell>v{form.version}</TableCell>
                                        <TableCell>
                                            <Badge
                                                variant={
                                                    form.status === 'published'
                                                        ? 'success'
                                                        : form.status === 'archived'
                                                            ? 'gray'
                                                            : 'warning'
                                                }
                                                className="capitalize"
                                            >
                                                {form.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{form.lastUpdated}</TableCell>
                                        <TableCell className="text-right">
                                            <div className="flex justify-end gap-2">
                                                <button
                                                    onClick={() => router.push(`/dashboard/consent-forms/builder?id=${form.id}`)}
                                                    className="p-1 text-gray-500 hover:text-purple transition-colors"
                                                    title="Edit"
                                                >
                                                    <svg
                                                        className="w-4 h-4"
                                                        fill="none"
                                                        viewBox="0 0 24 24"
                                                        strokeWidth={2}
                                                        stroke="currentColor"
                                                    >
                                                        <path
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125"
                                                        />
                                                    </svg>
                                                </button>
                                                <button
                                                    onClick={() => handleDelete(form.id)}
                                                    className="p-1 text-gray-500 hover:text-error transition-colors"
                                                    title="Delete"
                                                >
                                                    <svg
                                                        className="w-4 h-4"
                                                        fill="none"
                                                        viewBox="0 0 24 24"
                                                        strokeWidth={2}
                                                        stroke="currentColor"
                                                    >
                                                        <path
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
                                                        />
                                                    </svg>
                                                </button>
                                            </div>
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
