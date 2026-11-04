// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testAPIEndpoint   = "https://192.0.2.10:443"
	testAPIServerName = "kubernetes.test.svc"

	testInvalidAPIServerName                       = "bad_name"
	testForbiddenHTTPStatus                        = "HTTP 403"
	testParentHealthSecretName                     = "../health"
	testRSA2048Name                                = "RSA 2048"
	testApprovedKeySizeError                       = "below the approved size"
	testECDSAP384Name                              = "ECDSA P-384"
	testECDSAP224RejectedName                      = "ECDSA P-224 rejected"
	testAnotherPodClientCommonName                 = "another-pod-client"
	testOpaqueV1SecretError                        = "Opaque v1 Secret"
	testDuplicateSecretNameError                   = "Secret name is duplicated"
	testAbsoluteFrameworkConfigFile                = "/etc/framework.yaml"
	testDotPath                                    = "."
	testInvalidSecretName                          = "bad/name"
	testParentSecretName                           = "../secret"
	testInvalidCredentialMountError                = "credential Secret mount is invalid"
	testExistingOutputError                        = "already exists"
	testCleanAbsolutePathError                     = "clean absolute path"
	testFileOperationClose                         = "close"
	testCanceledError                              = "canceled"
	testThirdOccurrence                            = 3
	testInvalidPostgreSQLBindingError              = "binding is invalid"
	testDatabasePodZero                            = "db-0"
	testApplicationDatabaseRole                    = "application_role"
	testDecodeKubernetesSecretError                = "decode Kubernetes Secret"
	testIdentityLabelError                         = "identity label"
	testHealthSecretName                           = "health-1"
	testInvalidNamespaceCaseName                   = "invalid namespace"
	testInvalidSecretNameCaseName                  = "invalid Secret name"
	testInvalidPodIdentityCaseName                 = "invalid pod identity"
	testUntrustedKubernetesDNSAPIEndpoint          = "https://kubernetes.test.svc:443"
	testEndpointOrTLSServerNameError               = "endpoint or TLS server name"
	testInvalidIdentityError                       = "identity is invalid"
	testCurveNotApprovedError                      = "curve is not approved"
	testPostgreSQLServiceDNSName                   = "identity-authz-pg." + testNamespace + ".svc.cluster.local"
	testExtendedKeyUsageError                      = "extended key usage"
	testEveryRequiredPublicDNSIdentityError        = "every required public DNS identity"
	testExpectedAdministrativeDNSNameInvalidError  = "expected administrative DNS name is invalid"
	testFilePathInvalidError                       = "file path is invalid"
	testFixedPKILayoutError                        = "fixed PKI layout"
	testFrameworkConfigFilenameWithNUL             = "framework\x00.yaml"
	testDatabasePodNameOne                         = "db-1"
	testDatabaseClientCommonNameZero               = "client-db-0"
	testPostgreSQLPodNameOne                       = "identity-authz-pg-1"
	testUnconfiguredPostgreSQLPodName              = "identity-authz-pg-9"
	testIdentityAuthzClientDatabaseRole            = "identity_authz_client"
	testExactlyCoverConfiguredPodsError            = "exactly cover configured pods"
	testCreateMaterializerOutputFileError          = "create materializer output file"
	testExpectedSecretRequestCount                 = 3
	testPrivateKeyEncodingCaseCount                = 3
	testNamespace                                  = "namespace-17"
	testInvalidNamespaceValue                      = "namespace/17"
	testParseProjectedKubernetesAPICABundleError   = "parse projected Kubernetes API CA bundle"
	testMetadataIdentityError                      = "metadata identity"
	testMalformedLabelError                        = "malformed label"
	testSizeLimitError                             = "size limit"
	testMalformedError                             = "malformed"
	testNetworkFailureMessage                      = "network failure"
	testInvalidSizeError                           = "invalid size"
	testTypeIsNotApprovedError                     = "type is not approved"
	testNilRSAKeyCaseName                          = "nil RSA key"
	testNilKeyErrorSubstring                       = "nil"
	testUnsupportedValue                           = "unsupported"
	testMissingOrUnexpectedDataKeysError           = "missing or unexpected data keys"
	testWrongDNSName                               = "wrong.example"
	testExternalLoginSNI                           = "login.external.example"
	testMissingDerivedServiceOrPodFQDNError        = "omit a derived service or pod FQDN"
	testValidityPeriodError                        = "validity period"
	testNonCertificateDataError                    = "non-certificate data"
	testMissingOrUnexpectedError                   = "missing or unexpected"
	testTrailingYAMLDocumentError                  = "trailing YAML document"
	testOverlapsPKIDirectoryError                  = "overlaps the PKI directory"
	testOverlapsSensitiveSecretError               = "overlaps a trust, database, or credential Secret"
	testOtherWorkloadServiceName                   = "other-workload-i01"
	testPSIDOrPoolIdentityError                    = "psid or pool identity"
	testRunOrReleaseIdentityError                  = "run or release identity"
	testWorkloadOrPodIdentityError                 = "workload or pod identity is invalid"
	testSymlinkOrNonDirectoryError                 = "symlink or non-directory"
	testFileOperationLstat                         = "lstat"
	testFileOperationMkdir                         = "mkdir"
	testFileOperationMkdirTemp                     = "mkdir-temp"
	testFileOperationOpen                          = "open"
	testFileOperationRename                        = "rename"
	testFileOperationShortWrite                    = "short-write"
	testFileOperationSync                          = "sync"
	testInitialContextCancellationCaseName         = "initial context cancellation"
	testWriteMaterializerOutputError               = "write materializer output"
	testTargetCreatedDuringPublishRejectedCaseName = "target created during publish is rejected"
	testNotUniquePerPodError                       = "not unique per pod"
	testMaterializerTestLeafCommonName             = "materializer test leaf"
	testMismatchedIdentityValue                    = "other"
	testLabelValue                                 = "value"
	testMalformedECDSAKeyCaseName                  = "malformed ECDSA key"
	testMalformedEd25519CaseName                   = "malformed Ed25519"
	testUnsupportedKeyTypeCaseName                 = "unsupported key type"
)

