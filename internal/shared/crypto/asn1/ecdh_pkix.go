// Copyright (c) 2025-2026 Justin Cranford.
//

package asn1

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/x509"
	x509pkix "crypto/x509/pkix"
	stdasn1 "encoding/asn1"
	"fmt"
	"strconv"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type pkcs8PrivateKeyInfo struct {
	Version    int
	Algorithm  x509pkix.AlgorithmIdentifier
	PrivateKey []byte
}

type pkixPublicKeyInfo struct {
	Algorithm x509pkix.AlgorithmIdentifier
	PublicKey stdasn1.BitString
}

func marshalECDHPrivateKeyInfo(
	key *ecdh.PrivateKey,
	marshalPKCS8Fn func(any) ([]byte, error),
	parseOIDFn func(string) (stdasn1.ObjectIdentifier, error),
	marshalASN1Fn func(any) ([]byte, error),
) ([]byte, error) {
	der, err := marshalPKCS8Fn(key)
	if err != nil {
		return nil, fmt.Errorf("marshal PKCS#8 ECDH private key: %w", err)
	}

	switch curve := key.Curve(); curve {
	case ecdh.X25519():
		return der, nil
	case ecdh.P256(), ecdh.P384(), ecdh.P521():
	default:
		return nil, fmt.Errorf("unsupported ECDH curve: %s", curve)
	}

	var info pkcs8PrivateKeyInfo

	rest, err := stdasn1.Unmarshal(der, &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal PKCS#8 ECDH private key: %w", err)
	}

	switch {
	case len(rest) != 0:
		return nil, fmt.Errorf("PKCS#8 ECDH private key has trailing data")
	case info.Algorithm.Algorithm.String() != cryptoutilSharedMagic.StringOIDPublicKeyEC:
		return nil, fmt.Errorf("unexpected ECDH private-key algorithm OID: %s", info.Algorithm.Algorithm)
	case len(info.Algorithm.Parameters.FullBytes) == 0:
		return nil, fmt.Errorf("PKCS#8 ECDH private key is missing curve parameters")
	}

	oidECDH, err := parseOIDFn(cryptoutilSharedMagic.StringOIDPublicKeyECDH)
	if err != nil {
		return nil, fmt.Errorf("parse id-ecDH OID: %w", err)
	}

	info.Algorithm.Algorithm = oidECDH

	der, err = marshalASN1Fn(info)
	if err != nil {
		return nil, fmt.Errorf("marshal id-ecDH PKCS#8 private key: %w", err)
	}

	return der, nil
}

func marshalECDHPublicKeyInfo(
	key *ecdh.PublicKey,
	marshalPKIXFn func(any) ([]byte, error),
	parseOIDFn func(string) (stdasn1.ObjectIdentifier, error),
	marshalASN1Fn func(any) ([]byte, error),
) ([]byte, error) {
	der, err := marshalPKIXFn(key)
	if err != nil {
		return nil, fmt.Errorf("marshal PKIX ECDH public key: %w", err)
	}

	switch curve := key.Curve(); curve {
	case ecdh.X25519():
		return der, nil
	case ecdh.P256(), ecdh.P384(), ecdh.P521():
	default:
		return nil, fmt.Errorf("unsupported ECDH curve: %s", curve)
	}

	var info pkixPublicKeyInfo

	rest, err := stdasn1.Unmarshal(der, &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal PKIX ECDH public key: %w", err)
	}

	switch {
	case len(rest) != 0:
		return nil, fmt.Errorf("PKIX ECDH public key has trailing data")
	case info.Algorithm.Algorithm.String() != cryptoutilSharedMagic.StringOIDPublicKeyEC:
		return nil, fmt.Errorf("unexpected ECDH public-key algorithm OID: %s", info.Algorithm.Algorithm)
	case len(info.Algorithm.Parameters.FullBytes) == 0:
		return nil, fmt.Errorf("PKIX ECDH public key is missing curve parameters")
	}

	oidECDH, err := parseOIDFn(cryptoutilSharedMagic.StringOIDPublicKeyECDH)
	if err != nil {
		return nil, fmt.Errorf("parse id-ecDH OID: %w", err)
	}

	info.Algorithm.Algorithm = oidECDH

	der, err = marshalASN1Fn(info)
	if err != nil {
		return nil, fmt.Errorf("marshal id-ecDH PKIX public key: %w", err)
	}

	return der, nil
}

func parsePKCS8PrivateKey(der []byte) (any, error) {
	return parsePKCS8PrivateKeyWithFns(der, parseObjectIdentifier, stdasn1.Marshal, x509.ParsePKCS8PrivateKey)
}

