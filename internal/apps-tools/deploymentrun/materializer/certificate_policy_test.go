// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCertificateAndKeyPEMInputBounds(t *testing.T) {
	t.Parallel()

	oversized := bytes.Repeat([]byte("x"), cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes+1)
	_, certificateErr := parseCertificatePEM(oversized)
	require.ErrorContains(t, certificateErr, "certificate PEM exceeds the size limit")

	_, privateKeyErr := parsePrivateKeyPEM(oversized)
	require.ErrorContains(t, privateKeyErr, "private key PEM exceeds the size limit")
}

func TestApprovedPublicKeyPolicy(t *testing.T) {
	t.Parallel()

	p256 := elliptic.P256()
	p384 := elliptic.P384()
	p521 := elliptic.P521()
	p224 := elliptic.P224()

	tests := []struct {
		name    string
		key     any
		wantErr string
	}{
		{name: testRSA2048Name, key: testRSAPublicKey(2048, 65537)},
		{name: "RSA below minimum", key: testRSAPublicKey(1024, 65537), wantErr: testApprovedKeySizeError},
		{name: "RSA invalid exponent", key: testRSAPublicKey(2048, 2), wantErr: testApprovedKeySizeError},
		{name: "ECDSA P-256", key: testECDSAPublicKey(t, p256)},
		{name: testECDSAP384Name, key: testECDSAPublicKey(t, p384)},
		{name: "ECDSA P-521", key: testECDSAPublicKey(t, p521)},
		{name: testECDSAP224RejectedName, key: testECDSAPublicKey(t, p224), wantErr: testCurveNotApprovedError},
		{name: testMalformedECDSAKeyCaseName, key: &ecdsa.PublicKey{Curve: p256}, wantErr: testMalformedError},
		{name: cryptoutilSharedMagic.EdCurveEd25519, key: make(ed25519.PublicKey, ed25519.PublicKeySize)},
		{name: testMalformedEd25519CaseName, key: ed25519.PublicKey{1}, wantErr: testInvalidSizeError},
		{name: testUnsupportedKeyTypeCaseName, key: "not a public key", wantErr: testTypeIsNotApprovedError},
		{name: testNilRSAKeyCaseName, key: (*rsa.PublicKey)(nil), wantErr: testApprovedKeySizeError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateApprovedPublicKey(testCase.key)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestApprovedPrivateKeyPolicy(t *testing.T) {
	t.Parallel()

	validCurve := elliptic.P384()
	invalidCurve := elliptic.P224()

	tests := []struct {
		name    string
		key     any
		wantErr string
	}{
		{
			name: testECDSAP384Name,
			key:  testECDSAPrivateKey(t, validCurve),
		},
		{
			name: testRSA2048Name,
			key:  &rsa.PrivateKey{PublicKey: *testRSAPublicKey(2048, 65537)},
		},
		{
			name:    testECDSAP224RejectedName,
			key:     testECDSAPrivateKey(t, invalidCurve),
			wantErr: testCurveNotApprovedError,
		},
		{
			name:    "RSA 1024 rejected",
			key:     &rsa.PrivateKey{PublicKey: *testRSAPublicKey(1024, 65537)},
			wantErr: testApprovedKeySizeError,
		},
		{name: testNilRSAKeyCaseName, key: (*rsa.PrivateKey)(nil), wantErr: testNilKeyErrorSubstring},
		{name: cryptoutilSharedMagic.EdCurveEd25519, key: make(ed25519.PrivateKey, ed25519.PrivateKeySize)},
		{name: testMalformedEd25519CaseName, key: ed25519.PrivateKey{1}, wantErr: testInvalidSizeError},
		{name: testUnsupportedKeyTypeCaseName, key: struct{}{}, wantErr: testTypeIsNotApprovedError},
		{name: "nil ECDSA key", key: (*ecdsa.PrivateKey)(nil), wantErr: testNilKeyErrorSubstring},
		{
			name:    testMalformedECDSAKeyCaseName,
			key:     &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: validCurve}},
			wantErr: testMalformedError,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateApprovedPrivateKey(testCase.key)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestApprovedCertificatePolicy(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	certificates, err := parseCertificatePEM(fixture.tlsBundle[SecretDataPublicCertificate])
	require.NoError(t, err)

	unsupportedKey := *certificates[0]
	unsupportedKey.PublicKey = testUnsupportedValue
	weakSignature := *certificates[0]
	weakSignature.SignatureAlgorithm = x509.ECDSAWithSHA1

	tests := []struct {
		name        string
		certificate *x509.Certificate
		wantErr     string
	}{
		{name: "approved ECDSA certificate", certificate: certificates[0]},
		{name: "nil certificate", wantErr: "certificate is nil"},
		{name: "unapproved public key", certificate: &unsupportedKey, wantErr: testTypeIsNotApprovedError},
		{name: "unapproved signature", certificate: &weakSignature, wantErr: "signature algorithm is not approved"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			policyErr := validateApprovedCertificate(testCase.certificate)
			if testCase.wantErr != "" {
				require.ErrorContains(t, policyErr, testCase.wantErr)

				return
			}

			require.NoError(t, policyErr)
		})
	}
}

func TestCertificateMaterialRejectsNonapprovedPrivateKey(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	curve := elliptic.P224()
	privateKey := testECDSAPrivateKey(t, curve)
	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	leaf, err := parseCertificatePEM(fixture.tlsBundle[SecretDataPublicCertificate])
	require.NoError(t, err)

	materialErr := validateCertificateMaterial(certificateMaterial{
		certPEM: fixture.tlsBundle[SecretDataPublicCertificate],
		keyPEM: pem.EncodeToMemory(&pem.Block{
			Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: keyDER,
		}),
		trustPEM: fixture.rootPEM, expectedDNS: leaf[0].DNSNames[0],
		usage: x509.ExtKeyUsageServerAuth, certOutput: PublicServerCertificateFile,
		keyOutput: PublicServerPrivateKeyFile,
	}, fixture.now)
	require.ErrorContains(t, materialErr, "private key for public-server/tls.key is not FIPS-approved")
	require.ErrorContains(t, materialErr, testCurveNotApprovedError)
}

func TestTrustAndCertificateChainsRejectNonapprovedCAKeys(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	unapprovedCAPEM := newTestUnapprovedCAPEM(t, fixture)
	_, trustErr := parseTrustRoots(unapprovedCAPEM)
	require.ErrorContains(t, trustErr, "trust anchor is not FIPS-approved")

	leafChain := append(append([]byte(nil), fixture.tlsBundle[SecretDataPublicCertificate]...), unapprovedCAPEM...)
	leaf, err := parseCertificatePEM(fixture.tlsBundle[SecretDataPublicCertificate])
	require.NoError(t, err)

	materialErr := validateCertificateMaterial(certificateMaterial{
		certPEM: leafChain, keyPEM: fixture.tlsBundle[SecretDataPublicPrivateKey],
		trustPEM: fixture.rootPEM, expectedDNS: leaf[0].DNSNames[0],
		usage: x509.ExtKeyUsageServerAuth, certOutput: PublicServerCertificateFile,
		keyOutput: PublicServerPrivateKeyFile,
	}, fixture.now)
	require.ErrorContains(t, materialErr, "certificate chain for public-server/tls.crt is not FIPS-approved")
	require.ErrorContains(t, materialErr, testCurveNotApprovedError)
}

func TestApprovedCertificateSignaturePolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm x509.SignatureAlgorithm
		wantErr   bool
	}{
		{name: "RSA SHA-256", algorithm: x509.SHA256WithRSA},
		{name: "RSA SHA-384", algorithm: x509.SHA384WithRSA},
		{name: "RSA SHA-512", algorithm: x509.SHA512WithRSA},
		{name: "RSA PSS SHA-256", algorithm: x509.SHA256WithRSAPSS},
		{name: "RSA PSS SHA-384", algorithm: x509.SHA384WithRSAPSS},
		{name: "RSA PSS SHA-512", algorithm: x509.SHA512WithRSAPSS},
		{name: "ECDSA SHA-256", algorithm: x509.ECDSAWithSHA256},
		{name: "ECDSA SHA-384", algorithm: x509.ECDSAWithSHA384},
		{name: "ECDSA SHA-512", algorithm: x509.ECDSAWithSHA512},
		{name: cryptoutilSharedMagic.EdCurveEd25519, algorithm: x509.PureEd25519},
		{name: "RSA SHA-1 rejected", algorithm: x509.SHA1WithRSA, wantErr: true},
		{name: "ECDSA SHA-1 rejected", algorithm: x509.ECDSAWithSHA1, wantErr: true},
		{name: "RSA MD5 rejected", algorithm: x509.MD5WithRSA, wantErr: true},
		{name: "DSA rejected", algorithm: x509.DSAWithSHA1, wantErr: true},
		{name: "unknown rejected", algorithm: x509.UnknownSignatureAlgorithm, wantErr: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateApprovedSignatureAlgorithm(testCase.algorithm)
			if testCase.wantErr {
				require.ErrorContains(t, err, "signature algorithm is not approved")

				return
			}

			require.NoError(t, err)
		})
	}
}

