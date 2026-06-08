package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"infinitech2/services/api-go/internal/httpapi"
	"infinitech2/services/api-go/internal/platform"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store, closeStore := newRepository(ctx)
	defer closeStore()
	authSessionOption, closeAuthSessions := newAuthSessionOption(ctx)
	defer closeAuthSessions()
	server := &http.Server{
		Addr: ":" + env("PORT", "1029"),
		Handler: httpapi.NewRouter(
			store,
			httpapi.WithWechatMiniSessionResolver(newWechatMiniResolver()),
			authSessionOption,
			newAdminLoginOption(),
			httpapi.WithDevBearerAuth(devBearerAuthEnabled()),
			newNotificationProviderCallbackSecretOption(),
		),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("api-go listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func newWechatMiniResolver() httpapi.WechatMiniSessionResolver {
	if os.Getenv("WECHAT_MINI_LOGIN_MODE") == "dev" {
		log.Print("api-go using development WeChat mini login resolver")
		return httpapi.DevWechatMiniSessionResolver{}
	}
	resolver, err := httpapi.NewWechatCode2SessionResolver(os.Getenv("WECHAT_MINI_APP_ID"), os.Getenv("WECHAT_MINI_APP_SECRET"))
	if err != nil {
		log.Fatal("configure WECHAT_MINI_APP_ID and WECHAT_MINI_APP_SECRET, or set WECHAT_MINI_LOGIN_MODE=dev for local development")
	}
	log.Print("api-go using WeChat code2session login resolver")
	return resolver
}

func devBearerAuthEnabled() bool {
	return os.Getenv("WECHAT_MINI_LOGIN_MODE") == "dev" || os.Getenv("ALLOW_DEV_BEARER_AUTH") == "true"
}

func newAdminLoginOption() httpapi.RouterOption {
	accountID := strings.TrimSpace(os.Getenv("ADMIN_BOOTSTRAP_ACCOUNT_ID"))
	password := strings.TrimSpace(os.Getenv("ADMIN_BOOTSTRAP_PASSWORD"))
	if accountID == "" && password == "" {
		log.Print("api-go admin password login is not configured; set ADMIN_BOOTSTRAP_ACCOUNT_ID and ADMIN_BOOTSTRAP_PASSWORD to enable it")
		return func(*httpapi.Router) {}
	}
	if accountID == "" || password == "" {
		log.Fatal("configure both ADMIN_BOOTSTRAP_ACCOUNT_ID and ADMIN_BOOTSTRAP_PASSWORD, or unset both to disable admin password login")
	}
	if len(password) < 8 || len(password) > 72 {
		log.Fatal("ADMIN_BOOTSTRAP_PASSWORD length must be between 8 and 72 bytes")
	}
	log.Print("api-go admin password login configured from environment")
	return httpapi.WithAdminLoginCredential(accountID, password)
}

func newNotificationProviderCallbackSecretOption() httpapi.RouterOption {
	secret := strings.TrimSpace(os.Getenv("NOTIFICATION_PROVIDER_CALLBACK_SECRET"))
	if secret == "" {
		log.Print("api-go notification provider callback secret is empty for local development; set NOTIFICATION_PROVIDER_CALLBACK_SECRET for production callback verification")
	}
	return httpapi.WithNotificationProviderCallbackSecret(secret)
}

func newAuthSessionOption(ctx context.Context) (httpapi.RouterOption, func()) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Print("api-go using in-memory auth sessions; set DATABASE_URL to persist login sessions")
		return func(*httpapi.Router) {}, func() {}
	}
	store, err := httpapi.NewPostgresAuthSessionStore(ctx, databaseURL)
	if err != nil {
		log.Fatalf("open PostgreSQL auth session store: %v", err)
	}
	log.Print("api-go using PostgreSQL-backed auth sessions")
	return httpapi.WithAuthSessionStore(store), func() {
		if err := store.Close(); err != nil {
			log.Printf("close PostgreSQL auth session store: %v", err)
		}
	}
}

