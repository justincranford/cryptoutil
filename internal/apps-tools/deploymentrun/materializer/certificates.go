// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type certificateMaterial struct {
	certPEM            []byte
	keyPEM             []byte
	trustPEM           []byte
	expectedDNS        string
	expectedDNSNames   []string
	expectedCommonName string
	requireLoopback    bool
	usage              x509.ExtKeyUsage
	certOutput         string
	keyOutput          string
}

func buildMaterializedFiles(
	config InstanceConfig,
	service InstanceService,
	namespace string,
	expectedAdminDNS string,
	expectedPublicDNSNames []string,
	postgresBinding *PostgreSQLClientBinding,
	trust TrustInputs,
	tlsBundle map[string][]byte,
	healthSecret map[string][]byte,
	otlpSecret map[string][]byte,
	now time.Time,
) (map[string]materializedFile, error) {
	poolHasPostgreSQL := config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG

	bundleKeys := []string{
		SecretDataPublicCertificate, SecretDataPublicPrivateKey,
		SecretDataAdminCertificate, SecretDataAdminPrivateKey,
	}
	if poolHasPostgreSQL {
		bundleKeys = append(bundleKeys, SecretDataPostgreSQLCertificate, SecretDataPostgreSQLPrivateKey)
	}

	if err := requireExactDataKeys(tlsBundle, bundleKeys); err != nil {
		return nil, fmt.Errorf("TLS bundle Secret data is invalid: %w", err)
	}

	if err := requireExactDataKeys(healthSecret, []string{SecretDataCertificate, SecretDataPrivateKey}); err != nil {
		return nil, fmt.Errorf("health client Secret data is invalid: %w", err)
	}

	if err := requireExactDataKeys(otlpSecret, []string{SecretDataCertificate, SecretDataPrivateKey}); err != nil {
		return nil, fmt.Errorf("OTLP client Secret data is invalid: %w", err)
	}

	workloadDNS := service.Name + "." + namespace + ".svc." + config.ClusterDomain
	if _, err := validateExpectedPublicDNSNames(config, service, namespace, expectedPublicDNSNames); err != nil {
		return nil, err
	}

	if !validDNSSubdomain(expectedAdminDNS) {
		return nil, fmt.Errorf("expected administrative DNS name is invalid")
	}

	materials := []certificateMaterial{
		{
			certPEM: tlsBundle[SecretDataPublicCertificate], keyPEM: tlsBundle[SecretDataPublicPrivateKey],
			trustPEM: trust.PublicServerPEM, expectedDNS: workloadDNS, expectedDNSNames: expectedPublicDNSNames,
			usage: x509.ExtKeyUsageServerAuth, certOutput: PublicServerCertificateFile, keyOutput: PublicServerPrivateKeyFile,
		},
		{
			certPEM: tlsBundle[SecretDataAdminCertificate], keyPEM: tlsBundle[SecretDataAdminPrivateKey],
			trustPEM: trust.AdminServerPEM, expectedDNS: expectedAdminDNS, requireLoopback: true,
			usage: x509.ExtKeyUsageServerAuth, certOutput: AdminServerCertificateFile, keyOutput: AdminServerPrivateKeyFile,
		},
		{
			certPEM: healthSecret[SecretDataCertificate], keyPEM: healthSecret[SecretDataPrivateKey],
			trustPEM: trust.HealthClientPEM,
			usage:    x509.ExtKeyUsageClientAuth, certOutput: HealthClientCertificateFile, keyOutput: HealthClientPrivateKeyFile,
		},
		{
			certPEM: otlpSecret[SecretDataCertificate], keyPEM: otlpSecret[SecretDataPrivateKey],
			trustPEM: trust.OTLPClientPEM,
			usage:    x509.ExtKeyUsageClientAuth, certOutput: OTLPClientCertificateFile, keyOutput: OTLPClientPrivateKeyFile,
		},
	}

	if poolHasPostgreSQL {
		if postgresBinding == nil {
			return nil, fmt.Errorf("trusted PostgreSQL client binding is required for pg pool")
		}

		materials = append(materials, certificateMaterial{
			certPEM: tlsBundle[SecretDataPostgreSQLCertificate], keyPEM: tlsBundle[SecretDataPostgreSQLPrivateKey],
			trustPEM: trust.PostgreSQLClientPEM, expectedCommonName: postgresBinding.CommonName,
			usage: x509.ExtKeyUsageClientAuth, certOutput: PostgreSQLClientCertificateFile, keyOutput: PostgreSQLClientPrivateKeyFile,
		})
	}

	files := make(map[string]materializedFile, len(materials)*2)
	for _, material := range materials {
		if err := validateCertificateMaterial(material, now); err != nil {
			return nil, err
		}

		files[material.certOutput] = materializedFile{data: append([]byte(nil), material.certPEM...)}
		files[material.keyOutput] = materializedFile{data: append([]byte(nil), material.keyPEM...)}
	}

	return files, nil
}

