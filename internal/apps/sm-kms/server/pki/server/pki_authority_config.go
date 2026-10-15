// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"bytes"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

const (
	pkiAuthorityValidityTolerance = time.Second
	pkiAuthorityCurveP256         = "P-256"
	pkiAuthorityCurveP384         = "P-384"
	pkiAuthorityCurveP521         = "P-521"
)

// PKIAuthorityConfig defines the stable identity and cryptographic parameters for the
// persisted development authority. Existing persisted material must continue to satisfy
// this configuration; changing it under the same ID fails closed rather than rotating
// the issuer implicitly.
type PKIAuthorityConfig struct {
	ID            string
	Subject       pkix.Name
	KeySpec       cryptoutilCACrypto.KeySpec
	Validity      time.Duration
	MaxPathLength int
	CRLURL        string
	OCSPURL       string
}

// DefaultPKIAuthorityConfig returns the existing development authority identity and
// defaults, preserving compatibility with previously persisted authority material.
func DefaultPKIAuthorityConfig() PKIAuthorityConfig {
	return PKIAuthorityConfig{
		ID: cryptoutilSharedMagic.PKIDefaultAuthorityID,
		Subject: pkix.Name{
			CommonName:   "sm-kms PKI Development CA",
			Organization: []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault},
			Country:      []string{"US"},
		},
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: pkiAuthorityCurveP384,
		},
		Validity:      cryptoutilSharedMagic.TLSDefaultMaxCACertDuration,
		MaxPathLength: defaultRootCAPathLength,
	}
}

type pkiAuthorityYAML struct {
	ID            string   `yaml:"id"`
	CommonName    string   `yaml:"common-name"`
	Organization  []string `yaml:"organization"`
	Country       []string `yaml:"country"`
	KeyAlgorithm  string   `yaml:"key-algorithm"`
	ECDSACurve    string   `yaml:"ecdsa-curve"`
	RSABits       int      `yaml:"rsa-bits"`
	EdDSACurve    string   `yaml:"eddsa-curve"`
	Validity      string   `yaml:"validity"`
	MaxPathLength int      `yaml:"max-path-length"`
	CRLURL        string   `yaml:"crl-url"`
	OCSPURL       string   `yaml:"ocsp-url"`
}

// PKIAuthorityConfigFromConfigFiles loads the optional pki-authority section from the
// service's YAML configuration files. The default authority is used when no file defines it.
func PKIAuthorityConfigFromConfigFiles(paths []string) (PKIAuthorityConfig, error) {
	var (
		configured bool
		config     PKIAuthorityConfig
	)

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return PKIAuthorityConfig{}, fmt.Errorf("read PKI authority configuration file %q: %w", path, err)
		}

		var document map[string]yaml.Node
		if err := yaml.Unmarshal(data, &document); err != nil {
			return PKIAuthorityConfig{}, fmt.Errorf("parse service configuration file %q: %w", path, err)
		}

		authorityNode, exists := document["pki-authority"]
		if !exists {
			continue
		}

		if configured {
			return PKIAuthorityConfig{}, fmt.Errorf("PKI authority configuration is defined in multiple files")
		}

		var authorityYAML pkiAuthorityYAML
		if err := authorityNode.Decode(&authorityYAML); err != nil {
			return PKIAuthorityConfig{}, fmt.Errorf("parse pki-authority section in %q: %w", path, err)
		}

		config, err = authorityYAML.toConfig()
		if err != nil {
			return PKIAuthorityConfig{}, fmt.Errorf("invalid pki-authority section in %q: %w", path, err)
		}

		configured = true
	}

	if !configured {
		config = DefaultPKIAuthorityConfig()
	}

	if err := config.validate(); err != nil {
		return PKIAuthorityConfig{}, fmt.Errorf("invalid PKI authority configuration: %w", err)
	}

	return config, nil
}

