package azure

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"golang.org/x/sync/errgroup"
)

// AzureComputeClient wraps Azure Compute SDK clients.
type AzureComputeClient struct {
	vmClient       *armcompute.VirtualMachinesClient
	diskClient     *armcompute.DisksClient
	snapshotClient *armcompute.SnapshotsClient
}

// NewComputeClient creates a Compute API client for the given subscription.
func NewComputeClient(subscriptionID string) (*AzureComputeClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure credential: %w", err)
	}
	return NewComputeClientWithCred(subscriptionID, cred)
}

// NewComputeClientWithCred creates a Compute API client with an existing credential.
func NewComputeClientWithCred(subscriptionID string, cred azcore.TokenCredential) (*AzureComputeClient, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("vm client: %w", err)
	}
	diskClient, err := armcompute.NewDisksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("disk client: %w", err)
	}
	snapshotClient, err := armcompute.NewSnapshotsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("snapshot client: %w", err)
	}
	return &AzureComputeClient{
		vmClient:       vmClient,
		diskClient:     diskClient,
		snapshotClient: snapshotClient,
	}, nil
}

// ListVMs lists all VMs in the subscription with their power state.
func (c *AzureComputeClient) ListVMs(ctx context.Context, _ string) ([]VirtualMachine, error) {
	pager := c.vmClient.NewListAllPager(nil)
	var basicVMs []struct {
		rg   string
		name string
		vm   VirtualMachine
	}

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list VMs: %w", err)
		}
		for _, v := range page.Value {
			if v == nil || v.Properties == nil {
				continue
			}
			vm := convertVM(v)
			rg := vm.ResourceGroup
			basicVMs = append(basicVMs, struct {
				rg   string
				name string
				vm   VirtualMachine
			}{rg: rg, name: vm.Name, vm: vm})
		}
	}

	// Fetch instance view for each VM to get power state
	var mu sync.Mutex
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	result := make([]VirtualMachine, len(basicVMs))

	for i, bvm := range basicVMs {
		i, bvm := i, bvm
		g.Go(func() error {
			resp, err := c.vmClient.Get(ctx, bvm.rg, bvm.name, &armcompute.VirtualMachinesClientGetOptions{
				Expand: to.Ptr(armcompute.InstanceViewTypesInstanceView),
			})
			if err != nil {
				mu.Lock()
				result[i] = bvm.vm
				mu.Unlock()
				return nil
			}
			vm := bvm.vm
			vm.PowerState = extractPowerState(resp.VirtualMachine)
			mu.Lock()
			result[i] = vm
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// ListDisks lists all managed disks in the subscription.
func (c *AzureComputeClient) ListDisks(ctx context.Context, _ string) ([]ManagedDisk, error) {
	pager := c.diskClient.NewListPager(nil)
	var disks []ManagedDisk

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list disks: %w", err)
		}
		for _, d := range page.Value {
			if d == nil || d.Properties == nil {
				continue
			}
			disks = append(disks, convertDisk(d))
		}
	}
	return disks, nil
}

// ListSnapshots lists all disk snapshots in the subscription.
func (c *AzureComputeClient) ListSnapshots(ctx context.Context, _ string) ([]DiskSnapshot, error) {
	pager := c.snapshotClient.NewListPager(nil)
	var snapshots []DiskSnapshot

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list snapshots: %w", err)
		}
		for _, s := range page.Value {
			if s == nil || s.Properties == nil {
				continue
			}
			snapshots = append(snapshots, convertSnapshot(s))
		}
	}
	return snapshots, nil
}

// AzureNetworkClient wraps Azure Network SDK clients.
type AzureNetworkClient struct {
	ipClient  *armnetwork.PublicIPAddressesClient
	nsgClient *armnetwork.SecurityGroupsClient
	lbClient  *armnetwork.LoadBalancersClient
}

// NewNetworkClient creates a Network API client for the given subscription.
func NewNetworkClient(subscriptionID string) (*AzureNetworkClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure credential: %w", err)
	}
	return NewNetworkClientWithCred(subscriptionID, cred)
}

