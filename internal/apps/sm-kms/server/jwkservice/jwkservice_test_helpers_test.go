// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"encoding/base64"
	"testing"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createTestElasticKey(
	t *testing.T,
	tenantID googleUuid.UUID,
	algorithm, use string,
	maxMaterials int,
) (*cryptoutilAppsSmKmsModel.ElasticJWK, *cryptoutilAppsSmKmsModel.MaterialJWK, error) {
	t.Helper()

	return newTestElasticJWKService().CreateElasticJWK(
		testContext,
		tenantID,
		algorithm,
		use,
		maxMaterials,
	)
}

func encryptTestValue(t *testing.T, plaintext []byte) string {
	t.Helper()

	encrypted, err := testBarrierService.EncryptContentWithContext(testContext, plaintext)
	require.NoError(t, err)

	return base64.StdEncoding.EncodeToString(encrypted)
}
