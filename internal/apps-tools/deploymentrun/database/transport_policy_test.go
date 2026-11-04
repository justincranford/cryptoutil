// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestPostgresRejectsAmbientConfigurationBeforeParsing(t *testing.T) {
	t.Parallel()

	for _, key := range strings.Fields(postgresEnvironmentKeys) {
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			config, source, err := postgresConfigurationWithEnvironment("", nil, nil, nil, nil, nil, func(candidate string) string {
				if candidate == key {
					return "DO-NOT-EXPOSE-ENVIRONMENT"
				}

				return ""
			})
			require.ErrorContains(t, err, "ambient PostgreSQL")
			require.NotContains(t, err.Error(), "DO-NOT-EXPOSE-ENVIRONMENT")
			require.Nil(t, config)
			require.Nil(t, source)
		})
	}

	require.NoError(t, explicitPostgresEnvironment(func(string) string { return "" }))
}

func editedTransportCertificate(t *testing.T, original, key []byte, edit func(*x509.Certificate)) []byte {
	t.Helper()

	block, _ := pem.Decode(original)
	template, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	privateBlock, _ := pem.Decode(key)
	private, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	require.NoError(t, err)

	signer, ok := private.(*ecdsa.PrivateKey)
	require.True(t, ok)
	edit(template)
	// Parsed Extensions are output, not signing-template overrides.
	template.ExtraExtensions = nil
	raw, err := x509.CreateCertificate(crand.Reader, template, template, &signer.PublicKey, signer)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: testCertificatePEMType, Bytes: raw})
}

func TestPostgresTrustAndEncodingRejectIncompleteMaterial(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		edit func(*testing.T, []byte, []byte, []byte) ([]byte, []byte, []byte)
	}{
		{"leaf trust anchor", func(_ *testing.T, _, cert, key []byte) ([]byte, []byte, []byte) { return cert, cert, key }},
		{"mixed trust", func(_ *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return append(ca, cert...), cert, key
		}},
		{"expired root", func(t *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return editedTransportCertificate(t, ca, key, func(c *x509.Certificate) { c.NotAfter = time.Now().UTC().Add(-time.Minute) }), cert, key
		}},
		{"root lacks cert signing", func(t *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return editedTransportCertificate(t, ca, key, func(c *x509.Certificate) { c.KeyUsage = x509.KeyUsageDigitalSignature }), cert, key
		}},
		{"trailing certificate junk", func(_ *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return ca, append(cert, []byte(testInvalid)...), key
		}},
		{"trailing private key junk", func(_ *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return ca, cert, append(key, []byte(testInvalid)...)
		}},
		{"multiple private keys", func(_ *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return ca, cert, append(key, key...)
		}},
		{"wrong private block", func(_ *testing.T, ca, cert, _ []byte) ([]byte, []byte, []byte) { return ca, cert, ca }},
		{"certificate dual EKU", func(t *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return ca, editedTransportCertificate(t, cert, key, func(c *x509.Certificate) {
				c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
			}), key
		}},
		{"client missing signing usage", func(t *testing.T, ca, cert, key []byte) ([]byte, []byte, []byte) {
			return ca, editedTransportCertificate(t, cert, key, func(c *x509.Certificate) { c.KeyUsage = x509.KeyUsageKeyEncipherment }), key
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ca, cert, key := transportFixture(t)
			ca, cert, key = test.edit(t, ca, cert, key)
			config, source, err := postgresConfiguration("postgres://leader.example:5432/d_sm_kms", []byte(testMigrationOwner), []byte("protected-credential"), ca, cert, key)
			require.Error(t, err)
			require.Nil(t, config)
			require.Nil(t, source)
			require.NotContains(t, err.Error(), "protected-credential")
		})
	}
}