func newRepository(ctx context.Context) (platform.Repository, func()) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Print("api-go using in-memory store; set DATABASE_URL to enable PostgreSQL persistence")
		store := platform.NewStore(platform.DefaultHomeModules())
		configureAuditLogIntegrity(store)
		configureObjectStorage(store)
		configurePhoneVerification(store)
		return store, func() {}
	}
	store, err := platform.NewPostgresStore(ctx, databaseURL, platform.DefaultHomeModules())
	if err != nil {
		log.Fatalf("open PostgreSQL store: %v", err)
	}
	configureAuditLogIntegrity(store)
	configureObjectStorage(store)
	configurePhoneVerification(store)
	log.Print("api-go using PostgreSQL-backed store")
	return store, func() {
		if err := store.Close(); err != nil {
			log.Printf("close PostgreSQL store: %v", err)
		}
	}
}

type phoneVerificationConfigurator interface {
	ConfigurePhoneVerification(platform.PhoneVerificationConfig) error
}

type phoneVerificationHTTPDispatcher struct {
	endpoint string
	token    string
	client   *http.Client
}

func (dispatcher phoneVerificationHTTPDispatcher) DispatchPhoneVerificationCode(req platform.PhoneVerificationDispatchRequest) (*platform.PhoneVerificationDispatchResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest(http.MethodPost, dispatcher.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(dispatcher.token) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(dispatcher.token))
	}
	client := dispatcher.client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sms provider returned status %d", resp.StatusCode)
	}
	result := &platform.PhoneVerificationDispatchResult{
		Provider:  req.Provider,
		RequestID: req.RequestID,
		Status:    "delivered",
		SentAt:    time.Now().UTC(),
	}
	var envelope struct {
		Provider  string `json:"provider"`
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err == nil {
		if strings.TrimSpace(envelope.Provider) != "" {
			result.Provider = strings.TrimSpace(envelope.Provider)
		}
		if strings.TrimSpace(envelope.RequestID) != "" {
			result.RequestID = strings.TrimSpace(envelope.RequestID)
		}
		if strings.TrimSpace(envelope.Status) != "" {
			result.Status = strings.TrimSpace(envelope.Status)
		}
	}
	return result, nil
}

func configurePhoneVerification(store phoneVerificationConfigurator) {
	mode := strings.ToLower(strings.TrimSpace(env("PHONE_VERIFICATION_MODE", "dev")))
	config := platform.PhoneVerificationConfig{
		Mode:            mode,
		Provider:        env("SMS_PROVIDER_NAME", "dev"),
		TemplateID:      os.Getenv("SMS_TEMPLATE_PHONE_CODE"),
		Cooldown:        time.Duration(envInt64("PHONE_CODE_COOLDOWN_SECONDS", 60)) * time.Second,
		ExpiresIn:       time.Duration(envInt64("PHONE_CODE_EXPIRES_SECONDS", 10*60)) * time.Second,
		MaxPerPhoneHour: int(envInt64("PHONE_CODE_MAX_PER_PHONE_HOUR", 5)),
		MaxPerPhoneDay:  int(envInt64("PHONE_CODE_MAX_PER_PHONE_DAY", 20)),
		ReturnDevCode:   envBool("PHONE_CODE_RETURN_DEV_CODE", mode != "provider"),
	}
	if mode == "provider" {
		endpoint := strings.TrimSpace(os.Getenv("SMS_PROVIDER_ENDPOINT"))
		if endpoint == "" {
			log.Fatal("PHONE_VERIFICATION_MODE=provider requires SMS_PROVIDER_ENDPOINT")
		}
		config.Provider = env("SMS_PROVIDER_NAME", "sms_provider")
		config.Dispatcher = phoneVerificationHTTPDispatcher{
			endpoint: endpoint,
			token:    os.Getenv("SMS_PROVIDER_TOKEN"),
			client:   &http.Client{Timeout: time.Duration(envInt64("SMS_PROVIDER_TIMEOUT_SECONDS", 5)) * time.Second},
		}
		log.Print("api-go phone verification uses configured SMS provider")
	} else {
		log.Print("api-go phone verification uses development code return mode; set PHONE_VERIFICATION_MODE=provider for production SMS dispatch")
	}
	if err := store.ConfigurePhoneVerification(config); err != nil {
		log.Fatalf("configure phone verification: %v", err)
	}
}

