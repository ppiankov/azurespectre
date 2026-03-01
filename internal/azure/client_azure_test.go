package azure

import (
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

func TestConvertVM(t *testing.T) {
	created := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	vmSize := armcompute.VirtualMachineSizeTypesStandardB2S
	v := &armcompute.VirtualMachine{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"),
		Name:     to.Ptr("vm1"),
		Location: to.Ptr("eastus"),
		Tags:     map[string]*string{"env": to.Ptr("dev")},
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{VMSize: &vmSize},
			TimeCreated:     &created,
		},
	}

	vm := convertVM(v)
	if vm.Name != "vm1" {
		t.Errorf("Name = %q, want %q", vm.Name, "vm1")
	}
	if vm.ResourceGroup != "rg1" {
		t.Errorf("ResourceGroup = %q, want %q", vm.ResourceGroup, "rg1")
	}
	if vm.Location != "eastus" {
		t.Errorf("Location = %q, want %q", vm.Location, "eastus")
	}
	if vm.VMSize != "Standard_B2s" {
		t.Errorf("VMSize = %q, want %q", vm.VMSize, "Standard_B2s")
	}
	if vm.Tags["env"] != "dev" {
		t.Error("expected tag env=dev")
	}
	if vm.TimeCreated != created {
		t.Error("TimeCreated mismatch")
	}
}

func TestConvertVM_NilFields(t *testing.T) {
	v := &armcompute.VirtualMachine{
		Properties: &armcompute.VirtualMachineProperties{},
	}
	vm := convertVM(v)
	if vm.Name != "" {
		t.Errorf("Name = %q, want empty", vm.Name)
	}
	if vm.VMSize != "" {
		t.Errorf("VMSize = %q, want empty", vm.VMSize)
	}
}

func TestConvertDisk(t *testing.T) {
	sku := armcompute.DiskStorageAccountTypesPremiumLRS
	d := &armcompute.Disk{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/disks/disk1"),
		Name:     to.Ptr("disk1"),
		Location: to.Ptr("westeurope"),
		SKU:      &armcompute.DiskSKU{Name: &sku},
		Properties: &armcompute.DiskProperties{
			DiskSizeGB: to.Ptr[int32](128),
			DiskState:  to.Ptr(armcompute.DiskStateUnattached),
		},
		ManagedBy: to.Ptr("/subscriptions/sub-1/..."),
	}

	disk := convertDisk(d)
	if disk.Name != "disk1" {
		t.Errorf("Name = %q, want %q", disk.Name, "disk1")
	}
	if disk.SKU != "Premium_LRS" {
		t.Errorf("SKU = %q, want %q", disk.SKU, "Premium_LRS")
	}
	if disk.SizeGB != 128 {
		t.Errorf("SizeGB = %d, want 128", disk.SizeGB)
	}
	if disk.DiskState != "Unattached" {
		t.Errorf("DiskState = %q, want %q", disk.DiskState, "Unattached")
	}
	if disk.ManagedBy == "" {
		t.Error("ManagedBy should be set")
	}
}

func TestConvertSnapshot(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &armcompute.Snapshot{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/snapshots/snap1"),
		Name:     to.Ptr("snap1"),
		Location: to.Ptr("eastus"),
		Properties: &armcompute.SnapshotProperties{
			DiskSizeGB: to.Ptr[int32](64),
			CreationData: &armcompute.CreationData{
				SourceResourceID: to.Ptr("/subscriptions/sub-1/.../disks/disk1"),
			},
			TimeCreated: &created,
		},
	}

	snap := convertSnapshot(s)
	if snap.Name != "snap1" {
		t.Errorf("Name = %q, want %q", snap.Name, "snap1")
	}
	if snap.DiskSizeGB != 64 {
		t.Errorf("DiskSizeGB = %d, want 64", snap.DiskSizeGB)
	}
	if snap.SourceDisk == "" {
		t.Error("SourceDisk should be set")
	}
}

func TestConvertPublicIP(t *testing.T) {
	alloc := armnetwork.IPAllocationMethodStatic
	ip := &armnetwork.PublicIPAddress{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/publicIPAddresses/ip1"),
		Name:     to.Ptr("ip1"),
		Location: to.Ptr("eastus"),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			IPAddress:                to.Ptr("20.1.2.3"),
			PublicIPAllocationMethod: &alloc,
			IPConfiguration:          &armnetwork.IPConfiguration{ID: to.Ptr("/nic-config-id")},
		},
	}

	pip := convertPublicIP(ip)
	if pip.Name != "ip1" {
		t.Errorf("Name = %q, want %q", pip.Name, "ip1")
	}
	if pip.IPAddress != "20.1.2.3" {
		t.Errorf("IPAddress = %q, want %q", pip.IPAddress, "20.1.2.3")
	}
	if pip.AllocationMethod != "Static" {
		t.Errorf("AllocationMethod = %q, want %q", pip.AllocationMethod, "Static")
	}
	if pip.AssociatedResource == "" {
		t.Error("AssociatedResource should be set")
	}
}

