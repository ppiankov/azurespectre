package azure

import "time"

// Severity represents the impact level of a finding.
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// ResourceType identifies the Azure resource category.
type ResourceType string

const (
	ResourceVM           ResourceType = "virtual_machine"
	ResourceDisk         ResourceType = "managed_disk"
	ResourcePublicIP     ResourceType = "public_ip"
	ResourceSnapshot     ResourceType = "snapshot"
	ResourceNSG          ResourceType = "network_security_group"
	ResourceLoadBalancer ResourceType = "load_balancer"
	ResourceSQLDatabase  ResourceType = "sql_database"
	ResourceAppService   ResourceType = "app_service"
	ResourceStorage      ResourceType = "storage_account"
)

// FindingID uniquely identifies the type of waste or risk detected.
type FindingID string

const (
	FindingIdleVM         FindingID = "IDLE_VM"
	FindingStoppedVM      FindingID = "STOPPED_VM"
	FindingUnattachedDisk FindingID = "UNATTACHED_DISK"
	FindingUnusedIP       FindingID = "UNUSED_IP"
	FindingStaleSnapshot  FindingID = "STALE_SNAPSHOT"
	FindingUnusedNSG      FindingID = "UNUSED_NSG"
	FindingIdleLB         FindingID = "IDLE_LB"
	FindingIdleSQL        FindingID = "IDLE_SQL"
	FindingIdleAppService FindingID = "IDLE_APP_SERVICE"
	FindingUnusedStorage  FindingID = "UNUSED_STORAGE"
)

// Finding represents a single waste or risk detection.
type Finding struct {
	ID                    FindingID      `json:"id"`
	Severity              Severity       `json:"severity"`
	ResourceType          ResourceType   `json:"resource_type"`
	ResourceID            string         `json:"resource_id"`
	ResourceName          string         `json:"resource_name,omitempty"`
	Subscription          string         `json:"subscription"`
	Region                string         `json:"region,omitempty"`
	ResourceGroup         string         `json:"resource_group,omitempty"`
	Message               string         `json:"message"`
	EstimatedMonthlyWaste float64        `json:"estimated_monthly_waste"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

// ScanResult aggregates findings from a scan.
type ScanResult struct {
	Findings         []Finding `json:"findings"`
	Errors           []string  `json:"errors,omitempty"`
	ResourcesScanned int       `json:"resources_scanned"`
}

// ScanConfig controls scan behavior.
type ScanConfig struct {
	IdleDays       int
	StaleDays      int
	StoppedDays    int
	IdleCPU        float64
	MinMonthlyCost float64
	ResourceGroup  string
	Exclude        ExcludeConfig
}

// ExcludeConfig defines resource filtering rules.
type ExcludeConfig struct {
	ResourceIDs map[string]bool
	Tags        map[string]string
}

// ScanProgress reports scanner status.
type ScanProgress struct {
	Subscription string
	Scanner      string
	Message      string
	Timestamp    time.Time
}
