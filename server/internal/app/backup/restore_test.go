package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectArchiveRejectsTraversal(t *testing.T) {
	t.Parallel()
	archivePath := filepath.Join(t.TempDir(), "unsafe.tar.gz")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "../manifest.json", Mode: 0o600, Size: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("{}")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectArchive(context.Background(), archivePath, 1<<20, 1<<20); err == nil {
		t.Fatal("expected unsafe archive path to be rejected")
	}
}

func TestUploadSwapRollbackAndCommit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	uploads := filepath.Join(root, "uploads")
	source := filepath.Join(root, "restored")
	writeTestFile(t, filepath.Join(uploads, "old", "site.txt"), "old-site")
	writeTestFile(t, filepath.Join(source, "new", "site.txt"), "new-site")

	swap, err := prepareUploadSwap(context.Background(), uploads, source, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatal(err)
	}
	assertTestFile(t, filepath.Join(uploads, "new", "site.txt"), "new-site")
	if err := swap.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertTestFile(t, filepath.Join(uploads, "old", "site.txt"), "old-site")
	if _, err := os.Stat(filepath.Join(uploads, "new")); !os.IsNotExist(err) {
		t.Fatalf("new upload tree still exists after rollback: %v", err)
	}

	swap, err = prepareUploadSwap(context.Background(), uploads, source, "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatal(err)
	}
	if err := swap.Commit(); err != nil {
		t.Fatal(err)
	}
	assertTestFile(t, filepath.Join(uploads, "new", "site.txt"), "new-site")
	if _, err := os.Stat(filepath.Join(uploads, "old")); !os.IsNotExist(err) {
		t.Fatalf("old upload tree still exists after commit: %v", err)
	}
}

func TestRestorePostgresPassesDatabaseNameAndConnectionEnvironment(t *testing.T) {
	root := t.TempDir()
	argsPath := filepath.Join(root, "args")
	envPath := filepath.Join(root, "env")
	scriptPath := filepath.Join(root, "pg_restore")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$ARGS_OUTPUT\"\nprintf '%s|%s|%s|%s' \"$PGHOST\" \"$PGUSER\" \"$PGPASSWORD\" \"$PGDATABASE\" > \"$ENV_OUTPUT\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ARGS_OUTPUT", argsPath)
	t.Setenv("ENV_OUTPUT", envPath)
	if err := restorePostgres(context.Background(), scriptPath, "postgres://siteuser:secret@database:5432/site?sslmode=disable", filepath.Join(root, "site.dump")); err != nil {
		t.Fatal(err)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(args), "--dbname=site") {
		t.Fatalf("database name argument is missing: %s", args)
	}
	env, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(env) != "database|siteuser|secret|site" {
		t.Fatalf("unexpected postgres environment: %s", env)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertTestFile(t *testing.T, path, expected string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != expected {
		t.Fatalf("unexpected file content at %s: %q", path, raw)
	}
}
