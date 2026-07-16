package service

import "testing"

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