type auditLogIntegrityConfigurator interface {
	ConfigureAuditLogIntegrity(string)
}

func configureAuditLogIntegrity(store auditLogIntegrityConfigurator) {
	secret := os.Getenv("AUDIT_LOG_SIGNING_SECRET")
	store.ConfigureAuditLogIntegrity(secret)
	if strings.TrimSpace(secret) == "" {
		log.Print("api-go audit log integrity uses sha256 without a signing secret for local development; set AUDIT_LOG_SIGNING_SECRET for production HMAC sealing")
	}
}

type objectStorageConfigurator interface {
	ConfigureObjectStorage(platform.ObjectStorageConfig) error
}

func configureObjectStorage(store objectStorageConfigurator) {
	config := platform.ObjectStorageConfig{
		Provider:                        env("OBJECT_STORAGE_PROVIDER", platform.ObjectStorageProviderMinIO),
		Bucket:                          env("OBJECT_STORAGE_BUCKET", "infinitech-private"),
		UploadBaseURL:                   env("OBJECT_STORAGE_UPLOAD_BASE_URL", "https://object-storage.infinitech.local/upload"),
		PublicBaseURL:                   env("OBJECT_STORAGE_PUBLIC_BASE_URL", "https://cdn.infinitech.local"),
		HeadBaseURL:                     env("OBJECT_STORAGE_HEAD_BASE_URL", env("OBJECT_STORAGE_PUBLIC_BASE_URL", "https://cdn.infinitech.local")),
		AuditArchiveDownloadBaseURL:     env("AUDIT_ARCHIVE_DOWNLOAD_BASE_URL", "https://object-storage.infinitech.local/audit-archives"),
		SigningSecret:                   os.Getenv("OBJECT_STORAGE_SIGNING_SECRET"),
		CallbackSigningSecret:           os.Getenv("OBJECT_STORAGE_CALLBACK_SIGNING_SECRET"),
		TicketTTL:                       time.Duration(envInt64("OBJECT_STORAGE_TICKET_TTL_SECONDS", 15*60)) * time.Second,
		MaxUploadBytes:                  envInt64("OBJECT_STORAGE_MAX_UPLOAD_BYTES", platform.AfterSalesEvidenceMaxBytes),
		AuditArchiveMaxDownloadBytes:    envInt64("AUDIT_ARCHIVE_MAX_DOWNLOAD_BYTES", 100*1024*1024),
		RequireHeadVerification:         envBool("OBJECT_STORAGE_REQUIRE_HEAD_VERIFICATION", false),
		RequireUploadCallbackForConfirm: envBool("OBJECT_STORAGE_REQUIRE_UPLOAD_CALLBACK", false),
		RequireScanApprovalForConfirm:   envBool("OBJECT_STORAGE_REQUIRE_SCAN_APPROVAL", false),
		HeadTimeout:                     time.Duration(envInt64("OBJECT_STORAGE_HEAD_TIMEOUT_SECONDS", 3)) * time.Second,
		AuditArchiveDownloadTimeout:     time.Duration(envInt64("AUDIT_ARCHIVE_DOWNLOAD_TIMEOUT_SECONDS", 10)) * time.Second,
	}
	if err := store.ConfigureObjectStorage(config); err != nil {
		log.Fatalf("configure object storage: %v", err)
	}
	if strings.TrimSpace(os.Getenv("OBJECT_STORAGE_SIGNING_SECRET")) == "" {
		log.Print("api-go object storage signing secret is empty for local development; set OBJECT_STORAGE_SIGNING_SECRET for production")
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
