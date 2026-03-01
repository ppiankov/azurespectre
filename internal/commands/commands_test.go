package commands

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ppiankov/azurespectre/internal/azure"
	"github.com/ppiankov/azurespectre/internal/report"
)

func TestExecuteVersion(t *testing.T) {
	old := rootCmd.OutOrStdout()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})
	defer func() { rootCmd.SetOut(old); rootCmd.SetArgs(nil) }()

	version, commit, date = "0.1.0", "abc123", "2026-03-01"
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
}

func TestExecuteScanNoSubscription(t *testing.T) {
	rootCmd.SetArgs([]string{"scan"})
	defer rootCmd.SetArgs(nil)

	t.Setenv("AZURE_SUBSCRIPTION_ID", "")
	scanFlags.subscription = ""

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when no subscription")
	}
}

func TestSelectReporterText(t *testing.T) {
	r, cleanup, err := selectReporter("text", "")
	if err != nil {
		t.Fatalf("selectReporter error: %v", err)
	}
	if cleanup != nil {
		t.Error("cleanup should be nil for stdout")
	}
	if _, ok := r.(*report.TextReporter); !ok {
		t.Error("expected TextReporter")
	}
}

func TestSelectReporterJSON(t *testing.T) {
	r, _, err := selectReporter("json", "")
	if err != nil {
		t.Fatalf("selectReporter error: %v", err)
	}
	if _, ok := r.(*report.JSONReporter); !ok {
		t.Error("expected JSONReporter")
	}
}

func TestSelectReporterSARIF(t *testing.T) {
	r, _, err := selectReporter("sarif", "")
	if err != nil {
		t.Fatalf("selectReporter error: %v", err)
	}
	if _, ok := r.(*report.SARIFReporter); !ok {
		t.Error("expected SARIFReporter")
	}
}

func TestSelectReporterSpectreHub(t *testing.T) {
	r, _, err := selectReporter("spectrehub", "")
	if err != nil {
		t.Fatalf("selectReporter error: %v", err)
	}
	if _, ok := r.(*report.SpectreHubReporter); !ok {
		t.Error("expected SpectreHubReporter")
	}
}

func TestSelectReporterOutputFile(t *testing.T) {
	path := t.TempDir() + "/output.json"
	r, cleanup, err := selectReporter("json", path)
	if err != nil {
		t.Fatalf("selectReporter error: %v", err)
	}
	if cleanup == nil {
		t.Error("cleanup should be non-nil for file output")
	}
	defer cleanup()
	if _, ok := r.(*report.JSONReporter); !ok {
		t.Error("expected JSONReporter")
	}
}

func TestBuildExcludeConfig(t *testing.T) {
	cfg.Exclude.ResourceIDs = []string{"res1"}
	cfg.Exclude.Tags = []string{"env=prod"}
	scanFlags.excludeTags = []string{"team=platform"}
	defer func() {
		cfg.Exclude.ResourceIDs = nil
		cfg.Exclude.Tags = nil
		scanFlags.excludeTags = nil
	}()

	exc := buildExcludeConfig()
	if !exc.ResourceIDs["res1"] {
		t.Error("missing resource ID from config")
	}
	if exc.Tags["env"] != "prod" {
		t.Errorf("env tag = %q, want prod", exc.Tags["env"])
	}
	if exc.Tags["team"] != "platform" {
		t.Errorf("team tag = %q, want platform", exc.Tags["team"])
	}
}

func TestComputeTargetHash(t *testing.T) {
	h := computeTargetHash("sub-123")
	if !strings.HasPrefix(h, "sha256:") {
		t.Errorf("hash = %q, want sha256: prefix", h)
	}
}

func TestEnhanceError(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{"DefaultAzureCredential failed", "az login"},
		{"AuthorizationFailed", "Reader role"},
		{"SubscriptionNotFound", "subscription ID"},
		{"context deadline exceeded", "timed out"},
		{"random error", "random error"},
	}
	for _, tt := range tests {
		err := enhanceError("test", fmt.Errorf("%s", tt.msg))
		if !strings.Contains(err.Error(), tt.want) {
			t.Errorf("enhanceError(%q) = %q, want contains %q", tt.msg, err.Error(), tt.want)
		}
	}
}

func TestInitCommand(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	err := runInit(nil, nil)
	if err != nil {
		t.Fatalf("runInit() error: %v", err)
	}

	if _, err := os.Stat(".azurespectre.yaml"); os.IsNotExist(err) {
		t.Error(".azurespectre.yaml not created")
	}
	if _, err := os.Stat("azurespectre-role.json"); os.IsNotExist(err) {
		t.Error("azurespectre-role.json not created")
	}
}

func TestInitCommand_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	_ = os.WriteFile(".azurespectre.yaml", []byte("existing"), 0o644)

	err := runInit(nil, nil)
	if err != nil {
		t.Fatalf("runInit() error: %v", err)
	}

	data, _ := os.ReadFile(".azurespectre.yaml")
	if string(data) != "existing" {
		t.Error("should not overwrite existing config")
	}
}

func TestApplyConfigDefaults(t *testing.T) {
	origSub := cfg.Subscription
	origRG := cfg.ResourceGroup
	origFmt := cfg.Format
	defer func() {
		cfg.Subscription = origSub
		cfg.ResourceGroup = origRG
		cfg.Format = origFmt
		scanFlags.subscription = ""
		scanFlags.resourceGroup = ""
		scanFlags.format = "text"
	}()

	cfg.Subscription = "from-config"
	cfg.ResourceGroup = "rg-from-config"
	cfg.Format = "json"
	scanFlags.subscription = ""
	scanFlags.resourceGroup = ""
	scanFlags.format = "text"

	applyConfigDefaults()

	if scanFlags.subscription != "from-config" {
		t.Errorf("subscription = %q, want %q", scanFlags.subscription, "from-config")
	}
	if scanFlags.resourceGroup != "rg-from-config" {
		t.Errorf("resourceGroup = %q, want %q", scanFlags.resourceGroup, "rg-from-config")
	}
	if scanFlags.format != "json" {
		t.Errorf("format = %q, want %q", scanFlags.format, "json")
	}
}

func TestIndexByte(t *testing.T) {
	if indexByte("key=value", '=') != 3 {
		t.Error("expected index 3")
	}
	if indexByte("noequals", '=') != -1 {
		t.Error("expected -1")
	}
}

func TestExcludeConfigMerge(t *testing.T) {
	exc := azure.ExcludeConfig{
		ResourceIDs: map[string]bool{"a": true},
		Tags:        map[string]string{"k": "v"},
	}
	if !exc.ResourceIDs["a"] {
		t.Error("missing resource ID")
	}
	if exc.Tags["k"] != "v" {
		t.Error("missing tag")
	}
}
