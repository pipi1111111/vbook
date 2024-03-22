package web

import (
	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"time"
	"vbook/internal/domain"
	"vbook/internal/service"
	ijwt "vbook/internal/web/jwt"
)

const (
	emailRegexPattern = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = "^(?=.*[A-Za-z])(?=.*\\d)[A-Za-z\\d]{8,}$"
	bizLogin             = "login"
)

type UserHandler struct {
	ijwt.Handler
	emailRegexp    *regexp2.Regexp
	passwordRegexp *regexp2.Regexp
	us             service.UserService
	cs             service.CodeService
}

func NewUserHandler(us service.UserService, cs service.CodeService, hdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		emailRegexp:    regexp2.MustCompile(emailRegexPattern, regexp2.None),
		passwordRegexp: regexp2.MustCompile(passwordRegexPattern, regexp2.None),
		us:             us,
		cs:             cs,
		Handler:        hdl,
	}
}
func (h *UserHandler) RegisterRouters(server *gin.Engine) {
	u := server.Group("/users")
	u.POST("/register", h.Register)
	//u.POST("/login", h.Login)
	u.POST("/login", h.LoginJwt)
	u.POST("/edit", h.Edit)
	u.GET("/view", h.View)
	u.GET("/refresh_token", h.RefreshToken)
	u.POST("/sendSms", h.SendSMSLoginCode)
	u.POST("/loginSms", h.LoginSMS)
	u.GET("/logout", h.LogoutJWT)
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
	u, err := h.us.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			//十分组
			MaxAge:   600,
			HttpOnly: true,
		})
		err := sess.Save()
		if err != nil {
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvaliUserOrPassword:
		ctx.String(http.StatusOK, "账号或者密码不正确")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Name      string `json:"name"`
		Birthday  string `json:"birthday"`
		Introduce string `json:"introduce"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc, ok := ctx.MustGet("user").(ijwt.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式不正确")
		return
	}
	err = h.us.Update(ctx, domain.User{
		Id:        uc.Uid,
		Name:      req.Name,
		Birthday:  birthday,
		Introduce: req.Introduce,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "修改成功")

}

func (h *UserHandler) View(ctx *gin.Context) {
	uc, ok := ctx.MustGet("user").(ijwt.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	u, err := h.us.FindById(ctx, uc.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	type User struct {
		Email     string `json:"email"`
		Name      string `json:"name"`
		Introduce string `json:"introduce"`
		Birthday  string `json:"birthday"`
	}
	ctx.JSON(http.StatusOK, User{
		Email:     u.Email,
		Name:      u.Name,
		Introduce: u.Introduce,
		Birthday:  u.Birthday.Format(time.DateOnly),
	})
}

func (h *UserHandler) LoginJwt(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.us.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		err = h.SetLoginToken(ctx, u.Id)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvaliUserOrPassword:
		ctx.String(http.StatusOK, "账号或者密码不正确")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "请输入手机号"})
		return
	}
	err := h.cs.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{Msg: "发送成功"})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码发送太频繁"})
	default:
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		log.Println(err)
	}
}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := h.cs.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码不对，请重新输入"})
		return
	}
	u, err := h.us.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	err = h.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "登陆成功"})
}
func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	//约定 前端 Authorization里面带上这个refresh token
	tokenStr := h.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RCJWTKey, nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.CheckSession(ctx, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

//func (h *UserHandler) Logout(ctx *gin.Context) {
//	sess := sessions.Default(ctx)
//	sess.Options(sessions.Options{
//		MaxAge: -1,
//	})
//	sess.Save()
//}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录成功",
	})
}
