package web

import (
	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
	"vbook/internal/domain"
	"vbook/internal/service"
)

const (
	emailRegexPattern = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = "^(?=.*[A-Za-z])(?=.*\\d)[A-Za-z\\d]{8,}$"
)

type UserHandler struct {
	emailRegexp    *regexp2.Regexp
	passwordRegexp *regexp2.Regexp
	us             service.UserService
}

func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{
		emailRegexp:    regexp2.MustCompile(emailRegexPattern, regexp2.None),
		passwordRegexp: regexp2.MustCompile(passwordRegexPattern, regexp2.None),
		us:             us,
	}
}
func (h *UserHandler) RegisterRouters(server *gin.Engine) {
	u := server.Group("/users")
	u.POST("/register", h.Register)
	u.POST("/login", h.Login)
	u.POST("/edit", h.Edit)
	u.GET("/view", h.View)
}

func (h *UserHandler) Register(ctx *gin.Context) {
	type Req struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		RePassword string `json:"rePassword"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	isEmail, err := h.emailRegexp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式不正确")
		return
	}
	if req.Password != req.RePassword {
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}
	isPassword, err := h.passwordRegexp.MatchString(req.RePassword)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须为大小写加数字不少于八位")
		return
	}
	err = h.us.Register(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱已经被注册")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	_, err := h.us.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvaliUserOrPassword:
		ctx.String(http.StatusOK, "账号或者密码不正确")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Edit(ctx *gin.Context) {

}

func (h *UserHandler) View(ctx *gin.Context) {

}
