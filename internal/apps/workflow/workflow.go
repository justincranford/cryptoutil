// Copyright (c) 2025 Justin Cranford
//
//

// Package workflow provides workflow execution and orchestration utilities.
package workflow

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	return run(args[1:])
}

// run executes the workflow runner with the provided command line arguments.
func run(args []string) int {
	// Create flag set for parsing.
	fs := flag.NewFlagSet("workflow", flag.ExitOnError)

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
	availableWorkflows, err := getAvailableWorkflows(".github/workflows")
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
		if err := combinedLog.Close(); err != nil {
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

func executeWorkflow(wf WorkflowExecution, combinedLog *os.File, outputDir string, dryRun bool, actPath, actArgs string) WorkflowResult {
	result := WorkflowResult{
		Name:         wf.Name,
		TaskResults:  make(map[string]TaskResult),
		LogFile:      getWorkflowLogFile(outputDir, wf.Name),
		AnalysisFile: filepath.Join(outputDir, fmt.Sprintf("%s-analysis-%s.md", wf.Name, time.Now().UTC().Format("2006-01-02_15-04-05"))),
	}

	startTime := time.Now().UTC()
	result.StartTime = startTime

	// Capture initial memory stats
	var initialMemStats runtime.MemStats

	runtime.GC() // Force garbage collection to get accurate baseline

	runtime.ReadMemStats(&initialMemStats)

	// Print workflow header.
	header := fmt.Sprintf("\n%s\n", strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	header += fmt.Sprintf("%s🔧 Executing Workflow: %s%s\n", cryptoutilSharedMagic.ColorCyan, wf.Name, cryptoutilSharedMagic.ColorReset)
	header += fmt.Sprintf("%s\n", strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	header += fmt.Sprintf("📁 Log File: %s\n", result.LogFile)
	header += fmt.Sprintf("📄 Analysis File: %s\n", result.AnalysisFile)
	header += fmt.Sprintf("⏰ Started: %s\n", startTime.Format("15:04:05"))
	header += strings.Repeat("=", cryptoutilSharedMagic.LineWidth) + "\n"

	fmt.Print(header)

	if combinedLog != nil {
		_, _ = combinedLog.WriteString(header) //nolint:errcheck // Logging errors are non-fatal
	}

	if dryRun {
		// Build act command for dry-run display.
		// Use workflow_dispatch for DAST (supports inputs), push for others
		event := cryptoutilSharedMagic.EventTypePush
		if wf.Name == cryptoutilSharedMagic.WorkflowNameDAST {
			event = cryptoutilSharedMagic.EventTypeWorkflowDispatch
		}

		dryRunMsg := fmt.Sprintf("%s🔍 DRY RUN: Would execute act with workflow: %s%s\n", cryptoutilSharedMagic.ColorYellow, getWorkflowFile(wf.Name), cryptoutilSharedMagic.ColorReset)
		dryRunMsg += fmt.Sprintf("   Command: %s %s -W %s\n", actPath, event, getWorkflowFile(wf.Name))

		if actArgs != "" {
			dryRunMsg += fmt.Sprintf("   Extra Args: %s\n", actArgs)
		}

		dryRunMsg += strings.Repeat("=", cryptoutilSharedMagic.LineWidth) + "\n"

		fmt.Print(dryRunMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(dryRunMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = true
		result.EndTime = time.Now().UTC()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0) // No CPU time for dry run
		result.MemoryUsage = 0            // No memory usage for dry run
		createAnalysisFile(result)

		return result
	}

	// Build act command.
	// Use workflow_dispatch for DAST (supports inputs), push for others
	event := cryptoutilSharedMagic.EventTypePush
	if wf.Name == cryptoutilSharedMagic.WorkflowNameDAST {
		event = cryptoutilSharedMagic.EventTypeWorkflowDispatch
	}

	args := []string{event, "-W", getWorkflowFile(wf.Name)}

	if actArgs != "" {
		args = append(args, strings.Fields(actArgs)...)
	}

	args = append(args, "--artifact-server-path", outputDir)

	fmt.Printf("🚀 Executing: %s %s\n", actPath, strings.Join(args, " "))

	// Create workflow log file.
	workflowLog, err := os.OpenFile(result.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead)
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating workflow log file: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.EndTime = time.Now().UTC()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0)
		result.MemoryUsage = 0
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}

	defer func() {
		if err := workflowLog.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close workflow log: %v\n", err)
		}
	}()

	// Execute act command.
	cmd := exec.CommandContext(context.Background(), actPath, args...) //nolint:gosec // User-controlled input is intentional for local testing

	// Setup stdout and stderr pipes for dual logging.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating stdout pipe: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.EndTime = time.Now().UTC()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0)
		result.MemoryUsage = 0
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating stderr pipe: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.EndTime = time.Now().UTC()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0)
		result.MemoryUsage = 0
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}

	if err := cmd.Start(); err != nil {
		errMsg := fmt.Sprintf("%sError starting act command: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.EndTime = time.Now().UTC()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0)
		result.MemoryUsage = 0
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

	// Capture final metrics
	endTime := time.Now().UTC()
	result.EndTime = endTime
	result.Duration = endTime.Sub(startTime)

	// Capture final memory stats and calculate peak usage
	var finalMemStats runtime.MemStats

	runtime.GC() // Force garbage collection to get accurate final stats

	runtime.ReadMemStats(&finalMemStats)
	result.MemoryUsage = finalMemStats.Alloc // Current allocated memory

	// For CPU time, we'll use wall clock time as approximation since getting process CPU time is complex
	result.CPUTime = result.Duration // Approximation - in a real implementation we'd track actual CPU time

	// Analyze workflow log for detailed results.
	analyzeWorkflowLog(result.LogFile, &result)

	// Create analysis markdown file.
	createAnalysisFile(result)

	// Print workflow footer.
	footer := fmt.Sprintf("\n%s\n", strings.Repeat("=", cryptoutilSharedMagic.LineWidth))
	footer += fmt.Sprintf("%s✅ Workflow Completed: %s%s\n", cryptoutilSharedMagic.ColorCyan, wf.Name, cryptoutilSharedMagic.ColorReset)
	footer += fmt.Sprintf("⏰ Duration: %v\n", result.Duration.Round(time.Second))

	if result.Success {
		footer += fmt.Sprintf("%s✅ Status: SUCCESS%s\n", cryptoutilSharedMagic.ColorGreen, cryptoutilSharedMagic.ColorReset)
	} else {
		footer += fmt.Sprintf("%s❌ Status: FAILED%s\n", cryptoutilSharedMagic.ColorRed, cryptoutilSharedMagic.ColorReset)
	}

	footer += strings.Repeat("=", cryptoutilSharedMagic.LineWidth) + "\n"

	fmt.Print(footer)

	if combinedLog != nil {
		_, _ = combinedLog.WriteString(footer) //nolint:errcheck // Logging errors are non-fatal
	}

	return result
}

// getWorkflowFile returns the workflow file path for a given workflow name.
func getWorkflowFile(workflowName string) string {
	return fmt.Sprintf(".github/workflows/ci-%s.yml", workflowName)
}

// getWorkflowDescription returns the description for a given workflow name by reading it from the workflow file.
func getWorkflowDescription(workflowName string) string {
	workflowFile := getWorkflowFile(workflowName)

	content, err := os.ReadFile(workflowFile)
	if err != nil {
		// Fallback to a generic description if file cannot be read
		caser := cases.Title(language.English)

		return fmt.Sprintf("%s workflow", caser.String(workflowName))
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			// Extract the name value, removing quotes if present
			nameValue := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			nameValue = strings.Trim(nameValue, `"'`)

			return nameValue
		}
	}

	// Fallback if name not found
	caser := cases.Title(language.English)

	return fmt.Sprintf("%s workflow", caser.String(workflowName))
}

func getWorkflowLogFile(outputDir, workflowName string) string {
	// All workflows use the same timestamped naming pattern
	return filepath.Join(outputDir, fmt.Sprintf("%s-%s.log", workflowName, time.Now().UTC().Format("2006-01-02_15-04-05")))
}

func teeReader(reader io.Reader, writers ...io.Writer) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		for _, w := range writers {
			if w != nil {
				_, _ = fmt.Fprintln(w, line) //nolint:errcheck // Internal logging helper
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
}

func createAnalysisFile(result WorkflowResult) {
	analysis := strings.Builder{}

	analysis.WriteString(fmt.Sprintf("# Workflow Analysis: %s\n\n", result.Name))
	analysis.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().UTC().Format("2006-01-02 15:04:05")))

	analysis.WriteString("## Executive Summary\n\n")
	analysis.WriteString(fmt.Sprintf("- **Duration:** %v\n", result.Duration.Round(time.Second)))
	analysis.WriteString(fmt.Sprintf("- **Status:** %s\n", statusBadge(result.Success)))
	analysis.WriteString(fmt.Sprintf("- **Completion Found:** %v\n", result.CompletionFound))
	analysis.WriteString(fmt.Sprintf("- **Tasks Analyzed:** %d\n", len(result.TaskResults)))
	analysis.WriteString(fmt.Sprintf("- **Errors:** %d\n", len(result.ErrorMessages)))
	analysis.WriteString(fmt.Sprintf("- **Warnings:** %d\n\n", len(result.Warnings)))

	analysis.WriteString("## Execution Metrics\n\n")
	analysis.WriteString(fmt.Sprintf("- **Start Time:** %s\n", result.StartTime.Format("2006-01-02 15:04:05")))
	analysis.WriteString(fmt.Sprintf("- **End Time:** %s\n", result.EndTime.Format("2006-01-02 15:04:05")))
	analysis.WriteString(fmt.Sprintf("- **Duration:** %v\n", result.Duration.Round(time.Millisecond)))
	analysis.WriteString(fmt.Sprintf("- **CPU Time:** %v (approximated)\n", result.CPUTime.Round(time.Millisecond)))
	analysis.WriteString(fmt.Sprintf("- **Memory Usage:** %.2f MB\n\n", float64(result.MemoryUsage)/cryptoutilSharedMagic.BytesPerMB))

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
			if i >= cryptoutilSharedMagic.MaxErrorDisplay {
				analysis.WriteString(fmt.Sprintf("... and %d more errors (see log file)\n\n", len(result.ErrorMessages)-cryptoutilSharedMagic.MaxErrorDisplay))

				break
			}

			analysis.WriteString(fmt.Sprintf("- `%s`\n", err))
		}

		analysis.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		analysis.WriteString("## Warnings\n\n")

		for i, warn := range result.Warnings {
			if i >= cryptoutilSharedMagic.MaxWarningDisplay {
				analysis.WriteString(fmt.Sprintf("... and %d more warnings (see log file)\n\n", len(result.Warnings)-cryptoutilSharedMagic.MaxWarningDisplay))

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
	if err := os.WriteFile(result.AnalysisFile, []byte(analysis.String()), cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead); err != nil {
		fmt.Fprintf(os.Stderr, "%sError writing analysis file: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)
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
