package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidatePostgresSecretsDirSync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		sharedCompose       string
		envContentByService map[string]string
		wantValid           bool
		wantErrText         string
	}{
		{
			name:          "all service env files in sync",
			sharedCompose: "secrets:\n  postgres-username.secret:\n    file: ${POSTGRES_SECRETS_DIR:-./secrets}/postgres-username.secret\n",
			envContentByService: map[string]string{
				cryptoutilSharedMagic.OTLPServiceSMKMS:  "POSTGRES_SECRETS_DIR=../sm-kms/secrets\n",
				cryptoutilSharedMagic.OTLPServiceJoseJA: "POSTGRES_SECRETS_DIR=../jose-ja/secrets\n",
			},
			wantValid: true,
		},
		{
			name:          "service env value mismatch",
			sharedCompose: "secrets:\n  postgres-username.secret:\n    file: ${POSTGRES_SECRETS_DIR:-./secrets}/postgres-username.secret\n",
			envContentByService: map[string]string{
				cryptoutilSharedMagic.OTLPServiceSMIM: "POSTGRES_SECRETS_DIR=../wrong/secrets\n",
			},
			wantValid:   false,
			wantErrText: "mismatch",
		},
		{
			name:          "missing env key",
			sharedCompose: "secrets:\n  postgres-password.secret:\n    file: ${POSTGRES_SECRETS_DIR:-./secrets}/postgres-password.secret\n",
			envContentByService: map[string]string{
				cryptoutilSharedMagic.OTLPServicePKICA: "OTHER_KEY=value\n",
			},
			wantValid:   false,
			wantErrText: "Missing POSTGRES_SECRETS_DIR",
		},
		{
			name:          "shared compose missing reference pattern",
			sharedCompose: "secrets:\n  postgres-password.secret:\n    file: ./secrets/postgres-password.secret\n",
			envContentByService: map[string]string{
				cryptoutilSharedMagic.OTLPServiceIdentityAuthz: "POSTGRES_SECRETS_DIR=../identity-authz/secrets\n",
			},
			wantValid:   false,
			wantErrText: "missing '${POSTGRES_SECRETS_DIR:-./secrets}/'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deploymentsDir := t.TempDir()
			sharedDir := filepath.Join(deploymentsDir, postgresSharedDirName)
			require.NoError(t, os.MkdirAll(sharedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, os.WriteFile(filepath.Join(sharedDir, composeFileName), []byte(tc.sharedCompose), cryptoutilSharedMagic.CacheFilePermissions))

			for service, envContent := range tc.envContentByService {
				serviceDir := filepath.Join(deploymentsDir, service)
				require.NoError(t, os.MkdirAll(serviceDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(serviceDir, postgresEnvFileName), []byte(envContent), cryptoutilSharedMagic.CacheFilePermissions))
			}

			result, err := ValidatePostgresSecretsDirSync(deploymentsDir)
			require.NoError(t, err)
			require.Equal(t, tc.wantValid, result.Valid)

			if tc.wantErrText != "" {
				require.Contains(t, FormatPostgresSecretsDirSyncResult(result), tc.wantErrText)
			}
		})
	}
}
