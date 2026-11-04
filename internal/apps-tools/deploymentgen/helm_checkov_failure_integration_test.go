// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration && checkov

package deploymentgen

import (
	json "encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckovScanFailureMessagePreservesProcessDiagnostics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		subject string
		err     error
		stdout  []byte
		want    []string
	}{
		{
			name:    "process exit preserves stderr and stdout",
			subject: testCheckovDiagnosticSubject,
			err:     &exec.ExitError{Stderr: []byte("Checkov rejected the manifest")},
			stdout:  []byte(`{"check_type":"kubernetes"}`),
			want: []string{
				"Checkov Kubernetes scan rejected release 12",
				"stderr: Checkov rejected the manifest",
				`stdout: {"check_type":"kubernetes"}`,
			},
		},
		{
			name:    "process start failure preserves output",
			subject: testCheckovDiagnosticSubject,
			err:     errors.New("Checkov executable missing"),
			stdout:  []byte("partial output"),
			want: []string{
				"Checkov Kubernetes scan failed for release 12",
				"Checkov executable missing",
				"stdout: partial output",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			message := checkovScanFailureMessage(test.subject, test.err, test.stdout)
			for _, expected := range test.want {
				require.Contains(t, message, expected)
			}
		})
	}
}

func checkovScanFailureMessage(subject string, scanErr error, output []byte) string {
	var exitErr *exec.ExitError
	if errors.As(scanErr, &exitErr) {
		return fmt.Sprintf(
			"Checkov Kubernetes scan rejected %s with exit code %d: %v; stderr: %s; stdout: %s",
			subject, exitErr.ExitCode(), scanErr, exitErr.Stderr, output,
		)
	}

	return fmt.Sprintf("Checkov Kubernetes scan failed for %s: %v; stdout: %s", subject, scanErr, output)
}

func requireCheckovJSONReport(t *testing.T, output []byte) checkovJSONReport {
	t.Helper()

	var report checkovJSONReport
	require.NoError(t, json.Unmarshal(output, &report), "Checkov must produce valid JSON")

	return report
}

func checkovFailedCheckIDs(report checkovJSONReport) []string {
	failedCheckIDs := make([]string, 0, len(report.Results.FailedChecks))
	for _, failedCheck := range report.Results.FailedChecks {
		failedCheckIDs = append(failedCheckIDs, failedCheck.CheckID)
	}

	return failedCheckIDs
}
