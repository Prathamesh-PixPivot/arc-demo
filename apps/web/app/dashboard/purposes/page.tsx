'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Textarea } from '@/components/ui/Textarea';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/Table';
import { api } from '@/lib/api';

interface Purpose {
    id: string;
    name: string;
    description: string;
    legalBasis: string;
    retentionPeriod: string;
    status: 'active' | 'inactive' | 'archived';
    lastUpdated: string;
}

const purposeSchema = z.object({
    name: z.string().min(1, 'Name is required'),
    description: z.string().min(10, 'Description must be at least 10 characters'),
    legalBasis: z.enum([
        'consent',
        'contract',
        'legal_obligation',
        'vital_interests',
        'public_task',
        'legitimate_interests',
    ]),
    retentionPeriod: z.string().min(1, 'Retention period is required'),
    status: z.enum(['active', 'inactive', 'archived']),
});

type PurposeFormData = z.infer<typeof purposeSchema>;

export default function PurposesPage() {
    const [purposes, setPurposes] = useState<Purpose[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingPurpose, setEditingPurpose] = useState<Purpose | null>(null);

    const {
        register,
        handleSubmit,
        reset,
        formState: { errors },
        setValue,
    } = useForm<PurposeFormData>({
        resolver: zodResolver(purposeSchema),
        defaultValues: {
            status: 'active',
            legalBasis: 'consent',
        },
    });

    const fetchPurposes = async () => {
        try {
            const response = await api.get('/api/v1/fiduciary/purposes');
            setPurposes(response.data);
        } catch (error) {
            console.error('Failed to fetch purposes:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchPurposes();
    }, []);

    const handleOpenModal = (purpose?: Purpose) => {
        if (purpose) {
            setEditingPurpose(purpose);
            reset({
                name: purpose.name,
                description: purpose.description,
                legalBasis: purpose.legalBasis as any,
                retentionPeriod: purpose.retentionPeriod,
                status: purpose.status,
            });
        } else {
            setEditingPurpose(null);
            reset({
                name: '',
                description: '',
                legalBasis: 'consent',
                retentionPeriod: '',
                status: 'active',
            });
        }
        setIsModalOpen(true);
    };

    const onSubmit = async (data: PurposeFormData) => {
        try {
            if (editingPurpose) {
                await api.put(`/api/v1/fiduciary/purposes/${editingPurpose.id}`, data);
            } else {
                await api.post('/api/v1/fiduciary/purposes', data);
            }
            await fetchPurposes();
            setIsModalOpen(false);
        } catch (error) {
            console.error('Failed to save purpose:', error);
            alert('Failed to save purpose. Please try again.');
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this purpose?')) return;

        try {
            await api.delete(`/api/v1/fiduciary/purposes/${id}`);
            await fetchPurposes();
        } catch (error) {
            console.error('Failed to delete purpose:', error);
            alert('Failed to delete purpose. Please try again.');
        }
    };

    const filteredPurposes = purposes.filter((p) =>
        p.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        p.id.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const legalBasisOptions = [
        { value: 'consent', label: 'Consent' },
        { value: 'contract', label: 'Contract' },
        { value: 'legal_obligation', label: 'Legal Obligation' },
        { value: 'vital_interests', label: 'Vital Interests' },
        { value: 'public_task', label: 'Public Task' },
        { value: 'legitimate_interests', label: 'Legitimate Interests' },
    ];

    const statusOptions = [
        { value: 'active', label: 'Active' },
        { value: 'inactive', label: 'Inactive' },
        { value: 'archived', label: 'Archived' },
    ];

    return (
        <DashboardLayout>
            <div className="space-y-6">
                <div className="flex justify-between items-center">
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900">Data Purposes</h1>
                        <p className="text-gray-600">Define why and how you process user data</p>
                    </div>
                    <Button onClick={() => handleOpenModal()}>
                        <svg
                            className="w-5 h-5 mr-2"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2}
                            stroke="currentColor"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                        </svg>
                        Add Purpose
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
                            placeholder="Search purposes..."
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
                                <TableHead>Legal Basis</TableHead>
                                <TableHead>Retention</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Last Updated</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                                        Loading purposes...
                                    </TableCell>
                                </TableRow>
                            ) : filteredPurposes.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                                        No purposes found.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredPurposes.map((purpose) => (
                                    <TableRow key={purpose.id}>
                                        <TableCell className="font-medium">{purpose.id}</TableCell>
                                        <TableCell>
                                            <div>
                                                <div className="font-medium text-gray-900">{purpose.name}</div>
                                                <div className="text-xs text-gray-500 truncate max-w-[200px]">
                                                    {purpose.description}
                                                </div>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="info" className="capitalize">
                                                {purpose.legalBasis.replace('_', ' ')}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{purpose.retentionPeriod}</TableCell>
                                        <TableCell>
                                            <Badge
                                                variant={
                                                    purpose.status === 'active'
                                                        ? 'success'
                                                        : purpose.status === 'inactive'
                                                            ? 'gray'
                                                            : 'warning'
                                                }
                                                className="capitalize"
                                            >
                                                {purpose.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{new Date(purpose.lastUpdated).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <div className="flex justify-end gap-2">
                                                <button
                                                    onClick={() => handleOpenModal(purpose)}
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
                                                    onClick={() => handleDelete(purpose.id)}
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

                {/* Create/Edit Modal */}
                <Modal
                    isOpen={isModalOpen}
                    onClose={() => setIsModalOpen(false)}
                    title={editingPurpose ? 'Edit Purpose' : 'Create New Purpose'}
                    description="Define the purpose for data collection and processing."
                >
                    <form id="purpose-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                        <Input
                            label="Purpose Name"
                            placeholder="e.g., Marketing Communications"
                            required
                            {...register('name')}
                            error={errors.name?.message}
                        />

                        <Textarea
                            label="Description"
                            placeholder="Describe why this data is being collected..."
                            required
                            {...register('description')}
                            error={errors.description?.message}
                        />

                        <div className="grid grid-cols-2 gap-4">
                            <Select
                                label="Legal Basis"
                                options={legalBasisOptions}
                                required
                                {...register('legalBasis')}
                                error={errors.legalBasis?.message}
                            />

                            <Input
                                label="Retention Period"
                                placeholder="e.g., 2 years"
                                required
                                {...register('retentionPeriod')}
                                error={errors.retentionPeriod?.message}
                            />
                        </div>

                        <Select
                            label="Status"
                            options={statusOptions}
                            required
                            {...register('status')}
                            error={errors.status?.message}
                        />

                        <div className="flex justify-end gap-3 pt-4">
                            <Button type="button" variant="secondary" onClick={() => setIsModalOpen(false)}>
                                Cancel
                            </Button>
                            <Button type="submit" variant="primary">
                                {editingPurpose ? 'Save Changes' : 'Create Purpose'}
                            </Button>
                        </div>
                    </form>
                </Modal>
            </div>
        </DashboardLayout>
    );
}
