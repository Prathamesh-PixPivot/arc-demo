-- Migration to create purpose_templates table and seed with DPDP-compliant templates
-- Up Migration
CREATE TABLE IF NOT EXISTS purpose_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(20) NOT NULL CHECK (category IN ('marketing', 'analytics', 'functional', 'necessary')),
    legal_basis VARCHAR(30) NOT NULL CHECK (legal_basis IN ('consent', 'contract', 'legal_obligation', 'legitimate_interest')),
    regulatory_framework VARCHAR(10) NOT NULL CHECK (regulatory_framework IN ('dpdp', 'gdpr', 'ccpa')),
    required_data_objects TEXT[] DEFAULT '{}',
    suggested_retention_days INTEGER DEFAULT 365,
    compliance_notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_purpose_templates_category ON purpose_templates(category);
CREATE INDEX IF NOT EXISTS idx_purpose_templates_legal_basis ON purpose_templates(legal_basis);
CREATE INDEX IF NOT EXISTS idx_purpose_templates_framework ON purpose_templates(regulatory_framework);
CREATE INDEX IF NOT EXISTS idx_purpose_templates_active ON purpose_templates(is_active);

-- Seed DPDP-compliant purpose templates

-- NECESSARY PURPOSES (Legal Obligation/Contract)
INSERT INTO purpose_templates (name, description, category, legal_basis, regulatory_framework, required_data_objects, suggested_retention_days, compliance_notes) VALUES
('Transaction Processing', 'Processing payments and completing transactions for goods or services', 'necessary', 'contract', 'dpdp', '{"name", "email", "phone", "payment_details", "billing_address"}', 2555, 'Required for contract performance under DPDP Act Section 6(1)(a)'),
('Account Management', 'Creating and maintaining user accounts and profiles', 'necessary', 'contract', 'dpdp', '{"name", "email", "phone", "address", "date_of_birth"}', 2555, 'Essential for service provision and account security'),
('Order Fulfillment', 'Processing and delivering orders to customers', 'necessary', 'contract', 'dpdp', '{"name", "email", "phone", "shipping_address", "order_details"}', 1095, 'Required for contract performance and delivery'),
('Customer Authentication', 'Verifying user identity and securing accounts', 'necessary', 'contract', 'dpdp', '{"email", "phone", "authentication_credentials"}', 2555, 'Essential for account security and fraud prevention'),
('Legal Compliance', 'Meeting regulatory and legal obligations', 'necessary', 'legal_obligation', 'dpdp', '{"name", "email", "phone", "transaction_records", "kyc_documents"}', 2555, 'Required under various Indian laws including DPDP Act'),
('Tax Compliance', 'Maintaining records for tax purposes and GST compliance', 'necessary', 'legal_obligation', 'dpdp', '{"name", "pan_number", "gstin", "billing_address", "transaction_amount"}', 2555, 'Required under Income Tax Act and GST Act'),
('Fraud Prevention', 'Detecting and preventing fraudulent activities', 'necessary', 'legal_obligation', 'dpdp', '{"ip_address", "device_info", "transaction_patterns", "behavioral_data"}', 1095, 'Required for financial crime prevention under RBI guidelines'),
('Data Security', 'Protecting user data and maintaining system security', 'necessary', 'legal_obligation', 'dpdp', '{"access_logs", "security_events", "audit_trails"}', 2555, 'Required under DPDP Act for data protection'),
('Dispute Resolution', 'Handling customer complaints and disputes', 'necessary', 'contract', 'dpdp', '{"name", "email", "phone", "transaction_details", "complaint_details"}', 1095, 'Required for customer service and dispute resolution'),
('Refund Processing', 'Processing refunds and returns', 'necessary', 'contract', 'dpdp', '{"name", "email", "phone", "payment_details", "refund_reason"}', 1095, 'Required for contract performance and customer rights'),

