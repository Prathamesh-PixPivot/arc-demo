-- =====================================================
-- Create Translation Tables for Multi-language Support
-- DPDPA 2023 Compliance - 22+ Languages
-- =====================================================

-- 1. Create languages table
CREATE TABLE IF NOT EXISTS languages (
    code VARCHAR(10) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    native_name VARCHAR(50) NOT NULL,
    direction VARCHAR(3) DEFAULT 'ltr' CHECK (direction IN ('ltr', 'rtl')),
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add comment
COMMENT ON TABLE languages IS 'Supported languages for DPDPA 2023 compliance';
COMMENT ON COLUMN languages.code IS 'ISO 639-1 or custom language code';
COMMENT ON COLUMN languages.direction IS 'Text direction: ltr (left-to-right) or rtl (right-to-left)';

-- 2. Create translations table
CREATE TABLE IF NOT EXISTS translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    language_code VARCHAR(10) NOT NULL REFERENCES languages(code) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    context VARCHAR(100),
    metadata JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(language_code, key)
);

-- Add indexes
CREATE INDEX idx_translations_lang_key ON translations(language_code, key);
CREATE INDEX idx_translations_context ON translations(context) WHERE context IS NOT NULL;
CREATE INDEX idx_translations_active ON translations(is_active) WHERE is_active = true;

-- Add comments
COMMENT ON TABLE translations IS 'Translation entries for multi-language support';
COMMENT ON COLUMN translations.key IS 'Translation key (e.g., consent.title)';
COMMENT ON COLUMN translations.value IS 'Translated text value';
COMMENT ON COLUMN translations.context IS 'Context for translation (e.g., consent_form, dsr, notification)';

-- 3. Create translation_templates table
CREATE TABLE IF NOT EXISTS translation_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    placeholders JSONB, -- Array of placeholder names like ["{{name}}", "{{date}}"]
    context VARCHAR(100),
    category VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX idx_translation_templates_category ON translation_templates(category);
CREATE INDEX idx_translation_templates_context ON translation_templates(context);

-- Add comments
COMMENT ON TABLE translation_templates IS 'Templates for translations with placeholders';
COMMENT ON COLUMN translation_templates.placeholders IS 'Array of placeholder variables in the template';

-- 4. Create tenant_translations table (for tenant-specific overrides)
CREATE TABLE IF NOT EXISTS tenant_translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    language_code VARCHAR(10) NOT NULL REFERENCES languages(code) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, language_code, key)
);

-- Add indexes
CREATE INDEX idx_tenant_translations ON tenant_translations(tenant_id, language_code, key);
CREATE INDEX idx_tenant_translations_active ON tenant_translations(is_active) WHERE is_active = true;

-- Add comments
COMMENT ON TABLE tenant_translations IS 'Tenant-specific translation overrides';

-- 5. Create user_language_preferences table
CREATE TABLE IF NOT EXISTS user_language_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL,
    language_code VARCHAR(10) NOT NULL REFERENCES languages(code),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add index
CREATE INDEX idx_user_language_preferences_user ON user_language_preferences(user_id);

-- Add comment
COMMENT ON TABLE user_language_preferences IS 'User language preferences';

-- 6. Create triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_languages_updated_at BEFORE UPDATE ON languages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_translations_updated_at BEFORE UPDATE ON translations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_translation_templates_updated_at BEFORE UPDATE ON translation_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_translations_updated_at BEFORE UPDATE ON tenant_translations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_language_preferences_updated_at BEFORE UPDATE ON user_language_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 7. Insert DPDPA 2023 mandated languages (22 official languages of India)
INSERT INTO languages (code, name, native_name, direction, is_active, is_default, sort_order) VALUES
    ('en', 'English', 'English', 'ltr', true, true, 1),
    ('hi', 'Hindi', 'हिन्दी', 'ltr', true, false, 2),
    ('bn', 'Bengali', 'বাংলা', 'ltr', true, false, 3),
    ('te', 'Telugu', 'తెలుగు', 'ltr', true, false, 4),
    ('mr', 'Marathi', 'मराठी', 'ltr', true, false, 5),
    ('ta', 'Tamil', 'தமிழ்', 'ltr', true, false, 6),
    ('ur', 'Urdu', 'اردو', 'rtl', true, false, 7),
    ('gu', 'Gujarati', 'ગુજરાતી', 'ltr', true, false, 8),
    ('kn', 'Kannada', 'ಕನ್ನಡ', 'ltr', true, false, 9),
    ('ml', 'Malayalam', 'മലയാളം', 'ltr', true, false, 10),
    ('or', 'Odia', 'ଓଡ଼ିଆ', 'ltr', true, false, 11),
    ('pa', 'Punjabi', 'ਪੰਜਾਬੀ', 'ltr', true, false, 12),
    ('as', 'Assamese', 'অসমীয়া', 'ltr', true, false, 13),
    ('mai', 'Maithili', 'मैथिली', 'ltr', true, false, 14),
    ('sat', 'Santali', 'ᱥᱟᱱᱛᱟᱲᱤ', 'ltr', true, false, 15),
    ('ks', 'Kashmiri', 'कॉशुर', 'rtl', true, false, 16),
    ('ne', 'Nepali', 'नेपाली', 'ltr', true, false, 17),
    ('sd', 'Sindhi', 'سنڌي', 'rtl', true, false, 18),
    ('kok', 'Konkani', 'कोंकणी', 'ltr', true, false, 19),
    ('mni', 'Manipuri', 'মৈতৈলোন্', 'ltr', true, false, 20),
    ('doi', 'Dogri', 'डोगरी', 'ltr', true, false, 21),
    ('bodo', 'Bodo', 'बर''', 'ltr', true, false, 22)
