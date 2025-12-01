-- =====================================================
-- BSFI Industry Seed Data for DPDPA 2023 Compliance
-- =====================================================
-- This migration seeds comprehensive data for Banking, Financial Services, and Insurance (BSFI) industry
-- Fully compliant with Digital Personal Data Protection Act (DPDPA) 2023 of India

-- =====================================================
-- 1. SEED ORGANIZATIONS (BSFI Companies)
-- =====================================================

-- Sample BSFI organizations (can be customized per deployment)
INSERT INTO organizations (id, name, domain, industry, registration_number, dpo_email, dpo_phone, address, city, state, country, postal_code, is_active, created_at, updated_at)
VALUES 
    ('11111111-1111-1111-1111-111111111111', 'SecureBank India Ltd', 'securebank.in', 'Banking', 'CIN:U65110MH2020PLC123456', 'dpo@securebank.in', '+91-22-6789-0000', 'One BKC, Bandra Kurla Complex', 'Mumbai', 'Maharashtra', 'India', '400051', true, NOW(), NOW()),
    ('22222222-2222-2222-2222-222222222222', 'TrustInsure Life', 'trustinsure.co.in', 'Insurance', 'CIN:U66010DL2020PLC234567', 'privacy@trustinsure.co.in', '+91-11-4567-8900', 'Connaught Place, Central Delhi', 'New Delhi', 'Delhi', 'India', '110001', true, NOW(), NOW()),
    ('33333333-3333-3333-3333-333333333333', 'WealthGrow Mutual Fund', 'wealthgrow.in', 'Financial Services', 'CIN:U67120KA2020PLC345678', 'compliance@wealthgrow.in', '+91-80-2345-6789', 'UB City, Vittal Mallya Road', 'Bengaluru', 'Karnataka', 'India', '560001', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 2. SEED PURPOSE CATEGORIES (DPDPA 2023 Compliant)
-- =====================================================

INSERT INTO purpose_categories (id, name, description, regulatory_reference, created_at, updated_at)
VALUES
    ('cat-001', 'Account Management', 'Managing customer accounts and providing banking services', 'DPDPA 2023 - Section 7(1)', NOW(), NOW()),
    ('cat-002', 'Regulatory Compliance', 'Compliance with RBI, SEBI, IRDAI regulations', 'DPDPA 2023 - Section 7(2)', NOW(), NOW()),
    ('cat-003', 'Risk Assessment', 'Credit scoring, underwriting, and risk management', 'DPDPA 2023 - Section 7(3)', NOW(), NOW()),
    ('cat-004', 'Marketing & Communication', 'Product offers and promotional communications', 'DPDPA 2023 - Section 7(4)', NOW(), NOW()),
    ('cat-005', 'Fraud Prevention', 'Security and fraud detection measures', 'DPDPA 2023 - Section 7(5)', NOW(), NOW()),
    ('cat-006', 'Customer Service', 'Support and grievance redressal', 'DPDPA 2023 - Section 7(6)', NOW(), NOW()),
    ('cat-007', 'Analytics & Insights', 'Business intelligence and service improvement', 'DPDPA 2023 - Section 7(7)', NOW(), NOW()),
    ('cat-008', 'Third Party Sharing', 'Sharing with partners and service providers', 'DPDPA 2023 - Section 7(8)', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 3. SEED PURPOSES (BSFI Specific - DPDPA Compliant)
-- =====================================================

INSERT INTO purposes (id, tenant_id, name, description, legal_basis, retention_period, category_id, is_special_category, requires_explicit_consent, created_at, updated_at)
VALUES
    -- Banking Purposes
    ('pur-bank-001', '11111111-1111-1111-1111-111111111111', 'Account Opening and KYC', 'Collection and verification of identity documents for account opening as per RBI guidelines', 'legal_obligation', 2555, 'cat-001', false, false, NOW(), NOW()),
    ('pur-bank-002', '11111111-1111-1111-1111-111111111111', 'Transaction Processing', 'Processing deposits, withdrawals, transfers, and payment transactions', 'contract', 2555, 'cat-001', false, false, NOW(), NOW()),
    ('pur-bank-003', '11111111-1111-1111-1111-111111111111', 'Credit Assessment', 'Evaluating creditworthiness for loans and credit facilities', 'legitimate_interest', 2555, 'cat-003', true, true, NOW(), NOW()),
    ('pur-bank-004', '11111111-1111-1111-1111-111111111111', 'Regulatory Reporting', 'Reporting to RBI, Income Tax, and other regulatory authorities', 'legal_obligation', 2555, 'cat-002', false, false, NOW(), NOW()),
    ('pur-bank-005', '11111111-1111-1111-1111-111111111111', 'Marketing Communications', 'Sending information about new products, offers, and services', 'consent', 365, 'cat-004', false, true, NOW(), NOW()),
    ('pur-bank-006', '11111111-1111-1111-1111-111111111111', 'Fraud Monitoring', 'Real-time transaction monitoring for suspicious activities', 'legitimate_interest', 1825, 'cat-005', false, false, NOW(), NOW()),
    
    -- Insurance Purposes
    ('pur-ins-001', '22222222-2222-2222-2222-222222222222', 'Policy Underwriting', 'Assessing risk and determining premium for insurance policies', 'contract', 3650, 'cat-003', true, true, NOW(), NOW()),
    ('pur-ins-002', '22222222-2222-2222-2222-222222222222', 'Claims Processing', 'Processing and settling insurance claims', 'contract', 3650, 'cat-001', true, false, NOW(), NOW()),
    ('pur-ins-003', '22222222-2222-2222-2222-222222222222', 'Health Records Management', 'Managing medical records for health and life insurance', 'consent', 3650, 'cat-001', true, true, NOW(), NOW()),
    ('pur-ins-004', '22222222-2222-2222-2222-222222222222', 'Reinsurance', 'Sharing data with reinsurance partners', 'legitimate_interest', 3650, 'cat-008', false, true, NOW(), NOW()),
    ('pur-ins-005', '22222222-2222-2222-2222-222222222222', 'IRDAI Compliance', 'Regulatory reporting to Insurance Regulatory and Development Authority', 'legal_obligation', 3650, 'cat-002', false, false, NOW(), NOW()),
    
    -- Mutual Fund Purposes
    ('pur-mf-001', '33333333-3333-3333-3333-333333333333', 'Investment Account Management', 'Managing mutual fund investments and portfolios', 'contract', 2555, 'cat-001', false, false, NOW(), NOW()),
    ('pur-mf-002', '33333333-3333-3333-3333-333333333333', 'SEBI Compliance', 'Regulatory reporting to Securities and Exchange Board of India', 'legal_obligation', 2555, 'cat-002', false, false, NOW(), NOW()),
    ('pur-mf-003', '33333333-3333-3333-3333-333333333333', 'Risk Profiling', 'Assessing investor risk appetite and suitability', 'contract', 1825, 'cat-003', false, true, NOW(), NOW()),
    ('pur-mf-004', '33333333-3333-3333-3333-333333333333', 'Performance Analytics', 'Analyzing fund performance and generating reports', 'legitimate_interest', 1095, 'cat-007', false, false, NOW(), NOW()),
    ('pur-mf-005', '33333333-3333-3333-3333-333333333333', 'Tax Reporting', 'Generating tax statements and reporting to tax authorities', 'legal_obligation', 2555, 'cat-002', false, false, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 4. SEED DATA OBJECTS (BSFI Specific)
-- =====================================================

INSERT INTO data_objects (id, tenant_id, name, description, category, fields, is_sensitive, created_at, updated_at)
VALUES
    -- Personal Identification
    ('do-001', '11111111-1111-1111-1111-111111111111', 'Basic Identity Information', 'Name, date of birth, gender', 'personal_identification', '["full_name", "date_of_birth", "gender", "nationality"]', false, NOW(), NOW()),
    ('do-002', '11111111-1111-1111-1111-111111111111', 'Government IDs', 'PAN, Aadhaar, Passport, Voter ID', 'personal_identification', '["pan_number", "aadhaar_number", "passport_number", "voter_id"]', true, NOW(), NOW()),
    ('do-003', '11111111-1111-1111-1111-111111111111', 'Contact Information', 'Address, phone, email', 'contact_information', '["residential_address", "office_address", "mobile_number", "email_address"]', false, NOW(), NOW()),
    
    -- Financial Information
    ('do-004', '11111111-1111-1111-1111-111111111111', 'Income Details', 'Salary, business income, other sources', 'financial_information', '["annual_income", "income_source", "employer_name", "business_details"]', true, NOW(), NOW()),
    ('do-005', '11111111-1111-1111-1111-111111111111', 'Bank Account Details', 'Account numbers and transaction history', 'financial_information', '["account_number", "ifsc_code", "account_type", "balance"]', true, NOW(), NOW()),
    ('do-006', '11111111-1111-1111-1111-111111111111', 'Credit Information', 'Credit score, loan history, defaults', 'financial_information', '["credit_score", "existing_loans", "repayment_history", "defaults"]', true, NOW(), NOW()),
    
    -- Insurance Specific
    ('do-007', '22222222-2222-2222-2222-222222222222', 'Medical Records', 'Health conditions, medical history', 'health_information', '["medical_conditions", "medications", "hospitalizations", "family_history"]', true, NOW(), NOW()),
    ('do-008', '22222222-2222-2222-2222-222222222222', 'Nominee Information', 'Beneficiary details for policies', 'personal_identification', '["nominee_name", "nominee_relationship", "nominee_contact", "nominee_id"]', true, NOW(), NOW()),
    
    -- Investment Specific
    ('do-009', '33333333-3333-3333-3333-333333333333', 'Investment Profile', 'Risk tolerance, investment goals', 'financial_information', '["risk_appetite", "investment_horizon", "financial_goals", "investment_experience"]', false, NOW(), NOW()),
    ('do-010', '33333333-3333-3333-3333-333333333333', 'Portfolio Holdings', 'Current investments and assets', 'financial_information', '["mutual_funds", "stocks", "bonds", "other_assets"]', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 5. SEED PURPOSE-DATA OBJECT MAPPINGS
-- =====================================================

INSERT INTO purpose_data_objects (purpose_id, data_object_id)
VALUES
    -- Banking mappings
    ('pur-bank-001', 'do-001'), ('pur-bank-001', 'do-002'), ('pur-bank-001', 'do-003'),
    ('pur-bank-002', 'do-005'),
    ('pur-bank-003', 'do-004'), ('pur-bank-003', 'do-005'), ('pur-bank-003', 'do-006'),
    ('pur-bank-004', 'do-001'), ('pur-bank-004', 'do-002'), ('pur-bank-004', 'do-004'), ('pur-bank-004', 'do-005'),
    ('pur-bank-005', 'do-001'), ('pur-bank-005', 'do-003'),
    ('pur-bank-006', 'do-005'),
    
    -- Insurance mappings
    ('pur-ins-001', 'do-001'), ('pur-ins-001', 'do-007'), ('pur-ins-001', 'do-004'),
    ('pur-ins-002', 'do-001'), ('pur-ins-002', 'do-007'), ('pur-ins-002', 'do-008'),
    ('pur-ins-003', 'do-007'),
    ('pur-ins-004', 'do-001'), ('pur-ins-004', 'do-007'),
    ('pur-ins-005', 'do-001'), ('pur-ins-005', 'do-002'),
    
    -- Mutual Fund mappings
    ('pur-mf-001', 'do-001'), ('pur-mf-001', 'do-002'), ('pur-mf-001', 'do-003'),
    ('pur-mf-002', 'do-001'), ('pur-mf-002', 'do-002'), ('pur-mf-002', 'do-010'),
    ('pur-mf-003', 'do-009'),
    ('pur-mf-004', 'do-010'),
    ('pur-mf-005', 'do-001'), ('pur-mf-005', 'do-002'), ('pur-mf-005', 'do-010')
ON CONFLICT DO NOTHING;

-- =====================================================
-- 6. SEED CONSENT FORMS (BSFI Templates)
-- =====================================================

INSERT INTO consent_forms (id, tenant_id, name, description, version, purposes, status, is_active, created_by, created_at, updated_at)
VALUES
    -- Banking Forms
    ('cf-bank-001', '11111111-1111-1111-1111-111111111111', 'Account Opening Consent', 'Consent form for new account opening and KYC', '1.0', '["pur-bank-001", "pur-bank-004", "pur-bank-006"]', 'published', true, '11111111-1111-1111-1111-111111111111', NOW(), NOW()),
    ('cf-bank-002', '11111111-1111-1111-1111-111111111111', 'Loan Application Consent', 'Consent for loan processing and credit assessment', '1.0', '["pur-bank-003", "pur-bank-004"]', 'published', true, '11111111-1111-1111-1111-111111111111', NOW(), NOW()),
    ('cf-bank-003', '11111111-1111-1111-1111-111111111111', 'Marketing Preferences', 'Consent for promotional communications', '1.0', '["pur-bank-005"]', 'published', true, '11111111-1111-1111-1111-111111111111', NOW(), NOW()),
    
    -- Insurance Forms
    ('cf-ins-001', '22222222-2222-2222-2222-222222222222', 'Policy Application Consent', 'Consent for insurance policy underwriting', '1.0', '["pur-ins-001", "pur-ins-003", "pur-ins-005"]', 'published', true, '22222222-2222-2222-2222-222222222222', NOW(), NOW()),
    ('cf-ins-002', '22222222-2222-2222-2222-222222222222', 'Claims Processing Consent', 'Consent for processing insurance claims', '1.0', '["pur-ins-002", "pur-ins-003"]', 'published', true, '22222222-2222-2222-2222-222222222222', NOW(), NOW()),
    
    -- Mutual Fund Forms
    ('cf-mf-001', '33333333-3333-3333-3333-333333333333', 'Investment Account Consent', 'Consent for mutual fund account opening', '1.0', '["pur-mf-001", "pur-mf-002", "pur-mf-003"]', 'published', true, '33333333-3333-3333-3333-333333333333', NOW(), NOW()),
    ('cf-mf-002', '33333333-3333-3333-3333-333333333333', 'Portfolio Analytics Consent', 'Consent for investment analytics and reporting', '1.0', '["pur-mf-004", "pur-mf-005"]', 'published', true, '33333333-3333-3333-3333-333333333333', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 7. SEED WEBHOOK CONFIGURATIONS
-- =====================================================

INSERT INTO webhooks (id, tenant_id, name, url, events, headers, is_active, created_at, updated_at)
VALUES
    ('wh-001', '11111111-1111-1111-1111-111111111111', 'Core Banking System', 'https://corebanking.securebank.in/webhooks/consent', '["consent.granted", "consent.revoked", "consent.updated"]', '{"X-API-Key": "secure-key-here"}', true, NOW(), NOW()),
    ('wh-002', '22222222-2222-2222-2222-222222222222', 'Policy Management System', 'https://pms.trustinsure.co.in/webhooks/consent', '["consent.granted", "consent.revoked"]', '{"Authorization": "Bearer token-here"}', true, NOW(), NOW()),
    ('wh-003', '33333333-3333-3333-3333-333333333333', 'Portfolio Manager', 'https://portfolio.wealthgrow.in/webhooks/consent', '["consent.granted", "consent.updated"]', '{"X-Client-Id": "wealthgrow"}', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 8. SEED NOTIFICATION TEMPLATES
-- =====================================================

INSERT INTO notification_templates (id, tenant_id, name, type, subject, body, variables, is_active, created_at, updated_at)
VALUES
    ('nt-001', '11111111-1111-1111-1111-111111111111', 'Consent Confirmation', 'email', 'Your consent has been recorded', 'Dear {{customer_name}}, Your consent for {{purpose}} has been successfully recorded. Reference: {{consent_id}}', '["customer_name", "purpose", "consent_id"]', true, NOW(), NOW()),
    ('nt-002', '11111111-1111-1111-1111-111111111111', 'Consent Expiry Reminder', 'email', 'Your consent is expiring soon', 'Dear {{customer_name}}, Your consent for {{purpose}} will expire on {{expiry_date}}. Please renew if you wish to continue.', '["customer_name", "purpose", "expiry_date"]', true, NOW(), NOW()),
    ('nt-003', '22222222-2222-2222-2222-222222222222', 'Policy Consent Update', 'sms', 'Consent Updated', 'Your consent for {{policy_number}} has been updated. Call 1800-XXX-XXXX for queries.', '["policy_number"]', true, NOW(), NOW()),
    ('nt-004', '33333333-3333-3333-3333-333333333333', 'Investment Consent', 'email', 'Investment account consent recorded', 'Dear {{investor_name}}, We have recorded your consent for {{service}}. You can manage your preferences anytime through our portal.', '["investor_name", "service"]', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 9. SEED COMPLIANCE CONFIGURATIONS (DPDPA 2023)
-- =====================================================

INSERT INTO compliance_configs (id, tenant_id, regulation, config, is_active, created_at, updated_at)
VALUES
    ('cc-001', '11111111-1111-1111-1111-111111111111', 'DPDPA_2023', '{"data_fiduciary": true, "significant_data_fiduciary": true, "consent_age": 18, "data_retention_years": 7, "breach_notification_hours": 72, "languages": ["en", "hi", "mr", "gu", "ta", "te", "kn", "ml", "bn", "pa", "or", "as"]}', true, NOW(), NOW()),
    ('cc-002', '22222222-2222-2222-2222-222222222222', 'DPDPA_2023', '{"data_fiduciary": true, "significant_data_fiduciary": false, "consent_age": 18, "data_retention_years": 10, "breach_notification_hours": 72, "languages": ["en", "hi"]}', true, NOW(), NOW()),
    ('cc-003', '33333333-3333-3333-3333-333333333333', 'DPDPA_2023', '{"data_fiduciary": true, "significant_data_fiduciary": false, "consent_age": 18, "data_retention_years": 7, "breach_notification_hours": 72, "languages": ["en", "hi", "kn", "ta"]}', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 10. SEED AUDIT LOG CONFIGURATIONS
-- =====================================================

INSERT INTO audit_configs (id, tenant_id, events_to_log, retention_days, is_active, created_at, updated_at)
VALUES
    ('ac-001', '11111111-1111-1111-1111-111111111111', '["consent.*", "dsr.*", "user.*", "admin.*"]', 2555, true, NOW(), NOW()),
    ('ac-002', '22222222-2222-2222-2222-222222222222', '["consent.*", "dsr.*", "policy.*"]', 3650, true, NOW(), NOW()),
    ('ac-003', '33333333-3333-3333-3333-333333333333', '["consent.*", "dsr.*", "portfolio.*"]', 2555, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 11. CREATE INDEXES FOR PERFORMANCE
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_purposes_tenant_category ON purposes(tenant_id, category_id);
CREATE INDEX IF NOT EXISTS idx_purposes_legal_basis ON purposes(legal_basis);
CREATE INDEX IF NOT EXISTS idx_data_objects_tenant_category ON data_objects(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_consent_forms_tenant_status ON consent_forms(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_webhooks_tenant_active ON webhooks(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_notification_templates_tenant_type ON notification_templates(tenant_id, type);

-- =====================================================
-- 12. SEED DEFAULT ADMIN USERS (For Development/Demo)
-- =====================================================
-- Note: Passwords should be changed immediately after deployment
-- Default password: Admin@123 (hashed with bcrypt)

INSERT INTO admin_users (id, tenant_id, email, password_hash, full_name, role, is_active, created_at, updated_at)
VALUES
    ('admin-001', '11111111-1111-1111-1111-111111111111', 'admin@securebank.in', '$2a$10$YourHashedPasswordHere', 'Bank Administrator', 'super_admin', true, NOW(), NOW()),
    ('admin-002', '22222222-2222-2222-2222-222222222222', 'admin@trustinsure.co.in', '$2a$10$YourHashedPasswordHere', 'Insurance Administrator', 'super_admin', true, NOW(), NOW()),
    ('admin-003', '33333333-3333-3333-3333-333333333333', 'admin@wealthgrow.in', '$2a$10$YourHashedPasswordHere', 'MF Administrator', 'super_admin', true, NOW(), NOW()),
    ('dpo-001', '11111111-1111-1111-1111-111111111111', 'dpo@securebank.in', '$2a$10$YourHashedPasswordHere', 'Data Protection Officer', 'dpo', true, NOW(), NOW()),
    ('dpo-002', '22222222-2222-2222-2222-222222222222', 'dpo@trustinsure.co.in', '$2a$10$YourHashedPasswordHere', 'Data Protection Officer', 'dpo', true, NOW(), NOW()),
    ('dpo-003', '33333333-3333-3333-3333-333333333333', 'dpo@wealthgrow.in', '$2a$10$YourHashedPasswordHere', 'Data Protection Officer', 'dpo', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 13. SEED API KEYS (For Integration)
-- =====================================================

INSERT INTO api_keys (id, tenant_id, name, key_hash, permissions, rate_limit, is_active, created_at, updated_at, expires_at)
VALUES
    ('api-001', '11111111-1111-1111-1111-111111111111', 'Mobile Banking App', '$2a$10$ApiKeyHashHere', '["consent.read", "consent.create", "consent.update"]', 1000, true, NOW(), NOW(), NOW() + INTERVAL '1 year'),
    ('api-002', '11111111-1111-1111-1111-111111111111', 'Core Banking Integration', '$2a$10$ApiKeyHashHere', '["consent.*", "dsr.*"]', 5000, true, NOW(), NOW(), NOW() + INTERVAL '1 year'),
    ('api-003', '22222222-2222-2222-2222-222222222222', 'Claims Portal', '$2a$10$ApiKeyHashHere', '["consent.read", "consent.create"]', 500, true, NOW(), NOW(), NOW() + INTERVAL '1 year'),
    ('api-004', '33333333-3333-3333-3333-333333333333', 'Investment Portal', '$2a$10$ApiKeyHashHere', '["consent.*"]', 2000, true, NOW(), NOW(), NOW() + INTERVAL '1 year')
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- 14. SEED LANGUAGE TRANSLATIONS (DPDPA 22 Languages)
-- =====================================================

INSERT INTO language_translations (id, language_code, language_name, key, value, created_at, updated_at)
VALUES
    -- English
    ('lt-en-001', 'en', 'English', 'consent.title', 'Data Privacy Consent', NOW(), NOW()),
    ('lt-en-002', 'en', 'English', 'consent.agree', 'I Agree', NOW(), NOW()),
    ('lt-en-003', 'en', 'English', 'consent.disagree', 'I Disagree', NOW(), NOW()),
    ('lt-en-004', 'en', 'English', 'consent.withdraw', 'Withdraw Consent', NOW(), NOW()),
    
    -- Hindi
    ('lt-hi-001', 'hi', 'हिन्दी', 'consent.title', 'डेटा गोपनीयता सहमति', NOW(), NOW()),
    ('lt-hi-002', 'hi', 'हिन्दी', 'consent.agree', 'मैं सहमत हूं', NOW(), NOW()),
    ('lt-hi-003', 'hi', 'हिन्दी', 'consent.disagree', 'मैं असहमत हूं', NOW(), NOW()),
    ('lt-hi-004', 'hi', 'हिन्दी', 'consent.withdraw', 'सहमति वापस लें', NOW(), NOW()),
    
    -- Marathi
    ('lt-mr-001', 'mr', 'मराठी', 'consent.title', 'डेटा गोपनीयता संमती', NOW(), NOW()),
    ('lt-mr-002', 'mr', 'मराठी', 'consent.agree', 'मी सहमत आहे', NOW(), NOW()),
    
    -- Tamil
    ('lt-ta-001', 'ta', 'தமிழ்', 'consent.title', 'தரவு தனியுரிமை ஒப்புதல்', NOW(), NOW()),
    ('lt-ta-002', 'ta', 'தமிழ்', 'consent.agree', 'நான் ஒப்புக்கொள்கிறேன்', NOW(), NOW()),
    
    -- Add more translations as needed for all 22 languages
    -- Bengali, Telugu, Kannada, Malayalam, Gujarati, Punjabi, Odia, Assamese, etc.
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- END OF SEED MIGRATION
-- =====================================================

-- Summary:
-- - 3 BSFI Organizations (Bank, Insurance, Mutual Fund)
-- - 8 Purpose Categories aligned with DPDPA 2023
-- - 16 Specific Purposes for BSFI operations
-- - 10 Data Object types
-- - 7 Consent Form templates
-- - Webhook configurations
-- - Notification templates
-- - Compliance configurations for DPDPA 2023
-- - Admin users and API keys
-- - Multi-language support foundation
