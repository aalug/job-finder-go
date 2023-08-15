package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Store interface {
	Querier
	CreateMultipleUserSkills(ctx context.Context, arg []CreateMultipleUserSkillsParams, userID int32) ([]UserSkill, error)
	CreateMultipleJobSkills(ctx context.Context, skills []string, jobID int32) error
	DeleteJobPosting(ctx context.Context, jobID int32) error
	GetUserDetailsByEmail(ctx context.Context, email string) (User, []UserSkill, error)
	ListJobsByFilters(ctx context.Context, arg ListJobsByFiltersParams) ([]ListJobsByFiltersRow, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	ExecTx(ctx context.Context, fn func(*Queries) error) error
}

// SQLStore provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new Store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

type CreateMultipleUserSkillsParams struct {
	Skill      string
	Experience int32
}

// CreateMultipleUserSkills creates multiple user skills for a user with ID of userID
func (store SQLStore) CreateMultipleUserSkills(ctx context.Context, arg []CreateMultipleUserSkillsParams, userID int32) ([]UserSkill, error) {
	var skills []UserSkill

	for _, v := range arg {
		params := CreateUserSkillParams{
			Skill:      v.Skill,
			Experience: v.Experience,
			UserID:     userID,
		}

		skl, err := store.CreateUserSkill(ctx, params)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skl)
	}

	return skills, nil
}

func (store SQLStore) DeleteMultipleUserSkills(ctx context.Context, ids []int32) error {
	for _, id := range ids {
		err := store.DeleteUserSkill(ctx, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateMultipleJobSkills creates multiple job skills for a job with ID of jobID
func (store SQLStore) CreateMultipleJobSkills(ctx context.Context, skills []string, jobID int32) error {
	for _, skill := range skills {
		params := CreateJobSkillParams{
			Skill: skill,
			JobID: jobID,
		}

		_, err := store.CreateJobSkill(ctx, params)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteJobPosting deletes a job with all its skills
func (store SQLStore) DeleteJobPosting(ctx context.Context, jobID int32) error {
	// Delete job skills
	err := store.DeleteJobSkillsByJobID(ctx, jobID)
	if err != nil {
		return err
	}

	// Delete job
	err = store.DeleteJob(ctx, jobID)
	if err != nil {
		return err
	}

	return nil
}

// GetUserDetailsByEmail gets user details (user, user skills) by email
func (store SQLStore) GetUserDetailsByEmail(ctx context.Context, email string) (User, []UserSkill, error) {
	user, err := store.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, nil, err
	}

	params := ListUserSkillsParams{
		UserID: user.ID,
		Limit:  10,
		Offset: 0,
	}
	userSkills, err := store.ListUserSkills(ctx, params)
	if err != nil {
		return User{}, nil, err
	}

	return user, userSkills, nil
}

// This function could not be implemented using sqlc.
// Because of that, it is implemented manually.
const listJobsByFilters = `-- name: ListJobsByFilters :many
SELECT j.id, j.title, j.industry, j.company_id, j.description, j.location, j.salary_min, j.salary_max, j.requirements, j.created_at,
       c.name AS company_name
FROM jobs j
         JOIN companies c ON j.company_id = c.id
WHERE ($3::text IS NULL OR j.title ILIKE '%' || $3 || '%')
  AND ($4::text IS NULL OR j.location = $4)
  AND ($5::text IS NULL OR j.industry = $5)
  AND ($6::int IS NULL OR j.salary_min >= $6)
  AND ($7::int IS NULL OR j.salary_max <= $7)
LIMIT $1 OFFSET $2
`

type ListJobsByFiltersParams struct {
	Limit       int32          `json:"limit"`
	Offset      int32          `json:"offset"`
	Title       sql.NullString `json:"title"`
	JobLocation sql.NullString `json:"job_location"`
	Industry    sql.NullString `json:"industry"`
	SalaryMin   sql.NullInt32  `json:"salary_min"`
	SalaryMax   sql.NullInt32  `json:"salary_max"`
}

type ListJobsByFiltersRow struct {
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

func (store SQLStore) ListJobsByFilters(ctx context.Context, arg ListJobsByFiltersParams) ([]ListJobsByFiltersRow, error) {
	rows, err := store.db.QueryContext(ctx, listJobsByFilters,
		arg.Limit,
		arg.Offset,
		arg.Title,
		arg.JobLocation,
		arg.Industry,
		arg.SalaryMin,
		arg.SalaryMax,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListJobsByFiltersRow{}
	for rows.Next() {
		var i ListJobsByFiltersRow
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

// ExecTx executes a function within a database transaction
func (store SQLStore) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
