package main

import (
	"context"
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
		configureObjectStorage(store)
		return store, func() {}
	}
	store, err := platform.NewPostgresStore(ctx, databaseURL, platform.DefaultHomeModules())
	if err != nil {
		log.Fatalf("open PostgreSQL store: %v", err)
	}
	configureObjectStorage(store)
	log.Print("api-go using PostgreSQL-backed store")
	return store, func() {
		if err := store.Close(); err != nil {
			log.Printf("close PostgreSQL store: %v", err)
		}
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
		SigningSecret:                   os.Getenv("OBJECT_STORAGE_SIGNING_SECRET"),
		CallbackSigningSecret:           os.Getenv("OBJECT_STORAGE_CALLBACK_SIGNING_SECRET"),
		TicketTTL:                       time.Duration(envInt64("OBJECT_STORAGE_TICKET_TTL_SECONDS", 15*60)) * time.Second,
		MaxUploadBytes:                  envInt64("OBJECT_STORAGE_MAX_UPLOAD_BYTES", platform.AfterSalesEvidenceMaxBytes),
		RequireHeadVerification:         envBool("OBJECT_STORAGE_REQUIRE_HEAD_VERIFICATION", false),
		RequireUploadCallbackForConfirm: envBool("OBJECT_STORAGE_REQUIRE_UPLOAD_CALLBACK", false),
		RequireScanApprovalForConfirm:   envBool("OBJECT_STORAGE_REQUIRE_SCAN_APPROVAL", false),
		HeadTimeout:                     time.Duration(envInt64("OBJECT_STORAGE_HEAD_TIMEOUT_SECONDS", 3)) * time.Second,
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
