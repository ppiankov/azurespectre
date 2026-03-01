package azure

import "context"

type mockComputeAPI struct {
	vms       []VirtualMachine
	disks     []ManagedDisk
	snapshots []DiskSnapshot
	err       error
}

func (m *mockComputeAPI) ListVMs(_ context.Context, _ string) ([]VirtualMachine, error) {
	return m.vms, m.err
}

func (m *mockComputeAPI) ListDisks(_ context.Context, _ string) ([]ManagedDisk, error) {
	return m.disks, m.err
}

func (m *mockComputeAPI) ListSnapshots(_ context.Context, _ string) ([]DiskSnapshot, error) {
	return m.snapshots, m.err
}

type mockNetworkAPI struct {
	ips  []PublicIPAddress
	nsgs []NetworkSecurityGroup
	lbs  []LoadBalancer
	err  error
}

func (m *mockNetworkAPI) ListPublicIPs(_ context.Context, _ string) ([]PublicIPAddress, error) {
	return m.ips, m.err
}

func (m *mockNetworkAPI) ListNSGs(_ context.Context, _ string) ([]NetworkSecurityGroup, error) {
	return m.nsgs, m.err
}

func (m *mockNetworkAPI) ListLoadBalancers(_ context.Context, _ string) ([]LoadBalancer, error) {
	return m.lbs, m.err
}

type mockMonitorAPI struct {
	results map[string]float64
	err     error
}

func (m *mockMonitorAPI) FetchMetricMean(_ context.Context, _ []string, _ string, _ int) (map[string]float64, error) {
	return m.results, m.err
}

type mockSQLAPI struct {
	databases []SQLDatabase
	err       error
}

func (m *mockSQLAPI) ListSQLDatabases(_ context.Context, _ string) ([]SQLDatabase, error) {
	return m.databases, m.err
}

type mockAppServiceAPI struct {
	apps []AppServiceApp
	err  error
}

func (m *mockAppServiceAPI) ListAppServiceApps(_ context.Context, _ string) ([]AppServiceApp, error) {
	return m.apps, m.err
}

type mockStorageAPI struct {
	accounts []StorageAccount
	err      error
}

func (m *mockStorageAPI) ListStorageAccounts(_ context.Context, _ string) ([]StorageAccount, error) {
	return m.accounts, m.err
}
