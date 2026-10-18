package tls

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testPublishTierID       = "test-tier"
	testTLSConfigCertPrefix = "tls-mixed-ca-cert-pem: "
	testTLSConfigKeyPrefix  = "tls-mixed-ca-key-pem: "
	testPrivateKeyName      = "root.key"
	testCertificateName     = "root.crt"
)

func TestPublishValidationHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T, path string)
		directory bool
		wantErr   string
	}{
		{
			name: "missing destination",
		},
		{
			name: "directory destination exists",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.Mkdir(path, 0o700))
			},
			directory: true,
		},
		{
			name: "file destination exists",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("existing"), 0o600))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "destination")
			if tc.setup != nil {
				tc.setup(t, path)
			}

			err := validatePublishDestination(path, tc.directory)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCertificateAndPrivateKeyValidation(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	wrongKeyFixture := newPublishValidationFixture(t)
	otherFixture := newPublishValidationFixture(t)

	tests := []struct {
		name       string
		certPEM    []byte
		keyPEM     []byte
		certErr    string
		privateErr string
	}{
		{
			name:    "matching key and certificate",
			certPEM: fixture.certificatePEM,
			keyPEM:  fixture.privateKeyPEM,
		},
		{
			name:    testTLSCertificatePEMInvalid,
			certPEM: []byte("not pem"),
			keyPEM:  fixture.privateKeyPEM,
			certErr: testTLSInvalidPEM,
		},
		{
			name:    "PEM with no certificate block",
			certPEM: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: []byte("ignored")}),
			keyPEM:  fixture.privateKeyPEM,
			certErr: testTLSNoCertificates,
		},
		{
			name:    "malformed certificate DER",
			certPEM: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: []byte(cryptoutilSharedMagic.DevSetupTestInvalidName)}),
			keyPEM:  fixture.privateKeyPEM,
			certErr: "parse staged certificate",
		},
		{
			name:       "incorrect private key PEM type",
			certPEM:    fixture.certificatePEM,
			keyPEM:     pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey, Bytes: fixture.privateKeyDER}),
			privateErr: testTLSSinglePKCS8PEMBlock,
		},
		{
			name:       "malformed PKCS8 private key",
			certPEM:    fixture.certificatePEM,
			keyPEM:     pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: []byte(cryptoutilSharedMagic.DevSetupTestInvalidName)}),
			privateErr: "parse staged private key",
		},
		{
			name:       "extra private key PEM block",
			certPEM:    fixture.certificatePEM,
			keyPEM:     append(fixture.privateKeyPEM, fixture.privateKeyPEM...),
			privateErr: testTLSSinglePKCS8PEMBlock,
		},
		{
			name:       "private key does not match certificate",
			certPEM:    fixture.certificatePEM,
			keyPEM:     wrongKeyFixture.privateKeyPEM,
			privateErr: testTLSPrivateKeyDoesNotMatch,
		},
		{
			name:    testTLSCertificateChainInvalidSignature,
			certPEM: append(append([]byte{}, fixture.certificatePEM...), otherFixture.certificatePEM...),
			keyPEM:  fixture.privateKeyPEM,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			certificates, err := parseCertificatePEM(testCertificateName, tc.certPEM)
			if tc.certErr != "" {
				require.ErrorContains(t, err, tc.certErr)

				return
			}

			require.NoError(t, err)

			err = validatePrivateKeyMatchesCertificate(testPrivateKeyName, tc.keyPEM, certificates[0])
			if tc.privateErr != "" {
				require.ErrorContains(t, err, tc.privateErr)

				return
			}

			require.NoError(t, err)

			if tc.name == "certificate chain has invalid signature" {
				require.ErrorContains(t, validateCertificateChain(testCertificateName, certificates), "invalid link")
			}
		})
	}
}

