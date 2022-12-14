// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rabbitmq/amqp091-go (interfaces: Acknowledger)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockAcknowledger is a mock of Acknowledger interface.
type MockAcknowledger struct {
	ctrl     *gomock.Controller
	recorder *MockAcknowledgerMockRecorder
}

// MockAcknowledgerMockRecorder is the mock recorder for MockAcknowledger.
type MockAcknowledgerMockRecorder struct {
	mock *MockAcknowledger
}

// NewMockAcknowledger creates a new mock instance.
func NewMockAcknowledger(ctrl *gomock.Controller) *MockAcknowledger {
	mock := &MockAcknowledger{ctrl: ctrl}
	mock.recorder = &MockAcknowledgerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAcknowledger) EXPECT() *MockAcknowledgerMockRecorder {
	return m.recorder
}

// Ack mocks base method.
func (m *MockAcknowledger) Ack(arg0 uint64, arg1 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ack", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ack indicates an expected call of Ack.
func (mr *MockAcknowledgerMockRecorder) Ack(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ack", reflect.TypeOf((*MockAcknowledger)(nil).Ack), arg0, arg1)
}

// Nack mocks base method.
func (m *MockAcknowledger) Nack(arg0 uint64, arg1, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Nack", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Nack indicates an expected call of Nack.
func (mr *MockAcknowledgerMockRecorder) Nack(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Nack", reflect.TypeOf((*MockAcknowledger)(nil).Nack), arg0, arg1, arg2)
}

// Reject mocks base method.
func (m *MockAcknowledger) Reject(arg0 uint64, arg1 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reject", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reject indicates an expected call of Reject.
func (mr *MockAcknowledgerMockRecorder) Reject(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reject", reflect.TypeOf((*MockAcknowledger)(nil).Reject), arg0, arg1)
}
