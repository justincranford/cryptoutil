// Copyright (c) 2025 Justin Cranford
//
//

// Package workflow provides workflow execution and orchestration utilities.
package workflow

import (
	"bufio"
	"context"
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

// setupCmdPipes is a package-level function variable for creating stdout and stderr
// pipes from an exec.Cmd. It is defined as a variable to allow injection during
// testing to exercise the error handling paths that are otherwise unreachable.
var setupCmdPipes = func(cmd *exec.Cmd) (io.ReadCloser, io.ReadCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	return stdout, stderr, nil
}

// doCloseFile is a package-level function variable for closing a file.
// It is defined as a variable to allow injection during testing to exercise
// the close error handling path that is otherwise unreachable.
var doCloseFile = func(f *os.File) error {
	return f.Close()
}

// WorkflowConfig defines a GitHub Actions workflow that can be run locally with act.
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
		if err := doCloseFile(workflowLog); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close workflow log: %v\n", err)
		}
	}()

	// Execute act command.
	cmd := exec.CommandContext(context.Background(), actPath, args...) //nolint:gosec // User-controlled input is intentional for local testing

	// Setup stdout and stderr pipes for dual logging.
	stdoutPipe, stderrPipe, err := setupCmdPipes(cmd)
	if err != nil {
		errMsg := fmt.Sprintf("%sError creating command pipes: %v%s\n", cryptoutilSharedMagic.ColorRed, err, cryptoutilSharedMagic.ColorReset)
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

			analysis.WriteString(fmt.Sprintf("| %s | %s | %s |\n", task.Name, statusBadge(task.Status == cryptoutilSharedMagic.TaskSuccess), artifacts))
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
		return cryptoutilSharedMagic.StatusSuccess
	}

	return cryptoutilSharedMagic.StatusFailed
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
