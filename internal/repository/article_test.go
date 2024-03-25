package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"vbook/internal/domain"
	"vbook/internal/repository/dao"
	daomocks "vbook/internal/repository/dao/mocks"
)

func Test_articleRepository_SyncV1(t *testing.T) {
	testCase := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.ArticleAuthorDao, dao.ArticleReaderDao)
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "新建同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDao, dao.ArticleReaderDao) {
				authorDao := daomocks.NewMockArticleAuthorDao(ctrl)
				authorDao.EXPECT().Create(gomock.Any(), dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(int64(1), nil)
				readerDao := daomocks.NewMockArticleReaderDao(ctrl)
				readerDao.EXPECT().Save(gomock.Any(), dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(nil)
				return authorDao, readerDao
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 1,
		},
		{
			name: "修改同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDao, dao.ArticleReaderDao) {
				authorDao := daomocks.NewMockArticleAuthorDao(ctrl)
				authorDao.EXPECT().Update(gomock.Any(), dao.Article{
					Id:       11,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(nil)
				readerDao := daomocks.NewMockArticleReaderDao(ctrl)
				readerDao.EXPECT().Save(gomock.Any(), dao.Article{
					Id:       11,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(nil)
				return authorDao, readerDao
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 11,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authorDao, readerDao := tc.mock(ctrl)
			repo := NewArticleRepositoryV2(readerDao, authorDao)
			id, err := repo.SyncV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
