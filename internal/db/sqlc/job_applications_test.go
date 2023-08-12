package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// createRandomJobApplication create and return a random job application
func createRandomJobApplication(t *testing.T, userID, jobID int32) JobApplication {
	if userID == 0 {
		userID = createRandomUser(t).ID
	}
	if jobID == 0 {
		jobID = createRandomJob(t, nil, jobDetails{}).ID
	}

	// Generate random fake file data (e.g., 10KB size)
	fakeFileSize := 10 * 1024
	fakeFileData := make([]byte, fakeFileSize)
	_, err := rand.Read(fakeFileData)
	require.NoError(t, err)

	params := CreateJobApplicationParams{
		UserID: userID,
		JobID:  jobID,
		Message: sql.NullString{
			String: utils.RandomString(5),
			Valid:  true,
		},
		Cv: fakeFileData,
	}

	jobApplication, err := testQueries.CreateJobApplication(context.Background(), params)
	require.NoError(t, err)

	require.NotEmpty(t, jobApplication)
	require.Equal(t, params.UserID, jobApplication.UserID)
	require.Equal(t, params.JobID, jobApplication.JobID)
	require.Equal(t, params.Message.String, jobApplication.Message.String)
	require.Equal(t, params.Cv, jobApplication.Cv)
	require.Equal(t, jobApplication.Message, jobApplication.Message)
	require.NotZero(t, jobApplication.ID)
	require.NotZero(t, jobApplication.AppliedAt)
	require.Equal(t, jobApplication.Status, ApplicationStatusApplied)

	return jobApplication
}

func TestCreateJobApplication(t *testing.T) {
	createRandomJobApplication(t, 0, 0)
}

func TestQueries_GetJobApplicationForUser(t *testing.T) {
	jobApplication1 := createRandomJobApplication(t, 0, 0)
	jobApplication2, err := testQueries.GetJobApplicationForUser(context.Background(), jobApplication1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, jobApplication2)
	require.Equal(t, jobApplication1.ID, jobApplication2.ApplicationID)
	require.Equal(t, jobApplication1.UserID, jobApplication2.UserID)
	require.Equal(t, jobApplication1.JobID, jobApplication2.JobID)
	require.Equal(t, jobApplication1.Message, jobApplication2.ApplicationMessage)
	require.Equal(t, jobApplication1.Cv, jobApplication2.UserCv)
	require.Equal(t, jobApplication1.AppliedAt, jobApplication2.ApplicationDate)
	require.NotEmpty(t, jobApplication2.ApplicationStatus, jobApplication2.ApplicationStatus)
}

func TestQueries_GetJobApplicationForEmployer(t *testing.T) {
	jobApplication1 := createRandomJobApplication(t, 0, 0)
	jobApplication2, err := testQueries.GetJobApplicationForEmployer(context.Background(), jobApplication1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, jobApplication2)
	require.Equal(t, jobApplication1.ID, jobApplication2.ApplicationID)
	require.Equal(t, jobApplication1.UserID, jobApplication2.UserID)
	require.Equal(t, jobApplication1.JobID, jobApplication2.JobID)
	require.Equal(t, jobApplication1.Message, jobApplication2.ApplicationMessage)
	require.Equal(t, jobApplication1.Cv, jobApplication2.UserCv)
	require.Equal(t, jobApplication1.AppliedAt, jobApplication2.ApplicationDate)
	require.NotEmpty(t, jobApplication2.ApplicationStatus, jobApplication2.ApplicationStatus)
}

func TestQueries_DeleteJobApplication(t *testing.T) {
	jobApplication1 := createRandomJobApplication(t, 0, 0)
	err := testQueries.DeleteJobApplication(context.Background(), jobApplication1.ID)
	require.NoError(t, err)
	jobApplication2, err := testQueries.GetJobApplicationForUser(context.Background(), jobApplication1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, jobApplication2)
}

func TestQueries_ListJobApplicationsForUser(t *testing.T) {
	user := createRandomUser(t)
	for i := 0; i < 5; i++ {
		createRandomJobApplication(t, user.ID, 0)
	}

	params := ListJobApplicationsForUserParams{
		UserID:        user.ID,
		Limit:         5,
		Offset:        0,
		FilterStatus:  false,
		Status:        ApplicationStatusSeen,
		AppliedAtAsc:  false,
		AppliedAtDesc: false,
	}

	jobApplications, err := testQueries.ListJobApplicationsForUser(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobApplications, 5)
	for _, jobApplication := range jobApplications {
		require.NotEmpty(t, jobApplication)
		require.NotZero(t, jobApplication.ApplicationID)
		require.Equal(t, jobApplication.UserID, params.UserID)
		require.NotZero(t, jobApplication.ApplicationDate)
		require.NotEmpty(t, jobApplication.ApplicationStatus)
		require.Equal(t, ApplicationStatusApplied, jobApplication.ApplicationStatus)
	}

	params = ListJobApplicationsForUserParams{
		UserID:        user.ID,
		Limit:         5,
		Offset:        0,
		FilterStatus:  true,
		Status:        ApplicationStatusApplied,
		AppliedAtAsc:  true,
		AppliedAtDesc: false,
	}

	jobApplications, err = testQueries.ListJobApplicationsForUser(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobApplications, 5)
	for _, jobApplication := range jobApplications {
		require.NotEmpty(t, jobApplication)
		require.NotZero(t, jobApplication.ApplicationID)
		require.Equal(t, jobApplication.UserID, params.UserID)
		require.NotZero(t, jobApplication.ApplicationDate)
		require.Equal(t, ApplicationStatusApplied, jobApplication.ApplicationStatus)
	}
	for i := 1; i < len(jobApplications); i++ {
		require.True(t, jobApplications[i].ApplicationDate.After(jobApplications[i-1].ApplicationDate))
	}
}

