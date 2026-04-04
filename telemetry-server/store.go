package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides database operations for the telemetry collector.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// --- Collect (C4 fix: all in a transaction) ---

// InsertReport upserts a telemetry report (one per instance per day) and
// replaces its error digests, all within a single transaction.
func (s *Store) InsertReport(ctx context.Context, r *IncomingReport) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var reportID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO telemetry_report
			(instance_id, version, go_version, os, arch, deploy_mode,
			 uptime_sec, features, metrics, error_count, panic_count, payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (instance_id, ((received_at AT TIME ZONE 'UTC')::date))
		DO UPDATE SET
			version     = EXCLUDED.version,
			go_version  = EXCLUDED.go_version,
			os          = EXCLUDED.os,
			arch        = EXCLUDED.arch,
			deploy_mode = EXCLUDED.deploy_mode,
			uptime_sec  = EXCLUDED.uptime_sec,
			features    = EXCLUDED.features,
			metrics     = EXCLUDED.metrics,
			error_count = EXCLUDED.error_count,
			panic_count = EXCLUDED.panic_count,
			payload     = EXCLUDED.payload,
			received_at = NOW()
		RETURNING id
	`, r.InstanceID, r.Version, r.GoVersion, r.OS, r.Arch, r.DeployMode,
		r.UptimeSec, r.Features, r.Metrics, r.ErrorCount, r.PanicCount, r.Payload,
	).Scan(&reportID)
	if err != nil {
		return 0, err
	}

	// Delete old digests for this report.
	if _, err := tx.Exec(ctx, `DELETE FROM telemetry_error_digest WHERE report_id = $1`, reportID); err != nil {
		return 0, err
	}

	// Bulk insert new digests.
	if len(r.Digests) > 0 {
		_, err = tx.CopyFrom(ctx,
			pgx.Identifier{"telemetry_error_digest"},
			[]string{"report_id", "fingerprint", "kind", "biz_code", "location", "sample_message", "count", "first_seen", "last_seen"},
			pgx.CopyFromSlice(len(r.Digests), func(i int) ([]any, error) {
				d := r.Digests[i]
				return []any{reportID, d.Fingerprint, d.Kind, d.BizCode, d.Location, d.SampleMessage, d.Count, d.FirstSeen, d.LastSeen}, nil
			}),
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return reportID, nil
}

// --- Admin queries ---

type ReportRow struct {
	ID         int64     `json:"id"`
	InstanceID string    `json:"instanceId"`
	Version    string    `json:"version"`
	OS         string    `json:"os"`
	Arch       string    `json:"arch"`
	DeployMode string    `json:"deployMode"`
	ErrorCount int       `json:"errorCount"`
	PanicCount int       `json:"panicCount"`
	ReceivedAt time.Time `json:"receivedAt"`
}

func (s *Store) ListReports(ctx context.Context, limit, offset int) ([]ReportRow, int64, error) {
	var total int64
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM telemetry_report`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, instance_id, version, os, arch, deploy_mode, error_count, panic_count, received_at
		FROM telemetry_report
		ORDER BY received_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []ReportRow
	for rows.Next() {
		var r ReportRow
		if err := rows.Scan(&r.ID, &r.InstanceID, &r.Version, &r.OS, &r.Arch, &r.DeployMode, &r.ErrorCount, &r.PanicCount, &r.ReceivedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, r)
	}
	return result, total, nil
}

type TopErrorRow struct {
	Fingerprint   string `json:"fingerprint"`
	BizCode       string `json:"bizCode"`
	Location      string `json:"location"`
	SampleMessage string `json:"sampleMessage"`
	TotalCount    int64  `json:"totalCount"`
	InstanceCount int64  `json:"instanceCount"`
}

