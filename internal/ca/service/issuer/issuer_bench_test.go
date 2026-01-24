// Copyright (c) 2025 Justin Cranford

package issuer

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/ca/crypto"
	cryptoutilCAProfileSubject "cryptoutil/internal/ca/profile/subject"

	"github.com/stretchr/testify/require"
)

// BenchmarkCertificateIssuance_ECDSA measures end-entity certificate issuance with ECDSA keys.
func BenchmarkCertificateIssuance_ECDSA(b *testing.B) {
	// Setup issuing CA.
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(b, err)

	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	caCertDER, err := x509.CreateCertificate(crand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	require.NoError(b, err)

	caCert, err = x509.ParseCertificate(caCertDER)
	require.NoError(b, err)

	// Create minimal issuer without profiles (simplifies benchmark).
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	issuer, err := NewIssuer(provider, &IssuingCAConfig{
		Name:        "test-issuer",
		Certificate: caCert,
		PrivateKey:  caKey,
	})
	require.NoError(b, err)

	// Generate end-entity key once.
	eeKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(b, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
			DNSNames:     []string{"test.example.com", "www.test.example.com"},
		},
		PublicKey:        eeKey.Public(),
		ValidityDuration: 90 * 24 * time.Hour,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := issuer.Issue(req)
		if err != nil {
			b.Fatalf("failed to issue certificate: %v", err)
		}
	}
}

// BenchmarkCertificateIssuance_RSA measures end-entity certificate issuance with RSA keys.
func BenchmarkCertificateIssuance_RSA(b *testing.B) {
	// Setup issuing CA with RSA key.
	caKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(b, err)

	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA RSA",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	caCertDER, err := x509.CreateCertificate(crand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	require.NoError(b, err)

	caCert, err = x509.ParseCertificate(caCertDER)
	require.NoError(b, err)

	// Create minimal issuer without profiles.
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	issuer, err := NewIssuer(provider, &IssuingCAConfig{
		Name:        "test-issuer-rsa",
		Certificate: caCert,
		PrivateKey:  caKey,
	})
	require.NoError(b, err)

	// Generate end-entity RSA key once.
	eeKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(b, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName:   "rsa-test.example.com",
			Organization: []string{"Test Org"},
			DNSNames:     []string{"rsa-test.example.com", "www.rsa-test.example.com"},
		},
		PublicKey:        eeKey.Public(),
		ValidityDuration: 90 * 24 * time.Hour,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := issuer.Issue(req)
		if err != nil {
			b.Fatalf("failed to issue certificate: %v", err)
		}
	}
}

// BenchmarkCertificateIssuance_Parallel measures concurrent certificate issuance.
func BenchmarkCertificateIssuance_Parallel(b *testing.B) {
	// Setup issuing CA.
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(b, err)

	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA Parallel",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	caCertDER, err := x509.CreateCertificate(crand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	require.NoError(b, err)

	caCert, err = x509.ParseCertificate(caCertDER)
	require.NoError(b, err)

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	issuer, err := NewIssuer(provider, &IssuingCAConfig{
		Name:        "test-issuer-parallel",
		Certificate: caCert,
		PrivateKey:  caKey,
	})
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine generates its own key.
		eeKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(b, err)

		req := &CertificateRequest{
			SubjectRequest: &cryptoutilCAProfileSubject.Request{
				CommonName:   "parallel-test.example.com",
				Organization: []string{"Test Org"},
				DNSNames:     []string{"parallel-test.example.com"},
			},
			PublicKey:        eeKey.Public(),
			ValidityDuration: 90 * 24 * time.Hour,
		}

		for pb.Next() {
			_, _, err := issuer.Issue(req)
			if err != nil {
				b.Fatalf("failed to issue certificate: %v", err)
			}
		}
	})
}
