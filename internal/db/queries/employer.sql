-- name: CreateEmployer :one
INSERT INTO employers (company_id, full_name, email, hashed_password)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEmployerByID :one
SELECT *
FROM employers
WHERE id = $1;

-- name: GetEmployerByEmail :one
SELECT *
FROM employers
WHERE email = $1;

-- name: UpdateEmployer :one
UPDATE employers
SET company_id = $2,
    full_name  = $3,
    email      = $4
WHERE id = $1
RETURNING *;

-- name: UpdateEmployerPassword :exec
UPDATE employers
SET hashed_password = $2
WHERE id = $1;

-- name: VerifyEmployerEmail :one
UPDATE employers
SET is_email_verified = TRUE
WHERE email = $1
RETURNING *;

-- name: DeleteEmployer :exec
DELETE
FROM employers
WHERE id = $1;

-- name: GetEmployerAndCompanyDetails :one
SELECT c.name      AS company_name,
       c.industry  AS company_industry,
       c.location  AS company_location,
       c.id        AS company_id,
       e.id        AS employer_id,
       e.full_name AS employer_full_name,
       e.email     AS employer_email
FROM employers e
         JOIN companies c ON c.id = e.company_id
WHERE e.email = $1;
