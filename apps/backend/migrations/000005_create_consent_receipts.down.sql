-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_consent_receipts_updated_at ON consent_receipts;
DROP FUNCTION IF EXISTS update_consent_receipts_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_consent_receipts_user_consent_id;
DROP INDEX IF EXISTS idx_consent_receipts_tenant_id;
DROP INDEX IF EXISTS idx_consent_receipts_receipt_number;
DROP INDEX IF EXISTS idx_consent_receipts_generated_at;
DROP INDEX IF EXISTS idx_consent_receipts_is_valid;
DROP INDEX IF EXISTS idx_consent_receipts_expires_at;

-- Drop table
DROP TABLE IF EXISTS consent_receipts;
