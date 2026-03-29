-- 027_data_categories.sql
-- Seed data for data classification levels and data categories
-- Inserts 5 classification levels and 30+ data categories
-- Uses the system org '00000000-0000-0000-0000-000000000001' as template org

DO $$
DECLARE
    v_org UUID := '00000000-0000-0000-0000-000000000001';
    -- Classification IDs
    c_public       UUID;
    c_internal     UUID;
    c_confidential UUID;
    c_restricted   UUID;
    c_top_secret   UUID;
BEGIN

-- ============================================================
-- DATA CLASSIFICATION LEVELS
-- ============================================================

INSERT INTO data_classifications (id, organization_id, name, level, description, handling_requirements, encryption_required, access_restriction_required, data_masking_required, retention_policy, disposal_method, color_hex, is_system, sort_order)
VALUES
    (gen_random_uuid(), v_org, 'Public', 0,
     'Information intended for public disclosure. No confidentiality requirements.',
     'No special handling required. May be freely shared externally.',
     false, false, false,
     'Retain as long as useful. No mandatory retention period.',
     'Standard deletion or recycling.',
     '#22C55E', true, 0),

    (gen_random_uuid(), v_org, 'Internal', 1,
     'Information for internal use only. Low sensitivity but not for public release.',
     'Share only with employees and authorised contractors. Label as Internal.',
     false, false, false,
     'Retain for business need plus 1 year. Review annually.',
     'Secure deletion; shred paper documents.',
     '#3B82F6', true, 1),

    (gen_random_uuid(), v_org, 'Confidential', 2,
     'Sensitive business information. Unauthorised disclosure could cause harm.',
     'Encrypt in transit and at rest. Access restricted to need-to-know basis. Log all access.',
     true, true, false,
     'Retain for regulatory period or 7 years, whichever is longer.',
     'Cryptographic erasure or certified destruction.',
     '#F59E0B', true, 2),

    (gen_random_uuid(), v_org, 'Restricted', 3,
     'Highly sensitive data including personal data, financial records, trade secrets. Breach would cause significant harm.',
     'Encrypt at all times. Multi-factor authentication for access. Full audit trail. Data masking in non-production environments.',
     true, true, true,
     'Minimum retention only. Delete as soon as processing purpose fulfilled.',
     'Cryptographic erasure with certificate of destruction. Physical media: degaussing or incineration.',
     '#EF4444', true, 3),

    (gen_random_uuid(), v_org, 'Top Secret', 4,
     'Most sensitive data. Special category personal data, national security, critical infrastructure secrets. Compromise could cause catastrophic damage.',
     'Hardware security modules for encryption keys. Named-individual access only with manager approval. Real-time monitoring. No cloud storage without explicit CISO approval.',
     true, true, true,
     'Strict minimum retention. Quarterly review of continued need.',
     'NSA/NCSC-approved destruction methods. Witnessed destruction with signed certificate.',
     '#7C3AED', true, 4)
RETURNING id INTO c_public; -- We need to fetch IDs for FK references

-- Retrieve the IDs we just inserted
SELECT id INTO c_public FROM data_classifications WHERE organization_id = v_org AND level = 0;
SELECT id INTO c_internal FROM data_classifications WHERE organization_id = v_org AND level = 1;
SELECT id INTO c_confidential FROM data_classifications WHERE organization_id = v_org AND level = 2;
SELECT id INTO c_restricted FROM data_classifications WHERE organization_id = v_org AND level = 3;
SELECT id INTO c_top_secret FROM data_classifications WHERE organization_id = v_org AND level = 4;

-- ============================================================
-- DATA CATEGORIES — Personal Data (Art. 4 GDPR)
-- ============================================================

INSERT INTO data_categories (organization_id, name, category_type, gdpr_special_category, description, examples, classification_id, retention_period_months, is_system) VALUES

-- Basic Personal Identifiers
(v_org, 'Full Name', 'personal_data', false,
 'Individual''s first name, last name, middle name, and any known aliases.',
 ARRAY['John Smith', 'Jane Doe', 'Dr. Alice Brown'],
 c_confidential, 36, true),

(v_org, 'Email Address', 'personal_data', false,
 'Personal or work email addresses used to identify or contact individuals.',
 ARRAY['john@example.com', 'jane.doe@company.co.uk'],
 c_confidential, 36, true),

