export interface Purpose {
    id: string;
    name: string;
    description: string;
    legalBasis: 'consent' | 'contract' | 'legal_obligation' | 'vital_interests' | 'public_task' | 'legitimate_interests';
    retentionPeriod: string;
    status: 'active' | 'inactive' | 'draft';
    version: string;
    lastUpdated: string;
}

export interface ConsentForm {
    id: string;
    name: string;
    purposeIds: string[];
    version: string;
    status: 'published' | 'draft' | 'archived';
    lastUpdated: string;
}
