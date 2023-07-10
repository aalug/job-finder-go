-- name: CreateUserSkill :one
INSERT INTO user_skills (user_id, skill, experience)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListUserSkills :many
SELECT *
FROM user_skills
WHERE user_id = $1
ORDER BY skill
LIMIT $2 OFFSET $3;

-- name: UpdateUserSkill :one
UPDATE user_skills
SET experience = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUserSkill :exec
DELETE
FROM user_skills
WHERE id = $1;

-- name: ListUsersBySkill :many
SELECT u.*
FROM users u
JOIN user_skills us ON u.id = us.user_id
WHERE us.skill = $1
ORDER BY us.experience DESC
LIMIT $2 OFFSET $3;