func validateCertificateMaterial(material certificateMaterial, now time.Time) error {
	certificates, err := parseCertificatePEM(material.certPEM)
	if err != nil {
		return fmt.Errorf("parse certificate for %s: %w", material.certOutput, err)
	}

	privateKey, err := parsePrivateKeyPEM(material.keyPEM)
	if err != nil {
		return fmt.Errorf("parse private key for %s: %w", material.keyOutput, err)
	}

	if err := validateApprovedPrivateKey(privateKey); err != nil {
		return fmt.Errorf("private key for %s is not FIPS-approved: %w", material.keyOutput, err)
	}

	pair, err := tls.X509KeyPair(material.certPEM, material.keyPEM)
	if err != nil {
		return fmt.Errorf("certificate and private key do not match for %s", material.certOutput)
	}

	leaf := certificates[0]
	for _, certificate := range certificates {
		if err := validateApprovedCertificate(certificate); err != nil {
			return fmt.Errorf("certificate chain for %s is not FIPS-approved: %w", material.certOutput, err)
		}
	}

	pair.Leaf = leaf
	if leaf.IsCA {
		return fmt.Errorf("CA certificate cannot be materialized as a leaf")
	}

	if leaf.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return fmt.Errorf("certificate lacks digital-signature key usage")
	}

	if len(leaf.ExtKeyUsage) != 1 || leaf.ExtKeyUsage[0] != material.usage || len(leaf.UnknownExtKeyUsage) != 0 {
		return fmt.Errorf("certificate extended key usage does not match its materializer role")
	}

	if material.expectedDNS != "" && !containsDNSName(leaf.DNSNames, material.expectedDNS) {
		return fmt.Errorf("certificate SAN does not contain its expected DNS identity")
	}

	for _, expectedDNSName := range material.expectedDNSNames {
		if !containsDNSName(leaf.DNSNames, expectedDNSName) {
			return fmt.Errorf("certificate SAN does not contain every required public DNS identity")
		}
	}

	if material.expectedCommonName != "" && leaf.Subject.CommonName != material.expectedCommonName {
		return fmt.Errorf("certificate common name does not match the trusted PostgreSQL pod binding")
	}

	if material.requireLoopback && !containsIP(leaf.IPAddresses, net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)) {
		return fmt.Errorf("administrative certificate SAN does not contain 127.0.0.1")
	}

	if now.Before(leaf.NotBefore) || !now.Before(leaf.NotAfter) {
		return fmt.Errorf("certificate is outside its validity period")
	}

	roots, err := parseTrustRoots(material.trustPEM)
	if err != nil {
		return fmt.Errorf("load explicit trust anchors for %s: %w", material.certOutput, err)
	}

	intermediates := x509.NewCertPool()
	for _, certificate := range certificates[1:] {
		intermediates.AddCert(certificate)
	}

	_, err = leaf.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		DNSName:       material.expectedDNS,
		CurrentTime:   now,
		KeyUsages:     []x509.ExtKeyUsage{material.usage},
	})
	if err != nil {
		return fmt.Errorf("certificate chain or SAN verification failed for %s", material.certOutput)
	}

	return nil
}

func parseCertificatePEM(data []byte) ([]*x509.Certificate, error) {
	if len(data) > cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes {
		return nil, fmt.Errorf("certificate PEM exceeds the size limit")
	}

	remaining := bytes.TrimSpace(data)
	if len(remaining) == 0 {
		return nil, fmt.Errorf("certificate PEM is empty")
	}

	certificates := make([]*x509.Certificate, 0, 2)

	for len(remaining) > 0 {
		if !bytes.HasPrefix(remaining, []byte("-----BEGIN ")) {
			return nil, fmt.Errorf("certificate PEM contains non-certificate data")
		}

		block, rest := pem.Decode(remaining)
		if block == nil || block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate || len(block.Headers) != 0 {
			return nil, fmt.Errorf("certificate PEM contains an invalid block")
		}

		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse X.509 certificate")
		}

		certificates = append(certificates, certificate)
		remaining = bytes.TrimSpace(rest)
	}

	return certificates, nil
}

func parsePrivateKeyPEM(data []byte) (any, error) {
	if len(data) > cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes {
		return nil, fmt.Errorf("private key PEM exceeds the size limit")
	}

	remaining := bytes.TrimSpace(data)
	if !bytes.HasPrefix(remaining, []byte("-----BEGIN ")) {
		return nil, fmt.Errorf("private key PEM is empty or malformed")
	}

	block, rest := pem.Decode(remaining)
	if block == nil || len(block.Headers) != 0 || len(bytes.TrimSpace(rest)) != 0 {
		return nil, fmt.Errorf("private key PEM must contain one unencrypted key")
	}

	var (
		key any
		err error
	)

	switch block.Type {
	case cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey:
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case cryptoutilSharedMagic.StringPEMTypeECPrivateKey:
		key, err = x509.ParseECPrivateKey(block.Bytes)
	case cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey:
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("private key PEM block type is unsupported")
	}

	if err != nil {
		return nil, fmt.Errorf("parse private key PEM")
	}

	return key, nil
}

