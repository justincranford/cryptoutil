// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	rsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	postgresEnvironmentKeys = "PGHOST PGPORT PGDATABASE PGUSER PGPASSWORD PGPASSFILE PGSERVICE PGSERVICEFILE PGSSLMODE PGSSLCERT PGSSLKEY PGSSLROOTCERT PGSSLPASSWORD PGOPTIONS PGAPPNAME PGCONNECT_TIMEOUT PGTARGETSESSIONATTRS PGTZ PGMINPROTOCOLVERSION PGMAXPROTOCOLVERSION"
	rejectedCertificatePEM  = "certificate PEM rejected"
)

// pgx v5.11 parses ambient service files before callers can replace its config.
// Reject recognized ambient configuration before invoking that constructor.
func explicitPostgresEnvironment(lookup func(string) string) error {
	for _, key := range strings.Fields(postgresEnvironmentKeys) {
		if lookup(key) != "" {
			return errors.New("ambient PostgreSQL configuration is forbidden; use protected files")
		}
	}

	return nil
}

func postgresCertificates(data []byte) ([]*x509.Certificate, error) {
	if len(data) > cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes {
		return nil, errors.New(rejectedCertificatePEM)
	}

	remaining := bytes.TrimSpace(data)
	certificates := []*x509.Certificate{}

	for len(remaining) > 0 {
		if !bytes.HasPrefix(remaining, []byte("-----BEGIN CERTIFICATE-----")) {
			return nil, errors.New(rejectedCertificatePEM)
		}

		block, rest := pem.Decode(remaining)
		if block == nil || block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate || len(block.Headers) != 0 {
			return nil, errors.New(rejectedCertificatePEM)
		}

		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil || !approvedTransportCertificate(certificate) {
			return nil, errors.New("certificate policy rejected")
		}

		certificates = append(certificates, certificate)
		remaining = bytes.TrimSpace(rest)
	}

	if len(certificates) == 0 {
		return nil, errors.New("certificate PEM required")
	}

	return certificates, nil
}

// Match the existing materializer approved public-key/signature policy.
// These are input checks, not a new algorithm selection or a FIPS attestation.
func approvedTransportCertificate(certificate *x509.Certificate) (approved bool) {
	defer func() {
		if recover() != nil {
			approved = false
		}
	}()

	if certificate == nil {
		return false
	}

	switch key := certificate.PublicKey.(type) {
	case *rsa.PublicKey:
		if key == nil || key.N == nil || key.N.BitLen() < cryptoutilSharedMagic.RSAKeySize2048 || key.E < 3 || key.E%2 == 0 {
			return false
		}
	case *ecdsa.PublicKey:
		if key == nil || key.Curve == nil {
			return false
		}

		if key.Curve != elliptic.P256() && key.Curve != elliptic.P384() && key.Curve != elliptic.P521() {
			return false
		}

		if _, err := key.Bytes(); err != nil {
			return false
		}
	case ed25519.PublicKey:
		if len(key) != ed25519.PublicKeySize {
			return false
		}
	default:
		return false
	}

	switch certificate.SignatureAlgorithm {
	case x509.SHA256WithRSA, x509.SHA384WithRSA, x509.SHA512WithRSA, x509.SHA256WithRSAPSS, x509.SHA384WithRSAPSS, x509.SHA512WithRSAPSS,
		x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512, x509.PureEd25519:
		return true
	default:
		return false
	}
}

func postgresTrustRoots(data []byte, now time.Time) (*x509.CertPool, error) {
	certificates, err := postgresCertificates(data)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()

	for _, certificate := range certificates {
		if !certificate.IsCA || !certificate.BasicConstraintsValid || certificate.KeyUsage&x509.KeyUsageCertSign == 0 ||
			now.Before(certificate.NotBefore) || !now.Before(certificate.NotAfter) {
			return nil, errors.New("PostgreSQL trust anchor policy rejected")
		}

		roots.AddCert(certificate)
	}

	return roots, nil
}

func postgresClientCertificate(certificate *x509.Certificate, now time.Time) bool {
	return certificate != nil && !certificate.IsCA && certificate.KeyUsage&x509.KeyUsageDigitalSignature != 0 &&
		len(certificate.ExtKeyUsage) == 1 && certificate.ExtKeyUsage[0] == x509.ExtKeyUsageClientAuth && len(certificate.UnknownExtKeyUsage) == 0 &&
		!now.Before(certificate.NotBefore) && now.Before(certificate.NotAfter)
}

func singlePostgresPrivateKey(data []byte) bool {
	if len(data) > cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes {
		return false
	}

	remaining := bytes.TrimSpace(data)
	if !bytes.HasPrefix(remaining, []byte("-----BEGIN ")) {
		return false
	}

	block, rest := pem.Decode(remaining)
	if block == nil || len(block.Headers) != 0 || len(bytes.TrimSpace(rest)) != 0 {
		return false
	}

	switch block.Type {
	case cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, cryptoutilSharedMagic.StringPEMTypeECPrivateKey, cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey:
		return true
	default:
		return false
	}
}