-- FUNCTIONAL PURPOSES (Contract/Legitimate Interest)
('Customer Support', 'Providing customer service and technical support', 'functional', 'contract', 'dpdp', '{"name", "email", "phone", "support_history", "product_usage"}', 1095, 'Required for service quality and customer satisfaction'),
('Product Delivery', 'Delivering digital or physical products to customers', 'functional', 'contract', 'dpdp', '{"name", "email", "phone", "delivery_address", "product_preferences"}', 730, 'Essential for service delivery and customer experience'),
('Service Personalization', 'Customizing services based on user preferences', 'functional', 'legitimate_interest', 'dpdp', '{"preferences", "usage_history", "behavioral_data", "location"}', 730, 'Balancing user experience with privacy rights'),
('System Administration', 'Managing and maintaining IT systems and infrastructure', 'functional', 'legitimate_interest', 'dpdp', '{"system_logs", "performance_data", "error_logs", "usage_statistics"}', 365, 'Required for system stability and performance'),
('Quality Assurance', 'Monitoring and improving service quality', 'functional', 'legitimate_interest', 'dpdp', '{"feedback", "usage_patterns", "performance_metrics", "error_reports"}', 730, 'Balancing service improvement with privacy'),
('Content Delivery', 'Delivering personalized content and recommendations', 'functional', 'legitimate_interest', 'dpdp', '{"content_preferences", "viewing_history", "interaction_data"}', 365, 'Enhancing user experience while respecting privacy'),
('Notification Services', 'Sending important service notifications and updates', 'functional', 'contract', 'dpdp', '{"email", "phone", "notification_preferences", "device_tokens"}', 730, 'Required for service communication and user safety'),
('Backup and Recovery', 'Maintaining data backups for business continuity', 'functional', 'legitimate_interest', 'dpdp', '{"user_data", "system_backups", "recovery_logs"}', 1095, 'Essential for data protection and business continuity'),
('Performance Monitoring', 'Monitoring system and application performance', 'functional', 'legitimate_interest', 'dpdp', '{"performance_metrics", "system_logs", "user_interactions"}', 365, 'Required for service reliability and optimization'),
('User Experience Enhancement', 'Improving user interface and experience', 'functional', 'legitimate_interest', 'dpdp', '{"usage_patterns", "click_data", "navigation_behavior"}', 365, 'Balancing UX improvement with privacy'),

-- ANALYTICS PURPOSES (Legitimate Interest/Consent)
('Website Analytics', 'Analyzing website usage and user behavior', 'analytics', 'legitimate_interest', 'dpdp', '{"page_views", "session_data", "click_patterns", "referral_sources"}', 730, 'Balancing business insights with user privacy'),
('Business Intelligence', 'Generating insights for business decision making', 'analytics', 'legitimate_interest', 'dpdp', '{"aggregated_data", "usage_statistics", "performance_metrics"}', 1095, 'Using anonymized data for business insights'),
('Product Analytics', 'Understanding product usage and feature adoption', 'analytics', 'legitimate_interest', 'dpdp', '{"feature_usage", "user_journeys", "product_interactions"}', 730, 'Improving products based on usage patterns'),
('User Behavior Analysis', 'Analyzing user behavior patterns for insights', 'analytics', 'consent', 'dpdp', '{"behavioral_data", "interaction_patterns", "usage_frequency"}', 365, 'Requires explicit consent for detailed behavioral analysis'),
('Market Research', 'Conducting research for market understanding', 'analytics', 'consent', 'dpdp', '{"demographic_data", "preferences", "survey_responses", "feedback"}', 730, 'Requires consent for market research activities'),
('Trend Analysis', 'Identifying trends and patterns in data', 'analytics', 'legitimate_interest', 'dpdp', '{"aggregated_trends", "statistical_data", "pattern_analysis"}', 365, 'Using anonymized data for trend identification'),
('Performance Analytics', 'Measuring and analyzing system performance', 'analytics', 'legitimate_interest', 'dpdp', '{"performance_data", "load_metrics", "response_times"}', 365, 'Required for system optimization'),
('Usage Statistics', 'Collecting statistics on service usage', 'analytics', 'legitimate_interest', 'dpdp', '{"usage_counts", "session_duration", "feature_adoption"}', 365, 'Anonymized statistics for service improvement'),
('Conversion Analysis', 'Analyzing user conversion and funnel metrics', 'analytics', 'legitimate_interest', 'dpdp', '{"conversion_data", "funnel_metrics", "user_journeys"}', 365, 'Business optimization with privacy protection'),
('A/B Testing', 'Testing different versions of products or features', 'analytics', 'legitimate_interest', 'dpdp', '{"test_group_data", "experiment_results", "user_responses"}', 180, 'Short-term testing for product improvement'),

