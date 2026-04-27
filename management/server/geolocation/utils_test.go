package geolocation

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveOutputPathUsesExpectedFilename(t *testing.T) {
	destDir := t.TempDir()
	outPath, err := archiveOutputPath(destDir, "GeoLite2-City.mmdb")
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

func TestArchiveOutputPathRejectsTraversalFilename(t *testing.T) {
	if _, err := archiveOutputPath(t.TempDir(), "../GeoLite2-City.mmdb"); err == nil {
		t.Fatal("expected traversal filename to be rejected")
	}
}

func TestArchiveEntryMatchesExpectedNestedPath(t *testing.T) {
	if !archiveEntryMatchesExpected("GeoLite2-City_20260427/GeoLite2-City.mmdb", "GeoLite2-City.mmdb") {
		t.Fatal("expected nested archive path to match by basename")
	}
	if archiveEntryMatchesExpected("GeoLite2-City_20260427/other.mmdb", "GeoLite2-City.mmdb") {
		t.Fatal("expected non-matching basename to be skipped")
	}
	if archiveEntryMatchesExpected("../GeoLite2-City.mmdb", "GeoLite2-City.mmdb") {
		t.Fatal("expected traversal archive path to be skipped")
	}
}

func TestDecompressTarGzFileExtractsOnlyExpectedFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "geolite.tar.gz")
	if err := writeTarGz(src, map[string]string{
		"GeoLite2-City_20260427/GeoLite2-City.mmdb": "expected",
		"GeoLite2-City_20260427/ignored.mmdb":       "ignored",
		"../GeoLite2-City.mmdb":                     "traversal",
	}); err != nil {
		t.Fatalf("write tar.gz fixture: %v", err)
	}

	destDir := t.TempDir()
	if err := decompressTarGzFile(src, destDir, "GeoLite2-City.mmdb"); err != nil {
		t.Fatalf("decompressTarGzFile returned error: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "GeoLite2-City.mmdb"), "expected")
	if _, err := os.Stat(filepath.Join(destDir, "ignored.mmdb")); err == nil {
		t.Fatal("unexpected ignored file extracted")
	}
}

func TestDecompressZipFileExtractsOnlyExpectedFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "geolite.zip")
	if err := writeZip(src, map[string]string{
		"GeoLite2-City-CSV_20260427/GeoLite2-City-Locations-en.csv": "expected",
		"GeoLite2-City-CSV_20260427/ignored.csv":                    "ignored",
		"../GeoLite2-City-Locations-en.csv":                         "traversal",
	}); err != nil {
		t.Fatalf("write zip fixture: %v", err)
	}

	destDir := t.TempDir()
	if err := decompressZipFile(src, destDir, "GeoLite2-City-Locations-en.csv"); err != nil {
		t.Fatalf("decompressZipFile returned error: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "GeoLite2-City-Locations-en.csv"), "expected")
	if _, err := os.Stat(filepath.Join(destDir, "ignored.csv")); err == nil {
		t.Fatal("unexpected ignored file extracted")
	}
}

func writeTarGz(path string, files map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for name, content := range files {
		if err := tarWriter.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0o600,
			Size: int64(len(content)),
		}); err != nil {
			return err
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return err
		}
	}
	return nil
}

func writeZip(path string, files map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()
	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return err
		}
	}
	return nil
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(got) != want {
		t.Fatalf("extracted content = %q, want %q", string(got), want)
	}
}
