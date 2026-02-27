// Copyright (c) 2025 Justin Cranford

package certificate

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseProfile_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yaml     string
		wantName string
		wantType ProfileType
	}{
		{
			name: "root-ca",
			yaml: `
name: root-ca
description: Root CA certificate profile
type: root
validity:
  duration: 87600h
  max_duration: 175200h
key_usage:
  cert_sign: true
  crl_sign: true
basic_constraints:
  is_ca: true
  path_len_constraint: 2
`,
			wantName: "root-ca",
			wantType: ProfileTypeRoot,
		},
		{
			name: "tls-server",
			yaml: `
name: tls-server
description: TLS Server certificate profile
type: tls-server
validity:
  duration: 8760h
  max_duration: 8760h
  allow_custom: true
  backdate_buffer: 5m
key_usage:
  digital_signature: true
  key_encipherment: true
extended_key_usage:
  server_auth: true
`,
			wantName: "tls-server",
			wantType: ProfileTypeTLSServer,
		},
		{
			name: "code-signing",
			yaml: `
name: code-signing
description: Code signing certificate profile
type: code-signing
validity:
  duration: 8760h
key_usage:
  digital_signature: true
extended_key_usage:
  code_signing: true
`,
			wantName: "code-signing",
			wantType: ProfileTypeCodeSigning,
		},
		{
			name: "ocsp-responder",
			yaml: `
name: ocsp-responder
description: OCSP responder certificate
type: ocsp
validity:
  duration: 720h
key_usage:
  digital_signature: true
extended_key_usage:
  ocsp_signing: true
`,
			wantName: "ocsp-responder",
			wantType: ProfileTypeOCSP,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile, err := ParseProfile([]byte(tc.yaml))
			require.NoError(t, err)
			require.NotNil(t, profile)
			require.Equal(t, tc.wantName, profile.Name)
			require.Equal(t, tc.wantType, profile.Type)
		})
	}
}

