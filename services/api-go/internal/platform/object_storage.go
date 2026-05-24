package platform

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ObjectStorageProviderMinIO = "minio"
	objectUploadMethod         = "PUT"
)

type ObjectStorageConfig struct {
	Provider                        string
	Bucket                          string
	UploadBaseURL                   string
	PublicBaseURL                   string
	HeadBaseURL                     string
	AuditArchiveDownloadBaseURL     string
	SigningSecret                   string
	CallbackSigningSecret           string
	TicketTTL                       time.Duration
	MaxUploadBytes                  int64
	AuditArchiveMaxDownloadBytes    int64
	RequireHeadVerification         bool
	RequireUploadCallbackForConfirm bool
	RequireScanApprovalForConfirm   bool
	HeadTimeout                     time.Duration
	AuditArchiveDownloadTimeout     time.Duration
}

type objectUploadTicketInput struct {
	ObjectKey    string
	ContentType  string
	SizeBytes    int64
	MaxSizeBytes int64
	ExpiresAt    time.Time
}

type objectHeadCheckInput struct {
	ObjectKey   string
	ContentType string
	SizeBytes   int64
}

type objectUploadCallbackSignatureInput struct {
	TicketID    string
	ObjectKey   string
	ContentType string
	SizeBytes   int64
	ContentSHA  string
	UploadedAt  time.Time
}

type objectScanResultSignatureInput struct {
	TicketID      string
	ObjectKey     string
	ScanStatus    string
	ScanResult    string
	Scanner       string
	ScanCheckedAt time.Time
}

func DefaultObjectStorageConfig() ObjectStorageConfig {
	return ObjectStorageConfig{
		Provider:                     ObjectStorageProviderMinIO,
		Bucket:                       "infinitech-private",
		UploadBaseURL:                "https://object-storage.infinitech.local/upload",
		PublicBaseURL:                "https://cdn.infinitech.local",
		HeadBaseURL:                  "https://cdn.infinitech.local",
		TicketTTL:                    15 * time.Minute,
		MaxUploadBytes:               AfterSalesEvidenceMaxBytes,
		HeadTimeout:                  3 * time.Second,
		AuditArchiveDownloadBaseURL:  "https://object-storage.infinitech.local/audit-archives",
		AuditArchiveMaxDownloadBytes: 100 * 1024 * 1024,
		AuditArchiveDownloadTimeout:  10 * time.Second,
	}
}

func NormalizeObjectStorageConfig(config ObjectStorageConfig) (ObjectStorageConfig, error) {
	defaults := DefaultObjectStorageConfig()
	config.Provider = strings.ToLower(strings.TrimSpace(config.Provider))
	if config.Provider == "" {
		config.Provider = defaults.Provider
	}
	if config.Provider != ObjectStorageProviderMinIO {
		return ObjectStorageConfig{}, ErrInvalidArgument
	}
	config.Bucket = strings.Trim(strings.TrimSpace(config.Bucket), "/")
	if config.Bucket == "" {
		config.Bucket = defaults.Bucket
	}
	config.UploadBaseURL = strings.TrimRight(strings.TrimSpace(config.UploadBaseURL), "/")
	if config.UploadBaseURL == "" {
		config.UploadBaseURL = defaults.UploadBaseURL
	}
	config.PublicBaseURL = strings.TrimRight(strings.TrimSpace(config.PublicBaseURL), "/")
	if config.PublicBaseURL == "" {
		config.PublicBaseURL = defaults.PublicBaseURL
	}
	config.HeadBaseURL = strings.TrimRight(strings.TrimSpace(config.HeadBaseURL), "/")
	if config.HeadBaseURL == "" {
		config.HeadBaseURL = config.PublicBaseURL
	}
	config.AuditArchiveDownloadBaseURL = strings.TrimRight(strings.TrimSpace(config.AuditArchiveDownloadBaseURL), "/")
	if config.AuditArchiveDownloadBaseURL == "" {
		config.AuditArchiveDownloadBaseURL = defaults.AuditArchiveDownloadBaseURL
	}
	config.SigningSecret = strings.TrimSpace(config.SigningSecret)
	config.CallbackSigningSecret = strings.TrimSpace(config.CallbackSigningSecret)
	if config.CallbackSigningSecret == "" {
		config.CallbackSigningSecret = config.SigningSecret
	}
	if config.RequireScanApprovalForConfirm {
		config.RequireUploadCallbackForConfirm = true
	}
	if (config.RequireUploadCallbackForConfirm || config.RequireScanApprovalForConfirm) && config.CallbackSigningSecret == "" {
		return ObjectStorageConfig{}, ErrInvalidArgument
	}
	if config.TicketTTL <= 0 {
		config.TicketTTL = defaults.TicketTTL
	}
	if config.TicketTTL > time.Hour {
		return ObjectStorageConfig{}, ErrInvalidArgument
	}
	if config.MaxUploadBytes <= 0 {
		config.MaxUploadBytes = defaults.MaxUploadBytes
	}
	if config.MaxUploadBytes > AfterSalesEvidenceMaxBytes {
		return ObjectStorageConfig{}, ErrInvalidArgument
	}
	if config.AuditArchiveMaxDownloadBytes <= 0 {
		config.AuditArchiveMaxDownloadBytes = defaults.AuditArchiveMaxDownloadBytes
	}
	if config.AuditArchiveMaxDownloadBytes > 1024*1024*1024 {
		return ObjectStorageConfig{}, ErrInvalidArgument
	}
	if config.HeadTimeout <= 0 {
		config.HeadTimeout = defaults.HeadTimeout
	}
	if config.HeadTimeout > 30*time.Second {
		return ObjectStorageConfig{}, ErrInvalidArgument
	}
	if config.AuditArchiveDownloadTimeout <= 0 {
		config.AuditArchiveDownloadTimeout = defaults.AuditArchiveDownloadTimeout
	}
	if config.AuditArchiveDownloadTimeout > 120*time.Second {
		return ObjectStorageConfig{}, ErrInvalidArgument
	}
	return config, nil
}

