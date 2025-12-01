-- Drop indexes
DROP INDEX IF EXISTS idx_consent_forms_last_published_at;
DROP INDEX IF EXISTS idx_consent_forms_status;
DROP INDEX IF EXISTS idx_consent_form_versions_published_at;
DROP INDEX IF EXISTS idx_consent_form_versions_status;
DROP INDEX IF EXISTS idx_consent_form_versions_consent_form_id;

-- Drop consent_form_versions table
DROP TABLE IF EXISTS consent_form_versions;

-- Remove versioning and status fields from consent_forms table
ALTER TABLE consent_forms 
DROP COLUMN IF EXISTS last_published_by,
DROP COLUMN IF EXISTS last_published_at,
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS current_version;
