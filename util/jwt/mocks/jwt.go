// Code generated by MockGen. DO NOT EDIT.
// Source: ./jwt.go

// Package jwt_mocks is a generated GoMock package.
package jwt_mocks

import (
	context "context"
	reflect "reflect"

	jwt "github.com/golang-jwt/jwt/v5"
	gomock "github.com/golang/mock/gomock"
)

// MockJWTManager is a mock of JWTManager interface.
type MockJWTManager struct {
	ctrl     *gomock.Controller
	recorder *MockJWTManagerMockRecorder
}

// MockJWTManagerMockRecorder is the mock recorder for MockJWTManager.
type MockJWTManagerMockRecorder struct {
	mock *MockJWTManager
}

// NewMockJWTManager creates a new mock instance.
func NewMockJWTManager(ctrl *gomock.Controller) *MockJWTManager {
	mock := &MockJWTManager{ctrl: ctrl}
	mock.recorder = &MockJWTManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJWTManager) EXPECT() *MockJWTManagerMockRecorder {
	return m.recorder
}

// CreateToken mocks base method.
func (m *MockJWTManager) CreateToken(ctx context.Context, claims jwt.Claims) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateToken", ctx, claims)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateToken indicates an expected call of CreateToken.
func (mr *MockJWTManagerMockRecorder) CreateToken(ctx, claims interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateToken", reflect.TypeOf((*MockJWTManager)(nil).CreateToken), ctx, claims)
}

// ParseAndValidateToken mocks base method.
func (m *MockJWTManager) ParseAndValidateToken(ctx context.Context, tokenString string, claims jwt.Claims) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseAndValidateToken", ctx, tokenString, claims)
	ret0, _ := ret[0].(error)
	return ret0
}

// ParseAndValidateToken indicates an expected call of ParseAndValidateToken.
func (mr *MockJWTManagerMockRecorder) ParseAndValidateToken(ctx, tokenString, claims interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseAndValidateToken", reflect.TypeOf((*MockJWTManager)(nil).ParseAndValidateToken), ctx, tokenString, claims)
}