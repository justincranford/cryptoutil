// Copyright (c) 2025 Justin Cranford

package domain

import (
	"testing"

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
		{name: "ElasticJWK", model: &ElasticJWK{}, wantTable: "elastic_jwks"},
		{name: "MaterialJWK", model: &MaterialJWK{}, wantTable: "material_jwks"},
		{name: "AuditConfig", model: &AuditConfig{}, wantTable: "tenant_audit_config"},
		{name: "AuditLogEntry", model: &AuditLogEntry{}, wantTable: "audit_log"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantTable, tc.model.TableName())
		})
	}
}
