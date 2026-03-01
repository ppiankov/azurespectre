package commands

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/ppiankov/azurespectre/internal/analyzer"
	"github.com/ppiankov/azurespectre/internal/azure"
	"github.com/ppiankov/azurespectre/internal/report"
	"github.com/spf13/cobra"
)

var scanFlags struct {
	subscription   string
	resourceGroup  string
	idleDays       int
	staleDays      int
	stoppedDays    int
	idleCPU        float64
	format         string
	outputFile     string
	minMonthlyCost float64
	excludeTags    []string
	noProgress     bool
	timeout        time.Duration
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Azure subscription for resource waste",
	RunE:  runScan,
}

func init() {
	f := scanCmd.Flags()
	f.StringVar(&scanFlags.subscription, "subscription", "", "Azure subscription ID (env: AZURE_SUBSCRIPTION_ID)")
	f.StringVar(&scanFlags.resourceGroup, "resource-group", "", "Limit scan to resource group")
	f.IntVar(&scanFlags.idleDays, "idle-days", 7, "CPU utilization lookback window (days)")
	f.IntVar(&scanFlags.staleDays, "stale-days", 90, "Snapshot age threshold (days)")
	f.IntVar(&scanFlags.stoppedDays, "stopped-days", 30, "VM deallocated threshold (days)")
	f.Float64Var(&scanFlags.idleCPU, "idle-cpu", 5.0, "CPU percentage below which a VM is idle")
	f.StringVar(&scanFlags.format, "format", "text", "Output format: text, json, sarif, spectrehub")
	f.StringVarP(&scanFlags.outputFile, "output", "o", "", "Output file (default: stdout)")
	f.Float64Var(&scanFlags.minMonthlyCost, "min-monthly-cost", 5.0, "Minimum monthly waste to report ($)")
	f.StringSliceVar(&scanFlags.excludeTags, "exclude-tags", nil, "Exclude resources by tag Key=Value (repeatable)")
	f.BoolVar(&scanFlags.noProgress, "no-progress", false, "Disable progress output")
	f.DurationVar(&scanFlags.timeout, "timeout", 10*time.Minute, "Scan timeout")
}

func runScan(_ *cobra.Command, _ []string) error {
	applyConfigDefaults()

	sub := scanFlags.subscription
	if sub == "" {
		sub = os.Getenv("AZURE_SUBSCRIPTION_ID")
	}
	if sub == "" {
		return fmt.Errorf("--subscription is required (or set AZURE_SUBSCRIPTION_ID)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), scanFlags.timeout)
	defer cancel()

	slog.Info("starting scan", "subscription", sub)

	compute, err := azure.NewComputeClient(sub)
	if err != nil {
		return enhanceError("create compute client", err)
	}

	network, err := azure.NewNetworkClient(sub)
	if err != nil {
		return enhanceError("create network client", err)
	}

	monitor, err := azure.NewMonitorClient()
	if err != nil {
		slog.Warn("Azure Monitor unavailable, skipping metric-based checks", "error", err)
	}

	scanCfg := azure.ScanConfig{
		IdleDays:       scanFlags.idleDays,
		StaleDays:      scanFlags.staleDays,
		StoppedDays:    scanFlags.stoppedDays,
		IdleCPU:        scanFlags.idleCPU,
		MinMonthlyCost: scanFlags.minMonthlyCost,
		ResourceGroup:  scanFlags.resourceGroup,
		Exclude:        buildExcludeConfig(),
	}

	scanner := azure.NewSubscriptionScanner(compute, network, monitor, sub, scanCfg)

	sqlClient, err := azure.NewSQLClient(sub)
	if err != nil {
		slog.Warn("Azure SQL unavailable, skipping SQL checks", "error", err)
	} else {
		scanner.SetSQLAPI(sqlClient)
	}

	appServiceClient, err := azure.NewAppServiceClient(sub)
	if err != nil {
		slog.Warn("Azure App Service unavailable, skipping App Service checks", "error", err)
	} else {
		scanner.SetAppServiceAPI(appServiceClient)
	}

	storageClient, err := azure.NewStorageClient(sub)
	if err != nil {
		slog.Warn("Azure Storage unavailable, skipping storage checks", "error", err)
	} else {
		scanner.SetStorageAPI(storageClient)
	}

	if !scanFlags.noProgress {
		scanner.SetProgressFn(func(p azure.ScanProgress) {
			_, _ = fmt.Fprintf(os.Stderr, "  scanning %s...\n", p.Scanner)
		})
	}

	result, err := scanner.ScanAll(ctx)
	if err != nil {
		return enhanceError("scan", err)
	}

	analysis := analyzer.Analyze(result, analyzer.AnalyzerConfig{
		MinMonthlyCost: scanFlags.minMonthlyCost,
	})

	data := report.Data{
		Tool:      "azurespectre",
		Version:   version,
		Timestamp: time.Now().UTC(),
		Target: report.Target{
			Type:    "azure-subscription",
			URIHash: computeTargetHash(sub),
		},
		Config: report.ReportConfig{
			Subscription:   sub,
			ResourceGroup:  scanFlags.resourceGroup,
			IdleDays:       scanFlags.idleDays,
			StaleDays:      scanFlags.staleDays,
			StoppedDays:    scanFlags.stoppedDays,
			MinMonthlyCost: scanFlags.minMonthlyCost,
		},
		Findings: analysis.Findings,
		Summary:  analysis.Summary,
		Errors:   analysis.Errors,
	}

	reporter, cleanup, err := selectReporter(scanFlags.format, scanFlags.outputFile)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	return reporter.Generate(data)
}

func applyConfigDefaults() {
	if scanFlags.subscription == "" && cfg.Subscription != "" {
		scanFlags.subscription = cfg.Subscription
	}
	if scanFlags.resourceGroup == "" && cfg.ResourceGroup != "" {
		scanFlags.resourceGroup = cfg.ResourceGroup
	}
	if scanFlags.format == "text" && cfg.Format != "" {
		scanFlags.format = cfg.Format
	}
}

func buildExcludeConfig() azure.ExcludeConfig {
	exc := azure.ExcludeConfig{
		ResourceIDs: make(map[string]bool),
		Tags:        make(map[string]string),
	}

	for _, id := range cfg.Exclude.ResourceIDs {
		exc.ResourceIDs[id] = true
	}

	configTags := cfg.Exclude.ParseTags()
	for k, v := range configTags {
		exc.Tags[k] = v
	}

	for _, tag := range scanFlags.excludeTags {
		if idx := indexByte(tag, '='); idx >= 0 {
			exc.Tags[tag[:idx]] = tag[idx+1:]
		} else {
			exc.Tags[tag] = ""
		}
	}

	return exc
}

func indexByte(s string, c byte) int {
	for i := range s {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func selectReporter(format, outputFile string) (report.Reporter, func(), error) {
	var w io.Writer = os.Stdout
	var cleanup func()

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return nil, nil, fmt.Errorf("create output file: %w", err)
		}
		w = f
		cleanup = func() { _ = f.Close() }
	}

	switch format {
	case "json":
		return &report.JSONReporter{Writer: w}, cleanup, nil
	case "sarif":
		return &report.SARIFReporter{Writer: w}, cleanup, nil
	case "spectrehub":
		return &report.SpectreHubReporter{Writer: w}, cleanup, nil
	default:
		return &report.TextReporter{Writer: w}, cleanup, nil
	}
}

func computeTargetHash(subscription string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("subscription:%s", subscription)))
	return fmt.Sprintf("sha256:%x", h)
}
