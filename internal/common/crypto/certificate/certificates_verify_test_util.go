package certificate

import (
	"crypto/x509"
	"cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func verifyCertificateTemplate(t *testing.T, err error, certTemplate *x509.Certificate) {
	require.NoError(t, err, "Failed to create certificate template")
	require.NotNil(t, certTemplate, "Certificate template should not be nil")
}

func verifyCACertificate(t *testing.T, err error, certChain []*x509.Certificate, DERChain [][]byte, PEMChain [][]byte, expectedIssuerName string, expectedSubjectName string, expectedDuration time.Duration, expectedMaxPathLen int) {
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

func verifyEndEntityCertificate(t *testing.T, err error, cert *x509.Certificate, certDER []byte, certPEM []byte, expectedIssuerName string, expectedSubjectName string, expectedDuration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL) {
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
}

func verifyCertChain(t *testing.T, certificate *x509.Certificate, roots *x509.CertPool, intermediates *x509.CertPool) {
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
	require.NoError(t, err, "Failed to create CA subjects")

	for i, subject := range caSubjects {
		// Verify subject fields are properly populated
		require.NotEmpty(t, subject.SubjectName, "CA subject name should not be empty at index %d", i)
		require.NotEmpty(t, subject.IssuerName, "CA issuer name should not be empty at index %d", i)
		require.True(t, subject.IsCA, "CA should have IsCA=true at index %d", i)

		// For root CA (index 0), subject and issuer should be the same (self-signed)
		if i == 0 {
			require.Equal(t, subject.SubjectName, subject.IssuerName, "Root CA issuer name should be self-signed at index %d", i)
		} else {
			// For intermediate CAs, issuer should be the previous CA
			expectedIssuerName := caSubjects[i-1].SubjectName
			require.Equal(t, expectedIssuerName, subject.IssuerName, "Intermediate CA should be issued by previous CA at index %d", i)
		}

		// Verify MaxPathLen follows the expected pattern
		expectedMaxPathLen := len(caSubjects) - i - 1
		require.Equal(t, expectedMaxPathLen, subject.MaxPathLen, "CA MaxPathLen should be %d at index %d", expectedMaxPathLen, i)

		derChain := make([][]byte, len(subject.KeyMaterial.CertChain))
		pemChain := make([][]byte, len(subject.KeyMaterial.CertChain))
		for j, cert := range subject.KeyMaterial.CertChain {
			derChain[j] = cert.Raw
			pemChain[j] = toPEMCertificate(cert.Raw)
		}
		verifyCACertificate(t, nil, subject.KeyMaterial.CertChain, derChain, pemChain,
			subject.IssuerName, subject.SubjectName, 10*365*cryptoutilDateTime.Days1, expectedMaxPathLen)
	}
}

func verifyEndEntitySubject(t *testing.T, err error, endEntitySubject *Subject) {
	require.NoError(t, err, "Failed to create end entity subject")

	// Verify subject fields are properly populated
	require.NotEmpty(t, endEntitySubject.SubjectName, "End entity subject name should not be empty")
	require.NotEmpty(t, endEntitySubject.IssuerName, "End entity issuer name should not be empty")
	require.False(t, endEntitySubject.IsCA, "End entity should have IsCA=false")

	// Verify that MaxPathLen is not set for end entities (should be 0)
	require.Equal(t, 0, endEntitySubject.MaxPathLen, "End entity should have MaxPathLen=0")

	endEntityCert := endEntitySubject.KeyMaterial.CertChain[0]
	verifyEndEntityCertificate(t, nil, endEntityCert, endEntityCert.Raw, toPEMCertificate(endEntityCert.Raw),
		endEntitySubject.IssuerName, endEntitySubject.SubjectName, endEntitySubject.Duration,
		endEntitySubject.DNSNames, endEntitySubject.IPAddresses,
		endEntitySubject.EmailAddresses, endEntitySubject.URIs)
}

func getKeyPairs(numCAs int, keygenPool *cryptoutilPool.ValueGenPool[*keygen.KeyPair]) ([]*keygen.KeyPair, error) {
	var keyPairs []*keygen.KeyPair
	for i := range numCAs {
		keyPair := keygenPool.Get()
		if keyPair == nil {
			return nil, fmt.Errorf("keyPair should not be nil for CA %d", i)
		}
		keyPairs = append(keyPairs, keyPair)
	}
	return keyPairs, nil
}