func TestParseProfile_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "empty-name",
			yaml: `
type: tls-server
`,
			wantErr: "profile name is required",
		},
		{
			name: "empty-type",
			yaml: `
name: test
`,
			wantErr: "profile type is required",
		},
		{
			name: "invalid-type",
			yaml: `
name: test
type: invalid-type
`,
			wantErr: "invalid profile type",
		},
		{
			name: "invalid-duration",
			yaml: `
name: test
type: tls-server
validity:
  duration: not-a-duration
`,
			wantErr: "invalid validity duration",
		},
		{
			name: "invalid-max-duration",
			yaml: `
name: test
type: tls-server
validity:
  max_duration: bad
`,
			wantErr: "invalid max validity duration",
		},
		{
			name: "invalid-backdate",
			yaml: `
name: test
type: tls-server
validity:
  backdate_buffer: xyz
`,
			wantErr: "invalid backdate buffer",
		},
		{
			name: "ca-without-certsign",
			yaml: `
name: test
type: root
basic_constraints:
  is_ca: true
key_usage:
  crl_sign: true
`,
			wantErr: "CA profile must have cert_sign key usage",
		},
		{
			name:    "invalid-yaml",
			yaml:    `{{{invalid`,
			wantErr: "failed to parse certificate profile YAML",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile, err := ParseProfile([]byte(tc.yaml))
			require.Error(t, err)
			require.Nil(t, profile)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestKeyUsageConfig_ToX509KeyUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config KeyUsageConfig
		want   x509.KeyUsage
	}{
		{
			name:   "empty",
			config: KeyUsageConfig{},
			want:   0,
		},
		{
			name: "digital-signature",
			config: KeyUsageConfig{
				DigitalSignature: true,
			},
			want: x509.KeyUsageDigitalSignature,
		},
		{
			name: "ca-usage",
			config: KeyUsageConfig{
				CertSign: true,
				CRLSign:  true,
			},
			want: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		},
		{
			name: "tls-server-usage",
			config: KeyUsageConfig{
				DigitalSignature: true,
				KeyEncipherment:  true,
			},
			want: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		},
		{
			name: "all-usage",
			config: KeyUsageConfig{
				DigitalSignature:  true,
				ContentCommitment: true,
				KeyEncipherment:   true,
				DataEncipherment:  true,
				KeyAgreement:      true,
				CertSign:          true,
				CRLSign:           true,
				EncipherOnly:      true,
				DecipherOnly:      true,
			},
			want: x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment |
				x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment |
				x509.KeyUsageKeyAgreement | x509.KeyUsageCertSign |
				x509.KeyUsageCRLSign | x509.KeyUsageEncipherOnly |
				x509.KeyUsageDecipherOnly,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.config.ToX509KeyUsage()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestExtKeyUsageConfig_ToX509ExtKeyUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config ExtKeyUsageConfig
		want   []x509.ExtKeyUsage
	}{
		{
			name:   "empty",
			config: ExtKeyUsageConfig{},
			want:   nil,
		},
		{
			name: "server-auth",
			config: ExtKeyUsageConfig{
				ServerAuth: true,
			},
			want: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		{
			name: "tls-both",
			config: ExtKeyUsageConfig{
				ServerAuth: true,
				ClientAuth: true,
			},
			want: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		},
		{
			name: "all-standard",
			config: ExtKeyUsageConfig{
				ServerAuth:      true,
				ClientAuth:      true,
				CodeSigning:     true,
				EmailProtection: true,
				TimeStamping:    true,
				OCSPSigning:     true,
			},
			want: []x509.ExtKeyUsage{
				x509.ExtKeyUsageServerAuth,
				x509.ExtKeyUsageClientAuth,
				x509.ExtKeyUsageCodeSigning,
				x509.ExtKeyUsageEmailProtection,
				x509.ExtKeyUsageTimeStamping,
				x509.ExtKeyUsageOCSPSigning,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.config.ToX509ExtKeyUsage()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestValidityConfig_GetDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  ValidityConfig
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "empty",
			config:  ValidityConfig{},
			wantErr: true,
		},
		{
			name: "one-year",
			config: ValidityConfig{
				Duration: "8760h",
			},
			want: 8760 * time.Hour,
		},
		{
			name: "thirty-days",
			config: ValidityConfig{
				Duration: "720h",
			},
			want: 720 * time.Hour,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.config.GetDuration()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestValidityConfig_ValidateDuration(t *testing.T) {
	t.Parallel()

	oneYear := 8760 * time.Hour
	twoYears := 2 * oneYear

	tests := []struct {
		name      string
		config    ValidityConfig
		requested time.Duration
		wantErr   bool
	}{
		{
			name: "custom-allowed-within-max",
			config: ValidityConfig{
				Duration:    "8760h",
				MaxDuration: "17520h",
				AllowCustom: true,
			},
			requested: oneYear,
			wantErr:   false,
		},
		{
			name: "custom-allowed-exceeds-max",
			config: ValidityConfig{
				Duration:    "8760h",
				MaxDuration: "8760h",
				AllowCustom: true,
			},
			requested: twoYears,
			wantErr:   true,
		},
		{
			name: "custom-not-allowed-matches-default",
			config: ValidityConfig{
				Duration:    "8760h",
				AllowCustom: false,
			},
			requested: oneYear,
			wantErr:   false,
		},
		{
			name: "custom-not-allowed-differs-from-default",
			config: ValidityConfig{
				Duration:    "8760h",
				AllowCustom: false,
			},
			requested: twoYears,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.ValidateDuration(tc.requested)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidityConfig_GetBackdateBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  ValidityConfig
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "empty-returns-zero",
			config:  ValidityConfig{},
			want:    0,
			wantErr: false,
		},
		{
			name: "five-minutes",
			config: ValidityConfig{
				BackdateBuffer: "5m",
			},
			want:    cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Minute,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.config.GetBackdateBuffer()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