func TestValidateGeneratedMaterial(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	wrongKeyFixture := newPublishValidationFixture(t)
	otherFixture := newPublishValidationFixture(t)

	tests := []struct {
		name    string
		setup   func(t *testing.T, stagingRoot string)
		wantErr string
	}{
		{
			name:    "missing tier directory",
			wantErr: testTLSWalkStagedPath,
		},
		{
			name: "empty certificate tree",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()
				require.NoError(t, os.Mkdir(filepath.Join(stagingRoot, testPublishTierID), 0o700))
			},
			wantErr: "incomplete",
		},
		{
			name: "valid certificate material",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()
				writeValidPublishTree(t, stagingRoot, fixture)
			},
		},
		{
			name: "invalid certificate material",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()
				writeValidPublishTree(t, stagingRoot, fixture)
				writePublishFile(t, filepath.Join(stagingRoot, testPublishTierID, "invalid.crt"), []byte("not pem"), cryptoutilSharedMagic.PKIInitPublicCertFileMode)
			},
			wantErr: "invalid PEM",
		},
		{
			name: testTLSInvalidCertificateChain,
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()
				writeValidPublishTree(t, stagingRoot, fixture)
				certPath := filepath.Join(stagingRoot, testPublishTierID, testCertificateName)
				writePublishFile(t, certPath, append(append([]byte{}, fixture.certificatePEM...), otherFixture.certificatePEM...), cryptoutilSharedMagic.PKIInitPublicCertFileMode)
			},
			wantErr: "invalid link",
		},
		{
			name: "private key does not match staged certificate",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()
				writeValidPublishTree(t, stagingRoot, fixture)
				keyPath := filepath.Join(stagingRoot, testPublishTierID, testPrivateKeyName)
				writePublishFile(t, keyPath, wrongKeyFixture.privateKeyPEM, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
			},
			wantErr: testTLSPrivateKeyDoesNotMatch,
		},
		{
			name: "empty staged file",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()
				writeValidPublishTree(t, stagingRoot, fixture)
				writePublishFile(t, filepath.Join(stagingRoot, testPublishTierID, "empty.crt"), nil, cryptoutilSharedMagic.PKIInitPublicCertFileMode)
			},
			wantErr: cryptoutilSharedMagic.EMPTY,
		},
		{
			name: "missing paired certificate",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()

				tierPath := filepath.Join(stagingRoot, testPublishTierID)
				require.NoError(t, os.Mkdir(tierPath, 0o700))
				writePublishFile(t, filepath.Join(tierPath, testPrivateKeyName), fixture.privateKeyPEM, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
			},
			wantErr: "paired with staged key",
		},
		{
			name: testTLSInvalidPairedCertificate,
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()

				tierPath := filepath.Join(stagingRoot, testPublishTierID)
				require.NoError(t, os.Mkdir(tierPath, 0o700))
				writePublishFile(t, filepath.Join(tierPath, testPrivateKeyName), fixture.privateKeyPEM, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
				writePublishFile(t, filepath.Join(tierPath, testCertificateName), []byte("not pem"), cryptoutilSharedMagic.PKIInitPublicCertFileMode)
			},
			wantErr: testTLSInvalidPEM,
		},
		{
			name: "missing staged TLS config",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()

				tierPath := filepath.Join(stagingRoot, testPublishTierID)
				require.NoError(t, os.Mkdir(tierPath, 0o700))
				writePublishFile(t, filepath.Join(tierPath, testCertificateName), fixture.certificatePEM, cryptoutilSharedMagic.PKIInitPublicCertFileMode)
				writePublishFile(t, filepath.Join(tierPath, testPrivateKeyName), fixture.privateKeyPEM, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
				writePublishFile(t, filepath.Join(tierPath, "root.p12"), []byte("pkcs12"), cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
			},
			wantErr: "inspect staged tls-config.yml",
		},
		{
			name: "invalid staged file mode",
			setup: func(t *testing.T, stagingRoot string) {
				t.Helper()
				writeValidPublishTree(t, stagingRoot, fixture)
				certPath := filepath.Join(stagingRoot, testPublishTierID, testCertificateName)
				require.NoError(t, os.Chmod(certPath, cryptoutilSharedMagic.FilePermAllReadWrite))
			},
			wantErr: testTLSPermissions,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stagingRoot := t.TempDir()
			if tc.setup != nil {
				tc.setup(t, stagingRoot)
			}

			err := validateGeneratedMaterial(stagingRoot, testPublishTierID)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateAdminCABundle(t *testing.T) {
	t.Parallel()

	caFixture := newPublishValidationFixture(t)
	nonCAFixture := newPublishValidationFixtureWithCA(t, false)

	tests := []struct {
		name    string
		data    []byte
		missing bool
		wantErr string
	}{
		{name: "valid CA bundle", data: caFixture.certificatePEM},
		{name: "missing bundle", missing: true, wantErr: "read staged admin CA bundle"},
		{name: testTLSInvalidPEM, data: []byte("not pem"), wantErr: testTLSInvalidPEM},
		{name: testTLSNoCertificates, data: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: []byte("ignored")}), wantErr: testTLSNoCertificates},
		{name: testTLSNonCACertificate, data: nonCAFixture.certificatePEM, wantErr: "not a CA"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), cryptoutilSharedMagic.PKIInitAdminCABundleFilename)
			if !tc.missing {
				writePublishFile(t, path, tc.data, cryptoutilSharedMagic.PKIInitPublicCertFileMode)
			}

			err := validateAdminCABundle(path)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPublishGeneratedMaterial_RejectsMissingRenameFunction(t *testing.T) {
	t.Parallel()

	err := publishGeneratedMaterial(nil, nil)

	require.ErrorContains(t, err, "rename function is required")
}

func TestRollbackPublishedMaterial_RemoveFailure(t *testing.T) {
	t.Parallel()

	err := rollbackPublishedMaterialWithRemove(
		func(string) error {
			return errors.New("injected rollback removal failure")
		},
		[]publishPath{{destination: "published-output", published: true}},
	)

	require.ErrorContains(t, err, "remove partially published output")
	require.ErrorContains(t, err, "injected rollback removal failure")
}

type publishValidationFixture struct {
	certificatePEM []byte
	privateKeyPEM  []byte
	privateKeyDER  []byte
	certificate    *x509.Certificate
	privateKey     *ecdsa.PrivateKey
}

func newPublishValidationFixture(t *testing.T) publishValidationFixture {
	t.Helper()

	return newPublishValidationFixtureWithCA(t, true)
}

func newPublishValidationFixtureWithCA(t *testing.T, isCA bool) publishValidationFixture {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	id, err := googleUuid.NewV7()
	require.NoError(t, err)

	serial := new(big.Int).SetBytes(id[:])
	template := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             time.Now().UTC().Add(-time.Minute),
		NotAfter:              time.Now().UTC().Add(time.Hour),
		IsCA:                  isCA,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	certificateDER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)
	certificate, err := x509.ParseCertificate(certificateDER)
	require.NoError(t, err)

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	return publishValidationFixture{
		certificatePEM: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: certificateDER}),
		privateKeyPEM:  pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: privateKeyDER}),
		privateKeyDER:  privateKeyDER,
		certificate:    certificate,
		privateKey:     privateKey,
	}
}

