package esearch

import (
	"context"
	"github.com/aalug/go-gin-job-search/internal/db/mock"
	"github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadJobsFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock store
	mockStore := mockdb.NewMockStore(ctrl)

	// Define test data
	testJobsFromDB := []db.ListAllJobsForESRow{
		{
			ID:           utils.RandomInt(1, 1000),
			Title:        utils.RandomString(5),
			Industry:     utils.RandomString(2),
			Location:     utils.RandomString(2),
			Description:  utils.RandomString(4),
			CompanyName:  utils.RandomString(3),
			SalaryMin:    utils.RandomInt(100, 200),
			SalaryMax:    utils.RandomInt(201, 300),
			Requirements: utils.RandomString(2),
		},
	}

	// Mock the ListAllJobsForES function
	mockStore.EXPECT().
		ListAllJobsForES(gomock.Any()).
		Return(testJobsFromDB, nil)

	// Mock the ListAllJobSkillsByJobID function
	// You can add more specific expectations if needed for different job IDs and skills
	for _, job := range testJobsFromDB {
		mockStore.EXPECT().
			ListAllJobSkillsByJobID(gomock.Any(), gomock.Eq(job.ID)).
			Return([]string{"Skill1", "Skill2"}, nil)
	}

	// Call the function being tested
	ctx := context.Background()
	ctx, err := LoadJobsFromDB(ctx, mockStore)
	require.NoError(t, err)

	// Retrieve the loaded jobs from the context
	loadedJobs, ok := ctx.Value(JobKey).([]Job)
	if !ok {
		t.Fatalf("Expected value of type []Job, got %T", ctx.Value(JobKey))
	}

	// Check the number of jobs loaded from the database
	expectedNumJobs := len(testJobsFromDB)
	if len(loadedJobs) != expectedNumJobs {
		t.Fatalf("Expected %d jobs, got %d", expectedNumJobs, len(loadedJobs))
	}

	// Check specific attributes of the loaded jobs
	for i, expectedJob := range testJobsFromDB {
		require.Equal(t, loadedJobs[i].ID, expectedJob.ID)
		require.Equal(t, loadedJobs[i].Title, expectedJob.Title)
		require.Equal(t, loadedJobs[i].Industry, expectedJob.Industry)
		require.Equal(t, loadedJobs[i].Location, expectedJob.Location)
		require.Equal(t, loadedJobs[i].Description, expectedJob.Description)
		require.Equal(t, loadedJobs[i].CompanyName, expectedJob.CompanyName)
		require.Equal(t, loadedJobs[i].SalaryMin, expectedJob.SalaryMin)
		require.Equal(t, loadedJobs[i].SalaryMax, expectedJob.SalaryMax)
		require.Equal(t, loadedJobs[i].Requirements, expectedJob.Requirements)
		require.Equal(t, loadedJobs[i].JobSkills, []string{"Skill1", "Skill2"})
	}
}
