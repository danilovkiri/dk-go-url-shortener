// Package mocks is a generated GoMock package.
// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1 (interfaces: URLStorage)
package mocks

import (
	context "context"
	reflect "reflect"

	modelurl "github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	modelstorage "github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/modelstorage"
	gomock "github.com/golang/mock/gomock"
)

// MockURLStorage is a mock of URLStorage interface.
type MockURLStorage struct {
	ctrl     *gomock.Controller
	recorder *MockURLStorageMockRecorder
}

// MockURLStorageMockRecorder is the mock recorder for MockURLStorage.
type MockURLStorageMockRecorder struct {
	mock *MockURLStorage
}

// NewMockURLStorage creates a new mock instance.
func NewMockURLStorage(ctrl *gomock.Controller) *MockURLStorage {
	mock := &MockURLStorage{ctrl: ctrl}
	mock.recorder = &MockURLStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockURLStorage) EXPECT() *MockURLStorageMockRecorder {
	return m.recorder
}

// CloseDB mocks base method.
func (m *MockURLStorage) CloseDB() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseDB")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseDB indicates an expected call of CloseDB.
func (mr *MockURLStorageMockRecorder) CloseDB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseDB", reflect.TypeOf((*MockURLStorage)(nil).CloseDB))
}

// DeleteBatch mocks base method.
func (m *MockURLStorage) DeleteBatch(arg0 context.Context, arg1 []string, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBatch", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBatch indicates an expected call of DeleteBatch.
func (mr *MockURLStorageMockRecorder) DeleteBatch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBatch", reflect.TypeOf((*MockURLStorage)(nil).DeleteBatch), arg0, arg1, arg2)
}

// Dump mocks base method.
func (m *MockURLStorage) Dump(arg0 context.Context, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dump", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Dump indicates an expected call of Dump.
func (mr *MockURLStorageMockRecorder) Dump(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dump", reflect.TypeOf((*MockURLStorage)(nil).Dump), arg0, arg1, arg2, arg3)
}

// GetStats mocks base method.
func (m *MockURLStorage) GetStats(arg0 context.Context) (int64, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStats", arg0)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetStats indicates an expected call of GetStats.
func (mr *MockURLStorageMockRecorder) GetStats(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStats", reflect.TypeOf((*MockURLStorage)(nil).GetStats), arg0)
}

// PingDB mocks base method.
func (m *MockURLStorage) PingDB() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingDB")
	ret0, _ := ret[0].(error)
	return ret0
}

// PingDB indicates an expected call of PingDB.
func (mr *MockURLStorageMockRecorder) PingDB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingDB", reflect.TypeOf((*MockURLStorage)(nil).PingDB))
}

// Retrieve mocks base method.
func (m *MockURLStorage) Retrieve(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Retrieve", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Retrieve indicates an expected call of Retrieve.
func (mr *MockURLStorageMockRecorder) Retrieve(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Retrieve", reflect.TypeOf((*MockURLStorage)(nil).Retrieve), arg0, arg1)
}

// RetrieveByUserID mocks base method.
func (m *MockURLStorage) RetrieveByUserID(arg0 context.Context, arg1 string) ([]modelurl.FullURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetrieveByUserID", arg0, arg1)
	ret0, _ := ret[0].([]modelurl.FullURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RetrieveByUserID indicates an expected call of RetrieveByUserID.
func (mr *MockURLStorageMockRecorder) RetrieveByUserID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetrieveByUserID", reflect.TypeOf((*MockURLStorage)(nil).RetrieveByUserID), arg0, arg1)
}

// SendToQueue mocks base method.
func (m *MockURLStorage) SendToQueue(arg0 modelstorage.URLChannelEntry) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendToQueue", arg0)
}

// SendToQueue indicates an expected call of SendToQueue.
func (mr *MockURLStorageMockRecorder) SendToQueue(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendToQueue", reflect.TypeOf((*MockURLStorage)(nil).SendToQueue), arg0)
}
