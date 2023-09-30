package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"github.com/aalug/job-finder-go/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLStore_CreateJobApplicationTx(t *testing.T) {
	user := createRandomUser(t)
	company := createRandomCompany(t, "")
	job := createRandomJob(t, &company, jobDetails{})

	// Generate random fake file data (e.g., 10KB size)
	fakeFileSize := 10 * 1024
	fakeFileData := make([]byte, fakeFileSize)
	_, err := rand.Read(fakeFileData)
	require.NoError(t, err)

	params := CreateJobApplicationTxParams{
		CreateJobApplicationParams: CreateJobApplicationParams{
			UserID: user.ID,
			JobID:  job.ID,
			Cv:     fakeFileData,
			Message: sql.NullString{
				String: utils.RandomString(3),
				Valid:  true,
			},
		},
		AfterCreate: func(jobApplication JobApplication) error {
			return nil
		},
	}

	store := NewStore(testDB)
	result, err := store.CreateJobApplicationTx(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, result.JobApplication.JobID, job.ID)
	require.Equal(t, result.JobApplication.UserID, user.ID)
	require.Equal(t, result.JobApplication.Cv, fakeFileData)
	require.NotEmpty(t, result.JobApplication.ID)
}
