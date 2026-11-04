// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateRegistryScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(model, []Scope) (model, []Scope)
		wantErr string
	}{
		{
			name: testEmptyRegistryCase,
			mutate: func(m model, _ []Scope) (model, []Scope) {
				return m, nil
			},
			wantErr: "registry and modes must be nonempty",
		},
		{
			name: testInvalidScopeIDCase,
			mutate: func(m model, scopes []Scope) (model, []Scope) {
				scopes[0].ID = "Invalid_ID"

				return m, scopes
			},
			wantErr: "invalid scope ID",
		},
		{
			name: testDuplicateScopeCase,
			mutate: func(m model, scopes []Scope) (model, []Scope) {
				return m, append(scopes, scopes[0])
			},
			wantErr: "duplicate scope",
		},
		{
			name: "service scope without its PS-ID member",
			mutate: func(m model, scopes []Scope) (model, []Scope) {
				for i := range scopes {
					if scopes[i].Tier == TierService {
						scopes[i].Members = nil

						break
					}
				}

				return m, scopes
			},
			wantErr: "invalid or duplicate PS-ID scope",
		},
		{
			name: "empty product scope",
			mutate: func(m model, scopes []Scope) (model, []Scope) {
				for i := range scopes {
					if scopes[i].Tier == TierProduct {
						scopes[i].Members = nil

						break
					}
				}

				return m, scopes
			},
			wantErr: "empty registry scope",
		},
		{
			name: testUnknownRegistryTierCase,
			mutate: func(m model, scopes []Scope) (model, []Scope) {
				scopes[0].Tier = Tier(cryptoutilSharedMagic.STRATEGY_UNKNOWN)

				return m, scopes
			},
			wantErr: "unknown registry tier",
		},
		{
			name: "missing registry tier",
			mutate: func(m model, scopes []Scope) (model, []Scope) {
				filtered := make([]Scope, 0, len(scopes))
				for _, scope := range scopes {
					if scope.Tier != TierService {
						filtered = append(filtered, scope)
					}
				}

				return m, filtered
			},
			wantErr: "registry tiers must be nonempty",
		},
		{
			name: "unknown PS-ID member",
			mutate: func(m model, scopes []Scope) (model, []Scope) {
				for i := range scopes {
					if scopes[i].Tier == TierProduct {
						scopes[i].Members = []string{"unknown-service"}

						break
					}
				}

				return m, scopes
			},
			wantErr: "unknown PS-ID",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m, _ := fixture(t)

			scopes := append([]Scope(nil), RegisteredScopes()...)
			m, scopes = tc.mutate(m, scopes)

			require.ErrorContains(t, validate(m, scopes), tc.wantErr)
		})
	}
}
