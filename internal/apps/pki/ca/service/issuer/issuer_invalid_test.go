// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"crypto/x509"
	"testing"
	"time"

	cryptoutilCABootstrap "cryptoutil/internal/apps/pki/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAIntermediate "cryptoutil/internal/apps/pki/ca/intermediate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"

	"github.com/stretchr/testify/require"
)

func TestIssuer_Issue_InvalidRequest(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     *CertificateRequest
		wantErr string
	}{
		{
			name:    "nil-request",
			req:     nil,
			wantErr: "request cannot be nil",
		},
		{
			name: "no-public-key",
			req: &CertificateRequest{
				SubjectRequest: &cryptoutilCAProfileSubject.Request{
					CommonName: "test",
				},
				ValidityDuration: 24 * time.Hour,
			},
			wantErr: "public key is required",
		},
		{
			name: "zero-validity",
			req: &CertificateRequest{
				SubjectRequest: &cryptoutilCAProfileSubject.Request{
					CommonName: "test",
				},
				PublicKey:        keyPair.PublicKey,
				ValidityDuration: 0,
			},
			wantErr: "validity duration must be positive",
		},
		{
			name: "no-subject-request",
			req: &CertificateRequest{
				PublicKey:        keyPair.PublicKey,
				ValidityDuration: 24 * time.Hour,
			},
			wantErr: "subject request is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			issued, audit, err := issuer.Issue(tc.req)
			require.Error(t, err)
			require.Nil(t, issued)
			require.Nil(t, audit)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestIssuer_Issue_ValidityTruncation(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create short-lived root.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Short Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  180 * 24 * time.Hour, // 180 days.
		PathLenConstraint: 1,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	// Create issuing CA with short validity.
	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "Short Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  90 * 24 * time.Hour, // 90 days.
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	issuingCA, _, err := provisioner.Provision(intermediateConfig)
	require.NoError(t, err)

	caConfig := &IssuingCAConfig{
		Name:        "Short Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "long-validity-request",
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 365 * 24 * time.Hour, // Request 1 year.
	}

	issued, _, err := issuer.Issue(req)
	require.NoError(t, err)
	require.NotNil(t, issued)

	// Certificate should not outlive issuing CA.
	require.True(t, !issued.Certificate.NotAfter.After(issuingCA.Certificate.NotAfter))
}

func TestIssuer_Issue_ChainVerification(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create full chain: root -> intermediate -> end-entity.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Chain Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  20 * 365 * 24 * time.Hour,
		PathLenConstraint: 2,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "Chain Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	issuingCA, _, err := provisioner.Provision(intermediateConfig)
	require.NoError(t, err)

	caConfig := &IssuingCAConfig{
		Name:        "Chain Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "chain-test.example.com",
			DNSNames:   []string{"chain-test.example.com"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 90 * 24 * time.Hour,
	}

	issued, _, err := issuer.Issue(req)
	require.NoError(t, err)

	// Verify full chain.
	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCA.Certificate)

	intermediatePool := x509.NewCertPool()
	intermediatePool.AddCert(issuingCA.Certificate)

	opts := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediatePool,
		CurrentTime:   time.Now().UTC(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	chains, err := issued.Certificate.Verify(opts)
	require.NoError(t, err)
	require.NotEmpty(t, chains)

	// Chain should be: end-entity -> issuing CA -> root CA.
	require.Len(t, chains[0], 3)
}
