package backup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type PostgresDumper interface {
	Dump(ctx context.Context, snapshot, outputPath string) (string, error)
}

type CommandPostgresDumper struct {
	Binary string
	DSN    string
}

func (d CommandPostgresDumper) Dump(ctx context.Context, snapshot, outputPath string) (string, error) {
	binary := strings.TrimSpace(d.Binary)
	if binary == "" {
		binary = "pg_dump"
	}
	versionCmd := exec.CommandContext(ctx, binary, "--version")
	versionRaw, err := versionCmd.Output()
	if err != nil {
		return "", fmt.Errorf("inspect pg_dump version: %w", err)
	}
	args := []string{
		"--format=custom",
		"--compress=none",
		"--schema=public",
		"--strict-names",
		"--no-owner",
		"--no-privileges",
		"--no-comments",
		"--no-publications",
		"--no-subscriptions",
		"--no-security-labels",
		"--no-tablespaces",
		"--snapshot=" + snapshot,
		"--file=" + outputPath,
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), "PGDATABASE="+d.DSN)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if len(detail) > 4096 {
			detail = detail[len(detail)-4096:]
		}
		return "", fmt.Errorf("pg_dump failed: %w: %s", err, detail)
	}
	return strings.TrimSpace(string(versionRaw)), nil
}
