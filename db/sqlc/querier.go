// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package db

import (
	"context"
)

type Querier interface {
	CreateCompany(ctx context.Context, arg CreateCompanyParams) (Company, error)
	CreateEmployer(ctx context.Context, arg CreateEmployerParams) (Employer, error)
	CreateJob(ctx context.Context, arg CreateJobParams) (Job, error)
	CreateJobSkill(ctx context.Context, arg CreateJobSkillParams) (JobSkill, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	CreateUserSkill(ctx context.Context, arg CreateUserSkillParams) (UserSkill, error)
	DeleteAllUserSkills(ctx context.Context, userID int32) error
	DeleteCompany(ctx context.Context, id int32) error
	DeleteEmployer(ctx context.Context, id int32) error
	DeleteJob(ctx context.Context, id int32) error
	DeleteJobSkill(ctx context.Context, id int32) error
	DeleteJobSkillsByJobID(ctx context.Context, jobID int32) error
	DeleteMultipleJobSkills(ctx context.Context, ids []int32) error
	DeleteMultipleUserSkills(ctx context.Context, ids []int32) error
	DeleteUser(ctx context.Context, id int32) error
	DeleteUserSkill(ctx context.Context, id int32) error
	GetCompanyByID(ctx context.Context, id int32) (Company, error)
	GetCompanyByName(ctx context.Context, name string) (Company, error)
	GetEmployerByEmail(ctx context.Context, email string) (Employer, error)
	GetEmployerByID(ctx context.Context, id int32) (Employer, error)
	GetJob(ctx context.Context, id int32) (Job, error)
	GetJobDetails(ctx context.Context, id int32) (GetJobDetailsRow, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id int32) (User, error)
	ListJobSkillsByJobID(ctx context.Context, arg ListJobSkillsByJobIDParams) ([]ListJobSkillsByJobIDRow, error)
	ListJobsByCompanyExactName(ctx context.Context, arg ListJobsByCompanyExactNameParams) ([]Job, error)
	ListJobsByCompanyID(ctx context.Context, arg ListJobsByCompanyIDParams) ([]Job, error)
	ListJobsByCompanyName(ctx context.Context, arg ListJobsByCompanyNameParams) ([]Job, error)
	ListJobsByIndustry(ctx context.Context, arg ListJobsByIndustryParams) ([]Job, error)
	ListJobsByLocation(ctx context.Context, arg ListJobsByLocationParams) ([]Job, error)
	ListJobsBySalaryRange(ctx context.Context, arg ListJobsBySalaryRangeParams) ([]Job, error)
	ListJobsBySkill(ctx context.Context, arg ListJobsBySkillParams) ([]int32, error)
	ListJobsByTitle(ctx context.Context, arg ListJobsByTitleParams) ([]Job, error)
	ListJobsMatchingUserSkills(ctx context.Context, arg ListJobsMatchingUserSkillsParams) ([]ListJobsMatchingUserSkillsRow, error)
	ListUserSkills(ctx context.Context, arg ListUserSkillsParams) ([]UserSkill, error)
	ListUsersBySkill(ctx context.Context, arg ListUsersBySkillParams) ([]User, error)
	UpdateCompany(ctx context.Context, arg UpdateCompanyParams) (Company, error)
	UpdateEmployer(ctx context.Context, arg UpdateEmployerParams) (Employer, error)
	UpdateEmployerPassword(ctx context.Context, arg UpdateEmployerPasswordParams) error
	UpdateJob(ctx context.Context, arg UpdateJobParams) (Job, error)
	UpdateJobSkill(ctx context.Context, arg UpdateJobSkillParams) (JobSkill, error)
	UpdatePassword(ctx context.Context, arg UpdatePasswordParams) error
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserSkill(ctx context.Context, arg UpdateUserSkillParams) (UserSkill, error)
}

var _ Querier = (*Queries)(nil)
