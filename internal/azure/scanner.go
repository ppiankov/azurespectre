package azure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// ResourceScanner defines the interface for resource-specific scanners.
type ResourceScanner interface {
	Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error)
	Type() ResourceType
}

// SubscriptionScanner orchestrates scanning across resource types.
type SubscriptionScanner struct {
	compute      ComputeAPI
	network      NetworkAPI
	monitor      MonitorAPI
	sqlAPI       SQLAPI
	appService   AppServiceAPI
	storageAPI   StorageAPI
	subscription string
	scanConfig   ScanConfig
	progressFn   func(ScanProgress)
}

// NewSubscriptionScanner creates a scanner for the given subscription.
func NewSubscriptionScanner(compute ComputeAPI, network NetworkAPI, monitor MonitorAPI, subscription string, scanCfg ScanConfig) *SubscriptionScanner {
	return &SubscriptionScanner{
		compute:      compute,
		network:      network,
		monitor:      monitor,
		subscription: subscription,
		scanConfig:   scanCfg,
	}
}

// SetSQLAPI sets the SQL API client for SQL database scanning.
func (s *SubscriptionScanner) SetSQLAPI(api SQLAPI) { s.sqlAPI = api }

// SetAppServiceAPI sets the App Service API client for app scanning.
func (s *SubscriptionScanner) SetAppServiceAPI(api AppServiceAPI) { s.appService = api }

// SetStorageAPI sets the Storage API client for storage account scanning.
func (s *SubscriptionScanner) SetStorageAPI(api StorageAPI) { s.storageAPI = api }

// SetProgressFn sets the progress callback.
func (s *SubscriptionScanner) SetProgressFn(fn func(ScanProgress)) {
	s.progressFn = fn
}

// ScanAll runs all resource scanners concurrently and merges results.
func (s *SubscriptionScanner) ScanAll(ctx context.Context) (*ScanResult, error) {
	scanners := s.buildScanners()

	var mu sync.Mutex
	var result ScanResult

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for _, scanner := range scanners {
		scanner := scanner
		g.Go(func() error {
			if s.progressFn != nil {
				mu.Lock()
				s.progressFn(ScanProgress{
					Subscription: s.subscription,
					Scanner:      string(scanner.Type()),
					Message:      fmt.Sprintf("scanning %s", scanner.Type()),
					Timestamp:    time.Now(),
				})
				mu.Unlock()
			}
			sr, err := scanner.Scan(ctx, s.scanConfig)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s/%s: %v", s.subscription, scanner.Type(), err))
				mu.Unlock()
				return nil
			}
			mu.Lock()
			result.Findings = append(result.Findings, sr.Findings...)
			result.ResourcesScanned += sr.ResourcesScanned
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *SubscriptionScanner) buildScanners() []ResourceScanner {
	scanners := []ResourceScanner{
		NewVMScanner(s.compute, s.monitor, s.subscription),
		NewDiskScanner(s.compute, s.subscription),
		NewIPScanner(s.network, s.subscription),
		NewSnapshotScanner(s.compute, s.subscription),
		NewNSGScanner(s.network, s.subscription),
		NewLBScanner(s.network, s.subscription),
	}
	if s.sqlAPI != nil {
		scanners = append(scanners, NewSQLScanner(s.sqlAPI, s.monitor, s.subscription))
	}
	if s.appService != nil {
		scanners = append(scanners, NewAppServiceScanner(s.appService, s.monitor, s.subscription))
	}
	if s.storageAPI != nil {
		scanners = append(scanners, NewStorageScanner(s.storageAPI, s.monitor, s.subscription))
	}
	return scanners
}
