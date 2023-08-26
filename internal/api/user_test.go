package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aalug/go-gin-job-search/internal/db/mock"
	db "github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/worker"
	mockworker "github.com/aalug/go-gin-job-search/internal/worker/mock"
	"github.com/aalug/go-gin-job-search/pkg/token"
	utils "github.com/aalug/go-gin-job-search/pkg/utils"
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

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (e eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	err := utils.CheckPassword(e.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(e.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	err = actualArg.AfterCreate(e.user)
	return err == nil
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password, user}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := generateRandomUser(t)

	var skills []Skill
	var userSkills []db.UserSkill
	var createUserSkills []db.CreateMultipleUserSkillsParams
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
		buildStubs    func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				params := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
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
					},
				}
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(params, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Eq(createUserSkills), gomock.Eq(user.ID)).
					Times(1).
					Return(userSkills, nil)
				taskPayload := &worker.PayloadSendVerificationEmail{
					Email: user.Email,
				}
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user, userSkills)
			},
		},
		{
			name: "Invalid Body",
			body: gin.H{
				"full_name":          user.FullName,
				"skills_description": user.Skills,
				"desired_job_title":  user.DesiredJobTitle,
				"location":           user.Location,
				"skills":             skills,
			},
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "SalaryMin gt SalaryMax",
			body: gin.H{
				"email":              user.Email,
				"password":           password,
				"full_name":          user.FullName,
				"skills_description": user.Skills,
				"experience":         user.Experience,
				"desired_industry":   user.DesiredIndustry,
				"desired_salary_min": 10000,
				"desired_salary_max": 100,
				"desired_job_title":  user.DesiredJobTitle,
				"location":           user.Location,
				"skills":             skills,
			},
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unique Violation",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, &pq.Error{Code: "23505"})
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateUserTx",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateMultipleUserSkills",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)
				store.EXPECT().
					CreateMultipleUserSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.UserSkill{}, sql.ErrConnDone)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
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

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockworker.NewMockTaskDistributor(taskCtrl)

			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, nil, taskDistributor)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := BaseUrl + "/users"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestLoginUserAPI(t *testing.T) {
	user, password := generateRandomUser(t)
	user.IsEmailVerified = true
	var userSkills []db.UserSkill
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
				params := db.ListUserSkillsParams{
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
					Return(db.User{}, sql.ErrNoRows)
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
					Return(db.User{}, sql.ErrConnDone)
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
					Return([]db.UserSkill{}, sql.ErrConnDone)
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
				"password": fmt.Sprintf("%d, %s", utils.RandomInt(1, 1000), utils.RandomString(10)),
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
		{
			name: "Email Not Verified",
			body: gin.H{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				user.IsEmailVerified = false
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
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

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := BaseUrl + "/users/login"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestGetUserAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, _ := generateRandomEmployerAndCompany(t)
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
			name: "Unauthorized Only Users Access",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserDetailsByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, []db.UserSkill{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
					Return(db.User{}, []db.UserSkill{}, sql.ErrConnDone)
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

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			url := BaseUrl + "/users"
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
	employer, _, _ := generateRandomEmployerAndCompany(t)
	skills, userSkills, createUserSkills := generateSkills(user.ID)
	skillIDsToRemove := []int32{userSkills[0].ID}
	newDetails := db.UpdateUserParams{
		ID:               user.ID,
		FullName:         utils.RandomString(5),
		Email:            user.Email,
		Location:         utils.RandomString(5),
		DesiredJobTitle:  utils.RandomString(5),
		DesiredIndustry:  utils.RandomString(5),
		DesiredSalaryMin: utils.RandomInt(1000, 1100),
		DesiredSalaryMax: utils.RandomInt(1100, 1200),
		Skills:           user.Skills,
		Experience:       user.Experience,
	}

	updatedUser := db.User{
		ID:               user.ID,
		FullName:         utils.RandomString(5),
		Email:            user.Email,
		HashedPassword:   user.HashedPassword,
		Location:         utils.RandomString(5),
		DesiredJobTitle:  utils.RandomString(5),
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
				listSkillsParams := db.ListUserSkillsParams{
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
			name: "Unauthorized Only User Access",
			body: gin.H{
				"location":            newDetails.Location,
				"desired_job_title":   newDetails.DesiredJobTitle,
				"full_name":           newDetails.FullName,
				"skills_to_add":       skills,
				"skill_ids_to_remove": skillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
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
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
					Return(db.User{}, sql.ErrConnDone)
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
					Return(db.User{}, sql.ErrConnDone)
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
					Return([]db.UserSkill{}, sql.ErrConnDone)
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
				listSkillsParams := db.ListUserSkillsParams{
					UserID: user.ID,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListUserSkills(gomock.Any(), gomock.Eq(listSkillsParams)).
					Times(1).
					Return([]db.UserSkill{}, sql.ErrConnDone)
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

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := BaseUrl + "/users"
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
	employer, _, _ := generateRandomEmployerAndCompany(t)
	newPassword := utils.RandomString(6)
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
			name: "Unauthorized Only User Access",
			body: gin.H{
				"old_password": oldPassword,
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
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
					Return(db.User{}, sql.ErrConnDone)
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

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := BaseUrl + "/users/password"
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
	employer, _, _ := generateRandomEmployerAndCompany(t)
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
			name: "Unauthorized Only User Access",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					DeleteAllUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
					Return(db.User{}, sql.ErrConnDone)
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

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			url := BaseUrl + "/users"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestVerifyUserEmailAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	verifyEmail := db.VerifyEmail{
		ID:         int64(utils.RandomInt(1, 1000)),
		Email:      user.Email,
		SecretCode: utils.RandomString(32),
		IsUsed:     false,
		CreatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(15 * time.Minute),
	}

	type Query struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				ID:   verifyEmail.ID,
				Code: verifyEmail.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				params := db.VerifyEmailTxParams{
					ID:         verifyEmail.ID,
					SecretCode: verifyEmail.SecretCode,
				}
				verifyEmail.IsUsed = true
				user.IsEmailVerified = true
				store.EXPECT().
					VerifyUserEmailTx(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(db.VerifyUserEmailResult{
						User:        user,
						VerifyEmail: verifyEmail,
					}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Internal Server Error",
			query: Query{
				ID:   verifyEmail.ID,
				Code: verifyEmail.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					VerifyUserEmailTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.VerifyUserEmailResult{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Verify Email Not Found",
			query: Query{
				ID:   verifyEmail.ID,
				Code: verifyEmail.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					VerifyUserEmailTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.VerifyUserEmailResult{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Invalid Code Length",
			query: Query{
				ID:   verifyEmail.ID,
				Code: utils.RandomString(31),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					VerifyUserEmailTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid ID",
			query: Query{
				ID:   0,
				Code: verifyEmail.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					VerifyUserEmailTx(gomock.Any(), gomock.Any()).
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

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			url := BaseUrl + "/users/verify-email"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := req.URL.Query()
			q.Add("id", fmt.Sprintf("%d", tc.query.ID))
			q.Add("code", tc.query.Code)
			req.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestSendVerificationEmailToUserAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	testCases := []struct {
		name          string
		email         string
		buildStubs    func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			email: user.Email,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					DeleteVerifyEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(nil)
				taskPayload := &worker.PayloadSendVerificationEmail{
					Email: user.Email,
				}
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Eq(taskPayload), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "Invalid Email",
			email: "invalid",
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteVerifyEmail(gomock.Any(), gomock.Any()).
					Times(0)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "User Not Found",
			email: user.Email,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					DeleteVerifyEmail(gomock.Any(), gomock.Any()).
					Times(0)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error GetUserByEmail",
			email: user.Email,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteVerifyEmail(gomock.Any(), gomock.Any()).
					Times(0)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "DeleteVerifyEmail ErrNoRows Do Nothing",
			email: user.Email,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					DeleteVerifyEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(sql.ErrNoRows)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error DeleteVerifyEmail",
			email: user.Email,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					DeleteVerifyEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(sql.ErrConnDone)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error DistributeTaskSendVerificationEmail",
			email: user.Email,
			buildStubs: func(store *mockdb.MockStore, distributor *mockworker.MockTaskDistributor) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					DeleteVerifyEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(nil)
				distributor.EXPECT().
					DistributeTaskSendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("some error"))
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

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockworker.NewMockTaskDistributor(taskCtrl)

			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, nil, taskDistributor)
			recorder := httptest.NewRecorder()

			url := BaseUrl + "/users/send-verification-email"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := req.URL.Query()
			q.Add("email", tc.email)
			req.URL.RawQuery = q.Encode()

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
		ID:               utils.RandomInt(1, 1000),
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

func generateSkills(userID int32) ([]Skill, []db.UserSkill, []db.CreateMultipleUserSkillsParams) {
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
			UserID:     userID,
			Skill:      name,
			Experience: experience,
		})
	}
	return skills, userSkills, createUserSkills
}