func (authority pkiAuthorityYAML) toConfig() (PKIAuthorityConfig, error) {
	validity, err := time.ParseDuration(strings.TrimSpace(authority.Validity))
	if err != nil {
		return PKIAuthorityConfig{}, fmt.Errorf("parse validity: %w", err)
	}

	config := PKIAuthorityConfig{
		ID: authority.ID,
		Subject: pkix.Name{
			CommonName:   authority.CommonName,
			Organization: authority.Organization,
			Country:      authority.Country,
		},
		Validity:      validity,
		MaxPathLength: authority.MaxPathLength,
		CRLURL:        authority.CRLURL,
		OCSPURL:       authority.OCSPURL,
	}

	switch strings.ToUpper(strings.TrimSpace(authority.KeyAlgorithm)) {
	case "ECDSA":
		config.KeySpec = cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: authority.ECDSACurve,
		}
	case cryptoutilSharedMagic.KeyTypeRSA:
		config.KeySpec = cryptoutilCACrypto.KeySpec{
			Type:    cryptoutilCACrypto.KeyTypeRSA,
			RSABits: authority.RSABits,
		}
	case "EDDSA":
		config.KeySpec = cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeEdDSA,
			EdDSACurve: authority.EdDSACurve,
		}
	default:
		return PKIAuthorityConfig{}, fmt.Errorf("unsupported key algorithm %q", authority.KeyAlgorithm)
	}

	return config, nil
}

func (config PKIAuthorityConfig) validate() error {
	if strings.TrimSpace(config.ID) == "" || config.ID != strings.TrimSpace(config.ID) {
		return fmt.Errorf("authority ID must be non-empty and trimmed")
	}

	if url.PathEscape(config.ID) != config.ID {
		return fmt.Errorf("authority ID must be a single URL path segment")
	}

	if strings.TrimSpace(config.Subject.CommonName) == "" {
		return fmt.Errorf("authority subject common name is required")
	}

	if config.Validity <= 0 || config.Validity > cryptoutilSharedMagic.TLSDefaultMaxCACertDuration {
		return fmt.Errorf("authority validity must be positive and no greater than %s", cryptoutilSharedMagic.TLSDefaultMaxCACertDuration)
	}

	if config.MaxPathLength < 0 {
		return fmt.Errorf("authority maximum path length must not be negative")
	}

	if err := validateAuthorityDistributionURL(config.CRLURL, "/pki/cas/"+url.PathEscape(config.ID)+"/crl"); err != nil {
		return fmt.Errorf("invalid authority CRL URL: %w", err)
	}

	if err := validateAuthorityDistributionURL(config.OCSPURL, "/pki/ocsp/"+url.PathEscape(config.ID)); err != nil {
		return fmt.Errorf("invalid authority OCSP URL: %w", err)
	}

	switch config.KeySpec.Type {
	case cryptoutilCACrypto.KeyTypeECDSA:
		switch config.KeySpec.ECDSACurve {
		case pkiAuthorityCurveP256, pkiAuthorityCurveP384, pkiAuthorityCurveP521:
			return nil
		default:
			return fmt.Errorf("unsupported authority ECDSA curve %q", config.KeySpec.ECDSACurve)
		}
	case cryptoutilCACrypto.KeyTypeRSA:
		if config.KeySpec.RSABits < cryptoutilCACrypto.MinRSAKeyBits {
			return fmt.Errorf("authority RSA key size must be at least %d bits", cryptoutilCACrypto.MinRSAKeyBits)
		}

		return nil
	case cryptoutilCACrypto.KeyTypeEdDSA:
		if config.KeySpec.EdDSACurve == cryptoutilSharedMagic.EdCurveEd25519 {
			return nil
		}

		return fmt.Errorf("unsupported authority EdDSA curve %q", config.KeySpec.EdDSACurve)
	default:
		return fmt.Errorf("unsupported authority key type %q", config.KeySpec.Type)
	}
}

