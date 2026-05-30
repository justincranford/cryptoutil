package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateCertVolumePolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		deploymentName string
		composeContent string
		wantValid      bool
		wantErrText    string
	}{
		{
			name:           "valid ps-id named volume",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/certs\n  sm-kms-app:\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      true,
		},
		{
			name:           "forbidden cert bind mount",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - ./certs:/certs:rw\n  sm-kms-app:\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    "forbidden cert bind mount",
		},
		{
			name:           "missing top-level volume declaration",
			deploymentName: cryptoutilSharedMagic.OTLPServicePKICA,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - pki-ca-certs:/certs\n",
			wantValid:      false,
			wantErrText:    "missing top-level named volume declaration",
		},
		{
			name:           "valid template placeholder volume",
			deploymentName: cryptoutilSharedMagic.SkeletonTemplateServiceName,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - __PS_ID__-certs:/certs\n  app:\n    volumes:\n      - __PS_ID__-certs:/certs:ro\nvolumes:\n  __PS_ID__-certs:\n",
			wantValid:      true,
		},
		{
			name:           "valid infrastructure mount path",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/mnt/ps-certs-src\n  grafana-otel-lgtm:\n    volumes:\n      - sm-kms-certs:/mnt/ps-certs-src:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			composePath := filepath.Join(dir, composeFileName)
			require.NoError(t, os.WriteFile(composePath, []byte(tc.composeContent), cryptoutilSharedMagic.CacheFilePermissions))

			result, err := ValidateCertVolumePolicy(dir, tc.deploymentName)
			require.NoError(t, err)
			require.Equal(t, tc.wantValid, result.Valid)

			if tc.wantErrText != "" {
				require.Contains(t, FormatCertVolumePolicyResult(result), tc.wantErrText)
			}
		})
	}
}
