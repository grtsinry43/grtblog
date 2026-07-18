package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const (
	restoreRequestFilename = ".restore-request.json"
	restoreRunningFilename = ".restore-running.json"
	restoreStatusFilename  = ".restore-status.json"
)

type RestoreRequest struct {
	ID              string    `json:"id"`
	BackupID        string    `json:"backupId"`
	ArchiveFilename string    `json:"archiveFilename"`
	RequestedAt     time.Time `json:"requestedAt"`
}

type RestoreStatus struct {
	State           string     `json:"state"`
	RequestID       string     `json:"requestId,omitempty"`
	BackupID        string     `json:"backupId,omitempty"`
	ArchiveFilename string     `json:"archiveFilename,omitempty"`
	Message         string     `json:"message,omitempty"`
	RequestedAt     *time.Time `json:"requestedAt,omitempty"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
}

func loadRestoreRequest(rootDir string) (*RestoreRequest, string, error) {
	for _, filename := range []string{restoreRequestFilename, restoreRunningFilename} {
		path := filepath.Join(rootDir, filename)
		raw, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, "", err
		}
		var request RestoreRequest
		if err := json.Unmarshal(raw, &request); err != nil {
			return nil, "", fmt.Errorf("parse restore request: %w", err)
		}
		if _, parseErr := uuid.Parse(request.ID); parseErr != nil || request.BackupID == "" || filepath.Base(request.ArchiveFilename) != request.ArchiveFilename {
			return nil, "", errors.New("invalid restore request")
		}
		return &request, path, nil
	}
	return nil, "", os.ErrNotExist
}

func readRestoreStatus(rootDir string) (*RestoreStatus, error) {
	raw, err := os.ReadFile(filepath.Join(rootDir, restoreStatusFilename))
	if errors.Is(err, os.ErrNotExist) {
		return &RestoreStatus{State: "idle"}, nil
	}
	if err != nil {
		return nil, err
	}
	var status RestoreStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func writeRestoreStatus(rootDir string, status RestoreStatus) error {
	return writeJSONAtomic(rootDir, restoreStatusFilename, status)
}

func writeJSONAtomic(rootDir, filename string, value any) error {
	if err := os.MkdirAll(rootDir, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(rootDir, ".json-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	failed := true
	defer func() {
		_ = temp.Close()
		if failed {
			_ = os.Remove(tempPath)
		}
	}()
	if err := temp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := temp.Write(raw); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, filepath.Join(rootDir, filename)); err != nil {
		return err
	}
	failed = false
	return nil
}
