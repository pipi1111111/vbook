package repository

import (
	"context"
	"database/sql"
	"log"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository/cache"
	"vbook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, ud domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, uid int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, id string) (domain.User, error)
}
type CacheUserRepository struct {
	ud    dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(ud dao.UserDao, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{
		ud:    ud,
		cache: cache,
	}
}
func (ur *CacheUserRepository) Create(ctx context.Context, ud domain.User) error {
	return ur.ud.Insert(ctx, ur.toDaoUser(ud))
}
func (ur *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := ur.ud.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(u), nil
}
func (ur *CacheUserRepository) toDaoUser(ud domain.User) dao.User {
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
		WechatUnionId: sql.NullString{
			String: ud.WechatInfo.UnionId,
			Valid:  true,
		},
		WechatOpenId: sql.NullString{
			String: ud.WechatInfo.OpenId,
			Valid:  true,
		},
	}
}

func (ur *CacheUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:        u.Id,
		Email:     u.Email.String,
		Password:  u.Password,
		Name:      u.Name,
		Phone:     u.Phone.String,
		Birthday:  time.UnixMilli(u.Birthday),
		Introduce: u.Introduce,
		Ctime:     time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenId:  u.WechatOpenId.String,
			UnionId: u.WechatUnionId.String,
		},
	}
}

func (ur *CacheUserRepository) Update(ctx context.Context, user domain.User) error {
	return ur.ud.UpdateById(ctx, ur.toDaoUser(user))
}
func (ur *CacheUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := ur.cache.Get(ctx, uid)
	if err == nil {
		return domain.User{}, err
	}
	u, err := ur.ud.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	du = ur.toDomain(u)
	err = ur.cache.Set(ctx, du)
	if err != nil {
		log.Println(err)
	}
	return du, nil
}

func (ur *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := ur.ud.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(u), nil
}
func (ur *CacheUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	ue, err := ur.ud.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(ue), nil
}
