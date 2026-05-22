package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWechatCode2SessionResolverCallsProvider(t *testing.T) {
	var receivedQuery map[string]string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		receivedQuery = map[string]string{
			"appid":      query.Get("appid"),
			"secret":     query.Get("secret"),
			"js_code":    query.Get("js_code"),
			"grant_type": query.Get("grant_type"),
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"openid":      "real_openid_1",
			"session_key": "session_key_1",
			"unionid":     "union_1",
		})
	}))
	defer upstream.Close()

	resolver, err := NewWechatCode2SessionResolver("wx_app_1", "secret_1")
	if err != nil {
		t.Fatal(err)
	}
	resolver.endpoint = upstream.URL
	resolver.httpClient = upstream.Client()

	session, err := resolver.Resolve(context.Background(), "js_code_1")
	if err != nil {
		t.Fatal(err)
	}
	if session.OpenID != "real_openid_1" || session.UnionID != "union_1" {
		t.Fatalf("unexpected session: %+v", session)
	}
	if receivedQuery["appid"] != "wx_app_1" || receivedQuery["secret"] != "secret_1" || receivedQuery["js_code"] != "js_code_1" || receivedQuery["grant_type"] != "authorization_code" {
		t.Fatalf("unexpected code2session query: %+v", receivedQuery)
	}
}

func TestWechatCode2SessionResolverRejectsProviderErrors(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 40029,
			"errmsg":  "invalid code",
		})
	}))
	defer upstream.Close()

	resolver, err := NewWechatCode2SessionResolver("wx_app_1", "secret_1")
	if err != nil {
		t.Fatal(err)
	}
	resolver.endpoint = upstream.URL
	resolver.httpClient = upstream.Client()

	if _, err := resolver.Resolve(context.Background(), "bad_code"); err == nil {
		t.Fatal("expected provider error")
	}
}