func parsePKCS8PrivateKeyWithFns(
	der []byte,
	parseOIDFn func(string) (stdasn1.ObjectIdentifier, error),
	marshalASN1Fn func(any) ([]byte, error),
	parsePKCS8Fn func([]byte) (any, error),
) (any, error) {
	var info pkcs8PrivateKeyInfo

	rest, err := stdasn1.Unmarshal(der, &info)
	if err != nil {
		return parsePKCS8Fn(der)
	}

	switch {
	case info.Algorithm.Algorithm.String() != cryptoutilSharedMagic.StringOIDPublicKeyECDH:
		return parsePKCS8Fn(der)
	case len(rest) != 0:
		return nil, fmt.Errorf("PKCS#8 private key has trailing data")
	}

	oidEC, err := parseOIDFn(cryptoutilSharedMagic.StringOIDPublicKeyEC)
	if err != nil {
		return nil, fmt.Errorf("parse id-ecPublicKey OID: %w", err)
	}

	info.Algorithm.Algorithm = oidEC

	ecdsaDER, err := marshalASN1Fn(info)
	if err != nil {
		return nil, fmt.Errorf("marshal id-ecPublicKey PKCS#8 private key: %w", err)
	}

	parsedKey, err := parsePKCS8Fn(ecdsaDER)
	if err != nil {
		return nil, fmt.Errorf("parse id-ecDH PKCS#8 private key: %w", err)
	}

	ecdsaPrivateKey, ok := parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unexpected private-key type for id-ecDH: %T", parsedKey)
	}

	ecdhPrivateKey, err := ecdsaPrivateKey.ECDH()
	if err != nil {
		return nil, fmt.Errorf("convert id-ecDH private key: %w", err)
	}

	return ecdhPrivateKey, nil
}

func parsePKIXPublicKey(der []byte) (any, error) {
	return parsePKIXPublicKeyWithFns(der, parseObjectIdentifier, stdasn1.Marshal, x509.ParsePKIXPublicKey)
}

func parsePKIXPublicKeyWithFns(
	der []byte,
	parseOIDFn func(string) (stdasn1.ObjectIdentifier, error),
	marshalASN1Fn func(any) ([]byte, error),
	parsePKIXFn func([]byte) (any, error),
) (any, error) {
	var info pkixPublicKeyInfo

	rest, err := stdasn1.Unmarshal(der, &info)
	if err != nil {
		return parsePKIXFn(der)
	}

	switch {
	case info.Algorithm.Algorithm.String() != cryptoutilSharedMagic.StringOIDPublicKeyECDH:
		return parsePKIXFn(der)
	case len(rest) != 0:
		return nil, fmt.Errorf("PKIX public key has trailing data")
	}

	oidEC, err := parseOIDFn(cryptoutilSharedMagic.StringOIDPublicKeyEC)
	if err != nil {
		return nil, fmt.Errorf("parse id-ecPublicKey OID: %w", err)
	}

	info.Algorithm.Algorithm = oidEC

	ecdsaDER, err := marshalASN1Fn(info)
	if err != nil {
		return nil, fmt.Errorf("marshal id-ecPublicKey PKIX public key: %w", err)
	}

	parsedKey, err := parsePKIXFn(ecdsaDER)
	if err != nil {
		return nil, fmt.Errorf("parse id-ecDH PKIX public key: %w", err)
	}

	ecdsaPublicKey, ok := parsedKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unexpected public-key type for id-ecDH: %T", parsedKey)
	}

	ecdhPublicKey, err := ecdsaPublicKey.ECDH()
	if err != nil {
		return nil, fmt.Errorf("convert id-ecDH public key: %w", err)
	}

	return ecdhPublicKey, nil
}

func parseObjectIdentifier(value string) (stdasn1.ObjectIdentifier, error) {
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid object identifier %q", value)
	}

	oid := make(stdasn1.ObjectIdentifier, len(parts))

	for i, part := range parts {
		arc, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid object identifier arc %q: %w", part, err)
		}

		if arc < 0 {
			return nil, fmt.Errorf("negative object identifier arc %d", arc)
		}

		oid[i] = arc
	}

	switch {
	case oid[0] > 2:
		return nil, fmt.Errorf("invalid first object identifier arc %d", oid[0])
	case oid[0] < 2 && oid[1] > 39:
		return nil, fmt.Errorf("invalid second object identifier arc %d", oid[1])
	default:
		return oid, nil
	}
}
