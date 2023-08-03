package api

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/aalug/go-gin-job-search/db/mock"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/aalug/go-gin-job-search/utils"
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
			cv: nil,
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
			if tc.cv != nil {
				part, err := writer.CreateFormFile("cv", "test_file.pdf")
				require.NoError(t, err)
				part.Write(tc.cv)
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

func requireBodyMatchJobApplication(t *testing.T, body *bytes.Buffer, jobApplication db.JobApplication) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response jobApplicationResponse
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	require.Equal(t, response.ID, jobApplication.ID)
	require.Equal(t, response.JobID, jobApplication.JobID)
	require.Equal(t, response.Status, jobApplication.Status)

	expectedRounded := jobApplication.AppliedAt.Round(time.Microsecond)
	actualRounded := response.AppliedAt.Round(time.Microsecond)
	require.WithinDuration(t, expectedRounded, actualRounded, 1*time.Second)

	if jobApplication.Message.Valid {
		require.Equal(t, response.Message, jobApplication.Message.String)
	}
}

// errorReader is a io.Reader that always returns an error
type errorWriter struct{}

func (e errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("custom write error")
}
