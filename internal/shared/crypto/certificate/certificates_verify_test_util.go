// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"crypto/x509"
	"net"
	"net/url"
	"testing"
	"time"

	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"

	"github.com/stretchr/testify/require"
)

func verifyCertificateTemplate(t *testing.T, err error, certTemplate *x509.Certificate) {
	t.Helper()
	require.NoError(t, err, "Failed to create certificate template")
	require.NotNil(t, certTemplate, "Certificate template should not be nil")
}

func verifyCACertificate(t *testing.T, err error, certChain []*x509.Certificate, DERChain, PEMChain [][]byte, expectedIssuerName, expectedSubjectName string, expectedDuration time.Duration, expectedMaxPathLen int) {
	t.Helper()
	require.NoError(t, err, "Failed to sign certificate")
	require.NotNil(t, certChain, "Signed certificate should not be nil")
	require.NotEmpty(t, DERChain, "Certificate bytes should not be empty")
	require.NotEmpty(t, PEMChain, "Certificate PEM should not be empty")

	now := time.Now().UTC()

	require.Equal(t, expectedIssuerName, certChain[0].Issuer.CommonName, "Issuer name mismatch")
	require.Equal(t, expectedSubjectName, certChain[0].Subject.CommonName, "Subject name mismatch")
	require.True(t, certChain[0].IsCA, "Certificate should be a CA certificate")
	require.True(t, certChain[0].BasicConstraintsValid, "Basic constraints should be valid")
	require.Equal(t, expectedMaxPathLen, certChain[0].MaxPathLen, "MaxPathLen mismatch")
	require.Equal(t, expectedMaxPathLen == 0, certChain[0].MaxPathLenZero, "MaxPathLenZero mismatch")
	require.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, certChain[0].KeyUsage, "Key usage mismatch")
	require.Nil(t, certChain[0].ExtKeyUsage, "Extended key usage should be nil")
	require.True(t, certChain[0].NotBefore.Before(now), "NotBefore should be in the past")
	require.True(t, certChain[0].NotAfter.After(now), "NotAfter should be in the future")
	require.True(t, certChain[0].NotAfter.Sub(certChain[0].NotBefore) >= expectedDuration, "Certificate validity period should be >= duration")
}

func verifyEndEntityCertificate(t *testing.T, err error, cert *x509.Certificate, certDER, certPEM []byte, expectedIssuerName, expectedSubjectName string, expectedDuration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL) {
	t.Helper()
	require.NoError(t, err, "Failed to sign certificate")
	require.NotNil(t, cert, "Signed certificate should not be nil")
	require.NotEmpty(t, certDER, "Certificate bytes should not be empty")
	require.NotEmpty(t, certPEM, "Certificate PEM should not be empty")

	now := time.Now().UTC()

	require.Equal(t, expectedIssuerName, cert.Issuer.CommonName, "Issuer name mismatch")
	require.Equal(t, expectedSubjectName, cert.Subject.CommonName, "Subject name mismatch")
	require.False(t, cert.IsCA, "Certificate should not be a CA certificate")
	require.False(t, cert.BasicConstraintsValid, "Basic constraints should be invalid")
	require.Equal(t, cert.KeyUsage, x509.KeyUsageDigitalSignature, "Key usage mismatch")
	// require.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage, "Extended key usage mismatch")
	// require.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, cert.ExtKeyUsage, "Extended key usage mismatch")
	require.True(t, cert.NotBefore.Before(now), "NotBefore should be in the past")
	require.True(t, cert.NotAfter.After(now), "NotAfter should be in the future")
	require.True(t, cert.NotAfter.Sub(cert.NotBefore) >= expectedDuration, "Certificate validity period should be >= duration")
	require.ElementsMatch(t, dnsNames, cert.DNSNames, "DNS names mismatch")

	// Normalize IP addresses before comparison to handle IPv4/IPv6 representation differences
	expectedIPs := cryptoutilSharedUtilNetwork.NormalizeIPv4Addresses(ipAddresses)
	actualIPs := cryptoutilSharedUtilNetwork.NormalizeIPv4Addresses(cert.IPAddresses)
	require.ElementsMatch(t, expectedIPs, actualIPs, "IP addresses mismatch")

	require.ElementsMatch(t, emailAddresses, cert.EmailAddresses, "Email addresses mismatch")
	require.ElementsMatch(t, uris, cert.URIs, "URIs mismatch")
}

