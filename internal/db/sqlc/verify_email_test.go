package db

import (
	"context"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func createRandomVerifyEmail(t *testing.T) VerifyEmail {
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

func TestQueries_CreateVerifyEmail(t *testing.T) {
	createRandomUser(t)
}

func TestQueries_UpdateVerifyEmail(t *testing.T) {
	verifyEmail := createRandomVerifyEmail(t)
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
