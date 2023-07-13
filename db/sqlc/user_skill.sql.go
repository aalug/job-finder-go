// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: user_skill.sql

package db

import (
	"context"

	"github.com/lib/pq"
)

const createUserSkill = `-- name: CreateUserSkill :one
INSERT INTO user_skills (user_id, skill, experience)
VALUES ($1, $2, $3)
RETURNING id, user_id, skill, experience
`

type CreateUserSkillParams struct {
	UserID     int32  `json:"user_id"`
	Skill      string `json:"skill"`
	Experience int32  `json:"experience"`
}

func (q *Queries) CreateUserSkill(ctx context.Context, arg CreateUserSkillParams) (UserSkill, error) {
	row := q.db.QueryRowContext(ctx, createUserSkill, arg.UserID, arg.Skill, arg.Experience)
	var i UserSkill
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Skill,
		&i.Experience,
	)
	return i, err
}

const deleteAllUserSkills = `-- name: DeleteAllUserSkills :exec
DELETE
FROM user_skills
WHERE user_id = $1
`

func (q *Queries) DeleteAllUserSkills(ctx context.Context, userID int32) error {
	_, err := q.db.ExecContext(ctx, deleteAllUserSkills, userID)
	return err
}

const deleteMultipleUserSkills = `-- name: DeleteMultipleUserSkills :exec
DELETE
FROM user_skills
WHERE id = ANY($1::int[])
`

func (q *Queries) DeleteMultipleUserSkills(ctx context.Context, ids []int32) error {
	_, err := q.db.ExecContext(ctx, deleteMultipleUserSkills, pq.Array(ids))
	return err
}

const deleteUserSkill = `-- name: DeleteUserSkill :exec
DELETE
FROM user_skills
WHERE id = $1
`

func (q *Queries) DeleteUserSkill(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteUserSkill, id)
	return err
}

const listUserSkills = `-- name: ListUserSkills :many
SELECT id, user_id, skill, experience
FROM user_skills
WHERE user_id = $1
ORDER BY skill
LIMIT $2 OFFSET $3
`

type ListUserSkillsParams struct {
	UserID int32 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListUserSkills(ctx context.Context, arg ListUserSkillsParams) ([]UserSkill, error) {
	rows, err := q.db.QueryContext(ctx, listUserSkills, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UserSkill{}
	for rows.Next() {
		var i UserSkill
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Skill,
			&i.Experience,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUsersBySkill = `-- name: ListUsersBySkill :many
SELECT u.id, u.full_name, u.email, u.hashed_password, u.location, u.desired_job_title, u.desired_industry, u.desired_salary_min, u.desired_salary_max, u.skills, u.experience, u.created_at
FROM users u
JOIN user_skills us ON u.id = us.user_id
WHERE us.skill = $1
ORDER BY us.experience DESC
LIMIT $2 OFFSET $3
`

type ListUsersBySkillParams struct {
	Skill  string `json:"skill"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

func (q *Queries) ListUsersBySkill(ctx context.Context, arg ListUsersBySkillParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listUsersBySkill, arg.Skill, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.FullName,
			&i.Email,
			&i.HashedPassword,
			&i.Location,
			&i.DesiredJobTitle,
			&i.DesiredIndustry,
			&i.DesiredSalaryMin,
			&i.DesiredSalaryMax,
			&i.Skills,
			&i.Experience,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUserSkill = `-- name: UpdateUserSkill :one
UPDATE user_skills
SET experience = $2
WHERE id = $1
RETURNING id, user_id, skill, experience
`

type UpdateUserSkillParams struct {
	ID         int32 `json:"id"`
	Experience int32 `json:"experience"`
}

func (q *Queries) UpdateUserSkill(ctx context.Context, arg UpdateUserSkillParams) (UserSkill, error) {
	row := q.db.QueryRowContext(ctx, updateUserSkill, arg.ID, arg.Experience)
	var i UserSkill
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Skill,
		&i.Experience,
	)
	return i, err
}
