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

type eqCreateEmployerParamsMatcher struct {
	params   db2.CreateEmployerParams
	password string
}

func (e eqCreateEmployerParamsMatcher) Matches(arg interface{}) bool {
	params, ok := arg.(db2.CreateEmployerParams)
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

func (e eqCreateEmployerParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.params, e.password)
}

func EqCreateEmployerParams(arg db2.CreateEmployerParams, password string) gomock.Matcher {
	return eqCreateEmployerParamsMatcher{arg, password}
}

func TestCreateEmployerAPI(t *testing.T) {
	employer, password, company := generateRandomEmployerAndCompany(t)

	requestBody := gin.H{
		"email":            employer.Email,
		"full_name":        employer.FullName,
		"password":         password,
		"company_name":     company.Name,
		"company_industry": company.Industry,
		"company_location": company.Location,
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
				companyParams := db2.CreateCompanyParams{
					Name:     company.Name,
					Industry: company.Industry,
					Location: company.Location,
				}
				store.EXPECT().
					CreateCompany(gomock.Any(), gomock.Eq(companyParams)).
					Times(1).
					Return(company, nil)
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(employer, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchEmployerAndCompany(t, recorder.Body, employer, company)
			},
		},
		{
			name: "Internal Server Error CreateCompany",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				companyParams := db2.CreateCompanyParams{
					Name:     company.Name,
					Industry: company.Industry,
					Location: company.Location,
				}
				store.EXPECT().
					CreateCompany(gomock.Any(), gomock.Eq(companyParams)).
					Times(1).
					Return(db2.Company{}, sql.ErrConnDone)
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error CreateEmployer",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				companyParams := db2.CreateCompanyParams{
					Name:     company.Name,
					Industry: company.Industry,
					Location: company.Location,
				}
				store.EXPECT().
					CreateCompany(gomock.Any(), gomock.Eq(companyParams)).
					Times(1).
					Return(company, nil)
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Employer{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Email",
			body: gin.H{
				"email":            "invalid",
				"full_name":        employer.FullName,
				"password":         password,
				"company_name":     company.Name,
				"company_industry": company.Industry,
				"company_location": company.Location,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Duplicated Company Name",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				params := db2.CreateCompanyParams{
					Name:     company.Name,
					Industry: company.Industry,
					Location: company.Location,
				}
				store.EXPECT().
					CreateCompany(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(db2.Company{}, &pq.Error{Code: "23505"})
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "Duplicated Email",
			body: requestBody,
			buildStubs: func(store *mockdb.MockStore) {
				params := db2.CreateCompanyParams{
					Name:     company.Name,
					Industry: company.Industry,
					Location: company.Location,
				}
				store.EXPECT().
					CreateCompany(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(company, nil)
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Employer{}, &pq.Error{Code: "23505"})
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := baseUrl + "/employers"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestLoginEmployerAPI(t *testing.T) {
	employer, password, company := generateRandomEmployerAndCompany(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"email":    employer.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Eq(employer.CompanyID)).
					Times(1).
					Return(company, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Employer Not Found",
			body: gin.H{
				"email":    employer.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db2.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetEmployerByEmail",
			body: gin.H{
				"email":    employer.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetCompanyByID",
			body: gin.H{
				"email":    employer.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Company{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Company Not Found",
			body: gin.H{
				"email":    employer.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Company{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Incorrect Password",
			body: gin.H{
				"email":    employer.Email,
				"password": fmt.Sprintf("%d, %s", utils2.RandomInt(1, 1000), utils2.RandomString(10)),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Password Too Short",
			body: gin.H{
				"email":    employer.Email,
				"password": "abc",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
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

			url := baseUrl + "/employers/login"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestGetEmployerAPI(t *testing.T) {
	employer, _, company := generateRandomEmployerAndCompany(t)
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
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Eq(employer.CompanyID)).
					Times(1).
					Return(company, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployerAndCompany(t, recorder.Body, employer, company)
			},
		},
		{
			name: "Unauthorized Only Employer Access",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetEmployerByEmail",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db2.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetCompanyByID",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Eq(employer.CompanyID)).
					Times(1).
					Return(db2.Company{}, sql.ErrConnDone)
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

			url := baseUrl + "/employers"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateEmployerAPI(t *testing.T) {
	employer, _, company := generateRandomEmployerAndCompany(t)
	newEmployer, _, newCompany := generateRandomEmployerAndCompany(t)
	user, _ := generateRandomUser(t)

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
				"full_name":        newEmployer.FullName,
				"email":            newEmployer.Email,
				"company_name":     newCompany.Name,
				"company_industry": newCompany.Industry,
				"company_location": newCompany.Location,
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
					GetCompanyByID(gomock.Any(), gomock.Eq(employer.CompanyID)).
					Times(1).
					Return(company, nil)
				companyParams := db2.UpdateCompanyParams{
					ID:       company.ID,
					Name:     newCompany.Name,
					Industry: newCompany.Industry,
					Location: newCompany.Location,
				}
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Eq(companyParams)).
					Times(1).
					Return(newCompany, nil)
				employerParams := db2.UpdateEmployerParams{
					ID:        employer.ID,
					CompanyID: employer.CompanyID,
					FullName:  newEmployer.FullName,
					Email:     newEmployer.Email,
				}
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Eq(employerParams)).
					Times(1).
					Return(newEmployer, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployerAndCompany(t, recorder.Body, newEmployer, newCompany)
			},
		},
		{
			name: "Unauthorized Only Employer Access",
			body: gin.H{
				"email":            newEmployer.Email,
				"company_industry": newCompany.Industry,
				"company_location": newCompany.Location,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetEmployerByEmail",
			body: gin.H{
				"email":            newEmployer.Email,
				"company_industry": newCompany.Industry,
				"company_location": newCompany.Location,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(db2.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetCompanyByID",
			body: gin.H{
				"full_name":        newEmployer.FullName,
				"company_location": newCompany.Location,
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
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Company{}, sql.ErrConnDone)
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error UpdateCompany",
			body: gin.H{
				"full_name":    newEmployer.FullName,
				"company_name": newCompany.Name,
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
					GetCompanyByID(gomock.Any(), gomock.Eq(employer.CompanyID)).
					Times(1).
					Return(company, nil)
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Company{}, sql.ErrConnDone)
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error UpdateEmployer",
			body: gin.H{
				"company_name": newCompany.Name,
				"full_name":    newEmployer.FullName,
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
					GetCompanyByID(gomock.Any(), gomock.Eq(employer.CompanyID)).
					Times(1).
					Return(company, nil)
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Any()).
					Times(1).
					Return(newCompany, nil)
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Employer{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "invalid Body",
			body: gin.H{
				"company_name": 123,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "invalid Email",
			body: gin.H{
				"email": "invalid",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetCompanyByID(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
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

			url := baseUrl + "/employers"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateEmployerPasswordAPI(t *testing.T) {
	employer, password, _ := generateRandomEmployerAndCompany(t)
	newPassword := utils2.RandomString(6)
	user, _ := generateRandomUser(t)

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
				"old_password": password,
				"new_password": newPassword,
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
					UpdateEmployerPassword(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Old Password Too Short",
			body: gin.H{
				"old_password": "123",
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateEmployerPassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "New Password Too Short",
			body: gin.H{
				"old_password": password,
				"new_password": "123",
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateEmployerPassword(gomock.Any(), gomock.Any()).
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
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					UpdateEmployerPassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error UpdateEmployerPassword",
			body: gin.H{
				"old_password": password,
				"new_password": newPassword,
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
					UpdateEmployerPassword(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Unauthorized Only Employer Access",
			body: gin.H{
				"old_password": password,
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					UpdateEmployerPassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetEmployerByEmail",
			body: gin.H{
				"old_password": password,
				"new_password": newPassword,
			},
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					UpdateEmployerPassword(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := baseUrl + "/employers/password"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteEmployerAPI(t *testing.T) {
	employer, _, company := generateRandomEmployerAndCompany(t)
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
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(employer.Email)).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					DeleteCompany(gomock.Any(), gomock.Eq(company.ID)).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteEmployer(gomock.Any(), gomock.Eq(employer.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name: "Unauthorized Only Employer Access",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db2.Employer{}, sql.ErrNoRows)
				store.EXPECT().
					DeleteCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Internal Server Error GetEmployerByEmail",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db2.Employer{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteCompany(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error DeleteCompany",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					DeleteCompany(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
				store.EXPECT().
					DeleteEmployer(gomock.Any(), gomock.Eq(employer.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal Server Error DeleteEmployer",
			setupAuth: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthorization(t, r, maker, authorizationTypeBearer, employer.Email, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployerByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(employer, nil)
				store.EXPECT().
					DeleteCompany(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
				store.EXPECT().
					DeleteEmployer(gomock.Any(), gomock.Eq(employer.ID)).
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

			url := baseUrl + "/employers"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}

// generateRandomEmployer create a random employer and company
func generateRandomEmployerAndCompany(t *testing.T) (db2.Employer, string, db2.Company) {
	password := utils2.RandomString(6)
	hashedPassword, err := utils2.HashPassword(password)
	require.NoError(t, err)

	company := db2.Company{
		ID:       utils2.RandomInt(1, 100),
		Name:     utils2.RandomString(5),
		Industry: utils2.RandomString(5),
		Location: utils2.RandomString(6),
	}

	employer := db2.Employer{
		ID:             utils2.RandomInt(1, 100),
		CompanyID:      company.ID,
		FullName:       utils2.RandomString(5),
		Email:          utils2.RandomEmail(),
		HashedPassword: hashedPassword,
		CreatedAt:      time.Now(),
	}

	return employer, password, company
}

// requireBodyMatchEmployerAndCompany checks if the body of the response matches the employer and company
func requireBodyMatchEmployerAndCompany(t *testing.T, body *bytes.Buffer, employer db2.Employer, company db2.Company) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response employerResponse
	err = json.Unmarshal(data, &response)

	require.NoError(t, err)
	require.NotZero(t, response.EmployerID)
	require.Equal(t, employer.Email, response.Email)
	require.Equal(t, employer.FullName, response.FullName)
	require.Equal(t, employer.CompanyID, response.CompanyID)
	require.Equal(t, company.Name, response.CompanyName)
	require.Equal(t, company.Industry, response.CompanyIndustry)
	require.Equal(t, company.Location, response.CompanyLocation)
	require.WithinDuration(t, employer.CreatedAt, response.EmployerCreatedAt, time.Second)
}
