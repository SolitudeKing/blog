package service

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"solitude-blog/server/internal/storage"
)

func TestAssetMimeAllowedRejectsUnsanitizedSVG(t *testing.T) {
	t.Parallel()

	if assetMimeAllowed("image/svg+xml") {
		t.Fatal("assetMimeAllowed(image/svg+xml) = true, want false")
	}
	for _, mimeType := range []string{"image/jpeg", "image/png", "image/gif", "image/webp"} {
		if !assetMimeAllowed(mimeType) {
			t.Fatalf("assetMimeAllowed(%q) = false, want true", mimeType)
		}
	}
}

func TestAssetExtensionComesFromDetectedMIME(t *testing.T) {
	t.Parallel()

	wants := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	for mimeType, want := range wants {
		ext, ok := assetExtensionForMIME(mimeType)
		if !ok || ext != want {
			t.Fatalf("assetExtensionForMIME(%q) = %q, %v; want %q, true", mimeType, ext, ok, want)
		}
	}
	if ext, ok := assetExtensionForMIME("text/html"); ok || ext != "" {
		t.Fatalf("assetExtensionForMIME(text/html) = %q, %v; want empty, false", ext, ok)
	}
}

// TestUploadRoundTrip 验证文件上传后落盘字节完整。
// 修复前 sniff 后 reader 未重置导致前 512 字节被吞，本测试覆盖 1 KB / 100 KB / 1 MB 三档。
func TestUploadRoundTrip(t *testing.T) {
	t.Parallel()

	// PNG magic header 用于让 http.DetectContentType 识别为 image/png。
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	for _, size := range []int{1024, 100 * 1024, 1024 * 1024} {
		t.Run(fmt.Sprintf("%d-bytes", size), func(t *testing.T) {
			payload := make([]byte, size)
			copy(payload, pngMagic)
			// 用确定性的非零填充，避免任何隐式压缩或零字节优化掩盖问题。
			for i := len(pngMagic); i < size; i++ {
				payload[i] = byte((i*31 + 7) % 251)
			}

			tmpDir := t.TempDir()
			store, err := storage.NewLocal(tmpDir)
			if err != nil {
				t.Fatalf("storage.NewLocal: %v", err)
			}
			svc := NewAssetService(nil, store)
			fileHeader := buildMultipartFileHeader(t, payload, "test.png")

			item, err := svc.Upload(fileHeader, "test.png")
			if err != nil {
				t.Fatalf("svc.Upload: %v", err)
			}

			if item.Size != uint64(size) {
				t.Fatalf("item.Size = %d, want %d", item.Size, size)
			}

			fsPath := filepath.Join(tmpDir, item.StorageKey)
			saved, err := os.ReadFile(fsPath)
			if err != nil {
				t.Fatalf("os.ReadFile(%q): %v", fsPath, err)
			}
			if len(saved) != size {
				t.Fatalf("saved size = %d, want %d", len(saved), size)
			}
			if !bytes.Equal(saved, payload) {
				// 截断前若干字节定位首失配位置，便于排查。
				for i := 0; i < size; i++ {
					if saved[i] != payload[i] {
						t.Fatalf("first mismatch at byte %d: saved=%d want=%d", i, saved[i], payload[i])
					}
				}
			}
		})
	}
}

// buildMultipartFileHeader 模拟 Gin 解析 multipart/form-data 后得到的 FileHeader。
func buildMultipartFileHeader(t *testing.T, data []byte, filename string) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("multipart.CreateFormFile: %v", err)
	}
	if _, err := fileWriter.Write(data); err != nil {
		t.Fatalf("multipart write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart.Close: %v", err)
	}

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		t.Fatalf("multipart.ReadForm: %v", err)
	}
	files := form.File["file"]
	if len(files) == 0 {
		t.Fatal("multipart form has no file field")
	}
	return files[0]
}