func writeValidPublishTree(t *testing.T, stagingRoot string, fixture publishValidationFixture) {
	t.Helper()

	tierPath := filepath.Join(stagingRoot, testPublishTierID)
	require.NoError(t, os.Mkdir(tierPath, 0o700))
	writePublishFile(t, filepath.Join(tierPath, testCertificateName), fixture.certificatePEM, cryptoutilSharedMagic.PKIInitPublicCertFileMode)
	writePublishFile(t, filepath.Join(tierPath, testPrivateKeyName), fixture.privateKeyPEM, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
	writePublishFile(t, filepath.Join(tierPath, "root.p12"), []byte("pkcs12"), cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
	writePublishFile(t, filepath.Join(stagingRoot, "tls-config.yml"), makeTLSConfig(fixture.certificatePEM, fixture.privateKeyPEM), cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
	writePublishFile(t, filepath.Join(stagingRoot, cryptoutilSharedMagic.PKIInitAdminCABundleFilename), fixture.certificatePEM, cryptoutilSharedMagic.PKIInitPublicCertFileMode)
}

func writePublishFile(t *testing.T, path string, data []byte, mode os.FileMode) {
	t.Helper()

	if _, err := os.Lstat(path); err == nil {
		require.NoError(t, os.Chmod(path, cryptoutilSharedMagic.KeyFilePermissions))
	} else {
		require.ErrorIs(t, err, os.ErrNotExist)
	}

	require.NoError(t, os.WriteFile(path, data, mode))
	require.NoError(t, os.Chmod(path, mode))
}

func makeTLSConfig(certificatePEM, privateKeyPEM []byte) []byte {
	return []byte(testTLSConfigCertPrefix + base64.StdEncoding.EncodeToString(certificatePEM) + "\n" +
		testTLSConfigKeyPrefix + base64.StdEncoding.EncodeToString(privateKeyPEM) + "\n")
}
