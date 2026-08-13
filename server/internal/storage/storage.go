// Package storage 定义对象存储抽象。当前提供两种实现：本地文件系统（local）
// 与兼容 S3 协议的对象存储（s3，例如 MinIO）。上层业务只依赖 ObjectStorage
// 接口，由 factory 按 STORAGE_DRIVER 注入具体实现。
package storage

import (
	"context"
	"io"
)

// ObjectStorage 把上传/删除/URL 生成的差异封装在 driver 内部。
//
// Put 接收 io.Reader 与 size（-1 表示流式），由 driver 自行决定是否校验长度。
// 返回的字符串是可直接外链的 URL，本地驱动返回 "/uploads/<key>"，
// S3 驱动返回 "<STORAGE_S3_PUBLIC_URL>/<key>"。
type ObjectStorage interface {
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
	// Kind 用于日志与健康检查，例如 "local" 或 "s3"。
	Kind() string
}