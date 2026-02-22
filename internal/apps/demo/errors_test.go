// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDemoResult_ExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   DemoResult
		expected int
	}{
		{
			name:     "success",
			result:   DemoResult{Success: true, TotalSteps: 3, PassedSteps: 3},
			expected: ExitSuccess,
		},
		{
			name:     "partial failure",
			result:   DemoResult{Success: false, TotalSteps: 3, PassedSteps: 2, FailedSteps: 1},
			expected: ExitPartialFailure,
		},
		{
			name:     "total failure",
			result:   DemoResult{Success: false, TotalSteps: 3, PassedSteps: 0, FailedSteps: 3},
			expected: ExitFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.expected, tt.result.ExitCode())
		})
	}
}

func TestDemoError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      DemoError
		expected string
	}{
		{
			name:     "basic error",
			err:      DemoError{Phase: "kms", Step: "config", Message: "parse failed"},
			expected: "[kms/config] parse failed",
		},
		{
			name:     "error with details",
			err:      DemoError{Phase: "kms", Step: "config", Message: "parse failed", Details: "missing field"},
			expected: "[kms/config] parse failed: missing field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestDemoError_Unwrap(t *testing.T) {
	t.Parallel()

	t.Run("nil cause", func(t *testing.T) {
		t.Parallel()

		err := DemoError{Phase: "kms", Step: "config", Message: "failed"}
		require.NoError(t, err.Unwrap())
	})

	t.Run("with cause", func(t *testing.T) {
		t.Parallel()

		cause := DemoError{Phase: "inner", Step: "step", Message: "root cause"}
		err := DemoError{Phase: "kms", Step: "config", Message: "failed", Cause: &cause}
		unwrapped := err.Unwrap()
		require.Error(t, unwrapped)
		require.Contains(t, unwrapped.Error(), "root cause")
	})
}

func TestNewDemoError(t *testing.T) {
	t.Parallel()

	err := NewDemoError("identity", "login", "auth failed")
	require.Equal(t, "identity", err.Phase)
	require.Equal(t, "login", err.Step)
	require.Equal(t, "auth failed", err.Message)
}

func TestDemoError_WithDetails(t *testing.T) {
	t.Parallel()

	err := NewDemoError("kms", "encrypt", "failed").WithDetails("invalid key")
	require.Equal(t, "invalid key", err.Details)
	require.Contains(t, err.Error(), "invalid key")
}

func TestDemoError_WithCause(t *testing.T) {
	t.Parallel()

	t.Run("with DemoError cause", func(t *testing.T) {
		t.Parallel()

		cause := NewDemoError("inner", "step", "inner error")
		err := NewDemoError("outer", "step", "outer error").WithCause(cause)
		require.NotNil(t, err.Cause)
		require.Equal(t, "inner error", err.Cause.Message)
	})

	t.Run("with standard error cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("standard error")
		err := NewDemoError("outer", "step", "outer error").WithCause(cause)
		require.NotNil(t, err.Cause)
		require.Equal(t, "standard error", err.Cause.Message)
	})

	t.Run("with nil cause", func(t *testing.T) {
		t.Parallel()

		err := NewDemoError("outer", "step", "outer error").WithCause(nil)
		require.Nil(t, err.Cause)
	})
}

func TestErrorAggregator(t *testing.T) {
	t.Parallel()

	t.Run("empty aggregator", func(t *testing.T) {
		t.Parallel()

		agg := NewErrorAggregator("test")
		require.False(t, agg.HasErrors())
		require.Empty(t, agg.Errors())
		require.Equal(t, 0, agg.Count())
	})

	t.Run("add error", func(t *testing.T) {
		t.Parallel()

		agg := NewErrorAggregator("kms")
		agg.Add("step1", "failed", errors.New("cause"))
		require.True(t, agg.HasErrors())
		require.Len(t, agg.Errors(), 1)
		require.Equal(t, 1, agg.Count())
		require.Equal(t, "kms", agg.Errors()[0].Phase)
		require.Equal(t, "step1", agg.Errors()[0].Step)
	})

	t.Run("add error nil cause", func(t *testing.T) {
		t.Parallel()

		agg := NewErrorAggregator("kms")
		agg.Add("step1", "failed", nil)
		require.True(t, agg.HasErrors())
		require.Nil(t, agg.Errors()[0].Cause)
	})

	t.Run("add demo error directly", func(t *testing.T) {
		t.Parallel()

		agg := NewErrorAggregator("identity")
		de := NewDemoError("identity", "login", "auth failed")
		agg.AddError(de)
		require.Equal(t, 1, agg.Count())
		require.Equal(t, "auth failed", agg.Errors()[0].Message)
	})

	t.Run("multiple errors", func(t *testing.T) {
		t.Parallel()

		agg := NewErrorAggregator("integration")
		agg.Add("step1", "msg1", nil)
		agg.Add("step2", "msg2", errors.New("err"))
		agg.AddError(NewDemoError("integration", "step3", "msg3"))
		require.Equal(t, 3, agg.Count())
	})
}

