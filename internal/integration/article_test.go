package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"vbook/internal/integration/startup"
	"vbook/internal/repository/dao"
	ijwt "vbook/internal/web/jwt"
)

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
type Res[T any] struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func (a *ArticleHandlerSuite) TearDownTest() {
	a.db.Exec("truncate table `articles`")
}
func (a *ArticleHandlerSuite) SetupSuite() {
	a.db = startup.InitDB()
	hdl := startup.InitArticleHandler()
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", ijwt.UserClaims{
			Uid: 123,
		})
	})
	hdl.RegisterRouters(server)
	a.server = server
}
func (a *ArticleHandlerSuite) TestEdit() {
	t := a.T()
	testCase := []struct {
		name       string
		before     func(t *testing.T)
		after      func(t *testing.T)
		art        Article
		wantCode   int
		wantResult Res[int64]
	}{
		{
			name:   "新建帖子",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				//你要验证 保存到了数据库中
				var art dao.Article
				err := a.db.Where("author_id = ?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的文章", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)

			},
			art: Article{
				Title:   "我的标题",
				Content: "我的文章",
			},
			wantCode: http.StatusOK,
			wantResult: Res[int64]{
				//我希望你的Id是1
				Data: 1,
			},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			a.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var res Res[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, res)
		})
	}
}
func TestArticleHandler(t *testing.T) {
	suite.Run(t, &ArticleHandlerSuite{})
}