func TestConvertNSG(t *testing.T) {
	nsg := &armnetwork.SecurityGroup{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/networkSecurityGroups/nsg1"),
		Name:     to.Ptr("nsg1"),
		Location: to.Ptr("eastus"),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			Subnets: []*armnetwork.Subnet{
				{ID: to.Ptr("/subnet-1")},
			},
			NetworkInterfaces: []*armnetwork.Interface{
				{ID: to.Ptr("/nic-1")},
				{ID: to.Ptr("/nic-2")},
			},
		},
	}

	n := convertNSG(nsg)
	if n.Name != "nsg1" {
		t.Errorf("Name = %q, want %q", n.Name, "nsg1")
	}
	if len(n.Subnets) != 1 {
		t.Errorf("Subnets = %d, want 1", len(n.Subnets))
	}
	if len(n.NICs) != 2 {
		t.Errorf("NICs = %d, want 2", len(n.NICs))
	}
}

func TestConvertLoadBalancer(t *testing.T) {
	skuName := armnetwork.LoadBalancerSKUNameStandard
	lb := &armnetwork.LoadBalancer{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/loadBalancers/lb1"),
		Name:     to.Ptr("lb1"),
		Location: to.Ptr("eastus"),
		SKU:      &armnetwork.LoadBalancerSKU{Name: &skuName},
		Properties: &armnetwork.LoadBalancerPropertiesFormat{
			BackendAddressPools: []*armnetwork.BackendAddressPool{{}, {}},
			LoadBalancingRules:  []*armnetwork.LoadBalancingRule{{}, {}, {}},
		},
	}

	l := convertLoadBalancer(lb)
	if l.Name != "lb1" {
		t.Errorf("Name = %q, want %q", l.Name, "lb1")
	}
	if l.SKU != "Standard" {
		t.Errorf("SKU = %q, want %q", l.SKU, "Standard")
	}
	if l.BackendPoolCount != 2 {
		t.Errorf("BackendPoolCount = %d, want 2", l.BackendPoolCount)
	}
	if l.RuleCount != 3 {
		t.Errorf("RuleCount = %d, want 3", l.RuleCount)
	}
}

func TestConvertSQLDatabase(t *testing.T) {
	db := &armsql.Database{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Sql/servers/srv1/databases/db1"),
		Name:     to.Ptr("db1"),
		Location: to.Ptr("eastus"),
		Tags:     map[string]*string{"team": to.Ptr("platform")},
		SKU: &armsql.SKU{
			Name:     to.Ptr("S0"),
			Tier:     to.Ptr("Standard"),
			Capacity: to.Ptr[int32](10),
		},
		Properties: &armsql.DatabaseProperties{},
	}

	d := convertSQLDatabase(db, "srv1")
	if d.Name != "db1" {
		t.Errorf("Name = %q, want %q", d.Name, "db1")
	}
	if d.ServerName != "srv1" {
		t.Errorf("ServerName = %q, want %q", d.ServerName, "srv1")
	}
	if d.SKUName != "S0" {
		t.Errorf("SKUName = %q, want %q", d.SKUName, "S0")
	}
	if d.Capacity != 10 {
		t.Errorf("Capacity = %d, want 10", d.Capacity)
	}
}

func TestConvertAppServiceApp(t *testing.T) {
	site := &armappservice.Site{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Web/sites/app1"),
		Name:     to.Ptr("app1"),
		Location: to.Ptr("eastus"),
		Kind:     to.Ptr("app"),
		Properties: &armappservice.SiteProperties{
			State:        to.Ptr("Running"),
			ServerFarmID: to.Ptr("/plan-id"),
		},
	}

	app := convertAppServiceApp(site)
	if app.Name != "app1" {
		t.Errorf("Name = %q, want %q", app.Name, "app1")
	}
	if app.Kind != "app" {
		t.Errorf("Kind = %q, want %q", app.Kind, "app")
	}
	if app.State != "Running" {
		t.Errorf("State = %q, want %q", app.State, "Running")
	}
	if app.AppPlanID == "" {
		t.Error("AppPlanID should be set")
	}
}

func TestConvertStorageAccount(t *testing.T) {
	skuName := armstorage.SKUNameStandardLRS
	kind := armstorage.KindStorageV2
	acct := &armstorage.Account{
		ID:       to.Ptr("/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/sa1"),
		Name:     to.Ptr("sa1"),
		Location: to.Ptr("eastus"),
		SKU:      &armstorage.SKU{Name: &skuName},
		Kind:     &kind,
	}

	sa := convertStorageAccount(acct)
	if sa.Name != "sa1" {
		t.Errorf("Name = %q, want %q", sa.Name, "sa1")
	}
	if sa.SKU != "Standard_LRS" {
		t.Errorf("SKU = %q, want %q", sa.SKU, "Standard_LRS")
	}
	if sa.Kind != "StorageV2" {
		t.Errorf("Kind = %q, want %q", sa.Kind, "StorageV2")
	}
}

