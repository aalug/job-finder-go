package db

import (
	"context"
	"database/sql"
)

type Store interface {
	Querier
	CreateMultipleUserSkills(ctx context.Context, arg []CreateMultipleUserSkillsParams, userID int32) ([]UserSkill, error)
	CreateMultipleJobSkills(ctx context.Context, skills []string, jobID int32) error
	DeleteJobPosting(ctx context.Context, jobID int32) error
	GetUserDetailsByEmail(ctx context.Context, email string) (User, []UserSkill, error)
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
