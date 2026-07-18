package config

import "testing"

func TestLoadUsesCurrentUpdateRepositoryByDefault(t *testing.T) {
	t.Setenv("APP_UPDATE_CHECK_REPO", "")

	cfg := Load()
	if cfg.App.UpdateCheckRepo != "grtsinry43/grtblog" {
		t.Fatalf("unexpected default update repository: %q", cfg.App.UpdateCheckRepo)
	}
}

func TestLoadHonorsConfiguredUpdateRepository(t *testing.T) {
	t.Setenv("APP_UPDATE_CHECK_REPO", "example/custom-releases")

	cfg := Load()
	if cfg.App.UpdateCheckRepo != "example/custom-releases" {
		t.Fatalf("unexpected configured update repository: %q", cfg.App.UpdateCheckRepo)
	}
}
