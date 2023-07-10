package utils

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestPassword(t *testing.T) {
	// CASE 1 -when password is correct
	password := RandomString(10)

	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	// CASE 2 - when password is incorrect
	wrongPassword := "wrongPassword"
	err = CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	// CASE 3 - hashing password will return different value each time
	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEqual(t, hashedPassword1, hashedPassword2)
}
