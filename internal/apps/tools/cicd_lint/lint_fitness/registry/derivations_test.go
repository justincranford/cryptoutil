// Copyright (c) 2025 Justin Cranford

package registry

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------
// Port derivation tests (Task 3.1)
// -----------------------------------------------------------------------

func TestPublicPort_AllPSIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		psID string
		want int
	}{
		{psID: cryptoutilSharedMagic.OTLPServiceSMKMS, want: int(cryptoutilSharedMagic.KMSServicePort)},
		{psID: cryptoutilSharedMagic.OTLPServiceSMIM, want: int(cryptoutilSharedMagic.IMServicePort)},
		{psID: cryptoutilSharedMagic.OTLPServiceJoseJA, want: int(cryptoutilSharedMagic.JoseJAServicePort)},
		{psID: cryptoutilSharedMagic.OTLPServicePKICA, want: int(cryptoutilSharedMagic.PKICAServicePort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: int(cryptoutilSharedMagic.IdentityDefaultAuthZPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityIDP, want: int(cryptoutilSharedMagic.IdentityDefaultIDPPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRS, want: int(cryptoutilSharedMagic.IdentityDefaultRSPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRP, want: int(cryptoutilSharedMagic.DefaultSPARPPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentitySPA, want: int(cryptoutilSharedMagic.IdentitySPAServicePort)},
		{psID: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: int(cryptoutilSharedMagic.SkeletonTemplateServicePort)},
	}

	for _, tt := range tests {
		t.Run(tt.psID, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, PublicPort(tt.psID))
		})
	}
}

func TestPublicPort_UnknownPSID(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, PublicPort("unknown-service"))
}

func TestAdminPort_AllServicesReturn9090(t *testing.T) {
	t.Parallel()

	tests := []string{
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceSMIM,
		cryptoutilSharedMagic.OTLPServiceJoseJA,
		cryptoutilSharedMagic.OTLPServicePKICA,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		cryptoutilSharedMagic.OTLPServiceIdentityRS,
		cryptoutilSharedMagic.OTLPServiceIdentityRP,
		cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
	}

	for _, psID := range tests {
		t.Run(psID, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, int(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil), AdminPort(psID))
		})
	}
}

func TestAdminPort_UnknownPSIDStillReturns9090(t *testing.T) {
	t.Parallel()

	require.Equal(t, int(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil), AdminPort("unknown"))
}

func TestPostgresPort_AllPSIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		psID string
		want int
	}{
		{psID: cryptoutilSharedMagic.OTLPServiceSMKMS, want: 54320},
		{psID: cryptoutilSharedMagic.OTLPServiceSMIM, want: 54321},
		{psID: cryptoutilSharedMagic.OTLPServiceJoseJA, want: 54322},
		{psID: cryptoutilSharedMagic.OTLPServicePKICA, want: 54323},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: 54324},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityIDP, want: 54325},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRS, want: 54326},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRP, want: 54327},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentitySPA, want: 54328},
		{psID: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: int(cryptoutilSharedMagic.SkeletonTemplatePostgresPort)},
	}

	for _, tt := range tests {
		t.Run(tt.psID, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, PostgresPort(tt.psID))
		})
	}
}

func TestPostgresPort_UnknownPSID(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, PostgresPort("unknown-service"))
}

func TestProductPublicPort_AllPSIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		psID string
		want int
	}{
		{psID: cryptoutilSharedMagic.OTLPServiceSMKMS, want: int(cryptoutilSharedMagic.ProductTierPortMin)},
		{psID: cryptoutilSharedMagic.OTLPServiceSMIM, want: 18100},
		{psID: cryptoutilSharedMagic.OTLPServiceJoseJA, want: 18200}, // PRODUCT level: jose-ja base 8200 + offset 10000
		{psID: cryptoutilSharedMagic.OTLPServicePKICA, want: 18300},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: int(cryptoutilSharedMagic.IdentityE2EAuthzPublicPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityIDP, want: int(cryptoutilSharedMagic.IdentityE2EIDPPublicPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRS, want: int(cryptoutilSharedMagic.IdentityE2ERSPublicPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRP, want: int(cryptoutilSharedMagic.IdentityE2ERPPublicPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentitySPA, want: int(cryptoutilSharedMagic.IdentityE2ESPAPublicPort)},
		{psID: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: 18900}, // PRODUCT level: skeleton-template base 8900 + offset 10000
	}

	for _, tt := range tests {
		t.Run(tt.psID, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, ProductPublicPort(tt.psID))
		})
	}
}

func TestProductPublicPort_UnknownPSID(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, ProductPublicPort("unknown-service"))
}

