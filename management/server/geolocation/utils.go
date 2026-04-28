package geolocation

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// decompressTarGzFile decompresses the expected file from a .tar.gz archive.
func decompressTarGzFile(filepath, destDir, expectedName string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	extracted := false

	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if header.Typeflag == tar.TypeReg && archiveEntryMatchesExpected(header.Name, expectedName) {
			outPath, err := archiveOutputPath(destDir, expectedName)
			if err != nil {
				return err
			}
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, tarReader) // #nosec G110
			outFile.Close()
			if err != nil {
				return err
			}
			extracted = true
		}

	}

	if !extracted {
		return fmt.Errorf("archive does not contain expected file %q", expectedName)
	}

	return nil
}

// decompressZipFile decompresses the expected file from a .zip archive.
func decompressZipFile(filepath, destDir, expectedName string) error {
	r, err := zip.OpenReader(filepath)
	if err != nil {
		return err
	}
	defer r.Close()
	extracted := false

	for _, f := range r.File {
		if f.FileInfo().IsDir() || !archiveEntryMatchesExpected(f.Name, expectedName) {
			continue
		}

		outPath, err := archiveOutputPath(destDir, expectedName)
		if err != nil {
			return err
		}
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc) // #nosec G110
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
		extracted = true
	}

	if !extracted {
		return fmt.Errorf("archive does not contain expected file %q", expectedName)
	}

	return nil
}

func archiveOutputPath(destDir, expectedName string) (string, error) {
	if expectedName == "" || expectedName == "." || expectedName != filepath.Base(expectedName) {
		return "", fmt.Errorf("invalid expected archive filename %q", expectedName)
	}
	if strings.ContainsAny(expectedName, `/\`) {
		return "", fmt.Errorf("invalid expected archive filename %q", expectedName)
	}

	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return "", err
	}
	outAbs, err := filepath.Abs(filepath.Join(destAbs, expectedName))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(destAbs, outAbs)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("archive output path escapes destination: %q", outAbs)
	}
	return outAbs, nil
}

func archiveEntryMatchesExpected(archiveName, expectedName string) bool {
	if archiveName == "" {
		return false
	}
	if strings.Contains(archiveName, `\`) {
		return false
	}
	cleanName := filepath.Clean(archiveName)
	if cleanName == "." || cleanName == ".." || filepath.IsAbs(cleanName) || strings.HasPrefix(cleanName, "../") {
		return false
	}
	return filepath.Base(cleanName) == expectedName
}

// calculateFileSHA256 calculates the SHA256 checksum of a file.
func calculateFileSHA256(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// loadChecksumFromFile loads the first checksum from a file.
func loadChecksumFromFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) > 0 {
			return parts[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

// verifyChecksum compares the calculated SHA256 checksum of a file against the expected checksum.
func verifyChecksum(filepath, expectedChecksum string) error {
	calculatedChecksum, err := calculateFileSHA256(filepath)

	fileCheckSum := fmt.Sprintf("%x", calculatedChecksum)
	if err != nil {
		return err
	}

	if fileCheckSum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, fileCheckSum)
	}

	return nil
}

// downloadFile downloads a file from a URL and saves it to a local file path.
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected error occurred while downloading the file: %s", string(bodyBytes))
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewBuffer(bodyBytes))
	return err
}

func getFilenameFromURL(url string) (string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	_, params, err := mime.ParseMediaType(resp.Header["Content-Disposition"][0])
	if err != nil {
		return "", err
	}

	filename := params["filename"]

	return filename, nil
}
