-- =====================================================
-- Create Security and Platform Tables
-- Enhanced Security Architecture Implementation
-- =====================================================

-- 1. Platform Sessions Table
CREATE TABLE IF NOT EXISTS platform_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    platform VARCHAR(20) NOT NULL CHECK (platform IN ('web', 'desktop')),
    session_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for platform_sessions
CREATE INDEX idx_platform_sessions_user_id ON platform_sessions(user_id);
CREATE INDEX idx_platform_sessions_platform ON platform_sessions(platform);
CREATE INDEX idx_platform_sessions_expires_at ON platform_sessions(expires_at);
CREATE INDEX idx_platform_sessions_token ON platform_sessions(session_token);

-- Add comments
COMMENT ON TABLE platform_sessions IS 'User sessions per platform (web/desktop)';
COMMENT ON COLUMN platform_sessions.platform IS 'Platform type: web or desktop';
COMMENT ON COLUMN platform_sessions.session_token IS 'Unique session token for authentication';

-- 2. API Keys Table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    scopes JSONB,
    rate_limit INTEGER DEFAULT 1000,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for api_keys
CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX idx_api_keys_active ON api_keys(is_active) WHERE is_active = true;

-- Add comments
COMMENT ON TABLE api_keys IS 'API keys for external integrations';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of the API key (never store plain text)';
COMMENT ON COLUMN api_keys.scopes IS 'JSON array of allowed scopes/permissions';

-- 3. Websites Table
CREATE TABLE IF NOT EXISTS websites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for websites
CREATE INDEX idx_websites_tenant_id ON websites(tenant_id);
CREATE INDEX idx_websites_domain ON websites(domain);
CREATE INDEX idx_websites_active ON websites(is_active) WHERE is_active = true;

-- Add comments
COMMENT ON TABLE websites IS 'Client websites that integrate with consent manager';
COMMENT ON COLUMN websites.domain IS 'Website domain for CORS and validation';

-- 4. Public Consent Submissions Table
CREATE TABLE IF NOT EXISTS public_consent_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    website_id UUID REFERENCES websites(id),
    visitor_id VARCHAR(255),
    consent_data JSONB NOT NULL,
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for public_consent_submissions
CREATE INDEX idx_public_consent_submissions_tenant_id ON public_consent_submissions(tenant_id);
CREATE INDEX idx_public_consent_submissions_website_id ON public_consent_submissions(website_id);
CREATE INDEX idx_public_consent_submissions_visitor_id ON public_consent_submissions(visitor_id);
CREATE INDEX idx_public_consent_submissions_submitted_at ON public_consent_submissions(submitted_at);

-- Add comments
COMMENT ON TABLE public_consent_submissions IS 'Consent submissions from public websites';
COMMENT ON COLUMN public_consent_submissions.visitor_id IS 'Anonymous visitor identifier';
COMMENT ON COLUMN public_consent_submissions.consent_data IS 'JSON data of consent submission';

-- 5. Security Audit Log Table
CREATE TABLE IF NOT EXISTS security_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    user_id UUID,
    session_id UUID REFERENCES platform_sessions(id),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    ip_address INET,
    user_agent TEXT,
    resource VARCHAR(255),
    action VARCHAR(100),
    result VARCHAR(50),
    details JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for security_audit_logs
CREATE INDEX idx_security_audit_logs_tenant_id ON security_audit_logs(tenant_id);
CREATE INDEX idx_security_audit_logs_user_id ON security_audit_logs(user_id);
CREATE INDEX idx_security_audit_logs_event_type ON security_audit_logs(event_type);
CREATE INDEX idx_security_audit_logs_severity ON security_audit_logs(severity);
CREATE INDEX idx_security_audit_logs_timestamp ON security_audit_logs(timestamp);
CREATE INDEX idx_security_audit_logs_ip_address ON security_audit_logs(ip_address);

-- Add comments
COMMENT ON TABLE security_audit_logs IS 'Security-related audit events';
COMMENT ON COLUMN security_audit_logs.event_type IS 'Type of security event (login, access_denied, etc.)';
COMMENT ON COLUMN security_audit_logs.severity IS 'Event severity level';

-- 6. Rate Limit Entries Table
CREATE TABLE IF NOT EXISTS rate_limit_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identifier VARCHAR(255) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    count INTEGER DEFAULT 1,
    window_start TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

-- Add indexes for rate_limit_entries
CREATE INDEX idx_rate_limit_entries_identifier ON rate_limit_entries(identifier);
CREATE INDEX idx_rate_limit_entries_resource ON rate_limit_entries(resource);
CREATE INDEX idx_rate_limit_entries_window_start ON rate_limit_entries(window_start);
CREATE INDEX idx_rate_limit_entries_expires_at ON rate_limit_entries(expires_at);
CREATE UNIQUE INDEX idx_rate_limit_entries_unique ON rate_limit_entries(identifier, resource, window_start);

-- Add comments
COMMENT ON TABLE rate_limit_entries IS 'Rate limiting tracking data';
COMMENT ON COLUMN rate_limit_entries.identifier IS 'IP address, user ID, or API key identifier';

-- 7. OAuth Clients Table
CREATE TABLE IF NOT EXISTS oauth_clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uris JSONB,
    scopes JSONB,
    grant_types JSONB,
    is_public BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for oauth_clients
