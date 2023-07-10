package db

import (
	"context"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

// createRandomUserSkill creates and return a random user skill
func createRandomUserSkill(t *testing.T, userID int32, skill string) UserSkill {
	params := CreateUserSkillParams{
		Experience: utils.RandomInt(1, 10),
	}

	if userID == 0 {
		params.UserID = createRandomUser(t).ID
	} else {
		params.UserID = userID
	}

	if skill == "" {
		params.Skill = utils.RandomString(10)
	} else {
		params.Skill = skill
	}

	userSkill, err := testQueries.CreateUserSkill(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, userSkill)
	require.Equal(t, params.UserID, userSkill.UserID)
	require.Equal(t, params.Skill, userSkill.Skill)
	require.Equal(t, params.Experience, userSkill.Experience)
	require.NotZero(t, userSkill.ID)

	return userSkill
}

func TestQueries_CreateJobSkill(t *testing.T) {
	createRandomUserSkill(t, 0, "")
}

func TestQueries_DeleteUserSkill(t *testing.T) {
	userSkill := createRandomUserSkill(t, 0, "")
	err := testQueries.DeleteUserSkill(context.Background(), userSkill.ID)
	require.NoError(t, err)
}

func TestQueries_ListUserSkills(t *testing.T) {
	user := createRandomUser(t)
	for i := 0; i < 5; i++ {
		createRandomUserSkill(t, user.ID, "")
	}
	params := ListUserSkillsParams{
		UserID: user.ID,
		Limit:  5,
		Offset: 0,
	}

	userSkills, err := testQueries.ListUserSkills(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, userSkills, 5)
	for _, userSkill := range userSkills {
		require.NotEmpty(t, userSkill)
		require.Equal(t, user.ID, userSkill.UserID)
		require.NotZero(t, userSkill.ID)
	}
}

func TestQueries_ListUsersBySkill(t *testing.T) {
	skillName := utils.RandomString(10)
	var userIDs []int32
	for i := 0; i < 5; i++ {
		user := createRandomUser(t)
		userIDs = append(userIDs, user.ID)
		createRandomUserSkill(t, user.ID, skillName)
	}

	params := ListUsersBySkillParams{
		Skill:  skillName,
		Limit:  5,
		Offset: 0,
	}

	users, err := testQueries.ListUsersBySkill(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, users, 5)
	for _, user := range users {
		require.NotEmpty(t, user)
		require.Contains(t, userIDs, user.ID)
	}
}

func TestQueries_UpdateUserSkill(t *testing.T) {
	userSkill := createRandomUserSkill(t, 0, "")
	params := UpdateUserSkillParams{
		ID:         userSkill.ID,
		Experience: utils.RandomInt(5, 10),
	}

	updatedUserSkill, err := testQueries.UpdateUserSkill(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, userSkill.ID, updatedUserSkill.ID)
	require.Equal(t, params.Experience, updatedUserSkill.Experience)
	require.Equal(t, userSkill.UserID, updatedUserSkill.UserID)
	require.Equal(t, userSkill.Skill, updatedUserSkill.Skill)
}
