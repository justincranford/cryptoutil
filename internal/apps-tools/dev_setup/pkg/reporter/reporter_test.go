package reporter

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"strings"
	"testing"
)

func TestReportSummary_Happy(t *testing.T) {
	t.Run("All dependencies installed successfully", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		reporter := NewConsoleReporter(stdout, stderr)

		summary := &Summary{
			Groups: map[string]*GroupResult{
				cryptoutilSharedMagic.DevSetupTestGroupNamePython: {
					Name: cryptoutilSharedMagic.DevSetupTestGroupNamePython,
					Dependencies: []*DepResult{
						{
							Name:          cryptoutilSharedMagic.DevSetupTestToolNamePython,
							Version:       cryptoutilSharedMagic.DevSetupTestVersionPythonRequired,
							ActualVersion: cryptoutilSharedMagic.DevSetupTestVersionPythonActual,
							Status:        cryptoutilSharedMagic.DevSetupStatusInstalled,
						},
					},
					HasErrors: false,
				},
			},
			HasErrors: false,
		}

		err := reporter.ReportSummary(summary)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, cryptoutilSharedMagic.DevSetupStatusIconSuccess) {
			t.Errorf("expected success indicator in output")
		}

		if !strings.Contains(output, cryptoutilSharedMagic.DevSetupTestToolNamePython) {
			t.Errorf("expected tool name in output")
		}

		if !strings.Contains(output, "All dependencies installed successfully") {
			t.Errorf("expected success message in output")
		}
	})
}

func TestReportSummary_Sad(t *testing.T) {
	tests := []struct {
		name          string
		summary       *Summary
		expectError   bool
		expectWarning bool
	}{
		{
			name: "Dependency check failed",
			summary: &Summary{
				Groups: map[string]*GroupResult{
					cryptoutilSharedMagic.DevSetupTestGroupNameGo: {
						Name: cryptoutilSharedMagic.DevSetupTestGroupNameGo,
						Dependencies: []*DepResult{
							{
								Name:   cryptoutilSharedMagic.DevSetupTestToolNameGo,
								Status: cryptoutilSharedMagic.DevSetupStatusCheckFailed,
								Error:  fmt.Errorf("command not found"),
							},
						},
						HasErrors: true,
					},
				},
				HasErrors: true,
			},
			expectError:   false,
			expectWarning: true,
		},
		{
			name: "Installation failed",
			summary: &Summary{
				Groups: map[string]*GroupResult{
					cryptoutilSharedMagic.DevSetupTestGroupNamePython: {
						Name: cryptoutilSharedMagic.DevSetupTestGroupNamePython,
						Dependencies: []*DepResult{
							{
								Name:   "pre-commit",
								Status: cryptoutilSharedMagic.DevSetupStatusInstallFailed,
								Error:  fmt.Errorf("permission denied"),
							},
						},
						HasErrors: true,
					},
				},
				HasErrors: true,
			},
			expectError:   false,
			expectWarning: true,
		},
		{
			name: "Verification failed",
			summary: &Summary{
				Groups: map[string]*GroupResult{
					cryptoutilSharedMagic.DevSetupTestGroupNameGo: {
						Name: cryptoutilSharedMagic.DevSetupTestGroupNameGo,
						Dependencies: []*DepResult{
							{
								Name:   "golangci-lint",
								Status: cryptoutilSharedMagic.DevSetupStatusVerifyFailed,
								Error:  fmt.Errorf("tool not found after installation"),
							},
						},
						HasErrors: true,
					},
				},
				HasErrors: true,
			},
			expectError:   false,
			expectWarning: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			reporter := NewConsoleReporter(stdout, stderr)

			err := reporter.ReportSummary(tc.summary)
			if tc.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}

			stderrOutput := stderr.String()
			if tc.expectWarning && !strings.Contains(stderrOutput, "ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã¢â‚¬Â¦Ãƒâ€šÃ‚Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â ") && !strings.Contains(stderrOutput, "Error:") {
				t.Errorf("expected warning/error in stderr output")
			}
		})
	}
}