func fetchSecretAtTestEndpoint(
	ctx context.Context,
	doer HTTPDoer,
	token string,
	namespace string,
	name string,
	config InstanceConfig,
	service InstanceService,
	role string,
) (map[string][]byte, error) {
	return fetchSecret(ctx, doer, token, testAPIEndpoint, namespace, name, config, service, role)
}

type testCertificate struct {
	certPEM []byte
	keyPEM  []byte
}

type testPKIFixture struct {
	now        time.Time
	config     InstanceConfig
	service    InstanceService
	trust      TrustInputs
	tlsBundle  map[string][]byte
	health     map[string][]byte
	otlp       map[string][]byte
	rootPEM    []byte
	rootKeyPEM []byte
	apiCAPEM   []byte
}

func newTestPKIFixture(t *testing.T) *testPKIFixture {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	root, rootPEM, rootKeyPEM := newTestRoot(t, now)

	config, err := DecodeInstanceConfig([]byte(validInstanceYAML))
	if err != nil {
		t.Fatal(err)
	}

	service, err := SelectPodService(config, testPostgreSQLPodNameOne)
	if err != nil {
		t.Fatal(err)
	}

	adminDNS, err := expectedAdminServerDNS(config, service, testNamespace)
	if err != nil {
		t.Fatal(err)
	}

	workloadDNS := service.Name + "." + testNamespace + ".svc." + config.ClusterDomain
	publicDNSNames := testExpectedPublicDNSNames(config, service, testNamespace)
	public := newTestLeaf(t, root, rootPEM, now, x509.ExtKeyUsageServerAuth, publicDNSNames, nil, false, rootKeyPEM)
	admin := newTestLeaf(t, root, rootPEM, now, x509.ExtKeyUsageServerAuth, []string{adminDNS}, []string{cryptoutilSharedMagic.IPv4Loopback}, false, rootKeyPEM)
	health := newTestLeaf(t, root, rootPEM, now, x509.ExtKeyUsageClientAuth, []string{workloadDNS}, nil, false, rootKeyPEM)
	otlp := newTestLeaf(t, root, rootPEM, now, x509.ExtKeyUsageClientAuth, []string{workloadDNS}, nil, false, rootKeyPEM)
	postgres := newTestLeaf(t, root, rootPEM, now, x509.ExtKeyUsageClientAuth, []string{workloadDNS}, nil, false, rootKeyPEM)

	postgresBindings := make(map[string]PostgreSQLClientBinding, len(config.InstanceServices))
	for _, configuredService := range config.InstanceServices {
		commonName := "client-" + configuredService.PodName
		if configuredService.PodName == service.PodName {
			commonName = testMaterializerTestLeafCommonName
		}

		postgresBindings[configuredService.PodName] = PostgreSQLClientBinding{
			CommonName: commonName, DatabaseRole: testIdentityAuthzClientDatabaseRole,
		}
	}

	return &testPKIFixture{
		now:     now,
		config:  config,
		service: service,
		trust: TrustInputs{
			PublicServerPEM: rootPEM, AdminServerPEM: rootPEM,
			HealthClientPEM: rootPEM, OTLPClientPEM: rootPEM,
			PostgreSQLClientPEM:      rootPEM,
			PostgreSQLClientBindings: postgresBindings,
		},
		tlsBundle: map[string][]byte{
			cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate: public.certPEM, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey: public.keyPEM,
			cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminCertificate: admin.certPEM, cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminPrivateKey: admin.keyPEM,
			cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLCertificate: postgres.certPEM, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLPrivateKey: postgres.keyPEM,
		},
		health:     map[string][]byte{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate: health.certPEM, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey: health.keyPEM},
		otlp:       map[string][]byte{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate: otlp.certPEM, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey: otlp.keyPEM},
		rootPEM:    rootPEM,
		rootKeyPEM: rootKeyPEM,
		apiCAPEM:   rootPEM,
	}
}