func validateAuthorityDistributionURL(rawURL, expectedPath string) error {
	if rawURL == "" {
		return nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	if parsedURL.Scheme != cryptoutilSharedMagic.ProtocolHTTPS || parsedURL.Host == "" || parsedURL.User != nil || parsedURL.Fragment != "" {
		return fmt.Errorf("URL must be an absolute HTTPS URL without user information or fragment")
	}

	if parsedURL.Path != expectedPath || parsedURL.RawQuery != "" {
		return fmt.Errorf("URL path must be %q and must not include a query", expectedPath)
	}

	return nil
}

func validateAuthorityMaterial(certificate *x509.Certificate, privateKey crypto.PrivateKey, config PKIAuthorityConfig) error {
	if err := config.validate(); err != nil {
		return fmt.Errorf("invalid authority configuration: %w", err)
	}

	if certificate == nil {
		return fmt.Errorf("authority certificate is required")
	}

	if !certificate.IsCA || !certificate.BasicConstraintsValid {
		return fmt.Errorf("authority certificate must be a CA with valid basic constraints")
	}

	requiredKeyUsage := x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	if certificate.KeyUsage&requiredKeyUsage != requiredKeyUsage {
		return fmt.Errorf("authority certificate must permit certificate and CRL signing")
	}

	if len(certificate.SubjectKeyId) == 0 {
		return fmt.Errorf("authority certificate subject key identifier is required")
	}

	if err := certificate.CheckSignatureFrom(certificate); err != nil {
		return fmt.Errorf("authority certificate is not self-signed: %w", err)
	}

	now := time.Now().UTC()
	if certificate.NotBefore.After(now) || !certificate.NotAfter.After(now) {
		return fmt.Errorf("authority certificate is not currently valid")
	}

	certificateSubjectDER, err := asn1.Marshal(config.Subject.ToRDNSequence())
	if err != nil {
		return fmt.Errorf("encode configured authority subject: %w", err)
	}

	if !bytes.Equal(certificate.RawSubject, certificateSubjectDER) {
		return fmt.Errorf("persisted authority certificate subject does not match configured identity")
	}

	if certificate.NotAfter.Sub(certificate.NotBefore) > config.Validity+pkiAuthorityValidityTolerance ||
		certificate.NotAfter.Sub(certificate.NotBefore) < config.Validity-pkiAuthorityValidityTolerance {
		return fmt.Errorf("persisted authority certificate validity does not match configured lifetime")
	}

	if certificate.MaxPathLen != config.MaxPathLength ||
		certificate.MaxPathLenZero != (config.MaxPathLength == 0) {
		return fmt.Errorf("persisted authority certificate path length does not match configured identity")
	}

	if err := validateAuthorityPublicKey(certificate.PublicKey, config.KeySpec); err != nil {
		return err
	}

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return fmt.Errorf("authority private key is not a signer")
	}

	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		return fmt.Errorf("encode authority certificate public key: %w", err)
	}

	privateKeyPublicKey, err := x509.MarshalPKIXPublicKey(signer.Public())
	if err != nil {
		return fmt.Errorf("encode authority private key public component: %w", err)
	}

	if !bytes.Equal(certificatePublicKey, privateKeyPublicKey) {
		return fmt.Errorf("authority certificate and private key do not match")
	}

	return nil
}

func validateAuthorityPublicKey(publicKey crypto.PublicKey, keySpec cryptoutilCACrypto.KeySpec) error {
	switch keySpec.Type {
	case cryptoutilCACrypto.KeyTypeECDSA:
		ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
		if !ok || ecdsaPublicKey.Curve.Params().Name != keySpec.ECDSACurve {
			return fmt.Errorf("authority certificate public key does not match configured ECDSA curve %q", keySpec.ECDSACurve)
		}
	case cryptoutilCACrypto.KeyTypeRSA:
		rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
		if !ok || rsaPublicKey.N.BitLen() != keySpec.RSABits {
			return fmt.Errorf("authority certificate public key does not match configured RSA key size %d", keySpec.RSABits)
		}
	case cryptoutilCACrypto.KeyTypeEdDSA:
		if _, ok := publicKey.(ed25519.PublicKey); !ok {
			return fmt.Errorf("authority certificate public key does not match configured EdDSA curve")
		}
	default:
		return fmt.Errorf("unsupported configured authority public key type %q", keySpec.Type)
	}

	return nil
}
