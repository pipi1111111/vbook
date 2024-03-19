package repository

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx *gin.Context, ud domain.User) error
	FindByEmail(ctx *gin.Context, email string) (domain.User, error)
}
type userRepository struct {
	ud dao.UserDao
}

func NewUserRepository(ud dao.UserDao) UserRepository {
	return &userRepository{
		ud: ud,
	}
}
func (ur *userRepository) Create(ctx *gin.Context, ud domain.User) error {
	return ur.ud.Insert(ctx, ur.toDaoUser(ud))
}
func (ur *userRepository) FindByEmail(ctx *gin.Context, email string) (domain.User, error) {
	u, err := ur.ud.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(u), nil
}
func (ur *userRepository) toDaoUser(ud domain.User) dao.User {
	return dao.User{
		Id: ud.Id,
		Email: sql.NullString{
			String: ud.Email,
			Valid:  true,
		},
		Password:  ud.Password,
		Name:      ud.Name,
		Birthday:  ud.Birthday.UnixMilli(),
		Introduce: ud.Introduce,
		Phone: sql.NullString{
			String: ud.Phone,
			Valid:  true,
		},
	}
}

func (ur *userRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:        u.Id,
		Email:     u.Email.String,
		Password:  u.Password,
		Name:      u.Name,
		Phone:     u.Phone.String,
		Birthday:  time.UnixMilli(u.Birthday),
		Introduce: u.Introduce,
		Ctime:     time.UnixMilli(u.Ctime),
	}
}