func newTestRoot(t *testing.T, now time.Time) (*x509.Certificate, []byte, []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	serial, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}
	template.Subject.CommonName = "materializer test root"

	der, err := x509.CreateCertificate(crand.Reader, template, template, key.Public(), key)
	if err != nil {
		t.Fatal(err)
	}

	root, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	return root, pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: keyDER})
}

func newTestLeaf(
	t *testing.T,
	root *x509.Certificate,
	rootPEM []byte,
	now time.Time,
	usage x509.ExtKeyUsage,
	dnsNames []string,
	ipStrings []string,
	isCA bool,
	rootKeyPEM []byte,
) testCertificate {
	return newTestLeafWithKeyUsage(
		t, root, rootPEM, now, usage, dnsNames, ipStrings, isCA, rootKeyPEM, x509.KeyUsageDigitalSignature,
	)
}

func newTestLeafWithKeyUsage(
	t *testing.T,
	root *x509.Certificate,
	rootPEM []byte,
	now time.Time,
	usage x509.ExtKeyUsage,
	dnsNames []string,
	ipStrings []string,
	isCA bool,
	rootKeyPEM []byte,
	keyUsage x509.KeyUsage,
) testCertificate {
	t.Helper()

	rootKey, err := parsePrivateKeyPEM(rootKeyPEM)
	if err != nil {
		t.Fatal(err)
	}

	signer, ok := rootKey.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatal("test root key is not ECDSA")
	}

	key, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	serial, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	if err != nil {
		t.Fatal(err)
	}

	ips := make([]net.IP, 0, len(ipStrings))
	for _, address := range ipStrings {
		ips = append(ips, net.ParseIP(address))
	}

	template := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{usage},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
		DNSNames:              dnsNames,
		IPAddresses:           ips,
	}

	template.Subject.CommonName = "materializer test leaf"
	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	der, err := x509.CreateCertificate(crand.Reader, template, root, key.Public(), signer)
	if err != nil {
		t.Fatal(err)
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: der})
	certPEM = append(certPEM, rootPEM...)

	return testCertificate{
		certPEM: certPEM,
		keyPEM:  pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: keyDER}),
	}
}

