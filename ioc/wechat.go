package ioc

import (
	"os"
	"vbook/internal/service/oauth2/wechat"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到环境变量WECHAT_APP_SECRET")
	}
	return wechat.NewWechatService(appId, appSecret)
}
