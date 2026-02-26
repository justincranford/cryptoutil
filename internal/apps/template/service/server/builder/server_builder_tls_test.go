// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"crypto/elliptic"
	"fmt"
	"strings"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"testing/fstest"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestMergedMigrations_ReadDir_SubPath(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/subdir/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template_sub (id TEXT);"),
		},
	}

	domainFS := fstest.MapFS{
		"migrations/subdir/2001_domain.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE domain_sub (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     domainFS,
		domainPath:   "migrations",
	}

	// Read merged subdirectory.
	entries, err := merged.ReadDir("subdir")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 1)
}

// TestMergedMigrations_Open_RootDir tests Open with current directory.
func TestMergedMigrations_Open_RootDir(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	domainFS := fstest.MapFS{
		"migrations/2001_domain.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE domain (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     domainFS,
		domainPath:   "migrations",
	}

	// Open root directory (should work for both "." and "").
	rootFile, err := merged.Open(".")
	require.NoError(t, err)
	require.NotNil(t, rootFile)

	defer func() { _ = rootFile.Close() }()

	emptyPathFile, err := merged.Open("")
	require.NoError(t, err)
	require.NotNil(t, emptyPathFile)

	defer func() { _ = emptyPathFile.Close() }()
}

// TestMergedMigrations_ReadDir_NilDomainFS tests ReadDir when domainFS is nil.
func TestMergedMigrations_ReadDir_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Read merged directory with nil domainFS.
	entries, err := merged.ReadDir(".")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 1)
}

// TestMergedMigrations_Open_NilDomainFS tests Open when domainFS is nil.
func TestMergedMigrations_Open_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Open template file when domain FS is nil.
	templateFile, err := merged.Open("1001_template.up.sql")
	require.NoError(t, err)
	require.NotNil(t, templateFile)

	defer func() { _ = templateFile.Close() }()
}

// TestMergedMigrations_ReadFile_NilDomainFS tests ReadFile when domainFS is nil.
func TestMergedMigrations_ReadFile_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template_only (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Read template file when domain FS is nil.
	templateData, err := merged.ReadFile("1001_template.up.sql")
	require.NoError(t, err)
	require.Contains(t, string(templateData), "CREATE TABLE template_only")
}

// TestMergedMigrations_Stat_NilDomainFS tests Stat when domainFS is nil.
func TestMergedMigrations_Stat_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template_stat (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Stat template file when domain FS is nil.
	templateInfo, err := merged.Stat("1001_template.up.sql")
	require.NoError(t, err)
	require.NotNil(t, templateInfo)
}

// generateTestCA generates a valid CA certificate and key for testing.
func generateTestCA(t *testing.T) (caCertPEM, caKeyPEM []byte) {
	t.Helper()

	// Generate CA key pair.
	caKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	// Generate CA certificate.
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * cryptoutilSharedMagic.HoursPerDay * time.Hour //nolint:mnd // Duration calculation.
	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "Test CA", duration)
	require.NoError(t, err)
	require.Len(t, caSubjects, 1)

	caCert := caSubjects[0].KeyMaterial.CertificateChain[0]

	// Serialize CA certificate to PEM.
	caCertPEM = pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: caCert.Raw,
	})

	// Serialize CA private key to PEM.
	caKeyBytes, err := x509.MarshalPKCS8PrivateKey(caKeyPair.Private)
	require.NoError(t, err)

	caKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
		Bytes: caKeyBytes,
	})

	return caCertPEM, caKeyPEM
}

// getMinimalSettings returns minimal valid settings for testing.
// Uses same pattern as application_listener_test.go.
func getMinimalSettings() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:                    true,
		VerboseMode:                false,
		DatabaseURL:                fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(googleUuid.Must(googleUuid.NewV7()).String(), "-", "")),
		OTLPService:                "template-service-test",
		OTLPEnabled:                false,
		OTLPEndpoint:               cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		LogLevel:                   cryptoutilSharedMagic.DefaultLogLevelInfo,
		BrowserSessionAlgorithm:    cryptoutilSharedMagic.DefaultServiceSessionAlgorithm,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		BrowserSessionJWEAlgorithm: cryptoutilSharedMagic.JoseAlgRSAOAEP,
		BrowserSessionExpiration:   15 * time.Minute,
		ServiceSessionAlgorithm:    cryptoutilSharedMagic.DefaultServiceSessionAlgorithm,
		ServiceSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		ServiceSessionJWEAlgorithm: cryptoutilSharedMagic.JoseAlgRSAOAEP,
		ServiceSessionExpiration:   1 * time.Hour,
		SessionIdleTimeout:         cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Minute,
		SessionCleanupInterval:     1 * time.Hour,
		BindPublicProtocol:         cryptoutilSharedMagic.ProtocolHTTPS,
		BindPublicAddress:          cryptoutilSharedMagic.IPv4Loopback,
		BindPublicPort:             0,
		BindPrivateProtocol:        cryptoutilSharedMagic.ProtocolHTTPS,
		BindPrivateAddress:         cryptoutilSharedMagic.IPv4Loopback,
		BindPrivatePort:            0,
		TLSPublicMode:              cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
		TLSPrivateMode:             cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
		TLSPublicDNSNames:          []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		TLSPublicIPAddresses:       []string{cryptoutilSharedMagic.IPv4Loopback},
		TLSPrivateDNSNames:         []string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		TLSPrivateIPAddresses:      []string{cryptoutilSharedMagic.IPv4Loopback},
	}
}