ON CONFLICT (code) DO NOTHING;

-- 8. Insert default English translations
INSERT INTO translations (language_code, key, value, context) VALUES
    -- Consent related
    ('en', 'consent.title', 'Data Privacy Consent', 'consent_form'),
    ('en', 'consent.agree', 'I Agree', 'consent_form'),
    ('en', 'consent.disagree', 'I Disagree', 'consent_form'),
    ('en', 'consent.withdraw', 'Withdraw Consent', 'consent_form'),
    ('en', 'consent.purpose', 'Purpose', 'consent_form'),
    ('en', 'consent.data_object', 'Data Categories', 'consent_form'),
    ('en', 'consent.expiry', 'Consent Expiry', 'consent_form'),
    ('en', 'consent.granted_on', 'Granted On', 'consent_form'),
    ('en', 'consent.valid_until', 'Valid Until', 'consent_form'),
    
    -- DSR related
    ('en', 'dsr.title', 'Data Subject Rights', 'dsr'),
    ('en', 'dsr.access', 'Right to Access', 'dsr'),
    ('en', 'dsr.correction', 'Right to Correction', 'dsr'),
    ('en', 'dsr.erasure', 'Right to Erasure', 'dsr'),
    ('en', 'dsr.portability', 'Right to Data Portability', 'dsr'),
    ('en', 'dsr.objection', 'Right to Object', 'dsr'),
    ('en', 'dsr.submit_request', 'Submit Request', 'dsr'),
    ('en', 'dsr.request_status', 'Request Status', 'dsr'),
    
    -- Notification related
    ('en', 'notification.email', 'Email Notification', 'notification'),
    ('en', 'notification.sms', 'SMS Notification', 'notification'),
    ('en', 'notification.push', 'Push Notification', 'notification'),
    ('en', 'notification.consent_granted', 'Your consent has been recorded', 'notification'),
    ('en', 'notification.consent_withdrawn', 'Your consent has been withdrawn', 'notification'),
    
    -- Common UI elements
    ('en', 'ui.submit', 'Submit', 'ui'),
    ('en', 'ui.cancel', 'Cancel', 'ui'),
    ('en', 'ui.save', 'Save', 'ui'),
    ('en', 'ui.delete', 'Delete', 'ui'),
    ('en', 'ui.edit', 'Edit', 'ui'),
    ('en', 'ui.view', 'View', 'ui'),
    ('en', 'ui.search', 'Search', 'ui'),
    ('en', 'ui.filter', 'Filter', 'ui'),
    ('en', 'ui.export', 'Export', 'ui'),
    ('en', 'ui.import', 'Import', 'ui'),
    ('en', 'ui.loading', 'Loading...', 'ui'),
    ('en', 'ui.error', 'Error', 'ui'),
    ('en', 'ui.success', 'Success', 'ui'),
    ('en', 'ui.warning', 'Warning', 'ui'),
    ('en', 'ui.info', 'Information', 'ui'),
    
    -- DPDPA specific
    ('en', 'dpdpa.notice', 'Privacy Notice as per DPDPA 2023', 'dpdpa'),
    ('en', 'dpdpa.rights', 'Your Rights under DPDPA 2023', 'dpdpa'),
    ('en', 'dpdpa.grievance', 'Grievance Redressal', 'dpdpa'),
    ('en', 'dpdpa.dpo_contact', 'Data Protection Officer Contact', 'dpdpa'),
    ('en', 'dpdpa.data_principal', 'Data Principal', 'dpdpa'),
    ('en', 'dpdpa.data_fiduciary', 'Data Fiduciary', 'dpdpa'),
    ('en', 'dpdpa.consent_manager', 'Consent Manager', 'dpdpa'),
    
    -- Error messages
    ('en', 'error.required_field', 'This field is required', 'error'),
    ('en', 'error.invalid_email', 'Invalid email address', 'error'),
    ('en', 'error.invalid_phone', 'Invalid phone number', 'error'),
    ('en', 'error.consent_expired', 'This consent has expired', 'error'),
    ('en', 'error.unauthorized', 'You are not authorized to perform this action', 'error'),
    
    -- Success messages
    ('en', 'success.consent_granted', 'Consent granted successfully', 'success'),
    ('en', 'success.consent_withdrawn', 'Consent withdrawn successfully', 'success'),
    ('en', 'success.dsr_submitted', 'Your request has been submitted', 'success'),
    ('en', 'success.profile_updated', 'Profile updated successfully', 'success')
