-- Add the tenant_id column back to data_principals table
ALTER TABLE data_principals 
ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id) ON DELETE SET NULL;