(v_org, 'Phone Number', 'personal_data', false,
 'Mobile, landline, or work telephone numbers.',
 ARRAY['+44 7700 900000', '020 7946 0958'],
 c_confidential, 36, true),

(v_org, 'Postal Address', 'personal_data', false,
 'Home address, work address, or mailing address of individuals.',
 ARRAY['123 High Street, London, EC1A 1BB'],
 c_confidential, 36, true),

(v_org, 'Date of Birth', 'personal_data', false,
 'Individual''s date of birth or age.',
 ARRAY['1990-01-15', 'Age: 34'],
 c_confidential, 36, true),

(v_org, 'National ID Number', 'personal_data', false,
 'Government-issued identification numbers (NI number, SSN, etc.).',
 ARRAY['AB 12 34 56 C (UK NI)', '078-05-1120 (US SSN)'],
 c_restricted, 12, true),

(v_org, 'Passport Number', 'personal_data', false,
 'Passport identification numbers and associated data.',
 ARRAY['123456789'],
 c_restricted, 12, true),

(v_org, 'IP Address', 'personal_data', false,
 'Internet Protocol addresses that can identify a device or individual.',
 ARRAY['192.168.1.1', '2001:0db8:85a3:0000:0000:8a2e:0370:7334'],
 c_internal, 6, true),

(v_org, 'Cookie Identifiers', 'personal_data', false,
 'Browser cookies, tracking pixels, and device fingerprints used to identify users.',
 ARRAY['_ga=GA1.2.123456789', 'session_id=abc123'],
 c_internal, 6, true),

(v_org, 'Geolocation Data', 'personal_data', false,
 'GPS coordinates, cell tower data, or IP-derived location information.',
 ARRAY['51.5074, -0.1278 (London)', 'Wi-Fi positioning data'],
 c_confidential, 12, true),

(v_org, 'Biometric Data', 'personal_data', false,
 'Fingerprints, facial recognition templates, voice prints, retina scans (when used for identification).',
 ARRAY['Fingerprint template hash', 'FaceID enrollment data'],
 c_top_secret, 6, true),

(v_org, 'Photograph / Image', 'personal_data', false,
 'Photographs, video recordings, or CCTV footage containing identifiable individuals.',
 ARRAY['Profile photo', 'CCTV footage', 'ID card scan'],
 c_confidential, 24, true),

-- ============================================================
-- DATA CATEGORIES — Special Category Data (Art. 9 GDPR)
-- ============================================================

(v_org, 'Health Data', 'special_category', true,
 'Physical or mental health conditions, medical records, disability status, prescriptions.',
 ARRAY['Medical diagnosis', 'Prescription records', 'Sick leave records', 'Disability status'],
 c_top_secret, 60, true),

(v_org, 'Genetic Data', 'special_category', true,
 'Data relating to inherited or acquired genetic characteristics of a person.',
 ARRAY['DNA test results', 'Genetic predisposition data'],
 c_top_secret, 36, true),

(v_org, 'Racial or Ethnic Origin', 'special_category', true,
 'Data revealing racial or ethnic origin of individuals.',
 ARRAY['Ethnicity field in HR records', 'Diversity monitoring data'],
 c_top_secret, 36, true),

(v_org, 'Political Opinions', 'special_category', true,
 'Data revealing political opinions, party membership, or political activities.',
 ARRAY['Political party membership', 'Voting preferences'],
 c_top_secret, 12, true),

(v_org, 'Religious or Philosophical Beliefs', 'special_category', true,
 'Data revealing religious beliefs, philosophical convictions, or similar.',
 ARRAY['Religious affiliation', 'Dietary requirements linked to belief', 'Prayer room booking'],
 c_top_secret, 12, true),

(v_org, 'Trade Union Membership', 'special_category', true,
 'Data indicating membership or activity in trade unions.',
 ARRAY['Union membership records', 'Union representative status'],
 c_top_secret, 36, true),

(v_org, 'Sexual Orientation', 'special_category', true,
 'Data concerning a person''s sex life or sexual orientation.',
 ARRAY['Sexual orientation field', 'Partner benefits records'],
 c_top_secret, 12, true),

(v_org, 'Criminal Convictions and Offences', 'special_category', true,
 'Data relating to criminal convictions, offences, or related security measures (Art. 10).',
 ARRAY['DBS check results', 'Criminal record disclosure', 'Spent convictions'],
 c_top_secret, 36, true),