func TestExtractPowerState(t *testing.T) {
	tests := []struct {
		name string
		vm   armcompute.VirtualMachine
		want string
	}{
		{
			name: "deallocated",
			vm: armcompute.VirtualMachine{
				Properties: &armcompute.VirtualMachineProperties{
					InstanceView: &armcompute.VirtualMachineInstanceView{
						Statuses: []*armcompute.InstanceViewStatus{
							{Code: to.Ptr("ProvisioningState/succeeded")},
							{Code: to.Ptr("PowerState/deallocated")},
						},
					},
				},
			},
			want: "PowerState/deallocated",
		},
		{
			name: "running",
			vm: armcompute.VirtualMachine{
				Properties: &armcompute.VirtualMachineProperties{
					InstanceView: &armcompute.VirtualMachineInstanceView{
						Statuses: []*armcompute.InstanceViewStatus{
							{Code: to.Ptr("PowerState/running")},
						},
					},
				},
			},
			want: "PowerState/running",
		},
		{
			name: "nil properties",
			vm:   armcompute.VirtualMachine{},
			want: "",
		},
		{
			name: "nil instance view",
			vm: armcompute.VirtualMachine{
				Properties: &armcompute.VirtualMachineProperties{},
			},
			want: "",
		},
		{
			name: "nil status code",
			vm: armcompute.VirtualMachine{
				Properties: &armcompute.VirtualMachineProperties{
					InstanceView: &armcompute.VirtualMachineInstanceView{
						Statuses: []*armcompute.InstanceViewStatus{nil, {}},
					},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPowerState(tt.vm)
			if got != tt.want {
				t.Errorf("extractPowerState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractResourceGroup(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1", "rg1"},
		{"/subscriptions/sub-1/resourcegroups/RG-Upper/providers/foo", "RG-Upper"},
		{"no-resource-group-here", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := extractResourceGroup(tt.id)
		if got != tt.want {
			t.Errorf("extractResourceGroup(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestConvertTags(t *testing.T) {
	tags := map[string]*string{
		"env":  to.Ptr("dev"),
		"team": to.Ptr("platform"),
	}
	result := convertTags(tags)
	if result["env"] != "dev" {
		t.Errorf("env = %q, want %q", result["env"], "dev")
	}
	if result["team"] != "platform" {
		t.Errorf("team = %q, want %q", result["team"], "platform")
	}
}

func TestConvertTags_Nil(t *testing.T) {
	result := convertTags(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestDerefStr(t *testing.T) {
	s := "hello"
	if derefStr(&s) != "hello" {
		t.Error("expected hello")
	}
	if derefStr(nil) != "" {
		t.Error("expected empty for nil")
	}
}

func TestBuildTimespan(t *testing.T) {
	ts := buildTimespan(7)
	if len(ts) == 0 {
		t.Error("expected non-empty timespan")
	}
	// Should contain a / separator (ISO 8601 interval)
	found := false
	for i := range ts {
		if ts[i] == '/' {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("timespan %q should contain /", ts)
	}
}

func TestExtractMean(t *testing.T) {
	avg1 := 10.0
	avg2 := 20.0
	resp := azquery.MetricsClientQueryResourceResponse{
		Response: azquery.Response{
			Value: []*azquery.Metric{
				{
					TimeSeries: []*azquery.TimeSeriesElement{
						{
							Data: []*azquery.MetricValue{
								{Average: &avg1},
								{Average: &avg2},
							},
						},
					},
				},
			},
		},
	}

	mean, ok := extractMean(resp)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if mean != 15.0 {
		t.Errorf("mean = %.1f, want 15.0", mean)
	}
}

func TestExtractMean_Empty(t *testing.T) {
	resp := azquery.MetricsClientQueryResourceResponse{
		Response: azquery.Response{},
	}
	_, ok := extractMean(resp)
	if ok {
		t.Error("expected ok=false for empty response")
	}
}

func TestExtractMean_NilDataPoints(t *testing.T) {
	resp := azquery.MetricsClientQueryResourceResponse{
		Response: azquery.Response{
			Value: []*azquery.Metric{
				{
					TimeSeries: []*azquery.TimeSeriesElement{
						{
							Data: []*azquery.MetricValue{nil, {}},
						},
					},
				},
			},
		},
	}
	_, ok := extractMean(resp)
	if ok {
		t.Error("expected ok=false for nil data points")
	}
}

func TestExtractMean_NilMetric(t *testing.T) {
	resp := azquery.MetricsClientQueryResourceResponse{
		Response: azquery.Response{
			Value: []*azquery.Metric{nil},
		},
	}
	_, ok := extractMean(resp)
	if ok {
		t.Error("expected ok=false for nil metric")
	}
}

func TestExtractMean_NilTimeSeries(t *testing.T) {
	resp := azquery.MetricsClientQueryResourceResponse{
		Response: azquery.Response{
			Value: []*azquery.Metric{
				{TimeSeries: []*azquery.TimeSeriesElement{nil}},
			},
		},
	}
	_, ok := extractMean(resp)
	if ok {
		t.Error("expected ok=false for nil time series")
	}
}
