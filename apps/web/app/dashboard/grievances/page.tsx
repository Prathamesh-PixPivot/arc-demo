'use client';

import { useState, useEffect } from 'react';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Textarea } from '@/components/ui/Textarea';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/Table';
import { api } from '@/lib/api';

interface Grievance {
    id: string;
    subjectId: string;
    type: string;
    status: 'pending' | 'investigating' | 'resolved' | 'closed';
    dateSubmitted: string;
    dateResolved?: string;
    description: string;
}

export default function GrievancesPage() {
    const [grievances, setGrievances] = useState<Grievance[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedGrievance, setSelectedGrievance] = useState<Grievance | null>(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [responseNote, setResponseNote] = useState('');

    const fetchGrievances = async () => {
        try {
            const response = await api.get('/api/v1/fiduciary/grievances');
            setGrievances(response.data);
        } catch (error) {
            console.error('Failed to fetch grievances:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchGrievances();
    }, []);

    const handleViewGrievance = (grievance: Grievance) => {
        setSelectedGrievance(grievance);
        setResponseNote('');
        setIsModalOpen(true);
    };

    const handleUpdateStatus = async (status: Grievance['status']) => {
        if (!selectedGrievance) return;

        try {
            await api.put(`/api/v1/fiduciary/grievances/${selectedGrievance.id}`, { status, resolution: responseNote });
            await fetchGrievances();
            setIsModalOpen(false);
        } catch (error) {
            console.error('Failed to update grievance status:', error);
            alert('Failed to update grievance status. Please try again.');
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'pending': return 'error';
            case 'investigating': return 'warning';
            case 'resolved': return 'success';
            case 'closed': return 'gray';
            default: return 'gray';
        }
    };

    return (
        <DashboardLayout>
            <div className="space-y-6">
                <div className="flex justify-between items-center">
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900">Grievance Redressal</h1>
                        <p className="text-gray-600">Manage and resolve user complaints</p>
                    </div>
                </div>

                {/* Table */}
                <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>Subject ID</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Date Submitted</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                                        Loading grievances...
                                    </TableCell>
                                </TableRow>
                            ) : grievances.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                                        No grievances found.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                grievances.map((grievance) => (
                                    <TableRow key={grievance.id}>
                                        <TableCell className="font-medium">{grievance.id}</TableCell>
                                        <TableCell>{grievance.subjectId}</TableCell>
                                        <TableCell className="capitalize">{grievance.type.replace('_', ' ')}</TableCell>
                                        <TableCell>
                                            <Badge variant={getStatusColor(grievance.status)} className="capitalize">
                                                {grievance.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{new Date(grievance.dateSubmitted).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <Button
                                                variant="secondary"
                                                size="sm"
                                                onClick={() => handleViewGrievance(grievance)}
                                            >
                                                Manage
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>

                {/* Details Modal */}
                <Modal
                    isOpen={isModalOpen}
                    onClose={() => setIsModalOpen(false)}
                    title={`Grievance: ${selectedGrievance?.id}`}
                    description="Review details and update resolution status."
                >
                    {selectedGrievance && (
                        <div className="space-y-6">
                            <div className="grid grid-cols-2 gap-4 text-sm bg-gray-50 p-4 rounded-lg">
                                <div>
                                    <p className="text-gray-500">Subject ID</p>
                                    <p className="font-medium">{selectedGrievance.subjectId}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500">Date Submitted</p>
                                    <p className="font-medium">{new Date(selectedGrievance.dateSubmitted).toLocaleDateString()}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500">Type</p>
                                    <p className="font-medium capitalize">{selectedGrievance.type.replace('_', ' ')}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500">Current Status</p>
                                    <Badge variant={getStatusColor(selectedGrievance.status)} size="sm" className="mt-1 capitalize">
                                        {selectedGrievance.status}
                                    </Badge>
                                </div>
                            </div>

                            <div>
                                <p className="text-sm font-medium text-gray-900 mb-2">Description</p>
                                <div className="bg-white border border-gray-200 p-3 rounded-md text-sm text-gray-700">
                                    {selectedGrievance.description}
                                </div>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-900">Resolution Notes</label>
                                <Textarea
                                    placeholder="Enter details about the investigation or resolution..."
                                    value={responseNote}
                                    onChange={(e) => setResponseNote(e.target.value)}
                                    rows={3}
                                />
                            </div>

                            <div className="flex flex-col gap-3 pt-4 border-t border-gray-100">
                                <p className="text-sm font-medium text-gray-900">Update Status</p>
                                <div className="flex gap-2">
                                    <Button
                                        variant="secondary"
                                        className="flex-1"
                                        onClick={() => handleUpdateStatus('investigating')}
                                        disabled={selectedGrievance.status === 'investigating' || selectedGrievance.status === 'resolved'}
                                    >
                                        Investigate
                                    </Button>
                                    <Button
                                        variant="primary"
                                        className="flex-1 bg-success hover:bg-success-hover border-success"
                                        onClick={() => handleUpdateStatus('resolved')}
                                        disabled={selectedGrievance.status === 'resolved'}
                                    >
                                        Resolve
                                    </Button>
                                    <Button
                                        variant="secondary"
                                        className="flex-1"
                                        onClick={() => handleUpdateStatus('closed')}
                                        disabled={selectedGrievance.status === 'closed'}
                                    >
                                        Close
                                    </Button>
                                </div>
                            </div>
                        </div>
                    )}
                </Modal>
            </div>
        </DashboardLayout>
    );
}
