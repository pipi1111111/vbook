package domain

import "time"

type User struct {
	Email      string
	Id         int64
	Password   string
	Phone      string
	Name       string
	Birthday   time.Time
	Introduce  string
	Ctime      time.Time
	Utime      time.Time
	WechatInfo WechatInfo
}
