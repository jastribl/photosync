package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestConfig(t *testing.T, dir, key string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"places-api-key": "` + key + `"}`
	if err := os.WriteFile(filepath.Join(dir, "config", "config.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveAPIKeyOrder(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, "config-key")
	t.Chdir(dir)

	t.Setenv("GOOGLE_PLACES_API_KEY", "env-key")
	if got := resolveAPIKey("flag-key"); got != "flag-key" {
		t.Errorf("flag key should win, got %q", got)
	}
	if got := resolveAPIKey(""); got != "env-key" {
		t.Errorf("env key should win over config, got %q", got)
	}
	t.Setenv("GOOGLE_PLACES_API_KEY", "")
	if got := resolveAPIKey(""); got != "config-key" {
		t.Errorf("config key should be used when no flag/env, got %q", got)
	}
}

func TestResolveAPIKeyNoConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("GOOGLE_PLACES_API_KEY", "")
	if got := resolveAPIKey(""); got != "" {
		t.Errorf("expected empty key with no flag/env/config, got %q", got)
	}
}
