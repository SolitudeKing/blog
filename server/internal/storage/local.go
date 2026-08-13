package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage 把上传对象落到本地磁盘，URL 通过 "/uploads/<key>" 反向代理暴露。
// 与原 service/asset_files.go 相比，写入采用「临时文件 + os.Rename」以保证
// 上传过程出现异常时不会留下半截文件，调用方无需再做 defer 删除。
type LocalStorage struct {
	rootDir string
}

func NewLocal(rootDir string) (*LocalStorage, error) {
	if strings.TrimSpace(rootDir) == "" {
		return nil, errors.New("local storage root directory is required")
	}
	cleaned := filepath.Clean(rootDir)
	if err := os.MkdirAll(cleaned, 0o755); err != nil {
		return nil, err
	}
	return &LocalStorage{rootDir: cleaned}, nil
}

func (l *LocalStorage) Kind() string {
	return "local"
}

func (l *LocalStorage) PublicURL(key string) string {
	return "/uploads/" + strings.TrimLeft(key, "/")
}

func (l *LocalStorage) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	if reader == nil {
		return "", errors.New("local storage Put: reader is nil")
	}
	fullPath := filepath.Join(l.rootDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}

	// 原子写入：先写到同目录临时文件，成功后 rename 到目标位置，
	// 避免进程被中断或磁盘满时出现 0 字节文件被外链直接读取。
	tmp, err := os.CreateTemp(filepath.Dir(fullPath), ".upload-*.tmp")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	written, err := io.Copy(tmp, reader)
	if err != nil {
		_ = tmp.Close()
		return "", err
	}
	if size >= 0 && written != size {
		_ = tmp.Close()
		return "", errors.New("local storage Put: written size mismatch")
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, fullPath); err != nil {
		return "", err
	}
	cleanup = false
	return l.PublicURL(key), nil
}

func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	fullPath := filepath.Join(l.rootDir, filepath.FromSlash(key))
	if err := os.Remove(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}