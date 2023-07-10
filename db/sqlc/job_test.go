package db

import (
	"context"
	"database/sql"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

type jobDetails struct {
	title     string
	industry  string
	location  string
	salaryMin int32
	salaryMax int32
}

// createRandomJob  creates and return a random job
func createRandomJob(t *testing.T, company *Company, details jobDetails) Job {
	var c Company
	if company == nil {
		c = createRandomCompany(t, "")
	} else {
		c = *company
	}

	params := CreateJobParams{
		CompanyID:    c.ID,
		Description:  utils.RandomString(7),
		Requirements: utils.RandomString(5),
		CreatedAt:    time.Now(),
	}

	if details.title != "" {
		params.Title = details.title
	} else {
		params.Title = utils.RandomString(5)
	}

	if details.industry != "" {
		params.Industry = details.industry
	} else {
		params.Industry = utils.RandomString(5)
	}
	if details.location != "" {
		params.Location = details.location
	} else {
		params.Location = utils.RandomString(4)
	}
	if details.salaryMin != 0 && details.salaryMax != 0 {
		params.SalaryMin = details.salaryMin
		params.SalaryMax = details.salaryMax
	} else {
		params.SalaryMin = utils.RandomInt(100, 110)
		params.SalaryMax = utils.RandomInt(100, 110)
	}

	job, err := testQueries.CreateJob(context.Background(), params)

	require.NoError(t, err)
	require.NotEmpty(t, job)
	require.NotZero(t, job.ID)
	require.Equal(t, job.Title, params.Title)
	require.Equal(t, job.Industry, params.Industry)
	require.Equal(t, job.CompanyID, params.CompanyID)
	require.Equal(t, job.Description, params.Description)
	require.Equal(t, job.Location, params.Location)
	require.Equal(t, job.SalaryMin, params.SalaryMin)
	require.Equal(t, job.SalaryMax, params.SalaryMax)
	require.Equal(t, job.Requirements, params.Requirements)
	require.WithinDuration(t, job.CreatedAt, params.CreatedAt, time.Second)

	return job
}

func TestQueries_CreateJob(t *testing.T) {
	createRandomJob(t, nil, jobDetails{})
}

func TestQueries_GetJob(t *testing.T) {
	job := createRandomJob(t, nil, jobDetails{})
	job2, err := testQueries.GetJob(context.Background(), job.ID)
	require.NoError(t, err)
	require.NotEmpty(t, job2)
	require.Equal(t, job.ID, job2.ID)
	require.Equal(t, job.Title, job2.Title)
	require.Equal(t, job.Industry, job2.Industry)
	require.Equal(t, job.CompanyID, job2.CompanyID)
	require.Equal(t, job.Description, job2.Description)
	require.Equal(t, job.Location, job2.Location)
	require.Equal(t, job.SalaryMin, job2.SalaryMin)
	require.Equal(t, job.SalaryMax, job2.SalaryMax)
	require.Equal(t, job.Requirements, job2.Requirements)
	require.WithinDuration(t, job.CreatedAt, job2.CreatedAt, time.Second)
}

func TestQueries_UpdateJob(t *testing.T) {
	job := createRandomJob(t, nil, jobDetails{})
	company := createRandomCompany(t, "")
	params := UpdateJobParams{
		ID:           job.ID,
		Title:        utils.RandomString(7),
		Industry:     utils.RandomString(7),
		CompanyID:    company.ID,
		Description:  utils.RandomString(7),
		Location:     job.Location,
		SalaryMin:    utils.RandomInt(100, 110),
		SalaryMax:    utils.RandomInt(110, 120),
		Requirements: job.Requirements,
	}

	job2, err := testQueries.UpdateJob(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, job2)
	require.Equal(t, job.ID, job2.ID)
	require.Equal(t, params.Title, job2.Title)
	require.Equal(t, params.Industry, job2.Industry)
	require.Equal(t, params.CompanyID, job2.CompanyID)
	require.Equal(t, params.Description, job2.Description)
	require.Equal(t, params.Location, job2.Location)
	require.Equal(t, params.SalaryMin, job2.SalaryMin)
	require.Equal(t, params.SalaryMax, job2.SalaryMax)
	require.Equal(t, params.Requirements, job2.Requirements)
	require.WithinDuration(t, job.CreatedAt, job2.CreatedAt, time.Second)
}

func TestQueries_DeleteJob(t *testing.T) {
	job := createRandomJob(t, nil, jobDetails{})
	err := testQueries.DeleteJob(context.Background(), job.ID)
	require.NoError(t, err)
	job2, err := testQueries.GetJob(context.Background(), job.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, job2)
}

func TestQueries_ListJobsByCompanyExactName(t *testing.T) {
	company := createRandomCompany(t, "exactName")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomJob(t, &company, jobDetails{})
		} else {
			createRandomJob(t, nil, jobDetails{})
		}
	}

	params := ListJobsByCompanyExactNameParams{
		Name:   company.Name,
		Limit:  5,
		Offset: 0,
	}

	jobs, err := testQueries.ListJobsByCompanyExactName(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, jobs, 5)
	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.Equal(t, company.ID, job.CompanyID)
	}
}

