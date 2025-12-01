-- =====================================================
-- ROLLBACK: Remove BSFI Industry Seed Data
-- =====================================================

-- Remove in reverse order of dependencies

-- 14. Remove language translations
DELETE FROM language_translations WHERE id IN (
    'lt-en-001', 'lt-en-002', 'lt-en-003', 'lt-en-004',
    'lt-hi-001', 'lt-hi-002', 'lt-hi-003', 'lt-hi-004',
    'lt-mr-001', 'lt-mr-002',
    'lt-ta-001', 'lt-ta-002'
);

-- 13. Remove API keys
DELETE FROM api_keys WHERE id IN ('api-001', 'api-002', 'api-003', 'api-004');

-- 12. Remove admin users
DELETE FROM admin_users WHERE id IN ('admin-001', 'admin-002', 'admin-003', 'dpo-001', 'dpo-002', 'dpo-003');

-- 11. Remove indexes (if they were created by this migration)
DROP INDEX IF EXISTS idx_purposes_tenant_category;
DROP INDEX IF EXISTS idx_purposes_legal_basis;
DROP INDEX IF EXISTS idx_data_objects_tenant_category;
DROP INDEX IF EXISTS idx_consent_forms_tenant_status;
DROP INDEX IF EXISTS idx_webhooks_tenant_active;
DROP INDEX IF EXISTS idx_notification_templates_tenant_type;

-- 10. Remove audit configs
DELETE FROM audit_configs WHERE id IN ('ac-001', 'ac-002', 'ac-003');

-- 9. Remove compliance configs
DELETE FROM compliance_configs WHERE id IN ('cc-001', 'cc-002', 'cc-003');

-- 8. Remove notification templates
DELETE FROM notification_templates WHERE id IN ('nt-001', 'nt-002', 'nt-003', 'nt-004');

-- 7. Remove webhooks
DELETE FROM webhooks WHERE id IN ('wh-001', 'wh-002', 'wh-003');

-- 6. Remove consent forms
DELETE FROM consent_forms WHERE id IN (
    'cf-bank-001', 'cf-bank-002', 'cf-bank-003',
    'cf-ins-001', 'cf-ins-002',
    'cf-mf-001', 'cf-mf-002'
);

-- 5. Remove purpose-data object mappings
DELETE FROM purpose_data_objects WHERE purpose_id IN (
    'pur-bank-001', 'pur-bank-002', 'pur-bank-003', 'pur-bank-004', 'pur-bank-005', 'pur-bank-006',
    'pur-ins-001', 'pur-ins-002', 'pur-ins-003', 'pur-ins-004', 'pur-ins-005',
    'pur-mf-001', 'pur-mf-002', 'pur-mf-003', 'pur-mf-004', 'pur-mf-005'
);

-- 4. Remove data objects
DELETE FROM data_objects WHERE id IN (
    'do-001', 'do-002', 'do-003', 'do-004', 'do-005',
    'do-006', 'do-007', 'do-008', 'do-009', 'do-010'
);

-- 3. Remove purposes
DELETE FROM purposes WHERE id IN (
    'pur-bank-001', 'pur-bank-002', 'pur-bank-003', 'pur-bank-004', 'pur-bank-005', 'pur-bank-006',
    'pur-ins-001', 'pur-ins-002', 'pur-ins-003', 'pur-ins-004', 'pur-ins-005',
    'pur-mf-001', 'pur-mf-002', 'pur-mf-003', 'pur-mf-004', 'pur-mf-005'
);

-- 2. Remove purpose categories
DELETE FROM purpose_categories WHERE id IN (
    'cat-001', 'cat-002', 'cat-003', 'cat-004',
    'cat-005', 'cat-006', 'cat-007', 'cat-008'
);

-- 1. Remove organizations
DELETE FROM organizations WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);

-- =====================================================
-- END OF ROLLBACK
-- =====================================================
