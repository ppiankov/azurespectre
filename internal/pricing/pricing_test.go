package pricing

import (
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestMonthlyVMCost(t *testing.T) {
	cost := MonthlyVMCost("Standard_B2s", "eastus")
	want := 0.0416 * 730
	if !approxEqual(cost, want) {
		t.Errorf("MonthlyVMCost = %.2f, want %.2f", cost, want)
	}
}

func TestMonthlyVMCostCaseInsensitive(t *testing.T) {
	cost := MonthlyVMCost("STANDARD_B2S", "eastus")
	if cost <= 0 {
		t.Error("expected non-zero cost for uppercase VM size")
	}
}

func TestMonthlyDiskCost(t *testing.T) {
	cost := MonthlyDiskCost("Premium_LRS", 128, "eastus")
	want := 0.135 * 128
	if !approxEqual(cost, want) {
		t.Errorf("MonthlyDiskCost = %.2f, want %.2f", cost, want)
	}
}

func TestMonthlyPublicIPCost(t *testing.T) {
	cost := MonthlyPublicIPCost("eastus")
	if !approxEqual(cost, 3.65) {
		t.Errorf("MonthlyPublicIPCost = %.2f, want 3.65", cost)
	}
}

func TestMonthlySnapshotCost(t *testing.T) {
	cost := MonthlySnapshotCost(128, "eastus")
	want := 0.05 * 128
	if !approxEqual(cost, want) {
		t.Errorf("MonthlySnapshotCost = %.2f, want %.2f", cost, want)
	}
}

func TestUnknownVMSize(t *testing.T) {
	cost := MonthlyVMCost("Standard_XYZ_Unknown", "eastus")
	if cost != 0 {
		t.Errorf("unknown VM size should return 0, got %.2f", cost)
	}
}

func TestFallbackToEastUS(t *testing.T) {
	cost := MonthlyVMCost("Standard_B2s", "unknown-region-xyz")
	if cost <= 0 {
		t.Error("should fall back to eastus pricing")
	}
}

func TestMonthlyLBCost(t *testing.T) {
	cost := MonthlyLBCost("eastus")
	if !approxEqual(cost, 18.25) {
		t.Errorf("MonthlyLBCost = %.2f, want 18.25", cost)
	}
}

func TestMonthlySQLCost_DTU(t *testing.T) {
	cost := MonthlySQLCost("Basic", 5, "eastus")
	if !approxEqual(cost, 4.99) {
		t.Errorf("MonthlySQLCost(Basic) = %.2f, want 4.99", cost)
	}
}

func TestMonthlySQLCost_VCore(t *testing.T) {
	cost := MonthlySQLCost("GeneralPurpose", 2, "eastus")
	want := 0.2091 * 730 * 2
	if !approxEqual(cost, want) {
		t.Errorf("MonthlySQLCost(GP,2) = %.2f, want %.2f", cost, want)
	}
}

func TestMonthlySQLCost_Unknown(t *testing.T) {
	cost := MonthlySQLCost("UnknownTier", 1, "eastus")
	if cost != 0 {
		t.Errorf("unknown tier should return 0, got %.2f", cost)
	}
}

func TestMonthlyAppServiceCost(t *testing.T) {
	cost := MonthlyAppServiceCost("app", "eastus")
	if !approxEqual(cost, 54.75) {
		t.Errorf("MonthlyAppServiceCost = %.2f, want 54.75", cost)
	}
}

func TestMonthlyStorageCost(t *testing.T) {
	cost := MonthlyStorageCost("Standard_LRS", "eastus")
	if !approxEqual(cost, 5.00) {
		t.Errorf("MonthlyStorageCost = %.2f, want 5.00", cost)
	}
}

func TestMonthlyStorageCost_FallbackRegion(t *testing.T) {
	cost := MonthlyStorageCost("Standard_LRS", "unknown-region-xyz")
	if cost <= 0 {
		t.Error("should fall back to eastus pricing")
	}
}