// NewNetworkClientWithCred creates a Network API client with an existing credential.
func NewNetworkClientWithCred(subscriptionID string, cred azcore.TokenCredential) (*AzureNetworkClient, error) {
	ipClient, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("public IP client: %w", err)
	}
	nsgClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("NSG client: %w", err)
	}
	lbClient, err := armnetwork.NewLoadBalancersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("load balancer client: %w", err)
	}
	return &AzureNetworkClient{
		ipClient:  ipClient,
		nsgClient: nsgClient,
		lbClient:  lbClient,
	}, nil
}

// ListPublicIPs lists all public IP addresses in the subscription.
func (c *AzureNetworkClient) ListPublicIPs(ctx context.Context, _ string) ([]PublicIPAddress, error) {
	pager := c.ipClient.NewListAllPager(nil)
	var ips []PublicIPAddress

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list public IPs: %w", err)
		}
		for _, ip := range page.Value {
			if ip == nil || ip.Properties == nil {
				continue
			}
			ips = append(ips, convertPublicIP(ip))
		}
	}
	return ips, nil
}

// ListNSGs lists all network security groups in the subscription.
func (c *AzureNetworkClient) ListNSGs(ctx context.Context, _ string) ([]NetworkSecurityGroup, error) {
	pager := c.nsgClient.NewListAllPager(nil)
	var nsgs []NetworkSecurityGroup

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list NSGs: %w", err)
		}
		for _, nsg := range page.Value {
			if nsg == nil || nsg.Properties == nil {
				continue
			}
			nsgs = append(nsgs, convertNSG(nsg))
		}
	}
	return nsgs, nil
}

// ListLoadBalancers lists all load balancers in the subscription.
func (c *AzureNetworkClient) ListLoadBalancers(ctx context.Context, _ string) ([]LoadBalancer, error) {
	pager := c.lbClient.NewListAllPager(nil)
	var lbs []LoadBalancer

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list load balancers: %w", err)
		}
		for _, lb := range page.Value {
			if lb == nil || lb.Properties == nil {
				continue
			}
			lbs = append(lbs, convertLoadBalancer(lb))
		}
	}
	return lbs, nil
}

// AzureSQLClient wraps Azure SQL SDK clients.
type AzureSQLClient struct {
	serverClient   *armsql.ServersClient
	databaseClient *armsql.DatabasesClient
}

// NewSQLClient creates a SQL API client for the given subscription.
func NewSQLClient(subscriptionID string) (*AzureSQLClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure credential: %w", err)
	}
	return NewSQLClientWithCred(subscriptionID, cred)
}

// NewSQLClientWithCred creates a SQL API client with an existing credential.
func NewSQLClientWithCred(subscriptionID string, cred azcore.TokenCredential) (*AzureSQLClient, error) {
	serverClient, err := armsql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("sql server client: %w", err)
	}
	databaseClient, err := armsql.NewDatabasesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("sql database client: %w", err)
	}
	return &AzureSQLClient{
		serverClient:   serverClient,
		databaseClient: databaseClient,
	}, nil
}

// ListSQLDatabases lists all SQL databases across all servers in the subscription.
func (c *AzureSQLClient) ListSQLDatabases(ctx context.Context, _ string) ([]SQLDatabase, error) {
	// Azure SQL requires server enumeration first
	serverPager := c.serverClient.NewListPager(nil)
	var servers []struct {
		name string
		rg   string
	}

	for serverPager.More() {
		page, err := serverPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list SQL servers: %w", err)
		}
		for _, s := range page.Value {
			if s == nil || s.Name == nil || s.ID == nil {
				continue
			}
			servers = append(servers, struct {
				name string
				rg   string
			}{name: *s.Name, rg: extractResourceGroup(*s.ID)})
		}
	}

	var mu sync.Mutex
	var databases []SQLDatabase
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for _, srv := range servers {
		srv := srv
		g.Go(func() error {
			dbPager := c.databaseClient.NewListByServerPager(srv.rg, srv.name, nil)
			for dbPager.More() {
				page, err := dbPager.NextPage(ctx)
				if err != nil {
					return nil // skip individual server failures
				}
				for _, db := range page.Value {
					if db == nil || db.Properties == nil {
						continue
					}
					// Skip system databases
					if db.Name != nil && *db.Name == "master" {
						continue
					}
					d := convertSQLDatabase(db, srv.name)
					mu.Lock()
					databases = append(databases, d)
					mu.Unlock()
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return databases, nil
}

// AzureAppServiceClient wraps Azure App Service SDK clients.
type AzureAppServiceClient struct {
	webAppsClient *armappservice.WebAppsClient
}

// NewAppServiceClient creates an App Service API client for the given subscription.
func NewAppServiceClient(subscriptionID string) (*AzureAppServiceClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure credential: %w", err)
	}
	return NewAppServiceClientWithCred(subscriptionID, cred)
}

// NewAppServiceClientWithCred creates an App Service API client with an existing credential.
func NewAppServiceClientWithCred(subscriptionID string, cred azcore.TokenCredential) (*AzureAppServiceClient, error) {
	webAppsClient, err := armappservice.NewWebAppsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("web apps client: %w", err)
	}
	return &AzureAppServiceClient{webAppsClient: webAppsClient}, nil
}

// ListAppServiceApps lists all App Service web apps in the subscription.
func (c *AzureAppServiceClient) ListAppServiceApps(ctx context.Context, _ string) ([]AppServiceApp, error) {
	pager := c.webAppsClient.NewListPager(nil)
	var apps []AppServiceApp

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list app service apps: %w", err)
		}
		for _, site := range page.Value {
			if site == nil {
				continue
			}
			apps = append(apps, convertAppServiceApp(site))
		}
	}
	return apps, nil
}

