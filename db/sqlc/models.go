// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

type ApplicationStatus string

const (
	ApplicationStatusApplied      ApplicationStatus = "Applied"
	ApplicationStatusSeen         ApplicationStatus = "Seen"
	ApplicationStatusInterviewing ApplicationStatus = "Interviewing"
	ApplicationStatusOffered      ApplicationStatus = "Offered"
	ApplicationStatusRejected     ApplicationStatus = "Rejected"
)

func (e *ApplicationStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ApplicationStatus(s)
	case string:
		*e = ApplicationStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ApplicationStatus: %T", src)
	}
	return nil
}

type NullApplicationStatus struct {
	ApplicationStatus ApplicationStatus
	Valid             bool // Valid is true if ApplicationStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullApplicationStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ApplicationStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ApplicationStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullApplicationStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ApplicationStatus), nil
}

type Company struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Industry string `json:"industry"`
	Location string `json:"location"`
}

type Employer struct {
	ID             int32     `json:"id"`
	CompanyID      int32     `json:"company_id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
	CreatedAt      time.Time `json:"created_at"`
}

type Job struct {
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
}

type JobApplication struct {
	ID        int32             `json:"id"`
	UserID    int32             `json:"user_id"`
	JobID     int32             `json:"job_id"`
	Message   sql.NullString    `json:"message"`
	Cv        []byte            `json:"cv"`
	Status    ApplicationStatus `json:"status"`
	AppliedAt time.Time         `json:"applied_at"`
}

type JobSkill struct {
	ID    int32  `json:"id"`
	JobID int32  `json:"job_id"`
	Skill string `json:"skill"`
}

type User struct {
	ID               int32     `json:"id"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	HashedPassword   string    `json:"hashed_password"`
	Location         string    `json:"location"`
	DesiredJobTitle  string    `json:"desired_job_title"`
	DesiredIndustry  string    `json:"desired_industry"`
	DesiredSalaryMin int32     `json:"desired_salary_min"`
	DesiredSalaryMax int32     `json:"desired_salary_max"`
	Skills           string    `json:"skills"`
	Experience       string    `json:"experience"`
	CreatedAt        time.Time `json:"created_at"`
}

type UserSkill struct {
	ID         int32  `json:"id"`
	UserID     int32  `json:"user_id"`
	Skill      string `json:"skill"`
	Experience int32  `json:"experience"`
}
