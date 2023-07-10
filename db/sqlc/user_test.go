package db

import (
	"context"
	"database/sql"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// createRandomUser create and return a random user
func createRandomUser(t *testing.T) User {
	params := CreateUserParams{
		FullName:         utils.RandomString(6),
		Email:            utils.RandomEmail(),
		HashedPassword:   utils.RandomString(6),
		Location:         utils.RandomString(6),
		DesiredJobTitle:  utils.RandomString(6),
		DesiredIndustry:  utils.RandomString(6),
		DesiredSalaryMin: utils.RandomInt(100, 110),
		DesiredSalaryMax: utils.RandomInt(11, 120),
		Skills:           utils.RandomString(6),
		Experience:       utils.RandomString(6),
	}

	user, err := testQueries.CreateUser(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, params.FullName, user.FullName)
	require.Equal(t, params.Email, user.Email)
	require.Equal(t, params.HashedPassword, user.HashedPassword)
	require.Equal(t, params.Location, user.Location)
	require.Equal(t, params.DesiredJobTitle, user.DesiredJobTitle)
	require.Equal(t, params.DesiredIndustry, user.DesiredIndustry)
	require.Equal(t, params.DesiredSalaryMin, user.DesiredSalaryMin)
	require.Equal(t, params.DesiredSalaryMax, user.DesiredSalaryMax)
	require.Equal(t, params.Skills, user.Skills)
	require.Equal(t, params.Experience, user.Experience)
	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.ID)

	return user
}

func TestQueries_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestQueries_GetUserByEmail(t *testing.T) {
	user := createRandomUser(t)
	user2, err := testQueries.GetUserByEmail(context.Background(), user.Email)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	compareTwoUsers(t, user, user2)
}

func TestQueries_GetUserByID(t *testing.T) {
	user := createRandomUser(t)
	user2, err := testQueries.GetUserByID(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	compareTwoUsers(t, user, user2)
}

func TestQueries_DeleteUser(t *testing.T) {
	user := createRandomUser(t)
	err := testQueries.DeleteUser(context.Background(), user.ID)
	require.NoError(t, err)
	user2, err := testQueries.GetUserByID(context.Background(), user.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, user2)
}

func TestQueries_UpdateUser(t *testing.T) {
	user := createRandomUser(t)
	params := UpdateUserParams{
		ID:               user.ID,
		FullName:         utils.RandomString(6),
		Email:            user.Email,
		Location:         utils.RandomString(5),
		DesiredJobTitle:  user.DesiredJobTitle,
		DesiredIndustry:  user.DesiredIndustry,
		DesiredSalaryMin: utils.RandomInt(200, 300),
		DesiredSalaryMax: utils.RandomInt(300, 400),
		Skills:           user.Skills,
		Experience:       user.Experience,
	}

	user2, err := testQueries.UpdateUser(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, params.ID, user2.ID)
	require.Equal(t, params.FullName, user2.FullName)
	require.Equal(t, params.Email, user2.Email)
	require.Equal(t, params.Location, user2.Location)
	require.Equal(t, params.DesiredJobTitle, user2.DesiredJobTitle)
	require.Equal(t, params.DesiredIndustry, user2.DesiredIndustry)
	require.Equal(t, params.DesiredSalaryMin, user2.DesiredSalaryMin)
	require.Equal(t, params.DesiredSalaryMax, user2.DesiredSalaryMax)
	require.Equal(t, params.Skills, user2.Skills)
	require.Equal(t, params.Experience, user2.Experience)
}

func TestQueries_UpdatePassword(t *testing.T) {
	user := createRandomUser(t)
	params := UpdatePasswordParams{
		ID:             user.ID,
		HashedPassword: utils.RandomString(6),
	}

	err := testQueries.UpdatePassword(context.Background(), params)
	require.NoError(t, err)
	user2, err := testQueries.GetUserByID(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, params.HashedPassword, user2.HashedPassword)
}

// compareTwoUsers compare details of two users
func compareTwoUsers(t *testing.T, user1, user2 User) {
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Location, user2.Location)
	require.Equal(t, user1.DesiredJobTitle, user2.DesiredJobTitle)
	require.Equal(t, user1.DesiredIndustry, user2.DesiredIndustry)
	require.Equal(t, user1.DesiredSalaryMin, user2.DesiredSalaryMin)
	require.Equal(t, user1.DesiredSalaryMax, user2.DesiredSalaryMax)
	require.Equal(t, user1.Skills, user2.Skills)
	require.Equal(t, user1.Experience, user2.Experience)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}
