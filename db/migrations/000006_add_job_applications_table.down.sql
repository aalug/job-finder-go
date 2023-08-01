ALTER TABLE job_applications
    DROP CONSTRAINT IF EXISTS job_applications_status_check;
DROP TABLE IF EXISTS job_applications;
DROP TYPE IF EXISTS application_status;