func (s *Store) ConfigureObjectStorage(config ObjectStorageConfig) error {
	normalized, err := NormalizeObjectStorageConfig(config)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objectStorage = normalized
	return nil
}

func (s *PostgresStore) ConfigureObjectStorage(config ObjectStorageConfig) error {
	return s.Store.ConfigureObjectStorage(config)
}

func (s *Store) objectStorageConfigLocked() ObjectStorageConfig {
	config, err := NormalizeObjectStorageConfig(s.objectStorage)
	if err != nil {
		return DefaultObjectStorageConfig()
	}
	return config
}

func (s *Store) objectStorageSnapshot() ObjectStorageConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.objectStorageConfigLocked()
}

func (config ObjectStorageConfig) createObjectUploadTicket(input objectUploadTicketInput) (*ObjectUploadTicket, error) {
	config, err := NormalizeObjectStorageConfig(config)
	if err != nil {
		return nil, err
	}
	input.ObjectKey = strings.TrimSpace(input.ObjectKey)
	input.ContentType = strings.TrimSpace(input.ContentType)
	if input.ObjectKey == "" || input.ContentType == "" || input.SizeBytes <= 0 || input.SizeBytes > config.MaxUploadBytes {
		return nil, ErrInvalidArgument
	}
	if input.MaxSizeBytes <= 0 || input.MaxSizeBytes > config.MaxUploadBytes {
		input.MaxSizeBytes = config.MaxUploadBytes
	}
	if input.ExpiresAt.IsZero() {
		input.ExpiresAt = time.Now().UTC().Add(config.TicketTTL)
	} else {
		input.ExpiresAt = input.ExpiresAt.UTC()
	}
	signature := config.signObjectUpload(input)
	query := url.Values{}
	query.Set("bucket", config.Bucket)
	query.Set("content_type", input.ContentType)
	query.Set("expires", strconv.FormatInt(input.ExpiresAt.Unix(), 10))
	query.Set("max_size_bytes", strconv.FormatInt(input.MaxSizeBytes, 10))
	query.Set("signature", signature)
	return &ObjectUploadTicket{
		Provider:  config.Provider,
		Bucket:    config.Bucket,
		ObjectKey: input.ObjectKey,
		UploadURL: fmt.Sprintf(
			"%s?%s",
			joinObjectURL(config.UploadBaseURL, config.Bucket+"/"+input.ObjectKey),
			query.Encode(),
		),
		PublicURL: config.publicObjectURL(input.ObjectKey),
		Method:    objectUploadMethod,
		Headers: map[string]string{
			"Content-Type":       input.ContentType,
			"X-Content-SHA":      "required",
			"X-Max-Size-Bytes":   strconv.FormatInt(input.MaxSizeBytes, 10),
			"X-Object-Bucket":    config.Bucket,
			"X-Object-Key":       input.ObjectKey,
			"X-Upload-Signature": signature,
		},
		ExpiresAt:    input.ExpiresAt,
		MaxSizeBytes: input.MaxSizeBytes,
	}, nil
}