// AzureStorageClient wraps Azure Storage SDK clients.
type AzureStorageClient struct {
	accountsClient *armstorage.AccountsClient
}

// NewStorageClient creates a Storage API client for the given subscription.
func NewStorageClient(subscriptionID string) (*AzureStorageClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure credential: %w", err)
	}
	return NewStorageClientWithCred(subscriptionID, cred)
}

// NewStorageClientWithCred creates a Storage API client with an existing credential.
func NewStorageClientWithCred(subscriptionID string, cred azcore.TokenCredential) (*AzureStorageClient, error) {
	accountsClient, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("storage accounts client: %w", err)
	}
	return &AzureStorageClient{accountsClient: accountsClient}, nil
}

// ListStorageAccounts lists all storage accounts in the subscription.
func (c *AzureStorageClient) ListStorageAccounts(ctx context.Context, _ string) ([]StorageAccount, error) {
	pager := c.accountsClient.NewListPager(nil)
	var accounts []StorageAccount

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list storage accounts: %w", err)
		}
		for _, acct := range page.Value {
			if acct == nil {
				continue
			}
			accounts = append(accounts, convertStorageAccount(acct))
		}
	}
	return accounts, nil
}

// AzureMonitorClient wraps the Azure Monitor metrics client.
type AzureMonitorClient struct {
	client *azquery.MetricsClient
}

// NewMonitorClient creates a Monitor API client.
func NewMonitorClient() (*AzureMonitorClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure credential: %w", err)
	}
	return NewMonitorClientWithCred(cred)
}

// NewMonitorClientWithCred creates a Monitor API client with an existing credential.
func NewMonitorClientWithCred(cred azcore.TokenCredential) (*AzureMonitorClient, error) {
	client, err := azquery.NewMetricsClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("metrics client: %w", err)
	}
	return &AzureMonitorClient{client: client}, nil
}

