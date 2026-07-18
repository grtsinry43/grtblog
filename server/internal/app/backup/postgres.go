package backup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
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
	postgresEnv, err := postgresCommandEnv(d.DSN)
	if err != nil {
		return "", err
	}
	cmd.Env = append(os.Environ(), postgresEnv...)
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

func postgresCommandEnv(dsn string) ([]string, error) {
	parsed, err := url.Parse(strings.TrimSpace(dsn))
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
		return nil, errors.New("backup commands require a postgres:// or postgresql:// DB_DSN")
	}
	if parsed.User == nil || parsed.User.Username() == "" || parsed.Hostname() == "" {
		return nil, errors.New("postgres DB_DSN must include user and host")
	}
	database := strings.TrimPrefix(parsed.EscapedPath(), "/")
	database, err = url.PathUnescape(database)
	if err != nil || database == "" {
		return nil, errors.New("postgres DB_DSN must include a database name")
	}
	port := parsed.Port()
	if port == "" {
		port = strconv.Itoa(5432)
	}
	env := []string{
		"PGHOST=" + parsed.Hostname(), "PGPORT=" + port,
		"PGUSER=" + parsed.User.Username(), "PGDATABASE=" + database,
	}
	if password, ok := parsed.User.Password(); ok {
		env = append(env, "PGPASSWORD="+password)
	}
	queryToEnv := map[string]string{
		"sslmode": "PGSSLMODE", "sslcert": "PGSSLCERT", "sslkey": "PGSSLKEY",
		"sslrootcert": "PGSSLROOTCERT", "sslcrl": "PGSSLCRL", "sslcrldir": "PGSSLCRLDIR",
		"connect_timeout": "PGCONNECT_TIMEOUT", "application_name": "PGAPPNAME",
		"options": "PGOPTIONS", "target_session_attrs": "PGTARGETSESSIONATTRS",
		"channel_binding": "PGCHANNELBINDING", "gssencmode": "PGGSSENCMODE",
	}
	for queryKey, envKey := range queryToEnv {
		if value := parsed.Query().Get(queryKey); value != "" {
			env = append(env, envKey+"="+value)
		}
	}
	return env, nil
}
