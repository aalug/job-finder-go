package db

import (
	"context"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLStore_CreateUserTx(t *testing.T) {
	params := CreateUserTxParams{
		CreateUserParams: CreateUserParams{
			FullName:         utils.RandomString(4),
			Email:            utils.RandomEmail(),
			HashedPassword:   utils.RandomString(6),
			Location:         utils.RandomString(3),
			DesiredJobTitle:  utils.RandomString(2),
			DesiredIndustry:  utils.RandomString(2),
			DesiredSalaryMin: utils.RandomInt(1, 100),
			DesiredSalaryMax: utils.RandomInt(101, 200),
			Skills:           utils.RandomString(5),
			Experience:       utils.RandomString(3),
		},
		AfterCreate: func(user User) error {
			return nil
		},
	}

	store := NewStore(testDB)
	result, err := store.CreateUserTx(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, params.CreateUserParams.FullName, result.User.FullName)
	require.Equal(t, params.CreateUserParams.Email, result.User.Email)
	require.Equal(t, params.CreateUserParams.Location, result.User.Location)
	require.Equal(t, params.CreateUserParams.DesiredJobTitle, result.User.DesiredJobTitle)
	require.Equal(t, params.CreateUserParams.DesiredIndustry, result.User.DesiredIndustry)
	require.Equal(t, params.CreateUserParams.DesiredSalaryMin, result.User.DesiredSalaryMin)
	require.Equal(t, params.CreateUserParams.DesiredSalaryMax, result.User.DesiredSalaryMax)
	require.Equal(t, params.CreateUserParams.Skills, result.User.Skills)
	require.Equal(t, params.CreateUserParams.Experience, result.User.Experience)
	require.NotZero(t, result.User.CreatedAt)
}
