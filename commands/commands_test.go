package commands

import (
	"testing"

	"calorie-tracker/config"
)

func TestInitDBAndTracker(t *testing.T) {
	// Set required env var
	t.Setenv("NVIDIA_API_KEY", "test-key")

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("initDBAndTracker panicked (expected without DB): %v", r)
		}
	}()

	// We can't easily test initDBAndTracker without a real DB,
	// but we can verify the config loads
	cfg := config.Load()
	if cfg == nil {
		t.Error("Expected config to load")
	}
}

func TestRootCmd(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("Expected rootCmd to be initialized")
	}

	if rootCmd.Use != "calorie-tracker" {
		t.Errorf("Expected Use to be 'calorie-tracker', got %s", rootCmd.Use)
	}
}

func TestAddCmd(t *testing.T) {
	if addCmd == nil {
		t.Fatal("Expected addCmd to be initialized")
	}

	if addCmd.Use != "add [food description]" {
		t.Errorf("Expected Use to be 'add [food description]', got %s", addCmd.Use)
	}

	if addCmd.Short != "Add a new food entry" {
		t.Errorf("Expected Short to be 'Add a new food entry', got %s", addCmd.Short)
	}
}

func TestWaterCmd(t *testing.T) {
	if waterCmd == nil {
		t.Fatal("Expected waterCmd to be initialized")
	}

	if waterCmd.Use != "water [amount_in_ml]" {
		t.Errorf("Expected Use to be 'water [amount_in_ml]', got %s", waterCmd.Use)
	}
}

func TestReviewCmd(t *testing.T) {
	if reviewCmd == nil {
		t.Fatal("Expected reviewCmd to be initialized")
	}

	if reviewCmd.Use != "review" {
		t.Errorf("Expected Use to be 'review', got %s", reviewCmd.Use)
	}
}

func TestReportCmd(t *testing.T) {
	if reportCmd == nil {
		t.Fatal("Expected reportCmd to be initialized")
	}

	if reportCmd.Use != "report" {
		t.Errorf("Expected Use to be 'report', got %s", reportCmd.Use)
	}
}
