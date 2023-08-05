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

-- name: DeleteEmployer :exec
DELETE
FROM employers
WHERE id = $1;
