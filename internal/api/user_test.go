package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aalug/go-gin-job-search/internal/db/mock"
	db2 "github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/pkg/token"
	utils2 "github.com/aalug/go-gin-job-search/pkg/utils"
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
	params   db2.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(arg interface{}) bool {
	params, ok := arg.(db2.CreateUserParams)
	if !ok {
		return false
	}

	err := utils2.CheckPassword(e.password, params.HashedPassword)
	if err != nil {
		return false
	}

	e.params.HashedPassword = params.HashedPassword
	return reflect.DeepEqual(e.params, params)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.params, e.password)
}

func EqCreateUserParams(arg db2.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := generateRandomUser(t)

	var skills []Skill
	var userSkills []db2.UserSkill
	var createUserSkills []db2.CreateMultipleUserSkillsParams
	skills, userSkills, createUserSkills = generateSkills(user.ID)

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
				params := db2.CreateUserParams{
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
					Return(db2.User{}, sql.ErrConnDone)
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
				params := db2.CreateUserParams{
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
					Return(db2.User{}, &pq.Error{Code: "23505"})
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestLoginUserAPI(t *testing.T) {
	user, password := generateRandomUser(t)
	var userSkills []db2.UserSkill
	_, userSkills, _ = generateSkills(user.ID)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				params := db2.ListUserSkillsParams{
					UserID: user.ID,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(userSkills, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Not Found",
			body: gin.H{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.User{}, sql.ErrNoRows)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetUserByEmail",
			body: gin.H{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.User{}, sql.ErrConnDone)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error ListUserSkills",
			body: gin.H{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db2.UserSkill{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Email",
			body: gin.H{
				"email":    "invalid",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Incorrect Password",
			body: gin.H{
				"email":    user.Email,
				"password": fmt.Sprintf("%d, %s", utils2.RandomInt(1, 1000), utils2.RandomString(10)),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Password Too Short",
			body: gin.H{
				"email":    user.Email,
				"password": "abc",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/login"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestGetUserAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	_, userSkills, _ := generateSkills(user.ID)
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserDetailsByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, userSkills, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user, userSkills)
			},
		},
		{
			name: "Internal Server Error",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserDetailsByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.User{}, []db2.UserSkill{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := "/api/v1/users"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateUserAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	skills, userSkills, createUserSkills := generateSkills(user.ID)
	skillIDsToRemove := []int32{userSkills[0].ID}
	newDetails := db2.UpdateUserParams{
		ID:               user.ID,
		FullName:         utils2.RandomString(5),
		Email:            user.Email,
		Location:         utils2.RandomString(5),
		DesiredJobTitle:  utils2.RandomString(5),
		DesiredIndustry:  utils2.RandomString(5),
		DesiredSalaryMin: utils2.RandomInt(1000, 1100),
		DesiredSalaryMax: utils2.RandomInt(1100, 1200),
		Skills:           user.Skills,
		Experience:       user.Experience,
	}

	updatedUser := db2.User{
		ID:               user.ID,
		FullName:         utils2.RandomString(5),
		Email:            user.Email,
		HashedPassword:   user.HashedPassword,
		Location:         utils2.RandomString(5),
		DesiredJobTitle:  utils2.RandomString(5),
		DesiredIndustry:  user.Location,
		DesiredSalaryMin: user.DesiredSalaryMin,
		DesiredSalaryMax: user.DesiredSalaryMax,
		Skills:           user.Skills,
		Experience:       user.Experience,
		CreatedAt:        user.CreatedAt,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"location":            newDetails.Location,
				"desired_job_title":   newDetails.DesiredJobTitle,
				"full_name":           newDetails.FullName,
				"skills_to_add":       skills,
				"skill_ids_to_remove": skillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updatedUser, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Eq(createUserSkills), gomock.Eq(user.ID)).
					Times(1).
					Return(userSkills, nil)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Eq(skillIDsToRemove)).
					Times(1).
					Return(nil)
				listSkillsParams := db2.ListUserSkillsParams{
					UserID: user.ID,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Eq(listSkillsParams)).
					Times(1).
					Return(userSkills, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetUserByEmail",
			body: gin.H{
				"location":            newDetails.Location,
				"desired_job_title":   newDetails.DesiredJobTitle,
				"full_name":           newDetails.FullName,
				"skills_to_add":       skills,
				"skill_ids_to_remove": skillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.User{}, sql.ErrConnDone)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error UpdateUser",
			body: gin.H{
				"location":            newDetails.Location,
				"desired_job_title":   newDetails.DesiredJobTitle,
				"full_name":           newDetails.FullName,
				"skills_to_add":       skills,
				"skill_ids_to_remove": skillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.User{}, sql.ErrConnDone)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateMultipleUserSkills",
			body: gin.H{
				"location":            newDetails.Location,
				"desired_job_title":   newDetails.DesiredJobTitle,
				"full_name":           newDetails.FullName,
				"skills_to_add":       skills,
				"skill_ids_to_remove": skillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updatedUser, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Eq(createUserSkills), gomock.Eq(user.ID)).
					Times(1).
					Return([]db2.UserSkill{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error DeleteMultipleUserSkills",
			body: gin.H{
				"location":            newDetails.Location,
				"full_name":           newDetails.FullName,
				"skills_to_add":       skills,
				"skill_ids_to_remove": skillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updatedUser, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Eq(createUserSkills), gomock.Eq(user.ID)).
					Times(1).
					Return(userSkills, nil)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Eq(skillIDsToRemove)).
					Times(1).
					Return(sql.ErrConnDone)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error ListUserSkills",
			body: gin.H{
				"desired_job_title":   newDetails.DesiredJobTitle,
				"skills_to_add":       skills,
				"skill_ids_to_remove": skillIDsToRemove,
				"desired_salary_min":  1000,
				"desired_salary_max":  2000,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updatedUser, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Eq(createUserSkills), gomock.Eq(user.ID)).
					Times(1).
					Return(userSkills, nil)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Eq(skillIDsToRemove)).
					Times(1).
					Return(nil)
				listSkillsParams := db2.ListUserSkillsParams{
					UserID: user.ID,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Eq(listSkillsParams)).
					Times(1).
					Return([]db2.UserSkill{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Email",
			body: gin.H{
				"email": "invalid",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Body",
			body: gin.H{
				"location": 123,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Salary Min Greater Than Max",
			body: gin.H{
				"desired_salary_min": 10000,
				"desired_salary_max": 100,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateUserPasswordAPI(t *testing.T) {
	user, oldPassword := generateRandomUser(t)
	newPassword := utils2.RandomString(6)
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"old_password": oldPassword,
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdatePassword(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Invalid Old Password",
			body: gin.H{
				"old_password": "invalid",
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdatePassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "New Password Too Short",
			body: gin.H{
				"old_password": oldPassword,
				"new_password": "123",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdatePassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Incorrect Password",
			body: gin.H{
				"old_password": "incorrect",
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdatePassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetUserByEmail",
			body: gin.H{
				"old_password": oldPassword,
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.User{}, sql.ErrConnDone)
				store.EXPECT().
					UpdatePassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error UpdatePassword",
			body: gin.H{
				"old_password": oldPassword,
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					UpdatePassword(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/password"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteUserAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					DeleteAllUserSkills(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetUserByEmail",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.User{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteAllUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error DeleteAllUserSkills",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					DeleteAllUserSkills(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(sql.ErrConnDone)
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error DeleteAllUserSkills",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					DeleteAllUserSkills(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := "/api/v1/users"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

// generateRandomUser generates a random user and returns it with the password
func generateRandomUser(t *testing.T) (db2.User, string) {
	password := utils2.RandomString(6)
	hashedPassword, err := utils2.HashPassword(password)
	require.NoError(t, err)

	user := db2.User{
		ID:               utils2.RandomInt(1, 1000),
		FullName:         utils2.RandomString(6),
		Email:            utils2.RandomEmail(),
		HashedPassword:   hashedPassword,
		Location:         utils2.RandomString(4),
		DesiredJobTitle:  utils2.RandomString(3),
		DesiredIndustry:  utils2.RandomString(2),
		DesiredSalaryMin: utils2.RandomInt(1000, 1100),
		DesiredSalaryMax: utils2.RandomInt(1100, 1200),
		Skills:           utils2.RandomString(5),
		Experience:       utils2.RandomString(5),
		CreatedAt:        time.Now(),
	}

	return user, password
}

// requireBodyMatchUser checks if the body of the response matches the user
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db2.User, skills []db2.UserSkill) {
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

	// in the response skills are sent with ids
	// so we need to remove them from the response
	// to compare with the skills from the request
	var skls []Skill
	for _, s := range gotUser.Skills {
		skls = append(skls, Skill{
			SkillName:         s.SkillName,
			YearsOfExperience: s.YearsOfExperience,
		})
	}
	for _, skill := range userSkills {
		require.Contains(t, skls, skill)
	}
}

func generateSkills(userID int32) ([]Skill, []db2.UserSkill, []db2.CreateMultipleUserSkillsParams) {
	var skills []Skill
	var userSkills []db2.UserSkill
	var createUserSkills []db2.CreateMultipleUserSkillsParams
	for i := 0; i < 2; i++ {
		name := utils2.RandomString(3)
		experience := utils2.RandomInt(1, 5)
		skills = append(skills, Skill{
			SkillName:         name,
			YearsOfExperience: experience,
		})
		createUserSkills = append(createUserSkills, db2.CreateMultipleUserSkillsParams{
			Skill:      name,
			Experience: experience,
		})
		userSkills = append(userSkills, db2.UserSkill{
			ID:         utils2.RandomInt(1, 100),
			UserID:     userID,
			Skill:      name,
			Experience: experience,
		})
	}
	return skills, userSkills, createUserSkills
}