func TestReportSummary_Format(t *testing.T) {
	t.Run("Output contains all expected sections", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		reporter := NewConsoleReporter(stdout, stderr)

		summary := &Summary{
			Groups: map[string]*GroupResult{
				cryptoutilSharedMagic.DevSetupTestGroupNamePythonRuntime: {
					Name: cryptoutilSharedMagic.DevSetupTestGroupNamePythonRuntime,
					Dependencies: []*DepResult{
						{
							Name:          cryptoutilSharedMagic.DevSetupTestToolNamePython,
							Version:       cryptoutilSharedMagic.DevSetupTestVersionPythonRequired,
							ActualVersion: cryptoutilSharedMagic.DevSetupTestVersionPythonActual,
							Status:        cryptoutilSharedMagic.DevSetupStatusInstalled,
						},
						{
							Name:          "pip",
							Version:       cryptoutilSharedMagic.DevSetupTestVersionPythonNext,
							ActualVersion: cryptoutilSharedMagic.DevSetupTestVersionPythonNext,
							Status:        cryptoutilSharedMagic.DevSetupStatusInstalled,
						},
					},
					HasErrors: false,
				},
				cryptoutilSharedMagic.DevSetupTestGroupNameGoToolchain: {
					Name: cryptoutilSharedMagic.DevSetupTestGroupNameGoToolchain,
					Dependencies: []*DepResult{
						{
							Name:          cryptoutilSharedMagic.DevSetupTestToolNameGo,
							Version:       cryptoutilSharedMagic.DevSetupTestVersionGoRequired,
							ActualVersion: cryptoutilSharedMagic.DevSetupTestVersionGoActual,
							Status:        cryptoutilSharedMagic.DevSetupStatusInstalled,
						},
					},
					HasErrors: false,
				},
			},
			HasErrors: false,
		}

		err := reporter.ReportSummary(summary)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := stdout.String()

		// Check for header
		if !strings.Contains(output, "Development Environment Setup") {
			t.Errorf("expected header in output")
		}

		// Check for group names
		if !strings.Contains(output, "Python Runtime") {
			t.Errorf("expected 'Python Runtime' group in output")
		}

		if !strings.Contains(output, "Go Toolchain") {
			t.Errorf("expected 'Go Toolchain' group in output")
		}

		// Check for tool names
		if !strings.Contains(output, "python") {
			t.Errorf("expected 'python' tool in output")
		}

		if !strings.Contains(output, "go") {
			t.Errorf("expected 'go' tool in output")
		}

		// Check for versions
		if !strings.Contains(output, "3.14.0") && !strings.Contains(output, "1.26.1") {
			t.Errorf("expected version info in output")
		}
	})
}

func TestReportSummary_NilGroups(t *testing.T) {
	t.Run("Handles nil groups gracefully", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		reporter := NewConsoleReporter(stdout, stderr)

		summary := &Summary{
			Groups: map[string]*GroupResult{
				"valid": nil, // Nil group
				cryptoutilSharedMagic.DevSetupTestGroupNamePython: {
					Name: cryptoutilSharedMagic.DevSetupTestGroupNamePython,
					Dependencies: []*DepResult{
						{
							Name:          cryptoutilSharedMagic.DevSetupTestToolNamePython,
							ActualVersion: cryptoutilSharedMagic.DevSetupTestVersionPythonActual,
							Status:        cryptoutilSharedMagic.DevSetupStatusInstalled,
						},
					},
					HasErrors: false,
				},
			},
			HasErrors: false,
		}

		err := reporter.ReportSummary(summary)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should not panic and should report what's available
		output := stdout.String()
		if !strings.Contains(output, "Python") {
			t.Errorf("expected Python group in output")
		}
	})
}
