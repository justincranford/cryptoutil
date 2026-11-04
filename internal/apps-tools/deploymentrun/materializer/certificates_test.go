// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"net"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestBuildMaterializedFiles(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	tests := []struct {
		name                  string
		mutate                func(*testing.T, *testPKIFixture)
		now                   time.Time
		adminDNS              string
		publicDNSNames        []string
		omitPostgreSQLBinding bool
		wantErr               string
		wantFileNum           int
	}{
		{name: "verified role-separated files", wantFileNum: 10},
		{
			name: "unexpected TLS bundle key",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.tlsBundle["root.key"] = []byte("forbidden")
			},
			wantErr: testMissingOrUnexpectedDataKeysError,
		},
		{
			name: "malformed certificate PEM",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.tlsBundle[SecretDataPublicCertificate] = []byte(testMalformedError)
			},
			wantErr: "parse certificate for public-server",
		},
		{
			name: "malformed private key PEM",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.tlsBundle[SecretDataPublicPrivateKey] = []byte(testMalformedError)
			},
			wantErr: "parse private key for public-server",
		},
		{
			name: "missing private key",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				removeDataKey(fixture.health, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey)
			},
			wantErr: testMissingOrUnexpectedDataKeysError,
		},
		{
			name: "empty certificate value",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.otlp[cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate] = nil
			},
			wantErr: "missing required data key",
		},
		{
			name: "mismatched certificate key",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey] = fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminPrivateKey]
			},
			wantErr: "do not match",
		},
		{
			name: "CA cannot be leaf",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate] = fixture.rootPEM
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey] = fixture.rootKeyPEM
			},
			wantErr: "CA certificate cannot be materialized as a leaf",
		},
		{
			name: "wrong server extended usage",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				root, err := parseCertificatePEM(fixture.rootPEM)
				require.NoError(t, err)
				leaf := newTestLeaf(t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageClientAuth,
					[]string{testPostgreSQLServiceDNSName}, nil, false, fixture.rootKeyPEM)
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate] = leaf.certPEM
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey] = leaf.keyPEM
			},
			wantErr: testExtendedKeyUsageError,
		},
		{
			name: "client certificate does not require a server DNS SAN",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				root, err := parseCertificatePEM(fixture.rootPEM)
				require.NoError(t, err)
				leaf := newTestLeaf(t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageClientAuth,
					[]string{testWrongDNSName}, nil, false, fixture.rootKeyPEM)
				fixture.health[cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate] = leaf.certPEM
				fixture.health[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey] = leaf.keyPEM
			},
			wantFileNum: 10,
		},
		{
			name: "public leaf must contain derived service FQDN",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				root, err := parseCertificatePEM(fixture.rootPEM)
				require.NoError(t, err)
				leaf := newTestLeaf(t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageServerAuth,
					[]string{testWrongDNSName}, nil, false, fixture.rootKeyPEM)
				fixture.tlsBundle[SecretDataPublicCertificate] = leaf.certPEM
				fixture.tlsBundle[SecretDataPublicPrivateKey] = leaf.keyPEM
			},
			wantErr: "SAN does not contain its expected DNS identity",
		},
		{
			name: "public leaf must contain derived pod FQDN",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				root, err := parseCertificatePEM(fixture.rootPEM)
				require.NoError(t, err)

				serviceDNS := fixture.service.Name + ".namespace-17.svc." + fixture.config.ClusterDomain
				leaf := newTestLeaf(t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageServerAuth,
					[]string{serviceDNS}, nil, false, fixture.rootKeyPEM)
				fixture.tlsBundle[SecretDataPublicCertificate] = leaf.certPEM
				fixture.tlsBundle[SecretDataPublicPrivateKey] = leaf.keyPEM
			},
			wantErr: testEveryRequiredPublicDNSIdentityError,
		},
		{
			name: "public leaf must contain controller supplied external SNI",
			publicDNSNames: []string{
				"identity-authz-pg-i02.namespace-17.svc.cluster.local",
				"identity-authz-pg-1.identity-authz-pg-pods.namespace-17.svc.cluster.local",
				testPostgreSQLServiceDNSName,
				testExternalLoginSNI,
			},
			wantErr: testEveryRequiredPublicDNSIdentityError,
		},
		{
			name:           "expected public names must include derived pod FQDN",
			publicDNSNames: []string{testPostgreSQLServiceDNSName},
			wantErr:        testMissingDerivedServiceOrPodFQDNError,
		},
		{
			name:    "expired leaf",
			now:     fixture.now.Add(2 * time.Hour),
			wantErr: testValidityPeriodError,
		},
		{
			name:    "not-yet-valid leaf",
			now:     fixture.now.Add(-2 * time.Hour),
			wantErr: testValidityPeriodError,
		},
		{
			name: "wrong issuer",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				_, otherRoot, _ := newTestRoot(t, fixture.now)
				fixture.trust.PublicServerPEM = otherRoot
			},
			wantErr: "chain or SAN verification failed",
		},
		{
			name: "non-CA trust anchor",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.trust.HealthClientPEM = fixture.health[cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate]
			},
			wantErr: "trust anchor is not a CA certificate",
		},
		{
			name: "admin certificate must contain loopback IP",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				root, err := parseCertificatePEM(fixture.rootPEM)
				require.NoError(t, err)
				leaf := newTestLeaf(t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageServerAuth,
					[]string{testExpectedAdministrativeDNS(t, fixture.config, fixture.service)}, nil, false, fixture.rootKeyPEM)
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminCertificate] = leaf.certPEM
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminPrivateKey] = leaf.keyPEM
			},
			wantErr: "does not contain 127.0.0.1",
		},
		{
			name: "non-pg pool omits PostgreSQL pair",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				config := testInstanceConfigForPool(t, cryptoutilSharedMagic.DeploymentMaterializerPoolMem)
				fixture.config = config
				setTestPublicLeafDNS(t, fixture, testNamespace)
				setTestAdminLeafDNS(t, fixture, testNamespace)
				fixture.trust.PostgreSQLClientPEM = nil
				removeDataKey(fixture.tlsBundle, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLCertificate)
				removeDataKey(fixture.tlsBundle, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLPrivateKey)
			},
			wantFileNum: 8,
		},
		{
			name: "disk pool omits PostgreSQL pair",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				config := testInstanceConfigForPool(t, cryptoutilSharedMagic.DeploymentMaterializerPoolDisk)
				fixture.config = config
				setTestPublicLeafDNS(t, fixture, testNamespace)
				setTestAdminLeafDNS(t, fixture, testNamespace)
				fixture.trust.PostgreSQLClientPEM = nil
				removeDataKey(fixture.tlsBundle, SecretDataPostgreSQLCertificate)
				removeDataKey(fixture.tlsBundle, SecretDataPostgreSQLPrivateKey)
			},
			wantFileNum: 8,
		},
		{
			name: "non-pg pool rejects PostgreSQL files",
			mutate: func(t *testing.T, fixture *testPKIFixture) {
				config := testInstanceConfigForPool(t, cryptoutilSharedMagic.DeploymentMaterializerPoolMem)
				fixture.config = config
				setTestPublicLeafDNS(t, fixture, testNamespace)
				setTestAdminLeafDNS(t, fixture, testNamespace)
				fixture.trust.PostgreSQLClientPEM = nil
			},
			wantErr: testMissingOrUnexpectedDataKeysError,
		},
		{
			name: "PostgreSQL leaf must be a client certificate",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLCertificate] = fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate]
				fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLPrivateKey] = fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]
			},
			wantErr: testExtendedKeyUsageError,
		},
		{
			name: "PostgreSQL leaf common name must match trusted pod binding",
			mutate: func(_ *testing.T, fixture *testPKIFixture) {
				binding := fixture.trust.PostgreSQLClientBindings[fixture.service.PodName]
				binding.CommonName = testAnotherPodClientCommonName
				fixture.trust.PostgreSQLClientBindings[fixture.service.PodName] = binding
			},
			wantErr: "common name does not match",
		},
		{
			name:                  "PostgreSQL binding is required for pg pool",
			omitPostgreSQLBinding: true,
			wantErr:               "binding is required",
		},
		{
			name:     "invalid administrative expected DNS",
			adminDNS: "invalid/name",
			wantErr:  testExpectedAdministrativeDNSNameInvalidError,
		},
		{
			name:     "administrative certificate SAN differs from expected pod DNS",
			adminDNS: testWrongDNSName,
			wantErr:  "certificate SAN does not contain its expected DNS identity",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			current := cloneTestPKIFixture(fixture)
			if testCase.mutate != nil {
				testCase.mutate(t, current)
			}

			now := current.now
			if !testCase.now.IsZero() {
				now = testCase.now
			}

			adminDNS := testCase.adminDNS
			if adminDNS == "" {
				adminDNS = testExpectedAdministrativeDNS(t, current.config, current.service)
			}

			publicDNSNames := testCase.publicDNSNames
			if publicDNSNames == nil {
				publicDNSNames = testExpectedPublicDNSNames(current.config, current.service, testNamespace)
			}

			var postgresBinding *PostgreSQLClientBinding

			if current.config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG && !testCase.omitPostgreSQLBinding {
				binding := current.trust.PostgreSQLClientBindings[current.service.PodName]
				postgresBinding = &binding
			}

			files, err := buildMaterializedFiles(
				current.config, current.service, testNamespace, adminDNS, publicDNSNames, postgresBinding,
				current.trust, current.tlsBundle, current.health, current.otlp, now,
			)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)
				require.Nil(t, files)

				return
			}

			require.NoError(t, err)
			require.Len(t, files, testCase.wantFileNum)
			require.Contains(t, files, cryptoutilSharedMagic.DeploymentMaterializerPublicServerCertificateFile)
			require.Contains(t, files, cryptoutilSharedMagic.DeploymentMaterializerHealthClientPrivateKeyFile)

			if current.config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG {
				require.Contains(t, files, cryptoutilSharedMagic.DeploymentMaterializerPostgreSQLClientCertificateFile)
			} else {
				require.NotContains(t, files, cryptoutilSharedMagic.DeploymentMaterializerPostgreSQLClientCertificateFile)
			}
		})
	}
}

