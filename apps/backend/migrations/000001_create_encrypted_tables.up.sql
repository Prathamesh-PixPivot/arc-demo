-- Create extension for UUID generation if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create encrypted_consents table
CREATE TABLE encrypted_consents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    purposes JSONB,
    policy_snapshot JSONB,
    signature TEXT,
    geo_region VARCHAR(255),
    jurisdiction VARCHAR(255),
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for encrypted_consents
CREATE INDEX idx_encrypted_consents_user_id ON encrypted_consents(user_id);
CREATE INDEX idx_encrypted_consents_tenant_id ON encrypted_consents(tenant_id);

-- Create encrypted_breach_notifications table
CREATE TABLE encrypted_breach_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    description TEXT,
    breach_date TIMESTAMP WITH TIME ZONE NOT NULL,
    detection_date TIMESTAMP WITH TIME ZONE NOT NULL,
    notification_date TIMESTAMP WITH TIME ZONE,
    affected_users_count INTEGER DEFAULT 0,
    notified_users_count INTEGER DEFAULT 0,
    severity VARCHAR(20) DEFAULT 'medium',
    breach_type VARCHAR(50),
    status VARCHAR(50) DEFAULT 'investigating',
    requires_dpb_reporting BOOLEAN DEFAULT FALSE,
    dpb_reported BOOLEAN DEFAULT FALSE,
    dpb_reported_date TIMESTAMP WITH TIME ZONE,
    dpb_report_reference TEXT,
    remedial_actions JSONB,
    preventive_measures JSONB,
    investigation_summary TEXT,
    investigated_by TEXT,
    investigation_date TIMESTAMP WITH TIME ZONE,
    compliance_status VARCHAR(50) DEFAULT 'pending',
    compliance_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for encrypted_breach_notifications
CREATE INDEX idx_encrypted_breach_notifications_tenant_id ON encrypted_breach_notifications(tenant_id);
CREATE INDEX idx_encrypted_breach_notifications_breach_date ON encrypted_breach_notifications(breach_date);
CREATE INDEX idx_encrypted_breach_notifications_status ON encrypted_breach_notifications(status);
