package templates

// DPBNotificationTemplate is the official DPDP-compliant template for Data Protection Board notifications
const DPBNotificationTemplate = `Data Protection Board of India
Ministry of Electronics and Information Technology
Government of India

Subject: Data Breach Notification - {{breach_title}}

Dear Sir/Madam,

We are writing to notify the Data Protection Board of India about a data breach incident in accordance with our obligations under the Digital Personal Data Protection Act, 2023.

**BREACH DETAILS:**

Breach Reference: {{breach_id}}
Company/Organization: {{organization_name}}
Date of Breach: {{breach_date}}
Date of Detection: {{detection_date}}
Date of Containment: {{containment_date}}

**NATURE OF THE BREACH:**
{{breach_description}}

Breach Type: {{breach_type}}
Severity Level: {{severity}}

**DATA AFFECTED:**
Number of Data Principals Affected: {{affected_count}}
Categories of Personal Data Affected: {{data_categories}}

**LIKELY CONSEQUENCES:**
{{impact_description}}

Likelihood of Harm to Data Principals: {{likelihood_of_harm}}

**MEASURES TAKEN:**
Immediate Containment Actions:
{{remedial_actions}}

Preventive Measures Implemented:
{{preventive_measures}}

**DATA PRINCIPAL NOTIFICATION:**
Data Principals Notification Required: {{requires_data_principal_notification}}
Data Principals Notified On: {{data_principal_notification_date}}

**CONTACT INFORMATION:**
Data Protection Officer: {{dpo_name}}
Email: {{dpo_email}}
Phone: {{dpo_phone}}

This notification is submitted in compliance with Section 8 of the Digital Personal Data Protection Act, 2023.

Sincerely,
{{company_name}}
{{submission_date}}
`

// DataPrincipalNotificationTemplate is the DPDP-compliant template for notifying affected individuals
const DataPrincipalNotificationTemplate = `Subject: Important Security Notice - Data Breach Notification

Dear Valued Customer,

We are writing to inform you about a recent security incident that may have affected your personal data. We take the protection of your personal information very seriously and want to provide you with full transparency about what happened and the steps we are taking.

**WHAT HAPPENED:**
On {{breach_date}}, we discovered {{breach_description}}

**WHAT INFORMATION WAS AFFECTED:**
The following categories of your personal data may have been affected:
{{data_categories}}

**WHAT WE ARE DOING:**
We have taken immediate action to:
- {{remedial_actions}}
- {{preventive_measures}}

**WHAT YOU CAN DO:**
We recommend you take the following precautions:
1. Monitor your accounts for any suspicious activity
2. Consider changing your passwords
3. Be alert for phishing attempts or suspicious communications
4. Review your account statements regularly

**YOUR RIGHTS UNDER DPDP ACT 2023:**
As a Data Principal under the Digital Personal Data Protection Act, 2023, you have the right to:
- Access your personal data
- Correction of inaccurate data
- Erasure of personal data
- Grievance redressal

**CONTACT US:**
If you have any questions or concerns about this incident, please contact:

Data Protection Officer: {{dpo_name}}
Email: {{dpo_email}}
Phone: {{dpo_phone}}
Hours: Monday-Friday, 9:00 AM - 6:00 PM IST

We sincerely apologize for any inconvenience this incident may cause. Protecting your personal data is our top priority, and we are committed to preventing such incidents in the future.

Sincerely,
{{company_name}}

---
This notification is provided in compliance with the Digital Personal Data Protection Act, 2023.
If you believe you have not received adequate information or resolution, you may file a complaint with the Data Protection Board of India at https://dpb.gov.in
`

// DataPrincipalNotificationTemplateSMS for SMS notifications (160 char limit awareness)
const DataPrincipalNotificationTemplateSMS = `SECURITY ALERT: {{company_name}} detected a data breach on {{breach_date}} that may affect your personal data. Please check your email for full details and recommended actions. Contact: {{dpo_email}}`

// DPBFollowUpTemplate for providing updates to DPB
const DPBFollowUpTemplate = `Data Protection Board of India

Subject: Follow-up on Data Breach Notification - {{breach_title}} (Ref: {{breach_id}})

Dear Sir/Madam,

This is a follow-up to our initial breach notification submitted on {{initial_notification_date}}.

**UPDATE SUMMARY:**
{{update_description}}

**CURRENT STATUS:**
Status: {{current_status}}
Investigation Progress: {{investigation_progress}}
Root Cause Analysis: {{root_cause}}

**ADDITIONAL ACTIONS TAKEN:**
{{additional_actions}}

**LESSONS LEARNED:**
{{lessons_learned}}

**CLOSURE:**
We believe this incident has been fully resolved and all affected parties have been notified. We request the Board's acknowledgment of this follow-up and closure of this breach case.

Contact: {{dpo_name}} ({{dpo_email}})

Sincerely,
{{company_name}}
{{submission_date}}
`
