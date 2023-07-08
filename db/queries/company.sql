-- name: CreateCompany :one
INSERT INTO companies (name, industry, location)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCompanyByID :one
SELECT *
FROM companies
WHERE id = $1;

-- name: GetCompanyByName :one
SELECT *
FROM companies
WHERE name = $1;

-- name: UpdateCompany :one
UPDATE companies
SET name     = $2,
    industry = $3,
    location = $4
WHERE id = $1
RETURNING *;

-- name: DeleteCompany :exec
DELETE
FROM companies
WHERE id = $1;