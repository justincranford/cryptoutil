// Copyright (c) 2025 Justin Cranford

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestElasticJWK_TableName(t *testing.T) {
	elasticJWK := ElasticJWK{}
	require.Equal(t, "elastic_jwks", elasticJWK.TableName())
}

func TestMaterialJWK_TableName(t *testing.T) {
	materialJWK := MaterialJWK{}
	require.Equal(t, "material_jwks", materialJWK.TableName())
}

func TestAuditConfig_TableName(t *testing.T) {
	auditConfig := AuditConfig{}
	require.Equal(t, "tenant_audit_config", auditConfig.TableName())
}

func TestAuditLogEntry_TableName(t *testing.T) {
	auditLogEntry := AuditLogEntry{}
	require.Equal(t, "audit_log", auditLogEntry.TableName())
}
