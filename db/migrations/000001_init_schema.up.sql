CREATE TABLE companies
(
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    industry TEXT NOT NULL,
    location TEXT NOT NULL
);

CREATE INDEX idx_companies_name ON companies (name);


CREATE TABLE jobs
(
    id           SERIAL PRIMARY KEY,
    title        TEXT        NOT NULL,
    industry     TEXT        NOT NULL,
    company_id   INTEGER     NOT NULL,
    description  TEXT        NOT NULL,
    location     TEXT        NOT NULL,
    salary_min   INTEGER     NOT NULL,
    salary_max   INTEGER     NOT NULL,
    requirements TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (NOW()),
    FOREIGN KEY (company_id) REFERENCES companies (id)
);

CREATE INDEX idx_jobs_title ON jobs (title);
CREATE INDEX idx_jobs_location ON jobs (location);
CREATE INDEX idx_jobs_industry ON jobs (industry);
CREATE INDEX idx_jobs_salary_range ON jobs (salary_min, salary_max);
CREATE INDEX idx_jobs_created_at ON jobs (created_at);


CREATE TABLE users
(
    id                 SERIAL PRIMARY KEY,
    full_name          TEXT        NOT NULL,
    email              TEXT        NOT NULL,
    location           TEXT        NOT NULL,
    desired_job_title  TEXT        NOT NULL,
    desired_industry   TEXT        NOT NULL,
    desired_salary_min INTEGER     NOT NULL,
    desired_salary_max INTEGER     NOT NULL,
    skills             TEXT        NOT NULL,
    experience         TEXT        NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT (NOW())
);

CREATE INDEX idx_users_email ON users (email);


CREATE TABLE job_skills
(
    job_id INTEGER NOT NULL,
    skill  TEXT    NOT NULL,
    FOREIGN KEY (job_id) REFERENCES jobs (id)
);

CREATE TABLE user_skills
(
    user_id    INTEGER NOT NULL,
    skill      TEXT    NOT NULL,
    experience INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
);