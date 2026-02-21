// Copyright (c) 2025 Justin Cranford
//
//

// Package workflow provides workflow execution and orchestration utilities.
package workflow

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// WorkflowConfig defines a GitHub Actions workflow that can be run locally with act.
type WorkflowConfig struct {
	Description string
}

// WorkflowExecution represents a workflow to be executed with its name and config.
type WorkflowExecution struct {
	Name   string
	Config WorkflowConfig
}

// WorkflowResult captures the outcome of a workflow execution.
type WorkflowResult struct {
	Name            string
	Success         bool
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	CPUTime         time.Duration
	MemoryUsage     uint64 // Peak memory usage in bytes
	LogFile         string
	AnalysisFile    string
	TaskResults     map[string]TaskResult
	ErrorMessages   []string
	Warnings        []string
	CompletionFound bool
}

// TaskResult represents the outcome of an individual task/step within a workflow.
type TaskResult struct {
	Name      string
	Status    string // SUCCESS, FAILED, NOT_RUN, SKIPPED
	Artifacts []string
}

// ExecutionSummary aggregates results across all workflow executions.
type ExecutionSummary struct {
	StartTime     time.Time
	EndTime       time.Time
	TotalDuration time.Duration
	Workflows     []WorkflowResult
	TotalSuccess  int
	TotalFailed   int
	OutputDir     string
	CombinedLog   string
}

// Workflow executes the workflow runner with the provided command line arguments.
func Workflow(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return runWithWorkflowsDir(args[1:], ".github/workflows")
}

// run executes the workflow runner with the provided command line arguments.
// It uses the default .github/workflows directory for workflow discovery.
func run(args []string) int {
	return runWithWorkflowsDir(args, ".github/workflows")
}

// runWithWorkflowsDir executes the workflow runner with a configurable workflows directory.
// This separation enables testing without changing the process working directory.
func runWithWorkflowsDir(args []string, workflowsDir string) int {
	// Create flag set for parsing.
	fs := flag.NewFlagSet("workflow", flag.ContinueOnError)

	workflows := fs.String("workflows", "", "Comma-separated list of workflows to run (benchmark,coverage,dast,e2e,fuzz,gitleaks,load,quality,race,sast)")
	showHelp := fs.Bool("help", false, "Show usage with list available workflows and exit")
	outputDir := fs.String("output", "workflow-reports", "Output directory for logs and reports")
	dryRun := fs.Bool("dry-run", false, "Show what would be executed without running workflows")
	actPath := fs.String("act-path", "act", "Path to act executable")
	actArgs := fs.String("act-args", "", "Additional arguments to pass to act")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "%sError parsing flags: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)

		return 1
	}

	// Get available workflows - inline the call here
	availableWorkflows, err := getAvailableWorkflows(workflowsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError reading workflows directory: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)
		fmt.Fprintf(os.Stderr, "Make sure you're running from the project root with .github/workflows/ directory.\n")

		return 1
	}

	// Show available workflows if requested.
	if *showHelp {
		printHelp(availableWorkflows)

		return 0
	}

	// Validate parameter values
	if *workflows == "" {
		fmt.Fprintf(os.Stderr, "%sError: No workflows specified. Use -workflows flag.%s\n", cryptoutilSharedMagic.ColorRed, cryptoutilSharedMagic.ColorReset)
		fmt.Fprintf(os.Stderr, "Use -list to see available workflows.\n")

		return 1
	}

	selectedWorkflows := parseWorkflowNames(*workflows, availableWorkflows)
	if len(selectedWorkflows) == 0 {
		fmt.Fprintf(os.Stderr, "%sError: No valid workflows specified.%s\n", cryptoutilSharedMagic.ColorRed, cryptoutilSharedMagic.ColorReset)

		return 1
	}

	// Create execution summary.
	summary := &ExecutionSummary{
		StartTime:   time.Now().UTC(),
		Workflows:   make([]WorkflowResult, 0, len(selectedWorkflows)),
		OutputDir:   *outputDir,
		CombinedLog: filepath.Join(*outputDir, fmt.Sprintf("combined-%s.log", time.Now().UTC().Format("2006-01-02_15-04-05"))),
	}

	// Setup output directory.
	if err := os.MkdirAll(*outputDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute); err != nil {
		fmt.Fprintf(os.Stderr, "%sError creating output directory: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)

		return 1
	}

	// Open combined log file.
	combinedLog, err := os.OpenFile(summary.CombinedLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError creating combined log file: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)

		return 1
	}

	defer func() {
		if err := doCloseFile(combinedLog); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close combined log: %v\n", err)
		}
	}()

	// Print executive summary header.
	printExecutiveSummaryHeader(selectedWorkflows, combinedLog, *dryRun)

	// Execute workflows sequentially.
	for _, wf := range selectedWorkflows {
		result := executeWorkflow(wf, combinedLog, *outputDir, *dryRun, *actPath, *actArgs)
		summary.Workflows = append(summary.Workflows, result)

		if result.Success {
			summary.TotalSuccess++
		} else {
			summary.TotalFailed++
		}
	}

	// Complete summary.
	summary.EndTime = time.Now().UTC()
	summary.TotalDuration = summary.EndTime.Sub(summary.StartTime)

	// Print final executive summary.
	printExecutiveSummaryFooter(summary, combinedLog)

	// Exit with appropriate code.
	if summary.TotalFailed > 0 {
		return 1
	}

	return 0
}

