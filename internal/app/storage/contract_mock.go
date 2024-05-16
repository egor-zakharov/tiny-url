// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go

// Package storage is a generated GoMock package.
package storage

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockStorage) Add(ctx context.Context, shortURL, url, ID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, shortURL, url, ID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockStorageMockRecorder) Add(ctx, shortURL, url, ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockStorage)(nil).Add), ctx, shortURL, url, ID)
}

// AddBatch mocks base method.
func (m *MockStorage) AddBatch(ctx context.Context, URLs map[string]string, ID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddBatch", ctx, URLs, ID)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddBatch indicates an expected call of AddBatch.
func (mr *MockStorageMockRecorder) AddBatch(ctx, URLs, ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBatch", reflect.TypeOf((*MockStorage)(nil).AddBatch), ctx, URLs, ID)
}

// Backup mocks base method.
func (m *MockStorage) Backup() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Backup")
}

// Backup indicates an expected call of Backup.
func (mr *MockStorageMockRecorder) Backup() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Backup", reflect.TypeOf((*MockStorage)(nil).Backup))
}

// Get mocks base method.
func (m *MockStorage) Get(ctx context.Context, shortURL, ID string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, shortURL, ID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStorageMockRecorder) Get(ctx, shortURL, ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStorage)(nil).Get), ctx, shortURL, ID)
}
