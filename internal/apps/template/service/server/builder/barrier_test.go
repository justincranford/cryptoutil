// Copyright (c) 2025 Justin Cranford
// SPDX-License-Identifier: MIT

package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBarrierConfig(t *testing.T) {
	t.Parallel()

	config := NewBarrierConfig()

	require.NotNil(t, config)
	require.True(t, config.EnableRotationEndpoints)
	require.True(t, config.EnableStatusEndpoints)
}

func TestBarrierConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *BarrierConfig
		wantErr bool
	}{
		{
			name:    "default config",
			config:  NewBarrierConfig(),
			wantErr: false,
		},
		{
			name:    "custom config",
			config:  &BarrierConfig{EnableRotationEndpoints: false, EnableStatusEndpoints: false},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBarrierConfig_WithRotationEndpoints(t *testing.T) {
	t.Parallel()

	config := NewBarrierConfig()
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

	config := NewBarrierConfig()
	require.True(t, config.EnableStatusEndpoints)

	result := config.WithStatusEndpoints(false)

	require.Same(t, config, result)
	require.False(t, config.EnableStatusEndpoints)

	result = config.WithStatusEndpoints(true)

	require.Same(t, config, result)
	require.True(t, config.EnableStatusEndpoints)
}

func TestBarrierConfig_FluentChaining(t *testing.T) {
	t.Parallel()

	config := NewBarrierConfig().
		WithRotationEndpoints(false).
		WithStatusEndpoints(false)

	require.False(t, config.EnableRotationEndpoints)
	require.False(t, config.EnableStatusEndpoints)
}

func TestErrBarrierConfigRequired(t *testing.T) {
	t.Parallel()

	require.NotNil(t, ErrBarrierConfigRequired)
	require.Equal(t, "barrier config is required", ErrBarrierConfigRequired.Error())
}

func TestWithBarrierConfig_NilConfig(t *testing.T) {
	t.Parallel()

	builder := &ServerBuilder{}
	result := builder.WithBarrierConfig(nil)

	require.Same(t, builder, result)
	require.NoError(t, builder.err)
	require.Nil(t, builder.barrierConfig)
}

func TestWithBarrierConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	config := NewBarrierConfig()
	builder := &ServerBuilder{}
	result := builder.WithBarrierConfig(config)

	require.Same(t, builder, result)
	require.NoError(t, builder.err)
	require.Same(t, config, builder.barrierConfig)
}

func TestWithBarrierConfig_ErrorAccumulation(t *testing.T) {
	t.Parallel()

	builder := &ServerBuilder{err: ErrBarrierConfigRequired}
	result := builder.WithBarrierConfig(NewBarrierConfig())

	require.Same(t, builder, result)
	require.ErrorIs(t, builder.err, ErrBarrierConfigRequired)
	require.Nil(t, builder.barrierConfig)
}