CREATE INDEX idx_oauth_clients_tenant_id ON oauth_clients(tenant_id);
CREATE INDEX idx_oauth_clients_client_id ON oauth_clients(client_id);
CREATE INDEX idx_oauth_clients_active ON oauth_clients(is_active) WHERE is_active = true;

-- Add comments
COMMENT ON TABLE oauth_clients IS 'OAuth 2.0 client applications';
COMMENT ON COLUMN oauth_clients.client_secret IS 'Hashed client secret';
COMMENT ON COLUMN oauth_clients.is_public IS 'Whether client can store secrets securely';

-- 8. OAuth Tokens Table
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id VARCHAR(255) NOT NULL REFERENCES oauth_clients(client_id),
    user_id UUID,
    token_type VARCHAR(50) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    scopes JSONB,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP
);

-- Add indexes for oauth_tokens
CREATE INDEX idx_oauth_tokens_client_id ON oauth_tokens(client_id);
CREATE INDEX idx_oauth_tokens_user_id ON oauth_tokens(user_id);
CREATE INDEX idx_oauth_tokens_expires_at ON oauth_tokens(expires_at);
CREATE INDEX idx_oauth_tokens_access_token ON oauth_tokens(access_token);

-- Add comments
COMMENT ON TABLE oauth_tokens IS 'OAuth 2.0 access and refresh tokens';
COMMENT ON COLUMN oauth_tokens.access_token IS 'Hashed access token';
COMMENT ON COLUMN oauth_tokens.refresh_token IS 'Hashed refresh token';

-- 9. Security Configuration Table
CREATE TABLE IF NOT EXISTS security_configurations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID UNIQUE NOT NULL,
    password_policy JSONB,
    session_timeout INTEGER DEFAULT 3600,
    max_login_attempts INTEGER DEFAULT 5,
    lockout_duration INTEGER DEFAULT 900,
    require_mfa BOOLEAN DEFAULT false,
    allowed_ip_ranges JSONB,
    blocked_ip_ranges JSONB,
    enable_request_signing BOOLEAN DEFAULT false,
    enable_audit_logging BOOLEAN DEFAULT true,
    data_retention_days INTEGER DEFAULT 2555,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for security_configurations
CREATE INDEX idx_security_configurations_tenant_id ON security_configurations(tenant_id);

-- Add comments
COMMENT ON TABLE security_configurations IS 'Tenant-specific security settings';
COMMENT ON COLUMN security_configurations.session_timeout IS 'Session timeout in seconds';
COMMENT ON COLUMN security_configurations.data_retention_days IS 'Data retention period (default 7 years for DPDPA)';

-- 10. Create triggers for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers
CREATE TRIGGER update_platform_sessions_updated_at BEFORE UPDATE ON platform_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_websites_updated_at BEFORE UPDATE ON websites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_oauth_clients_updated_at BEFORE UPDATE ON oauth_clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_security_configurations_updated_at BEFORE UPDATE ON security_configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 11. Create cleanup functions for expired data
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM platform_sessions WHERE expires_at < CURRENT_TIMESTAMP;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_expired_rate_limits()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM rate_limit_entries WHERE expires_at < CURRENT_TIMESTAMP;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_expired_oauth_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM oauth_tokens WHERE expires_at < CURRENT_TIMESTAMP AND revoked_at IS NULL;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 12. Insert default security configurations for existing tenants
INSERT INTO security_configurations (tenant_id, password_policy, session_timeout, max_login_attempts, lockout_duration)
SELECT 
    id,
    '{"min_length": 8, "require_uppercase": true, "require_lowercase": true, "require_numbers": true, "require_special": true}'::jsonb,
    3600,
    5,
    900
FROM organizations 
WHERE id NOT IN (SELECT tenant_id FROM security_configurations)
ON CONFLICT (tenant_id) DO NOTHING;

-- 13. Create indexes for performance optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_security_audit_logs_composite 
ON security_audit_logs(tenant_id, event_type, timestamp DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_platform_sessions_composite 
ON platform_sessions(user_id, platform, expires_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_keys_composite 
ON api_keys(tenant_id, is_active, expires_at);

-- 14. Add row-level security policies
ALTER TABLE platform_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE websites ENABLE ROW LEVEL SECURITY;
ALTER TABLE public_consent_submissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE oauth_clients ENABLE ROW LEVEL SECURITY;
ALTER TABLE oauth_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_configurations ENABLE ROW LEVEL SECURITY;

-- Create RLS policies (example for tenant isolation)
CREATE POLICY tenant_isolation_platform_sessions ON platform_sessions
    USING (user_id IN (SELECT id FROM users WHERE tenant_id = current_setting('app.current_tenant_id')::uuid));

CREATE POLICY tenant_isolation_api_keys ON api_keys
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE POLICY tenant_isolation_websites ON websites
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE POLICY tenant_isolation_public_consent_submissions ON public_consent_submissions
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE POLICY tenant_isolation_security_audit_logs ON security_audit_logs
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE POLICY tenant_isolation_oauth_clients ON oauth_clients
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE POLICY tenant_isolation_security_configurations ON security_configurations
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- =====================================================
-- END OF MIGRATION
-- =====================================================
