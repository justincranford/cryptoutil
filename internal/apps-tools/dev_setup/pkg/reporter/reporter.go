package reporter

import (
	"fmt"
	"io"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DepResult represents the result of checking/installing a single dependency.
type DepResult struct {
	Name          string
	Version       string
	ActualVersion string
	Status        string
	Error         error
}

// GroupResult represents the result of processing a dependency group.
type GroupResult struct {
	Name         string
	Dependencies []*DepResult
	HasErrors    bool
}

// Summary contains the overall setup result.
type Summary struct {
	Groups    map[string]*GroupResult
	HasErrors bool
}

// Reporter defines the interface for reporting setup results.
type Reporter interface {
	// ReportSummary outputs the setup summary
	ReportSummary(summary *Summary) error
}

// ConsoleReporter prints results to console.
type ConsoleReporter struct {
	stdout io.Writer
	stderr io.Writer
}

// NewConsoleReporter creates a console reporter.
func NewConsoleReporter(stdout, stderr io.Writer) Reporter {
	return &ConsoleReporter{
		stdout: stdout,
		stderr: stderr,
	}
}

// ReportSummary outputs setup results to console.
func (cr *ConsoleReporter) ReportSummary(summary *Summary) error {
	_, _ = fmt.Fprintf(cr.stdout, "\n=== Development Environment Setup ===\n\n")

	for _, group := range summary.Groups {
		if group == nil {
			continue
		}

		status := cryptoutilSharedMagic.DevSetupStatusIconSuccess
		if group.HasErrors {
			status = cryptoutilSharedMagic.DevSetupStatusIconFailure
		}

		_, _ = fmt.Fprintf(cr.stdout, "%s %s\n", status, group.Name)

		for _, dep := range group.Dependencies {
			icon := cryptoutilSharedMagic.DevSetupStatusIconSuccess
			if dep.Status != cryptoutilSharedMagic.DevSetupStatusInstalled {
				icon = cryptoutilSharedMagic.DevSetupStatusIconFailure
			}

			versionStr := ""
			if dep.ActualVersion != "" {
				versionStr = fmt.Sprintf(" (%s)", dep.ActualVersion)
			}

			_, _ = fmt.Fprintf(cr.stdout, "  %s %s%s\n", icon, dep.Name, versionStr)

			if dep.Error != nil {
				_, _ = fmt.Fprintf(cr.stderr, "    Error: %v\n", dep.Error)
			}
		}

		_, _ = fmt.Fprintf(cr.stdout, "\n")
	}

	if summary.HasErrors {
		_, _ = fmt.Fprintf(cr.stderr, "ÃƒÂ¢Ã…Â¡Ã‚Â  Some dependencies failed to install. Please check the errors above.\n")
	} else {
		_, _ = fmt.Fprintf(cr.stdout, "ÃƒÂ¢Ã…â€œÃ¢â‚¬Å“ All dependencies installed successfully!\n")
	}

	return nil
}
