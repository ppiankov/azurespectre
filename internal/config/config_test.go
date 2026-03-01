package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	content := `
subscription: "sub-123"
idle_days: 14
stale_days: 60
min_monthly_cost: 10.0
format: json
exclude:
  resource_ids:
    - "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"
  tags:
    - "Environment=production"
    - "DoNotDelete"
`
	if err := os.WriteFile(filepath.Join(dir, ".azurespectre.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Subscription != "sub-123" {
		t.Errorf("Subscription = %q, want sub-123", cfg.Subscription)
	}
	if cfg.IdleDays != 14 {
		t.Errorf("IdleDays = %d, want 14", cfg.IdleDays)
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want json", cfg.Format)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Subscription != "" {
		t.Error("expected empty config when file not found")
	}
}

func TestParseTags(t *testing.T) {
	e := Exclude{Tags: []string{"Environment=production", "DoNotDelete"}}
	tags := e.ParseTags()
	if tags["Environment"] != "production" {
		t.Errorf("Environment = %q, want production", tags["Environment"])
	}
	if v, ok := tags["DoNotDelete"]; !ok || v != "" {
		t.Errorf("DoNotDelete = %q (%v), want empty string", v, ok)
	}
}

func TestTimeoutDuration(t *testing.T) {
	cfg := Config{Timeout: "5m"}
	if cfg.TimeoutDuration() != 5*time.Minute {
		t.Errorf("timeout = %v, want 5m", cfg.TimeoutDuration())
	}
}

func TestTimeoutDurationDefault(t *testing.T) {
	cfg := Config{}
	if cfg.TimeoutDuration() != 10*time.Minute {
		t.Errorf("timeout = %v, want 10m default", cfg.TimeoutDuration())
	}
}
