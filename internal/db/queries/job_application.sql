-- name: CreateJobApplication :one
INSERT INTO job_applications (user_id, job_id, message, cv)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- this function will be used by users only
-- name: GetJobApplicationForUser :one
SELECT ja.id         AS application_id,
       j.id          AS job_id,
       j.title       AS job_title,
       c.name        AS company_name,
       ja.status     AS application_status,
       ja.applied_at AS application_date,
       ja.message    AS application_message,
       ja.cv         AS user_cv,
       ja.user_id    AS user_id
FROM job_applications ja
         JOIN jobs j ON ja.job_id = j.id
         JOIN companies c ON j.company_id = c.id
WHERE ja.id = $1;

-- this function will be used by employers
-- name: GetJobApplicationForEmployer :one
SELECT ja.id         AS application_id,
       j.title       AS job_title,
       j.id          AS job_id,
       ja.status     AS application_status,
       ja.applied_at AS application_date,
       ja.message    AS application_message,
       ja.cv         AS user_cv,
       ja.user_id    AS user_id,
       u.email       AS user_email,
       u.full_name   AS user_full_name,
       u.location    AS user_location,
       c.id          AS company_id
FROM job_applications ja
         JOIN jobs j ON ja.job_id = j.id
         JOIN companies c ON j.company_id = c.id
         JOIN users u ON ja.user_id = u.id
WHERE ja.id = $1;

-- name: ListJobApplicationsForUser :many
SELECT ja.user_id    AS user_id,
       ja.id         AS application_id,
       j.title       AS job_title,
       j.id          AS job_id,
       c.name        AS company_name,
       ja.status     AS application_status,
       ja.applied_at AS application_date
FROM job_applications ja
         JOIN jobs j ON ja.job_id = j.id
         JOIN companies c ON j.company_id = c.id
WHERE ja.user_id = $1
  AND (@filter_status::bool = TRUE AND ja.status = @status OR @filter_status::bool = FALSE)
ORDER BY CASE WHEN @applied_at_asc::bool THEN ja.applied_at END ASC,
         CASE WHEN @applied_at_desc::bool THEN ja.applied_at END DESC,
         ja.applied_at DESC
LIMIT $2 OFFSET $3;

-- name: ListJobApplicationsForEmployer :many
SELECT ja.id         AS application_id,
       ja.user_id    AS user_id,
       u.email       AS user_email,
       u.full_name   AS user_full_name,
       ja.status     AS application_status,
       ja.applied_at AS application_date
FROM job_applications ja
         JOIN users u ON u.id = ja.user_id
WHERE ja.job_id = $1
  AND (@filter_status::bool = TRUE AND ja.status = @status OR @filter_status::bool = FALSE)
ORDER BY CASE WHEN @applied_at_asc::bool THEN ja.applied_at END ASC,
         CASE WHEN @applied_at_desc::bool THEN ja.applied_at END DESC,
         ja.applied_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateJobApplication :one
UPDATE job_applications
SET message = COALESCE($2, message),
    cv      = COALESCE($3, cv)
WHERE id = $1
RETURNING *;

-- name: UpdateJobApplicationStatus :exec
UPDATE job_applications
SET status = $2
WHERE id = $1;

-- name: DeleteJobApplication :exec
DELETE
FROM job_applications
WHERE id = $1;

-- name: GetJobApplicationUserIDAndStatus :one
SELECT user_id, status
FROM job_applications
WHERE id = $1;

-- name: GetJobApplicationUserID :one
SELECT user_id
FROM job_applications
WHERE id = $1;

-- name: GetJobIDOfJobApplication :one
SELECT job_id
FROM job_applications
WHERE id = $1;




