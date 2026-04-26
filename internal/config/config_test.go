package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "portwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
scan_interval: 10s
ports:
  - port: 80
    protocol: tcp
    name: http
  - port: 443
    protocol: tcp
    name: https
alerts:
  webhook_url: https://example.com/hook
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ScanInterval != 10*time.Second {
		t.Errorf("scan_interval: got %v, want 10s", cfg.ScanInterval)
	}
	if len(cfg.Ports) != 2 {
		t.Fatalf("ports: got %d, want 2", len(cfg.Ports))
	}
	if cfg.Ports[0].Port != 80 {
		t.Errorf("ports[0].port: got %d, want 80", cfg.Ports[0].Port)
	}
	if cfg.Alerts.WebhookURL != "https://example.com/hook" {
		t.Errorf("webhook_url: got %q", cfg.Alerts.WebhookURL)
	}
}

func TestLoad_DefaultScanInterval(t *testing.T) {
	path := writeTempConfig(t, `
ports:
  - port: 22
    name: ssh
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ScanInterval != 30*time.Second {
		t.Errorf("default scan_interval: got %v, want 30s", cfg.ScanInterval)
	}
	if cfg.Ports[0].Protocol != "tcp" {
		t.Errorf("default protocol: got %q, want tcp", cfg.Ports[0].Protocol)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	path := writeTempConfig(t, `
ports:
  - port: 99999
    protocol: tcp
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
