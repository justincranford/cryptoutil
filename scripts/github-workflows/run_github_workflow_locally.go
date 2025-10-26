package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// WorkflowConfig defines a GitHub Actions workflow that can be run locally with act.
type WorkflowConfig struct {
	Name         string
	WorkflowFile string
	Description  string
	DefaultArgs  []string
}

// WorkflowResult captures the outcome of a workflow execution.
type WorkflowResult struct {
	Name            string
	Success         bool
	Duration        time.Duration
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

// ANSI color codes for console output.
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"

	// Status constants.
	statusSuccess = "✅ SUCCESS"
	statusFailed  = "❌ FAILED"
	taskSuccess   = "SUCCESS"
	taskFailed    = "FAILED"

	// File permissions.
	filePermissions = 0o644
	dirPermissions  = 0o755

	// Layout constants.
	lineWidth         = 80
	maxErrorDisplay   = 20
	maxWarningDisplay = 10
)

var (
	// Available workflows.
	workflows = map[string]WorkflowConfig{
		"e2e": {
			Name:         "e2e",
			WorkflowFile: ".github/workflows/ci-e2e.yml",
			Description:  "End-to-End Testing - Full system integration with Docker Compose",
			DefaultArgs:  []string{},
		},
		"dast": {
			Name:         "dast",
			WorkflowFile: ".github/workflows/ci-dast.yml",
			Description:  "Dynamic Application Security Testing - OWASP ZAP and Nuclei scans",
			DefaultArgs:  []string{"--input", "scan_profile=quick"},
		},
		"sast": {
			Name:         "sast",
			WorkflowFile: ".github/workflows/ci-sast.yml",
			Description:  "Static Application Security Testing - gosec and golangci-lint security checks",
			DefaultArgs:  []string{},
		},
		"robust": {
			Name:         "robust",
			WorkflowFile: ".github/workflows/ci-robust.yml",
			Description:  "Robustness Testing - Concurrency, race detection, fuzz tests, benchmarks",
			DefaultArgs:  []string{},
		},
		"quality": {
			Name:         "quality",
			WorkflowFile: ".github/workflows/ci-quality.yml",
			Description:  "Code Quality - Unit tests, coverage, linting, formatting checks",
			DefaultArgs:  []string{},
		},
	}

	// Command line flags.
	workflowNames = flag.String("workflows", "", "Comma-separated list of workflows to run (e2e,dast,sast,robust,quality)")
	outputDir     = flag.String("output", "workflow-reports", "Output directory for logs and reports")
	dryRun        = flag.Bool("dry-run", false, "Show what would be executed without running workflows")
	actPath       = flag.String("act-path", "act", "Path to act executable")
	actArgs       = flag.String("act-args", "", "Additional arguments to pass to act")
	showList      = flag.Bool("list", false, "List available workflows and exit")
)

