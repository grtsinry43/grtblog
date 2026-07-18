package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const databaseArchivePath = "database/site.dump"

func snapshotUploads(ctx context.Context, source, destination string) (map[string]string, int64, error) {
	checksums := make(map[string]string)
	var count int64
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return nil, 0, err
	}
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return checksums, 0, nil
	} else if err != nil {
		return nil, 0, err
	}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("upload tree contains unsupported symlink: %s", rel)
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("upload tree contains unsupported file: %s", rel)
		}
		sum, err := copyFileWithChecksum(path, target)
		if err != nil {
			return err
		}
		archivePath := filepath.ToSlash(filepath.Join("files", "uploads", rel))
		checksums[archivePath] = sum
		count++
		return nil
	})
	return checksums, count, err
}

func writeArchive(ctx context.Context, outputPath, dumpPath, uploadsPath string, manifest Manifest) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	failed := true
	defer func() {
		_ = file.Close()
		if failed {
			_ = os.Remove(outputPath)
		}
	}()
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)

	dumpSum, err := hashPath(dumpPath)
	if err != nil {
		return err
	}
	manifest.Checksums[databaseArchivePath] = dumpSum
	manifestRaw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := writeTarBytes(tw, "manifest.json", manifestRaw, 0o600); err != nil {
		return err
	}
	if err := writeTarFile(ctx, tw, databaseArchivePath, dumpPath); err != nil {
		return err
	}

	paths := make([]string, 0, len(manifest.Checksums))
	for path := range manifest.Checksums {
		if path != databaseArchivePath {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, archivePath := range paths {
		rel := strings.TrimPrefix(archivePath, "files/uploads/")
		if err := writeTarFile(ctx, tw, archivePath, filepath.Join(uploadsPath, filepath.FromSlash(rel))); err != nil {
			return err
		}
	}
	var checksumText strings.Builder
	allPaths := append([]string{databaseArchivePath}, paths...)
	for _, path := range allPaths {
		fmt.Fprintf(&checksumText, "%s  %s\n", manifest.Checksums[path], path)
	}
	if err := writeTarBytes(tw, "checksums.sha256", []byte(checksumText.String()), 0o600); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	failed = false
	return nil
}

func writeTarBytes(tw *tar.Writer, name string, raw []byte, mode int64) error {
	header := &tar.Header{Name: name, Mode: mode, Size: int64(len(raw))}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(raw)
	return err
}

func writeTarFile(ctx context.Context, tw *tar.Writer, archivePath, diskPath string) error {
	info, err := os.Stat(diskPath)
	if err != nil {
		return err
	}
	header := &tar.Header{Name: archivePath, Mode: 0o600, Size: info.Size()}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	file, err := os.Open(diskPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(tw, &contextReader{ctx: ctx, reader: file})
	return err
}

func copyFileWithChecksum(source, destination string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return "", err
	}
	in, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	failed := true
	defer func() {
		_ = out.Close()
		if failed {
			_ = os.Remove(destination)
		}
	}()
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, h), in); err != nil {
		return "", err
	}
	if err := out.Sync(); err != nil {
		return "", err
	}
	failed = false
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashPath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