ON CONFLICT (language_code, key) DO NOTHING;

-- 9. Insert sample Hindi translations
INSERT INTO translations (language_code, key, value, context) VALUES
    -- Consent related
    ('hi', 'consent.title', 'डेटा गोपनीयता सहमति', 'consent_form'),
    ('hi', 'consent.agree', 'मैं सहमत हूं', 'consent_form'),
    ('hi', 'consent.disagree', 'मैं असहमत हूं', 'consent_form'),
    ('hi', 'consent.withdraw', 'सहमति वापस लें', 'consent_form'),
    ('hi', 'consent.purpose', 'उद्देश्य', 'consent_form'),
    ('hi', 'consent.data_object', 'डेटा श्रेणियां', 'consent_form'),
    ('hi', 'consent.expiry', 'सहमति समाप्ति', 'consent_form'),
    
    -- DSR related
    ('hi', 'dsr.title', 'डेटा विषय अधिकार', 'dsr'),
    ('hi', 'dsr.access', 'पहुंच का अधिकार', 'dsr'),
    ('hi', 'dsr.correction', 'सुधार का अधिकार', 'dsr'),
    ('hi', 'dsr.erasure', 'मिटाने का अधिकार', 'dsr'),
    
    -- Common UI elements
    ('hi', 'ui.submit', 'जमा करें', 'ui'),
    ('hi', 'ui.cancel', 'रद्द करें', 'ui'),
    ('hi', 'ui.save', 'सहेजें', 'ui'),
    ('hi', 'ui.delete', 'हटाएं', 'ui'),
    ('hi', 'ui.edit', 'संपादित करें', 'ui'),
    ('hi', 'ui.view', 'देखें', 'ui'),
    ('hi', 'ui.search', 'खोजें', 'ui'),
    
    -- DPDPA specific
    ('hi', 'dpdpa.notice', 'DPDPA 2023 के अनुसार गोपनीयता सूचना', 'dpdpa'),
    ('hi', 'dpdpa.rights', 'DPDPA 2023 के तहत आपके अधिकार', 'dpdpa'),
    ('hi', 'dpdpa.grievance', 'शिकायत निवारण', 'dpdpa'),
    ('hi', 'dpdpa.dpo_contact', 'डेटा संरक्षण अधिकारी संपर्क', 'dpdpa')
ON CONFLICT (language_code, key) DO NOTHING;

-- 10. Create translation templates
INSERT INTO translation_templates (key, description, placeholders, context, category) VALUES
    ('consent.confirmation_email', 'Email template for consent confirmation', '["{{user_name}}", "{{purpose}}", "{{date}}"]', 'notification', 'email'),
    ('consent.expiry_reminder', 'Reminder for consent expiry', '["{{user_name}}", "{{purpose}}", "{{expiry_date}}"]', 'notification', 'email'),
    ('dsr.request_confirmation', 'DSR request confirmation', '["{{user_name}}", "{{request_type}}", "{{request_id}}"]', 'notification', 'email'),
    ('dsr.request_completed', 'DSR request completion notification', '["{{user_name}}", "{{request_type}}", "{{completion_date}}"]', 'notification', 'email')
ON CONFLICT (key) DO NOTHING;

-- =====================================================
-- END OF MIGRATION
-- =====================================================