func TestStrictPEMAndDataKeyParsing(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)

	tests := []struct {
		name     string
		certPEM  []byte
		keyPEM   []byte
		trustPEM []byte
		wantErr  string
	}{
		{name: "valid certificate chain", certPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate], keyPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey], trustPEM: fixture.rootPEM},
		{name: "empty certificate", keyPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey], trustPEM: fixture.rootPEM, wantErr: "certificate PEM is empty"},
		{name: "non-PEM certificate", certPEM: []byte(cryptoutilSharedMagic.DevSetupTestInvalidName), keyPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey], trustPEM: fixture.rootPEM, wantErr: testNonCertificateDataError},
		{
			name:    "wrong certificate block",
			certPEM: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: []byte{1}}),
			keyPEM:  fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey], trustPEM: fixture.rootPEM, wantErr: "invalid block",
		},
		{
			name:    "invalid certificate DER",
			certPEM: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: []byte{1}}),
			keyPEM:  fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey], trustPEM: fixture.rootPEM, wantErr: "parse X.509 certificate",
		},
		{
			name:    "trailing certificate data",
			certPEM: append(append([]byte(nil), fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate]...), []byte("trailing")...),
			keyPEM:  fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey], trustPEM: fixture.rootPEM, wantErr: testNonCertificateDataError,
		},
		{
			name:     "multiple private keys",
			certPEM:  fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate],
			keyPEM:   append(append([]byte(nil), fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]...), fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]...),
			trustPEM: fixture.rootPEM, wantErr: "one unencrypted key",
		},
		{
			name:     "unsupported private key block",
			certPEM:  fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate],
			keyPEM:   pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: []byte{1}}),
			trustPEM: fixture.rootPEM, wantErr: testUnsupportedValue,
		},
		{
			name:     "invalid private key DER",
			certPEM:  fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate],
			keyPEM:   pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: []byte{1}}),
			trustPEM: fixture.rootPEM, wantErr: "parse private key PEM",
		},
		{
			name:    "trust anchor must be CA",
			certPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate], keyPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey],
			trustPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate], wantErr: "not a CA",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			certificates, certErr := parseCertificatePEM(testCase.certPEM)
			if testCase.wantErr != "" && certErr != nil {
				require.ErrorContains(t, certErr, testCase.wantErr)

				return
			}

			require.NoError(t, certErr)
			require.NotEmpty(t, certificates)

			if _, keyErr := parsePrivateKeyPEM(testCase.keyPEM); keyErr != nil {
				require.ErrorContains(t, keyErr, testCase.wantErr)

				return
			}

			roots, trustErr := parseTrustRoots(testCase.trustPEM)
			if testCase.wantErr != "" {
				require.ErrorContains(t, trustErr, testCase.wantErr)

				return
			}

			require.NoError(t, trustErr)
			require.NotNil(t, roots)
			require.True(t, bytes.Contains(testCase.certPEM, []byte(cryptoutilSharedMagic.StringPEMTypeCertificate)))
		})
	}

	keys := map[string][]byte{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate: []byte("cert"), cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey: []byte("key")}

	keyCases := []struct {
		name    string
		data    map[string][]byte
		wantErr string
	}{
		{name: "exact keys", data: keys},
		{name: "missing key", data: map[string][]byte{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate: []byte("cert")}, wantErr: testMissingOrUnexpectedError},
		{name: "empty value", data: map[string][]byte{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate: []byte("cert"), cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey: nil}, wantErr: "missing required"},
		{name: "extra key", data: map[string][]byte{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate: []byte("cert"), cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey: []byte("key"), "ca.key": []byte("extra")}, wantErr: testMissingOrUnexpectedError},
	}
	for _, testCase := range keyCases {
		t.Run("data keys "+testCase.name, func(t *testing.T) {
			t.Parallel()

			err := requireExactDataKeys(testCase.data, []string{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey})
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}

	require.True(t, containsDNSName([]string{"expected.example"}, "expected.example"))
	require.False(t, containsDNSName([]string{testWrongDNSName}, "expected.example"))
	require.True(t, containsIP([]net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}, net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)))
	require.False(t, containsIP([]net.IP{net.ParseIP("127.0.0.2")}, net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)))
	require.ErrorContains(t, errParsePrivateKey([]byte("not PEM")), testMalformedError)
}