func main() {
	flag.Parse()

	// Show available workflows if requested.
	if *showList {
		listWorkflows()
		os.Exit(0)
	}

	// Validate workflow names.
	if *workflowNames == "" {
		fmt.Fprintf(os.Stderr, "%sError: No workflows specified. Use -workflows flag.%s\n", colorRed, colorReset)
		fmt.Fprintf(os.Stderr, "Use -list to see available workflows.\n")
		os.Exit(1)
	}

	selectedWorkflows := parseWorkflowNames(*workflowNames)
	if len(selectedWorkflows) == 0 {
		fmt.Fprintf(os.Stderr, "%sError: No valid workflows specified.%s\n", colorRed, colorReset)
		os.Exit(1)
	}

	// Create execution summary.
	summary := &ExecutionSummary{
		StartTime:   time.Now(),
		Workflows:   make([]WorkflowResult, 0, len(selectedWorkflows)),
		OutputDir:   *outputDir,
		CombinedLog: filepath.Join(*outputDir, fmt.Sprintf("combined-%s.log", time.Now().Format("2006-01-02_15-04-05"))),
	}

	// Setup output directory.
	if err := os.MkdirAll(*outputDir, dirPermissions); err != nil {
		fmt.Fprintf(os.Stderr, "%sError creating output directory: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	// Open combined log file.
	combinedLog, err := os.OpenFile(summary.CombinedLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError creating combined log file: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	defer combinedLog.Close()

	// Print executive summary header.
	printExecutiveSummaryHeader(selectedWorkflows, combinedLog)

	// Execute workflows sequentially.
	for _, wf := range selectedWorkflows {
		result := executeWorkflow(wf, combinedLog)
		summary.Workflows = append(summary.Workflows, result)

		if result.Success {
			summary.TotalSuccess++
		} else {
			summary.TotalFailed++
		}
	}

	// Complete summary.
	summary.EndTime = time.Now()
	summary.TotalDuration = summary.EndTime.Sub(summary.StartTime)

	// Print final executive summary.
	printExecutiveSummaryFooter(summary, combinedLog)

	// Exit with appropriate code.
	if summary.TotalFailed > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

func listWorkflows() {
	fmt.Println("\n" + strings.Repeat("=", lineWidth))
	fmt.Printf("%s📋 Available GitHub Actions Workflows%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("=", lineWidth))

	for _, name := range []string{"e2e", "dast", "sast", "robust", "quality"} {
		wf := workflows[name]
		fmt.Printf("\n%s%-10s%s %s\n", colorGreen, wf.Name, colorReset, wf.Description)
		fmt.Printf("           File: %s\n", wf.WorkflowFile)

		if len(wf.DefaultArgs) > 0 {
			fmt.Printf("           Args: %s\n", strings.Join(wf.DefaultArgs, " "))
		}
	}

	fmt.Println("\n" + strings.Repeat("=", lineWidth))
	fmt.Println("\nUsage:")
	fmt.Println("  go run ./scripts/github-workflows/run_github_workflow_locally.go -workflows=e2e,dast")
	fmt.Println("  go run ./scripts/github-workflows/run_github_workflow_locally.go -workflows=quality -dry-run")
	fmt.Println("  go run ./scripts/github-workflows/run_github_workflow_locally.go -list")
	fmt.Println()
}

func parseWorkflowNames(names string) []WorkflowConfig {
	parts := strings.Split(names, ",")
	result := make([]WorkflowConfig, 0, len(parts))

	for _, name := range parts {
		name = strings.TrimSpace(name)
		if wf, ok := workflows[name]; ok {
			result = append(result, wf)
		} else {
			fmt.Fprintf(os.Stderr, "%sWarning: Unknown workflow '%s' (skipping)%s\n", colorYellow, name, colorReset)
		}
	}

	return result
}

func printExecutiveSummaryHeader(selectedWorkflows []WorkflowConfig, logFile *os.File) {
	header := fmt.Sprintf("\n%s\n", strings.Repeat("=", lineWidth))
	header += fmt.Sprintf("%s🚀 GITHUB ACTIONS LOCAL WORKFLOW EXECUTION%s\n", colorCyan, colorReset)
	header += fmt.Sprintf("%s\n", strings.Repeat("=", lineWidth))
	header += fmt.Sprintf("\n📅 Execution Started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	header += fmt.Sprintf("📊 Workflows Selected: %d\n", len(selectedWorkflows))

	if *dryRun {
		header += fmt.Sprintf("%s🔍 DRY RUN MODE - No workflows will be executed%s\n", colorYellow, colorReset)
	}

	header += "\n📋 Workflow Execution Plan:\n"
	header += strings.Repeat("-", lineWidth) + "\n"

	for i, wf := range selectedWorkflows {
		header += fmt.Sprintf("%2d. %s%-10s%s - %s\n", i+1, colorGreen, wf.Name, colorReset, wf.Description)
		header += fmt.Sprintf("    File: %s\n", wf.WorkflowFile)

		if len(wf.DefaultArgs) > 0 {
			header += fmt.Sprintf("    Args: %s\n", strings.Join(wf.DefaultArgs, " "))
		}
	}

	header += "\n" + strings.Repeat("=", lineWidth) + "\n"

	// Print to console and log file.
	fmt.Print(header)

	if logFile != nil {
		_, _ = logFile.WriteString(header) //nolint:errcheck // Logging errors are non-fatal
	}
}

func printExecutiveSummaryFooter(summary *ExecutionSummary, logFile *os.File) {
	footer := fmt.Sprintf("\n%s\n", strings.Repeat("=", lineWidth))
	footer += fmt.Sprintf("%s🎯 EXECUTION SUMMARY REPORT%s\n", colorCyan, colorReset)
	footer += fmt.Sprintf("%s\n", strings.Repeat("=", lineWidth))

	footer += fmt.Sprintf("\n📅 Execution Completed: %s\n", summary.EndTime.Format("2006-01-02 15:04:05"))
	footer += fmt.Sprintf("⏱️  Total Duration: %v\n", summary.TotalDuration.Round(time.Second))
	footer += fmt.Sprintf("📊 Total Workflows: %d\n", len(summary.Workflows))
	footer += fmt.Sprintf("%s✅ Successful: %d%s\n", colorGreen, summary.TotalSuccess, colorReset)

	if summary.TotalFailed > 0 {
		footer += fmt.Sprintf("%s❌ Failed: %d%s\n", colorRed, summary.TotalFailed, colorReset)
	} else {
		footer += fmt.Sprintf("❌ Failed: %d\n", summary.TotalFailed)
	}

	footer += "\n" + strings.Repeat("-", lineWidth) + "\n"
	footer += "📋 Workflow Results:\n"
	footer += strings.Repeat("-", lineWidth) + "\n"

	for i, wf := range summary.Workflows {
		status := statusSuccess
		color := colorGreen

		if !wf.Success {
			status = statusFailed
			color = colorRed
		}

		footer += fmt.Sprintf("%2d. %s%-10s%s %s%s%s (took %v)\n",
			i+1, colorGreen, wf.Name, colorReset, color, status, colorReset, wf.Duration.Round(time.Second))
		footer += fmt.Sprintf("    Log: %s\n", wf.LogFile)
		footer += fmt.Sprintf("    Analysis: %s\n", wf.AnalysisFile)

		if len(wf.ErrorMessages) > 0 {
			footer += fmt.Sprintf("    %sErrors: %d%s\n", colorRed, len(wf.ErrorMessages), colorReset)
		}

		if len(wf.Warnings) > 0 {
			footer += fmt.Sprintf("    %sWarnings: %d%s\n", colorYellow, len(wf.Warnings), colorReset)
		}
	}

	footer += "\n" + strings.Repeat("=", lineWidth) + "\n"
	footer += fmt.Sprintf("📁 Output Directory: %s\n", summary.OutputDir)
	footer += fmt.Sprintf("📄 Combined Log: %s\n", summary.CombinedLog)

	if summary.TotalFailed > 0 {
		footer += fmt.Sprintf("\n%s⚠️  EXECUTION STATUS: PARTIAL SUCCESS%s\n", colorYellow, colorReset)
	} else {
		footer += fmt.Sprintf("\n%s🎉 EXECUTION STATUS: FULL SUCCESS%s\n", colorGreen, colorReset)
	}

	footer += strings.Repeat("=", lineWidth) + "\n"

	// Print to console and log file.
	fmt.Print(footer)

	if logFile != nil {
		_, _ = logFile.WriteString(footer) //nolint:errcheck // Logging errors are non-fatal
	}
}

func executeWorkflow(wf WorkflowConfig, combinedLog *os.File) WorkflowResult {
	result := WorkflowResult{
		Name:         wf.Name,
		TaskResults:  make(map[string]TaskResult),
		LogFile:      filepath.Join(*outputDir, fmt.Sprintf("%s-%s.log", wf.Name, time.Now().Format("2006-01-02_15-04-05"))),
		AnalysisFile: filepath.Join(*outputDir, fmt.Sprintf("%s-analysis-%s.md", wf.Name, time.Now().Format("2006-01-02_15-04-05"))),
	}

	startTime := time.Now()

	// Print workflow header.
	header := fmt.Sprintf("\n%s\n", strings.Repeat("=", lineWidth))
	header += fmt.Sprintf("%s🔧 Executing Workflow: %s%s\n", colorCyan, wf.Name, colorReset)
	header += fmt.Sprintf("%s\n", strings.Repeat("=", lineWidth))
	header += fmt.Sprintf("📁 Log File: %s\n", result.LogFile)
	header += fmt.Sprintf("📄 Analysis File: %s\n", result.AnalysisFile)
	header += fmt.Sprintf("⏰ Started: %s\n", startTime.Format("15:04:05"))
	header += strings.Repeat("=", lineWidth) + "\n"

	fmt.Print(header)

	if combinedLog != nil {
		_, _ = combinedLog.WriteString(header) //nolint:errcheck // Logging errors are non-fatal
	}

	if *dryRun {
		dryRunMsg := fmt.Sprintf("%s🔍 DRY RUN: Would execute act with workflow: %s%s\n", colorYellow, wf.WorkflowFile, colorReset)
		dryRunMsg += fmt.Sprintf("   Command: %s workflow_dispatch -W %s\n", *actPath, wf.WorkflowFile)

		if len(wf.DefaultArgs) > 0 {
			dryRunMsg += fmt.Sprintf("   Args: %s\n", strings.Join(wf.DefaultArgs, " "))
		}

		if *actArgs != "" {
			dryRunMsg += fmt.Sprintf("   Extra Args: %s\n", *actArgs)
		}

		dryRunMsg += strings.Repeat("=", lineWidth) + "\n"

		fmt.Print(dryRunMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(dryRunMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = true
		result.Duration = time.Since(startTime)
		createAnalysisFile(result)

		return result
	}

	// Build act command.
	args := []string{"workflow_dispatch", "-W", wf.WorkflowFile}
	args = append(args, wf.DefaultArgs...)

	if *actArgs != "" {
		args = append(args, strings.Fields(*actArgs)...)
	}

	args = append(args, "--artifact-server-path", *outputDir)

	fmt.Printf("🚀 Executing: %s %s\n", *actPath, strings.Join(args, " "))

	// Create workflow log file.
	workflowLog, err := os.OpenFile(result.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating workflow log file: %v%s\n", colorRed, err, colorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.Duration = time.Since(startTime)
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}
	defer workflowLog.Close()

	// Execute act command.
	cmd := exec.Command(*actPath, args...) //nolint:gosec // User-controlled input is intentional for local testing

	// Setup stdout and stderr pipes for dual logging.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating stdout pipe: %v%s\n", colorRed, err, colorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.Duration = time.Since(startTime)
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating stderr pipe: %v%s\n", colorRed, err, colorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.Duration = time.Since(startTime)
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}

	if err := cmd.Start(); err != nil {
		errMsg := fmt.Sprintf("%sError starting act command: %v%s\n", colorRed, err, colorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.Duration = time.Since(startTime)
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}

	// Setup dual logging for stdout and stderr.
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		teeReader(stdoutPipe, os.Stdout, workflowLog, combinedLog)
	}()

	go func() {
		defer wg.Done()
		teeReader(stderrPipe, os.Stderr, workflowLog, combinedLog)
	}()

	// Wait for output to complete.
	wg.Wait()

	// Wait for command to finish.
	if err := cmd.Wait(); err != nil {
		result.Success = false
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
	} else {
		result.Success = true
	}

	result.Duration = time.Since(startTime)

	// Analyze workflow log for detailed results.
	analyzeWorkflowLog(result.LogFile, &result)

	// Create analysis markdown file.
	createAnalysisFile(result)

	// Print workflow footer.
	footer := fmt.Sprintf("\n%s\n", strings.Repeat("=", lineWidth))
	footer += fmt.Sprintf("%s✅ Workflow Completed: %s%s\n", colorCyan, wf.Name, colorReset)
	footer += fmt.Sprintf("⏰ Duration: %v\n", result.Duration.Round(time.Second))

	if result.Success {
		footer += fmt.Sprintf("%s✅ Status: SUCCESS%s\n", colorGreen, colorReset)
	} else {
		footer += fmt.Sprintf("%s❌ Status: FAILED%s\n", colorRed, colorReset)
	}

	footer += strings.Repeat("=", lineWidth) + "\n"

	fmt.Print(footer)

	if combinedLog != nil {
		_, _ = combinedLog.WriteString(footer) //nolint:errcheck // Logging errors are non-fatal
	}

	return result
}

func teeReader(reader io.Reader, writers ...io.Writer) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		for _, w := range writers {
			if w != nil {
				fmt.Fprintln(w, line)
			}
		}
	}
}

func analyzeWorkflowLog(logFile string, result *WorkflowResult) {
	content, err := os.ReadFile(logFile)
	if err != nil {
		result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("Failed to read log file: %v", err))

		return
	}

	logContent := string(content)

	// Check for workflow completion.
	if strings.Contains(logContent, "Job succeeded") {
		result.CompletionFound = true
		result.Success = true
	} else if strings.Contains(logContent, "Job failed") {
		result.CompletionFound = true
		result.Success = false
	}

	// Extract error messages.
	errorPattern := regexp.MustCompile(`(?i)(error|ERROR|failed|FAILED):.*`)

	errorMatches := errorPattern.FindAllString(logContent, -1)
	for _, match := range errorMatches {
		if !contains(result.ErrorMessages, match) {
			result.ErrorMessages = append(result.ErrorMessages, match)
		}
	}

	// Extract warnings.
	warningPattern := regexp.MustCompile(`(?i)(warning|WARNING|warn|WARN):.*`)

	warningMatches := warningPattern.FindAllString(logContent, -1)
	for _, match := range warningMatches {
		if !contains(result.Warnings, match) {
			result.Warnings = append(result.Warnings, match)
		}
	}

	// Workflow-specific task analysis.
	switch result.Name {
	case "dast":
		analyzeDastWorkflow(logContent, result)
	case "e2e":
		analyzeE2EWorkflow(logContent, result)
	case "sast":
		analyzeSastWorkflow(logContent, result)
	case "robust":
		analyzeRobustWorkflow(logContent, result)
	case "quality":
		analyzeQualityWorkflow(logContent, result)
	}
}

func analyzeDastWorkflow(logContent string, result *WorkflowResult) {
	// Check for Nuclei scan.
	if strings.Contains(logContent, "Nuclei - Vulnerability Scan") {
		status := taskFailed
		artifacts := []string{}

		if strings.Contains(logContent, "nuclei.log") || strings.Contains(logContent, "nuclei.sarif") {
			status = taskSuccess

			artifacts = append(artifacts, "nuclei.log", "nuclei.sarif")
		}

		result.TaskResults["Nuclei Scan"] = TaskResult{Name: "Nuclei Scan", Status: status, Artifacts: artifacts}
	}

	// Check for ZAP Full Scan.
	if strings.Contains(logContent, "OWASP ZAP DAST Scan") {
		status := taskFailed
		artifacts := []string{}

		if strings.Contains(logContent, "zap-report") {
			status = taskSuccess

			artifacts = append(artifacts, "zap-report.html")
		}

		result.TaskResults["ZAP Full Scan"] = TaskResult{Name: "ZAP Full Scan", Status: status, Artifacts: artifacts}
	}

	// Check for ZAP API Scan.
	if strings.Contains(logContent, "OWASP ZAP API Scan") {
		status := taskFailed
		artifacts := []string{}

		if strings.Contains(logContent, "zap-api-report") {
			status = taskSuccess

			artifacts = append(artifacts, "zap-api-report.html")
		}

		result.TaskResults["ZAP API Scan"] = TaskResult{Name: "ZAP API Scan", Status: status, Artifacts: artifacts}
	}

	// Check for header capture.
	if strings.Contains(logContent, "response-headers.txt") {
		result.TaskResults["Header Capture"] = TaskResult{
			Name:      "Header Capture",
			Status:    taskSuccess,
			Artifacts: []string{"response-headers.txt"},
		}
	}
}

func analyzeE2EWorkflow(logContent string, result *WorkflowResult) {
	// Check for Docker Compose services.
	if strings.Contains(logContent, "docker compose") {
		status := taskSuccess
		if strings.Contains(logContent, "failed") || strings.Contains(logContent, "error") {
			status = taskFailed
		}

		result.TaskResults["Docker Compose Setup"] = TaskResult{Name: "Docker Compose Setup", Status: status}
	}

	// Check for E2E tests.
	if strings.Contains(logContent, "go test -tags=e2e") {
		status := taskSuccess
		if strings.Contains(logContent, "FAIL:") {
			status = taskFailed
		}

		result.TaskResults["E2E Tests"] = TaskResult{Name: "E2E Tests", Status: status}
	}
}

func analyzeSastWorkflow(logContent string, result *WorkflowResult) {
	// Check for gosec scan.
	if strings.Contains(logContent, "gosec") {
		status := taskSuccess
		if strings.Contains(logContent, "Issues : ") && !strings.Contains(logContent, "Issues : 0") {
			status = taskFailed
		}

		result.TaskResults["Gosec Scan"] = TaskResult{Name: "Gosec Scan", Status: status}
	}

	// Check for golangci-lint.
	if strings.Contains(logContent, "golangci-lint") {
		status := taskSuccess
		if strings.Contains(logContent, "found issues") {
			status = taskFailed
		}

		result.TaskResults["Golangci-Lint"] = TaskResult{Name: "Golangci-Lint", Status: status}
	}
}

func analyzeRobustWorkflow(logContent string, result *WorkflowResult) {
	// Check for race detection.
	if strings.Contains(logContent, "-race") {
		status := taskSuccess
		if strings.Contains(logContent, "DATA RACE") {
			status = taskFailed
		}

		result.TaskResults["Race Detection"] = TaskResult{Name: "Race Detection", Status: status}
	}

	// Check for fuzz tests.
	if strings.Contains(logContent, "-fuzz") {
		status := taskSuccess
		if strings.Contains(logContent, "fuzz: elapsed") && strings.Contains(logContent, "FAIL") {
			status = taskFailed
		}

		result.TaskResults["Fuzz Tests"] = TaskResult{Name: "Fuzz Tests", Status: status}
	}

	// Check for benchmarks.
	if strings.Contains(logContent, "-bench") {
		result.TaskResults["Benchmarks"] = TaskResult{Name: "Benchmarks", Status: taskSuccess}
	}
}

func analyzeQualityWorkflow(logContent string, result *WorkflowResult) {
	// Check for unit tests.
	if strings.Contains(logContent, "go test") {
		status := taskSuccess
		if strings.Contains(logContent, "FAIL:") {
			status = taskFailed
		}

		result.TaskResults["Unit Tests"] = TaskResult{Name: "Unit Tests", Status: status}
	}

	// Check for coverage.
	if strings.Contains(logContent, "-coverprofile") {
		result.TaskResults["Coverage Report"] = TaskResult{Name: "Coverage Report", Status: taskSuccess}
	}
}

func createAnalysisFile(result WorkflowResult) {
	analysis := strings.Builder{}

	analysis.WriteString(fmt.Sprintf("# Workflow Analysis: %s\n\n", result.Name))
	analysis.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	analysis.WriteString("## Executive Summary\n\n")
	analysis.WriteString(fmt.Sprintf("- **Duration:** %v\n", result.Duration.Round(time.Second)))
	analysis.WriteString(fmt.Sprintf("- **Status:** %s\n", statusBadge(result.Success)))
	analysis.WriteString(fmt.Sprintf("- **Completion Found:** %v\n", result.CompletionFound))
	analysis.WriteString(fmt.Sprintf("- **Tasks Analyzed:** %d\n", len(result.TaskResults)))
	analysis.WriteString(fmt.Sprintf("- **Errors:** %d\n", len(result.ErrorMessages)))
	analysis.WriteString(fmt.Sprintf("- **Warnings:** %d\n\n", len(result.Warnings)))

	if len(result.TaskResults) > 0 {
		analysis.WriteString("## Task Results\n\n")
		analysis.WriteString("| Task | Status | Artifacts |\n")
		analysis.WriteString("|------|--------|----------|\n")

		for _, task := range result.TaskResults {
			artifacts := "None"
			if len(task.Artifacts) > 0 {
				artifacts = strings.Join(task.Artifacts, ", ")
			}

			analysis.WriteString(fmt.Sprintf("| %s | %s | %s |\n", task.Name, statusBadge(task.Status == "SUCCESS"), artifacts))
		}

		analysis.WriteString("\n")
	}

	if len(result.ErrorMessages) > 0 {
		analysis.WriteString("## Error Messages\n\n")

		for i, err := range result.ErrorMessages {
			if i >= maxErrorDisplay {
				analysis.WriteString(fmt.Sprintf("... and %d more errors (see log file)\n\n", len(result.ErrorMessages)-maxErrorDisplay))

				break
			}

			analysis.WriteString(fmt.Sprintf("- `%s`\n", err))
		}

		analysis.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		analysis.WriteString("## Warnings\n\n")

		for i, warn := range result.Warnings {
			if i >= maxWarningDisplay {
				analysis.WriteString(fmt.Sprintf("... and %d more warnings (see log file)\n\n", len(result.Warnings)-maxWarningDisplay))

				break
			}

			analysis.WriteString(fmt.Sprintf("- `%s`\n", warn))
		}

		analysis.WriteString("\n")
	}

	analysis.WriteString("## Log Files\n\n")
	analysis.WriteString(fmt.Sprintf("- **Workflow Log:** `%s`\n", result.LogFile))

	analysis.WriteString("\n## Recommendations\n\n")

	if !result.Success {
		analysis.WriteString("### Critical Issues\n\n")

		if !result.CompletionFound {
			analysis.WriteString("- ⚠️ Workflow did not complete (no completion marker found)\n")
		}

		if len(result.ErrorMessages) > 0 {
			analysis.WriteString(fmt.Sprintf("- ❌ %d error(s) detected - review log file for details\n", len(result.ErrorMessages)))
		}

		analysis.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		analysis.WriteString("### Warnings to Address\n\n")
		analysis.WriteString(fmt.Sprintf("- ⚠️ %d warning(s) detected - consider investigating\n\n", len(result.Warnings)))
	}

	if result.Success && len(result.ErrorMessages) == 0 && len(result.Warnings) == 0 {
		analysis.WriteString("✅ No issues detected - workflow executed successfully!\n\n")
	}

	// Write analysis to file.
	if err := os.WriteFile(result.AnalysisFile, []byte(analysis.String()), filePermissions); err != nil {
		fmt.Fprintf(os.Stderr, "%sError writing analysis file: %v%s\n", colorRed, err, colorReset)
	}
}

func statusBadge(success bool) string {
	if success {
		return "✅ SUCCESS"
	}

	return "❌ FAILED"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
