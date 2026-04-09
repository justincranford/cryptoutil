package businesslogic

import (
	"context"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm-kms/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm-kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

// seedImportableElasticKeyWithStatus creates an ElasticKey with ImportAllowed=true and the given status.
func seedImportableElasticKeyWithStatus(t *testing.T, stack *testStack, name string, status cryptoutilKmsServer.ElasticKeyStatus) googleUuid.UUID {
	t.Helper()

	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	ekID := googleUuid.New()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       "test-import-status",
		ElasticKeyProvider:          "Internal",
		ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMDir,
		ElasticKeyVersioningAllowed: false,
		ElasticKeyImportAllowed:     true,
		ElasticKeyStatus:            status,
	}

	testify.NoError(t, stack.core.DB.Create(ek).Error)

	return ekID
}

func TestNoTenantContext(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	noTenant := context.Background()
	id := googleUuid.New()

	tests := []struct {
		name string
		fn   func() error
	}{
		{name: "GetElasticKeyByID", fn: func() error {
			_, err := stack.service.GetElasticKeyByElasticKeyID(noTenant, &id)

			return err
		}},
		{name: "GetElasticKeys", fn: func() error {
			_, err := stack.service.GetElasticKeys(noTenant, nil)

			return err
		}},
		{name: "GetMaterialKeysForElasticKey", fn: func() error {
			_, err := stack.service.GetMaterialKeysForElasticKey(noTenant, &id, nil)

			return err
		}},
		{name: "GetMaterialKeys", fn: func() error {
			_, err := stack.service.GetMaterialKeys(noTenant, nil)

			return err
		}},
		{name: "GetMaterialKeyByIDs", fn: func() error {
			_, err := stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(noTenant, &id, &id)

			return err
		}},
		{name: "UpdateElasticKey", fn: func() error {
			_, err := stack.service.UpdateElasticKey(noTenant, &id, &cryptoutilKmsServer.ElasticKeyUpdate{Name: "x"})

			return err
		}},
		{name: "DeleteElasticKey", fn: func() error {
			return stack.service.DeleteElasticKey(noTenant, &id)
		}},
		{name: "AddElasticKey", fn: func() error {
			_, err := stack.service.AddElasticKey(noTenant, &cryptoutilKmsServer.ElasticKeyCreate{
				Name: "x", Algorithm: string(cryptoutilOpenapiModel.A256GCMDir), Provider: providerInternal,
			})

			return err
		}},
		{name: "GenerateMaterialKey", fn: func() error {
			_, err := stack.service.GenerateMaterialKeyInElasticKey(noTenant, &id, nil)

			return err
		}},
		{name: "PostEncrypt", fn: func() error {
			_, err := stack.service.PostEncryptByElasticKeyID(noTenant, &id, nil, []byte("x"))

			return err
		}},
		{name: "PostSign", fn: func() error {
			_, err := stack.service.PostSignByElasticKeyID(noTenant, &id, []byte("x"))

			return err
		}},
		{name: "ImportMaterialKey", fn: func() error {
			_, err := stack.service.ImportMaterialKey(noTenant, &id, &cryptoutilKmsServer.MaterialKeyImport{JWK: "x"})

			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			testify.Error(t, err)
			testify.Contains(t, err.Error(), "tenant context required")
		})
	}
}

func TestAddElasticKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		create   cryptoutilKmsServer.ElasticKeyCreate
		wantErr  string
		wantName string
	}{
		{
			name: "success",
			create: cryptoutilKmsServer.ElasticKeyCreate{
				Name: "add-ek-success", Algorithm: string(cryptoutilOpenapiModel.A256GCMDir),
				Provider: providerInternal, Description: ptr("add-ek-desc"),
			},
			wantName: "add-ek-success",
		},
		{
			name: "import allowed rejected",
			create: cryptoutilKmsServer.ElasticKeyCreate{
				Name: "import-test", Algorithm: string(cryptoutilOpenapiModel.A256GCMDir),
				Provider: providerInternal, ImportAllowed: ptr(true),
			},
			wantErr: "elasticKeyImportAllowed=true not supported yet",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stack := setupTestStack(t)
			result, err := stack.service.AddElasticKey(stack.ctx, &tc.create)

			if tc.wantErr != "" {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.wantErr)

				return
			}

			testify.NoError(t, err)
			testify.NotNil(t, result)
			testify.Equal(t, tc.wantName, *result.Name)
			testify.NotNil(t, result.ElasticKeyID)
		})
	}
}

func TestGenerateMaterialKeyInElasticKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func(t *testing.T, stack *testStack) googleUuid.UUID
		wantOK bool
	}{
		{
			name: "success",
			setup: func(t *testing.T, stack *testStack) googleUuid.UUID {
				t.Helper()

				return seedElasticKey(t, stack, "gen-mk-success", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
			},
			wantOK: true,
		},
		{
			name: "not found",
			setup: func(t *testing.T, _ *testStack) googleUuid.UUID {
				t.Helper()

				return googleUuid.New()
			},
		},
		{
			name: "invalid status",
			setup: func(t *testing.T, stack *testStack) googleUuid.UUID {
				t.Helper()

				return seedElasticKey(t, stack, "gen-mk-invalid-status", cryptoutilOpenapiModel.A256GCMDir,
					cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stack := setupTestStack(t)
			ekID := tc.setup(t, stack)
			result, err := stack.service.GenerateMaterialKeyInElasticKey(stack.ctx, &ekID, nil)

			if !tc.wantOK {
				testify.Error(t, err)

				return
			}

			testify.NoError(t, err)
			testify.NotNil(t, result)
			testify.NotNil(t, result.ElasticKeyID)
			testify.Equal(t, ekID, *result.ElasticKeyID)
		})
	}
}

func TestImportMaterialKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, stack *testStack) googleUuid.UUID
		wantOK  bool
		wantErr string
	}{
		{
			name: "success",
			setup: func(t *testing.T, stack *testStack) googleUuid.UUID {
				t.Helper()

				return seedImportableElasticKeyWithStatus(t, stack, "import-mk-success", cryptoutilKmsServer.Active)
			},
			wantOK: true,
		},
		{
			name: "not found",
			setup: func(t *testing.T, _ *testStack) googleUuid.UUID {
				t.Helper()

				return googleUuid.New()
			},
		},
		{
			name: "invalid status",
			setup: func(t *testing.T, stack *testStack) googleUuid.UUID {
				t.Helper()

				return seedImportableElasticKeyWithStatus(t, stack, "import-mk-bad-status",
					cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled))
			},
			wantErr: "invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stack := setupTestStack(t)
			ekID := tc.setup(t, stack)
			result, err := stack.service.ImportMaterialKey(stack.ctx, &ekID, &cryptoutilKmsServer.MaterialKeyImport{
				JWK: `{"kty":"oct","k":"dGVzdGtleWRhdGExMjM0NTY3ODkwMTIzNDU2"}`,
			})

			if !tc.wantOK {
				testify.Error(t, err)

				if tc.wantErr != "" {
					testify.Contains(t, err.Error(), tc.wantErr)
				}

				return
			}

			testify.NoError(t, err)
			testify.NotNil(t, result)
			testify.NotNil(t, result.ElasticKeyID)
			testify.Equal(t, ekID, *result.ElasticKeyID)
		})
	}
}

func TestPostOperation_NoMaterialKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(*testStack, googleUuid.UUID) error
	}{
		{name: "encrypt", fn: func(s *testStack, id googleUuid.UUID) error {
			_, err := s.service.PostEncryptByElasticKeyID(s.ctx, &id, &cryptoutilOpenapiModel.EncryptParams{}, []byte("test"))

			return err
		}},
		{name: "sign", fn: func(s *testStack, id googleUuid.UUID) error {
			_, err := s.service.PostSignByElasticKeyID(s.ctx, &id, []byte("test"))

			return err
		}},
		{name: "generate", fn: func(s *testStack, id googleUuid.UUID) error {
			validAlg := cryptoutilOpenapiModel.Oct256
			_, _, _, err := s.service.PostGenerateByElasticKeyID(s.ctx, &id, &cryptoutilOpenapiModel.GenerateParams{Alg: &validAlg})

			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stack := setupTestStack(t)
			ekID := seedElasticKey(t, stack, "op-nomk-"+tc.name, cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
			err := tc.fn(stack, ekID)
			testify.Error(t, err)
		})
	}
}

func TestPostOperation_InvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fn      func(*testStack, googleUuid.UUID) error
		wantErr string
	}{
		{name: "decrypt invalid JWE", fn: func(s *testStack, id googleUuid.UUID) error {
			_, err := s.service.PostDecryptByElasticKeyID(s.ctx, &id, []byte("not-a-jwe"))

			return err
		}, wantErr: "failed to parse JWE message bytes"},
		{name: "verify invalid JWS", fn: func(s *testStack, id googleUuid.UUID) error {
			_, err := s.service.PostVerifyByElasticKeyID(s.ctx, &id, []byte("not-a-jws"))

			return err
		}, wantErr: "failed to parse JWS message bytes"},
		{name: "generate invalid algorithm", fn: func(s *testStack, id googleUuid.UUID) error {
			badAlg := cryptoutilOpenapiModel.GenerateAlgorithm("BAD")
			_, _, _, err := s.service.PostGenerateByElasticKeyID(s.ctx, &id, &cryptoutilOpenapiModel.GenerateParams{Alg: &badAlg})

			return err
		}, wantErr: "failed to map generate algorithm"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stack := setupTestStack(t)
			ekID := googleUuid.New()
			err := tc.fn(stack, ekID)
			testify.Error(t, err)
			testify.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
