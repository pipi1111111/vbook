package ioc

import (
	"vbook/internal/service/sms"
	"vbook/internal/service/sms/localsms"
)

func InitSmsService() sms.Service {
	return localsms.NewService()
}
