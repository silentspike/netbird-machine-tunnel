package geolocation

import (
	"path/filepath"
	"testing"
)

func TestArchiveOutputPathUsesBasename(t *testing.T) {
	destDir := t.TempDir()
	outPath, err := archiveOutputPath(destDir, "../nested/GeoLite2-City.mmdb")
	if err != nil {
		t.Fatalf("archiveOutputPath returned error: %v", err)
	}

	expected := filepath.Join(destDir, "GeoLite2-City.mmdb")
	if outPath != expected {
		t.Fatalf("expected %q, got %q", expected, outPath)
	}
}

func TestArchiveOutputPathRejectsEmptyBasename(t *testing.T) {
	if _, err := archiveOutputPath(t.TempDir(), "."); err == nil {
		t.Fatal("expected empty archive basename to be rejected")
	}
}
