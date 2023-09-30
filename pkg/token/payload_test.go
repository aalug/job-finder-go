package token

import (
	"github.com/aalug/job-finder-go/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewPayload(t *testing.T) {
	email := utils.RandomEmail()
	duration := time.Hour

	payload, err := NewPayload(email, duration)
	require.NoError(t, err)

	// Check that the payload fields are set correctly
	require.NotEqual(t, uuid.Nil, payload.ID, "ID should not be nil")
	require.Equal(t, email, payload.Email)
	require.WithinDuration(t, time.Now(), payload.IssuedAt, 5*time.Second, "IssuedAt should be close to the current time")
	require.WithinDuration(t, time.Now().Add(duration), payload.ExpiredAt, 5*time.Second, "ExpiredAt should be close to current time + duration")
}

func TestPayload_Valid(t *testing.T) {
	// Create a payload that has not expired
	validPayload := &Payload{
		ID:        uuid.New(),
		Email:     utils.RandomEmail(),
		IssuedAt:  time.Now().Add(-time.Hour), // Issued 1 hour ago
		ExpiredAt: time.Now().Add(time.Hour),  // Expires in 1 hour
	}

	// Create a payload that has already expired
	expiredPayload := &Payload{
		ID:        uuid.New(),
		Email:     utils.RandomEmail(),
		IssuedAt:  time.Now().Add(-2 * time.Hour), // Issued 2 hours ago
		ExpiredAt: time.Now().Add(-time.Hour),     // Expired 1 hour ago
	}

	// Check that a valid payload does not return an error
	err := validPayload.Valid()
	require.NoError(t, err)

	// Check that an expired payload returns an error
	err = expiredPayload.Valid()
	require.EqualError(t, err, ErrExpiredToken.Error())
}
