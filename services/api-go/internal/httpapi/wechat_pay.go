package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

var errInvalidWechatSignature = errors.New("invalid wechat pay signature")

type WechatPayVerifier interface {
	Verify(req *http.Request, body []byte) error
}

type HMACWechatPayVerifier struct {
	secret []byte
}

func NewWechatPayVerifier(secret string) HMACWechatPayVerifier {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		secret = "infinitech-wechat-pay-dev-secret"
	}
	return HMACWechatPayVerifier{secret: []byte(secret)}
}

func (verifier HMACWechatPayVerifier) Verify(req *http.Request, body []byte) error {
	timestamp := strings.TrimSpace(req.Header.Get("Wechatpay-Timestamp"))
	nonce := strings.TrimSpace(req.Header.Get("Wechatpay-Nonce"))
	signature := strings.TrimSpace(req.Header.Get("Wechatpay-Signature"))
	if timestamp == "" || nonce == "" || signature == "" {
		return errInvalidWechatSignature
	}
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	mac := hmac.New(sha256.New, verifier.secret)
	mac.Write([]byte(message))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errInvalidWechatSignature
	}
	return nil
}
