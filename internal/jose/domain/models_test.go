// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestElasticJWK_TableName(t *testing.T) {
	t.Parallel()

	model := &ElasticJWK{}
	require.Equal(t, "elastic_jwks", model.TableName())
}

func TestElasticJWK_Fields(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	model := &ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  "test-kid",
		KTY:                  "RSA",
		ALG:                  "RS256",
		USE:                  "sig",
		MaxMaterials:         1000,
		CurrentMaterialCount: 0,
		CreatedAt:            1234567890,
	}

	require.NotEqual(t, googleUuid.Nil, model.ID)
	require.Equal(t, tenantID, model.TenantID)
	require.Equal(t, realmID, model.RealmID)
	require.Equal(t, "test-kid", model.KID)
	require.Equal(t, "RSA", model.KTY)
	require.Equal(t, "RS256", model.ALG)
	require.Equal(t, "sig", model.USE)
	require.Equal(t, 1000, model.MaxMaterials)
	require.Equal(t, 0, model.CurrentMaterialCount)
	require.Equal(t, int64(1234567890), model.CreatedAt)
}

func TestMaterialJWK_TableName(t *testing.T) {
	t.Parallel()

	model := &MaterialJWK{}
	require.Equal(t, "material_jwks", model.TableName())
}

func TestMaterialJWK_Fields(t *testing.T) {
	t.Parallel()

	elasticJWKID := googleUuid.New()
	retiredAt := int64(9876543210)

	model := &MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWKID,
		MaterialKID:    "material-kid-001",
		PrivateJWKJWE:  "eyJhbGc...",
		PublicJWKJWE:   "eyJhbGc...",
		Active:         true,
		CreatedAt:      1234567890,
		RetiredAt:      &retiredAt,
		BarrierVersion: 1,
	}

	require.NotEqual(t, googleUuid.Nil, model.ID)
	require.Equal(t, elasticJWKID, model.ElasticJWKID)
	require.Equal(t, "material-kid-001", model.MaterialKID)
	require.Equal(t, "eyJhbGc...", model.PrivateJWKJWE)
	require.Equal(t, "eyJhbGc...", model.PublicJWKJWE)
	require.True(t, model.Active)
	require.Equal(t, int64(1234567890), model.CreatedAt)
	require.NotNil(t, model.RetiredAt)
	require.Equal(t, int64(9876543210), *model.RetiredAt)
	require.Equal(t, 1, model.BarrierVersion)
}

func TestAuditConfig_TableName(t *testing.T) {
	t.Parallel()

	model := &AuditConfig{}
	require.Equal(t, "tenant_audit_config", model.TableName())
}

func TestAuditConfig_Fields(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	model := &AuditConfig{
		TenantID:     tenantID,
		Operation:    "encrypt",
		Enabled:      true,
		SamplingRate: 0.01,
	}

	require.Equal(t, tenantID, model.TenantID)
	require.Equal(t, "encrypt", model.Operation)
	require.True(t, model.Enabled)
	require.Equal(t, 0.01, model.SamplingRate)
}

func TestAuditLogEntry_TableName(t *testing.T) {
	t.Parallel()

	model := &AuditLogEntry{}
	require.Equal(t, "tenant_audit_log", model.TableName())
}

func TestAuditLogEntry_Fields(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	userID := googleUuid.New()
	errorMsg := "test error"
	metadata := `{"key":"value"}`

	model := &AuditLogEntry{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		RealmID:      realmID,
		UserID:       &userID,
		Operation:    "encrypt",
		ResourceType: "elastic_jwk",
		ResourceID:   "resource-123",
		Success:      false,
		ErrorMessage: &errorMsg,
		Metadata:     &metadata,
		CreatedAt:    1234567890,
	}

	require.NotEqual(t, googleUuid.Nil, model.ID)
	require.Equal(t, tenantID, model.TenantID)
	require.Equal(t, realmID, model.RealmID)
	require.NotNil(t, model.UserID)
	require.Equal(t, userID, *model.UserID)
	require.Equal(t, "encrypt", model.Operation)
	require.Equal(t, "elastic_jwk", model.ResourceType)
	require.Equal(t, "resource-123", model.ResourceID)
	require.False(t, model.Success)
	require.NotNil(t, model.ErrorMessage)
	require.Equal(t, "test error", *model.ErrorMessage)
	require.NotNil(t, model.Metadata)
	require.Equal(t, `{"key":"value"}`, *model.Metadata)
	require.Equal(t, int64(1234567890), model.CreatedAt)
}