func (config ObjectStorageConfig) signObjectUpload(input objectUploadTicketInput) string {
	payload := strings.Join([]string{
		config.Provider,
		config.Bucket,
		input.ObjectKey,
		input.ContentType,
		strconv.FormatInt(input.SizeBytes, 10),
		strconv.FormatInt(input.MaxSizeBytes, 10),
		strconv.FormatInt(input.ExpiresAt.Unix(), 10),
	}, "\n")
	mac := hmac.New(sha256.New, []byte(config.SigningSecret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (config ObjectStorageConfig) signObjectUploadCallback(input objectUploadCallbackSignatureInput) string {
	input.ContentType = normalizeEvidenceContentType(input.ContentType)
	input.ContentSHA = strings.TrimSpace(input.ContentSHA)
	input.UploadedAt = input.UploadedAt.UTC()
	payload := strings.Join([]string{
		strings.TrimSpace(input.TicketID),
		strings.TrimSpace(input.ObjectKey),
		input.ContentType,
		strconv.FormatInt(input.SizeBytes, 10),
		input.ContentSHA,
		strconv.FormatInt(input.UploadedAt.Unix(), 10),
	}, "\n")
	mac := hmac.New(sha256.New, []byte(config.CallbackSigningSecret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (config ObjectStorageConfig) verifyObjectUploadCallback(input objectUploadCallbackSignatureInput, signature string) error {
	config, err := NormalizeObjectStorageConfig(config)
	if err != nil {
		return err
	}
	if config.CallbackSigningSecret == "" {
		return nil
	}
	expected := config.signObjectUploadCallback(input)
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature))) {
		return ErrInvalidArgument
	}
	return nil
}

func (config ObjectStorageConfig) signObjectScanResult(input objectScanResultSignatureInput) string {
	input.ScanStatus = normalizeAfterSalesUploadScanStatus(input.ScanStatus)
	input.ScanResult = strings.TrimSpace(input.ScanResult)
	input.Scanner = strings.TrimSpace(input.Scanner)
	input.ScanCheckedAt = input.ScanCheckedAt.UTC()
	payload := strings.Join([]string{
		strings.TrimSpace(input.TicketID),
		strings.TrimSpace(input.ObjectKey),
		input.ScanStatus,
		input.ScanResult,
		input.Scanner,
		strconv.FormatInt(input.ScanCheckedAt.Unix(), 10),
	}, "\n")
	mac := hmac.New(sha256.New, []byte(config.CallbackSigningSecret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (config ObjectStorageConfig) verifyObjectScanResult(input objectScanResultSignatureInput, signature string) error {
	config, err := NormalizeObjectStorageConfig(config)
	if err != nil {
		return err
	}
	if config.CallbackSigningSecret == "" {
		return nil
	}
	expected := config.signObjectScanResult(input)
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature))) {
		return ErrInvalidArgument
	}
	return nil
}

func (config ObjectStorageConfig) publicObjectURL(objectKey string) string {
	return joinObjectURL(config.PublicBaseURL, objectKey)
}

func (config ObjectStorageConfig) verifyUploadedObject(input objectHeadCheckInput) error {
	config, err := NormalizeObjectStorageConfig(config)
	if err != nil {
		return err
	}
	if !config.RequireHeadVerification {
		return nil
	}
	input.ObjectKey = strings.TrimSpace(input.ObjectKey)
	input.ContentType = normalizeEvidenceContentType(input.ContentType)
	if input.ObjectKey == "" || input.ContentType == "" || input.SizeBytes <= 0 || input.SizeBytes > config.MaxUploadBytes {
		return ErrInvalidArgument
	}
	request, err := http.NewRequest(http.MethodHead, joinObjectURL(config.HeadBaseURL, config.Bucket+"/"+input.ObjectKey), nil)
	if err != nil {
		return ErrInvalidArgument
	}
	response, err := (&http.Client{Timeout: config.HeadTimeout}).Do(request)
	if err != nil {
		return ErrInvalidOrderState
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return ErrInvalidOrderState
	}
	if response.ContentLength >= 0 && response.ContentLength != input.SizeBytes {
		return ErrInvalidArgument
	}
	headContentType := normalizeEvidenceContentType(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	if headContentType != "" && headContentType != input.ContentType {
		return ErrInvalidArgument
	}
	return nil
}

func (config ObjectStorageConfig) downloadAuditArchiveObject(storageKey string, expectedBytes int64) ([]byte, error) {
	config, err := NormalizeObjectStorageConfig(config)
	if err != nil {
		return nil, err
	}
	path := normalizeAuditArchiveStoragePath(storageKey)
	if path == "" {
		return nil, ErrInvalidArgument
	}
	request, err := http.NewRequest(http.MethodGet, joinObjectURL(config.AuditArchiveDownloadBaseURL, path), nil)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	response, err := (&http.Client{Timeout: config.AuditArchiveDownloadTimeout}).Do(request)
	if err != nil {
		return nil, ErrInvalidOrderState
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, ErrInvalidOrderState
	}
	maxBytes := config.AuditArchiveMaxDownloadBytes
	if expectedBytes > 0 && expectedBytes < maxBytes {
		maxBytes = expectedBytes + 1
	}
	if response.ContentLength > maxBytes {
		return nil, ErrInvalidArgument
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return nil, ErrInvalidOrderState
	}
	if int64(len(body)) > maxBytes {
		return nil, ErrInvalidArgument
	}
	return body, nil
}

func normalizeAuditArchiveStoragePath(value string) string {
	raw := strings.TrimSpace(value)
	raw = strings.TrimPrefix(raw, "worm://")
	raw = strings.TrimPrefix(raw, "s3://")
	raw = strings.TrimPrefix(raw, "minio://")
	parts := strings.Split(raw, "/")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		normalized = append(normalized, url.PathEscape(part))
	}
	return strings.Join(normalized, "/")
}

func joinObjectURL(baseURL string, path string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	path = strings.TrimLeft(strings.TrimSpace(path), "/")
	if baseURL == "" {
		return path
	}
	if path == "" {
		return baseURL
	}
	return baseURL + "/" + path
}
