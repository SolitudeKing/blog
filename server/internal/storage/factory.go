package storage

import (
	"context"
	"fmt"
	"strings"

	"solitude-blog/server/internal/config"
)

// NewFromConfig 根据 cfg.StorageDriver 选择 driver 并返回实例。
// s3 驱动会立即做一次 BucketExists 探测；探测失败将阻止服务启动。
//
// 注意：driver 合法性也由 config.Validate() 提前校验，本函数对未知 driver
// 仍返回错误，作为防御性兜底。
func NewFromConfig(ctx context.Context, cfg config.Config) (ObjectStorage, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.StorageDriver)) {
	case "", "local":
		return NewLocal(cfg.StorageLocalRoot)
	case "s3":
		store, err := NewS3(S3Config{
			Endpoint:  cfg.StorageS3Endpoint,
			AccessKey: cfg.StorageS3AccessKey,
			SecretKey: cfg.StorageS3SecretKey,
			Bucket:    cfg.StorageS3Bucket,
			Region:    cfg.StorageS3Region,
			UseSSL:    cfg.StorageS3UseSSL,
			PublicURL: cfg.StorageS3PublicURL,
		})
		if err != nil {
			return nil, err
		}
		if err := store.Probe(ctx); err != nil {
			return nil, err
		}
		return store, nil
	default:
		return nil, fmt.Errorf("storage: unsupported driver %q", cfg.StorageDriver)
	}
}