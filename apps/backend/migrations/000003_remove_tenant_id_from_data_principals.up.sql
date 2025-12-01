-- Remove the tenant_id column from data_principals table
ALTER TABLE data_principals DROP COLUMN IF EXISTS tenant_id;
