package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	manifestArchivePath  = "manifest.json"
	checksumsArchivePath = "checksums.sha256"
	maxManifestBytes     = 1 << 20
	maxArchiveEntries    = 1_000_000
)

func inspectArchive(ctx context.Context, archivePath string, maxArchiveBytes, maxExtractedBytes int64) (*Manifest, error) {
	return readAndValidateArchive(ctx, archivePath, "", maxArchiveBytes, maxExtractedBytes)
}

func extractArchive(ctx context.Context, archivePath, destination string, maxArchiveBytes, maxExtractedBytes int64) (*Manifest, error) {
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return nil, err
	}
	return readAndValidateArchive(ctx, archivePath, destination, maxArchiveBytes, maxExtractedBytes)
}

func readAndValidateArchive(ctx context.Context, archivePath, destination string, maxArchiveBytes, maxExtractedBytes int64) (*Manifest, error) {
	info, err := os.Stat(archivePath)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("backup archive must be a regular file")
	}
	if maxArchiveBytes > 0 && info.Size() > maxArchiveBytes {
		return nil, fmt.Errorf("backup archive exceeds %d bytes", maxArchiveBytes)
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("open gzip stream: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := make(map[string]struct{})
	verified := make(map[string]struct{})
	var manifest *Manifest
	var totalBytes int64
	var uploadCount int64
	for entryCount := 0; ; entryCount++ {
		if entryCount >= maxArchiveEntries {
			return nil, errors.New("backup archive has too many entries")
		}
		header, nextErr := tr.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return nil, fmt.Errorf("read tar stream: %w", nextErr)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name, err := validateRestoreArchivePath(header.Name)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("duplicate archive entry: %s", name)
		}
		seen[name] = struct{}{}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			return nil, fmt.Errorf("unsupported archive entry type for %s", name)
		}
		if header.Size < 0 {
			return nil, fmt.Errorf("invalid archive entry size for %s", name)
		}
		if header.Size > int64(^uint64(0)>>1)-totalBytes {
			return nil, errors.New("backup archive size overflows supported range")
		}
		totalBytes += header.Size
		if maxExtractedBytes > 0 && totalBytes > maxExtractedBytes {
			return nil, fmt.Errorf("extracted backup exceeds %d bytes", maxExtractedBytes)
		}

		if name == manifestArchivePath {
			if manifest != nil || len(seen) != 1 {
				return nil, errors.New("manifest.json must be the first archive entry")
			}
			if header.Size > maxManifestBytes {
				return nil, errors.New("backup manifest is too large")
			}
			raw, readErr := io.ReadAll(io.LimitReader(tr, maxManifestBytes+1))
			if readErr != nil {
				return nil, readErr
			}
			var parsed Manifest
			if err := json.Unmarshal(raw, &parsed); err != nil {
				return nil, fmt.Errorf("parse backup manifest: %w", err)
			}
			if err := validateManifest(&parsed); err != nil {
				return nil, err
			}
			manifest = &parsed
			continue
		}
		if manifest == nil {
			return nil, errors.New("backup manifest is missing")
		}
		if name == checksumsArchivePath {
			if _, err := io.Copy(io.Discard, &contextReader{ctx: ctx, reader: tr}); err != nil {
				return nil, err
			}
			continue
		}
		expected, ok := manifest.Checksums[name]
		if !ok {
			return nil, fmt.Errorf("archive entry is not declared in manifest: %s", name)
		}
		var destinationFile *os.File
		if destination != "" {
			diskPath := filepath.Join(destination, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(diskPath), 0o700); err != nil {
				return nil, err
			}
			destinationFile, err = os.OpenFile(diskPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
			if err != nil {
				return nil, err
			}
		}
		h := sha256.New()
		writer := io.Writer(h)
		if destinationFile != nil {
			writer = io.MultiWriter(destinationFile, h)
		}
		_, copyErr := io.Copy(writer, &contextReader{ctx: ctx, reader: tr})
		if destinationFile != nil {
			closeErr := destinationFile.Close()
			if copyErr == nil {
				copyErr = closeErr
			}
		}
		if copyErr != nil {
			return nil, copyErr
		}
		actual := hex.EncodeToString(h.Sum(nil))
		if !strings.EqualFold(actual, expected) {
			return nil, fmt.Errorf("checksum mismatch for %s", name)
		}
		verified[name] = struct{}{}
		if strings.HasPrefix(name, "files/uploads/") {
			uploadCount++
		}
	}
	if manifest == nil {
		return nil, errors.New("backup manifest is missing")
	}
	if _, ok := seen[checksumsArchivePath]; !ok {
		return nil, errors.New("checksums.sha256 is missing")
	}
	for name := range manifest.Checksums {
		if _, ok := verified[name]; !ok {
			return nil, fmt.Errorf("manifest entry is missing from archive: %s", name)
		}
	}
	if uploadCount != manifest.UploadFileCount {
		return nil, fmt.Errorf("upload count mismatch: manifest=%d archive=%d", manifest.UploadFileCount, uploadCount)
	}
	return manifest, nil
}

func validateManifest(manifest *Manifest) error {
	if manifest.FormatVersion != ArchiveFormatVersion {
		return fmt.Errorf("unsupported backup format version: %d", manifest.FormatVersion)
	}
	if strings.TrimSpace(manifest.BackupID) == "" {
		return errors.New("backup manifest has no backup id")
	}
	if manifest.Checksums == nil {
		return errors.New("backup manifest has no checksums")
	}
	if _, ok := manifest.Checksums[databaseArchivePath]; !ok {
		return errors.New("backup manifest has no database dump")
	}
	for name, checksum := range manifest.Checksums {
		validated, err := validateRestoreArchivePath(name)
		if err != nil || validated == manifestArchivePath || validated == checksumsArchivePath {
			return fmt.Errorf("invalid manifest path: %s", name)
		}
		if validated != databaseArchivePath && !strings.HasPrefix(validated, "files/uploads/") {
			return fmt.Errorf("unsupported manifest path: %s", name)
		}
		if len(checksum) != sha256.Size*2 {
			return fmt.Errorf("invalid checksum for %s", name)
		}
		if _, err := hex.DecodeString(checksum); err != nil {
			return fmt.Errorf("invalid checksum for %s", name)
		}
	}
	return nil
}

func validateRestoreArchivePath(name string) (string, error) {
	if strings.ContainsRune(name, '\x00') || strings.Contains(name, "\\") {
		return "", fmt.Errorf("unsafe archive path: %q", name)
	}
	cleaned := path.Clean(name)
	if cleaned != name || cleaned == "." || strings.HasPrefix(cleaned, "/") || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("unsafe archive path: %q", name)
	}
	return cleaned, nil
}
