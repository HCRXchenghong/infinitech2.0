package platform

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestObjectStorageConfigCreatesSignedUploadTicket(t *testing.T) {
	config, err := NormalizeObjectStorageConfig(ObjectStorageConfig{
		Provider:       ObjectStorageProviderMinIO,
		Bucket:         "after-sales",
		UploadBaseURL:  "https://minio.example.com/upload/",
		PublicBaseURL:  "https://cdn.example.com/assets/",
		SigningSecret:  "storage-secret",
		TicketTTL:      10 * time.Minute,
		MaxUploadBytes: AfterSalesEvidenceMaxBytes,
	})
	if err != nil {
		t.Fatal(err)
	}
	expiresAt := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	ticket, err := config.createObjectUploadTicket(objectUploadTicketInput{
		ObjectKey:    "after-sales/asr_1/sig/evidence.jpg",
		ContentType:  "image/jpeg",
		SizeBytes:    1024,
		MaxSizeBytes: AfterSalesEvidenceMaxBytes,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Provider != ObjectStorageProviderMinIO || ticket.Bucket != "after-sales" || ticket.Method != objectUploadMethod {
		t.Fatalf("expected minio upload ticket metadata, got %+v", ticket)
	}
	if !strings.HasPrefix(ticket.UploadURL, "https://minio.example.com/upload/after-sales/after-sales/asr_1/sig/evidence.jpg?") {
		t.Fatalf("expected upload URL to include endpoint, bucket and object key, got %s", ticket.UploadURL)
	}
	if ticket.PublicURL != "https://cdn.example.com/assets/after-sales/asr_1/sig/evidence.jpg" {
		t.Fatalf("expected public URL to use configured CDN base, got %s", ticket.PublicURL)
	}
	if ticket.Headers["X-Upload-Signature"] == "" || ticket.Headers["X-Object-Key"] != ticket.ObjectKey || !strings.Contains(ticket.UploadURL, "signature=") {
		t.Fatalf("expected signed upload URL and headers, got url=%s headers=%+v", ticket.UploadURL, ticket.Headers)
	}
}

func TestObjectStorageConfigRejectsUnsafeProductionBounds(t *testing.T) {
	if _, err := NormalizeObjectStorageConfig(ObjectStorageConfig{Provider: "s3"}); err != ErrInvalidArgument {
		t.Fatalf("expected unsupported provider to be rejected, got %v", err)
	}
	if _, err := NormalizeObjectStorageConfig(ObjectStorageConfig{TicketTTL: 2 * time.Hour}); err != ErrInvalidArgument {
		t.Fatalf("expected overly long upload ticket ttl to be rejected, got %v", err)
	}
	if _, err := NormalizeObjectStorageConfig(ObjectStorageConfig{MaxUploadBytes: AfterSalesEvidenceMaxBytes + 1}); err != ErrInvalidArgument {
		t.Fatalf("expected max upload size above evidence limit to be rejected, got %v", err)
	}
	if _, err := NormalizeObjectStorageConfig(ObjectStorageConfig{HeadTimeout: 31 * time.Second}); err != ErrInvalidArgument {
		t.Fatalf("expected overly long object head timeout to be rejected, got %v", err)
	}
	if _, err := NormalizeObjectStorageConfig(ObjectStorageConfig{RequireUploadCallbackForConfirm: true}); err != ErrInvalidArgument {
		t.Fatalf("expected callback gate without signing secret to be rejected, got %v", err)
	}
	config, err := NormalizeObjectStorageConfig(ObjectStorageConfig{SigningSecret: "upload-secret", RequireScanApprovalForConfirm: true})
	if err != nil {
		t.Fatal(err)
	}
	if !config.RequireUploadCallbackForConfirm || config.CallbackSigningSecret != "upload-secret" {
		t.Fatalf("expected scan gate to imply upload callback gate and fallback callback secret, got %+v", config)
	}
}

func TestObjectStorageConfigVerifiesUploadedObjectWithHead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodHead {
			t.Errorf("expected HEAD request, got %s", req.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if req.URL.Path != "/objects/after-sales/after-sales/asr_1/sig/evidence.jpg" {
			t.Errorf("unexpected object head path: %s", req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", "1024")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config, err := NormalizeObjectStorageConfig(ObjectStorageConfig{
		Provider:                ObjectStorageProviderMinIO,
		Bucket:                  "after-sales",
		PublicBaseURL:           "https://cdn.example.com/assets",
		HeadBaseURL:             server.URL + "/objects",
		RequireHeadVerification: true,
		HeadTimeout:             time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := config.verifyUploadedObject(objectHeadCheckInput{
		ObjectKey:   "after-sales/asr_1/sig/evidence.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   1024,
	}); err != nil {
		t.Fatalf("expected uploaded object HEAD verification to pass, got %v", err)
	}
	if err := config.verifyUploadedObject(objectHeadCheckInput{
		ObjectKey:   "after-sales/asr_1/sig/evidence.jpg",
		ContentType: "image/png",
		SizeBytes:   1024,
	}); err != ErrInvalidArgument {
		t.Fatalf("expected content type mismatch to be rejected, got %v", err)
	}
}
