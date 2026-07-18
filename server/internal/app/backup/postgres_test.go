package backup

import (
	"slices"
	"testing"
)

func TestPostgresCommandEnvParsesURLWithoutCommandArguments(t *testing.T) {
	t.Parallel()
	env, err := postgresCommandEnv("postgres://backup%20user:s%40cret@db.internal:5544/site%20db?sslmode=require&connect_timeout=7")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"PGHOST=db.internal", "PGPORT=5544", "PGUSER=backup user", "PGPASSWORD=s@cret",
		"PGDATABASE=site db", "PGSSLMODE=require", "PGCONNECT_TIMEOUT=7",
	} {
		if !slices.Contains(env, expected) {
			t.Fatalf("missing %q from postgres environment: %#v", expected, env)
		}
	}
}

func TestPostgresCommandEnvRejectsKeywordDSN(t *testing.T) {
	t.Parallel()
	if _, err := postgresCommandEnv("host=localhost dbname=site"); err == nil {
		t.Fatal("expected keyword DSN to be rejected")
	}
}
