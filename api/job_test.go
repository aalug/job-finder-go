package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/aalug/go-gin-job-search/db/mock"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateJobAPI(t *testing.T) {
	employer, _, _ := generateRandomEmployerAndCompany(t)

	job := generateRandomJob()

	requiredSkills := []string{"skill1", "skill2"}
	var jobSkills []db.ListJobSkillsByJobIDRow
	for _, skill := range requiredSkills {
		js := db.ListJobSkillsByJobIDRow{
			ID:    utils.RandomInt(1, 1000),
			Skill: skill,
		}
		jobSkills = append(jobSkills, js)
	}

	requestBody := gin.H{
		"title":           job.Title,
		"description":     job.Description,
		"industry":        job.Industry,
		"location":        job.Location,
		"salary_min":      job.SalaryMin,
		"salary_max":      job.SalaryMax,
		"requirements":    job.Requirements,
		"required_skills": requiredSkills,
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
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				params := db.CreateJobParams{
					Title:        job.Title,
					Industry:     job.Industry,
					CompanyID:    employer.CompanyID,
					Description:  job.Description,
					Location:     job.Location,
					SalaryMin:    job.SalaryMin,
					SalaryMax:    job.SalaryMax,
					Requirements: job.Requirements,
				}
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkills), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				listSkillsParams := db.ListJobSkillsByJobIDParams{
					JobID:  job.ID,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Eq(listSkillsParams)).
					Times(1).
					Return(jobSkills, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchJob(t, recorder.Body, job, jobSkills)
			},
		},
		{
			name: "Internal Server Error ListJobSkillsByJobID",
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				params := db.CreateJobParams{
					Title:        job.Title,
					Industry:     job.Industry,
					CompanyID:    employer.CompanyID,
					Description:  job.Description,
					Location:     job.Location,
					SalaryMin:    job.SalaryMin,
					SalaryMax:    job.SalaryMax,
					Requirements: job.Requirements,
				}
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkills), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobSkillsByJobIDRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateMultipleJobSkills",
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				params := db.CreateJobParams{
					Title:        job.Title,
					Industry:     job.Industry,
					CompanyID:    employer.CompanyID,
					Description:  job.Description,
					Location:     job.Location,
					SalaryMin:    job.SalaryMin,
					SalaryMax:    job.SalaryMax,
					Requirements: job.Requirements,
				}
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateJob",
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Job{}, sql.ErrConnDone)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetEmployerByEmail",
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Body",
			body: gin.H{
				"title":    job.Title,
				"industry": job.Industry,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
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

			url := "/jobs"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteJobAPI(t *testing.T) {
	employer, _, _ := generateRandomEmployerAndCompany(t)
	job := generateRandomJob()

	// set the company ID to the employer's company ID
	// so that the job belongs to the employer
	job.CompanyID = employer.CompanyID

	testCases := []struct {
		name          string
		jobID         int32
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:  "Unauthorized User",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, "unauthorized@example.com", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq("unauthorized@example.com")).
					Times(1).
					Return(db.Employer{
						ID:        employer.ID + 1,
						CompanyID: employer.CompanyID + 1,
						Email:     "unauthorized@example.com",
					}, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:  "Invalid Job ID",
			jobID: 0,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error GetEmployerByEmail",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error GetJob",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Job{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error DeleteJob",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/jobs/%d", tc.jobID)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestGetJobAPI(t *testing.T) {
	employer, _, company := generateRandomEmployerAndCompany(t)
	job := generateRandomJob()

	// set the company ID to the employer's company ID
	// so that the job belongs to the employer
	job.CompanyID = employer.CompanyID

	getJobRow := db.GetJobDetailsRow{
		ID:               job.ID,
		Title:            job.Title,
		Industry:         job.Industry,
		CompanyID:        job.CompanyID,
		Description:      job.Description,
		Location:         job.Location,
		SalaryMin:        job.SalaryMin,
		SalaryMax:        job.SalaryMax,
		Requirements:     job.Requirements,
		CreatedAt:        job.CreatedAt,
		CompanyName:      company.Name,
		CompanyLocation:  company.Location,
		CompanyIndustry:  company.Industry,
		EmployerID:       employer.ID,
		EmployerEmail:    employer.Email,
		EmployerFullName: employer.FullName,
	}

	testCases := []struct {
		name          string
		jobID         int32
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			jobID: job.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetJobDetails(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(getJobRow, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobDetails(t, recorder.Body, getJobRow)
			},
		},
		{
			name:  "Not Found",
			jobID: job.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetJobDetails(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetJobDetailsRow{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error",
			jobID: job.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetJobDetails(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetJobDetailsRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Invalid Job ID",
			jobID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetJobDetails(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/jobs/%d", tc.jobID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func generateRandomJob() db.Job {
	return db.Job{
		ID:           utils.RandomInt(1, 1000),
		Title:        utils.RandomString(4),
		Industry:     utils.RandomString(2),
		Description:  utils.RandomString(5),
		Location:     utils.RandomString(4),
		SalaryMin:    utils.RandomInt(100, 1000),
		SalaryMax:    utils.RandomInt(100, 1000),
		Requirements: utils.RandomString(5),
	}
}

func requireBodyMatchJob(t *testing.T, body *bytes.Buffer, job db.Job, skills []db.ListJobSkillsByJobIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotJob jobResponse
	err = json.Unmarshal(data, &gotJob)
	require.NoError(t, err)

	require.Equal(t, job.Title, gotJob.Title)
	require.Equal(t, job.Industry, gotJob.Industry)
	require.Equal(t, job.Description, gotJob.Description)
	require.Equal(t, job.Location, gotJob.Location)
	require.Equal(t, job.SalaryMin, gotJob.SalaryMin)
	require.Equal(t, job.SalaryMax, gotJob.SalaryMax)
	require.Equal(t, job.Requirements, gotJob.Requirements)
	require.Equal(t, skills, gotJob.RequiredSkills)
}

func requireBodyMatchJobDetails(t *testing.T, body *bytes.Buffer, row db.GetJobDetailsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotJobRow db.GetJobDetailsRow
	err = json.Unmarshal(data, &gotJobRow)
	require.NoError(t, err)
	require.Equal(t, row, gotJobRow)
}
