package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLStore_VerifyEmailTx(t *testing.T) {
	verifyEmail := createRandomVerifyEmail(t)
	params := VerifyEmailTxParams{
		ID:         verifyEmail.ID,
		SecretCode: verifyEmail.SecretCode,
	}

	store := NewStore(testDB)
	result, err := store.VerifyEmailTx(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, verifyEmail.ID, result.VerifyEmail.ID)
	require.Equal(t, verifyEmail.Email, result.VerifyEmail.Email)
	require.Equal(t, verifyEmail.SecretCode, result.VerifyEmail.SecretCode)
	require.NotEmpty(t, result.User)
	require.NotEmpty(t, result.User.ID)
	require.NotEmpty(t, result.User.Email)
	require.Equal(t, verifyEmail.Email, result.User.Email)
}