// getAvailableWorkflows returns a map of available workflows by reading ci-*.yml files from .github/workflows/.
func getAvailableWorkflows(workflowsDir string) (map[string]WorkflowConfig, error) {
	files, err := os.ReadDir(workflowsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflows directory: %w", err)
	}

	workflows := make(map[string]WorkflowConfig)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// Match files with pattern ci-*.yml
		if strings.HasPrefix(filename, "ci-") && strings.HasSuffix(filename, ".yml") {
			// Extract workflow name from filename (remove ci- prefix and .yml suffix)
			workflowName := strings.TrimPrefix(filename, "ci-")
			workflowName = strings.TrimSuffix(workflowName, ".yml")

			workflows[workflowName] = WorkflowConfig{Description: getWorkflowDescription(workflowName)}
		}
	}

	if len(workflows) == 0 {
		fmt.Fprintf(os.Stderr, "%sError: No workflows found in %s directory%s\n", cryptoutilSharedMagic.ColorRed, workflowsDir, cryptoutilSharedMagic.ColorReset)
		fmt.Fprintf(os.Stderr, "Make sure workflow files %s/ci-*.yml exist.\n", workflowsDir)

		return nil, fmt.Errorf("no workflows found in %s directory", workflowsDir)
	}

	return workflows, nil
}