-- ============================================================
-- DATA CATEGORIES — Financial Data
-- ============================================================

(v_org, 'Bank Account Details', 'financial', false,
 'Bank account numbers, sort codes, IBAN, SWIFT/BIC codes.',
 ARRAY['Sort code: 12-34-56', 'Account: 12345678', 'IBAN: GB29 NWBK 6016 1331 9268 19'],
 c_restricted, 84, true),

(v_org, 'Credit/Debit Card Data', 'financial', false,
 'Payment card numbers, expiry dates, CVV (PCI DSS scope).',
 ARRAY['Card number: 4111-XXXX-XXXX-1234', 'Expiry: 12/26'],
 c_restricted, 12, true),

(v_org, 'Salary and Compensation', 'financial', false,
 'Employee salary, bonus, commission, benefits, and total compensation data.',
 ARRAY['Annual salary: 65000', 'Bonus: 5000', 'Benefits package value'],
 c_restricted, 84, true),

(v_org, 'Credit Score', 'financial', false,
 'Credit ratings, credit history, and financial risk scores.',
 ARRAY['Experian score: 750', 'Credit check result: Pass'],
 c_restricted, 36, true),

(v_org, 'Tax Information', 'financial', false,
 'Tax identification numbers, tax returns, P60/P45 forms, VAT records.',
 ARRAY['UTR: 1234567890', 'P60 annual statement', 'VAT registration number'],
 c_restricted, 84, true),

-- ============================================================
-- DATA CATEGORIES — Employment Data
-- ============================================================

(v_org, 'Employment History', 'business', false,
 'Previous employers, job titles, dates of employment, reasons for leaving.',
 ARRAY['CV/resume', 'Reference letters', 'LinkedIn profile data'],
 c_confidential, 84, true),

(v_org, 'Performance Reviews', 'business', false,
 'Employee performance appraisals, ratings, goals, and development plans.',
 ARRAY['Annual review scores', 'Quarterly objectives', 'PIP documentation'],
 c_restricted, 60, true),

(v_org, 'Disciplinary Records', 'business', false,
 'Warnings, disciplinary hearings, grievances, and outcomes.',
 ARRAY['Written warning', 'Investigation notes', 'Grievance records'],
 c_restricted, 72, true),

(v_org, 'Training Records', 'business', false,
 'Certifications, courses completed, training attendance, and skill assessments.',
 ARRAY['ISO 27001 Lead Auditor certificate', 'GDPR training completion', 'First aid certification'],
 c_internal, 60, true),

-- ============================================================
-- DATA CATEGORIES — Technical Data
-- ============================================================

(v_org, 'System Logs', 'technical', false,
 'Application logs, server logs, access logs, and error logs that may contain user identifiers.',
 ARRAY['Apache access logs', 'Application error traces', 'Authentication logs'],
 c_internal, 12, true),

(v_org, 'Authentication Credentials', 'technical', false,
 'Usernames, password hashes, API keys, tokens, and multi-factor authentication data.',
 ARRAY['Password hash (bcrypt)', 'OAuth tokens', 'MFA seed values'],
 c_restricted, 3, true),

-- ============================================================
-- DATA CATEGORIES — Public / Business
-- ============================================================

(v_org, 'Business Contact Information', 'business', false,
 'Work email addresses, office phone numbers, job titles used in a professional context.',
 ARRAY['CEO name and office email on website', 'Sales team contact directory'],
 c_internal, 36, true),

(v_org, 'Published Company Information', 'public', false,
 'Information already in the public domain: press releases, annual reports, marketing materials.',
 ARRAY['Annual report', 'Press release', 'Product brochure'],
 c_public, NULL, true),

(v_org, 'Trade Secrets', 'proprietary', false,
 'Proprietary formulas, algorithms, processes, designs, or methods that provide competitive advantage.',
 ARRAY['Algorithm source code', 'Manufacturing process', 'Pricing model'],
 c_top_secret, NULL, true),

(v_org, 'Customer Lists', 'proprietary', false,
 'Lists of customers, client details, and associated purchase or contract history.',
 ARRAY['CRM customer database', 'Client contact list', 'Sales pipeline data'],
 c_confidential, 60, true);

END $$;
