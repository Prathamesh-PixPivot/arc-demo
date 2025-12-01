-- =====================================================
-- Rollback: Remove Translation Tables
-- =====================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_user_language_preferences_updated_at ON user_language_preferences;
DROP TRIGGER IF EXISTS update_tenant_translations_updated_at ON tenant_translations;
DROP TRIGGER IF EXISTS update_translation_templates_updated_at ON translation_templates;
DROP TRIGGER IF EXISTS update_translations_updated_at ON translations;
DROP TRIGGER IF EXISTS update_languages_updated_at ON languages;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS user_language_preferences;
DROP TABLE IF EXISTS tenant_translations;
DROP TABLE IF EXISTS translation_templates;
DROP TABLE IF EXISTS translations;
DROP TABLE IF EXISTS languages;

-- =====================================================
-- END OF ROLLBACK
-- =====================================================