func TestQueries_ListJobApplicationsForEmployer(t *testing.T) {
	job := createRandomJob(t, nil, jobDetails{})
	for i := 0; i < 5; i++ {
		createRandomJobApplication(t, 0, job.ID)
	}

	params := ListJobApplicationsForEmployerParams{
		JobID:         job.ID,
		Limit:         5,
		Offset:        0,
		FilterStatus:  false,
		Status:        ApplicationStatusApplied,
		AppliedAtAsc:  false,
		AppliedAtDesc: false,
	}

	jobApplications, err := testQueries.ListJobApplicationsForEmployer(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobApplications, 5)
	for _, jobApplication := range jobApplications {
		require.NotEmpty(t, jobApplication)
		require.NotZero(t, jobApplication.ApplicationID)
		require.NotZero(t, jobApplication.UserID)
		require.NotZero(t, jobApplication.ApplicationDate)
		require.NotEmpty(t, jobApplication.ApplicationStatus)
		require.NotEmpty(t, jobApplication.UserEmail)
		require.NotEmpty(t, jobApplication.UserFullName)
	}

	params = ListJobApplicationsForEmployerParams{
		JobID:         job.ID,
		Limit:         5,
		Offset:        0,
		FilterStatus:  true,
		Status:        ApplicationStatusApplied,
		AppliedAtAsc:  true,
		AppliedAtDesc: false,
	}

	jobApplications, err = testQueries.ListJobApplicationsForEmployer(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobApplications, 5)
	for _, jobApplication := range jobApplications {
		require.NotEmpty(t, jobApplication)
		require.NotZero(t, jobApplication.ApplicationID)
		require.NotZero(t, jobApplication.ApplicationDate)
		require.Equal(t, ApplicationStatusApplied, jobApplication.ApplicationStatus)
	}
	for i := 1; i < len(jobApplications); i++ {
		require.True(t, jobApplications[i].ApplicationDate.After(jobApplications[i-1].ApplicationDate))
	}
}

func TestQueries_UpdateJobApplication(t *testing.T) {
	jobApplication := createRandomJobApplication(t, 0, 0)
	fakeFileData := make([]byte, 5*1024)
	_, err := rand.Read(fakeFileData)
	require.NoError(t, err)
	params := UpdateJobApplicationParams{
		ID: jobApplication.ID,
		Message: sql.NullString{
			String: utils.RandomString(3),
			Valid:  true,
		},
		Cv: fakeFileData,
	}

	jobApplication2, err := testQueries.UpdateJobApplication(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, jobApplication2)
	require.Equal(t, jobApplication.ID, jobApplication2.ID)
	require.Equal(t, jobApplication.UserID, jobApplication2.UserID)
	require.Equal(t, jobApplication.JobID, jobApplication2.JobID)
	require.Equal(t, params.Message.String, jobApplication2.Message.String)
	require.Equal(t, params.Cv, jobApplication2.Cv)
	require.WithinDuration(t, jobApplication.AppliedAt, jobApplication2.AppliedAt, 1*time.Second)
}

func TestQueries_UpdateJobApplicationStatus(t *testing.T) {
	jobApplication := createRandomJobApplication(t, 0, 0)
	status := ApplicationStatusSeen
	params := UpdateJobApplicationStatusParams{
		ID:     jobApplication.ID,
		Status: status,
	}

	err := testQueries.UpdateJobApplicationStatus(context.Background(), params)
	require.NoError(t, err)
	jobApplication2, err := testQueries.GetJobApplicationForUser(context.Background(), jobApplication.ID)
	require.NoError(t, err)
	require.Equal(t, jobApplication2.ApplicationStatus, status)
}

func TestQueries_GetJobApplicationUserID(t *testing.T) {
	jobApplication := createRandomJobApplication(t, 0, 0)

	userID, err := testQueries.GetJobApplicationUserID(context.Background(), jobApplication.ID)
	require.NoError(t, err)
	require.Equal(t, userID, jobApplication.UserID)
}

func TestQueries_GetJobApplicationUserIDAndStatus(t *testing.T) {
	jobApplication := createRandomJobApplication(t, 0, 0)

	details, err := testQueries.GetJobApplicationUserIDAndStatus(context.Background(), jobApplication.ID)
	require.NoError(t, err)
	require.Equal(t, details.UserID, jobApplication.UserID)
	require.Equal(t, details.Status, jobApplication.Status)
}

func TestQueries_GetJobIDOfJobApplication(t *testing.T) {
	jobApplication := createRandomJobApplication(t, 0, 0)
	jobID, err := testQueries.GetJobIDOfJobApplication(context.Background(), jobApplication.ID)
	require.NoError(t, err)
	require.Equal(t, jobID, jobApplication.JobID)
}