func errParsePrivateKey(data []byte) error {
	_, err := parsePrivateKeyPEM(data)

	return err
}

func TestCertificateRoleVerificationRejectsBadKeyUsage(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	root, err := parseCertificatePEM(fixture.rootPEM)
	require.NoError(t, err)
	badLeaf := newTestLeafWithKeyUsage(
		t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageServerAuth,
		[]string{testPostgreSQLServiceDNSName}, nil, false, fixture.rootKeyPEM,
		x509.KeyUsageKeyEncipherment,
	)
	err = validateCertificateMaterial(certificateMaterial{
		certPEM: badLeaf.certPEM, keyPEM: badLeaf.keyPEM, trustPEM: fixture.rootPEM,
		expectedDNS: testPostgreSQLServiceDNSName,
		usage:       x509.ExtKeyUsageServerAuth, certOutput: cryptoutilSharedMagic.DeploymentMaterializerPublicServerCertificateFile, keyOutput: cryptoutilSharedMagic.DeploymentMaterializerPublicServerPrivateKeyFile,
	}, fixture.now)
	require.ErrorContains(t, err, "digital-signature")
	require.False(
		t,
		bytes.Contains(
			fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate],
			[]byte(cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey),
		),
	)
}

func TestDecodeTrustBundleRejectsEmptyRoots(t *testing.T) {
	t.Parallel()

	_, err := parseTrustRoots(nil)
	require.ErrorContains(t, err, "certificate PEM is empty")
}

func TestCertificateInputTrailingPrivateKeyData(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	keyWithTrailing := append(append([]byte(nil), fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]...), []byte("ignored")...)
	_, err := parsePrivateKeyPEM(keyWithTrailing)
	require.ErrorContains(t, err, "one unencrypted key")
	require.True(
		t,
		strings.Contains(
			string(fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]),
			cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey,
		),
	)
}
