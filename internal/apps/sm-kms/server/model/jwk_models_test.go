// Copyright (c) 2025-2026 Justin Cranford.
package model

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type tableNamer interface {
	TableName() string
}

func TestTableNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		model     tableNamer
		wantTable string
	}{
		{name: "ElasticJWK", model: &ElasticJWK{}, wantTable: cryptoutilSharedMagic.ELASTIC_JWKS},
		{name: "MaterialJWK", model: &MaterialJWK{}, wantTable: cryptoutilSharedMagic.MATERIAL_JWKS},
		{name: "AuditConfig", model: &AuditConfig{}, wantTable: cryptoutilSharedMagic.TENANT_AUDIT_CONFIG},
		{name: "AuditLogEntry", model: &AuditLogEntry{}, wantTable: cryptoutilSharedMagic.AUDIT_LOG},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantTable, tc.model.TableName())
		})
	}
}
