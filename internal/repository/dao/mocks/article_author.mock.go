// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/repository/dao/article_author.go
//
// Generated by this command:
//
//	mockgen -source=./internal/repository/dao/article_author.go -package=daomocks -destination=./internal/repository/dao/mocks/article_author.mock.go
//

// Package daomocks is a generated GoMock package.
package daomocks

import (
	context "context"
	reflect "reflect"
	dao "vbook/internal/repository/dao"

	gomock "go.uber.org/mock/gomock"
)

// MockArticleAuthorDao is a mock of ArticleAuthorDao interface.
type MockArticleAuthorDao struct {
	ctrl     *gomock.Controller
	recorder *MockArticleAuthorDaoMockRecorder
}

// MockArticleAuthorDaoMockRecorder is the mock recorder for MockArticleAuthorDao.
type MockArticleAuthorDaoMockRecorder struct {
	mock *MockArticleAuthorDao
}

// NewMockArticleAuthorDao creates a new mock instance.
func NewMockArticleAuthorDao(ctrl *gomock.Controller) *MockArticleAuthorDao {
	mock := &MockArticleAuthorDao{ctrl: ctrl}
	mock.recorder = &MockArticleAuthorDaoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArticleAuthorDao) EXPECT() *MockArticleAuthorDaoMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockArticleAuthorDao) Create(ctx context.Context, art dao.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockArticleAuthorDaoMockRecorder) Create(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockArticleAuthorDao)(nil).Create), ctx, art)
}

// Update mocks base method.
func (m *MockArticleAuthorDao) Update(ctx context.Context, art dao.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, art)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockArticleAuthorDaoMockRecorder) Update(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockArticleAuthorDao)(nil).Update), ctx, art)
}
