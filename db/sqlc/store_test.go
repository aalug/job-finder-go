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

	userSkills, err := testStore.CreateMultipleUserSkills(context.Background(), params, user.ID)
	require.NoError(t, err)
	require.Len(t, userSkills, 5)
	for _, userSkill := range userSkills {
		require.NotEmpty(t, userSkill)
		require.Equal(t, user.ID, userSkill.UserID)
		require.NotZero(t, userSkill.ID)
	}
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

func TestSQLStore_GetUserDetailsByEmail(t *testing.T) {
	user := createRandomUser(t)
	params := CreateUserSkillParams{
		UserID:     user.ID,
		Skill:      utils.RandomString(4),
		Experience: utils.RandomInt(1, 5),
	}

	skills, err := testStore.CreateUserSkill(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, skills)
	require.Equal(t, skills.Skill, params.Skill)
	require.Equal(t, skills.Experience, params.Experience)
	require.Equal(t, skills.UserID, params.UserID)

	user, userSkills, err := testStore.GetUserDetailsByEmail(context.Background(), user.Email)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotEmpty(t, userSkills)
	require.Equal(t, user.ID, userSkills[0].UserID)
	require.Equal(t, userSkills[0].Skill, params.Skill)
	require.Equal(t, userSkills[0].Experience, params.Experience)
}
