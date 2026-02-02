// Copyright (c) 2025 Justin Cranford
//

package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDefaultBarrierConfig(t *testing.T) {
	t.Parallel()

	config := NewDefaultBarrierConfig()

	require.NotNil(t, config)
	require.Equal(t, BarrierModeTemplate, config.Mode)
	require.True(t, config.EnableRotationEndpoints)
	require.True(t, config.EnableStatusEndpoints)
}

func TestNewSharedBarrierConfig(t *testing.T) {
	t.Parallel()

	config := NewSharedBarrierConfig()

	require.NotNil(t, config)
	require.Equal(t, BarrierModeShared, config.Mode)
	require.False(t, config.EnableRotationEndpoints)
	require.False(t, config.EnableStatusEndpoints)
}

func TestNewDisabledBarrierConfig(t *testing.T) {
	t.Parallel()

	config := NewDisabledBarrierConfig()

	require.NotNil(t, config)
	require.Equal(t, BarrierModeDisabled, config.Mode)
	require.False(t, config.EnableRotationEndpoints)
	require.False(t, config.EnableStatusEndpoints)
}

func TestBarrierConfig_Validate_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *BarrierConfig
	}{
		{
			name:   "template mode",
			config: NewDefaultBarrierConfig(),
		},
		{
			name:   "shared mode",
			config: NewSharedBarrierConfig(),
		},
		{
			name:   "disabled mode",
			config: NewDisabledBarrierConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			require.NoError(t, err)
		})
	}
}

func TestBarrierConfig_Validate_EmptyMode(t *testing.T) {
	t.Parallel()

	config := &BarrierConfig{
		Mode: "",
	}

	err := config.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrBarrierModeRequired)
}

func TestBarrierConfig_Validate_InvalidMode(t *testing.T) {
	t.Parallel()

	config := &BarrierConfig{
		Mode: "invalid-mode",
	}

	err := config.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid barrier mode")
}

func TestBarrierConfig_WithMode(t *testing.T) {
	t.Parallel()

	config := NewDefaultBarrierConfig()
	require.Equal(t, BarrierModeTemplate, config.Mode)

	result := config.WithMode(BarrierModeShared)

	require.Same(t, config, result)
	require.Equal(t, BarrierModeShared, config.Mode)
}

func TestBarrierConfig_WithRotationEndpoints(t *testing.T) {
	t.Parallel()

	config := NewDefaultBarrierConfig()
	require.True(t, config.EnableRotationEndpoints)

	result := config.WithRotationEndpoints(false)

	require.Same(t, config, result)
	require.False(t, config.EnableRotationEndpoints)

	result = config.WithRotationEndpoints(true)

	require.Same(t, config, result)
	require.True(t, config.EnableRotationEndpoints)
}

func TestBarrierConfig_WithStatusEndpoints(t *testing.T) {
	t.Parallel()

	config := NewDefaultBarrierConfig()
	require.True(t, config.EnableStatusEndpoints)

	result := config.WithStatusEndpoints(false)

	require.Same(t, config, result)
	require.False(t, config.EnableStatusEndpoints)

	result = config.WithStatusEndpoints(true)

	require.Same(t, config, result)
	require.True(t, config.EnableStatusEndpoints)
}

func TestBarrierConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *BarrierConfig
		expected bool
	}{
		{
			name:     "template mode is enabled",
			config:   NewDefaultBarrierConfig(),
			expected: true,
		},
		{
			name:     "shared mode is enabled",
			config:   NewSharedBarrierConfig(),
			expected: true,
		},
		{
			name:     "disabled mode is not enabled",
			config:   NewDisabledBarrierConfig(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.config.IsEnabled())
		})
	}
}

func TestBarrierConfig_FluentChaining(t *testing.T) {
	t.Parallel()

	config := NewDefaultBarrierConfig().
		WithMode(BarrierModeShared).
		WithRotationEndpoints(false).
		WithStatusEndpoints(false)

	require.Equal(t, BarrierModeShared, config.Mode)
	require.False(t, config.EnableRotationEndpoints)
	require.False(t, config.EnableStatusEndpoints)
}

func TestBarrierModeConstants(t *testing.T) {
	t.Parallel()

	// Verify constant values don't change unexpectedly.
	require.Equal(t, BarrierMode("template"), BarrierModeTemplate)
	require.Equal(t, BarrierMode("shared"), BarrierModeShared)
	require.Equal(t, BarrierMode("disabled"), BarrierModeDisabled)
}

func TestErrBarrierModeRequired(t *testing.T) {
	t.Parallel()

	require.NotNil(t, ErrBarrierModeRequired)
	require.Equal(t, "barrier mode is required", ErrBarrierModeRequired.Error())
}

func TestWithBarrierConfig_NilConfig(t *testing.T) {
	t.Parallel()

	// ServerBuilder should accept nil config without error (uses default).
	builder := &ServerBuilder{}
	result := builder.WithBarrierConfig(nil)

	require.Same(t, builder, result)
	require.NoError(t, builder.err)
	require.Nil(t, builder.barrierConfig)
}

func TestWithBarrierConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *BarrierConfig
	}{
		{
			name:   "template mode",
			config: NewDefaultBarrierConfig(),
		},
		{
			name:   "shared mode",
			config: NewSharedBarrierConfig(),
		},
		{
			name:   "disabled mode",
			config: NewDisabledBarrierConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := &ServerBuilder{}
			result := builder.WithBarrierConfig(tt.config)

			require.Same(t, builder, result)
			require.NoError(t, builder.err)
			require.Same(t, tt.config, builder.barrierConfig)
		})
	}
}

func TestWithBarrierConfig_InvalidConfig(t *testing.T) {
	t.Parallel()

	config := &BarrierConfig{
		Mode: "invalid-mode",
	}

	builder := &ServerBuilder{}
	result := builder.WithBarrierConfig(config)

	require.Same(t, builder, result)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "invalid barrier config")
}

func TestWithBarrierConfig_ErrorAccumulation(t *testing.T) {
	t.Parallel()

	// Builder with existing error should not process new config.
	builder := &ServerBuilder{err: ErrBarrierModeRequired}
	result := builder.WithBarrierConfig(NewDefaultBarrierConfig())

	require.Same(t, builder, result)
	require.ErrorIs(t, builder.err, ErrBarrierModeRequired)
	require.Nil(t, builder.barrierConfig)
}
