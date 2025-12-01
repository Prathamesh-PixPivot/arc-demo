-- Add versioning and status fields to consent_forms table
ALTER TABLE consent_forms 
ADD COLUMN IF NOT EXISTS current_version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'draft',
ADD COLUMN IF NOT EXISTS last_published_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_published_by UUID;

-- Create consent_form_versions table
CREATE TABLE IF NOT EXISTS consent_form_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consent_form_id UUID NOT NULL REFERENCES consent_forms(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    snapshot JSONB NOT NULL,
    published_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_by UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    change_log TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_form_version UNIQUE (consent_form_id, version_number)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_consent_form_versions_consent_form_id ON consent_form_versions(consent_form_id);
CREATE INDEX IF NOT EXISTS idx_consent_form_versions_status ON consent_form_versions(status);
CREATE INDEX IF NOT EXISTS idx_consent_form_versions_published_at ON consent_form_versions(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_consent_forms_status ON consent_forms(status);
CREATE INDEX IF NOT EXISTS idx_consent_forms_last_published_at ON consent_forms(last_published_at DESC);

-- Add comments for documentation
COMMENT ON COLUMN consent_forms.current_version IS 'Current version number of the consent form';
COMMENT ON COLUMN consent_forms.status IS 'Status of the consent form: draft, review, published, archived';
COMMENT ON COLUMN consent_forms.last_published_at IS 'Timestamp when the form was last published';
COMMENT ON COLUMN consent_forms.last_published_by IS 'UUID of the fiduciary user who last published the form';

COMMENT ON TABLE consent_form_versions IS 'Version history and snapshots of consent forms';
COMMENT ON COLUMN consent_form_versions.snapshot IS 'Complete snapshot of the consent form state at publication time';
COMMENT ON COLUMN consent_form_versions.change_log IS 'Description of changes made in this version';
