package workflow

import (
	"bufio"
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

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// WorkflowConfig defines a GitHub Actions workflow that can be run locally with act.
type WorkflowConfig struct {
	// Description is computed dynamically from the workflow file
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

// ANSI color codes for console output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"

	// Status constants.
	statusSuccess = "✅ SUCCESS"
	statusFailed  = "❌ FAILED"
	taskSuccess   = "SUCCESS"
	taskFailed    = "FAILED"

	// Workflow names.
	workflowNameDAST = "dast"
	workflowNameLoad = "load"

	// Event types.
	eventTypePush             = "push"
	eventTypeWorkflowDispatch = "workflow_dispatch"

	// File permissions.
	filePermissions = 0o644
	dirPermissions  = 0o755

	// Layout constants.
	lineWidth         = 80
	maxErrorDisplay   = 20
	maxWarningDisplay = 10

	// Memory conversion constants.
	bytesPerMB = 1024 * 1024
)

// getAvailableWorkflows returns a map of available workflows by reading ci-*.yml files from .github/workflows/.
func getAvailableWorkflows() (map[string]WorkflowConfig, error) {
	workflowsDir := ".github/workflows"

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

			workflows[workflowName] = WorkflowConfig{}
		}
	}

	return workflows, nil
}

// Available workflowNames. In alphabetical order.
var workflowNames = func() map[string]WorkflowConfig {
	workflows, err := getAvailableWorkflows()
	if err != nil {
		// Fallback to hardcoded list if directory read fails
		return map[string]WorkflowConfig{
			"benchmark": {},
			"coverage":  {},
			"dast":      {},
			"e2e":       {},
			"fuzz":      {},
			"gitleaks":  {},
			"load":      {},
			"quality":   {},
			"race":      {},
			"sast":      {},
		}
	}

	return workflows
}()

// Run executes the workflow runner with the provided command line arguments.
func Run(args []string) int {
	// Create flag set for parsing.
	fs := flag.NewFlagSet("workflow", flag.ExitOnError)

	workflowNames := fs.String("workflows", "", "Comma-separated list of workflows to run (quality,coverage,benchmark,gitleaks,sast,race,fuzz,e2e,dast,load)")
	outputDir := fs.String("output", "workflow-reports", "Output directory for logs and reports")
	dryRun := fs.Bool("dry-run", false, "Show what would be executed without running workflows")
	actPath := fs.String("act-path", "act", "Path to act executable")
	actArgs := fs.String("act-args", "", "Additional arguments to pass to act")
	showList := fs.Bool("list", false, "List available workflows and exit")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "%sError parsing flags: %v%s\n", colorRed, err, colorReset)

		return 1
	}

	// Show available workflows if requested.
	if *showList {
		listWorkflows()

		return 0
	}

	// Validate workflow names.
	if *workflowNames == "" {
		fmt.Fprintf(os.Stderr, "%sError: No workflows specified. Use -workflows flag.%s\n", colorRed, colorReset)
		fmt.Fprintf(os.Stderr, "Use -list to see available workflows.\n")

		return 1
	}

	selectedWorkflows := parseWorkflowNames(*workflowNames)
	if len(selectedWorkflows) == 0 {
		fmt.Fprintf(os.Stderr, "%sError: No valid workflows specified.%s\n", colorRed, colorReset)

		return 1
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

		return 1
	}

	// Open combined log file.
	combinedLog, err := os.OpenFile(summary.CombinedLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError creating combined log file: %v%s\n", colorRed, err, colorReset)

		return 1
	}
	defer combinedLog.Close()

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
	summary.EndTime = time.Now()
	summary.TotalDuration = summary.EndTime.Sub(summary.StartTime)

	// Print final executive summary.
	printExecutiveSummaryFooter(summary, combinedLog)

	// Exit with appropriate code.
	if summary.TotalFailed > 0 {
		return 1
	}

	return 0
}