func TestSuitePublicPort_AllPSIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		psID string
		want int
	}{
		{psID: cryptoutilSharedMagic.OTLPServiceSMKMS, want: int(cryptoutilSharedMagic.SuiteTierPortMin)},
		{psID: cryptoutilSharedMagic.OTLPServiceSMIM, want: int(cryptoutilSharedMagic.DefaultPublicPortSmIM) + int(cryptoutilSharedMagic.ServiceToSuitePortOffset)},
		{psID: cryptoutilSharedMagic.OTLPServiceJoseJA, want: 28200},
		{psID: cryptoutilSharedMagic.OTLPServicePKICA, want: 28300},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: 28400},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityIDP, want: 28500},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRS, want: 28600},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRP, want: 28700},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentitySPA, want: int(cryptoutilSharedMagic.IMEnterpriseSessionAbsoluteMax)},
		{psID: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: 28900},
	}

	for _, tt := range tests {
		t.Run(tt.psID, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, SuitePublicPort(tt.psID))
		})
	}
}

func TestSuitePublicPort_UnknownPSID(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, SuitePublicPort("unknown-service"))
}

// -----------------------------------------------------------------------
// SQL identifier derivation tests (Task 3.2)
// -----------------------------------------------------------------------

func TestPSIDToSQLID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMKMS, input: cryptoutilSharedMagic.OTLPServiceSMKMS, want: "sm_kms"},
		{name: cryptoutilSharedMagic.OTLPServiceSMIM, input: cryptoutilSharedMagic.OTLPServiceSMIM, want: "sm_im"},
		{name: cryptoutilSharedMagic.OTLPServiceJoseJA, input: cryptoutilSharedMagic.OTLPServiceJoseJA, want: "jose_ja"},
		{name: cryptoutilSharedMagic.OTLPServicePKICA, input: cryptoutilSharedMagic.OTLPServicePKICA, want: "pki_ca"},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, input: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: "identity_authz"},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityIDP, input: cryptoutilSharedMagic.OTLPServiceIdentityIDP, want: "identity_idp"},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityRS, input: cryptoutilSharedMagic.OTLPServiceIdentityRS, want: "identity_rs"},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityRP, input: cryptoutilSharedMagic.OTLPServiceIdentityRP, want: "identity_rp"},
		{name: cryptoutilSharedMagic.OTLPServiceIdentitySPA, input: cryptoutilSharedMagic.OTLPServiceIdentitySPA, want: "identity_spa"},
		{name: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, input: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: "skeleton_template"},
		{name: "no-hyphens", input: "nohyphens", want: "nohyphens"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, PSIDToSQLID(tt.input))
		})
	}
}

func TestDatabaseName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMKMS, input: cryptoutilSharedMagic.OTLPServiceSMKMS, want: "sm_kms_database"},
		{name: cryptoutilSharedMagic.OTLPServiceJoseJA, input: cryptoutilSharedMagic.OTLPServiceJoseJA, want: "jose_ja_database"},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, input: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: "identity_authz_database"},
		{name: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, input: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: "skeleton_template_database"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, DatabaseName(tt.input))
		})
	}
}

func TestDatabaseUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMKMS, input: cryptoutilSharedMagic.OTLPServiceSMKMS, want: "sm_kms_database_user"},
		{name: cryptoutilSharedMagic.OTLPServiceJoseJA, input: cryptoutilSharedMagic.OTLPServiceJoseJA, want: "jose_ja_database_user"},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, input: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: "identity_authz_database_user"},
		{name: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, input: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: "skeleton_template_database_user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, DatabaseUser(tt.input))
		})
	}
}

func TestPostgresServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMKMS, input: cryptoutilSharedMagic.OTLPServiceSMKMS, want: cryptoutilSharedMagic.OTLPServiceSMKMS + PostgresServiceSuffix},
		{name: cryptoutilSharedMagic.OTLPServiceJoseJA, input: cryptoutilSharedMagic.OTLPServiceJoseJA, want: cryptoutilSharedMagic.OTLPServiceJoseJA + PostgresServiceSuffix},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, input: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: cryptoutilSharedMagic.OTLPServiceIdentityAuthz + PostgresServiceSuffix},
		{name: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, input: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate + PostgresServiceSuffix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, PostgresServiceName(tt.input))
		})
	}
}

func TestDBServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMKMS, input: cryptoutilSharedMagic.OTLPServiceSMKMS, want: cryptoutilSharedMagic.OTLPServiceSMKMS + DBServiceSuffix},
		{name: cryptoutilSharedMagic.OTLPServiceJoseJA, input: cryptoutilSharedMagic.OTLPServiceJoseJA, want: cryptoutilSharedMagic.OTLPServiceJoseJA + DBServiceSuffix},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, input: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: cryptoutilSharedMagic.OTLPServiceIdentityAuthz + DBServiceSuffix},
		{name: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, input: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate + DBServiceSuffix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, DBServiceName(tt.input))
		})
	}
}

// -----------------------------------------------------------------------
// Service name derivation tests (Task 3.3)
// -----------------------------------------------------------------------

func TestOTLPServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMKMS, input: cryptoutilSharedMagic.OTLPServiceSMKMS, want: OTLPServicePrefix + cryptoutilSharedMagic.OTLPServiceSMKMS},
		{name: cryptoutilSharedMagic.OTLPServiceJoseJA, input: cryptoutilSharedMagic.OTLPServiceJoseJA, want: OTLPServicePrefix + cryptoutilSharedMagic.OTLPServiceJoseJA},
		{name: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, input: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: OTLPServicePrefix + cryptoutilSharedMagic.OTLPServiceIdentityAuthz},
		{name: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, input: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: OTLPServicePrefix + cryptoutilSharedMagic.OTLPServiceSkeletonTemplate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, OTLPServiceName(tt.input))
		})
	}
}

func TestComposeServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		psID    string
		variant string
		want    string
	}{
		{
			name:    cryptoutilSharedMagic.OTLPServiceSMKMS + "-" + ComposeVariantSQLite1,
			psID:    cryptoutilSharedMagic.OTLPServiceSMKMS,
			variant: ComposeVariantSQLite1,
			want:    cryptoutilSharedMagic.OTLPServiceSMKMS + ComposeAppSuffix + ComposeVariantSQLite1,
		},
		{
			name:    cryptoutilSharedMagic.OTLPServiceSMKMS + "-" + ComposeVariantPostgres1,
			psID:    cryptoutilSharedMagic.OTLPServiceSMKMS,
			variant: ComposeVariantPostgres1,
			want:    cryptoutilSharedMagic.OTLPServiceSMKMS + ComposeAppSuffix + ComposeVariantPostgres1,
		},
		{
			name:    cryptoutilSharedMagic.OTLPServiceSMKMS + "-" + ComposeVariantPostgres2,
			psID:    cryptoutilSharedMagic.OTLPServiceSMKMS,
			variant: ComposeVariantPostgres2,
			want:    cryptoutilSharedMagic.OTLPServiceSMKMS + ComposeAppSuffix + ComposeVariantPostgres2,
		},
		{
			name:    cryptoutilSharedMagic.OTLPServiceJoseJA + "-" + ComposeVariantSQLite1,
			psID:    cryptoutilSharedMagic.OTLPServiceJoseJA,
			variant: ComposeVariantSQLite1,
			want:    cryptoutilSharedMagic.OTLPServiceJoseJA + ComposeAppSuffix + ComposeVariantSQLite1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, ComposeServiceName(tt.psID, tt.variant))
		})
	}
}

func TestValidOTLPServiceNames(t *testing.T) {
	t.Parallel()

	names := ValidOTLPServiceNames()

	require.Len(t, names, cryptoutilSharedMagic.PostgreSQLMaxIdleConns)
	require.Contains(t, names, OTLPServicePrefix+cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.Contains(t, names, OTLPServicePrefix+cryptoutilSharedMagic.OTLPServiceJoseJA)
	require.Contains(t, names, OTLPServicePrefix+cryptoutilSharedMagic.OTLPServiceSkeletonTemplate)
}

func TestValidComposeServiceNames(t *testing.T) {
	t.Parallel()

	names := ValidComposeServiceNames()
	allVariants := []string{ComposeVariantSQLite1, ComposeVariantSQLite2, ComposeVariantPostgres1, ComposeVariantPostgres2}

	// 10 PS-IDs × 4 variants = 40 names (sqlite-1, sqlite-2, postgresql-1, postgresql-2).
	require.Len(t, names, len(AllProductServices())*len(allVariants))
	require.Contains(t, names, cryptoutilSharedMagic.OTLPServiceSMKMS+ComposeAppSuffix+ComposeVariantSQLite1)
	require.Contains(t, names, cryptoutilSharedMagic.OTLPServiceSMKMS+ComposeAppSuffix+ComposeVariantSQLite2)
	require.Contains(t, names, cryptoutilSharedMagic.OTLPServiceSMKMS+ComposeAppSuffix+ComposeVariantPostgres1)
	require.Contains(t, names, cryptoutilSharedMagic.OTLPServiceSMKMS+ComposeAppSuffix+ComposeVariantPostgres2)
	require.Contains(t, names, cryptoutilSharedMagic.OTLPServiceJoseJA+ComposeAppSuffix+ComposeVariantSQLite1)
	require.Contains(t, names, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate+ComposeAppSuffix+ComposeVariantPostgres2)
}

// -----------------------------------------------------------------------
// Dockerfile derivation tests (Task 5.2)
// -----------------------------------------------------------------------

func TestDockerfileEntrypoint_AllPSIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		psID string
		want []string
	}{
		{psID: cryptoutilSharedMagic.OTLPServiceSMKMS, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceSMIM, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceJoseJA, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServicePKICA, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityIDP, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRS, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentityRP, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceIdentitySPA, want: []string{"/sbin/tini", "--"}},
		{psID: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, want: []string{"/sbin/tini", "--"}},
	}

	for _, tt := range tests {
		t.Run(tt.psID, func(t *testing.T) {
			t.Parallel()

			got := DockerfileEntrypoint(tt.psID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDockerfileEntrypoint_UnknownPSID(t *testing.T) {
	t.Parallel()

	require.Nil(t, DockerfileEntrypoint("unknown-service"))
}
