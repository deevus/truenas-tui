package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deevus/truenas-tui/config"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(`
[servers.home]
host = "truenas.local"
port = 443
username = "admin"
api_key = "1-abc"

[servers.offsite]
host = "backup.example.com"
port = 443
username = "admin"
api_key = "1-xyz"
insecure_skip_verify = true
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(cfg.Servers))
	}

	home := cfg.Servers["home"]
	if home.Host != "truenas.local" {
		t.Errorf("expected host truenas.local, got %s", home.Host)
	}
	if home.Port != 443 {
		t.Errorf("expected port 443, got %d", home.Port)
	}
	if home.APIKey != "1-abc" {
		t.Errorf("expected api_key 1-abc, got %s", home.APIKey)
	}

	offsite := cfg.Servers["offsite"]
	if !offsite.InsecureSkipVerify {
		t.Error("expected insecure_skip_verify=true for offsite")
	}
}

func TestLoad_WithSSH(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(`
[servers.home]
host = "truenas.local"
port = 443
username = "admin"
api_key = "1-abc"

[servers.home.ssh]
host = "truenas.local"
port = 22
username = "root"
private_key_path = "/home/test/.ssh/id_ed25519"
host_key_fingerprint = "SHA256:abc123"
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ssh := cfg.Servers["home"].SSH
	if ssh == nil {
		t.Fatal("expected SSH config")
	}
	if ssh.Host != "truenas.local" {
		t.Errorf("expected ssh host truenas.local, got %s", ssh.Host)
	}
	if ssh.Port != 22 {
		t.Errorf("expected ssh port 22, got %d", ssh.Port)
	}
	if ssh.Username != "root" {
		t.Errorf("expected ssh username root, got %s", ssh.Username)
	}
}

func TestLoad_SSHDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(`
[servers.home]
host = "truenas.local"
port = 443
username = "admin"
api_key = "1-abc"

[servers.home.ssh]
private_key_path = "/home/test/.ssh/id_ed25519"
host_key_fingerprint = "SHA256:abc123"
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ssh := cfg.Servers["home"].SSH
	if ssh == nil {
		t.Fatal("expected SSH config")
	}
	if ssh.Port != 22 {
		t.Errorf("expected default ssh port 22, got %d", ssh.Port)
	}
	if ssh.Username != "admin" {
		t.Errorf("expected ssh username to default to server username 'admin', got %s", ssh.Username)
	}
}

func TestLoad_ExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err = os.WriteFile(path, []byte(`
[servers.home]
host = "truenas.local"
port = 443
username = "admin"
api_key = "1-abc"

[servers.home.ssh]
private_key_path = "~/.ssh/id_ed25519"
host_key_fingerprint = "SHA256:abc123"
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(home, ".ssh", "id_ed25519")
	if cfg.Servers["home"].SSH.PrivateKeyPath != expected {
		t.Errorf("expected expanded path %s, got %s", expected, cfg.Servers["home"].SSH.PrivateKeyPath)
	}
}

func TestLoad_ExpandEnvVar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TEST_KEY_DIR", "/custom/keys")

	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(`
[servers.home]
host = "truenas.local"
port = 443
username = "admin"
api_key = "1-abc"

[servers.home.ssh]
private_key_path = "$TEST_KEY_DIR/id_ed25519"
host_key_fingerprint = "SHA256:abc123"
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "/custom/keys/id_ed25519"
	if cfg.Servers["home"].SSH.PrivateKeyPath != expected {
		t.Errorf("expected expanded path %s, got %s", expected, cfg.Servers["home"].SSH.PrivateKeyPath)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.LoadFrom("/nonexistent/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_EmptyServers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(``), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = config.LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestConfig_ServerNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(`
[servers.alpha]
host = "a.local"
port = 443
username = "admin"
api_key = "1-a"

[servers.beta]
host = "b.local"
port = 443
username = "admin"
api_key = "1-b"
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := cfg.ServerNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("expected [alpha beta], got %v", names)
	}
}

func TestDefaultPath(t *testing.T) {
	path := config.DefaultPath()
	if path == "" {
		t.Fatal("expected non-empty default path")
	}
}
