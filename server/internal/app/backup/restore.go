package backup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/config"
)

func ExecutePendingRestore(ctx context.Context, cfg config.BackupConfig, databaseDSN string) error {
	request, requestPath, err := loadRestoreRequest(cfg.RootDir)
	if err != nil {
		return err
	}
	runningPath := filepath.Join(cfg.RootDir, restoreRunningFilename)
	if filepath.Base(requestPath) == restoreRequestFilename {
		if err := os.Rename(requestPath, runningPath); err != nil {
			return fmt.Errorf("claim restore request: %w", err)
		}
		requestPath = runningPath
	}
	now := time.Now().UTC()
	status := RestoreStatus{
		State: "running", RequestID: request.ID, BackupID: request.BackupID,
		ArchiveFilename: request.ArchiveFilename, Message: "正在离线恢复站点",
		RequestedAt: &request.RequestedAt, StartedAt: &now,
	}
	if err := writeRestoreStatus(cfg.RootDir, status); err != nil {
		return err
	}
	fail := func(restoreErr error) error {
		completed := time.Now().UTC()
		status.State, status.Message, status.CompletedAt = "failed", restoreErr.Error(), &completed
		_ = writeRestoreStatus(cfg.RootDir, status)
		_ = os.Remove(requestPath)
		return restoreErr
	}

	archivePath := filepath.Join(cfg.RootDir, filepath.Base(request.ArchiveFilename))
	workDir := filepath.Join(cfg.RootDir, ".restore-work-"+request.ID)
	if err := os.RemoveAll(workDir); err != nil {
		return fail(err)
	}
	defer os.RemoveAll(workDir)
	manifest, err := extractArchive(ctx, archivePath, workDir, cfg.RestoreMaxArchiveBytes, cfg.RestoreMaxExtractedBytes)
	if err != nil {
		return fail(fmt.Errorf("validate and extract backup: %w", err))
	}
	if manifest.BackupID != request.BackupID {
		return fail(errors.New("restore request does not match backup manifest"))
	}

	swap, err := prepareUploadSwap(ctx, cfg.UploadDir, filepath.Join(workDir, "files", "uploads"), request.ID)
	if err != nil {
		return fail(fmt.Errorf("prepare upload restore: %w", err))
	}
	if err := restorePostgres(ctx, cfg.PGRestoreBin, databaseDSN, filepath.Join(workDir, filepath.FromSlash(databaseArchivePath))); err != nil {
		if rollbackErr := swap.Rollback(); rollbackErr != nil {
			err = fmt.Errorf("%w; upload rollback also failed: %v", err, rollbackErr)
		}
		return fail(err)
	}
	cleanupErr := swap.Commit()

	completed := time.Now().UTC()
	status.State, status.Message, status.CompletedAt = "succeeded", "站点已从备份完整恢复", &completed
	if cleanupErr != nil {
		status.Message = "站点恢复成功，但旧上传文件暂存目录清理失败: " + cleanupErr.Error()
	}
	if err := writeRestoreStatus(cfg.RootDir, status); err != nil {
		return err
	}
	return os.Remove(requestPath)
}

func restorePostgres(ctx context.Context, binary, databaseDSN, dumpPath string) error {
	if strings.TrimSpace(binary) == "" {
		binary = "pg_restore"
	}
	postgresEnv, err := postgresCommandEnv(databaseDSN)
	if err != nil {
		return err
	}
	databaseName := ""
	for _, value := range postgresEnv {
		if strings.HasPrefix(value, "PGDATABASE=") {
			databaseName = strings.TrimPrefix(value, "PGDATABASE=")
			break
		}
	}
	if databaseName == "" {
		return errors.New("postgres database name is missing")
	}
	args := []string{
		"--clean", "--if-exists", "--no-owner", "--no-privileges", "--single-transaction",
		"--dbname=" + databaseName, dumpPath,
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), postgresEnv...)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(output.String())
		if len(detail) > 8192 {
			detail = detail[len(detail)-8192:]
		}
		return fmt.Errorf("pg_restore failed: %w: %s", err, detail)
	}
	return nil
}

type uploadSwap struct {
	root       string
	stagingDir string
	oldDir     string
}

func prepareUploadSwap(ctx context.Context, uploadDir, source, requestID string) (*uploadSwap, error) {
	if strings.TrimSpace(uploadDir) == "" || filepath.Clean(uploadDir) == "." || filepath.Clean(uploadDir) == string(filepath.Separator) {
		return nil, errors.New("unsafe upload directory")
	}
	if err := os.MkdirAll(uploadDir, 0o700); err != nil {
		return nil, err
	}
	swap := &uploadSwap{
		root: uploadDir, stagingDir: filepath.Join(uploadDir, ".restore-new-"+requestID),
		oldDir: filepath.Join(uploadDir, ".restore-old-"+requestID),
	}
	if _, err := os.Stat(swap.oldDir); err == nil {
		if rollbackErr := swap.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("recover interrupted upload swap: %w", rollbackErr)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err := os.RemoveAll(swap.stagingDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(swap.stagingDir, 0o700); err != nil {
		return nil, err
	}
	if err := copyRegularTree(ctx, source, swap.stagingDir); err != nil {
		_ = os.RemoveAll(swap.stagingDir)
		return nil, err
	}
	if err := os.MkdirAll(swap.oldDir, 0o700); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		current := filepath.Join(uploadDir, entry.Name())
		if current == swap.stagingDir || current == swap.oldDir {
			continue
		}
		if err := os.Rename(current, filepath.Join(swap.oldDir, entry.Name())); err != nil {
			_ = swap.Rollback()
			return nil, err
		}
	}
	newEntries, err := os.ReadDir(swap.stagingDir)
	if err != nil {
		_ = swap.Rollback()
		return nil, err
	}
	for _, entry := range newEntries {
		if err := os.Rename(filepath.Join(swap.stagingDir, entry.Name()), filepath.Join(uploadDir, entry.Name())); err != nil {
			_ = swap.Rollback()
			return nil, err
		}
	}
	return swap, nil
}

func (s *uploadSwap) Rollback() error {
	if _, err := os.Stat(s.oldDir); errors.Is(err, os.ErrNotExist) {
		return os.RemoveAll(s.stagingDir)
	} else if err != nil {
		return err
	}
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		current := filepath.Join(s.root, entry.Name())
		if current == s.oldDir || current == s.stagingDir {
			continue
		}
		if err := os.RemoveAll(current); err != nil {
			return err
		}
	}
	oldEntries, err := os.ReadDir(s.oldDir)
	if err != nil {
		return err
	}
	for _, entry := range oldEntries {
		if err := os.Rename(filepath.Join(s.oldDir, entry.Name()), filepath.Join(s.root, entry.Name())); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(s.oldDir); err != nil {
		return err
	}
	return os.RemoveAll(s.stagingDir)
}

func (s *uploadSwap) Commit() error {
	if err := os.RemoveAll(s.oldDir); err != nil {
		return err
	}
	return os.RemoveAll(s.stagingDir)
}

func copyRegularTree(ctx context.Context, source, destination string) error {
	if _, err := os.Stat(source); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(source, current)
		if err != nil || rel == "." {
			return err
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsupported restored upload entry: %s", rel)
		}
		in, err := os.Open(current)
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			_ = in.Close()
			return err
		}
		_, copyErr := io.Copy(out, &contextReader{ctx: ctx, reader: in})
		closeInErr, closeOutErr := in.Close(), out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeInErr != nil {
			return closeInErr
		}
		return closeOutErr
	})
}
