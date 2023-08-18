-- name: CreateVerifyEmail :one
INSERT INTO verify_emails
    (email, secret_code)
VALUES ($1, $2)
RETURNING *;