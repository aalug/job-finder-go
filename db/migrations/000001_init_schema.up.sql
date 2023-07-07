CREATE TABLE jobs
(
    id           SERIAL PRIMARY KEY,
    title        TEXT NOT NULL,
    industry     TEXT NOT NULL,
    description  TEXT NOT NULL,
    location     TEXT NOT NULL,
    salary_min   INTEGER,
    salary_max   INTEGER,
    requirements TEXT
);

CREATE INDEX idx_jobs_title ON jobs (title);
CREATE INDEX idx_jobs_location ON jobs (location);
CREATE INDEX idx_jobs_industry ON jobs (industry);
CREATE INDEX idx_jobs_salary_range ON jobs (salary_min, salary_max);


CREATE TABLE companies
(
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    industry TEXT NOT NULL,
    location TEXT NOT NULL
);

CREATE INDEX idx_companies_name ON companies (name);


CREATE TABLE users
(
    id                 SERIAL PRIMARY KEY,
    full_name          TEXT NOT NULL,
    email              TEXT NOT NULL,
    location           TEXT NOT NULL,
    desired_job_title  TEXT,
    desired_industry   TEXT,
    desired_salary_min INTEGER,
    desired_salary_max INTEGER,
    skills             TEXT,
    experience         TEXT
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
    experience INTEGER,
    FOREIGN KEY (user_id) REFERENCES users (id)
);