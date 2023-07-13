// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aalug/go-gin-job-search/db/sqlc (interfaces: Store)

// Package mockdb is a generated GoMock package.
package mockdb

import (
	context "context"
	reflect "reflect"

	db "github.com/aalug/go-gin-job-search/db/sqlc"
	gomock "github.com/golang/mock/gomock"
)

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// CreateCompany mocks base method.
func (m *MockStore) CreateCompany(arg0 context.Context, arg1 db.CreateCompanyParams) (db.Company, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCompany", arg0, arg1)
	ret0, _ := ret[0].(db.Company)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCompany indicates an expected call of CreateCompany.
func (mr *MockStoreMockRecorder) CreateCompany(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCompany", reflect.TypeOf((*MockStore)(nil).CreateCompany), arg0, arg1)
}

// CreateEmployer mocks base method.
func (m *MockStore) CreateEmployer(arg0 context.Context, arg1 db.CreateEmployerParams) (db.Employer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEmployer", arg0, arg1)
	ret0, _ := ret[0].(db.Employer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEmployer indicates an expected call of CreateEmployer.
func (mr *MockStoreMockRecorder) CreateEmployer(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEmployer", reflect.TypeOf((*MockStore)(nil).CreateEmployer), arg0, arg1)
}

// CreateJob mocks base method.
func (m *MockStore) CreateJob(arg0 context.Context, arg1 db.CreateJobParams) (db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateJob", arg0, arg1)
	ret0, _ := ret[0].(db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateJob indicates an expected call of CreateJob.
func (mr *MockStoreMockRecorder) CreateJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateJob", reflect.TypeOf((*MockStore)(nil).CreateJob), arg0, arg1)
}

// CreateJobSkill mocks base method.
func (m *MockStore) CreateJobSkill(arg0 context.Context, arg1 db.CreateJobSkillParams) (db.JobSkill, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateJobSkill", arg0, arg1)
	ret0, _ := ret[0].(db.JobSkill)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateJobSkill indicates an expected call of CreateJobSkill.
func (mr *MockStoreMockRecorder) CreateJobSkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateJobSkill", reflect.TypeOf((*MockStore)(nil).CreateJobSkill), arg0, arg1)
}

// CreateMultipleJobSkills mocks base method.
func (m *MockStore) CreateMultipleJobSkills(arg0 context.Context, arg1 []string, arg2 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMultipleJobSkills", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateMultipleJobSkills indicates an expected call of CreateMultipleJobSkills.
func (mr *MockStoreMockRecorder) CreateMultipleJobSkills(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMultipleJobSkills", reflect.TypeOf((*MockStore)(nil).CreateMultipleJobSkills), arg0, arg1, arg2)
}

// CreateMultipleUserSkills mocks base method.
func (m *MockStore) CreateMultipleUserSkills(arg0 context.Context, arg1 []db.CreateMultipleUserSkillsParams, arg2 int32) ([]db.UserSkill, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMultipleUserSkills", arg0, arg1, arg2)
	ret0, _ := ret[0].([]db.UserSkill)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMultipleUserSkills indicates an expected call of CreateMultipleUserSkills.
func (mr *MockStoreMockRecorder) CreateMultipleUserSkills(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMultipleUserSkills", reflect.TypeOf((*MockStore)(nil).CreateMultipleUserSkills), arg0, arg1, arg2)
}

// CreateUser mocks base method.
func (m *MockStore) CreateUser(arg0 context.Context, arg1 db.CreateUserParams) (db.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0, arg1)
	ret0, _ := ret[0].(db.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockStoreMockRecorder) CreateUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockStore)(nil).CreateUser), arg0, arg1)
}

// CreateUserSkill mocks base method.
func (m *MockStore) CreateUserSkill(arg0 context.Context, arg1 db.CreateUserSkillParams) (db.UserSkill, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserSkill", arg0, arg1)
	ret0, _ := ret[0].(db.UserSkill)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUserSkill indicates an expected call of CreateUserSkill.
func (mr *MockStoreMockRecorder) CreateUserSkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserSkill", reflect.TypeOf((*MockStore)(nil).CreateUserSkill), arg0, arg1)
}

// DeleteAllUserSkills mocks base method.
func (m *MockStore) DeleteAllUserSkills(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAllUserSkills", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAllUserSkills indicates an expected call of DeleteAllUserSkills.
func (mr *MockStoreMockRecorder) DeleteAllUserSkills(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllUserSkills", reflect.TypeOf((*MockStore)(nil).DeleteAllUserSkills), arg0, arg1)
}

// DeleteCompany mocks base method.
func (m *MockStore) DeleteCompany(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCompany", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCompany indicates an expected call of DeleteCompany.
func (mr *MockStoreMockRecorder) DeleteCompany(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCompany", reflect.TypeOf((*MockStore)(nil).DeleteCompany), arg0, arg1)
}

// DeleteEmployer mocks base method.
func (m *MockStore) DeleteEmployer(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEmployer", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEmployer indicates an expected call of DeleteEmployer.
func (mr *MockStoreMockRecorder) DeleteEmployer(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEmployer", reflect.TypeOf((*MockStore)(nil).DeleteEmployer), arg0, arg1)
}

// DeleteJob mocks base method.
func (m *MockStore) DeleteJob(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJob", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJob indicates an expected call of DeleteJob.
func (mr *MockStoreMockRecorder) DeleteJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJob", reflect.TypeOf((*MockStore)(nil).DeleteJob), arg0, arg1)
}

// DeleteJobPosting mocks base method.
func (m *MockStore) DeleteJobPosting(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJobPosting", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJobPosting indicates an expected call of DeleteJobPosting.
func (mr *MockStoreMockRecorder) DeleteJobPosting(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJobPosting", reflect.TypeOf((*MockStore)(nil).DeleteJobPosting), arg0, arg1)
}

// DeleteJobSkill mocks base method.
func (m *MockStore) DeleteJobSkill(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJobSkill", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJobSkill indicates an expected call of DeleteJobSkill.
func (mr *MockStoreMockRecorder) DeleteJobSkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJobSkill", reflect.TypeOf((*MockStore)(nil).DeleteJobSkill), arg0, arg1)
}

// DeleteJobSkillsByJobID mocks base method.
func (m *MockStore) DeleteJobSkillsByJobID(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJobSkillsByJobID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJobSkillsByJobID indicates an expected call of DeleteJobSkillsByJobID.
func (mr *MockStoreMockRecorder) DeleteJobSkillsByJobID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJobSkillsByJobID", reflect.TypeOf((*MockStore)(nil).DeleteJobSkillsByJobID), arg0, arg1)
}

// DeleteMultipleJobSkills mocks base method.
func (m *MockStore) DeleteMultipleJobSkills(arg0 context.Context, arg1 []int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMultipleJobSkills", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMultipleJobSkills indicates an expected call of DeleteMultipleJobSkills.
func (mr *MockStoreMockRecorder) DeleteMultipleJobSkills(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMultipleJobSkills", reflect.TypeOf((*MockStore)(nil).DeleteMultipleJobSkills), arg0, arg1)
}

// DeleteMultipleUserSkills mocks base method.
func (m *MockStore) DeleteMultipleUserSkills(arg0 context.Context, arg1 []int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMultipleUserSkills", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMultipleUserSkills indicates an expected call of DeleteMultipleUserSkills.
func (mr *MockStoreMockRecorder) DeleteMultipleUserSkills(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMultipleUserSkills", reflect.TypeOf((*MockStore)(nil).DeleteMultipleUserSkills), arg0, arg1)
}

// DeleteUser mocks base method.
func (m *MockStore) DeleteUser(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockStoreMockRecorder) DeleteUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockStore)(nil).DeleteUser), arg0, arg1)
}

// DeleteUserSkill mocks base method.
func (m *MockStore) DeleteUserSkill(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUserSkill", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUserSkill indicates an expected call of DeleteUserSkill.
func (mr *MockStoreMockRecorder) DeleteUserSkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUserSkill", reflect.TypeOf((*MockStore)(nil).DeleteUserSkill), arg0, arg1)
}

// GetCompanyByID mocks base method.
func (m *MockStore) GetCompanyByID(arg0 context.Context, arg1 int32) (db.Company, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCompanyByID", arg0, arg1)
	ret0, _ := ret[0].(db.Company)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCompanyByID indicates an expected call of GetCompanyByID.
func (mr *MockStoreMockRecorder) GetCompanyByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCompanyByID", reflect.TypeOf((*MockStore)(nil).GetCompanyByID), arg0, arg1)
}

// GetCompanyByName mocks base method.
func (m *MockStore) GetCompanyByName(arg0 context.Context, arg1 string) (db.Company, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCompanyByName", arg0, arg1)
	ret0, _ := ret[0].(db.Company)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCompanyByName indicates an expected call of GetCompanyByName.
func (mr *MockStoreMockRecorder) GetCompanyByName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCompanyByName", reflect.TypeOf((*MockStore)(nil).GetCompanyByName), arg0, arg1)
}

// GetEmployerByEmail mocks base method.
func (m *MockStore) GetEmployerByEmail(arg0 context.Context, arg1 string) (db.Employer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEmployerByEmail", arg0, arg1)
	ret0, _ := ret[0].(db.Employer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEmployerByEmail indicates an expected call of GetEmployerByEmail.
func (mr *MockStoreMockRecorder) GetEmployerByEmail(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEmployerByEmail", reflect.TypeOf((*MockStore)(nil).GetEmployerByEmail), arg0, arg1)
}

// GetEmployerByID mocks base method.
func (m *MockStore) GetEmployerByID(arg0 context.Context, arg1 int32) (db.Employer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEmployerByID", arg0, arg1)
	ret0, _ := ret[0].(db.Employer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEmployerByID indicates an expected call of GetEmployerByID.
func (mr *MockStoreMockRecorder) GetEmployerByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEmployerByID", reflect.TypeOf((*MockStore)(nil).GetEmployerByID), arg0, arg1)
}

// GetJob mocks base method.
func (m *MockStore) GetJob(arg0 context.Context, arg1 int32) (db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJob", arg0, arg1)
	ret0, _ := ret[0].(db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJob indicates an expected call of GetJob.
func (mr *MockStoreMockRecorder) GetJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJob", reflect.TypeOf((*MockStore)(nil).GetJob), arg0, arg1)
}

// GetUserByEmail mocks base method.
func (m *MockStore) GetUserByEmail(arg0 context.Context, arg1 string) (db.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", arg0, arg1)
	ret0, _ := ret[0].(db.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByEmail indicates an expected call of GetUserByEmail.
func (mr *MockStoreMockRecorder) GetUserByEmail(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockStore)(nil).GetUserByEmail), arg0, arg1)
}

// GetUserByID mocks base method.
func (m *MockStore) GetUserByID(arg0 context.Context, arg1 int32) (db.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByID", arg0, arg1)
	ret0, _ := ret[0].(db.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByID indicates an expected call of GetUserByID.
func (mr *MockStoreMockRecorder) GetUserByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByID", reflect.TypeOf((*MockStore)(nil).GetUserByID), arg0, arg1)
}

// GetUserDetailsByEmail mocks base method.
func (m *MockStore) GetUserDetailsByEmail(arg0 context.Context, arg1 string) (db.User, []db.UserSkill, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserDetailsByEmail", arg0, arg1)
	ret0, _ := ret[0].(db.User)
	ret1, _ := ret[1].([]db.UserSkill)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetUserDetailsByEmail indicates an expected call of GetUserDetailsByEmail.
func (mr *MockStoreMockRecorder) GetUserDetailsByEmail(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserDetailsByEmail", reflect.TypeOf((*MockStore)(nil).GetUserDetailsByEmail), arg0, arg1)
}

// ListJobSkillsByJobID mocks base method.
func (m *MockStore) ListJobSkillsByJobID(arg0 context.Context, arg1 db.ListJobSkillsByJobIDParams) ([]db.ListJobSkillsByJobIDRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobSkillsByJobID", arg0, arg1)
	ret0, _ := ret[0].([]db.ListJobSkillsByJobIDRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobSkillsByJobID indicates an expected call of ListJobSkillsByJobID.
func (mr *MockStoreMockRecorder) ListJobSkillsByJobID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobSkillsByJobID", reflect.TypeOf((*MockStore)(nil).ListJobSkillsByJobID), arg0, arg1)
}

// ListJobsByCompanyExactName mocks base method.
func (m *MockStore) ListJobsByCompanyExactName(arg0 context.Context, arg1 db.ListJobsByCompanyExactNameParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsByCompanyExactName", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsByCompanyExactName indicates an expected call of ListJobsByCompanyExactName.
func (mr *MockStoreMockRecorder) ListJobsByCompanyExactName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsByCompanyExactName", reflect.TypeOf((*MockStore)(nil).ListJobsByCompanyExactName), arg0, arg1)
}

// ListJobsByCompanyID mocks base method.
func (m *MockStore) ListJobsByCompanyID(arg0 context.Context, arg1 db.ListJobsByCompanyIDParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsByCompanyID", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsByCompanyID indicates an expected call of ListJobsByCompanyID.
func (mr *MockStoreMockRecorder) ListJobsByCompanyID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsByCompanyID", reflect.TypeOf((*MockStore)(nil).ListJobsByCompanyID), arg0, arg1)
}

// ListJobsByCompanyName mocks base method.
func (m *MockStore) ListJobsByCompanyName(arg0 context.Context, arg1 db.ListJobsByCompanyNameParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsByCompanyName", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsByCompanyName indicates an expected call of ListJobsByCompanyName.
func (mr *MockStoreMockRecorder) ListJobsByCompanyName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsByCompanyName", reflect.TypeOf((*MockStore)(nil).ListJobsByCompanyName), arg0, arg1)
}

// ListJobsByIndustry mocks base method.
func (m *MockStore) ListJobsByIndustry(arg0 context.Context, arg1 db.ListJobsByIndustryParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsByIndustry", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsByIndustry indicates an expected call of ListJobsByIndustry.
func (mr *MockStoreMockRecorder) ListJobsByIndustry(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsByIndustry", reflect.TypeOf((*MockStore)(nil).ListJobsByIndustry), arg0, arg1)
}

// ListJobsByLocation mocks base method.
func (m *MockStore) ListJobsByLocation(arg0 context.Context, arg1 db.ListJobsByLocationParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsByLocation", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsByLocation indicates an expected call of ListJobsByLocation.
func (mr *MockStoreMockRecorder) ListJobsByLocation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsByLocation", reflect.TypeOf((*MockStore)(nil).ListJobsByLocation), arg0, arg1)
}

// ListJobsBySalaryRange mocks base method.
func (m *MockStore) ListJobsBySalaryRange(arg0 context.Context, arg1 db.ListJobsBySalaryRangeParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsBySalaryRange", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsBySalaryRange indicates an expected call of ListJobsBySalaryRange.
func (mr *MockStoreMockRecorder) ListJobsBySalaryRange(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsBySalaryRange", reflect.TypeOf((*MockStore)(nil).ListJobsBySalaryRange), arg0, arg1)
}

// ListJobsBySkill mocks base method.
func (m *MockStore) ListJobsBySkill(arg0 context.Context, arg1 db.ListJobsBySkillParams) ([]int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsBySkill", arg0, arg1)
	ret0, _ := ret[0].([]int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsBySkill indicates an expected call of ListJobsBySkill.
func (mr *MockStoreMockRecorder) ListJobsBySkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsBySkill", reflect.TypeOf((*MockStore)(nil).ListJobsBySkill), arg0, arg1)
}

// ListJobsByTitle mocks base method.
func (m *MockStore) ListJobsByTitle(arg0 context.Context, arg1 db.ListJobsByTitleParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsByTitle", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsByTitle indicates an expected call of ListJobsByTitle.
func (mr *MockStoreMockRecorder) ListJobsByTitle(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsByTitle", reflect.TypeOf((*MockStore)(nil).ListJobsByTitle), arg0, arg1)
}

// ListJobsMatchingUserSkills mocks base method.
func (m *MockStore) ListJobsMatchingUserSkills(arg0 context.Context, arg1 db.ListJobsMatchingUserSkillsParams) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobsMatchingUserSkills", arg0, arg1)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobsMatchingUserSkills indicates an expected call of ListJobsMatchingUserSkills.
func (mr *MockStoreMockRecorder) ListJobsMatchingUserSkills(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobsMatchingUserSkills", reflect.TypeOf((*MockStore)(nil).ListJobsMatchingUserSkills), arg0, arg1)
}

// ListUserSkills mocks base method.
func (m *MockStore) ListUserSkills(arg0 context.Context, arg1 db.ListUserSkillsParams) ([]db.UserSkill, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUserSkills", arg0, arg1)
	ret0, _ := ret[0].([]db.UserSkill)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUserSkills indicates an expected call of ListUserSkills.
func (mr *MockStoreMockRecorder) ListUserSkills(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUserSkills", reflect.TypeOf((*MockStore)(nil).ListUserSkills), arg0, arg1)
}

// ListUsersBySkill mocks base method.
func (m *MockStore) ListUsersBySkill(arg0 context.Context, arg1 db.ListUsersBySkillParams) ([]db.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsersBySkill", arg0, arg1)
	ret0, _ := ret[0].([]db.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsersBySkill indicates an expected call of ListUsersBySkill.
func (mr *MockStoreMockRecorder) ListUsersBySkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsersBySkill", reflect.TypeOf((*MockStore)(nil).ListUsersBySkill), arg0, arg1)
}

// UpdateCompany mocks base method.
func (m *MockStore) UpdateCompany(arg0 context.Context, arg1 db.UpdateCompanyParams) (db.Company, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCompany", arg0, arg1)
	ret0, _ := ret[0].(db.Company)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateCompany indicates an expected call of UpdateCompany.
func (mr *MockStoreMockRecorder) UpdateCompany(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCompany", reflect.TypeOf((*MockStore)(nil).UpdateCompany), arg0, arg1)
}

// UpdateEmployer mocks base method.
func (m *MockStore) UpdateEmployer(arg0 context.Context, arg1 db.UpdateEmployerParams) (db.Employer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateEmployer", arg0, arg1)
	ret0, _ := ret[0].(db.Employer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateEmployer indicates an expected call of UpdateEmployer.
func (mr *MockStoreMockRecorder) UpdateEmployer(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEmployer", reflect.TypeOf((*MockStore)(nil).UpdateEmployer), arg0, arg1)
}

// UpdateEmployerPassword mocks base method.
func (m *MockStore) UpdateEmployerPassword(arg0 context.Context, arg1 db.UpdateEmployerPasswordParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateEmployerPassword", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateEmployerPassword indicates an expected call of UpdateEmployerPassword.
func (mr *MockStoreMockRecorder) UpdateEmployerPassword(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEmployerPassword", reflect.TypeOf((*MockStore)(nil).UpdateEmployerPassword), arg0, arg1)
}

// UpdateJob mocks base method.
func (m *MockStore) UpdateJob(arg0 context.Context, arg1 db.UpdateJobParams) (db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJob", arg0, arg1)
	ret0, _ := ret[0].(db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateJob indicates an expected call of UpdateJob.
func (mr *MockStoreMockRecorder) UpdateJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJob", reflect.TypeOf((*MockStore)(nil).UpdateJob), arg0, arg1)
}

// UpdateJobSkill mocks base method.
func (m *MockStore) UpdateJobSkill(arg0 context.Context, arg1 db.UpdateJobSkillParams) (db.JobSkill, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobSkill", arg0, arg1)
	ret0, _ := ret[0].(db.JobSkill)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateJobSkill indicates an expected call of UpdateJobSkill.
func (mr *MockStoreMockRecorder) UpdateJobSkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobSkill", reflect.TypeOf((*MockStore)(nil).UpdateJobSkill), arg0, arg1)
}

// UpdatePassword mocks base method.
func (m *MockStore) UpdatePassword(arg0 context.Context, arg1 db.UpdatePasswordParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePassword", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePassword indicates an expected call of UpdatePassword.
func (mr *MockStoreMockRecorder) UpdatePassword(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockStore)(nil).UpdatePassword), arg0, arg1)
}

// UpdateUser mocks base method.
func (m *MockStore) UpdateUser(arg0 context.Context, arg1 db.UpdateUserParams) (db.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", arg0, arg1)
	ret0, _ := ret[0].(db.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockStoreMockRecorder) UpdateUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockStore)(nil).UpdateUser), arg0, arg1)
}

// UpdateUserSkill mocks base method.
func (m *MockStore) UpdateUserSkill(arg0 context.Context, arg1 db.UpdateUserSkillParams) (db.UserSkill, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserSkill", arg0, arg1)
	ret0, _ := ret[0].(db.UserSkill)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUserSkill indicates an expected call of UpdateUserSkill.
func (mr *MockStoreMockRecorder) UpdateUserSkill(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserSkill", reflect.TypeOf((*MockStore)(nil).UpdateUserSkill), arg0, arg1)
}
