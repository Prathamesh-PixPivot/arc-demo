-- Create consent_receipts table for DPDP-compliant receipt generation
CREATE TABLE IF NOT EXISTS consent_receipts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_consent_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    receipt_number VARCHAR(50) UNIQUE NOT NULL,
    pdf_path TEXT,
    qr_code_data TEXT,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    emailed_at TIMESTAMP WITH TIME ZONE,
    download_count INTEGER DEFAULT 0,
    is_valid BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_consent_receipts_user_consent_id ON consent_receipts(user_consent_id);
CREATE INDEX IF NOT EXISTS idx_consent_receipts_tenant_id ON consent_receipts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_consent_receipts_receipt_number ON consent_receipts(receipt_number);
CREATE INDEX IF NOT EXISTS idx_consent_receipts_generated_at ON consent_receipts(generated_at);
CREATE INDEX IF NOT EXISTS idx_consent_receipts_is_valid ON consent_receipts(is_valid);
CREATE INDEX IF NOT EXISTS idx_consent_receipts_expires_at ON consent_receipts(expires_at);

-- Add foreign key constraints
ALTER TABLE consent_receipts 
ADD CONSTRAINT fk_consent_receipts_user_consent 
FOREIGN KEY (user_consent_id) REFERENCES user_consents(id) ON DELETE CASCADE;

-- Add check constraints
ALTER TABLE consent_receipts 
ADD CONSTRAINT chk_receipt_number_format 
CHECK (receipt_number ~ '^RCP-[0-9]{8}-[A-F0-9]{6}$');

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_consent_receipts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_consent_receipts_updated_at
    BEFORE UPDATE ON consent_receipts
    FOR EACH ROW
    EXECUTE FUNCTION update_consent_receipts_updated_at();

-- Add comments for documentation
COMMENT ON TABLE consent_receipts IS 'DPDP-compliant consent receipts with PDF generation and verification';
COMMENT ON COLUMN consent_receipts.receipt_number IS 'Unique receipt number in format RCP-YYYYMMDD-XXXXXX';
COMMENT ON COLUMN consent_receipts.pdf_path IS 'Path to stored PDF file (local or S3)';
COMMENT ON COLUMN consent_receipts.qr_code_data IS 'QR code data for verification';
COMMENT ON COLUMN consent_receipts.download_count IS 'Number of times receipt has been downloaded';
COMMENT ON COLUMN consent_receipts.is_valid IS 'Whether the receipt is still valid';
COMMENT ON COLUMN consent_receipts.expires_at IS 'Optional expiration date for the receipt';
COMMENT ON COLUMN consent_receipts.metadata IS 'Additional receipt metadata in JSON format';
