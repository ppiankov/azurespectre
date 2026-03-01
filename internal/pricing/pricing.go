package pricing

import (
	"encoding/json"
	"log/slog"
	"strings"
)

const hoursPerMonth = 730

var pricingDB map[string]map[string]map[string]float64

func init() {
	if err := json.Unmarshal(pricingData, &pricingDB); err != nil {
		slog.Warn("failed to parse embedded pricing data", "error", err)
		pricingDB = make(map[string]map[string]map[string]float64)
	}
}

func lookupHourly(resourceType, key, region string) (float64, bool) {
	types, ok := pricingDB[resourceType]
	if !ok {
		return 0, false
	}
	regions, ok := types[key]
	if !ok {
		return 0, false
	}
	price, ok := regions[region]
	if !ok {
		price, ok = regions["eastus"]
		if !ok {
			return 0, false
		}
	}
	return price, true
}

func lookupMonthly(resourceType, region string) (float64, bool) {
	types, ok := pricingDB[resourceType]
	if !ok {
		return 0, false
	}
	regions, ok := types["default"]
	if !ok {
		return 0, false
	}
	price, ok := regions[region]
	if !ok {
		price, ok = regions["eastus"]
		if !ok {
			return 0, false
		}
	}
	return price, true
}

// MonthlyVMCost returns estimated monthly cost for a VM size.
func MonthlyVMCost(vmSize, region string) float64 {
	hourly, ok := lookupHourly("virtual_machine", strings.ToLower(vmSize), region)
	if !ok {
		return 0
	}
	return hourly * hoursPerMonth
}

// MonthlyDiskCost returns estimated monthly cost for a managed disk.
func MonthlyDiskCost(sku string, sizeGB int, region string) float64 {
	perGB, ok := lookupHourly("managed_disk", strings.ToLower(sku), region)
	if !ok {
		return 0
	}
	return perGB * float64(sizeGB)
}

// MonthlyPublicIPCost returns estimated monthly cost for a static public IP.
func MonthlyPublicIPCost(region string) float64 {
	cost, _ := lookupMonthly("public_ip", region)
	return cost
}

// MonthlySnapshotCost returns estimated monthly cost for a disk snapshot.
func MonthlySnapshotCost(sizeGB int, region string) float64 {
	perGB, ok := lookupHourly("snapshot", "default", region)
	if !ok {
		return 0
	}
	return perGB * float64(sizeGB)
}

// MonthlyLBCost returns estimated monthly cost for a load balancer.
func MonthlyLBCost(region string) float64 {
	cost, _ := lookupMonthly("load_balancer", region)
	return cost
}

// MonthlySQLCost returns estimated monthly cost for a SQL database.
// For DTU tiers (Basic, Standard), uses flat monthly pricing.
// For vCore tiers (GeneralPurpose), uses hourly × 730 × capacity.
func MonthlySQLCost(tier string, capacity int32, region string) float64 {
	key := strings.ToLower(tier)
	switch key {
	case "generalpurpose", "businesscritical", "hyperscale":
		hourly, ok := lookupHourly("sql_database", key, region)
		if !ok {
			return 0
		}
		return hourly * hoursPerMonth * float64(capacity)
	default:
		cost, ok := lookupHourly("sql_database", key, region)
		if !ok {
			return 0
		}
		return cost
	}
}

// MonthlyAppServiceCost returns estimated monthly cost for an App Service app.
func MonthlyAppServiceCost(_ string, region string) float64 {
	cost, _ := lookupMonthly("app_service", region)
	return cost
}

// MonthlyStorageCost returns estimated monthly cost for a storage account.
func MonthlyStorageCost(_ string, region string) float64 {
	cost, _ := lookupMonthly("storage_account", region)
	return cost
}