func TestErrorAggregator_ToResult(t *testing.T) {
	t.Parallel()

	t.Run("no errors", func(t *testing.T) {
		t.Parallel()

		agg := NewErrorAggregator("kms")
		result := agg.ToResult(3, 0)
		require.True(t, result.Success)
		require.Equal(t, 3, result.TotalSteps)
		require.Equal(t, 3, result.PassedSteps)
		require.Equal(t, 0, result.FailedSteps)
		require.Equal(t, 0, result.SkippedSteps)
	})

	t.Run("with errors and skipped", func(t *testing.T) {
		t.Parallel()

		agg := NewErrorAggregator("kms")
		agg.Add("step2", "failed", nil)
		result := agg.ToResult(1, 2)
		require.False(t, result.Success)
		require.Equal(t, 4, result.TotalSteps)
		require.Equal(t, 1, result.PassedSteps)
		require.Equal(t, 1, result.FailedSteps)
		require.Equal(t, 2, result.SkippedSteps)
	})
}

func TestOutputFormatter_FormatResult(t *testing.T) {
	t.Parallel()

	result := &DemoResult{
		Success:      false,
		TotalSteps:   3,
		PassedSteps:  1,
		FailedSteps:  1,
		SkippedSteps: 1,
		DurationMS:   42,
		Errors:       []DemoError{{Phase: "kms", Step: "config", Message: "parse failed"}},
	}

	t.Run("json format", func(t *testing.T) {
		t.Parallel()

		formatter := NewOutputFormatter(OutputJSON)
		output := formatter.FormatResult(result)
		require.Contains(t, output, `"success": false`)
		require.Contains(t, output, `"total_steps": 3`)
		require.Contains(t, output, "parse failed")
	})

	t.Run("structured format", func(t *testing.T) {
		t.Parallel()

		formatter := NewOutputFormatter(OutputStructured)
		output := formatter.FormatResult(result)
		require.Contains(t, output, "level=info")
		require.Contains(t, output, "success=false")
		require.Contains(t, output, "level=error")
		require.Contains(t, output, "parse failed")
	})

	t.Run("human format", func(t *testing.T) {
		t.Parallel()

		formatter := NewOutputFormatter(OutputHuman)
		output := formatter.FormatResult(result)
		require.Contains(t, output, "Demo Summary")
		require.Contains(t, output, "Failed Steps")
		require.Contains(t, output, "parse failed")
	})

	t.Run("human format success", func(t *testing.T) {
		t.Parallel()

		successResult := &DemoResult{Success: true, TotalSteps: 2, PassedSteps: 2}
		formatter := NewOutputFormatter(OutputHuman)
		output := formatter.FormatResult(successResult)
		require.Contains(t, output, "successfully")
	})

	t.Run("human format all failed", func(t *testing.T) {
		t.Parallel()

		failResult := &DemoResult{
			Success:     false,
			TotalSteps:  2,
			PassedSteps: 0,
			FailedSteps: 2,
			Errors: []DemoError{
				{Phase: "kms", Step: "s1", Message: "err1"},
				{Phase: "kms", Step: "s2", Message: "err2"},
			},
		}
		formatter := NewOutputFormatter(OutputHuman)
		output := formatter.FormatResult(failResult)
		require.Contains(t, output, "Demo failed")
	})
}
