package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePrecedence(t *testing.T) {
	cfg := &Config{
		Current: "prod",
		Profiles: map[string]*Profile{
			"prod": {Server: "https://prod.example.com", Token: "gt_prod"},
			"dev":  {Server: "http://127.0.0.1:8080", Token: "gt_dev"},
		},
	}

	t.Run("profile from current", func(t *testing.T) {
		r := cfg.Resolve("", "", "")
		if r.Profile != "prod" || r.Server != "https://prod.example.com" || r.Token != "gt_prod" {
			t.Errorf("unexpected: %+v", r)
		}
	})

	t.Run("flag overrides profile", func(t *testing.T) {
		r := cfg.Resolve("dev", "https://flag.example.com", "gt_flag")
		if r.Profile != "dev" || r.Server != "https://flag.example.com" || r.Token != "gt_flag" {
			t.Errorf("unexpected: %+v", r)
		}
	})

	t.Run("env overrides profile", func(t *testing.T) {
		t.Setenv(EnvServer, "https://env.example.com")
		t.Setenv(EnvToken, "gt_env")
		r := cfg.Resolve("dev", "", "")
		if r.Server != "https://env.example.com" || r.Token != "gt_env" {
			t.Errorf("unexpected: %+v", r)
		}
	})

	t.Run("flag overrides env", func(t *testing.T) {
		t.Setenv(EnvServer, "https://env.example.com")
		r := cfg.Resolve("", "https://flag.example.com", "")
		if r.Server != "https://flag.example.com" {
			t.Errorf("unexpected: %+v", r)
		}
	})
}

func TestNormalizeServer(t *testing.T) {
	cases := map[string]string{
		"blog.example.com":         "https://blog.example.com",
		"https://blog.example.com": "https://blog.example.com",
		"http://127.0.0.1:8080/":   "http://127.0.0.1:8080",
		"":                         "",
	}
	for in, want := range cases {
		if got := NormalizeServer(in); got != want {
			t.Errorf("NormalizeServer(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.yaml")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom missing file: %v", err)
	}
	cfg.UpsertProfile("default", Profile{Server: "https://a.example.com", Token: "gt_a"})
	cfg.UpsertProfile("dev", Profile{Server: "http://127.0.0.1:8080", Token: "gt_b"})
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config perm = %o, want 600", perm)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if loaded.Current != "default" {
		t.Errorf("Current = %q, want default", loaded.Current)
	}
	if p := loaded.Profiles["dev"]; p == nil || p.Token != "gt_b" {
		t.Errorf("dev profile not persisted: %+v", p)
	}
}

func TestRemoveProfile(t *testing.T) {
	cfg := &Config{Current: "a", Profiles: map[string]*Profile{"a": {}, "b": {}}}
	if !cfg.RemoveProfile("a") {
		t.Fatal("expected removal")
	}
	if cfg.Current != "b" {
		t.Errorf("Current = %q, want b", cfg.Current)
	}
	if cfg.RemoveProfile("missing") {
		t.Error("unexpected removal of missing profile")
	}
}