func TestQueries_ListJobsByCompanyName(t *testing.T) {
	company1 := createRandomCompany(t, "companyName")
	company2 := createRandomCompany(t, "otherName")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomJob(t, &company1, jobDetails{})
		} else {
			createRandomJob(t, &company2, jobDetails{})
		}
	}

	params := ListJobsByCompanyNameParams{
		Name:   company1.Name[0:5],
		Limit:  5,
		Offset: 0,
	}

	jobs, err := testQueries.ListJobsByCompanyName(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, jobs, 5)
	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.Equal(t, company1.ID, job.CompanyID)
	}
}

func TestQueries_ListJobsByIndustry(t *testing.T) {
	details := jobDetails{
		industry: "testIndustry",
	}
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomJob(t, nil, details)
		} else {
			createRandomJob(t, nil, jobDetails{})
		}
	}

	params := ListJobsByIndustryParams{
		Industry: details.industry,
		Limit:    5,
		Offset:   0,
	}

	jobs, err := testQueries.ListJobsByIndustry(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, jobs, 5)
	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.Equal(t, params.Industry, job.Industry)
	}
}

func TestQueries_ListJobsByCompanyID(t *testing.T) {
	company := createRandomCompany(t, "")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomJob(t, &company, jobDetails{})
		} else {
			createRandomJob(t, nil, jobDetails{})
		}
	}

	params := ListJobsByCompanyIDParams{
		CompanyID: company.ID,
		Limit:     5,
		Offset:    0,
	}

	jobs, err := testQueries.ListJobsByCompanyID(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, jobs, 5)
	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.Equal(t, company.ID, job.CompanyID)
	}
}

func TestQueries_ListJobsByLocation(t *testing.T) {
	details := jobDetails{
		location: "testLocation",
	}
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomJob(t, nil, details)
		} else {
			createRandomJob(t, nil, jobDetails{})
		}
	}

	params := ListJobsByLocationParams{
		Location: details.location,
		Limit:    5,
		Offset:   0,
	}

	jobs, err := testQueries.ListJobsByLocation(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, jobs, 5)
	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.Equal(t, details.location, job.Location)
	}
}

func TestQueries_ListJobsByTitle(t *testing.T) {
	details := jobDetails{
		title: "testTitle",
	}
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomJob(t, nil, details)
		} else {
			createRandomJob(t, nil, jobDetails{})
		}
	}

	params := ListJobsByTitleParams{
		Title:  details.title,
		Limit:  5,
		Offset: 0,
	}

	jobs, err := testQueries.ListJobsByTitle(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobs, 5)
	require.NotEmpty(t, jobs)

	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.True(t, strings.Contains(job.Title, params.Title))
	}
}

func TestQueries_ListJobsBySalaryRange(t *testing.T) {
	details := jobDetails{
		salaryMin: 1000,
		salaryMax: 1200,
	}

	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomJob(t, nil, details)
		} else {
			createRandomJob(t, nil, jobDetails{})
		}
	}

	params := ListJobsBySalaryRangeParams{
		SalaryMin: details.salaryMin - 10,
		SalaryMax: details.salaryMax + 10,
		Limit:     5,
		Offset:    0,
	}

	jobs, err := testQueries.ListJobsBySalaryRange(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobs, 5)
	require.NotEmpty(t, jobs)
	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.True(t, job.SalaryMin >= params.SalaryMin)
		require.True(t, job.SalaryMax <= params.SalaryMax)
	}
}

func TestQueries_ListJobsMatchingUserSkills(t *testing.T) {
	skillName := utils.RandomString(10)
	user := createRandomUser(t)
	createRandomUserSkill(t, user.ID, skillName)

	var jobIDs []int32
	for i := 0; i < 5; i++ {
		job := createRandomJob(t, nil, jobDetails{})
		jobIDs = append(jobIDs, job.ID)
		createRandomJobSkill(t, &job, skillName)
	}

	params := ListJobsMatchingUserSkillsParams{
		UserID: user.ID,
		Limit:  5,
		Offset: 0,
	}

	jobs, err := testQueries.ListJobsMatchingUserSkills(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobs, 5)
	for _, job := range jobs {
		require.NotEmpty(t, job)
		require.Contains(t, jobIDs, job.ID)
	}
}
