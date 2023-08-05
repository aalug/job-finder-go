package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aalug/go-gin-job-search/internal/db/mock"
	db "github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/esearch"
	mockesearch "github.com/aalug/go-gin-job-search/internal/esearch/mock"
	"github.com/aalug/go-gin-job-search/pkg/token"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestCreateJobAPI(t *testing.T) {
	employer, _, company := generateRandomEmployerAndCompany(t)

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
		buildStubs    func(store *mockdb.MockStore, client *mockesearch.MockESearchClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Eq(employer.CompanyID)).
					Times(1).
					Return(company.Name, nil)
				j := esearch.Job{
					ID:           job.ID,
					Title:        job.Title,
					Industry:     job.Industry,
					CompanyName:  company.Name,
					Description:  job.Description,
					Location:     job.Location,
					SalaryMin:    job.SalaryMin,
					SalaryMax:    job.SalaryMax,
					Requirements: job.Requirements,
					JobSkills:    requiredSkills,
				}
				client.EXPECT().
					IndexJobAsDocument(gomock.Eq(1), gomock.Eq(j)).
					Times(1).
					Return(nil)
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
					Times(0)
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetCompanyNameByID",
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(jobSkills, nil)
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(1).
					Return("", sql.ErrConnDone)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error IndexJobAsDocument",
			body: requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(jobSkills, nil)
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(company.Name, nil)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("some error"))
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Salary Min Greater Than Max",
			body: gin.H{
				"title":           job.Title,
				"description":     job.Description,
				"industry":        job.Industry,
				"location":        job.Location,
				"salary_min":      1000,
				"salary_max":      10,
				"requirements":    job.Requirements,
				"required_skills": requiredSkills,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
				store.EXPECT().
					GetCompanyNameByID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					IndexJobAsDocument(gomock.Any(), gomock.Any()).
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
			client := mockesearch.NewMockESearchClient(ctrl)
			tc.buildStubs(store, client)

			server := newTestServer(t, store, client)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/jobs"
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
		buildStubs    func(store *mockdb.MockStore, client *mockesearch.MockESearchClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				documentID := strconv.Itoa(int(job.ID))
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Eq(int(job.ID))).
					Times(1).
					Return(documentID, nil)
				client.EXPECT().
					DeleteJobDocument(gomock.Eq(documentID)).
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
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
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Job{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error GetDocumentIDByJobID",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Eq(int(job.ID))).
					Times(1).
					Return("", errors.New("some error"))
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error DeleteJobDocument",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(nil)
				documentID := strconv.Itoa(int(job.ID))
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Eq(int(job.ID))).
					Times(1).
					Return(documentID, nil)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
					Times(1).
					Return(errors.New("some error"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Not Found",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(db.Job{}, sql.ErrNoRows)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error DeleteJobSkillsByJobID",
			jobID: job.ID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					DeleteJobSkillsByJobID(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(sql.ErrConnDone)
				store.EXPECT().
					DeleteJob(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					DeleteJobDocument(gomock.Any()).
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
			client := mockesearch.NewMockESearchClient(ctrl)
			tc.buildStubs(store, client)

			server := newTestServer(t, store, client)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/jobs/%d", tc.jobID)
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/jobs/%d", tc.jobID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestFilterAndListJobsAPI(t *testing.T) {
	_, _, company := generateRandomEmployerAndCompany(t)
	var jobs []db.ListJobsByFiltersRow
	title := utils.RandomString(5)
	industry := utils.RandomString(4)
	jobLocation := utils.RandomString(6)
	salaryMin := utils.RandomInt(100, 150)
	salaryMax := utils.RandomInt(151, 200)
	title2 := utils.RandomString(8)
	industry2 := utils.RandomString(5)
	jobLocation2 := utils.RandomString(7)
	salaryMin2 := utils.RandomInt(201, 250)
	salaryMax2 := utils.RandomInt(251, 300)

	job := generateJob(
		title,
		industry,
		jobLocation,
		salaryMin,
		salaryMax,
	)
	job2 := generateJob(
		title2,
		industry2,
		jobLocation2,
		salaryMin2,
		salaryMax2,
	)

	for i := 0; i < 10; i++ {
		j := job
		if i%2 == 0 {
			j = job2
		}
		row := db.ListJobsByFiltersRow{
			ID:           j.ID,
			Title:        j.Title,
			Industry:     j.Industry,
			CompanyID:    j.CompanyID,
			Description:  j.Description,
			Location:     j.Location,
			SalaryMin:    j.SalaryMin,
			SalaryMax:    j.SalaryMax,
			Requirements: j.Requirements,
			CreatedAt:    j.CreatedAt,
			CompanyName:  company.Name,
		}
		jobs = append(jobs, row)
	}

	type Query struct {
		page        int32
		pageSize    int32
		industry    string
		jobLocation string
		title       string
		salaryMin   int32
		salaryMax   int32
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
				page:        1,
				pageSize:    10,
				industry:    industry2,
				jobLocation: jobLocation2,
			},
			buildStubs: func(store *mockdb.MockStore) {
				params := db.ListJobsByFiltersParams{
					Limit:  10,
					Offset: 0,
					Title: sql.NullString{
						String: "",
						Valid:  false,
					},
					JobLocation: sql.NullString{
						String: jobLocation2,
						Valid:  true,
					},
					Industry: sql.NullString{
						String: industry2,
						Valid:  true,
					},
					SalaryMin: sql.NullInt32{
						Int32: 0,
						Valid: false,
					},
					SalaryMax: sql.NullInt32{
						Int32: 0,
						Valid: false,
					},
				}
				store.EXPECT().
					ListJobsByFilters(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobs, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobs(t, recorder.Body, jobs)
			},
		},
		{
			name: "No Page In Query",
			query: Query{
				pageSize:    10,
				industry:    industry,
				jobLocation: jobLocation,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByFilters(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No Page Size In Query",
			query: Query{
				page:        1,
				industry:    industry,
				jobLocation: jobLocation,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByFilters(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page",
			query: Query{
				page:        0,
				pageSize:    10,
				industry:    industry,
				jobLocation: jobLocation,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByFilters(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page Size",
			query: Query{
				page:        1,
				pageSize:    50,
				industry:    industry,
				jobLocation: jobLocation,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByFilters(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal Server Error",
			query: Query{
				page:     1,
				pageSize: 10,
				title:    title,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByFilters(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobsByFiltersRow{}, sql.ErrConnDone)
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

			url := "/api/v1/jobs"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query params
			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			q.Add("industry", tc.query.industry)
			q.Add("job_location", tc.query.jobLocation)
			q.Add("title", tc.query.title)
			q.Add("salary_min", fmt.Sprintf("%d", tc.query.salaryMin))
			q.Add("salary_max", fmt.Sprintf("%d", tc.query.salaryMax))
			req.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestListJobsByMatchingSkillsAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, _ := generateRandomEmployerAndCompany(t)
	_, _, company := generateRandomEmployerAndCompany(t)
	var jobs []db.ListJobsMatchingUserSkillsRow
	title := utils.RandomString(5)
	industry := utils.RandomString(4)
	jobLocation := utils.RandomString(6)
	salaryMin := utils.RandomInt(100, 150)
	salaryMax := utils.RandomInt(151, 200)

	job := generateJob(
		title,
		industry,
		jobLocation,
		salaryMin,
		salaryMax,
	)

	for i := 0; i < 10; i++ {
		row := db.ListJobsMatchingUserSkillsRow{
			ID:           job.ID,
			Title:        job.Title,
			Industry:     job.Industry,
			CompanyID:    job.CompanyID,
			Description:  job.Description,
			Location:     job.Location,
			SalaryMin:    job.SalaryMin,
			SalaryMax:    job.SalaryMax,
			Requirements: job.Requirements,
			CreatedAt:    job.CreatedAt,
			CompanyName:  company.Name,
		}
		jobs = append(jobs, row)
	}

	type Query struct {
		page     int32
		pageSize int32
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				page:     1,
				pageSize: 10,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				params := db.ListJobsMatchingUserSkillsParams{
					UserID: user.ID,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobs, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobs(t, recorder.Body, jobs)
			},
		},
		{
			name: "Employer Making Request",
			query: Query{
				page:     1,
				pageSize: 10,
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
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetUserByEmail",
			query: Query{
				page:     1,
				pageSize: 10,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error ListJobsMatchingUserSkills",
			query: Query{
				page:     1,
				pageSize: 10,
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
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobsMatchingUserSkillsRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Page Size",
			query: Query{
				page:     1,
				pageSize: 50,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page",
			query: Query{
				page:     0,
				pageSize: 10,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No Page Size",
			query: Query{
				page: 1,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No Page",
			query: Query{
				pageSize: 10,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsMatchingUserSkills(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/jobs/match-skills"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query params
			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			req.URL.RawQuery = q.Encode()

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateJobAPI(t *testing.T) {
	employer, _, _ := generateRandomEmployerAndCompany(t)
	employer2, _, _ := generateRandomEmployerAndCompany(t)

	job := generateRandomJob()
	newJob := generateRandomJob()

	// set the company id to the employer company id
	// so that the job is created under the same company
	// and as a result the job is treated as owned by  the employer
	job.CompanyID = employer.CompanyID
	newJob.CompanyID = employer.CompanyID

	requiredSkillsToAdd := []string{"skill1", "skill2"}
	requiredSkillIDsToRemove := []int32{1, 2}

	requestBody := gin.H{
		"title":                        newJob.Title,
		"description":                  newJob.Description,
		"industry":                     newJob.Industry,
		"location":                     newJob.Location,
		"salary_min":                   newJob.SalaryMin,
		"salary_max":                   newJob.SalaryMax,
		"requirements":                 newJob.Requirements,
		"required_skills_to_add":       requiredSkillsToAdd,
		"required_skill_ids_to_remove": requiredSkillIDsToRemove,
	}
	listedSkills := []db.ListJobSkillsByJobIDRow{
		{
			ID:    utils.RandomInt(1, 100),
			Skill: utils.RandomString(5),
		},
	}

	testCases := []struct {
		name          string
		jobID         int32
		body          gin.H
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore, client *mockesearch.MockESearchClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(newJob, nil)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkillIDsToRemove)).
					Times(1).
					Return(nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkillsToAdd), gomock.Eq(newJob.ID)).
					Times(1).
					Return(nil)
				listSkillsParams := db.ListJobSkillsByJobIDParams{
					JobID:  newJob.ID,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Eq(listSkillsParams)).
					Times(1).
					Return(listedSkills, nil)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(1).
					Return(strconv.Itoa(int(job.ID)), nil)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJob(t, recorder.Body, newJob, listedSkills)
			},
		},
		{
			name:  "Not Found",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(db.Job{}, sql.ErrNoRows)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error GetJob",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(db.Job{}, sql.ErrConnDone)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error GetEmployerByEmail",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error UpdateJob",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Job{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error DeleteMultipleJobSkills",
			jobID: job.ID,
			body: gin.H{
				"requirements ":                "new requirements",
				"required_skills_to_add":       requiredSkillsToAdd,
				"required_skill_ids_to_remove": requiredSkillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(newJob, nil)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error CreateMultipleJobSkills",
			jobID: job.ID,
			body: gin.H{
				"title":                        "new title",
				"required_skills_to_add":       requiredSkillsToAdd,
				"required_skill_ids_to_remove": requiredSkillIDsToRemove,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(newJob, nil)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error ListJobSkillsByJobID",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(newJob, nil)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobSkillsByJobIDRow{}, sql.ErrConnDone)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Invalid Job ID",
			jobID: 0,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "Invalid Body",
			jobID: job.ID,
			body: gin.H{
				"salary_min": "invalid",
				"title":      100,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "Employer Not Job Owner",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer2.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer2.Email)).
					Times(1).
					Return(employer2, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:  "Salary Min Greater Than Max",
			jobID: job.ID,
			body: gin.H{
				"salary_min": 1000,
				"salary_max": 5,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "Unauthorized",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(0)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(0)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error GetDocumentIDByJobID",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(newJob, nil)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkillIDsToRemove)).
					Times(1).
					Return(nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkillsToAdd), gomock.Eq(newJob.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobSkillsByJobIDRow{}, nil)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(1).
					Return("", errors.New("some error"))
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Internal Server Error UpdateJobDocument",
			jobID: job.ID,
			body:  requestBody,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, client *mockesearch.MockESearchClient) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(job, nil)
				store.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(newJob, nil)
				store.EXPECT().
					DeleteMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkillIDsToRemove)).
					Times(1).
					Return(nil)
				store.EXPECT().
					CreateMultipleJobSkills(gomock.Any(), gomock.Eq(requiredSkillsToAdd), gomock.Eq(newJob.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					ListJobSkillsByJobID(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobSkillsByJobIDRow{}, nil)
				client.EXPECT().
					GetDocumentIDByJobID(gomock.Any()).
					Times(1).
					Return(strconv.Itoa(int(job.ID)), nil)
				client.EXPECT().
					UpdateJobDocument(gomock.Any(), gomock.Any()).
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
			client := mockesearch.NewMockESearchClient(ctrl)
			tc.buildStubs(store, client)

			server := newTestServer(t, store, client)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/jobs/%d", tc.jobID)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestListJobsByCompanyAPI(t *testing.T) {
	_, _, company := generateRandomEmployerAndCompany(t)
	var jobsByExactName []db.ListJobsByCompanyExactNameRow
	var jobByName []db.ListJobsByCompanyNameRow
	var jobByID []db.ListJobsByCompanyIDRow
	title := utils.RandomString(5)
	industry := utils.RandomString(4)
	jobLocation := utils.RandomString(6)
	salaryMin := utils.RandomInt(100, 150)
	salaryMax := utils.RandomInt(151, 200)

	job := generateJob(
		title,
		industry,
		jobLocation,
		salaryMin,
		salaryMax,
	)

	for i := 0; i < 5; i++ {
		row := db.ListJobsByCompanyExactNameRow{
			ID:           job.ID,
			Title:        job.Title,
			Industry:     job.Industry,
			CompanyID:    job.CompanyID,
			Description:  job.Description,
			Location:     job.Location,
			SalaryMin:    job.SalaryMin,
			SalaryMax:    job.SalaryMax,
			Requirements: job.Requirements,
			CreatedAt:    job.CreatedAt,
			CompanyName:  company.Name,
		}
		jobsByExactName = append(jobsByExactName, row)

		row2 := db.ListJobsByCompanyNameRow{
			ID:           job.ID,
			Title:        job.Title,
			Industry:     job.Industry,
			CompanyID:    job.CompanyID,
			Description:  job.Description,
			Location:     job.Location,
			SalaryMin:    job.SalaryMin,
			SalaryMax:    job.SalaryMax,
			Requirements: job.Requirements,
			CreatedAt:    job.CreatedAt,
			CompanyName:  company.Name,
		}
		jobByName = append(jobByName, row2)

		row3 := db.ListJobsByCompanyIDRow{
			ID:           job.ID,
			Title:        job.Title,
			Industry:     job.Industry,
			CompanyID:    job.CompanyID,
			Description:  job.Description,
			Location:     job.Location,
			SalaryMin:    job.SalaryMin,
			SalaryMax:    job.SalaryMax,
			Requirements: job.Requirements,
			CreatedAt:    job.CreatedAt,
			CompanyName:  company.Name,
		}
		jobByID = append(jobByID, row3)
	}

	type Query struct {
		page         int32
		pageSize     int32
		id           int32
		name         string
		nameContains string
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK Name",
			query: Query{
				page:     1,
				pageSize: 10,
				name:     company.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				params := db.ListJobsByCompanyExactNameParams{
					Name:   company.Name,
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobsByExactName, nil)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobs(t, recorder.Body, jobsByExactName)
			},
		},
		{
			name: "OK ID",
			query: Query{
				page:     1,
				pageSize: 10,
				id:       company.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				params := db.ListJobsByCompanyIDParams{
					CompanyID: company.ID,
					Limit:     10,
					Offset:    0,
				}
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobByID, nil)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobs(t, recorder.Body, jobByID)
			},
		},
		{
			name: "OK Name Contains",
			query: Query{
				page:         1,
				pageSize:     10,
				nameContains: company.Name[1:3],
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				params := db.ListJobsByCompanyNameParams{
					Name:   company.Name[1:3],
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobByName, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobs(t, recorder.Body, jobByName)
			},
		},
		{
			name: "Invalid Page Size",
			query: Query{
				page:     1,
				pageSize: 50,
				id:       company.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page",
			query: Query{
				page:     0,
				pageSize: 10,
				id:       company.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No Parameters",
			query: Query{
				page:     1,
				pageSize: 10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "To Many Parameters",
			query: Query{
				page:     1,
				pageSize: 10,
				id:       company.ID,
				name:     company.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal Server Error ListJobsByCompanyID",
			query: Query{
				page:     1,
				pageSize: 10,
				id:       company.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobsByCompanyIDRow{}, sql.ErrConnDone)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error ListJobsByCompanyExactName",
			query: Query{
				page:     1,
				pageSize: 10,
				name:     company.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobsByCompanyExactNameRow{}, sql.ErrConnDone)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error ListJobsByCompanyExactName",
			query: Query{
				page:         1,
				pageSize:     10,
				nameContains: company.Name[1:3],
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListJobsByCompanyID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyExactName(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobsByCompanyName(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobsByCompanyNameRow{}, sql.ErrConnDone)
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

			url := "/api/v1/jobs/company"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query params
			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			q.Add("name", tc.query.name)
			q.Add("name_contains", tc.query.nameContains)
			q.Add("id", fmt.Sprintf("%d", tc.query.id))
			req.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestSearchJobsAPI(t *testing.T) {
	_, _, company := generateRandomEmployerAndCompany(t)
	var jobs []*esearch.Job
	title := utils.RandomString(5)
	industry := utils.RandomString(4)
	jobLocation := utils.RandomString(6)
	salaryMin := utils.RandomInt(100, 150)
	salaryMax := utils.RandomInt(151, 200)

	job := generateJob(
		title,
		industry,
		jobLocation,
		salaryMin,
		salaryMax,
	)

	for i := 0; i < 10; i++ {
		row := esearch.Job{
			ID:           job.ID,
			Title:        job.Title,
			Industry:     job.Industry,
			CompanyName:  company.Name,
			Description:  job.Description,
			Location:     job.Location,
			SalaryMin:    job.SalaryMin,
			SalaryMax:    job.SalaryMax,
			Requirements: job.Requirements,
		}
		jobs = append(jobs, &row)
	}

	var page int32 = 1
	var pageSize int32 = 10

	type Query struct {
		page     int32
		pageSize int32
		search   string
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(client *mockesearch.MockESearchClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				page:     page,
				pageSize: pageSize,
				search:   title,
			},
			buildStubs: func(client *mockesearch.MockESearchClient) {
				client.EXPECT().
					SearchJobs(gomock.Any(), gomock.Eq(title), gomock.Eq(page), gomock.Eq(pageSize)).
					Times(1).
					Return(jobs, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobs(t, recorder.Body, jobs)
			},
		},
		{
			name: "Internal Server Error",
			query: Query{
				page:     page,
				pageSize: pageSize,
				search:   title,
			},
			buildStubs: func(client *mockesearch.MockESearchClient) {
				client.EXPECT().
					SearchJobs(gomock.Any(), gomock.Eq(title), gomock.Eq(page), gomock.Eq(pageSize)).
					Times(1).
					Return([]*esearch.Job{}, errors.New("some error"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "No Page",
			query: Query{
				pageSize: pageSize,
				search:   title,
			},
			buildStubs: func(client *mockesearch.MockESearchClient) {
				client.EXPECT().
					SearchJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No Search",
			query: Query{
				pageSize: pageSize,
				page:     page,
			},
			buildStubs: func(client *mockesearch.MockESearchClient) {
				client.EXPECT().
					SearchJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No Page Size",
			query: Query{
				page:   page,
				search: title,
			},
			buildStubs: func(client *mockesearch.MockESearchClient) {
				client.EXPECT().
					SearchJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page Size",
			query: Query{
				pageSize: 50,
				search:   title,
			},
			buildStubs: func(client *mockesearch.MockESearchClient) {
				client.EXPECT().
					SearchJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page",
			query: Query{
				page:   0,
				search: title,
			},
			buildStubs: func(client *mockesearch.MockESearchClient) {
				client.EXPECT().
					SearchJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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

			client := mockesearch.NewMockESearchClient(ctrl)
			tc.buildStubs(client)

			server := newTestServer(t, store, client)
			recorder := httptest.NewRecorder()

			url := "/api/v1/jobs/search"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query params
			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			q.Add("search", tc.query.search)
			req.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func generateJob(title, industry, jobLocation string, salaryMin, salaryMax int32) db.Job {
	return db.Job{
		ID:           utils.RandomInt(1, 1000),
		Title:        title,
		Industry:     industry,
		Description:  utils.RandomString(5),
		Location:     jobLocation,
		SalaryMin:    salaryMin,
		SalaryMax:    salaryMax,
		Requirements: utils.RandomString(5),
	}
}

func generateRandomJob() db.Job {
	return db.Job{
		ID:           utils.RandomInt(1, 1000),
		Title:        utils.RandomString(4),
		Industry:     utils.RandomString(2),
		Description:  utils.RandomString(5),
		Location:     utils.RandomString(4),
		SalaryMin:    utils.RandomInt(100, 200),
		SalaryMax:    utils.RandomInt(201, 300),
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

func requireBodyMatchJobs(t *testing.T, body *bytes.Buffer, jobs interface{}) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	switch j := jobs.(type) {
	case []db.ListJobsByFiltersRow:
		var gotJobRows []db.ListJobsByFiltersRow
		err = json.Unmarshal(data, &gotJobRows)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i], gotJobRows[i])
		}
	case []db.ListJobsMatchingUserSkillsRow:
		var gotJobRows []db.ListJobsMatchingUserSkillsRow
		err = json.Unmarshal(data, &gotJobRows)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i], gotJobRows[i])
		}
	case []db.ListJobsByCompanyExactNameRow:
		var gotJobRows []db.ListJobsByCompanyExactNameRow
		err = json.Unmarshal(data, &gotJobRows)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i], gotJobRows[i])
		}
	case []db.ListJobsByCompanyNameRow:
		var gotJobRows []db.ListJobsByCompanyNameRow
		err = json.Unmarshal(data, &gotJobRows)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i], gotJobRows[i])
		}
	case []db.ListJobsByCompanyIDRow:
		var gotJobRows []db.ListJobsByCompanyIDRow
		err = json.Unmarshal(data, &gotJobRows)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i], gotJobRows[i])
		}
	case []*esearch.Job:
		var gotJobRows []esearch.Job
		err = json.Unmarshal(data, &gotJobRows)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i].Title, gotJobRows[i].Title)
			require.Equal(t, j[i].Industry, gotJobRows[i].Industry)
			require.Equal(t, j[i].Description, gotJobRows[i].Description)
			require.Equal(t, j[i].Location, gotJobRows[i].Location)
			require.Equal(t, j[i].SalaryMin, gotJobRows[i].SalaryMin)
			require.Equal(t, j[i].SalaryMax, gotJobRows[i].SalaryMax)
			require.Equal(t, j[i].Requirements, gotJobRows[i].Requirements)
			require.Equal(t, j[i].CompanyName, gotJobRows[i].CompanyName)
			require.Equal(t, j[i].ID, gotJobRows[i].ID)
			require.Equal(t, j[i].JobSkills, gotJobRows[i].JobSkills)
		}
	default:
		t.Fatalf("unsupported type %T", jobs)
	}
}
