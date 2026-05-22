package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	errWechatMiniLoginInvalidCode  = errors.New("invalid wechat mini login code")
	errWechatMiniLoginUnavailable  = errors.New("wechat mini login unavailable")
	errWechatMiniLoginUnauthorized = errors.New("wechat mini login unauthorized")
)

type WechatMiniSession struct {
	OpenID  string
	UnionID string
}

type WechatMiniSessionResolver interface {
	Resolve(ctx context.Context, code string) (WechatMiniSession, error)
}

type WechatCode2SessionResolver struct {
	appID      string
	secret     string
	endpoint   string
	httpClient *http.Client
}

func NewWechatCode2SessionResolver(appID string, secret string) (*WechatCode2SessionResolver, error) {
	appID = strings.TrimSpace(appID)
	secret = strings.TrimSpace(secret)
	if appID == "" || secret == "" {
		return nil, errWechatMiniLoginUnavailable
	}
	return &WechatCode2SessionResolver{
		appID:    appID,
		secret:   secret,
		endpoint: "https://api.weixin.qq.com/sns/jscode2session",
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}, nil
}

func (resolver *WechatCode2SessionResolver) Resolve(ctx context.Context, code string) (WechatMiniSession, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return WechatMiniSession{}, errWechatMiniLoginInvalidCode
	}
	values := url.Values{}
	values.Set("appid", resolver.appID)
	values.Set("secret", resolver.secret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")
	endpoint := resolver.endpoint
	if endpoint == "" {
		endpoint = "https://api.weixin.qq.com/sns/jscode2session"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return WechatMiniSession{}, errWechatMiniLoginUnavailable
	}
	response, err := resolver.httpClient.Do(request)
	if err != nil {
		return WechatMiniSession{}, errWechatMiniLoginUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return WechatMiniSession{}, errWechatMiniLoginUnavailable
	}
	var payload code2SessionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return WechatMiniSession{}, errWechatMiniLoginUnavailable
	}
	if payload.ErrCode != 0 {
		return WechatMiniSession{}, fmt.Errorf("%w: %d %s", errWechatMiniLoginUnauthorized, payload.ErrCode, payload.ErrMsg)
	}
	openID := strings.TrimSpace(payload.OpenID)
	if openID == "" {
		return WechatMiniSession{}, errWechatMiniLoginUnavailable
	}
	return WechatMiniSession{
		OpenID:  openID,
		UnionID: strings.TrimSpace(payload.UnionID),
	}, nil
}

type code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type DevWechatMiniSessionResolver struct{}

func (DevWechatMiniSessionResolver) Resolve(_ context.Context, code string) (WechatMiniSession, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return WechatMiniSession{}, errWechatMiniLoginInvalidCode
	}
	return WechatMiniSession{OpenID: "wx_" + shortHash(code)}, nil
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:16]
}
