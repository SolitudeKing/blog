package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStoragePutCreatesFileAndPublicURL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store, err := NewLocal(root)
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}

	ctx := context.Background()
	payload := []byte("hello-storage")
	url, err := store.Put(ctx, "2026/08/sample.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain")
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if url != "/uploads/2026/08/sample.txt" {
		t.Fatalf("PublicURL = %q, want /uploads/2026/08/sample.txt", url)
	}

	got, err := readFile(t, filepath.Join(root, "2026", "08", "sample.txt"))
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("uploaded content = %q, want %q", got, payload)
	}

	if err := store.Delete(ctx, "2026/08/sample.txt"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := readFile(t, filepath.Join(root, "2026", "08", "sample.txt")); err == nil {
		t.Fatal("expected file to be removed after Delete()")
	}

	// 删除不存在的对象应当容忍而非报错。
	if err := store.Delete(ctx, "2026/08/missing.txt"); err != nil {
		t.Fatalf("Delete(missing) error = %v, want nil", err)
	}

	if store.Kind() != "local" {
		t.Fatalf("Kind() = %q, want local", store.Kind())
	}
}

func TestLocalStorageRejectsInvalidRoot(t *testing.T) {
	t.Parallel()

	if _, err := NewLocal(""); err == nil {
		t.Fatal("NewLocal(\"\") error = nil, want validation error")
	}
}

func TestLocalStoragePublicURLNormalizesLeadingSlash(t *testing.T) {
	t.Parallel()

	store, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	if got := store.PublicURL("/foo/bar.jpg"); got != "/uploads/foo/bar.jpg" {
		t.Fatalf("PublicURL(/foo/bar.jpg) = %q, want /uploads/foo/bar.jpg", got)
	}
}

func TestParseEndpointAcceptsURLAndHost(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input        string
		wantHost     string
		wantScheme   string
		wantErrorSub string
	}{
		{input: "minio.example.com:9000", wantHost: "minio.example.com:9000", wantScheme: "http"},
		{input: "https://minio.solitude.love", wantHost: "minio.solitude.love", wantScheme: "https"},
		{input: "http://127.0.0.1:9000", wantHost: "127.0.0.1:9000", wantScheme: "http"},
		{input: "ftp://bad", wantErrorSub: "scheme must be http or https"},
		{input: "https://", wantErrorSub: "endpoint must include host"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			host, scheme, err := parseEndpoint(tc.input)
			if tc.wantErrorSub != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrorSub) {
					t.Fatalf("parseEndpoint(%q) error = %v, want substring %q", tc.input, err, tc.wantErrorSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseEndpoint(%q) error = %v", tc.input, err)
			}
			if host != tc.wantHost {
				t.Fatalf("host = %q, want %q", host, tc.wantHost)
			}
			if scheme != tc.wantScheme {
				t.Fatalf("scheme = %q, want %q", scheme, tc.wantScheme)
			}
		})
	}
}

func readFile(t *testing.T, path string) ([]byte, error) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}