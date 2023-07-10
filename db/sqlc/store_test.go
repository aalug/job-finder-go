package db

import (
	"context"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLStore_CreateMultipleUserSkills(t *testing.T) {
	user := createRandomUser(t)
	var params []CreateMultipleUserSkillsParams
	for i := 0; i < 5; i++ {
		params = append(params, CreateMultipleUserSkillsParams{
			Skill:      utils.RandomString(9),
			Experience: utils.RandomInt(0, 10),
		})
	}

	err := testStore.CreateMultipleUserSkills(context.Background(), params, user.ID)
	require.NoError(t, err)
}

func TestSQLStore_CreateMultipleJobSkills(t *testing.T) {
	job := createRandomJob(t, nil, jobDetails{})
	var skills []string
	for i := 0; i < 5; i++ {
		skills = append(skills, utils.RandomString(9))
	}

	err := testStore.CreateMultipleJobSkills(context.Background(), skills, job.ID)
	require.NoError(t, err)
}

func TestSQLStore_DeleteJobPosting(t *testing.T) {
	job := createRandomJob(t, nil, jobDetails{})
	err := testStore.DeleteJobPosting(context.Background(), job.ID)
	require.NoError(t, err)

	params := ListJobSkillsByJobIDParams{
		JobID:  job.ID,
		Limit:  5,
		Offset: 0,
	}
	jobSkills, err := testStore.ListJobSkillsByJobID(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, jobSkills, 0)
	require.Empty(t, jobSkills)
}
