// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aalug/go-gin-job-search/esearch (interfaces: ESearchClient)

// Package mockesearch is a generated GoMock package.
package mockesearch

import (
	context "context"
	reflect "reflect"

	esearch "github.com/aalug/go-gin-job-search/esearch"
	gomock "github.com/golang/mock/gomock"
)

// MockESearchClient is a mock of ESearchClient interface.
type MockESearchClient struct {
	ctrl     *gomock.Controller
	recorder *MockESearchClientMockRecorder
}

// MockESearchClientMockRecorder is the mock recorder for MockESearchClient.
type MockESearchClientMockRecorder struct {
	mock *MockESearchClient
}

// NewMockESearchClient creates a new mock instance.
func NewMockESearchClient(ctrl *gomock.Controller) *MockESearchClient {
	mock := &MockESearchClient{ctrl: ctrl}
	mock.recorder = &MockESearchClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockESearchClient) EXPECT() *MockESearchClientMockRecorder {
	return m.recorder
}

// DeleteJobDocument mocks base method.
func (m *MockESearchClient) DeleteJobDocument(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJobDocument", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJobDocument indicates an expected call of DeleteJobDocument.
func (mr *MockESearchClientMockRecorder) DeleteJobDocument(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJobDocument", reflect.TypeOf((*MockESearchClient)(nil).DeleteJobDocument), arg0)
}

// GetDocumentIDByJobID mocks base method.
func (m *MockESearchClient) GetDocumentIDByJobID(arg0 int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDocumentIDByJobID", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDocumentIDByJobID indicates an expected call of GetDocumentIDByJobID.
func (mr *MockESearchClientMockRecorder) GetDocumentIDByJobID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDocumentIDByJobID", reflect.TypeOf((*MockESearchClient)(nil).GetDocumentIDByJobID), arg0)
}

// IndexJobAsDocument mocks base method.
func (m *MockESearchClient) IndexJobAsDocument(arg0 int, arg1 esearch.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IndexJobAsDocument", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// IndexJobAsDocument indicates an expected call of IndexJobAsDocument.
func (mr *MockESearchClientMockRecorder) IndexJobAsDocument(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IndexJobAsDocument", reflect.TypeOf((*MockESearchClient)(nil).IndexJobAsDocument), arg0, arg1)
}

// IndexJobsAsDocuments mocks base method.
func (m *MockESearchClient) IndexJobsAsDocuments(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IndexJobsAsDocuments", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// IndexJobsAsDocuments indicates an expected call of IndexJobsAsDocuments.
func (mr *MockESearchClientMockRecorder) IndexJobsAsDocuments(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IndexJobsAsDocuments", reflect.TypeOf((*MockESearchClient)(nil).IndexJobsAsDocuments), arg0)
}

// SearchJobs mocks base method.
func (m *MockESearchClient) SearchJobs(arg0 context.Context, arg1 string, arg2, arg3 int32) ([]*esearch.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchJobs", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*esearch.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchJobs indicates an expected call of SearchJobs.
func (mr *MockESearchClientMockRecorder) SearchJobs(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchJobs", reflect.TypeOf((*MockESearchClient)(nil).SearchJobs), arg0, arg1, arg2, arg3)
}

// UpdateJobDocument mocks base method.
func (m *MockESearchClient) UpdateJobDocument(arg0 string, arg1 esearch.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobDocument", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobDocument indicates an expected call of UpdateJobDocument.
func (mr *MockESearchClientMockRecorder) UpdateJobDocument(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobDocument", reflect.TypeOf((*MockESearchClient)(nil).UpdateJobDocument), arg0, arg1)
}
