package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLStore_VerifyUserEmailTx(t *testing.T) {
	verifyEmail := createVerifyEmailForUser(t)
	params := VerifyEmailTxParams{
		ID:         verifyEmail.ID,
		SecretCode: verifyEmail.SecretCode,
	}

	store := NewStore(testDB)
	result, err := store.VerifyUserEmailTx(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, verifyEmail.ID, result.VerifyEmail.ID)
	require.Equal(t, verifyEmail.Email, result.VerifyEmail.Email)
	require.Equal(t, verifyEmail.SecretCode, result.VerifyEmail.SecretCode)
	require.NotEmpty(t, result.User)
	require.NotEmpty(t, result.User.ID)
	require.NotEmpty(t, result.User.Email)
	require.Equal(t, verifyEmail.Email, result.User.Email)
}

func TestSQLStore_VerifyEmployerEmailTx(t *testing.T) {
	verifyEmail := createVerifyEmailForEmployer(t)
	params := VerifyEmailTxParams{
		ID:         verifyEmail.ID,
		SecretCode: verifyEmail.SecretCode,
	}

	store := NewStore(testDB)
	result, err := store.VerifyEmployerEmailTx(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, verifyEmail.ID, result.VerifyEmail.ID)
	require.Equal(t, verifyEmail.Email, result.VerifyEmail.Email)
	require.Equal(t, verifyEmail.SecretCode, result.VerifyEmail.SecretCode)
	require.NotEmpty(t, result.Employer)
	require.NotEmpty(t, result.Employer.ID)
	require.NotEmpty(t, result.Employer.Email)
	require.Equal(t, verifyEmail.Email, result.Employer.Email)
}
