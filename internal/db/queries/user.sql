-- name: CreateUser :one
INSERT INTO users (full_name, email, hashed_password, location, desired_job_title, desired_industry, desired_salary_min,
                   desired_salary_max, skills, experience)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET full_name          = $2,
    email              = $3,
    location           = $4,
    desired_job_title  = $5,
    desired_industry   = $6,
    desired_salary_min = $7,
    desired_salary_max = $8,
    skills             = $9,
    experience         = $10
WHERE id = $1
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
SET hashed_password = $2
WHERE id = $1;

-- name: VerifyUserEmail :one
UPDATE users
SET is_email_verified = TRUE
WHERE email = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;