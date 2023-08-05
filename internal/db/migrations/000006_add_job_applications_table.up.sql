-- Create the enum type
CREATE TYPE application_status AS ENUM ('Applied', 'Seen', 'Interviewing', 'Offered', 'Rejected');

CREATE TABLE job_applications
(
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER            NOT NULL,
    job_id     INTEGER            NOT NULL,
    message    TEXT,
    cv         BYTEA              NOT NULL,
    status     application_status NOT NULL DEFAULT 'Applied' CHECK (status IN ('Applied', 'Seen', 'Interviewing', 'Offered', 'Rejected')),
    applied_at TIMESTAMPTZ        NOT NULL DEFAULT (NOW()),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (job_id) REFERENCES jobs (id),
    CONSTRAINT unique_user_job_combination UNIQUE (user_id, job_id)
);

CREATE INDEX idx_job_applications_user_id ON job_applications (user_id);
CREATE INDEX idx_job_applications_job_id ON job_applications (job_id);
