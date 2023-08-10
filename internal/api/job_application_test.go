package api

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aalug/go-gin-job-search/internal/db/mock"
	db "github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/pkg/token"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateJobApplicationAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, _ := generateRandomEmployerAndCompany(t)
	job := generateRandomJob()

	fakeFileSize := 10 * 1024
	fakeFileData := make([]byte, fakeFileSize)
	_, err := rand.Read(fakeFileData)
	require.NoError(t, err)

	message := utils.RandomString(5)

	jobApplication := db.JobApplication{
		ID:     utils.RandomInt(1, 1000),
		UserID: user.ID,
		JobID:  job.ID,
		Message: sql.NullString{
			String: message,
			Valid:  len(message) > 0,
		},
		Cv:        fakeFileData,
		Status:    db.ApplicationStatusApplied,
		AppliedAt: time.Now(),
	}

	type body struct {
		Message string `json:"message"`
		JobID   int32  `json:"job_id"`
	}

	testCases := []struct {
		name          string
		body          body
		cv            []byte
		setupAuth     func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: body{
				Message: message,
				JobID:   job.ID,
			},
			cv: fakeFileData,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				params := db.CreateJobApplicationParams{
					UserID: user.ID,
					JobID:  job.ID,
					Message: sql.NullString{
						String: message,
						Valid:  true,
					},
					Cv: fakeFileData,
				}
				store.EXPECT().
					CreateJobApplication(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobApplication, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchJobApplication(t, recorder.Body, jobApplication)
			},
		},
		{
			name: "Unauthorized",
			body: body{
				Message: message,
				JobID:   job.ID,
			},
			cv: fakeFileData,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					CreateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetUserByEmail",
			body: body{
				Message: message,
				JobID:   job.ID,
			},
			cv: fakeFileData,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					CreateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "No CV File",
			body: body{
				Message: message,
				JobID:   job.ID,
			},
			cv: []byte{},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Job ID",
			body: body{
				Message: message,
				JobID:   0,
			},
			cv: fakeFileData,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateJobApplication",
			body: body{
				Message: message,
				JobID:   job.ID,
			},
			cv: fakeFileData,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateJobApplication(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.JobApplication{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Unique Constraint Violated",
			body: body{
				Message: message,
				JobID:   job.ID,
			},
			cv: fakeFileData,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateJobApplication(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.JobApplication{}, &pq.Error{Code: "23505"})
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := "/api/v1/job-applications"

			formData := &bytes.Buffer{}
			writer := multipart.NewWriter(formData)

			// Add the message field to the request body.
			err := writer.WriteField("message", tc.body.Message)
			require.NoError(t, err)

			// Add the job_id
			err = writer.WriteField("job_id", fmt.Sprintf("%d", tc.body.JobID))
			require.NoError(t, err)

			// Add the CV file
			if len(tc.cv) > 0 {
				part, err := writer.CreateFormFile("cv", "test_file.pdf")
				require.NoError(t, err)
				_, err = part.Write(tc.cv)
				require.NoError(t, err)
			}
			writer.Close()

			// Create the HTTP request with the updated request body.
			req, err := http.NewRequest(http.MethodPost, url, formData)
			require.NoError(t, err)

			req.Header.Set("Content-Type", writer.FormDataContentType())

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestGetJobApplicationForUserAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, company := generateRandomEmployerAndCompany(t)
	job := generateRandomJob()
	jobApplicationID := utils.RandomInt(1, 1000)

	fakeFileSize := 10 * 1024
	fakeFileData := make([]byte, fakeFileSize)
	_, err := rand.Read(fakeFileData)
	require.NoError(t, err)

	getJobApplicationForUserRow := db.GetJobApplicationForUserRow{
		ApplicationID:      jobApplicationID,
		JobID:              job.ID,
		JobTitle:           job.Title,
		CompanyName:        company.Name,
		ApplicationStatus:  db.ApplicationStatusApplied,
		ApplicationDate:    time.Now(),
		ApplicationMessage: sql.NullString{},
		UserCv:             fakeFileData,
		UserID:             user.ID,
	}

	testCases := []struct {
		name             string
		JobApplicationID int32
		setupAuth        func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs       func(store *mockdb.MockStore)
		checkResponse    func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:             "OK",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				getJobApplicationForUserRow.ApplicationMessage.Valid = true
				getJobApplicationForUserRow.ApplicationMessage.String = utils.RandomString(5)
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(getJobApplicationForUserRow, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobApplication(t, recorder.Body, getJobApplicationForUserRow)
			},
		},
		{
			name:             "Invalid Job Application ID",
			JobApplicationID: 0,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:             "Unauthorized",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetUserByEmail",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Not Found",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetJobApplicationForUserRow{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetJobApplicationForUser",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetJobApplicationForUserRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetJobApplicationForUser",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetJobApplicationForUserRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Forbidden Not Owner",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)

				// change userID so that the job application does not belong to the user
				getJobApplicationForUserRow.UserID = user.ID + 1
				store.EXPECT().
					GetJobApplicationForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(getJobApplicationForUserRow, nil)
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

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/job-applications/user/%d", tc.JobApplicationID)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)

			if tc.name == "OK" {
				rec := httptest.NewRecorder()
				url := fmt.Sprintf("/assets/cvs/cv_%d.pdf", jobApplicationID)
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)

				server.router.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)

				contentType := rec.Header().Get("Content-Type")
				require.Equal(t, "application/pdf", contentType)

				contentDisposition := rec.Header().Get("Content-Disposition")
				expectedContentDisposition := fmt.Sprintf("inline; filename=cv_%d.pdf", jobApplicationID)
				require.Equal(t, expectedContentDisposition, contentDisposition)
			}
		})
	}
}

func TestGetJobApplicationForEmployerAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, company := generateRandomEmployerAndCompany(t)
	job := generateRandomJob()
	jobApplicationID := utils.RandomInt(1, 1000)

	fakeFileSize := 10 * 1024
	fakeFileData := make([]byte, fakeFileSize)
	_, err := rand.Read(fakeFileData)
	require.NoError(t, err)

	getJobApplicationForEmployerRow := db.GetJobApplicationForEmployerRow{
		ApplicationID:      jobApplicationID,
		JobTitle:           job.Title,
		JobID:              job.ID,
		ApplicationStatus:  db.ApplicationStatusSeen,
		ApplicationDate:    time.Now(),
		ApplicationMessage: sql.NullString{},
		UserCv:             fakeFileData,
		UserID:             user.ID,
		UserEmail:          user.Email,
		UserFullName:       user.FullName,
		UserLocation:       user.Location,
		CompanyID:          company.ID,
	}

	testCases := []struct {
		name             string
		JobApplicationID int32
		setupAuth        func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs       func(store *mockdb.MockStore)
		checkResponse    func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:             "OK",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				getJobApplicationForEmployerRow.ApplicationMessage.Valid = true
				getJobApplicationForEmployerRow.ApplicationMessage.String = utils.RandomString(5)
				getJobApplicationForEmployerRow.ApplicationStatus = db.ApplicationStatusApplied
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(getJobApplicationForEmployerRow, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(company.ID, nil)
				params := db.UpdateJobApplicationStatusParams{
					ID:     jobApplicationID,
					Status: db.ApplicationStatusSeen,
				}
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobApplication(t, recorder.Body, getJobApplicationForEmployerRow)
			},
		},
		{
			name:             "Invalid Job Application ID",
			JobApplicationID: 0,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(0)
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:             "Unauthorized",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetEmployerByEmail",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Not Found",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetJobApplicationForEmployerRow{}, sql.ErrNoRows)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetJobApplicationForEmployer",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(db.GetJobApplicationForEmployerRow{}, sql.ErrConnDone)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Forbidden Only Owner",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(getJobApplicationForEmployerRow, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(company.ID+1, nil)
				// +1 to raise error because of this id is different from employer.CompanyID
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetCompanyIDOfJob",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(getJobApplicationForEmployerRow, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(int32(0), sql.ErrConnDone)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error UpdateJobApplicationStatus",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				getJobApplicationForEmployerRow.ApplicationStatus = db.ApplicationStatusApplied
				store.EXPECT().
					GetJobApplicationForEmployer(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(getJobApplicationForEmployerRow, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(company.ID, nil)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/job-applications/employer/%d", tc.JobApplicationID)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)

			if tc.name == "OK" {
				rec := httptest.NewRecorder()
				url := fmt.Sprintf("/assets/cvs/cv_%d.pdf", jobApplicationID)
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)

				server.router.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)

				contentType := rec.Header().Get("Content-Type")
				require.Equal(t, "application/pdf", contentType)

				contentDisposition := rec.Header().Get("Content-Disposition")
				expectedContentDisposition := fmt.Sprintf("inline; filename=cv_%d.pdf", jobApplicationID)
				require.Equal(t, expectedContentDisposition, contentDisposition)
			}
		})
	}
}

func TestChangeJobApplicationStatusAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, company := generateRandomEmployerAndCompany(t)
	jobApplicationID := utils.RandomInt(1, 1000)

	testCases := []struct {
		name             string
		JobApplicationID int32
		body             gin.H
		setupAuth        func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs       func(store *mockdb.MockStore)
		checkResponse    func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:             "OK",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(company.ID, nil)
				params := db.UpdateJobApplicationStatusParams{
					ID:     jobApplicationID,
					Status: db.ApplicationStatusRejected,
				}
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				var response changeJobApplicationStatusResponse
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				require.Equal(t, response.ApplicationID, jobApplicationID)
				require.Equal(t, db.ApplicationStatusRejected, response.Status)
				require.Equal(t, "Status updated successfully", response.Message)
			},
		},
		{
			name:             "Invalid Status",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Invalid",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:             "Invalid Job Application ID",
			JobApplicationID: 0,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:             "Unauthorized",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetEmployerByEmail",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Job Application Not Found",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int32(0), sql.ErrNoRows)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetCompanyIDOfJob",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int32(0), sql.ErrConnDone)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Forbidden Employer Not Job Owner",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(company.ID+1, nil)
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:             "Forbidden Employer Not Job Owner",
			JobApplicationID: jobApplicationID,
			body: gin.H{
				"new_status": "Rejected",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(company.ID, nil)
				params := db.UpdateJobApplicationStatusParams{
					ID:     jobApplicationID,
					Status: db.ApplicationStatusRejected,
				}
				store.EXPECT().
					UpdateJobApplicationStatus(gomock.Any(), gomock.Eq(params)).
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

			url := fmt.Sprintf("/api/v1/job-applications/employer/%d/status", tc.JobApplicationID)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateJobApplicationAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, _ := generateRandomEmployerAndCompany(t)
	job := generateRandomJob()

	fakeFileSize := 10 * 1024
	fakeFileData := make([]byte, fakeFileSize)
	_, err := rand.Read(fakeFileData)
	require.NoError(t, err)

	message := utils.RandomString(5)

	jobApplication := db.JobApplication{
		ID:     utils.RandomInt(1, 1000),
		UserID: user.ID,
		JobID:  job.ID,
		Message: sql.NullString{
			String: message,
			Valid:  true,
		},
		Cv:        fakeFileData,
		Status:    db.ApplicationStatusApplied,
		AppliedAt: time.Now(),
	}

	getJobApplicationUserIDAndStatusRow := db.GetJobApplicationUserIDAndStatusRow{
		UserID: user.ID,
		Status: db.ApplicationStatusApplied,
	}

	testCases := []struct {
		name             string
		JobApplicationID int32
		message          string
		cv               []byte
		cvProvided       string
		setupAuth        func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs       func(store *mockdb.MockStore)
		checkResponse    func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:             "OK",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cv:               fakeFileData,
			cvProvided:       "true",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Eq(jobApplication.ID)).
					Times(1).
					Return(getJobApplicationUserIDAndStatusRow, nil)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(1).
					Return(jobApplication, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobApplication(t, recorder.Body, jobApplication)
			},
		},
		{
			name:             "Invalid Job Application ID",
			JobApplicationID: 0,
			message:          message,
			cv:               fakeFileData,
			cvProvided:       "true",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:             "Unauthorized Only Users Can Access",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cv:               fakeFileData,
			cvProvided:       "1",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetUserByEmail",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cvProvided:       "false",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Job Application Not Found",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cv:               fakeFileData,
			cvProvided:       "1",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Eq(jobApplication.ID)).
					Times(1).
					Return(db.GetJobApplicationUserIDAndStatusRow{}, sql.ErrNoRows)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetJobApplicationUserIDAndStatus",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cv:               fakeFileData,
			cvProvided:       "1",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Eq(jobApplication.ID)).
					Times(1).
					Return(db.GetJobApplicationUserIDAndStatusRow{}, sql.ErrConnDone)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Forbidden Application Seen",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cv:               fakeFileData,
			cvProvided:       "1",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Eq(jobApplication.ID)).
					Times(1).
					Return(db.GetJobApplicationUserIDAndStatusRow{
						UserID: user.ID,
						Status: db.ApplicationStatusSeen,
					}, nil)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:             "Forbidden User Not Application Owner",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cv:               fakeFileData,
			cvProvided:       "1",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Eq(jobApplication.ID)).
					Times(1).
					Return(db.GetJobApplicationUserIDAndStatusRow{
						UserID: user.ID + 1,
						Status: db.ApplicationStatusApplied,
					}, nil)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error UpdateJobApplication",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cvProvided:       "0",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Eq(jobApplication.ID)).
					Times(1).
					Return(getJobApplicationUserIDAndStatusRow, nil)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.JobApplication{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "CV Provided True But No File",
			JobApplicationID: jobApplication.ID,
			message:          message,
			cvProvided:       "1",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserIDAndStatus(gomock.Any(), gomock.Eq(jobApplication.ID)).
					Times(1).
					Return(getJobApplicationUserIDAndStatusRow, nil)
				store.EXPECT().
					UpdateJobApplication(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/job-applications/user/%d", tc.JobApplicationID)
			formData := &bytes.Buffer{}

			writer := multipart.NewWriter(formData)

			if tc.message != "" {
				// Add the message field to the request body.
				err := writer.WriteField("message", tc.message)
				require.NoError(t, err)
			}
			if tc.cvProvided == "true" || tc.cvProvided == "1" {
				err := writer.WriteField("cv_provided", tc.cvProvided)
				require.NoError(t, err)

				// Add the CV file
				// additional check to test error that is returned
				// when cv_provided is true, but no file was provided
				if len(tc.cv) > 0 {
					part, err := writer.CreateFormFile("cv", "test_file.pdf")
					require.NoError(t, err)

					_, err = part.Write(tc.cv)
					require.NoError(t, err)
				}
			}
			writer.Close()

			// Create the HTTP request with the updated request body.
			req, err := http.NewRequest(http.MethodPatch, url, formData)
			require.NoError(t, err)

			req.Header.Set("Content-Type", writer.FormDataContentType())

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteJobApplicationAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, _ := generateRandomEmployerAndCompany(t)
	jobApplicationID := utils.RandomInt(1, 1000)

	testCases := []struct {
		name             string
		JobApplicationID int32
		setupAuth        func(t *testing.T, r *http.Request, maker token.Maker)
		buildStubs       func(store *mockdb.MockStore)
		checkResponse    func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:             "OK",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(user.ID, nil)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:             "Invalid Job Application ID",
			JobApplicationID: 0,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:             "Unauthorized Only User Access",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetUserByEmail",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Job Application Not Found",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(int32(0), sql.ErrNoRows)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error GetJobApplicationUserID",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(int32(0), sql.ErrConnDone)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:             "Forbidden User Not Owner",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(user.ID+1, nil)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:             "Internal Server Error DeleteJobApplication",
			JobApplicationID: jobApplicationID,
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetJobApplicationUserID(gomock.Any(), gomock.Eq(jobApplicationID)).
					Times(1).
					Return(user.ID, nil)
				store.EXPECT().
					DeleteJobApplication(gomock.Any(), gomock.Eq(jobApplicationID)).
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

			url := fmt.Sprintf("/api/v1/job-applications/user/%d", tc.JobApplicationID)

			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestLustJobApplicationsForUserAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, _ := generateRandomEmployerAndCompany(t)

	var jobApplications []db.ListJobApplicationsForUserRow
	for i := 0; i < 10; i++ {
		jobApplications = append(jobApplications, db.ListJobApplicationsForUserRow{
			UserID:            user.ID,
			ApplicationID:     int32(i),
			JobTitle:          utils.RandomString(5),
			JobID:             int32(i * 2),
			CompanyName:       utils.RandomString(4),
			ApplicationStatus: db.ApplicationStatusApplied,
			ApplicationDate:   time.Now().Add(-time.Hour * time.Duration(i)),
		})
	}

	type Query struct {
		page     int
		pageSize int
		sort     string // 'date-asc' or 'date-desc'
		status   db.ApplicationStatus
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
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
				params := db.ListJobApplicationsForUserParams{
					UserID:        user.ID,
					Limit:         10,
					Offset:        0,
					FilterStatus:  true,
					Status:        db.ApplicationStatusApplied,
					AppliedAtAsc:  false,
					AppliedAtDesc: true,
				}
				store.EXPECT().
					ListJobApplicationsForUser(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobApplications, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobApplications(t, recorder.Body, jobApplications)
			},
		},
		{
			name: "Invalid Page",
			query: Query{
				page:     0,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page Size",
			query: Query{
				page:     1,
				pageSize: 50,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Sort",
			query: Query{
				page:     1,
				pageSize: 10,
				sort:     "invalid",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Sort",
			query: Query{
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   "invalid",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unauthorized Only Users Access",
			query: Query{
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
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
					ListJobApplicationsForUser(gomock.Any(), gomock.Any()).
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
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					ListJobApplicationsForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetUserByEmail",
			query: Query{
				page:     1,
				pageSize: 10,
				sort:     "date-asc",
				status:   db.ApplicationStatusApplied,
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
					ListJobApplicationsForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobApplicationsForUserRow{}, sql.ErrConnDone)
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

			url := "/api/v1/job-applications/user"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// add query params
			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			q.Add("sort", tc.query.sort)
			q.Add("status", fmt.Sprintf("%s", tc.query.status))
			req.URL.RawQuery = q.Encode()

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestLustJobApplicationsForEmployerAPI(t *testing.T) {
	user, _ := generateRandomUser(t)
	employer, _, company := generateRandomEmployerAndCompany(t)
	job := generateRandomJob()

	var jobApplications []db.ListJobApplicationsForEmployerRow
	for i := 0; i < 10; i++ {
		jobApplications = append(jobApplications, db.ListJobApplicationsForEmployerRow{
			ApplicationID:     int32(i),
			UserID:            user.ID,
			UserEmail:         user.Email,
			UserFullName:      user.FullName,
			ApplicationStatus: db.ApplicationStatusApplied,
			ApplicationDate:   time.Now().Add(-time.Hour * time.Duration(i)),
		})
	}

	type Query struct {
		jobID    int32
		page     int
		pageSize int
		sort     string // 'date-asc' or 'date-desc'
		status   db.ApplicationStatus
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
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(company.ID, nil)
				params := db.ListJobApplicationsForEmployerParams{
					JobID:         job.ID,
					Limit:         10,
					Offset:        0,
					FilterStatus:  true,
					Status:        db.ApplicationStatusApplied,
					AppliedAtAsc:  false,
					AppliedAtDesc: true,
				}
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(jobApplications, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobApplications(t, recorder.Body, jobApplications)
			},
		},
		{
			name: "Invalid Page",
			query: Query{
				jobID:    job.ID,
				page:     0,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Page Size",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 50,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Sort",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "invalid",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Invalid Status",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   "invalid",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "No Job ID",
			query: Query{
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unauthorized Only Employer Access",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetEmployerByEmail",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Job Not Found",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(int32(0), sql.ErrNoRows)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetCompanyIDOfJob",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-desc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int32(0), sql.ErrConnDone)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Employer Not Owner Of Job",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-asc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(company.ID+1, nil)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "Internal Server Error ListJobApplicationsForEmployer",
			query: Query{
				jobID:    job.ID,
				page:     1,
				pageSize: 10,
				sort:     "date-asc",
				status:   db.ApplicationStatusApplied,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyIDOfJob(gomock.Any(), gomock.Eq(job.ID)).
					Times(1).
					Return(company.ID, nil)
				store.EXPECT().
					ListJobApplicationsForEmployer(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListJobApplicationsForEmployerRow{}, sql.ErrConnDone)
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

			url := "/api/v1/job-applications/employer"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// add query params
			q := req.URL.Query()
			q.Add("job_id", fmt.Sprintf("%d", tc.query.jobID))
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			q.Add("sort", tc.query.sort)
			q.Add("status", fmt.Sprintf("%s", tc.query.status))
			req.URL.RawQuery = q.Encode()

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchJobApplication(t *testing.T, body *bytes.Buffer, jobApplication interface{}) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	switch ja := jobApplication.(type) {
	case db.JobApplication:
		var response jobApplicationResponse
		err = json.Unmarshal(data, &response)
		require.NoError(t, err)

		require.Equal(t, response.ID, ja.ID)
		require.Equal(t, response.JobID, ja.JobID)
		require.Equal(t, response.Status, ja.Status)

		expectedRounded := ja.AppliedAt.Round(time.Microsecond)
		actualRounded := response.AppliedAt.Round(time.Microsecond)
		require.WithinDuration(t, expectedRounded, actualRounded, 1*time.Second)

		if ja.Message.Valid {
			require.Equal(t, response.Message, ja.Message.String)
		}
	case db.GetJobApplicationForUserRow:
		var response getJobApplicationForUserResponse
		err = json.Unmarshal(data, &response)
		require.NoError(t, err)

		require.Equal(t, response.ApplicationID, ja.ApplicationID)
		require.Equal(t, response.JobID, ja.JobID)
		require.Equal(t, response.JobTitle, ja.JobTitle)
		require.Equal(t, response.CompanyName, ja.CompanyName)
		require.Equal(t, response.ApplicationStatus, ja.ApplicationStatus)
		require.WithinDuration(t, response.ApplicationDate, ja.ApplicationDate, 1*time.Second)
		if ja.ApplicationMessage.Valid {
			require.Equal(t, response.ApplicationMessage, ja.ApplicationMessage.String)
		}
	case db.GetJobApplicationForEmployerRow:
		var response getJobApplicationForEmployerResponse
		err = json.Unmarshal(data, &response)
		require.NoError(t, err)

		require.Equal(t, response.ApplicationID, ja.ApplicationID)
		require.Equal(t, response.JobID, ja.JobID)
		require.Equal(t, response.JobTitle, ja.JobTitle)
		require.Equal(t, response.UserFullName, ja.UserFullName)
		require.Equal(t, response.UserEmail, ja.UserEmail)
		require.Equal(t, response.UserLocation, ja.UserLocation)
		require.Equal(t, response.UserID, ja.UserID)
		require.WithinDuration(t, response.ApplicationDate, ja.ApplicationDate, 1*time.Second)
		if ja.ApplicationMessage.Valid {
			require.Equal(t, response.ApplicationMessage, ja.ApplicationMessage.String)
		}
	default:
		t.Fatalf("unsupported type %T", jobApplication)
	}
}

func requireBodyMatchJobApplications(t *testing.T, body *bytes.Buffer, jobApplication interface{}) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	switch j := jobApplication.(type) {
	case []db.ListJobApplicationsForUserRow:
		var response []db.ListJobApplicationsForUserRow
		err = json.Unmarshal(data, &response)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i].JobID, response[i].JobID)
			require.Equal(t, j[i].UserID, response[i].UserID)
			require.Equal(t, j[i].ApplicationID, response[i].ApplicationID)
			require.Equal(t, j[i].JobTitle, response[i].JobTitle)
			require.Equal(t, j[i].CompanyName, response[i].CompanyName)
			require.Equal(t, j[i].ApplicationStatus, response[i].ApplicationStatus)
			require.Equal(t, j[i].ApplicationStatus, response[i].ApplicationStatus)
			require.WithinDuration(t, j[i].ApplicationDate, response[i].ApplicationDate, 1*time.Second)
		}
	case []db.ListJobApplicationsForEmployerRow:
		var response []db.ListJobApplicationsForEmployerRow
		err = json.Unmarshal(data, &response)
		require.NoError(t, err)

		for i := 0; i < len(j); i++ {
			require.Equal(t, j[i].UserID, response[i].UserID)
			require.Equal(t, j[i].ApplicationID, response[i].ApplicationID)
			require.Equal(t, j[i].UserFullName, response[i].UserFullName)
			require.Equal(t, j[i].UserEmail, response[i].UserEmail)
			require.Equal(t, j[i].ApplicationStatus, response[i].ApplicationStatus)
			require.WithinDuration(t, j[i].ApplicationDate, response[i].ApplicationDate, 1*time.Second)
		}
	default:
		t.Fatalf("unsupported type %T", jobApplication)
	}
}