// verifyCertChain verifies a certificate chain - kept for future use
//
//nolint:unused // Test utility function for future use
func verifyCertChain(t *testing.T, certificate *x509.Certificate, roots, intermediates *x509.CertPool) {
	t.Helper()

	x509VerifyOptions := x509.VerifyOptions{
		CurrentTime:   time.Now().UTC(),
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	chains, err := certificate.Verify(x509VerifyOptions)
	require.NoError(t, err, "Failed to verify intermediate certificate using root certificate")
	require.NotEmpty(t, chains, "Certificate chains should not be empty")
}

func verifyCASubjects(t *testing.T, err error, caSubjects []*Subject) {
	t.Helper()
	require.NoError(t, err, "Failed to create CA subjects")

	for i, subject := range caSubjects {
		require.NotEmpty(t, subject.SubjectName, "CA subject name should not be empty at index %d", i)
		require.NotEmpty(t, subject.IssuerName, "CA issuer name should not be empty at index %d", i)
		require.True(t, subject.IsCA, "CA should have IsCA=true at index %d", i)

		lastIndex := len(caSubjects) - 1
		if i == lastIndex {
			require.Equal(t, subject.SubjectName, subject.IssuerName, "Root CA issuer name should be self-signed at index %d", i)
		} else {
			expectedIssuerName := caSubjects[i+1].SubjectName
			require.Equal(t, expectedIssuerName, subject.IssuerName, "Intermediate CA should be issued by next CA at index %d", i)
		}

		require.LessOrEqual(t, 0, subject.MaxPathLen, "CA MaxPathLen should be %d at index %d", i)

		derChain := make([][]byte, len(subject.KeyMaterial.CertificateChain))
		pemChain := make([][]byte, len(subject.KeyMaterial.CertificateChain))

		for j, cert := range subject.KeyMaterial.CertificateChain {
			derChain[j] = cert.Raw
			pemChain[j] = toCertificatePEM(cert.Raw)
		}

		verifyCACertificate(t, nil, subject.KeyMaterial.CertificateChain, derChain, pemChain,
			subject.IssuerName, subject.SubjectName, subject.Duration, subject.MaxPathLen)
	}
}

func verifyEndEntitySubject(t *testing.T, err error, endEntitySubject *Subject) {
	t.Helper()
	require.NoError(t, err, "Failed to create end entity subject")

	// Verify subject fields are properly populated
	require.NotEmpty(t, endEntitySubject.SubjectName, "End entity subject name should not be empty")
	require.NotEmpty(t, endEntitySubject.IssuerName, "End entity issuer name should not be empty")
	require.False(t, endEntitySubject.IsCA, "End entity should have IsCA=false")

	// Verify that MaxPathLen is not set for end entities (should be 0)
	require.Equal(t, 0, endEntitySubject.MaxPathLen, "End entity should have MaxPathLen=0")

	endEntityCert := endEntitySubject.KeyMaterial.CertificateChain[0]
	verifyEndEntityCertificate(t, nil, endEntityCert, endEntityCert.Raw, toCertificatePEM(endEntityCert.Raw),
		endEntitySubject.IssuerName, endEntitySubject.SubjectName, endEntitySubject.Duration,
		endEntitySubject.DNSNames, endEntitySubject.IPAddresses,
		endEntitySubject.EmailAddresses, endEntitySubject.URIs)
}
