// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewProgressDisplay(t *testing.T) {
	t.Parallel()

	config := &Config{NoColor: true, Quiet: false, Verbose: true}
	p := NewProgressDisplay(config)
	require.NotNil(t, p)
	require.True(t, p.noColor)
	require.False(t, p.quiet)
	require.True(t, p.verbose)
}

func TestProgressDisplay_SetTotalSteps(t *testing.T) {
	t.Parallel()

	p := NewProgressDisplay(&Config{NoColor: true})
	p.SetTotalSteps(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, p.stepTotal)
}

func TestProgressDisplay_StartStep(t *testing.T) {
	t.Parallel()

	t.Run("normal mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.SetTotalSteps(3)
		p.StartStep("test step")
		require.Contains(t, buf.String(), "[1/3] test step...")
	})

	t.Run("quiet mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true, Quiet: true})
		p.writer = &buf
		p.StartStep("test step")
		require.Empty(t, buf.String())
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.StartStep("test step")
		require.Contains(t, buf.String(), "test step...")
	})

	t.Run("no total steps", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.StartStep("test step")
		require.Contains(t, buf.String(), "test step...")
		require.NotContains(t, buf.String(), "[")
	})
}

func TestProgressDisplay_CompleteStep(t *testing.T) {
	t.Parallel()

	t.Run("normal mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.CompleteStep("done step")
		require.Contains(t, buf.String(), "[OK] done step")
	})

	t.Run("quiet mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true, Quiet: true})
		p.writer = &buf
		p.CompleteStep("done step")
		require.Empty(t, buf.String())
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.CompleteStep("done step")
		require.Contains(t, buf.String(), "done step")
	})
}

func TestProgressDisplay_FailStep(t *testing.T) {
	t.Parallel()

	t.Run("no color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.FailStep("fail step", errors.New("test error"))
		require.Contains(t, buf.String(), "[FAIL] fail step: test error")
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.FailStep("fail step", errors.New("test error"))
		require.Contains(t, buf.String(), "fail step: test error")
	})
}

func TestProgressDisplay_SkipStep(t *testing.T) {
	t.Parallel()

	t.Run("normal mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.SkipStep("skip step", "not needed")
		require.Contains(t, buf.String(), "[SKIP] skip step: not needed")
	})

	t.Run("quiet mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true, Quiet: true})
		p.writer = &buf
		p.SkipStep("skip step", "not needed")
		require.Empty(t, buf.String())
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.SkipStep("skip step", "reason")
		require.Contains(t, buf.String(), "skip step: reason")
	})
}

func TestProgressDisplay_Info(t *testing.T) {
	t.Parallel()

	t.Run("normal mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.Info("info %s", "msg")
		require.Contains(t, buf.String(), "[INFO] info msg")
	})

	t.Run("quiet mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true, Quiet: true})
		p.writer = &buf
		p.Info("info msg")
		require.Empty(t, buf.String())
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.Info("info msg")
		require.Contains(t, buf.String(), "info msg")
	})
}

func TestProgressDisplay_Debug(t *testing.T) {
	t.Parallel()

	t.Run("verbose mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true, Verbose: true})
		p.writer = &buf
		p.Debug("debug %s", "msg")
		require.Contains(t, buf.String(), "[DEBUG] debug msg")
	})

	t.Run("non-verbose mode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true, Verbose: false})
		p.writer = &buf
		p.Debug("debug msg")
		require.Empty(t, buf.String())
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false, Verbose: true})
		p.writer = &buf
		p.Debug("debug msg")
		require.Contains(t, buf.String(), "debug msg")
	})
}

func TestProgressDisplay_Warn(t *testing.T) {
	t.Parallel()

	t.Run("no color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.Warn("warn %s", "msg")
		require.Contains(t, buf.String(), "[WARN] warn msg")
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.Warn("warn msg")
		require.Contains(t, buf.String(), "warn msg")
	})
}

func TestProgressDisplay_Error(t *testing.T) {
	t.Parallel()

	t.Run("no color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.Error("error %s", "msg")
		require.Contains(t, buf.String(), "[ERROR] error msg")
	})

	t.Run("with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.Error("error msg")
		require.Contains(t, buf.String(), "error msg")
	})
}

func TestProgressDisplay_PrintSummary(t *testing.T) {
	t.Parallel()

	t.Run("success no color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.PrintSummary(&DemoResult{Success: true, TotalSteps: 2, PassedSteps: 2})

		output := buf.String()
		require.Contains(t, output, "Demo Summary")
		require.Contains(t, output, "[SUCCESS]")
	})

	t.Run("success with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.PrintSummary(&DemoResult{Success: true, TotalSteps: 2, PassedSteps: 2})

		output := buf.String()
		require.Contains(t, output, "Demo Summary")
		require.Contains(t, output, "successfully")
	})

	t.Run("partial failure no color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.PrintSummary(&DemoResult{
			Success:     false,
			TotalSteps:  3,
			PassedSteps: 2,
			FailedSteps: 1,
			Errors:      []DemoError{{Phase: cryptoutilSharedMagic.KMSServiceName, Step: "s1", Message: "err1"}},
		})

		output := buf.String()
		require.Contains(t, output, "[PARTIAL]")
		require.Contains(t, output, "Failed Steps:")
	})

	t.Run("partial failure with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.PrintSummary(&DemoResult{
			Success:     false,
			TotalSteps:  3,
			PassedSteps: 2,
			FailedSteps: 1,
			Errors:      []DemoError{{Phase: cryptoutilSharedMagic.KMSServiceName, Step: "s1", Message: "err1"}},
		})

		output := buf.String()
		require.Contains(t, output, "some failures")
	})

	t.Run("total failure no color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: true})
		p.writer = &buf
		p.PrintSummary(&DemoResult{
			Success:     false,
			TotalSteps:  2,
			PassedSteps: 0,
			FailedSteps: 2,
			Errors:      []DemoError{{Phase: cryptoutilSharedMagic.KMSServiceName, Step: "s1", Message: "err1"}},
		})

		output := buf.String()
		require.Contains(t, output, "[FAILURE]")
	})

	t.Run("total failure with color", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := NewProgressDisplay(&Config{NoColor: false})
		p.writer = &buf
		p.PrintSummary(&DemoResult{
			Success:     false,
			TotalSteps:  2,
			PassedSteps: 0,
			FailedSteps: 2,
			Errors:      []DemoError{{Phase: cryptoutilSharedMagic.KMSServiceName, Step: "s1", Message: "err1"}},
		})

		output := buf.String()
		require.Contains(t, output, "Demo failed")
	})
}

func TestSpinner_StartStop(t *testing.T) {
	t.Parallel()

	spinner := NewSpinner()
	require.NotNil(t, spinner)

	// Start and immediately stop.
	spinner.Start("loading")
	time.Sleep(cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond)
	spinner.Stop()

	// Double stop should not panic.
	spinner.Stop()
}

func TestSpinner_DoubleStart(t *testing.T) {
	t.Parallel()

	spinner := NewSpinner()

	spinner.Start("loading")
	spinner.Start("loading again") // Should be no-op since already running.

	time.Sleep(cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond)
	spinner.Stop()
}
