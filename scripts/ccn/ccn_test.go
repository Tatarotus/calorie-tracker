package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const maxComplexity = 20

func findGoFiles() []string {
	var files []string
	excludeDirs := map[string]bool{
		"commands": true,
		"db":       true,
	}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if excludeDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files
}

func TestCyclomaticComplexity(t *testing.T) {
	files := findGoFiles()
	if len(files) == 0 {
		t.Fatal("No Go files found to analyze")
	}

	// Find gocyclo in PATH
	gocycloPath, err := exec.LookPath("gocyclo")
	if err != nil {
		// Try GOBIN
		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			gobin = filepath.Join(os.Getenv("HOME"), "go", "bin")
		}
		gocycloPath = filepath.Join(gobin, "gocyclo")
	}

	// Run gocyclo
	cmd := exec.Command(gocycloPath, "-top", "1000")
	cmd.Args = append(cmd.Args, files...)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "executable file not found") ||
			strings.Contains(err.Error(), "executable file not found") ||
			strings.Contains(err.Error(), "no such file or directory") {
			t.Skipf("gocyclo not found, skipping complexity test")
		}
		t.Fatalf("gocyclo error: %v\n%s", err, output)
	}

	// Parse output
	lines := strings.Split(string(output), "\n")
	var failures []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		complexity, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if complexity > maxComplexity {
			failures = append(failures, line)
		}
	}

	if len(failures) > 0 {
		t.Errorf("The following functions exceed cyclomatic complexity limit of %d:\n%s",
			maxComplexity, strings.Join(failures, "\n"))
	}
}
