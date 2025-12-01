-- =====================================================
-- Rollback: Remove Security and Platform Tables
-- =====================================================

-- Drop RLS policies first
DROP POLICY IF EXISTS tenant_isolation_security_configurations ON security_configurations;
DROP POLICY IF EXISTS tenant_isolation_oauth_clients ON oauth_clients;
DROP POLICY IF EXISTS tenant_isolation_security_audit_logs ON security_audit_logs;
DROP POLICY IF EXISTS tenant_isolation_public_consent_submissions ON public_consent_submissions;
DROP POLICY IF EXISTS tenant_isolation_websites ON websites;
DROP POLICY IF EXISTS tenant_isolation_api_keys ON api_keys;
DROP POLICY IF EXISTS tenant_isolation_platform_sessions ON platform_sessions;

-- Drop triggers
DROP TRIGGER IF EXISTS update_security_configurations_updated_at ON security_configurations;
DROP TRIGGER IF EXISTS update_oauth_clients_updated_at ON oauth_clients;
DROP TRIGGER IF EXISTS update_websites_updated_at ON websites;
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS update_platform_sessions_updated_at ON platform_sessions;

-- Drop functions
DROP FUNCTION IF EXISTS cleanup_expired_oauth_tokens();
DROP FUNCTION IF EXISTS cleanup_expired_rate_limits();
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS security_configurations;
DROP TABLE IF EXISTS oauth_tokens;
DROP TABLE IF EXISTS oauth_clients;
DROP TABLE IF EXISTS rate_limit_entries;
DROP TABLE IF EXISTS security_audit_logs;
DROP TABLE IF EXISTS public_consent_submissions;
DROP TABLE IF EXISTS websites;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS platform_sessions;

-- =====================================================
-- END OF ROLLBACK
-- =====================================================
