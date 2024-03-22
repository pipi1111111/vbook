package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"vbook/internal/domain"
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

var redirtctURL = url.PathEscape(`https://meoying.com/oauth2/wechat/callback`)

type service struct {
	appID     string
	appSecret string
	client    *http.Client
}

func NewWechatService(appID string, appSecret string) Service {
	return &service{
		appID:     appID,
		client:    http.DefaultClient,
		appSecret: appSecret,
	}
}
func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	accessTokenUrl := fmt.Sprintf(`https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`,
		s.appID, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenUrl, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	httpResp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	var res Result
	err = json.NewDecoder(httpResp.Body).Decode(&res)
	if err != nil {
		//转JSON为结构体出错
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{},
			fmt.Errorf("调用微信接口失败 ErrCode %d,ErrMag %s", res.ErrCode, res.ErrMsg)
	}
	return domain.WechatInfo{
		UnionId: res.UnionId,
		OpenId:  res.OpenId,
	}, nil
}
func (s service) AuthURL(ctx context.Context, state string) (string, error) {
	const authUOLPattern = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`
	return fmt.Sprintf(authUOLPattern, s.appID, redirtctURL, state), nil
}

type Result struct {
	//接口调用凭证
	AccessToken string `json:"access_token"`
	//access_token接口调用凭证超时时间，单位（秒）
	ExpiresIn int64 `json:"expires_in"`
	//用户刷新access_token
	RefreshToken string `json:"refresh_token"`
	//授权用户唯一标识
	OpenId string `json:"openid"`
	//用户授权的作用域，使用逗号（,）分隔
	Scope string `json:"scope"`
	//当且仅当该网站应用已获得该用户的userinfo授权时，才会出现该字段。
	UnionId string `json:"unionid"`

	//错误返回
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
