package db

import (
	"context"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func createVerifyEmailForUser(t *testing.T) VerifyEmail {
	user := createRandomUser(t)
	params := CreateVerifyEmailParams{
		Email:      user.Email,
		SecretCode: utils.RandomString(32),
	}

	verifyEmail, err := testQueries.CreateVerifyEmail(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, verifyEmail.Email, user.Email)
	require.Equal(t, verifyEmail.SecretCode, params.SecretCode)
	require.NotZero(t, verifyEmail.ID)
	require.NotZero(t, verifyEmail.CreatedAt)
	require.NotZero(t, verifyEmail.ExpiredAt)
	require.True(t, verifyEmail.ExpiredAt.After(verifyEmail.CreatedAt))

	return verifyEmail
}

func createVerifyEmailForEmployer(t *testing.T) VerifyEmail {
	employer := createRandomEmployer(t, 0)
	params := CreateVerifyEmailParams{
		Email:      employer.Email,
		SecretCode: utils.RandomString(32),
	}

	verifyEmail, err := testQueries.CreateVerifyEmail(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, verifyEmail.Email, employer.Email)
	require.Equal(t, verifyEmail.SecretCode, params.SecretCode)
	require.NotZero(t, verifyEmail.ID)
	require.NotZero(t, verifyEmail.CreatedAt)
	require.NotZero(t, verifyEmail.ExpiredAt)
	require.True(t, verifyEmail.ExpiredAt.After(verifyEmail.CreatedAt))

	return verifyEmail
}

func TestQueries_CreateVerifyEmail(t *testing.T) {
	createVerifyEmailForUser(t)
	createVerifyEmailForEmployer(t)
}

func TestQueries_UpdateVerifyEmail(t *testing.T) {
	verifyEmail := createVerifyEmailForUser(t)
	params := UpdateVerifyEmailParams{
		ID:         verifyEmail.ID,
		SecretCode: verifyEmail.SecretCode,
	}

	updatedVerifyEmail, err := testQueries.UpdateVerifyEmail(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, updatedVerifyEmail.ID, verifyEmail.ID)
	require.Equal(t, updatedVerifyEmail.Email, verifyEmail.Email)
	require.Equal(t, updatedVerifyEmail.SecretCode, verifyEmail.SecretCode)
	require.True(t, updatedVerifyEmail.IsUsed)
}
