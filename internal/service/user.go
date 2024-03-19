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
	return domain.User{}, nil
}
func (us *userService) Update(ctx *gin.Context, user domain.User) error {
	return us.repo.Update(ctx, user)
}
