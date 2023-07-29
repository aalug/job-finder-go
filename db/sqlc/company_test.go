package db

import (
	"context"
	"database/sql"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"testing"
)

// createRandomCompany get or creates and return a random company
func createRandomCompany(t *testing.T, name string) Company {
	params := CreateCompanyParams{
		Industry: utils.RandomString(4),
		Location: utils.RandomString(6),
	}
	if name != "" {
		params.Name = name
	} else {
		params.Name = utils.RandomString(6)
	}

	company, err := testQueries.CreateCompany(context.Background(), params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				c, err := testQueries.GetCompanyByName(context.Background(), params.Name)
				require.NoError(t, err)
				return c
			}
		}
	}

	require.NoError(t, err)
	require.NotEmpty(t, company)
	require.Equal(t, params.Name, company.Name)
	require.Equal(t, params.Industry, company.Industry)
	require.Equal(t, params.Location, company.Location)
	require.NotZero(t, company.ID)

	return company
}

func TestQueries_CreateCompany(t *testing.T) {
	createRandomCompany(t, "")
}

func TestQueries_GetCompanyByID(t *testing.T) {
	company := createRandomCompany(t, "")
	company2, err := testQueries.GetCompanyByID(context.Background(), company.ID)

	require.NoError(t, err)
	require.NotEmpty(t, company2)
	require.Equal(t, company.ID, company2.ID)
	require.Equal(t, company.Name, company2.Name)
	require.Equal(t, company.Industry, company2.Industry)
	require.Equal(t, company.Location, company2.Location)
}

func TestQueries_GetCompanyNameByID(t *testing.T) {
	company := createRandomCompany(t, "")
	companyName, err := testQueries.GetCompanyNameByID(context.Background(), company.ID)

	require.NoError(t, err)
	require.Equal(t, company.Name, companyName)
}

func TestQueries_GetCompanyByName(t *testing.T) {
	company := createRandomCompany(t, "")
	company2, err := testQueries.GetCompanyByName(context.Background(), company.Name)

	require.NoError(t, err)
	require.NotEmpty(t, company2)
	require.Equal(t, company.ID, company2.ID)
	require.Equal(t, company.Name, company2.Name)
	require.Equal(t, company.Industry, company2.Industry)
	require.Equal(t, company.Location, company2.Location)
}

func TestQueries_UpdateCompany(t *testing.T) {
	company := createRandomCompany(t, "")
	params := UpdateCompanyParams{
		ID:       company.ID,
		Name:     utils.RandomString(3),
		Industry: utils.RandomString(3),
		Location: utils.RandomString(4),
	}

	company2, err := testQueries.UpdateCompany(context.Background(), params)

	require.NoError(t, err)
	require.NotEmpty(t, company2)
	require.Equal(t, company.ID, company2.ID)
	require.Equal(t, params.Name, company2.Name)
	require.Equal(t, params.Industry, company2.Industry)
	require.Equal(t, params.Location, company2.Location)
}

func TestQueries_DeleteCompany(t *testing.T) {
	company := createRandomCompany(t, "")
	err := testQueries.DeleteCompany(context.Background(), company.ID)

	require.NoError(t, err)

	company2, err := testQueries.GetCompanyByID(context.Background(), company.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, company2)
}
