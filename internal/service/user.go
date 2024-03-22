package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"vbook/internal/domain"
	"vbook/internal/repository"
)

var (
	ErrDuplicateEmail       = repository.ErrDuplicateUser
	ErrInvaliUserOrPassword = errors.New("用户不存在或者密码不对")
)

type UserService interface {
	Register(ctx *gin.Context, ud domain.User) error
	Login(ctx *gin.Context, email string, password string) (domain.User, error)
	Update(ctx *gin.Context, user domain.User) error
	FindById(ctx *gin.Context, uid int64) (domain.User, error)
	FindOrCreate(ctx *gin.Context, phone string) (domain.User, error)
	FindOrCreateWechat(ctx *gin.Context, info domain.WechatInfo) (domain.User, error)
}
type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (us *userService) Register(ctx *gin.Context, ud domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(ud.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	ud.Password = string(hash)
	return us.repo.Create(ctx, ud)
}
func (us *userService) Login(ctx *gin.Context, email string, password string) (domain.User, error) {
	u, err := us.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, err
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvaliUserOrPassword
	}
	return u, nil
}
func (us *userService) Update(ctx *gin.Context, user domain.User) error {
	return us.repo.Update(ctx, user)
}
func (us *userService) FindById(ctx *gin.Context, uid int64) (domain.User, error) {
	return us.repo.FindById(ctx, uid)
}
func (us *userService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	u, err := us.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	err = us.repo.Create(ctx, domain.User{Phone: phone})
	if err != nil && err == repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	return us.repo.FindByPhone(ctx, phone)
}
func (svc *userService) FindOrCreateWechat(ctx *gin.Context, info domain.WechatInfo) (domain.User, error) {
	//先找一找，我们认为大部分用户是已经存在的用户
	u, err := svc.repo.FindByWechat(ctx, info.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {

		return u, err
	}
	//如果没找到 意味着是一个新用户
	//JSON格式的wechatInfo
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: info,
	})
	if err != nil && !errors.Is(err, repository.ErrDuplicateUser) {
		return domain.User{}, err
	}

	return svc.repo.FindByWechat(ctx, info.OpenId)
}
