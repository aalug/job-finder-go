// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: user.sql

package db

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (full_name, email, hashed_password, location, desired_job_title, desired_industry, desired_salary_min,
                   desired_salary_max, skills, experience)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, full_name, email, hashed_password, location, desired_job_title, desired_industry, desired_salary_min, desired_salary_max, skills, experience, created_at, is_email_verified
`

type CreateUserParams struct {
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	HashedPassword   string `json:"hashed_password"`
	Location         string `json:"location"`
	DesiredJobTitle  string `json:"desired_job_title"`
	DesiredIndustry  string `json:"desired_industry"`
	DesiredSalaryMin int32  `json:"desired_salary_min"`
	DesiredSalaryMax int32  `json:"desired_salary_max"`
	Skills           string `json:"skills"`
	Experience       string `json:"experience"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.FullName,
		arg.Email,
		arg.HashedPassword,
		arg.Location,
		arg.DesiredJobTitle,
		arg.DesiredIndustry,
		arg.DesiredSalaryMin,
		arg.DesiredSalaryMax,
		arg.Skills,
		arg.Experience,
	)
	var i User
	err := row.Scan(
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
		&i.IsEmailVerified,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteUser, id)
	return err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, full_name, email, hashed_password, location, desired_job_title, desired_industry, desired_salary_min, desired_salary_max, skills, experience, created_at, is_email_verified
FROM users
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
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
		&i.IsEmailVerified,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, full_name, email, hashed_password, location, desired_job_title, desired_industry, desired_salary_min, desired_salary_max, skills, experience, created_at, is_email_verified
FROM users
WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(
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
		&i.IsEmailVerified,
	)
	return i, err
}

const updatePassword = `-- name: UpdatePassword :exec
UPDATE users
SET hashed_password = $2
WHERE id = $1
`

type UpdatePasswordParams struct {
	ID             int32  `json:"id"`
	HashedPassword string `json:"hashed_password"`
}

func (q *Queries) UpdatePassword(ctx context.Context, arg UpdatePasswordParams) error {
	_, err := q.db.ExecContext(ctx, updatePassword, arg.ID, arg.HashedPassword)
	return err
}

const updateUser = `-- name: UpdateUser :one
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
RETURNING id, full_name, email, hashed_password, location, desired_job_title, desired_industry, desired_salary_min, desired_salary_max, skills, experience, created_at, is_email_verified
`

type UpdateUserParams struct {
	ID               int32  `json:"id"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	Location         string `json:"location"`
	DesiredJobTitle  string `json:"desired_job_title"`
	DesiredIndustry  string `json:"desired_industry"`
	DesiredSalaryMin int32  `json:"desired_salary_min"`
	DesiredSalaryMax int32  `json:"desired_salary_max"`
	Skills           string `json:"skills"`
	Experience       string `json:"experience"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUser,
		arg.ID,
		arg.FullName,
		arg.Email,
		arg.Location,
		arg.DesiredJobTitle,
		arg.DesiredIndustry,
		arg.DesiredSalaryMin,
		arg.DesiredSalaryMax,
		arg.Skills,
		arg.Experience,
	)
	var i User
	err := row.Scan(
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
		&i.IsEmailVerified,
	)
	return i, err
}

const verifyUserEmail = `-- name: VerifyUserEmail :one
UPDATE users
SET is_email_verified = TRUE
WHERE email = $1
RETURNING id, full_name, email, hashed_password, location, desired_job_title, desired_industry, desired_salary_min, desired_salary_max, skills, experience, created_at, is_email_verified
`

func (q *Queries) VerifyUserEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, verifyUserEmail, email)
	var i User
	err := row.Scan(
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
		&i.IsEmailVerified,
	)
	return i, err
}
