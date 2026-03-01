package azure

import (
	"context"
	"time"
)

// VirtualMachine represents an Azure VM.
type VirtualMachine struct {
	ID            string
	Name          string
	ResourceGroup string
	Location      string
	VMSize        string
	PowerState    string
	Tags          map[string]string
	TimeCreated   time.Time
}

// ManagedDisk represents an Azure managed disk.
type ManagedDisk struct {
	ID            string
	Name          string
	ResourceGroup string
	Location      string
	SKU           string
	SizeGB        int32
	DiskState     string
	ManagedBy     string
	Tags          map[string]string
	TimeCreated   time.Time
}

// PublicIPAddress represents an Azure public IP.
type PublicIPAddress struct {
	ID                 string
	Name               string
	ResourceGroup      string
	Location           string
	IPAddress          string
	AllocationMethod   string
	AssociatedResource string
	Tags               map[string]string
}

// DiskSnapshot represents an Azure disk snapshot.
type DiskSnapshot struct {
	ID            string
	Name          string
	ResourceGroup string
	Location      string
	DiskSizeGB    int32
	SourceDisk    string
	Tags          map[string]string
	TimeCreated   time.Time
}

// NetworkSecurityGroup represents an Azure NSG.
type NetworkSecurityGroup struct {
	ID            string
	Name          string
	ResourceGroup string
	Location      string
	Subnets       []string
	NICs          []string
	Tags          map[string]string
}

// LoadBalancer represents an Azure load balancer.
type LoadBalancer struct {
	ID               string
	Name             string
	ResourceGroup    string
	Location         string
	SKU              string
	BackendPoolCount int
	RuleCount        int
	Tags             map[string]string
}

// SQLDatabase represents an Azure SQL database.
type SQLDatabase struct {
	ID            string
	Name          string
	ResourceGroup string
	Location      string
	ServerName    string
	SKUName       string
	SKUTier       string
	Capacity      int32
	Tags          map[string]string
}

// AppServiceApp represents an Azure App Service web app.
type AppServiceApp struct {
	ID            string
	Name          string
	ResourceGroup string
	Location      string
	Kind          string
	State         string
	AppPlanID     string
	Tags          map[string]string
}

// StorageAccount represents an Azure storage account.
type StorageAccount struct {
	ID            string
	Name          string
	ResourceGroup string
	Location      string
	SKU           string
	Kind          string
	Tags          map[string]string
}

// ComputeAPI abstracts Azure Compute operations.
type ComputeAPI interface {
	ListVMs(ctx context.Context, subscriptionID string) ([]VirtualMachine, error)
	ListDisks(ctx context.Context, subscriptionID string) ([]ManagedDisk, error)
	ListSnapshots(ctx context.Context, subscriptionID string) ([]DiskSnapshot, error)
}

// NetworkAPI abstracts Azure Network operations.
type NetworkAPI interface {
	ListPublicIPs(ctx context.Context, subscriptionID string) ([]PublicIPAddress, error)
	ListNSGs(ctx context.Context, subscriptionID string) ([]NetworkSecurityGroup, error)
	ListLoadBalancers(ctx context.Context, subscriptionID string) ([]LoadBalancer, error)
}

// MonitorAPI abstracts Azure Monitor metric queries.
type MonitorAPI interface {
	FetchMetricMean(ctx context.Context, resourceURIs []string, metricName string, lookbackDays int) (map[string]float64, error)
}

// SQLAPI abstracts Azure SQL operations.
type SQLAPI interface {
	ListSQLDatabases(ctx context.Context, subscriptionID string) ([]SQLDatabase, error)
}

// AppServiceAPI abstracts Azure App Service operations.
type AppServiceAPI interface {
	ListAppServiceApps(ctx context.Context, subscriptionID string) ([]AppServiceApp, error)
}

// StorageAPI abstracts Azure Storage operations.
type StorageAPI interface {
	ListStorageAccounts(ctx context.Context, subscriptionID string) ([]StorageAccount, error)
}
