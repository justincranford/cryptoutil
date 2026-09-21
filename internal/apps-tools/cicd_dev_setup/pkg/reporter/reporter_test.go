package reporter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestReportSummary_Happy(t *testing.T) {
	t.Parallel()
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
		require.NoError(t, err)

		output := stdout.String()
		require.Contains(t, output, cryptoutilSharedMagic.DevSetupStatusIconSuccess, "expected success indicator in output")
		require.Contains(t, output, cryptoutilSharedMagic.DevSetupTestToolNamePython, "expected tool name in output")
		require.Contains(t, output, "All dependencies installed successfully")
	})
}

func TestReportSummary_Sad(t *testing.T) {
	t.Parallel()

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
			if tc.expectError {
				require.Error(t, err)
			}

			stderrOutput := stderr.String()
			if tc.expectWarning {
				hasIndicator := strings.Contains(stderrOutput, "ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã¢â‚¬Â¦Ãƒâ€šÃ‚Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â ") || strings.Contains(stderrOutput, "Error:")
				require.True(t, hasIndicator, "expected warning/error in stderr output")
			}
		})
	}
}

func TestReportSummary_Format(t *testing.T) {
	t.Parallel()
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
		require.NoError(t, err)

		output := stdout.String()

		// Check for header
		require.Contains(t, output, "Development Environment Setup")

		// Check for group names
		require.Contains(t, output, cryptoutilSharedMagic.DevSetupTestGroupPythonRuntime)
		require.Contains(t, output, cryptoutilSharedMagic.DevSetupTestGroupGoToolchain)

		// Check for tool names
		require.Contains(t, output, cryptoutilSharedMagic.DevSetupTestToolNamePython)
		require.Contains(t, output, cryptoutilSharedMagic.DevSetupTestToolNameGo)

		// Check for versions
		hasVersion := strings.Contains(output, cryptoutilSharedMagic.DevSetupTestVersionPythonRequired) || strings.Contains(output, cryptoutilSharedMagic.DevSetupTestVersionGoRequired)
		require.True(t, hasVersion, "expected version info in output")
	})
}

func TestReportSummary_NilGroups(t *testing.T) {
	t.Parallel()
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
		require.NoError(t, err)

		// Should not panic and should report what's available
		output := stdout.String()
		require.Contains(t, output, cryptoutilSharedMagic.DevSetupTestGroupNamePython, "expected Python group in output")
	})
}
