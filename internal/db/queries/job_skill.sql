-- name: CreateJobSkill :one
INSERT INTO job_skills (job_id, skill)
VALUES ($1, $2)
RETURNING *;

-- name: ListJobSkillsByJobID :many
SELECT id, skill
FROM job_skills
WHERE job_id = $1
LIMIT $2 OFFSET $3;

-- name: ListAllJobSkillsByJobID :many
SELECT skill
FROM job_skills
WHERE job_id = $1;

-- name: ListJobsBySkill :many
SELECT job_id
FROM job_skills
WHERE skill = $1
LIMIT $2 OFFSET $3;

-- name: UpdateJobSkill :one
UPDATE job_skills
SET skill = $2
WHERE id = $1
RETURNING *;

-- name: DeleteJobSkill :exec
DELETE
FROM job_skills
WHERE id = $1;

-- name: DeleteMultipleJobSkills :exec
DELETE
FROM job_skills
WHERE id = ANY (@IDs::int[]);

-- name: DeleteJobSkillsByJobID :exec
DELETE
FROM job_skills
WHERE job_id = $1;