package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSnapshotUploadsAndWriteArchive(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	uploads := filepath.Join(root, "uploads")
	if err := os.MkdirAll(filepath.Join(uploads, "pictures"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(uploads, "pictures", "sample.txt"), []byte("uploaded-content"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot := filepath.Join(root, "snapshot")
	checksums, count, err := snapshotUploads(context.Background(), uploads, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 upload, got %d", count)
	}
	archiveUploadPath := "files/uploads/pictures/sample.txt"
	if checksums[archiveUploadPath] == "" {
		t.Fatalf("missing upload checksum: %#v", checksums)
	}

	dumpPath := filepath.Join(root, "site.dump")
	if err := os.WriteFile(dumpPath, []byte("database-dump"), 0o600); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(root, "backup.tar.gz")
	manifest := Manifest{
		FormatVersion: ArchiveFormatVersion,
		BackupID:      "backup-id",
		CreatedAt:     time.Now().UTC(),
		Checksums:     checksums,
	}
	if err := writeArchive(context.Background(), archivePath, dumpPath, snapshot, manifest); err != nil {
		t.Fatal(err)
	}

	entries := readTestArchive(t, archivePath)
	for _, name := range []string{"manifest.json", databaseArchivePath, archiveUploadPath, "checksums.sha256"} {
		if _, ok := entries[name]; !ok {
			t.Fatalf("archive entry %q is missing", name)
		}
	}
	var archivedManifest Manifest
	if err := json.Unmarshal(entries["manifest.json"], &archivedManifest); err != nil {
		t.Fatal(err)
	}
	if archivedManifest.Checksums[databaseArchivePath] == "" {
		t.Fatal("database checksum is missing from manifest")
	}
	if string(entries[archiveUploadPath]) != "uploaded-content" {
		t.Fatalf("unexpected upload content: %q", entries[archiveUploadPath])
	}
}

func TestSnapshotUploadsRejectsSymlinks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	source := filepath.Join(root, "uploads")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "secret")
	if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(source, "link")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := snapshotUploads(context.Background(), source, filepath.Join(root, "out")); err == nil {
		t.Fatal("expected a symlink validation error")
	}
}

func readTestArchive(t *testing.T, path string) map[string][]byte {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	entries := make(map[string][]byte)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		entries[header.Name] = body
	}
	return entries
}
