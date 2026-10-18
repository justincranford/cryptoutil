package tls

import (
	"encoding/base64"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestConfigMaterialValidation(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	nonCAFixture := newPublishValidationFixtureWithCA(t, false)

	tests := []struct {
		name        string
		certificate []byte
		key         []byte
		wantErr     string
	}{
		{
			name:        "valid CA config",
			certificate: fixture.certificatePEM,
			key:         fixture.privateKeyPEM,
		},
		{
			name:    testTLSCertificateValueMissing,
			key:     fixture.privateKeyPEM,
			wantErr: testTLSMissingCertPEMSetting,
		},
		{
			name:        testTLSKeyValueMalformed,
			certificate: fixture.certificatePEM,
			key:         []byte(cryptoutilSharedMagic.DevSetupTestInvalidName),
			wantErr:     "decode tls-mixed-ca-key-pem",
		},
		{
			name:        "certificate is not a CA",
			certificate: nonCAFixture.certificatePEM,
			key:         nonCAFixture.privateKeyPEM,
			wantErr:     "does not contain a CA",
		},
		{
			name:        "CA key mismatch",
			certificate: fixture.certificatePEM,
			key:         nonCAFixture.privateKeyPEM,
			wantErr:     testTLSPrivateKeyDoesNotMatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			configPath := filepath.Join(t.TempDir(), "tls-config.yml")

			config := makeTLSConfig(tc.certificate, tc.key)
			if tc.name == testTLSKeyValueMalformed {
				config = []byte(testTLSConfigCertPrefix + base64.StdEncoding.EncodeToString(tc.certificate) + "\n" +
					testTLSConfigKeyPrefix + "invalid\n")
			}

			if tc.name == testTLSCertificateValueMissing {
				config = []byte(testTLSConfigKeyPrefix + base64.StdEncoding.EncodeToString(tc.key) + "\n")
			}

			writePublishFile(t, configPath, config, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)

			err := validateTLSConfig(configPath)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}