func parseTrustRoots(data []byte) (*x509.CertPool, error) {
	certificates, err := parseCertificatePEM(data)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()

	for _, certificate := range certificates {
		if !certificate.IsCA || certificate.KeyUsage&x509.KeyUsageCertSign == 0 {
			return nil, fmt.Errorf("trust anchor is not a CA certificate")
		}

		if err := validateApprovedCertificate(certificate); err != nil {
			return nil, fmt.Errorf("trust anchor is not FIPS-approved: %w", err)
		}

		roots.AddCert(certificate)
	}

	return roots, nil
}

func validateApprovedCertificate(certificate *x509.Certificate) error {
	if certificate == nil {
		return fmt.Errorf("certificate is nil")
	}

	if err := validateApprovedPublicKey(certificate.PublicKey); err != nil {
		return err
	}

	return validateApprovedSignatureAlgorithm(certificate.SignatureAlgorithm)
}

func validateApprovedPublicKey(publicKey any) error {
	switch key := publicKey.(type) {
	case *rsa.PublicKey:
		if key == nil || key.N == nil || key.N.BitLen() < cryptoutilSharedMagic.RSAKeySize2048 || key.E < 3 || key.E%2 == 0 {
			return fmt.Errorf("RSA public key is below the approved size or exponent policy")
		}
	case *ecdsa.PublicKey:
		if key == nil || key.Curve == nil || key.Params() == nil {
			return fmt.Errorf("ECDSA public key is malformed")
		}

		if _, err := safeECDSAPublicKeyBytes(key); err != nil {
			return fmt.Errorf("ECDSA public key is malformed: %w", err)
		}

		curveName := strings.ReplaceAll(strings.ToUpper(key.Curve.Params().Name), "-", "")
		switch curveName {
		case cryptoutilSharedMagic.ECCurveP256, cryptoutilSharedMagic.ECCurveP384, cryptoutilSharedMagic.ECCurveP521:
		default:
			return fmt.Errorf("ECDSA public key curve is not approved")
		}
	case ed25519.PublicKey:
		if len(key) != ed25519.PublicKeySize {
			return fmt.Errorf("Ed25519 public key has an invalid size")
		}
	default:
		return fmt.Errorf("public key type is not approved")
	}

	return nil
}

func validateApprovedPrivateKey(privateKey any) error {
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		if key == nil {
			return fmt.Errorf("RSA private key is nil")
		}

		return validateApprovedPublicKey(&key.PublicKey)
	case *ecdsa.PrivateKey:
		if key == nil {
			return fmt.Errorf("ECDSA private key is nil")
		}

		if _, err := safeECDSAPrivateKeyBytes(key); err != nil {
			return fmt.Errorf("ECDSA private key is malformed: %w", err)
		}

		return validateApprovedPublicKey(&key.PublicKey)
	case ed25519.PrivateKey:
		if len(key) != ed25519.PrivateKeySize {
			return fmt.Errorf("Ed25519 private key has an invalid size")
		}

		return validateApprovedPublicKey(key.Public())
	default:
		return fmt.Errorf("private key type is not approved")
	}
}

func safeECDSAPublicKeyBytes(key *ecdsa.PublicKey) (encoded []byte, err error) {
	defer func() {
		if recover() != nil {
			encoded = nil
			err = fmt.Errorf("ECDSA public key encoding failed")
		}
	}()

	encoded, err = key.Bytes()
	if err != nil {
		return nil, fmt.Errorf("encode ECDSA public key: %w", err)
	}

	return encoded, nil
}

func safeECDSAPrivateKeyBytes(key *ecdsa.PrivateKey) (encoded []byte, err error) {
	defer func() {
		if recover() != nil {
			encoded = nil
			err = fmt.Errorf("ECDSA private key encoding failed")
		}
	}()

	encoded, err = key.Bytes()
	if err != nil {
		return nil, fmt.Errorf("encode ECDSA private key: %w", err)
	}

	return encoded, nil
}

func validateApprovedSignatureAlgorithm(algorithm x509.SignatureAlgorithm) error {
	switch algorithm {
	case x509.SHA256WithRSA, x509.SHA384WithRSA, x509.SHA512WithRSA,
		x509.SHA256WithRSAPSS, x509.SHA384WithRSAPSS, x509.SHA512WithRSAPSS,
		x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512,
		x509.PureEd25519:
		return nil
	default:
		return fmt.Errorf("certificate signature algorithm is not approved")
	}
}

func requireExactDataKeys(data map[string][]byte, expected []string) error {
	if len(data) != len(expected) {
		return fmt.Errorf("the Secret contains missing or unexpected data keys")
	}

	for _, key := range expected {
		value, exists := data[key]
		if !exists || len(value) == 0 {
			return fmt.Errorf("the Secret is missing required data key %q", key)
		}
	}

	return nil
}

func containsDNSName(names []string, expected string) bool {
	for _, name := range names {
		if name == expected {
			return true
		}
	}

	return false
}

func containsIP(addresses []net.IP, expected net.IP) bool {
	for _, address := range addresses {
		if address.Equal(expected) {
			return true
		}
	}

	return false
}