func listWorkflows() {
	fmt.Println("\n" + strings.Repeat("=", lineWidth))
	fmt.Printf("%s📋 Available GitHub Actions Workflows%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("=", lineWidth))

	for workflowName := range workflowNames {
		fmt.Printf("\n%s%-12s%s %s\n", colorGreen, workflowName, colorReset, getWorkflowDescription(workflowName))
		fmt.Printf("           File: %s\n", getWorkflowFile(workflowName))

		if len(getWorkflowDefaultArgs(workflowName)) > 0 {
			fmt.Printf("           Args: %s\n", strings.Join(getWorkflowDefaultArgs(workflowName), " "))
		}
	}

	fmt.Println("\n" + strings.Repeat("=", lineWidth))
	fmt.Println("\nUsage:")
	fmt.Println("  go run ./cmd/workflow -workflows=e2e,dast")
	fmt.Println("  go run ./cmd/workflow -workflows=quality -dry-run")
	fmt.Println("  go run ./cmd/workflow -list")
	fmt.Println()
}

func parseWorkflowNames(names string) []WorkflowExecution {
	parts := strings.Split(names, ",")
	result := make([]WorkflowExecution, 0, len(parts))

	for _, name := range parts {
		name = strings.TrimSpace(name)
		if wf, ok := workflowNames[name]; ok {
			result = append(result, WorkflowExecution{Name: name, Config: wf})
		} else {
			fmt.Fprintf(os.Stderr, "%sWarning: Unknown workflow '%s' (skipping)%s\n", colorYellow, name, colorReset)
		}
	}

	return result
}

func printExecutiveSummaryHeader(selectedWorkflows []WorkflowExecution, logFile *os.File, dryRun bool) {
	header := fmt.Sprintf("\n%s\n", strings.Repeat("=", lineWidth))
	header += fmt.Sprintf("%s🚀 GITHUB ACTIONS LOCAL WORKFLOW EXECUTION%s\n", colorCyan, colorReset)
	header += fmt.Sprintf("%s\n", strings.Repeat("=", lineWidth))
	header += fmt.Sprintf("\n📅 Execution Started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	header += fmt.Sprintf("📊 Workflows Selected: %d\n", len(selectedWorkflows))

	if dryRun {
		header += fmt.Sprintf("%s🔍 DRY RUN MODE - No workflows will be executed%s\n", colorYellow, colorReset)
	}

	header += "\n📋 Workflow Execution Plan:\n"
	header += strings.Repeat("-", lineWidth) + "\n"

	for i, wf := range selectedWorkflows {
		header += fmt.Sprintf("%2d. %s%-10s%s - %s\n", i+1, colorGreen, wf.Name, colorReset, getWorkflowDescription(wf.Name))
		header += fmt.Sprintf("    File: %s\n", getWorkflowFile(wf.Name))

		if len(getWorkflowDefaultArgs(wf.Name)) > 0 {
			header += fmt.Sprintf("    Args: %s\n", strings.Join(getWorkflowDefaultArgs(wf.Name), " "))
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

func executeWorkflow(wf WorkflowExecution, combinedLog *os.File, outputDir string, dryRun bool, actPath, actArgs string) WorkflowResult {
	result := WorkflowResult{
		Name:         wf.Name,
		TaskResults:  make(map[string]TaskResult),
		LogFile:      getWorkflowLogFile(outputDir, wf.Name),
		AnalysisFile: filepath.Join(outputDir, fmt.Sprintf("%s-analysis-%s.md", wf.Name, time.Now().Format("2006-01-02_15-04-05"))),
	}

	startTime := time.Now()
	result.StartTime = startTime

	// Capture initial memory stats
	var initialMemStats runtime.MemStats

	runtime.GC() // Force garbage collection to get accurate baseline

	runtime.ReadMemStats(&initialMemStats)

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

	if dryRun {
		// Build act command for dry-run display.
		// Use workflow_dispatch for DAST (supports inputs), push for others
		event := eventTypePush
		if wf.Name == workflowNameDAST {
			event = eventTypeWorkflowDispatch
		}

		dryRunMsg := fmt.Sprintf("%s🔍 DRY RUN: Would execute act with workflow: %s%s\n", colorYellow, getWorkflowFile(wf.Name), colorReset)
		dryRunMsg += fmt.Sprintf("   Command: %s %s -W %s\n", actPath, event, getWorkflowFile(wf.Name))

		if len(getWorkflowDefaultArgs(wf.Name)) > 0 {
			dryRunMsg += fmt.Sprintf("   Args: %s\n", strings.Join(getWorkflowDefaultArgs(wf.Name), " "))
		}

		if actArgs != "" {
			dryRunMsg += fmt.Sprintf("   Extra Args: %s\n", actArgs)
		}

		dryRunMsg += strings.Repeat("=", lineWidth) + "\n"

		fmt.Print(dryRunMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(dryRunMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = true
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0) // No CPU time for dry run
		result.MemoryUsage = 0            // No memory usage for dry run
		createAnalysisFile(result)

		return result
	}

	// Build act command.
	// Use workflow_dispatch for DAST (supports inputs), push for others
	event := eventTypePush
	if wf.Name == workflowNameDAST {
		event = eventTypeWorkflowDispatch
	}

	args := []string{event, "-W", getWorkflowFile(wf.Name)}
	args = append(args, getWorkflowDefaultArgs(wf.Name)...)

	if actArgs != "" {
		args = append(args, strings.Fields(actArgs)...)
	}

	args = append(args, "--artifact-server-path", outputDir)

	fmt.Printf("🚀 Executing: %s %s\n", actPath, strings.Join(args, " "))

	// Create workflow log file.
	workflowLog, err := os.OpenFile(result.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating workflow log file: %v%s\n", colorRed, err, colorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0)
		result.MemoryUsage = 0
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		createAnalysisFile(result)

		return result
	}
	defer workflowLog.Close()

	// Execute act command.
	cmd := exec.Command(actPath, args...) //nolint:gosec // User-controlled input is intentional for local testing

	// Setup stdout and stderr pipes for dual logging.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating stdout pipe: %v%s\n", colorRed, err, colorReset)
		fmt.Print(errMsg)

		if combinedLog != nil {
			_, _ = combinedLog.WriteString(errMsg) //nolint:errcheck // Logging errors are non-fatal
		}

		result.Success = false
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0)
		result.MemoryUsage = 0
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
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.CPUTime = time.Duration(0)
		result.MemoryUsage = 0
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
		result.EndTime = time.Now()
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
	endTime := time.Now()
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

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
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

// getWorkflowDefaultArgs returns the default arguments for a given workflow name.
func getWorkflowDefaultArgs(workflowName string) []string {
	switch workflowName {
	case workflowNameDAST:
		return []string{"--input", "scan_profile=quick"}
	case workflowNameLoad:
		return []string{"--input", "load_profile=quick"}
	default:
		return []string{}
	}
}

func getWorkflowLogFile(outputDir, workflowName string) string {
	// All workflows use the same timestamped naming pattern
	return filepath.Join(outputDir, fmt.Sprintf("%s-%s.log", workflowName, time.Now().Format("2006-01-02_15-04-05")))
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
	case "race":
		analyzeRaceWorkflow(logContent, result)
	case "fuzz":
		analyzeFuzzWorkflow(logContent, result)
	case "quality":
		analyzeQualityWorkflow(logContent, result)
	case "coverage":
		analyzeCoverageWorkflow(logContent, result)
	case "benchmark":
		analyzeBenchmarkWorkflow(logContent, result)
	case "gitleaks":
		analyzeGitleaksWorkflow(logContent, result)
	case "load":
		analyzeLoadWorkflow(logContent, result)
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

func analyzeLoadWorkflow(logContent string, result *WorkflowResult) {
	// Check for Gatling load tests.
	if strings.Contains(logContent, "gatling") || strings.Contains(logContent, "Gatling") {
		status := taskSuccess
		if strings.Contains(logContent, "failed") || strings.Contains(logContent, "error") {
			status = taskFailed
		}

		result.TaskResults["Gatling Load Tests"] = TaskResult{Name: "Gatling Load Tests", Status: status}
	}

	// Check for Docker Compose services.
	if strings.Contains(logContent, "docker compose") {
		status := taskSuccess
		if strings.Contains(logContent, "failed") || strings.Contains(logContent, "error") {
			status = taskFailed
		}

		result.TaskResults["Docker Compose Setup"] = TaskResult{Name: "Docker Compose Setup", Status: status}
	}
}

func analyzeGitleaksWorkflow(logContent string, result *WorkflowResult) {
	// Check for Gitleaks scan.
	if strings.Contains(logContent, "gitleaks") || strings.Contains(logContent, "Gitleaks") {
		status := taskSuccess
		if strings.Contains(logContent, "leaks found") || strings.Contains(logContent, "failed") {
			status = taskFailed
		}

		result.TaskResults["Gitleaks Secrets Scan"] = TaskResult{Name: "Gitleaks Secrets Scan", Status: status}
	}
}

func analyzeBenchmarkWorkflow(logContent string, result *WorkflowResult) {
	// Check for benchmarks.
	if strings.Contains(logContent, "-bench") {
		status := taskSuccess
		if strings.Contains(logContent, "FAIL") {
			status = taskFailed
		}

		result.TaskResults["Benchmark Tests"] = TaskResult{Name: "Benchmark Tests", Status: status}
	}
}

func analyzeCoverageWorkflow(logContent string, result *WorkflowResult) {
	// Check for coverage collection.
	if strings.Contains(logContent, "-coverprofile") {
		status := taskSuccess
		if strings.Contains(logContent, "failed") || strings.Contains(logContent, "error") {
			status = taskFailed
		}

		result.TaskResults["Coverage Collection"] = TaskResult{Name: "Coverage Collection", Status: status}
	}

	// Check for coverage reporting.
	if strings.Contains(logContent, "codecov") || strings.Contains(logContent, "coverage.html") {
		result.TaskResults["Coverage Reporting"] = TaskResult{Name: "Coverage Reporting", Status: taskSuccess}
	}
}

func analyzeFuzzWorkflow(logContent string, result *WorkflowResult) {
	// Check for fuzz tests - keygen package.
	if strings.Contains(logContent, "FuzzGenerateRSAKeyPair") ||
		strings.Contains(logContent, "FuzzGenerateECDSAKeyPair") ||
		strings.Contains(logContent, "FuzzGenerateECDHKeyPair") ||
		strings.Contains(logContent, "FuzzGenerateEDDSAKeyPair") ||
		strings.Contains(logContent, "FuzzGenerateAESKey") ||
		strings.Contains(logContent, "FuzzGenerateAESHSKey") ||
		strings.Contains(logContent, "FuzzGenerateHMACKey") {
		status := taskSuccess
		if strings.Contains(logContent, "fuzz: elapsed") && strings.Contains(logContent, "FAIL") {
			status = taskFailed
		}

		result.TaskResults["Fuzz Tests - Keygen"] = TaskResult{Name: "Fuzz Tests - Keygen", Status: status}
	}

	// Check for fuzz tests - digests package.
	if strings.Contains(logContent, "FuzzHKDFAllVariants") ||
		strings.Contains(logContent, "FuzzHKDFwithSHA256") ||
		strings.Contains(logContent, "FuzzHKDFwithSHA384") ||
		strings.Contains(logContent, "FuzzHKDFwithSHA512") ||
		strings.Contains(logContent, "FuzzHKDFwithSHA224") ||
		strings.Contains(logContent, "FuzzSHA512") ||
		strings.Contains(logContent, "FuzzSHA384") ||
		strings.Contains(logContent, "FuzzSHA256") ||
		strings.Contains(logContent, "FuzzSHA224") {
		status := taskSuccess
		if strings.Contains(logContent, "fuzz: elapsed") && strings.Contains(logContent, "FAIL") {
			status = taskFailed
		}

		result.TaskResults["Fuzz Tests - Digests"] = TaskResult{Name: "Fuzz Tests - Digests", Status: status}
	}
}

func analyzeRaceWorkflow(logContent string, result *WorkflowResult) {
	// Check for race condition detection.
	if strings.Contains(logContent, "-race") {
		status := taskSuccess
		if strings.Contains(logContent, "DATA RACE") {
			status = taskFailed
		}

		result.TaskResults["Race Condition Detection"] = TaskResult{Name: "Race Condition Detection", Status: status}
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

	analysis.WriteString("## Execution Metrics\n\n")
	analysis.WriteString(fmt.Sprintf("- **Start Time:** %s\n", result.StartTime.Format("2006-01-02 15:04:05")))
	analysis.WriteString(fmt.Sprintf("- **End Time:** %s\n", result.EndTime.Format("2006-01-02 15:04:05")))
	analysis.WriteString(fmt.Sprintf("- **Duration:** %v\n", result.Duration.Round(time.Millisecond)))
	analysis.WriteString(fmt.Sprintf("- **CPU Time:** %v (approximated)\n", result.CPUTime.Round(time.Millisecond)))
	analysis.WriteString(fmt.Sprintf("- **Memory Usage:** %.2f MB\n\n", float64(result.MemoryUsage)/bytesPerMB))

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