-- MARKETING PURPOSES (Consent Required)
('Email Marketing', 'Sending promotional emails and newsletters', 'marketing', 'consent', 'dpdp', '{"email", "name", "preferences", "engagement_history"}', 1095, 'Requires explicit consent under DPDP Act Section 6(1)(c)'),
('SMS Marketing', 'Sending promotional SMS messages', 'marketing', 'consent', 'dpdp', '{"phone", "name", "preferences", "opt_in_status"}', 1095, 'Requires explicit consent and TRAI DND compliance'),
('Push Notifications', 'Sending promotional push notifications', 'marketing', 'consent', 'dpdp', '{"device_tokens", "preferences", "engagement_data"}', 365, 'Requires explicit consent for promotional content'),
('Targeted Advertising', 'Displaying personalized advertisements', 'marketing', 'consent', 'dpdp', '{"behavioral_data", "interests", "demographics", "ad_interactions"}', 365, 'Requires explicit consent for behavioral targeting'),
('Social Media Marketing', 'Marketing through social media platforms', 'marketing', 'consent', 'dpdp', '{"social_profiles", "interests", "engagement_data", "demographic_info"}', 730, 'Requires consent for social media marketing'),
('Retargeting Campaigns', 'Re-engaging users who have shown interest', 'marketing', 'consent', 'dpdp', '{"website_behavior", "product_views", "cart_abandonment"}', 365, 'Requires consent for retargeting activities'),
('Promotional Campaigns', 'Running promotional and discount campaigns', 'marketing', 'consent', 'dpdp', '{"contact_info", "purchase_history", "preferences", "campaign_responses"}', 730, 'Requires explicit consent for promotional activities'),
('Customer Segmentation', 'Segmenting customers for targeted marketing', 'marketing', 'consent', 'dpdp', '{"demographic_data", "behavioral_data", "purchase_patterns"}', 730, 'Requires consent for marketing segmentation'),
('Lead Generation', 'Generating and nurturing sales leads', 'marketing', 'consent', 'dpdp', '{"contact_info", "interests", "engagement_data", "lead_scores"}', 1095, 'Requires consent for lead generation activities'),
('Affiliate Marketing', 'Marketing through affiliate partners', 'marketing', 'consent', 'dpdp', '{"referral_data", "conversion_tracking", "affiliate_interactions"}', 365, 'Requires consent for affiliate marketing'),
('Content Marketing', 'Delivering personalized marketing content', 'marketing', 'consent', 'dpdp', '{"content_preferences", "engagement_history", "behavioral_data"}', 730, 'Requires consent for personalized marketing content'),
('Event Marketing', 'Marketing events and webinars', 'marketing', 'consent', 'dpdp', '{"contact_info", "event_preferences", "attendance_history"}', 365, 'Requires consent for event marketing'),
('Survey and Feedback', 'Collecting feedback for marketing purposes', 'marketing', 'consent', 'dpdp', '{"contact_info", "feedback_responses", "survey_data"}', 365, 'Requires consent when used for marketing'),
('Loyalty Programs', 'Managing customer loyalty and rewards programs', 'marketing', 'consent', 'dpdp', '{"purchase_history", "loyalty_points", "reward_preferences"}', 1095, 'Requires consent for loyalty program marketing'),
('Cross-selling', 'Promoting related products and services', 'marketing', 'consent', 'dpdp', '{"purchase_history", "product_preferences", "behavioral_data"}', 730, 'Requires consent for cross-selling activities'),

-- ADDITIONAL SPECIALIZED PURPOSES
('Third-party Integrations', 'Integrating with third-party services and APIs', 'functional', 'legitimate_interest', 'dpdp', '{"integration_data", "api_usage", "service_responses"}', 365, 'Required for service functionality and integrations'),
('Compliance Monitoring', 'Monitoring compliance with regulations and policies', 'necessary', 'legal_obligation', 'dpdp', '{"compliance_data", "audit_logs", "policy_violations"}', 2555, 'Required under DPDP Act and other regulations'),
('Risk Assessment', 'Assessing and managing business risks', 'functional', 'legitimate_interest', 'dpdp', '{"risk_indicators", "assessment_data", "mitigation_measures"}', 1095, 'Required for business risk management'),
('Vendor Management', 'Managing relationships with vendors and suppliers', 'functional', 'contract', 'dpdp', '{"vendor_data", "contract_details", "performance_metrics"}', 2555, 'Required for vendor relationship management'),
('Research and Development', 'Conducting R&D for product improvement', 'analytics', 'legitimate_interest', 'dpdp', '{"anonymized_usage_data", "feature_requests", "innovation_metrics"}', 1095, 'Using anonymized data for R&D purposes');

-- Update the updated_at timestamp trigger
CREATE OR REPLACE FUNCTION update_purpose_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_purpose_templates_updated_at
    BEFORE UPDATE ON purpose_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_purpose_templates_updated_at();

-- Down Migration (for rollback)
-- DROP TRIGGER IF EXISTS update_purpose_templates_updated_at ON purpose_templates;
-- DROP FUNCTION IF EXISTS update_purpose_templates_updated_at();
-- DROP TABLE IF EXISTS purpose_templates;
