package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/aalug/go-gin-job-search/db/mock"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type eqCreateUserParamsMatcher struct {
	params   db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(arg interface{}) bool {
	params, ok := arg.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := utils.CheckPassword(e.password, params.HashedPassword)
	if err != nil {
		return false
	}

	e.params.HashedPassword = params.HashedPassword
	return reflect.DeepEqual(e.params, params)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.params, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := generateRandomUser(t)

	var skills []Skill
	var userSkills []db.UserSkill
	var createUserSkills []db.CreateMultipleUserSkillsParams
	for i := 0; i < 2; i++ {
		name := utils.RandomString(3)
		experience := utils.RandomInt(1, 5)
		skills = append(skills, Skill{
			SkillName:         name,
			YearsOfExperience: experience,
		})
		createUserSkills = append(createUserSkills, db.CreateMultipleUserSkillsParams{
			Skill:      name,
			Experience: experience,
		})
		userSkills = append(userSkills, db.UserSkill{
			ID:         utils.RandomInt(1, 100),
			UserID:     user.ID,
			Skill:      name,
			Experience: experience,
		})
	}

	requestBody := gin.H{
		"email":              user.Email,
		"password":           password,
		"full_name":          user.FullName,
		"skills_description": user.Skills,
		"experience":         user.Experience,
		"desired_industry":   user.DesiredIndustry,
		"desired_salary_min": user.DesiredSalaryMin,
		"desired_salary_max": user.DesiredSalaryMax,
		"desired_job_title":  user.DesiredJobTitle,
		"location":           user.Location,
		"skills":             skills,
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.CreateUserParams{
					FullName:         user.FullName,
					Email:            user.Email,
					HashedPassword:   user.HashedPassword,
					Location:         user.Location,
					DesiredJobTitle:  user.DesiredJobTitle,
					DesiredIndustry:  user.DesiredIndustry,
					DesiredSalaryMin: user.DesiredSalaryMin,
					DesiredSalaryMax: user.DesiredSalaryMax,
					Skills:           user.Skills,
					Experience:       user.Experience,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(params, password)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Eq(createUserSkills), gomock.Eq(user.ID)).
					Times(1).
					Return(userSkills, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user, userSkills)
			},
		},
		{
			name: "Internal Server Error CreateUser",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateMultipleUserSkills",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				params := db.CreateUserParams{
					FullName:         user.FullName,
					Email:            user.Email,
					HashedPassword:   user.HashedPassword,
					Location:         user.Location,
					DesiredJobTitle:  user.DesiredJobTitle,
					DesiredIndustry:  user.DesiredIndustry,
					DesiredSalaryMin: user.DesiredSalaryMin,
					DesiredSalaryMax: user.DesiredSalaryMax,
					Skills:           user.Skills,
					Experience:       user.Experience,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(params, password)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Eq(createUserSkills), gomock.Eq(user.ID)).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Duplicated Email",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "Invalid Body",
			body: gin.H{
				"password": password,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Email",
			body: gin.H{
				"email":              "invalid",
				"password":           password,
				"full_name":          user.FullName,
				"skills_description": user.Skills,
				"experience":         user.Experience,
				"desired_industry":   user.DesiredIndustry,
				"desired_salary_min": user.DesiredSalaryMin,
				"desired_salary_max": user.DesiredSalaryMax,
				"desired_job_title":  user.DesiredJobTitle,
				"location":           user.Location,
				"skills":             skills,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Password Too Short",
			body: gin.H{
				"email":              user.Email,
				"password":           "123",
				"full_name":          user.FullName,
				"skills_description": user.Skills,
				"experience":         user.Experience,
				"desired_industry":   user.DesiredIndustry,
				"desired_salary_min": user.DesiredSalaryMin,
				"desired_salary_max": user.DesiredSalaryMax,
				"desired_job_title":  user.DesiredJobTitle,
				"location":           user.Location,
				"skills":             skills,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Password Too Short",
			body: gin.H{
				"email":              user.Email,
				"password":           password,
				"full_name":          user.FullName,
				"skills_description": user.Skills,
				"experience":         user.Experience,
				"desired_industry":   user.DesiredIndustry,
				"desired_salary_min": 1000,
				"desired_salary_max": 10,
				"desired_job_title":  user.DesiredJobTitle,
				"location":           user.Location,
				"skills":             skills,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

// generateRandomUser generates a random user and returns it with the password
func generateRandomUser(t *testing.T) (db.User, string) {
	password := utils.RandomString(6)
	hashedPassword, err := utils.HashPassword(password)
	require.NoError(t, err)

	user := db.User{
		FullName:         utils.RandomString(6),
		Email:            utils.RandomEmail(),
		HashedPassword:   hashedPassword,
		Location:         utils.RandomString(4),
		DesiredJobTitle:  utils.RandomString(3),
		DesiredIndustry:  utils.RandomString(2),
		DesiredSalaryMin: utils.RandomInt(1000, 1100),
		DesiredSalaryMax: utils.RandomInt(1100, 1200),
		Skills:           utils.RandomString(5),
		Experience:       utils.RandomString(5),
		CreatedAt:        time.Now(),
	}

	return user, password
}

// requireBodyMatchUser checks if the body of the response matches the user
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User, skills []db.UserSkill) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser userResponse
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Location, gotUser.Location)
	require.Equal(t, user.DesiredJobTitle, gotUser.DesiredJobTitle)
	require.Equal(t, user.DesiredIndustry, gotUser.DesiredIndustry)
	require.Equal(t, user.DesiredSalaryMin, gotUser.DesiredSalaryMin)
	require.Equal(t, user.DesiredSalaryMax, gotUser.DesiredSalaryMax)
	require.Equal(t, user.Skills, gotUser.SkillsDescription)
	require.Equal(t, user.Experience, gotUser.Experience)
	require.WithinDuration(t, user.CreatedAt, gotUser.CreatedAt, time.Second)

	var userSkills []Skill
	for _, skill := range skills {
		userSkills = append(userSkills, Skill{
			SkillName:         skill.Skill,
			YearsOfExperience: skill.Experience,
		})
	}

	for _, skill := range userSkills {
		require.Contains(t, gotUser.Skills, skill)
	}
}
