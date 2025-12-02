// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultScopeConfig(t *testing.T) {
	t.Parallel()

	config := DefaultScopeConfig()

	// Verify coarse scopes.
	require.Contains(t, config.CoarseScopes, "kms:admin")
	require.Contains(t, config.CoarseScopes, "kms:read")
	require.Contains(t, config.CoarseScopes, "kms:write")

	// Verify fine scopes.
	require.Contains(t, config.FineScopes, "kms:encrypt")
	require.Contains(t, config.FineScopes, "kms:decrypt")
	require.Contains(t, config.FineScopes, "kms:sign")

	// Verify hierarchy.
	require.NotEmpty(t, config.ScopeHierarchy["kms:admin"])
	require.Contains(t, config.ScopeHierarchy["kms:admin"], "kms:read")
	require.Contains(t, config.ScopeHierarchy["kms:admin"], "kms:write")
}

func TestScopeValidator_ExpandScopes(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name     string
		scopes   []string
		contains []string
	}{
		{
			name:     "admin expands to all",
			scopes:   []string{"kms:admin"},
			contains: []string{"kms:admin", "kms:read", "kms:write", "kms:encrypt", "kms:decrypt"},
		},
		{
			name:     "read expands to read operations",
			scopes:   []string{"kms:read"},
			contains: []string{"kms:read", "kms:pool:read", "kms:key:read"},
		},
		{
			name:     "write expands to write operations",
			scopes:   []string{"kms:write"},
			contains: []string{"kms:write", "kms:encrypt", "kms:decrypt", "kms:sign"},
		},
		{
			name:     "fine scope does not expand",
			scopes:   []string{"kms:encrypt"},
			contains: []string{"kms:encrypt"},
		},
		{
			name:     "multiple scopes combine",
			scopes:   []string{"kms:read", "kms:encrypt"},
			contains: []string{"kms:read", "kms:encrypt", "kms:pool:read"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expanded := validator.ExpandScopes(tc.scopes)

			for _, expected := range tc.contains {
				require.Contains(t, expanded, expected, "expanded scopes should contain %s", expected)
			}
		})
	}
}

