CREATE TABLE employers
(
    id              SERIAL PRIMARY KEY,
    company_id      INTEGER     NOT NULL,
    full_name       TEXT        NOT NULL,
    email           TEXT UNIQUE NOT NULL,
    hashed_password TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT (NOW()),
    FOREIGN KEY (company_id) REFERENCES companies (id)
);

CREATE INDEX employers_email_idx ON employers (email);
