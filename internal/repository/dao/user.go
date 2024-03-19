package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱已经被注册")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type User struct {
	Id        int64          `gorm:"primaryKey,autoIncrement"`
	Email     sql.NullString `gorm:"unique"`
	Password  string
	Name      string `gorm:"type=varchar(128)"`
	Birthday  int64
	Introduce string         `gorm:"type=varchar(4096)"`
	Phone     sql.NullString `gorm:"unique"`
	Ctime     int64
	Utime     int64
}
type UserDao interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdateById(ctx *gin.Context, user User) error
}
type GormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{
		db: db,
	}
}
func (ud *GormUserDao) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := ud.db.WithContext(ctx).Create(&u).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		const duplicateErr = 1062
		if me.Number == duplicateErr {
			return ErrDuplicateEmail
		}
	}
	return err
}
func (ud *GormUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("email = ?", email).Find(&u).Error
	return u, err
}
func (ud *GormUserDao) UpdateById(ctx *gin.Context, user User) error {
	return ud.db.WithContext(ctx).Model(&user).Where("id = ?", user.Id).Updates(map[string]any{
		"utime":     time.Now().UnixMilli(),
		"name":      user.Name,
		"birthday":  user.Birthday,
		"introduce": user.Introduce,
	}).Error
}