func TestScopeValidator_HasScope(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name     string
		scopes   []string
		required string
		expected bool
	}{
		{
			name:     "admin has encrypt via hierarchy",
			scopes:   []string{"kms:admin"},
			required: "kms:encrypt",
			expected: true,
		},
		{
			name:     "read has pool:read via hierarchy",
			scopes:   []string{"kms:read"},
			required: "kms:pool:read",
			expected: true,
		},
		{
			name:     "read does not have encrypt",
			scopes:   []string{"kms:read"},
			required: "kms:encrypt",
			expected: false,
		},
		{
			name:     "direct scope match",
			scopes:   []string{"kms:encrypt"},
			required: "kms:encrypt",
			expected: true,
		},
		{
			name:     "missing scope",
			scopes:   []string{"kms:encrypt"},
			required: "kms:decrypt",
			expected: false,
		},
		{
			name:     "empty scopes",
			scopes:   []string{},
			required: "kms:read",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := validator.HasScope(tc.scopes, tc.required)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestScopeValidator_HasAnyScope(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name     string
		scopes   []string
		required []string
		expected bool
	}{
		{
			name:     "has one of required",
			scopes:   []string{"kms:read"},
			required: []string{"kms:read", "kms:write"},
			expected: true,
		},
		{
			name:     "has none of required",
			scopes:   []string{"kms:encrypt"},
			required: []string{"kms:read", "kms:admin"},
			expected: false,
		},
		{
			name:     "via hierarchy",
			scopes:   []string{"kms:admin"},
			required: []string{"kms:encrypt"},
			expected: true,
		},
		{
			name:     "empty required",
			scopes:   []string{"kms:read"},
			required: []string{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := validator.HasAnyScope(tc.scopes, tc.required)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestScopeValidator_HasAllScopes(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name     string
		scopes   []string
		required []string
		expected bool
	}{
		{
			name:     "has all required",
			scopes:   []string{"kms:read", "kms:write"},
			required: []string{"kms:read", "kms:write"},
			expected: true,
		},
		{
			name:     "missing one",
			scopes:   []string{"kms:read"},
			required: []string{"kms:read", "kms:write"},
			expected: false,
		},
		{
			name:     "admin has all via hierarchy",
			scopes:   []string{"kms:admin"},
			required: []string{"kms:read", "kms:write"},
			expected: true,
		},
		{
			name:     "empty required (vacuously true)",
			scopes:   []string{"kms:read"},
			required: []string{},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := validator.HasAllScopes(tc.scopes, tc.required)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestParseScopeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "space-separated",
			input:    "kms:read kms:write",
			expected: []string{"kms:read", "kms:write"},
		},
		{
			name:     "comma-separated",
			input:    "kms:read,kms:write",
			expected: []string{"kms:read", "kms:write"},
		},
		{
			name:     "mixed separators",
			input:    "kms:read kms:write,kms:admin",
			expected: []string{"kms:read", "kms:write", "kms:admin"},
		},
		{
			name:     "extra whitespace",
			input:    "  kms:read   kms:write  ",
			expected: []string{"kms:read", "kms:write"},
		},
		{
			name:     "single scope",
			input:    "kms:read",
			expected: []string{"kms:read"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ParseScopeString(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestScopeValidator_ValidateScope(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name    string
		scope   string
		wantErr bool
	}{
		{
			name:    "valid coarse scope",
			scope:   "kms:admin",
			wantErr: false,
		},
		{
			name:    "valid fine scope",
			scope:   "kms:encrypt",
			wantErr: false,
		},
		{
			name:    "unknown scope",
			scope:   "kms:unknown",
			wantErr: true,
		},
		{
			name:    "completely invalid",
			scope:   "invalid",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validator.ValidateScope(tc.scope)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestScopeValidator_ValidateScopes(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name    string
		scopes  []string
		wantErr bool
	}{
		{
			name:    "all valid",
			scopes:  []string{"kms:admin", "kms:read", "kms:encrypt"},
			wantErr: false,
		},
		{
			name:    "one invalid",
			scopes:  []string{"kms:read", "kms:unknown"},
			wantErr: true,
		},
		{
			name:    "empty scopes",
			scopes:  []string{},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validator.ValidateScopes(tc.scopes)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() context.Context
		expected []string
	}{
		{
			name: "from scope context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ScopeContextKey{}, []string{"kms:read"})
			},
			expected: []string{"kms:read"},
		},
		{
			name: "from JWT claims",
			setup: func() context.Context {
				claims := &JWTClaims{Scopes: []string{"kms:write"}}

				return context.WithValue(context.Background(), JWTContextKey{}, claims)
			},
			expected: []string{"kms:write"},
		},
		{
			name: "from service auth info",
			setup: func() context.Context {
				info := &ServiceAuthInfo{Scopes: []string{"kms:admin"}}

				return context.WithValue(context.Background(), ServiceAuthContextKey{}, info)
			},
			expected: []string{"kms:admin"},
		},
		{
			name: "no scopes in context",
			setup: func() context.Context {
				return context.Background()
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setup()
			scopes := GetScopes(ctx)
			require.Equal(t, tc.expected, scopes)
		})
	}
}

func TestNewScopeValidator(t *testing.T) {
	t.Parallel()

	config := ScopeConfig{
		CoarseScopes:     []string{"custom:admin"},
		FineScopes:       []string{"custom:read"},
		ScopeHierarchy:   map[string][]string{"custom:admin": {"custom:read"}},
		ErrorDetailLevel: "verbose",
	}

	validator := NewScopeValidator(config)

	require.NotNil(t, validator)
	require.Equal(t, config.CoarseScopes, validator.config.CoarseScopes)
	require.Equal(t, config.FineScopes, validator.config.FineScopes)
	require.Equal(t, "verbose", validator.config.ErrorDetailLevel)
}
