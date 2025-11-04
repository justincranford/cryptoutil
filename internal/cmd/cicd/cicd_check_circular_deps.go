// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the go-check-circular-package-dependencies command implementation.
package cicd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type PackageInfo struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

// goCheckCircularPackageDeps checks for circular dependencies in Go packages.
// It analyzes the dependency graph of all packages in the project and reports any circular dependencies.
// This command exits with code 1 if circular dependencies are found.
func goCheckCircularPackageDeps(logger *LogUtil) {
	logger.Log("Checking for circular dependencies in Go packages")

	logger.Log("Running: go list -json ./...")

	cmd := exec.Command("go", "list", "-json", "./...")

	output, err := cmd.Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error running go list: %v", err))
		os.Exit(1)
	}

	// Use the extracted function for the core logic
	if err := checkCircularDependencies(string(output)); err != nil {
		logger.Log(fmt.Sprintf("❌ RESULT: %v", err))

		logger.Log("goCheckCircularPackageDeps completed (circular dependencies found)")

		os.Exit(1) // Fail the build
	}

	fmt.Fprintln(os.Stderr, "✅ RESULT: No circular dependencies found")
	fmt.Fprintln(os.Stderr, "All internal package dependencies are acyclic.")

	logger.Log("goCheckCircularPackageDeps completed (no circular dependencies)")
}

// checkCircularDependencies analyzes the JSON output from 'go list -json ./...' for circular dependencies.
// Returns an error if circular dependencies are found, nil if no circular dependencies exist.
func checkCircularDependencies(jsonOutput string) error {
	// Parse JSON output (multiple JSON objects in stream)
	decoder := json.NewDecoder(strings.NewReader(jsonOutput))
	packages := make([]PackageInfo, 0)

	for {
		var pkg PackageInfo
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("failed to parse package info: %w", err)
		}

		packages = append(packages, pkg)
	}

	if len(packages) == 0 {
		return fmt.Errorf("no packages found")
	}

	// Build dependency graph
	dependencyGraph := make(map[string][]string)

	for _, pkg := range packages {
		dependencyGraph[pkg.ImportPath] = pkg.Imports
	}

	// Find circular dependencies using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	circularDeps := [][]string{}

	var dfs func(string, []string)
	dfs = func(pkg string, path []string) {
		// Check if package exists in dependency graph
		if _, exists := dependencyGraph[pkg]; !exists {
			return
		}

		visited[pkg] = true
		recursionStack[pkg] = true

		for _, dep := range dependencyGraph[pkg] {
			// Only check internal packages (those starting with our module name)
			if !strings.HasPrefix(dep, "cryptoutil/") {
				continue
			}

			if !visited[dep] {
				newPath := append(path, dep)
				dfs(dep, newPath)
			} else if recursionStack[dep] {
				// Found a cycle
				cycleStart := -1

				for i, p := range path {
					if p == dep {
						cycleStart = i

						break
					}
				}

				if cycleStart >= 0 {
					cycle := append(path[cycleStart:], dep)
					circularDeps = append(circularDeps, cycle)
				}
			}
		}

		recursionStack[pkg] = false
	}

	// Check each package for circular dependencies
	for pkg := range dependencyGraph {
		if !visited[pkg] {
			dfs(pkg, []string{pkg})
		}
	}

	if len(circularDeps) == 0 {
		return nil // No circular dependencies found
	}

	// Build error message with details about circular dependencies
	var errorMsg strings.Builder

	errorMsg.WriteString(fmt.Sprintf("Found %d circular dependency chain(s):\n\n", len(circularDeps)))

	for i, cycle := range circularDeps {
		errorMsg.WriteString(fmt.Sprintf("Chain %d (%d packages):\n", i+1, len(cycle)))

		for j, pkg := range cycle {
			prefix := "  "
			if j > 0 {
				prefix = "  → "
			}

			errorMsg.WriteString(fmt.Sprintf("%s%s\n", prefix, pkg))
		}

		errorMsg.WriteString("\n")
	}

	errorMsg.WriteString("Circular dependencies can prevent enabling advanced linters like gomnd.\n")
	errorMsg.WriteString("Consider refactoring to break these cycles.")

	return fmt.Errorf("circular dependencies detected: %s", errorMsg.String())
}
