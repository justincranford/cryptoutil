// Package main implements the identity-requirements-check tool for validating
// requirements traceability from acceptance criteria to test implementations.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultRequirementsFile = "docs/02-identityV2/requirements.yml"
	defaultRootPath         = "./internal/identity"
	defaultReportFile       = "docs/02-identityV2/REQUIREMENTS-COVERAGE.md"
)

type Requirement struct {
	Task               string   `yaml:"task"`
	ID                 string   `yaml:"id"`
	Description        string   `yaml:"description"`
	Category           string   `yaml:"category"`
	Priority           string   `yaml:"priority"`
	AcceptanceCriteria string   `yaml:"acceptance_criteria"`
	TestFiles          []string `yaml:"test_files,omitempty"`
	TestFunctions      []string `yaml:"test_functions,omitempty"`
	Validated          bool     `yaml:"validated"`
}

type RequirementsDoc struct {
	Metadata struct {
		Version      string `yaml:"version"`
		LastUpdated  string `yaml:"last_updated"`
		Source       string `yaml:"source"`
		TotalReqs    int    `yaml:"total_requirements"`
		TasksCovered int    `yaml:"tasks_covered"`
	} `yaml:"metadata"`
	Requirements map[string]Requirement
}

type TestMapping struct {
	FilePath       string
	FunctionName   string
	RequirementIDs []string
}

type CoverageStats struct {
	TotalRequirements     int
	ValidatedRequirements int
	UncoveredCritical     int
	UncoveredHigh         int
	UncoveredMedium       int
	UncoveredLow          int
	ByTask                map[string]*TaskStats
	ByCategory            map[string]*CategoryStats
	ByPriority            map[string]*PriorityStats
}

type TaskStats struct {
	TaskID    string
	Total     int
	Validated int
	Uncovered []Requirement
}

type CategoryStats struct {
	Category  string
	Total     int
	Validated int
}

type PriorityStats struct {
	Priority  string
	Total     int
	Validated int
}