func printHelp(availableWorkflows map[string]WorkflowConfig) {
	fmt.Println("\n" + strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	fmt.Printf("%s📋 Available GitHub Actions Workflows%s\n", cryptoutilSharedMagic.ColorCyan, cryptoutilSharedMagic.ColorReset)
	fmt.Println(strings.Repeat("=", cryptoutilSharedMagic.LineWidth))

	for workflowName, workflowConfig := range availableWorkflows {
		fmt.Printf("\n%s%-12s%s %s\n", cryptoutilSharedMagic.ColorGreen, workflowName, cryptoutilSharedMagic.ColorReset, workflowConfig.Description)
		fmt.Printf("           File: %s\n", getWorkflowFile(workflowName))
	}

	fmt.Println("\n" + strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	fmt.Println("\nUsage:")
	fmt.Println("  go run ./cmd/workflow -workflows=e2e,dast")
	fmt.Println("  go run ./cmd/workflow -workflows=quality -dry-run")
	fmt.Println("  go run ./cmd/workflow -list")
	fmt.Println()
}

func parseWorkflowNames(names string, availableWorkflows map[string]WorkflowConfig) []WorkflowExecution {
	parts := strings.Split(names, ",")
	result := make([]WorkflowExecution, 0, len(parts))

	for _, name := range parts {
		name = strings.TrimSpace(name)
		if wf, ok := availableWorkflows[name]; ok {
			result = append(result, WorkflowExecution{Name: name, Config: wf})
		} else {
			fmt.Fprintf(os.Stderr, "%sWarning: Unknown workflow '%s' (skipping)%s\n", cryptoutilSharedMagic.ColorYellow, name, cryptoutilSharedMagic.ColorReset)
		}
	}

	return result
}

func printExecutiveSummaryHeader(selectedWorkflows []WorkflowExecution, logFile *os.File, dryRun bool) {
	header := fmt.Sprintf("\n%s\n", strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	header += fmt.Sprintf("%s🚀 GITHUB ACTIONS LOCAL WORKFLOW EXECUTION%s\n", cryptoutilSharedMagic.ColorCyan, cryptoutilSharedMagic.ColorReset)
	header += fmt.Sprintf("%s\n", strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	header += fmt.Sprintf("\n📅 Execution Started: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	header += fmt.Sprintf("📊 Workflows Selected: %d\n", len(selectedWorkflows))

	if dryRun {
		header += fmt.Sprintf("%s🔍 DRY RUN MODE - No workflows will be executed%s\n", cryptoutilSharedMagic.ColorYellow, cryptoutilSharedMagic.ColorReset)
	}

	header += "\n📋 Workflow Execution Plan:\n"
	header += strings.Repeat("-", cryptoutilSharedMagic.LineWidth) + "\n"

	for i, wf := range selectedWorkflows {
		header += fmt.Sprintf("%2d. %s%-10s%s - %s\n", i+1, cryptoutilSharedMagic.ColorGreen, wf.Name, cryptoutilSharedMagic.ColorReset, wf.Config.Description)
		header += fmt.Sprintf("    File: %s\n", getWorkflowFile(wf.Name))
	}

	header += "\n" + strings.Repeat("=", cryptoutilSharedMagic.LineWidth) + "\n"

	// Print to console and log file.
	fmt.Print(header)

	if logFile != nil {
		_, _ = logFile.WriteString(header) //nolint:errcheck // Logging errors are non-fatal
	}
}

func printExecutiveSummaryFooter(summary *ExecutionSummary, logFile *os.File) {
	footer := fmt.Sprintf("\n%s\n", strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	footer += fmt.Sprintf("%s🎯 EXECUTION SUMMARY REPORT%s\n", cryptoutilSharedMagic.ColorCyan, cryptoutilSharedMagic.ColorReset)
	footer += fmt.Sprintf("%s\n", strings.Repeat("=", cryptoutilSharedMagic.LineWidth))

	footer += fmt.Sprintf("\n📅 Execution Completed: %s\n", summary.EndTime.Format("2006-01-02 15:04:05"))
	footer += fmt.Sprintf("⏱️  Total Duration: %v\n", summary.TotalDuration.Round(time.Second))
	footer += fmt.Sprintf("📊 Total Workflows: %d\n", len(summary.Workflows))
	footer += fmt.Sprintf("%s✅ Successful: %d%s\n", cryptoutilSharedMagic.ColorGreen, summary.TotalSuccess, cryptoutilSharedMagic.ColorReset)

	if summary.TotalFailed > 0 {
		footer += fmt.Sprintf("%s❌ Failed: %d%s\n", cryptoutilSharedMagic.ColorRed, summary.TotalFailed, cryptoutilSharedMagic.ColorReset)
	} else {
		footer += fmt.Sprintf("❌ Failed: %d\n", summary.TotalFailed)
	}

	footer += "\n" + strings.Repeat("-", cryptoutilSharedMagic.LineWidth) + "\n"
	footer += "📋 Workflow Results:\n"
	footer += strings.Repeat("-", cryptoutilSharedMagic.LineWidth) + "\n"

	for i, wf := range summary.Workflows {
		status := cryptoutilSharedMagic.StatusSuccess
		color := cryptoutilSharedMagic.ColorGreen

		if !wf.Success {
			status = cryptoutilSharedMagic.StatusFailed
			color = cryptoutilSharedMagic.ColorRed
		}

		footer += fmt.Sprintf("%2d. %s%-10s%s %s%s%s (took %v)\n",
			i+1, cryptoutilSharedMagic.ColorGreen, wf.Name, cryptoutilSharedMagic.ColorReset, color, status, cryptoutilSharedMagic.ColorReset, wf.Duration.Round(time.Second))
		footer += fmt.Sprintf("    Log: %s\n", wf.LogFile)
		footer += fmt.Sprintf("    Analysis: %s\n", wf.AnalysisFile)

		if len(wf.ErrorMessages) > 0 {
			footer += fmt.Sprintf("    %sErrors: %d%s\n", cryptoutilSharedMagic.ColorRed, len(wf.ErrorMessages), cryptoutilSharedMagic.ColorReset)
		}

		if len(wf.Warnings) > 0 {
			footer += fmt.Sprintf("    %sWarnings: %d%s\n", cryptoutilSharedMagic.ColorYellow, len(wf.Warnings), cryptoutilSharedMagic.ColorReset)
		}
	}

	footer += "\n" + strings.Repeat("=", cryptoutilSharedMagic.LineWidth) + "\n"
	footer += fmt.Sprintf("📁 Output Directory: %s\n", summary.OutputDir)
	footer += fmt.Sprintf("📄 Combined Log: %s\n", summary.CombinedLog)

	if summary.TotalFailed > 0 {
		footer += fmt.Sprintf("\n%s⚠️  EXECUTION STATUS: PARTIAL SUCCESS%s\n", cryptoutilSharedMagic.ColorYellow, cryptoutilSharedMagic.ColorReset)
	} else {
		footer += fmt.Sprintf("\n%s🎉 EXECUTION STATUS: FULL SUCCESS%s\n", cryptoutilSharedMagic.ColorGreen, cryptoutilSharedMagic.ColorReset)
	}

	footer += strings.Repeat("=", cryptoutilSharedMagic.LineWidth) + "\n"

	// Print to console and log file.
	fmt.Print(footer)

	if logFile != nil {
		_, _ = logFile.WriteString(footer) //nolint:errcheck // Logging errors are non-fatal
	}
}