// FetchMetricMean queries Azure Monitor for the average of a metric across resources.
func (c *AzureMonitorClient) FetchMetricMean(ctx context.Context, resourceURIs []string, metricName string, lookbackDays int) (map[string]float64, error) {
	results := make(map[string]float64)
	var mu sync.Mutex
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for _, uri := range resourceURIs {
		uri := uri
		g.Go(func() error {
			ts := azquery.TimeInterval(buildTimespan(lookbackDays))
			avg := azquery.AggregationTypeAverage
			opts := &azquery.MetricsClientQueryResourceOptions{
				MetricNames: to.Ptr(metricName),
				Timespan:    &ts,
				Aggregation: []*azquery.AggregationType{&avg},
				Interval:    to.Ptr("PT1H"),
			}
			resp, err := c.client.QueryResource(ctx, uri, opts)
			if err != nil {
				return nil // skip individual failures
			}
			mean, ok := extractMean(resp)
			if ok {
				mu.Lock()
				results[uri] = mean
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// Conversion helpers

func convertVM(v *armcompute.VirtualMachine) VirtualMachine {
	vm := VirtualMachine{
		Name: derefStr(v.Name),
		Tags: convertTags(v.Tags),
	}
	if v.ID != nil {
		vm.ID = *v.ID
		vm.ResourceGroup = extractResourceGroup(*v.ID)
	}
	if v.Location != nil {
		vm.Location = *v.Location
	}
	if v.Properties != nil && v.Properties.HardwareProfile != nil && v.Properties.HardwareProfile.VMSize != nil {
		vm.VMSize = string(*v.Properties.HardwareProfile.VMSize)
	}
	if v.Properties != nil && v.Properties.TimeCreated != nil {
		vm.TimeCreated = *v.Properties.TimeCreated
	}
	return vm
}

func convertDisk(d *armcompute.Disk) ManagedDisk {
	disk := ManagedDisk{
		Name: derefStr(d.Name),
		Tags: convertTags(d.Tags),
	}
	if d.ID != nil {
		disk.ID = *d.ID
		disk.ResourceGroup = extractResourceGroup(*d.ID)
	}
	if d.Location != nil {
		disk.Location = *d.Location
	}
	if d.SKU != nil && d.SKU.Name != nil {
		disk.SKU = string(*d.SKU.Name)
	}
	if d.Properties != nil {
		if d.Properties.DiskSizeGB != nil {
			disk.SizeGB = *d.Properties.DiskSizeGB
		}
		if d.Properties.DiskState != nil {
			disk.DiskState = string(*d.Properties.DiskState)
		}
		if d.Properties.TimeCreated != nil {
			disk.TimeCreated = *d.Properties.TimeCreated
		}
	}
	if d.ManagedBy != nil {
		disk.ManagedBy = *d.ManagedBy
	}
	return disk
}

func convertSnapshot(s *armcompute.Snapshot) DiskSnapshot {
	snap := DiskSnapshot{
		Name: derefStr(s.Name),
		Tags: convertTags(s.Tags),
	}
	if s.ID != nil {
		snap.ID = *s.ID
		snap.ResourceGroup = extractResourceGroup(*s.ID)
	}
	if s.Location != nil {
		snap.Location = *s.Location
	}
	if s.Properties != nil {
		if s.Properties.DiskSizeGB != nil {
			snap.DiskSizeGB = *s.Properties.DiskSizeGB
		}
		if s.Properties.CreationData != nil && s.Properties.CreationData.SourceResourceID != nil {
			snap.SourceDisk = *s.Properties.CreationData.SourceResourceID
		}
		if s.Properties.TimeCreated != nil {
			snap.TimeCreated = *s.Properties.TimeCreated
		}
	}
	return snap
}

func convertPublicIP(ip *armnetwork.PublicIPAddress) PublicIPAddress {
	pip := PublicIPAddress{
		Name: derefStr(ip.Name),
		Tags: convertTags(ip.Tags),
	}
	if ip.ID != nil {
		pip.ID = *ip.ID
		pip.ResourceGroup = extractResourceGroup(*ip.ID)
	}
	if ip.Location != nil {
		pip.Location = *ip.Location
	}
	if ip.Properties != nil {
		if ip.Properties.IPAddress != nil {
			pip.IPAddress = *ip.Properties.IPAddress
		}
		if ip.Properties.PublicIPAllocationMethod != nil {
			pip.AllocationMethod = string(*ip.Properties.PublicIPAllocationMethod)
		}
		if ip.Properties.IPConfiguration != nil && ip.Properties.IPConfiguration.ID != nil {
			pip.AssociatedResource = *ip.Properties.IPConfiguration.ID
		}
	}
	return pip
}

func convertNSG(nsg *armnetwork.SecurityGroup) NetworkSecurityGroup {
	n := NetworkSecurityGroup{
		Name: derefStr(nsg.Name),
		Tags: convertTags(nsg.Tags),
	}
	if nsg.ID != nil {
		n.ID = *nsg.ID
		n.ResourceGroup = extractResourceGroup(*nsg.ID)
	}
	if nsg.Location != nil {
		n.Location = *nsg.Location
	}
	if nsg.Properties != nil {
		for _, s := range nsg.Properties.Subnets {
			if s != nil && s.ID != nil {
				n.Subnets = append(n.Subnets, *s.ID)
			}
		}
		for _, nic := range nsg.Properties.NetworkInterfaces {
			if nic != nil && nic.ID != nil {
				n.NICs = append(n.NICs, *nic.ID)
			}
		}
	}
	return n
}

func extractPowerState(vm armcompute.VirtualMachine) string {
	if vm.Properties == nil || vm.Properties.InstanceView == nil {
		return ""
	}
	for _, s := range vm.Properties.InstanceView.Statuses {
		if s == nil || s.Code == nil {
			continue
		}
		if strings.HasPrefix(*s.Code, "PowerState/") {
			return *s.Code
		}
	}
	return ""
}

func extractResourceGroup(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, p := range parts {
		if strings.EqualFold(p, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func convertTags(tags map[string]*string) map[string]string {
	if tags == nil {
		return nil
	}
	result := make(map[string]string, len(tags))
	for k, v := range tags {
		if v != nil {
			result[k] = *v
		}
	}
	return result
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func buildTimespan(lookbackDays int) string {
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -lookbackDays)
	return fmt.Sprintf("%s/%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
}

func extractMean(resp azquery.MetricsClientQueryResourceResponse) (float64, bool) {
	for _, metric := range resp.Value {
		if metric == nil {
			continue
		}
		for _, ts := range metric.TimeSeries {
			if ts == nil {
				continue
			}
			var sum float64
			var count int
			for _, dp := range ts.Data {
				if dp == nil || dp.Average == nil {
					continue
				}
				sum += *dp.Average
				count++
			}
			if count > 0 {
				return sum / float64(count), true
			}
		}
	}
	return 0, false
}

func convertLoadBalancer(lb *armnetwork.LoadBalancer) LoadBalancer {
	l := LoadBalancer{
		Name: derefStr(lb.Name),
		Tags: convertTags(lb.Tags),
	}
	if lb.ID != nil {
		l.ID = *lb.ID
		l.ResourceGroup = extractResourceGroup(*lb.ID)
	}
	if lb.Location != nil {
		l.Location = *lb.Location
	}
	if lb.SKU != nil && lb.SKU.Name != nil {
		l.SKU = string(*lb.SKU.Name)
	}
	if lb.Properties != nil {
		l.BackendPoolCount = len(lb.Properties.BackendAddressPools)
		l.RuleCount = len(lb.Properties.LoadBalancingRules)
	}
	return l
}

func convertSQLDatabase(db *armsql.Database, serverName string) SQLDatabase {
	d := SQLDatabase{
		Name:       derefStr(db.Name),
		ServerName: serverName,
		Tags:       convertTags(db.Tags),
	}
	if db.ID != nil {
		d.ID = *db.ID
		d.ResourceGroup = extractResourceGroup(*db.ID)
	}
	if db.Location != nil {
		d.Location = *db.Location
	}
	if db.SKU != nil {
		d.SKUName = derefStr(db.SKU.Name)
		d.SKUTier = derefStr(db.SKU.Tier)
		if db.SKU.Capacity != nil {
			d.Capacity = *db.SKU.Capacity
		}
	}
	return d
}

func convertAppServiceApp(site *armappservice.Site) AppServiceApp {
	app := AppServiceApp{
		Name: derefStr(site.Name),
		Kind: derefStr(site.Kind),
		Tags: convertTags(site.Tags),
	}
	if site.ID != nil {
		app.ID = *site.ID
		app.ResourceGroup = extractResourceGroup(*site.ID)
	}
	if site.Location != nil {
		app.Location = *site.Location
	}
	if site.Properties != nil {
		if site.Properties.State != nil {
			app.State = *site.Properties.State
		}
		if site.Properties.ServerFarmID != nil {
			app.AppPlanID = *site.Properties.ServerFarmID
		}
	}
	return app
}

func convertStorageAccount(acct *armstorage.Account) StorageAccount {
	sa := StorageAccount{
		Name: derefStr(acct.Name),
		Tags: convertTags(acct.Tags),
	}
	if acct.ID != nil {
		sa.ID = *acct.ID
		sa.ResourceGroup = extractResourceGroup(*acct.ID)
	}
	if acct.Location != nil {
		sa.Location = *acct.Location
	}
	if acct.SKU != nil && acct.SKU.Name != nil {
		sa.SKU = string(*acct.SKU.Name)
	}
	if acct.Kind != nil {
		sa.Kind = string(*acct.Kind)
	}
	return sa
}
