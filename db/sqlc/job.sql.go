// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: job.sql

package db

import (
	"context"
	"time"
)

const createJob = `-- name: CreateJob :one
INSERT INTO jobs (title,
                  industry,
                  company_id,
                  description,
                  location,
                  salary_min,
                  salary_max,
                  requirements)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, title, industry, company_id, description, location, salary_min, salary_max, requirements, created_at
`

type CreateJobParams struct {
	Title        string `json:"title"`
	Industry     string `json:"industry"`
	CompanyID    int32  `json:"company_id"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	SalaryMin    int32  `json:"salary_min"`
	SalaryMax    int32  `json:"salary_max"`
	Requirements string `json:"requirements"`
}

func (q *Queries) CreateJob(ctx context.Context, arg CreateJobParams) (Job, error) {
	row := q.db.QueryRowContext(ctx, createJob,
		arg.Title,
		arg.Industry,
		arg.CompanyID,
		arg.Description,
		arg.Location,
		arg.SalaryMin,
		arg.SalaryMax,
		arg.Requirements,
	)
	var i Job
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Industry,
		&i.CompanyID,
		&i.Description,
		&i.Location,
		&i.SalaryMin,
		&i.SalaryMax,
		&i.Requirements,
		&i.CreatedAt,
	)
	return i, err
}

const deleteJob = `-- name: DeleteJob :exec
DELETE
FROM jobs
WHERE id = $1
`

func (q *Queries) DeleteJob(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteJob, id)
	return err
}

const getJob = `-- name: GetJob :one
SELECT id, title, industry, company_id, description, location, salary_min, salary_max, requirements, created_at
FROM jobs
WHERE id = $1
`

func (q *Queries) GetJob(ctx context.Context, id int32) (Job, error) {
	row := q.db.QueryRowContext(ctx, getJob, id)
	var i Job
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Industry,
		&i.CompanyID,
		&i.Description,
		&i.Location,
		&i.SalaryMin,
		&i.SalaryMax,
		&i.Requirements,
		&i.CreatedAt,
	)
	return i, err
}

const getJobDetails = `-- name: GetJobDetails :one
SELECT j.id, j.title, j.industry, j.company_id, j.description, j.location, j.salary_min, j.salary_max, j.requirements, j.created_at,
       c.name      AS company_name,
       c.location  AS company_location,
       c.industry  AS company_industry,
       e.id        AS employer_id,
       e.email     AS employer_email,
       e.full_name AS employer_full_name
FROM jobs j
         JOIN companies c ON j.company_id = c.id
         JOIN employers e ON c.id = e.company_id
WHERE j.id = $1
`

type GetJobDetailsRow struct {
	ID               int32     `json:"id"`
	Title            string    `json:"title"`
	Industry         string    `json:"industry"`
	CompanyID        int32     `json:"company_id"`
	Description      string    `json:"description"`
	Location         string    `json:"location"`
	SalaryMin        int32     `json:"salary_min"`
	SalaryMax        int32     `json:"salary_max"`
	Requirements     string    `json:"requirements"`
	CreatedAt        time.Time `json:"created_at"`
	CompanyName      string    `json:"company_name"`
	CompanyLocation  string    `json:"company_location"`
	CompanyIndustry  string    `json:"company_industry"`
	EmployerID       int32     `json:"employer_id"`
	EmployerEmail    string    `json:"employer_email"`
	EmployerFullName string    `json:"employer_full_name"`
}

func (q *Queries) GetJobDetails(ctx context.Context, id int32) (GetJobDetailsRow, error) {
	row := q.db.QueryRowContext(ctx, getJobDetails, id)
	var i GetJobDetailsRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Industry,
		&i.CompanyID,
		&i.Description,
		&i.Location,
		&i.SalaryMin,
		&i.SalaryMax,
		&i.Requirements,
		&i.CreatedAt,
		&i.CompanyName,
		&i.CompanyLocation,
		&i.CompanyIndustry,
		&i.EmployerID,
		&i.EmployerEmail,
		&i.EmployerFullName,
	)
	return i, err
}

const listAllJobsForES = `-- name: ListAllJobsForES :many
SELECT j.id,
       j.title,
       j.industry,
       j.location,
       j.description,
       c.name AS company_name,
       j.salary_min,
       j.salary_max,
       j.requirements
FROM jobs j
         JOIN companies c ON j.company_id = c.id
`

type ListAllJobsForESRow struct {
	ID           int32  `json:"id"`
	Title        string `json:"title"`
	Industry     string `json:"industry"`
	Location     string `json:"location"`
	Description  string `json:"description"`
	CompanyName  string `json:"company_name"`
	SalaryMin    int32  `json:"salary_min"`
	SalaryMax    int32  `json:"salary_max"`
	Requirements string `json:"requirements"`
}

func (q *Queries) ListAllJobsForES(ctx context.Context) ([]ListAllJobsForESRow, error) {
	rows, err := q.db.QueryContext(ctx, listAllJobsForES)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListAllJobsForESRow{}
	for rows.Next() {
		var i ListAllJobsForESRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.Location,
			&i.Description,
			&i.CompanyName,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
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

const listJobsByCompanyExactName = `-- name: ListJobsByCompanyExactName :many
SELECT j.id, j.title, j.industry, j.company_id, j.description, j.location, j.salary_min, j.salary_max, j.requirements, j.created_at,
       c.name AS company_name
FROM jobs j
         JOIN companies c ON j.company_id = c.id
WHERE c.name = $1
LIMIT $2 OFFSET $3
`

type ListJobsByCompanyExactNameParams struct {
	Name   string `json:"name"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

type ListJobsByCompanyExactNameRow struct {
	ID           int32     `json:"id"`
	Title        string    `json:"title"`
	Industry     string    `json:"industry"`
	CompanyID    int32     `json:"company_id"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	SalaryMin    int32     `json:"salary_min"`
	SalaryMax    int32     `json:"salary_max"`
	Requirements string    `json:"requirements"`
	CreatedAt    time.Time `json:"created_at"`
	CompanyName  string    `json:"company_name"`
}

func (q *Queries) ListJobsByCompanyExactName(ctx context.Context, arg ListJobsByCompanyExactNameParams) ([]ListJobsByCompanyExactNameRow, error) {
	rows, err := q.db.QueryContext(ctx, listJobsByCompanyExactName, arg.Name, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListJobsByCompanyExactNameRow{}
	for rows.Next() {
		var i ListJobsByCompanyExactNameRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
			&i.CreatedAt,
			&i.CompanyName,
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

const listJobsByCompanyID = `-- name: ListJobsByCompanyID :many
SELECT j.id, j.title, j.industry, j.company_id, j.description, j.location, j.salary_min, j.salary_max, j.requirements, j.created_at,
       c.name AS company_name
FROM jobs j
         JOIN companies c ON j.company_id = c.id
WHERE j.company_id = $1
LIMIT $2 OFFSET $3
`

type ListJobsByCompanyIDParams struct {
	CompanyID int32 `json:"company_id"`
	Limit     int32 `json:"limit"`
	Offset    int32 `json:"offset"`
}

type ListJobsByCompanyIDRow struct {
	ID           int32     `json:"id"`
	Title        string    `json:"title"`
	Industry     string    `json:"industry"`
	CompanyID    int32     `json:"company_id"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	SalaryMin    int32     `json:"salary_min"`
	SalaryMax    int32     `json:"salary_max"`
	Requirements string    `json:"requirements"`
	CreatedAt    time.Time `json:"created_at"`
	CompanyName  string    `json:"company_name"`
}

func (q *Queries) ListJobsByCompanyID(ctx context.Context, arg ListJobsByCompanyIDParams) ([]ListJobsByCompanyIDRow, error) {
	rows, err := q.db.QueryContext(ctx, listJobsByCompanyID, arg.CompanyID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListJobsByCompanyIDRow{}
	for rows.Next() {
		var i ListJobsByCompanyIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
			&i.CreatedAt,
			&i.CompanyName,
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

const listJobsByCompanyName = `-- name: ListJobsByCompanyName :many
SELECT j.id, j.title, j.industry, j.company_id, j.description, j.location, j.salary_min, j.salary_max, j.requirements, j.created_at,
       c.name AS company_name
FROM jobs j
         JOIN companies c ON j.company_id = c.id
WHERE c.name ILIKE '%' || $3::text || '%'
LIMIT $1 OFFSET $2
`

type ListJobsByCompanyNameParams struct {
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
	Name   string `json:"name"`
}

type ListJobsByCompanyNameRow struct {
	ID           int32     `json:"id"`
	Title        string    `json:"title"`
	Industry     string    `json:"industry"`
	CompanyID    int32     `json:"company_id"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	SalaryMin    int32     `json:"salary_min"`
	SalaryMax    int32     `json:"salary_max"`
	Requirements string    `json:"requirements"`
	CreatedAt    time.Time `json:"created_at"`
	CompanyName  string    `json:"company_name"`
}

func (q *Queries) ListJobsByCompanyName(ctx context.Context, arg ListJobsByCompanyNameParams) ([]ListJobsByCompanyNameRow, error) {
	rows, err := q.db.QueryContext(ctx, listJobsByCompanyName, arg.Limit, arg.Offset, arg.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListJobsByCompanyNameRow{}
	for rows.Next() {
		var i ListJobsByCompanyNameRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
			&i.CreatedAt,
			&i.CompanyName,
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

const listJobsByIndustry = `-- name: ListJobsByIndustry :many
SELECT id, title, industry, company_id, description, location, salary_min, salary_max, requirements, created_at
FROM jobs
WHERE industry = $1
LIMIT $2 OFFSET $3
`

type ListJobsByIndustryParams struct {
	Industry string `json:"industry"`
	Limit    int32  `json:"limit"`
	Offset   int32  `json:"offset"`
}

func (q *Queries) ListJobsByIndustry(ctx context.Context, arg ListJobsByIndustryParams) ([]Job, error) {
	rows, err := q.db.QueryContext(ctx, listJobsByIndustry, arg.Industry, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Job{}
	for rows.Next() {
		var i Job
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
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

const listJobsByLocation = `-- name: ListJobsByLocation :many
SELECT id, title, industry, company_id, description, location, salary_min, salary_max, requirements, created_at
FROM jobs
WHERE location = $1
LIMIT $2 OFFSET $3
`

type ListJobsByLocationParams struct {
	Location string `json:"location"`
	Limit    int32  `json:"limit"`
	Offset   int32  `json:"offset"`
}

func (q *Queries) ListJobsByLocation(ctx context.Context, arg ListJobsByLocationParams) ([]Job, error) {
	rows, err := q.db.QueryContext(ctx, listJobsByLocation, arg.Location, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Job{}
	for rows.Next() {
		var i Job
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
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

const listJobsBySalaryRange = `-- name: ListJobsBySalaryRange :many
SELECT id, title, industry, company_id, description, location, salary_min, salary_max, requirements, created_at
FROM jobs
WHERE salary_min >= $1
  AND salary_max <= $2
LIMIT $3 OFFSET $4
`

type ListJobsBySalaryRangeParams struct {
	SalaryMin int32 `json:"salary_min"`
	SalaryMax int32 `json:"salary_max"`
	Limit     int32 `json:"limit"`
	Offset    int32 `json:"offset"`
}

func (q *Queries) ListJobsBySalaryRange(ctx context.Context, arg ListJobsBySalaryRangeParams) ([]Job, error) {
	rows, err := q.db.QueryContext(ctx, listJobsBySalaryRange,
		arg.SalaryMin,
		arg.SalaryMax,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Job{}
	for rows.Next() {
		var i Job
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
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

const listJobsByTitle = `-- name: ListJobsByTitle :many
SELECT id, title, industry, company_id, description, location, salary_min, salary_max, requirements, created_at
FROM jobs
WHERE title ILIKE '%' || $3::text || '%'
LIMIT $1 OFFSET $2
`

type ListJobsByTitleParams struct {
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
	Title  string `json:"title"`
}

func (q *Queries) ListJobsByTitle(ctx context.Context, arg ListJobsByTitleParams) ([]Job, error) {
	rows, err := q.db.QueryContext(ctx, listJobsByTitle, arg.Limit, arg.Offset, arg.Title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Job{}
	for rows.Next() {
		var i Job
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
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

const listJobsMatchingUserSkills = `-- name: ListJobsMatchingUserSkills :many
SELECT j.id, j.title, j.industry, j.company_id, j.description, j.location, j.salary_min, j.salary_max, j.requirements, j.created_at,
       c.name AS company_name
FROM jobs j
         JOIN companies c ON j.company_id = c.id
WHERE j.id IN (SELECT job_id
               FROM job_skills
               WHERE skill IN (SELECT skill
                               FROM user_skills
                               WHERE user_id = $1))
LIMIT $2 OFFSET $3
`

type ListJobsMatchingUserSkillsParams struct {
	UserID int32 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListJobsMatchingUserSkillsRow struct {
	ID           int32     `json:"id"`
	Title        string    `json:"title"`
	Industry     string    `json:"industry"`
	CompanyID    int32     `json:"company_id"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	SalaryMin    int32     `json:"salary_min"`
	SalaryMax    int32     `json:"salary_max"`
	Requirements string    `json:"requirements"`
	CreatedAt    time.Time `json:"created_at"`
	CompanyName  string    `json:"company_name"`
}

func (q *Queries) ListJobsMatchingUserSkills(ctx context.Context, arg ListJobsMatchingUserSkillsParams) ([]ListJobsMatchingUserSkillsRow, error) {
	rows, err := q.db.QueryContext(ctx, listJobsMatchingUserSkills, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListJobsMatchingUserSkillsRow{}
	for rows.Next() {
		var i ListJobsMatchingUserSkillsRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Industry,
			&i.CompanyID,
			&i.Description,
			&i.Location,
			&i.SalaryMin,
			&i.SalaryMax,
			&i.Requirements,
			&i.CreatedAt,
			&i.CompanyName,
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

const updateJob = `-- name: UpdateJob :one
UPDATE jobs
SET title        = $2,
    industry     = $3,
    company_id   = $4,
    description  = $5,
    location     = $6,
    salary_min   = $7,
    salary_max   = $8,
    requirements = $9
WHERE id = $1
RETURNING id, title, industry, company_id, description, location, salary_min, salary_max, requirements, created_at
`

type UpdateJobParams struct {
	ID           int32  `json:"id"`
	Title        string `json:"title"`
	Industry     string `json:"industry"`
	CompanyID    int32  `json:"company_id"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	SalaryMin    int32  `json:"salary_min"`
	SalaryMax    int32  `json:"salary_max"`
	Requirements string `json:"requirements"`
}

func (q *Queries) UpdateJob(ctx context.Context, arg UpdateJobParams) (Job, error) {
	row := q.db.QueryRowContext(ctx, updateJob,
		arg.ID,
		arg.Title,
		arg.Industry,
		arg.CompanyID,
		arg.Description,
		arg.Location,
		arg.SalaryMin,
		arg.SalaryMax,
		arg.Requirements,
	)
	var i Job
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Industry,
		&i.CompanyID,
		&i.Description,
		&i.Location,
		&i.SalaryMin,
		&i.SalaryMax,
		&i.Requirements,
		&i.CreatedAt,
	)
	return i, err
}
