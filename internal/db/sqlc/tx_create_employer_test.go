package db

import (
	"context"
	"github.com/aalug/job-finder-go/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLStore_CreateEmployerTx(t *testing.T) {
	company := createRandomCompany(t, "")
	params := CreateEmployerTxParams{
		CreateEmployerParams: CreateEmployerParams{
			FullName:       utils.RandomString(4),
			Email:          utils.RandomEmail(),
			HashedPassword: utils.RandomString(6),
			CompanyID:      company.ID,
		},
		AfterCreate: func(employer Employer) error {
			return nil
		},
	}

	store := NewStore(testDB)
	result, err := store.CreateEmployerTx(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, params.CreateEmployerParams.FullName, result.Employer.FullName)
	require.Equal(t, params.CreateEmployerParams.Email, result.Employer.Email)
	require.Equal(t, params.CreateEmployerParams.CompanyID, result.Employer.CompanyID)
	require.NotZero(t, result.Employer.ID)
	require.NotZero(t, result.Employer.CreatedAt)
}
