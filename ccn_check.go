//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const maxComplexity = 20

func findGoFiles() []string {
	var files []string
	excludeDirs := map[string]bool{
		"commands": true,
		"db":       true,
	}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
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
			strings.HasSuffix(path, "_test.go") ||
			strings.HasSuffix(path, "ccn_check.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
	}
	return files
}

func main() {
	files := findGoFiles()
	if len(files) == 0 {
		fmt.Println("No Go files found to analyze")
		os.Exit(0)
	}

	// Run gocyclo
	cmd := exec.Command("gocyclo", "-top", "1000")
	cmd.Args = append(cmd.Args, files...)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") ||
			strings.Contains(string(output), "executable file not found") {
			fmt.Println("Error: gocyclo is not installed.")
			fmt.Println("Install with: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "gocyclo error: %v\n%s\n", err, output)
		os.Exit(1)
	}

	// Parse output - format: "complexity function_name path:line:col"
	// Function name can contain spaces, so split from the right
	lines := strings.Split(string(output), "\n")
	exceeded := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Find the last space-separated integer (complexity) at the start
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		complexity, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if complexity > maxComplexity {
			exceeded = true
			// Reconstruct the full line for the error message
			fmt.Printf("FAIL: %s has cyclomatic complexity %d (max: %d)\n",
				line, complexity, maxComplexity)
		}
	}

	if exceeded {
		fmt.Printf("\nFAILED: Some functions exceed cyclomatic complexity limit of %d\n", maxComplexity)
		os.Exit(1)
	}

	fmt.Printf("PASS: All functions have cyclomatic complexity <= %d\n", maxComplexity)
	os.Exit(0)
}