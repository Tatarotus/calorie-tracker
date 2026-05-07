package commands

import (
	"testing"
)

func TestAddCmd_HasCorrectArgs(t *testing.T) {
	if addCmd.Args == nil {
		t.Error("Expected addCmd to have Args validator")
	}
}

func TestWaterCmd_HasCorrectArgs(t *testing.T) {
	if waterCmd.Args == nil {
		t.Error("Expected waterCmd to have Args validator")
	}
}

func TestRootCmd_HasSubcommands(t *testing.T) {
	subcommands := rootCmd.Commands()
	if len(subcommands) < 4 {
		t.Errorf("Expected at least 4 subcommands, got %d", len(subcommands))
	}
}

func TestReviewCmd_Short(t *testing.T) {
	if reviewCmd.Short != "Get an AI-powered review of your recent progress" {
		t.Errorf("Unexpected Short: %s", reviewCmd.Short)
	}
}

func TestReportCmd_Short(t *testing.T) {
	if reportCmd.Short != "View your daily stats" {
		t.Errorf("Unexpected Short: %s", reportCmd.Short)
	}
}

func TestWaterCmd_Short(t *testing.T) {
	if waterCmd.Short != "Add water intake in ml" {
		t.Errorf("Unexpected Short: %s", waterCmd.Short)
	}
}
