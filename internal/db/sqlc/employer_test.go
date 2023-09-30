package db

import (
	"context"
	"database/sql"
	"github.com/aalug/job-finder-go/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// createRandomEmployer create and return a random employer
func createRandomEmployer(t *testing.T, companyID int32) Employer {
	params := CreateEmployerParams{
		FullName:       utils.RandomString(6),
		Email:          utils.RandomEmail(),
		HashedPassword: utils.RandomString(6),
	}

	if companyID == 0 {
		params.CompanyID = createRandomCompany(t, "").ID
	} else {
		params.CompanyID = companyID
	}

	employer, err := testQueries.CreateEmployer(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, employer)
	require.Equal(t, params.FullName, employer.FullName)
	require.Equal(t, params.Email, employer.Email)
	require.Equal(t, params.HashedPassword, employer.HashedPassword)
	require.Equal(t, params.CompanyID, employer.CompanyID)
	require.NotZero(t, employer.ID)
	require.NotZero(t, employer.CreatedAt)

	return employer
}

func TestQueries_CreateEmployer(t *testing.T) {
	createRandomEmployer(t, 0)
}

func TestQueries_GetEmployerByID(t *testing.T) {
	employer := createRandomEmployer(t, 0)
	employer2, err := testQueries.GetEmployerByID(context.Background(), employer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, employer2)
	compareTwoEmployers(t, employer, employer2)
}

func TestQueries_GetEmployerByEmail(t *testing.T) {
	employer := createRandomEmployer(t, 0)
	employer2, err := testQueries.GetEmployerByEmail(context.Background(), employer.Email)
	require.NoError(t, err)
	require.NotEmpty(t, employer2)
	compareTwoEmployers(t, employer, employer2)
}

func TestQueries_DeleteEmployer(t *testing.T) {
	employer := createRandomEmployer(t, 0)
	err := testQueries.DeleteEmployer(context.Background(), employer.ID)
	require.NoError(t, err)
	employer2, err := testQueries.GetEmployerByID(context.Background(), employer.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, employer2)
}

func TestQueries_UpdateEmployer(t *testing.T) {
	employer := createRandomEmployer(t, 0)
	params := UpdateEmployerParams{
		ID:        employer.ID,
		FullName:  utils.RandomString(6),
		Email:     utils.RandomEmail(),
		CompanyID: employer.CompanyID,
	}

	employer2, err := testQueries.UpdateEmployer(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, employer2)
	require.Equal(t, params.FullName, employer2.FullName)
	require.Equal(t, params.Email, employer2.Email)
	require.Equal(t, params.CompanyID, employer2.CompanyID)
	require.Equal(t, employer.ID, employer2.ID)
}

func TestQueries_UpdateEmployerPassword(t *testing.T) {
	employer := createRandomEmployer(t, 0)
	params := UpdateEmployerPasswordParams{
		ID:             employer.ID,
		HashedPassword: utils.RandomString(6),
	}
	err := testQueries.UpdateEmployerPassword(context.Background(), params)
	require.NoError(t, err)

	employer2, err := testQueries.GetEmployerByID(context.Background(), employer.ID)
	require.NoError(t, err)
	require.Equal(t, params.HashedPassword, employer2.HashedPassword)
}

func compareTwoEmployers(t *testing.T, employer1, employer2 Employer) {
	require.Equal(t, employer1.ID, employer2.ID)
	require.Equal(t, employer1.FullName, employer2.FullName)
	require.Equal(t, employer1.Email, employer2.Email)
	require.Equal(t, employer1.HashedPassword, employer2.HashedPassword)
	require.Equal(t, employer1.CompanyID, employer2.CompanyID)
	require.WithinDuration(t, employer1.CreatedAt, employer2.CreatedAt, time.Second)
}

func TestQueries_GetEmployerAndCompanyDetails(t *testing.T) {
	company := createRandomCompany(t, "")
	employer := createRandomEmployer(t, company.ID)
	details, err := testQueries.GetEmployerAndCompanyDetails(context.Background(), employer.Email)
	require.NoError(t, err)
	require.NotEmpty(t, details)
	require.Equal(t, employer.ID, details.EmployerID)
	require.Equal(t, company.ID, details.CompanyID)
	require.Equal(t, employer.CompanyID, details.CompanyID)
	require.Equal(t, company.Name, details.CompanyName)
	require.Equal(t, company.Location, details.CompanyLocation)
	require.Equal(t, company.Industry, details.CompanyIndustry)
	require.Equal(t, employer.FullName, details.EmployerFullName)
	require.Equal(t, employer.Email, details.EmployerEmail)
}