func main() {
	ctx := context.Background()

	requirementsFile := flag.String("requirements", defaultRequirementsFile, "Path to requirements YAML file")
	rootPath := flag.String("root", defaultRootPath, "Root path for test file scanning")
	reportFile := flag.String("report", defaultReportFile, "Path to output coverage report")
	failOnUncovered := flag.Bool("fail", true, "Exit with error if critical requirements uncovered")
	flag.Parse()

	reqDoc, err := loadRequirements(*requirementsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load requirements: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Loaded %d requirements from %s\n", len(reqDoc.Requirements), *requirementsFile)

	testMappings, err := scanTestFiles(ctx, *rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to scan test files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 Found %d test mappings in %s\n", len(testMappings), *rootPath)

	coverage := mapRequirementsToTests(reqDoc, testMappings)

	stats := calculateCoverageStats(reqDoc, coverage)

	report := generateCoverageReport(reqDoc, stats)

	if err := os.WriteFile(*reportFile, []byte(report), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Coverage report written to %s\n", *reportFile)

	printSummary(stats)

	if *failOnUncovered && stats.UncoveredCritical > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ FAILED: %d critical requirements not validated by tests\n", stats.UncoveredCritical)
		os.Exit(1)
	}
}

func loadRequirements(filePath string) (*RequirementsDoc, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	doc := &RequirementsDoc{
		Requirements: make(map[string]Requirement),
	}

	if metadataRaw, ok := raw["metadata"].(map[string]interface{}); ok {
		metadataBytes, _ := yaml.Marshal(metadataRaw)
		yaml.Unmarshal(metadataBytes, &doc.Metadata)
	}

	for key, value := range raw {
		if key == "metadata" {
			continue
		}

		reqBytes, err := yaml.Marshal(value)
		if err != nil {
			continue
		}

		var req Requirement
		if err := yaml.Unmarshal(reqBytes, &req); err != nil {
			continue
		}

		if req.ID != "" {
			doc.Requirements[req.ID] = req
		}
	}

	return doc, nil
}

func scanTestFiles(ctx context.Context, rootPath string) ([]TestMapping, error) {
	var mappings []TestMapping

	requirementIDPattern := regexp.MustCompile(`[A-Z]\d+-\d+`)
	functionPattern := regexp.MustCompile(`func\s+(Test[A-Za-z0-9_]+)\s*\(`)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read test file %s: %w", path, err)
		}

		lines := strings.Split(string(content), "\n")
		var currentReqIDs []string
		inRequirementsBlock := false

		for i, line := range lines {
			if strings.Contains(line, "Validates requirements:") {
				inRequirementsBlock = true
				continue
			}

			if inRequirementsBlock {
				if strings.HasPrefix(strings.TrimSpace(line), "// -") {
					reqMatches := requirementIDPattern.FindAllString(line, -1)
					currentReqIDs = append(currentReqIDs, reqMatches...)
				} else if !strings.HasPrefix(strings.TrimSpace(line), "//") {
					inRequirementsBlock = false
				}
			}

			if matches := functionPattern.FindStringSubmatch(line); len(matches) > 1 {
				currentFunc := matches[1]

				if len(currentReqIDs) > 0 {
					mappings = append(mappings, TestMapping{
						FilePath:       path,
						FunctionName:   currentFunc,
						RequirementIDs: currentReqIDs,
					})
					fmt.Printf("   Found mapping: %s -> %v (line %d)\n", currentFunc, currentReqIDs, i+1)
				}

				currentReqIDs = nil
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	return mappings, nil
}

func mapRequirementsToTests(doc *RequirementsDoc, mappings []TestMapping) map[string]Requirement {
	coverage := make(map[string]Requirement)

	for id, req := range doc.Requirements {
		coverage[id] = req
	}

	for _, mapping := range mappings {
		for _, reqID := range mapping.RequirementIDs {
			if req, exists := coverage[reqID]; exists {
				req.TestFiles = append(req.TestFiles, mapping.FilePath)
				req.TestFunctions = append(req.TestFunctions, mapping.FunctionName)
				req.Validated = true
				coverage[reqID] = req
			}
		}
	}

	return coverage
}

func calculateCoverageStats(doc *RequirementsDoc, coverage map[string]Requirement) *CoverageStats {
	stats := &CoverageStats{
		ByTask:     make(map[string]*TaskStats),
		ByCategory: make(map[string]*CategoryStats),
		ByPriority: make(map[string]*PriorityStats),
	}

	for _, req := range coverage {
		stats.TotalRequirements++

		if req.Validated {
			stats.ValidatedRequirements++
		} else {
			switch req.Priority {
			case "CRITICAL":
				stats.UncoveredCritical++
			case "HIGH":
				stats.UncoveredHigh++
			case "MEDIUM":
				stats.UncoveredMedium++
			case "LOW":
				stats.UncoveredLow++
			}
		}

		if _, exists := stats.ByTask[req.Task]; !exists {
			stats.ByTask[req.Task] = &TaskStats{
				TaskID:    req.Task,
				Uncovered: []Requirement{},
			}
		}
		stats.ByTask[req.Task].Total++
		if req.Validated {
			stats.ByTask[req.Task].Validated++
		} else {
			stats.ByTask[req.Task].Uncovered = append(stats.ByTask[req.Task].Uncovered, req)
		}

		if _, exists := stats.ByCategory[req.Category]; !exists {
			stats.ByCategory[req.Category] = &CategoryStats{Category: req.Category}
		}
		stats.ByCategory[req.Category].Total++
		if req.Validated {
			stats.ByCategory[req.Category].Validated++
		}

		if _, exists := stats.ByPriority[req.Priority]; !exists {
			stats.ByPriority[req.Priority] = &PriorityStats{Priority: req.Priority}
		}
		stats.ByPriority[req.Priority].Total++
		if req.Validated {
			stats.ByPriority[req.Priority].Validated++
		}
	}

	return stats
}

func generateCoverageReport(doc *RequirementsDoc, stats *CoverageStats) string {
	var sb strings.Builder

	sb.WriteString("# Identity V2 Requirements Coverage Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n", doc.Metadata.LastUpdated))
	sb.WriteString(fmt.Sprintf("**Total Requirements**: %d\n", stats.TotalRequirements))
	sb.WriteString(fmt.Sprintf("**Validated**: %d (%.1f%%)\n", stats.ValidatedRequirements, float64(stats.ValidatedRequirements)/float64(stats.TotalRequirements)*100))
	sb.WriteString(fmt.Sprintf("**Uncovered CRITICAL**: %d\n", stats.UncoveredCritical))
	sb.WriteString(fmt.Sprintf("**Uncovered HIGH**: %d\n", stats.UncoveredHigh))
	sb.WriteString(fmt.Sprintf("**Uncovered MEDIUM**: %d\n", stats.UncoveredMedium))
	sb.WriteString("\n## Summary by Task\n\n")
	sb.WriteString("| Task | Requirements | Validated | Coverage |\n")
	sb.WriteString("|------|--------------|-----------|----------|\n")

	taskIDs := make([]string, 0, len(stats.ByTask))
	for taskID := range stats.ByTask {
		taskIDs = append(taskIDs, taskID)
	}
	sort.Strings(taskIDs)

	for _, taskID := range taskIDs {
		taskStats := stats.ByTask[taskID]
		pct := float64(taskStats.Validated) / float64(taskStats.Total) * 100
		status := "✅"
		if pct < 100 {
			status = "⚠️"
		}
		if pct == 0 {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %.1f%% %s |\n", taskID, taskStats.Total, taskStats.Validated, pct, status))
	}

	sb.WriteString("\n## Coverage by Category\n\n")

	categories := make([]string, 0, len(stats.ByCategory))
	for cat := range stats.ByCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		catStats := stats.ByCategory[cat]
		pct := float64(catStats.Validated) / float64(catStats.Total) * 100
		status := "✅"
		if pct < 100 {
			status = "⚠️"
		}
		if pct == 0 {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("### %s: %d/%d (%.1f%%) %s\n", cat, catStats.Validated, catStats.Total, pct, status))
	}

	sb.WriteString("\n## Coverage by Priority\n\n")

	priorities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}
	for _, pri := range priorities {
		if priStats, exists := stats.ByPriority[pri]; exists {
			pct := float64(priStats.Validated) / float64(priStats.Total) * 100
			status := "✅"
			if pct < 100 {
				status = "⚠️"
			}
			if pct == 0 {
				status = "❌"
			}
			sb.WriteString(fmt.Sprintf("### %s: %d/%d (%.1f%%) %s\n", pri, priStats.Validated, priStats.Total, pct, status))
		}
	}

	sb.WriteString("\n## Uncovered Requirements\n\n")

	for _, taskID := range taskIDs {
		taskStats := stats.ByTask[taskID]
		if len(taskStats.Uncovered) > 0 {
			sb.WriteString(fmt.Sprintf("### %s\n\n", taskID))
			sb.WriteString("| ID | Priority | Description |\n")
			sb.WriteString("|----|----------|-------------|\n")
			for _, req := range taskStats.Uncovered {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", req.ID, req.Priority, req.Description))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("**Report Generation Command**: `go run ./internal/cmd/cicd/identity-requirements-check`\n")
	sb.WriteString("**CI/CD Integration**: Add to `.github/workflows/ci-identity.yml` as quality gate\n")

	return sb.String()
}

func printSummary(stats *CoverageStats) {
	fmt.Println()
	fmt.Println("📊 Coverage Summary:")
	fmt.Printf("   Total Requirements: %d\n", stats.TotalRequirements)
	fmt.Printf("   Validated: %d (%.1f%%)\n", stats.ValidatedRequirements, float64(stats.ValidatedRequirements)/float64(stats.TotalRequirements)*100)
	fmt.Printf("   Uncovered CRITICAL: %d\n", stats.UncoveredCritical)
	fmt.Printf("   Uncovered HIGH: %d\n", stats.UncoveredHigh)
	fmt.Printf("   Uncovered MEDIUM: %d\n", stats.UncoveredMedium)
	fmt.Printf("   Uncovered LOW: %d\n", stats.UncoveredLow)
}
