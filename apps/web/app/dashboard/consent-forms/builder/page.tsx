'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Purpose } from '@/lib/types';
import { api } from '@/lib/api';
import { useEffect } from 'react';

// Purposes will be fetched from API

export default function ConsentFormBuilder() {
    const router = useRouter();
    const [formName, setFormName] = useState('New Consent Form');
    const [description, setDescription] = useState('Please review and accept our privacy terms.');
    const [selectedPurposes, setSelectedPurposes] = useState<string[]>([]);
    const [previewMode, setPreviewMode] = useState<'desktop' | 'mobile'>('desktop');
    const [availablePurposes, setAvailablePurposes] = useState<Purpose[]>([]);

    useEffect(() => {
        const fetchPurposes = async () => {
            try {
                const response = await api.get('/api/v1/fiduciary/purposes');
                setAvailablePurposes(response.data);
            } catch (error) {
                console.error('Failed to fetch purposes:', error);
            }
        };
        fetchPurposes();
    }, []);

    const togglePurpose = (id: string) => {
        setSelectedPurposes((prev) =>
            prev.includes(id) ? prev.filter((p) => p !== id) : [...prev, id]
        );
    };

    const handleSave = () => {
        // TODO: Save form
        console.log('Saving form:', { formName, description, selectedPurposes });
        router.push('/dashboard/consent-forms');
    };

    return (
        <DashboardLayout>
            <div className="h-[calc(100vh-8rem)] flex flex-col">
                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                    <div>
                        <div className="flex items-center gap-2 text-sm text-gray-500 mb-1">
                            <button onClick={() => router.back()} className="hover:text-gray-900">
                                Consent Forms
                            </button>
                            <span>/</span>
                            <span>Builder</span>
                        </div>
                        <h1 className="text-2xl font-bold text-gray-900">Form Builder</h1>
                    </div>
                    <div className="flex gap-3">
                        <Button variant="secondary" onClick={() => router.back()}>
                            Cancel
                        </Button>
                        <Button variant="primary" onClick={handleSave}>
                            Save Form
                        </Button>
                    </div>
                </div>

                {/* Builder Interface */}
                <div className="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-6 min-h-0">
                    {/* Configuration Panel */}
                    <div className="lg:col-span-4 flex flex-col gap-6 overflow-y-auto pr-2">
                        <Card>
                            <div className="p-4 space-y-4">
                                <h3 className="font-semibold text-gray-900">General Settings</h3>
                                <Input
                                    label="Form Name"
                                    value={formName}
                                    onChange={(e) => setFormName(e.target.value)}
                                />
                                <Textarea
                                    label="Header Description"
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    rows={3}
                                />
                            </div>
                        </Card>

                        <Card>
                            <div className="p-4 space-y-4">
                                <div className="flex items-center justify-between">
                                    <h3 className="font-semibold text-gray-900">Select Purposes</h3>
                                    <Badge variant="info">{selectedPurposes.length} selected</Badge>
                                </div>
                                <div className="space-y-2">
                                    {availablePurposes.map((purpose) => (
                                        <div
                                            key={purpose.id}
                                            className={`p-3 rounded-lg border cursor-pointer transition-all ${selectedPurposes.includes(purpose.id)
                                                ? 'border-purple bg-purple-light/10 ring-1 ring-purple'
                                                : 'border-gray-200 hover:border-purple/50'
                                                }`}
                                            onClick={() => togglePurpose(purpose.id)}
                                        >
                                            <div className="flex items-start gap-3">
                                                <div
                                                    className={`mt-0.5 w-4 h-4 rounded border flex items-center justify-center ${selectedPurposes.includes(purpose.id)
                                                        ? 'bg-purple border-purple'
                                                        : 'border-gray-300 bg-white'
                                                        }`}
                                                >
                                                    {selectedPurposes.includes(purpose.id) && (
                                                        <svg
                                                            className="w-3 h-3 text-white"
                                                            fill="none"
                                                            viewBox="0 0 24 24"
                                                            strokeWidth={3}
                                                            stroke="currentColor"
                                                        >
                                                            <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                                                        </svg>
                                                    )}
                                                </div>
                                                <div>
                                                    <p className="text-sm font-medium text-gray-900">{purpose.name}</p>
                                                    <p className="text-xs text-gray-500 mt-0.5 line-clamp-2">
                                                        {purpose.description}
                                                    </p>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </Card>
                    </div>

                    {/* Preview Panel */}
                    <div className="lg:col-span-8 flex flex-col bg-gray-100 rounded-xl border border-gray-200 overflow-hidden">
                        <div className="p-3 border-b border-gray-200 bg-white flex items-center justify-between">
                            <span className="text-sm font-medium text-gray-600">Live Preview</span>
                            <div className="flex items-center gap-2 bg-gray-100 p-1 rounded-lg">
                                <button
                                    onClick={() => setPreviewMode('desktop')}
                                    className={`p-1.5 rounded ${previewMode === 'desktop' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-900'
                                        }`}
                                    title="Desktop View"
                                >
                                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m-9-12V15a2.25 2.25 0 002.25 2.25h9.5A2.25 2.25 0 0019 15V5.25m-9-12h9.5a2.25 2.25 0 012.25 2.25v9a2.25 2.25 0 01-2.25 2.25h-9.5a2.25 2.25 0 01-2.25-2.25v-9a2.25 2.25 0 012.25-2.25z" />
                                    </svg>
                                </button>
                                <button
                                    onClick={() => setPreviewMode('mobile')}
                                    className={`p-1.5 rounded ${previewMode === 'mobile' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-900'
                                        }`}
                                    title="Mobile View"
                                >
                                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3" />
                                    </svg>
                                </button>
                            </div>
                        </div>

                        <div className="flex-1 flex items-center justify-center p-8 overflow-y-auto">
                            <div
                                className={`bg-white shadow-2xl rounded-lg overflow-hidden transition-all duration-300 ${previewMode === 'mobile' ? 'w-[375px]' : 'w-full max-w-2xl'
                                    }`}
                            >
                                {/* Simulated Form Header */}
                                <div className="bg-purple p-6 text-white">
                                    <h2 className="text-xl font-bold">{formName || 'Form Title'}</h2>
                                    <p className="mt-2 text-purple-100 text-sm opacity-90">
                                        {description || 'Form description goes here...'}
                                    </p>
                                </div>

                                {/* Simulated Form Content */}
                                <div className="p-6 space-y-6">
                                    {selectedPurposes.length === 0 ? (
                                        <div className="text-center py-8 text-gray-400 border-2 border-dashed border-gray-200 rounded-lg">
                                            <p>Select purposes from the left panel to see them here</p>
                                        </div>
                                    ) : (
                                        selectedPurposes.map((id) => {
                                            const purpose = availablePurposes.find((p) => p.id === id);
                                            if (!purpose) return null;
                                            return (
                                                <div key={id} className="flex items-start gap-3 pb-4 border-b border-gray-100 last:border-0">
                                                    <div className="mt-1">
                                                        <input
                                                            type="checkbox"
                                                            className="w-4 h-4 text-purple border-gray-300 rounded focus:ring-purple"
                                                            defaultChecked
                                                        />
                                                    </div>
                                                    <div>
                                                        <p className="font-medium text-gray-900">{purpose.name}</p>
                                                        <p className="text-sm text-gray-600 mt-1">{purpose.description}</p>
                                                    </div>
                                                </div>
                                            );
                                        })
                                    )}

                                    <div className="pt-4">
                                        <button className="w-full bg-purple text-white py-2.5 rounded-lg font-medium hover:bg-purple-hover transition-colors">
                                            Accept Selected
                                        </button>
                                        <button className="w-full mt-3 text-gray-600 py-2.5 rounded-lg font-medium hover:bg-gray-50 transition-colors">
                                            Reject All
                                        </button>
                                    </div>

                                    <div className="text-center text-xs text-gray-400">
                                        Powered by Arc Privacy
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </DashboardLayout>
    );
}
