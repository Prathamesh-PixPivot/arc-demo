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

interface DSRRequest {
    id: string;
    userId: string;
    userName: string;
    type: 'access' | 'delete' | 'rectify' | 'portability';
    status: 'pending' | 'processing' | 'completed' | 'rejected';
    dateSubmitted: string;
    deadline: string;
    description: string;
}

export default function DSRPage() {
    const [requests, setRequests] = useState<DSRRequest[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedRequest, setSelectedRequest] = useState<DSRRequest | null>(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [responseNote, setResponseNote] = useState('');

    const fetchRequests = async () => {
        try {
            const response = await api.get('/api/v1/fiduciary/dsr');
            setRequests(response.data);
        } catch (error) {
            console.error('Failed to fetch DSR requests:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchRequests();
    }, []);

    const handleViewRequest = (request: DSRRequest) => {
        setSelectedRequest(request);
        setResponseNote('');
        setIsModalOpen(true);
    };

    const handleUpdateStatus = async (status: DSRRequest['status']) => {
        if (!selectedRequest) return;

        try {
            await api.post(`/api/v1/fiduciary/dsr/${selectedRequest.id}/${status}`, { note: responseNote });
            await fetchRequests();
            setIsModalOpen(false);
        } catch (error) {
            console.error('Failed to update DSR status:', error);
            alert('Failed to update request status. Please try again.');
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'pending': return 'warning';
            case 'processing': return 'info';
            case 'completed': return 'success';
            case 'rejected': return 'error';
            default: return 'gray';
        }
    };

    return (
        <DashboardLayout>
            <div className="space-y-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Data Subject Requests (DSR)</h1>
                    <p className="text-gray-600">Manage and respond to user data requests</p>
                </div>

                {/* Stats */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    {[
                        { label: 'Total Requests', value: requests.length, color: 'bg-purple-light text-purple' },
                        { label: 'Pending', value: requests.filter(r => r.status === 'pending').length, color: 'bg-warning-light text-warning' },
                        { label: 'Processing', value: requests.filter(r => r.status === 'processing').length, color: 'bg-info-light text-info' },
                        { label: 'Completed', value: requests.filter(r => r.status === 'completed').length, color: 'bg-success-light text-success' },
                    ].map((stat, idx) => (
                        <div key={idx} className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm">
                            <p className="text-sm font-medium text-gray-500">{stat.label}</p>
                            <p className={`mt-2 text-2xl font-bold ${stat.color.split(' ')[1]}`}>{stat.value}</p>
                        </div>
                    ))}
                </div>

                {/* Table */}
                <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Request ID</TableHead>
                                <TableHead>User</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Submitted</TableHead>
                                <TableHead>Deadline</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                                        Loading requests...
                                    </TableCell>
                                </TableRow>
                            ) : requests.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                                        No requests found.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                requests.map((request) => (
                                    <TableRow key={request.id}>
                                        <TableCell className="font-medium">{request.id}</TableCell>
                                        <TableCell>
                                            <div>
                                                <div className="font-medium text-gray-900">{request.userName}</div>
                                                <div className="text-xs text-gray-500">{request.userId}</div>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="purple" className="capitalize">
                                                {request.type}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant={getStatusColor(request.status)} className="capitalize">
                                                {request.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{new Date(request.dateSubmitted).toLocaleDateString()}</TableCell>
                                        <TableCell>
                                            <span className={new Date(request.deadline) < new Date() ? 'text-error font-medium' : ''}>
                                                {new Date(request.deadline).toLocaleDateString()}
                                            </span>
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <Button
                                                variant="secondary"
                                                size="sm"
                                                onClick={() => handleViewRequest(request)}
                                            >
                                                View Details
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
                    title={`Request Details: ${selectedRequest?.id}`}
                    description="Review request details and update status."
                >
                    {selectedRequest && (
                        <div className="space-y-6">
                            <div className="grid grid-cols-2 gap-4 text-sm">
                                <div>
                                    <p className="text-gray-500">User</p>
                                    <p className="font-medium">{selectedRequest.userName}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500">Type</p>
                                    <p className="font-medium capitalize">{selectedRequest.type}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500">Date Submitted</p>
                                    <p className="font-medium">{new Date(selectedRequest.dateSubmitted).toLocaleDateString()}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500">Deadline</p>
                                    <p className="font-medium text-error">{new Date(selectedRequest.deadline).toLocaleDateString()}</p>
                                </div>
                            </div>

                            <div className="bg-gray-50 p-4 rounded-lg">
                                <p className="text-sm font-medium text-gray-900 mb-1">Request Description</p>
                                <p className="text-sm text-gray-600">{selectedRequest.description}</p>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-900">Response Note (Internal)</label>
                                <Textarea
                                    placeholder="Add notes about the resolution..."
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
                                        onClick={() => handleUpdateStatus('processing')}
                                        disabled={selectedRequest.status === 'processing' || selectedRequest.status === 'completed'}
                                    >
                                        Processing
                                    </Button>
                                    <Button
                                        variant="primary"
                                        className="flex-1 bg-success hover:bg-success-hover border-success"
                                        onClick={() => handleUpdateStatus('completed')}
                                        disabled={selectedRequest.status === 'completed'}
                                    >
                                        Complete
                                    </Button>
                                    <Button
                                        variant="danger"
                                        className="flex-1"
                                        onClick={() => handleUpdateStatus('rejected')}
                                        disabled={selectedRequest.status === 'rejected' || selectedRequest.status === 'completed'}
                                    >
                                        Reject
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