func cloneTestPKIFixture(source *testPKIFixture) *testPKIFixture {
	clone := *source

	clone.trust.PostgreSQLClientBindings = make(map[string]PostgreSQLClientBinding, len(source.trust.PostgreSQLClientBindings))
	for podName, binding := range source.trust.PostgreSQLClientBindings {
		clone.trust.PostgreSQLClientBindings[podName] = binding
	}

	clone.tlsBundle = cloneByteMap(source.tlsBundle)
	clone.health = cloneByteMap(source.health)
	clone.otlp = cloneByteMap(source.otlp)

	return &clone
}

func testExpectedPublicDNSNames(config InstanceConfig, service InstanceService, namespace string) []string {
	return []string{
		service.Name + "." + namespace + ".svc." + config.ClusterDomain,
		service.PodName + "." + config.PSID + "-" + config.Pool + "-pods." + namespace + ".svc." + config.ClusterDomain,
	}
}

func testExpectedAdministrativeDNS(t *testing.T, config InstanceConfig, service InstanceService) string {
	t.Helper()

	expected, err := expectedAdminServerDNS(config, service, testNamespace)
	if err != nil {
		t.Fatal(err)
	}

	return expected
}

func testInstanceConfigForPool(t *testing.T, pool string) InstanceConfig {
	t.Helper()

	instanceYAML := strings.Replace(validInstanceYAML, "pool: pg\n", "pool: "+pool+"\n", 1)
	instanceYAML = strings.ReplaceAll(instanceYAML, "identity-authz-pg", "identity-authz-"+pool)

	config, err := DecodeInstanceConfig([]byte(instanceYAML))
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func setTestPublicLeafDNS(t *testing.T, fixture *testPKIFixture, namespace string) {
	t.Helper()

	root, err := parseCertificatePEM(fixture.rootPEM)
	if err != nil {
		t.Fatal(err)
	}

	names := testExpectedPublicDNSNames(fixture.config, fixture.service, namespace)
	leaf := newTestLeaf(
		t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageServerAuth, names, nil, false, fixture.rootKeyPEM,
	)
	fixture.tlsBundle[SecretDataPublicCertificate] = leaf.certPEM
	fixture.tlsBundle[SecretDataPublicPrivateKey] = leaf.keyPEM
}

func setTestAdminLeafDNS(t *testing.T, fixture *testPKIFixture, namespace string) {
	t.Helper()

	root, err := parseCertificatePEM(fixture.rootPEM)
	if err != nil {
		t.Fatal(err)
	}

	adminDNS, err := expectedAdminServerDNS(fixture.config, fixture.service, namespace)
	if err != nil {
		t.Fatal(err)
	}

	leaf := newTestLeaf(
		t, root[0], fixture.rootPEM, fixture.now, x509.ExtKeyUsageServerAuth,
		[]string{adminDNS}, []string{cryptoutilSharedMagic.IPv4Loopback}, false, fixture.rootKeyPEM,
	)
	fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminCertificate] = leaf.certPEM
	fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminPrivateKey] = leaf.keyPEM
}

func cloneByteMap(source map[string][]byte) map[string][]byte {
	clone := make(map[string][]byte, len(source))
	for key, value := range source {
		clone[key] = append([]byte(nil), value...)
	}

	return clone
}

func removeDataKey(data map[string][]byte, key string) {
	delete(data, key)
}