func (s *Store) TopErrors(ctx context.Context, limit int, since time.Time) ([]TopErrorRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.fingerprint, d.biz_code, d.location,
			   (array_agg(d.sample_message ORDER BY d.count DESC))[1] AS sample_message,
			   SUM(d.count) AS total_count,
			   COUNT(DISTINCT r.instance_id) AS instance_count
		FROM telemetry_error_digest d
		JOIN telemetry_report r ON r.id = d.report_id
		WHERE r.received_at >= $1
		GROUP BY d.fingerprint, d.biz_code, d.location
		ORDER BY total_count DESC
		LIMIT $2
	`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopErrorRow
	for rows.Next() {
		var r TopErrorRow
		if err := rows.Scan(&r.Fingerprint, &r.BizCode, &r.Location, &r.SampleMessage, &r.TotalCount, &r.InstanceCount); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

type OverviewStats struct {
	TotalReports    int64            `json:"totalReports"`
	ActiveInstances int64            `json:"activeInstances"`
	VersionDist     map[string]int64 `json:"versionDist"`
	DeployModeDist  map[string]int64 `json:"deployModeDist"`
	FeatureAdoption map[string]int64 `json:"featureAdoption"`
	TotalErrors     int64            `json:"totalErrors"`
	TotalPanics     int64            `json:"totalPanics"`
}

func (s *Store) GetStats(ctx context.Context) (*OverviewStats, error) {
	stats := &OverviewStats{
		VersionDist:     make(map[string]int64),
		DeployModeDist:  make(map[string]int64),
		FeatureAdoption: make(map[string]int64),
	}

	since7d := time.Now().UTC().Add(-7 * 24 * time.Hour)

	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM telemetry_report`).Scan(&stats.TotalReports); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT instance_id) FROM telemetry_report WHERE received_at >= $1`, since7d).Scan(&stats.ActiveInstances); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(error_count),0), COALESCE(SUM(panic_count),0) FROM telemetry_report WHERE received_at >= $1`, since7d).Scan(&stats.TotalErrors, &stats.TotalPanics); err != nil {
		return nil, err
	}

	// Version distribution (latest report per instance in 7d).
	scanDist := func(query string, dest map[string]int64) {
		rows, err := s.pool.Query(ctx, query, since7d)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var k string
			var v int64
			if rows.Scan(&k, &v) == nil {
				dest[k] = v
			}
		}
	}

	scanDist(`SELECT version, COUNT(*) FROM (SELECT DISTINCT ON (instance_id) version FROM telemetry_report WHERE received_at >= $1 ORDER BY instance_id, received_at DESC) sub GROUP BY version`, stats.VersionDist)
	scanDist(`SELECT deploy_mode, COUNT(*) FROM (SELECT DISTINCT ON (instance_id) deploy_mode FROM telemetry_report WHERE received_at >= $1 ORDER BY instance_id, received_at DESC) sub GROUP BY deploy_mode`, stats.DeployModeDist)

	// Feature adoption.
	fRows, err := s.pool.Query(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE features->>'federationEnabled' = 'true') AS federation,
			COUNT(*) FILTER (WHERE features->>'activityPubEnabled' = 'true') AS activitypub,
			COUNT(*) FILTER (WHERE features->>'emailEnabled' = 'true') AS email,
			COUNT(*) FILTER (WHERE features->>'turnstileEnabled' = 'true') AS turnstile
		FROM (
			SELECT DISTINCT ON (instance_id) features
			FROM telemetry_report WHERE received_at >= $1
			ORDER BY instance_id, received_at DESC
		) sub
	`, since7d)
	if err == nil {
		defer fRows.Close()
		if fRows.Next() {
			var fed, ap, em, ts int64
			if fRows.Scan(&fed, &ap, &em, &ts) == nil {
				stats.FeatureAdoption["federation"] = fed
				stats.FeatureAdoption["activitypub"] = ap
				stats.FeatureAdoption["email"] = em
				stats.FeatureAdoption["turnstile"] = ts
			}
		}
	}

	return stats, nil
}

// --- WebAuthn credential storage ---

func (s *Store) SaveCredentialJSON(ctx context.Context, id string, credJSON []byte, displayName string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO passkey_credential (id, credential_json, display_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET credential_json = $2
	`, id, credJSON, displayName)
	return err
}

func (s *Store) GetCredentialJSON(ctx context.Context, id string) ([]byte, error) {
	var data []byte
	err := s.pool.QueryRow(ctx, `SELECT credential_json FROM passkey_credential WHERE id = $1`, id).Scan(&data)
	return data, err
}

func (s *Store) ListCredentialJSONs(ctx context.Context) ([][]byte, error) {
	rows, err := s.pool.Query(ctx, `SELECT credential_json FROM passkey_credential`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result [][]byte
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		result = append(result, data)
	}
	return result, nil
}

func (s *Store) HasAnyCredential(ctx context.Context) (bool, error) {
	var count int64
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM passkey_credential`).Scan(&count)
	return count > 0, err
}

// --- WebAuthn ceremony session storage ---

func (s *Store) SaveWebAuthnSession(ctx context.Context, id string, sessionJSON []byte, sessionType string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO webauthn_session (id, session_json, session_type, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET session_json = $2, expires_at = $4
	`, id, sessionJSON, sessionType, expiresAt)
	return err
}

func (s *Store) GetWebAuthnSession(ctx context.Context, id string, expectedType string) ([]byte, error) {
	var data []byte
	err := s.pool.QueryRow(ctx, `SELECT session_json FROM webauthn_session WHERE id = $1 AND session_type = $2 AND expires_at > NOW()`, id, expectedType).Scan(&data)
	return data, err
}

func (s *Store) DeleteWebAuthnSession(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM webauthn_session WHERE id = $1`, id)
	return err
}

// --- App session management ---

func (s *Store) SaveAdminSession(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO admin_session (token, expires_at) VALUES ($1, $2)
	`, token, expiresAt)
	return err
}

func (s *Store) ValidateAdminSession(ctx context.Context, token string) (bool, error) {
	var count int64
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_session WHERE token = $1 AND expires_at > NOW()`, token).Scan(&count)
	return count > 0, err
}

func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM admin_session WHERE expires_at < NOW()`)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `DELETE FROM webauthn_session WHERE expires_at < NOW()`)
	return err
}

// --- Helpers ---

func mustMarshalJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return b
}