func newTestUnapprovedCAPEM(t *testing.T, fixture *testPKIFixture) []byte {
	t.Helper()

	issuer, err := parseCertificatePEM(fixture.rootPEM)
	require.NoError(t, err)
	issuerKeyValue, err := parsePrivateKeyPEM(fixture.rootKeyPEM)
	require.NoError(t, err)

	issuerKey, ok := issuerKeyValue.(*ecdsa.PrivateKey)
	require.True(t, ok)

	curve := elliptic.P224()
	unapprovedKey := testECDSAPrivateKey(t, curve)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(777),
		NotBefore:             fixture.now.Add(-time.Hour),
		NotAfter:              fixture.now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}
	template.Subject.CommonName = "unapproved test CA"
	der, err := x509.CreateCertificate(crand.Reader, template, issuer[0], &unapprovedKey.PublicKey, issuerKey)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: der})
}

func testRSAPublicKey(bits int, exponent int) *rsa.PublicKey {
	modulus := new(big.Int).Lsh(big.NewInt(1), uint(bits-1))

	return &rsa.PublicKey{N: modulus, E: exponent}
}

func testECDSAPublicKey(t *testing.T, curve elliptic.Curve) *ecdsa.PublicKey {
	t.Helper()

	privateKey := testECDSAPrivateKey(t, curve)

	return &privateKey.PublicKey
}

func testECDSAPrivateKey(t *testing.T, curve elliptic.Curve) *ecdsa.PrivateKey {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(curve, crand.Reader)
	require.NoError(t, err)

	return privateKey
}
