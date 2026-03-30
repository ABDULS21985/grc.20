-- 000034_training_certification.down.sql
-- Rollback: Training & Certification Tracking

DROP FUNCTION IF EXISTS generate_training_ref(UUID);

DROP TABLE IF EXISTS phishing_simulation_results CASCADE;
DROP TABLE IF EXISTS phishing_simulations CASCADE;
DROP TABLE IF EXISTS professional_certifications CASCADE;
DROP TABLE IF EXISTS training_assignments CASCADE;
DROP TABLE IF EXISTS training_content CASCADE;
DROP TABLE IF EXISTS training_programmes CASCADE;

DROP TYPE IF EXISTS certification_status;
DROP TYPE IF EXISTS phishing_action;
DROP TYPE IF EXISTS phishing_difficulty;
DROP TYPE IF EXISTS phishing_simulation_status;
DROP TYPE IF EXISTS content_type;
DROP TYPE IF EXISTS assignment_status;
DROP TYPE IF EXISTS training_target_audience;
DROP TYPE IF EXISTS training_category;
DROP TYPE IF EXISTS training_programme_status;
