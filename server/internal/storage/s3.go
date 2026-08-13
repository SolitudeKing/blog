package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config 是兼容 S3 协议的对象存储驱动所需的连接配置。
// 通常对应 MinIO 服务，也适用于任何实现 AWS SigV4 + S3 API 的对象存储。
type S3Config struct {
	Endpoint  string // 例如 https://minio.solitude.love，host[:port] 也接受
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
	// PublicURL 是对象的外链前缀，不含末尾 /。推荐形如
	// "https://minio.solitude.love/blog" 或 CDN 域名。
	// 留空时按 endpoint + bucket 拼接。
	PublicURL string
}

// S3Storage 通过 minio-go 操作 S3 兼容的对象存储。
// 启动时只做 BucketExists 探测，不自动创建 bucket（由运维手工预置）。
type S3Storage struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewS3(cfg S3Config) (*S3Storage, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, errors.New("s3 storage: endpoint is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, errors.New("s3 storage: bucket is required")
	}
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, errors.New("s3 storage: access key and secret key are required")
	}
	if strings.TrimSpace(cfg.Region) == "" {
		cfg.Region = "us-east-1"
	}

	host, scheme, err := parseEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	client, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: scheme == "https" || cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 storage: create client: %w", err)
	}

	publicURL := strings.TrimRight(cfg.PublicURL, "/")
	if publicURL == "" {
		publicURL = fmt.Sprintf("%s://%s/%s", scheme, host, cfg.Bucket)
	}

	return &S3Storage{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: publicURL,
	}, nil
}

// Probe 启动时调用一次：确认 bucket 存在且凭据可用。
// 由 caller 在 NewS3 之后立即调用，错误将阻止服务启动。
func (s *S3Storage) Probe(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("s3 storage: probe bucket %q: %w", s.bucket, err)
	}
	if !exists {
		return fmt.Errorf("s3 storage: bucket %q does not exist (auto-create disabled)", s.bucket)
	}
	return nil
}

// parseEndpoint 接受 "host[:port]" 或 "http(s)://host[:port]"，
// 返回 host（minio.New 要求）与 scheme。
func parseEndpoint(endpoint string) (string, string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if strings.Contains(endpoint, "://") {
		u, err := url.Parse(endpoint)
		if err != nil {
			return "", "", fmt.Errorf("s3 storage: invalid endpoint %q: %w", endpoint, err)
		}
		if u.Host == "" {
			return "", "", errors.New("s3 storage: endpoint must include host")
		}
		scheme := strings.ToLower(u.Scheme)
		if scheme != "http" && scheme != "https" {
			return "", "", fmt.Errorf("s3 storage: endpoint scheme must be http or https, got %q", scheme)
		}
		return u.Host, scheme, nil
	}
	return endpoint, "http", nil
}

func (s *S3Storage) Kind() string {
	return "s3"
}

func (s *S3Storage) PublicURL(key string) string {
	return s.publicURL + "/" + strings.TrimLeft(key, "/")
}

func (s *S3Storage) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	if reader == nil {
		return "", errors.New("s3 storage Put: reader is nil")
	}
	info := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if _, err := s.client.PutObject(ctx, s.bucket, key, reader, size, info); err != nil {
		return "", fmt.Errorf("s3 storage: put object %q: %w", key, err)
	}
	return s.PublicURL(key), nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		// 兼容"对象不存在"语义：忽略 ErrObjectNotFound。
		errStr := err.Error()
		if strings.Contains(errStr, "NoSuchKey") || strings.Contains(errStr, "Not Found") {
			return nil
		}
		return fmt.Errorf("s3 storage: remove object %q: %w", key, err)
	}
	return nil
}