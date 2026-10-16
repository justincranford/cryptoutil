package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testCertVolumeMissingActivationBarrier = "without depending on"
	testUnresolvedCertVolumeTarget         = "unresolved variable interpolation in volume target"
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
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/certs\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      true,
		},
		{
			name:           "cert consumer does not depend on initialization",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/certs\n  sm-kms-app:\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testCertVolumeMissingActivationBarrier,
		},
		{
			name:           "cert consumer only waits for initialization to start",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/certs\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_started\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testCertVolumeMissingActivationBarrier,
		},
		{
			name:           "cert consumer uses shorthand dependency",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/certs\n  sm-kms-app:\n    depends_on:\n      - pki-init\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testCertVolumeMissingActivationBarrier,
		},
		{
			name:           testForbiddenCertBindMount,
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - ./certs:/certs:rw\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testForbiddenCertBindMount,
		},
		{
			name:           "trailing slash bind target at certificate root rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - ./certs:/certs/:rw\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testForbiddenCertBindMount,
		},
		{
			name:           "short syntax interpolated certificate target rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:${CERT_TARGET}\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testUnresolvedCertVolumeTarget,
		},
		{
			name:           "short syntax unbraced interpolated certificate target rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:$CERT_TARGET\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testUnresolvedCertVolumeTarget,
		},
		{
			name:           "short syntax interpolated default certificate target rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:${CERT_TARGET:-/certs}\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testUnresolvedCertVolumeTarget,
		},
		{
			name:           "long syntax interpolated certificate target rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: `services:
  pki-init:
    volumes:
      - type: volume
        source: sm-kms-certs
        target: ${CERT_TARGET}
  sm-kms-app:
    depends_on:
      pki-init:
        condition: service_completed_successfully
    volumes:
      - type: volume
        source: sm-kms-certs
        target: /certs
volumes:
  sm-kms-certs:
`,
			wantValid:   false,
			wantErrText: testUnresolvedCertVolumeTarget,
		},
		{
			name:           "source interpolation and escaped target dollar allowed",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/certs\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\n      - ${DATA_VOLUME}:/data:ro\n      - data-volume:/data/$$DATA_SEGMENT:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      true,
		},
		{
			name:           "bind target nested under certificate root rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - ./private:/certs/private:rw\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testForbiddenCertBindMount,
		},
		{
			name:           "bind target overlapping infrastructure certificate root rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - /host/certs:/mnt:rw\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/mnt/ps-certs-src:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testForbiddenCertBindMount,
		},
		{
			name:           "root bind target overlapping certificate roots rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - /host:/\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testForbiddenCertBindMount,
		},
		{
			name:           "bind target beside certificate root remains allowed",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - ./backup:/certs-backup:rw\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      true,
		},
		{
			name:           "named volume target with trailing slash accepted",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/certs/\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/certs/:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      true,
		},
		{
			name:           "bind source containing expected volume name rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - ./sm-kms-certs:/certs:rw\n  sm-kms-app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - ./sm-kms-certs:/certs:ro\nvolumes:\n  sm-kms-certs:\n",
			wantValid:      false,
			wantErrText:    testForbiddenCertBindMount,
		},
		{
			name:           "valid long syntax named volume",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: `services:
  pki-init:
    volumes:
      - type: volume
        source: sm-kms-certs
        target: /certs
  sm-kms-app:
    depends_on:
      pki-init:
        condition: service_completed_successfully
    volumes:
      - type: volume
        source: sm-kms-certs
        target: /certs
        read_only: true
volumes:
  sm-kms-certs:
`,
			wantValid: true,
		},
		{
			name:           "long syntax bind mount to cert target rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: `services:
  pki-init:
    volumes:
      - type: bind
        source: ./sm-kms-certs
        target: /certs
  sm-kms-app:
    depends_on:
      pki-init:
        condition: service_completed_successfully
    volumes:
      - type: volume
        source: sm-kms-certs
        target: /certs
volumes:
  sm-kms-certs:
`,
			wantValid:   false,
			wantErrText: testForbiddenCertBindMount,
		},
		{
			name:           "long syntax trailing slash bind target rejected",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: `services:
  pki-init:
    volumes:
      - type: bind
        source: ./certs
        target: /certs/
  sm-kms-app:
    depends_on:
      pki-init:
        condition: service_completed_successfully
    volumes:
      - type: volume
        source: sm-kms-certs
        target: /certs
volumes:
  sm-kms-certs:
`,
			wantValid:   false,
			wantErrText: testForbiddenCertBindMount,
		},
		{
			name:           "long syntax named volume with trailing slash accepted",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: `services:
  pki-init:
    volumes:
      - type: volume
        source: sm-kms-certs
        target: /certs/
  sm-kms-app:
    depends_on:
      pki-init:
        condition: service_completed_successfully
    volumes:
      - type: volume
        source: sm-kms-certs
        target: /certs/
volumes:
  sm-kms-certs:
`,
			wantValid: true,
		},
		{
			name:           "missing top-level volume declaration",
			deploymentName: cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - identity-authz-certs:/certs\n",
			wantValid:      false,
			wantErrText:    "missing top-level named volume declaration",
		},
		{
			name:           "valid template placeholder volume",
			deploymentName: cryptoutilSharedMagic.SkeletonTemplateServiceName,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - __PS_ID__-certs:/certs\n  app:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - __PS_ID__-certs:/certs:ro\nvolumes:\n  __PS_ID__-certs:\n",
			wantValid:      true,
		},
		{
			name:           "valid infrastructure mount path",
			deploymentName: cryptoutilSharedMagic.OTLPServiceSMKMS,
			composeContent: "services:\n  pki-init:\n    volumes:\n      - sm-kms-certs:/mnt/ps-certs-src\n  grafana-otel-lgtm:\n    depends_on:\n      pki-init:\n        condition: service_completed_successfully\n    volumes:\n      - sm-kms-certs:/mnt/ps-certs-src:ro\nvolumes:\n  sm-kms-certs:\n",
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

func TestValidateCertVolumePolicyDeploymentTiers(t *testing.T) {
	t.Parallel()

	deployments := []string{
		cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		cryptoutilSharedMagic.OTLPServiceIdentityRP,
		cryptoutilSharedMagic.OTLPServiceIdentityRS,
		cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
	}

	for _, deploymentName := range deployments {
		t.Run(deploymentName, func(t *testing.T) {
			t.Parallel()

			deploymentPath := filepath.Join("..", "..", "..", "..", "deployments", deploymentName)
			result, err := ValidateCertVolumePolicy(deploymentPath, deploymentName)
			require.NoError(t, err)
			require.True(t, result.Valid, "%s", result.Errors)
		})
	}
}

func TestFormatCertVolumePolicyResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result *CertVolumePolicyResult
		want   string
	}{
		{name: testNilResult, want: "[ValidateCertVolumePolicy] nil result"},
		{
			name:   testValidResult,
			result: &CertVolumePolicyResult{Path: composeFileName, Valid: true},
			want:   "[ValidateCertVolumePolicy] PASS: compose.yml",
		},
		{
			name:   testInvalidResultWithoutErrors,
			result: &CertVolumePolicyResult{Path: composeFileName},
			want:   "[ValidateCertVolumePolicy] FAIL: compose.yml",
		},
		{
			name:   testInvalidResultWithErrors,
			result: &CertVolumePolicyResult{Errors: []string{"missing cert volume", "invalid mount"}},
			want:   "missing cert volume\ninvalid mount",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, FormatCertVolumePolicyResult(tc.result))
		})
	}
}