func TestPostgresPEMGuards(t *testing.T) {
	t.Parallel()

	const (
		header      = "pem-header"
		headerValue = "ignored-header-content"
	)
	for _, data := range [][]byte{
		nil, []byte("-----BEGIN CERTIFICATE-----\ninvalid\n"),
		pem.EncodeToMemory(&pem.Block{Type: testCertificatePEMType, Bytes: []byte(testInvalid)}),
		pem.EncodeToMemory(&pem.Block{Type: testCertificatePEMType, Headers: map[string]string{header: headerValue}, Bytes: []byte(testInvalid)}),
		bytes.Repeat([]byte("x"), cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes+1),
	} {
		_, err := postgresCertificates(data)
		require.Error(t, err)
	}

	for _, data := range [][]byte{
		nil, []byte("-----BEGIN PRIVATE KEY-----"), bytes.Repeat([]byte("x"), cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes+1),
		pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Headers: map[string]string{header: headerValue}, Bytes: []byte(testInvalid)}),
	} {
		require.False(t, singlePostgresPrivateKey(data))
	}

	for _, kind := range []string{cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, cryptoutilSharedMagic.StringPEMTypeECPrivateKey, cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey} {
		require.True(t, singlePostgresPrivateKey(pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: []byte("encoding-only")})))
	}
}

func TestPostgresClientCertificateRoleGuards(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	for _, test := range []struct {
		name string
		edit func(*x509.Certificate)
	}{
		{"CA leaf", func(c *x509.Certificate) { c.IsCA = true }},
		{"unsigned usage", func(c *x509.Certificate) { c.KeyUsage = 0 }},
		{"wrong EKU", func(c *x509.Certificate) { c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth} }},
		{"unknown EKU", func(c *x509.Certificate) { c.UnknownExtKeyUsage = []asn1.ObjectIdentifier{{1, 2, 999}} }},
		{"future leaf", func(c *x509.Certificate) { c.NotBefore = now.Add(time.Hour) }},
		{"expired leaf", func(c *x509.Certificate) { c.NotAfter = now }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cert := &x509.Certificate{NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
			test.edit(cert)
			require.False(t, postgresClientCertificate(cert, now))
		})
	}

	require.False(t, postgresClientCertificate(nil, now))
}

func TestPostgresApprovedTransportAlgorithms(t *testing.T) {
	t.Parallel()

	n := new(big.Int).Lsh(big.NewInt(1), 2047)
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	for _, test := range []struct {
		name  string
		key   any
		valid bool
	}{
		{cryptoutilSharedMagic.KeyTypeRSA, &rsa.PublicKey{N: n, E: 65537}, true},
		{"nil RSA", (*rsa.PublicKey)(nil), false},
		{"empty RSA", &rsa.PublicKey{}, false},
		{"weak RSA", &rsa.PublicKey{N: big.NewInt(7), E: 3}, false},
		{"even exponent", &rsa.PublicKey{N: n, E: 4}, false},
		{"low exponent", &rsa.PublicKey{N: n, E: 1}, false},
		{"EC", &ecKey.PublicKey, true},
		{"nil EC", (*ecdsa.PublicKey)(nil), false},
		{"empty EC", &ecdsa.PublicKey{}, false},
		{"unapproved curve", &ecdsa.PublicKey{Curve: elliptic.P224()}, false},
		{"malformed point", &ecdsa.PublicKey{Curve: elliptic.P256()}, false},
		{cryptoutilSharedMagic.EdCurveEd25519, ed25519.PublicKey(make([]byte, ed25519.PublicKeySize)), true},
		{"invalid Ed25519", ed25519.PublicKey{}, false},
		{"unknown key", pkix.Name{}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.valid, approvedTransportCertificate(&x509.Certificate{PublicKey: test.key, SignatureAlgorithm: x509.ECDSAWithSHA256}))
		})
	}

	require.False(t, approvedTransportCertificate(nil))
	require.False(t, approvedTransportCertificate(&x509.Certificate{PublicKey: &rsa.PublicKey{N: n, E: 65537}, SignatureAlgorithm: x509.SHA1WithRSA}))
}

func TestDatabaseContextEntryGuards(t *testing.T) {
	t.Parallel()

	var absent context.Context
	require.ErrorContains(t, Apply(absent, sharedSQLite, psidKMS, SQLite), "requires context and database")
	require.ErrorContains(t, Apply(context.Background(), nil, psidKMS, SQLite), "requires context and database")
	require.ErrorContains(t, Execute(absent, Migrate, Input{}, nil), "context required")
}
