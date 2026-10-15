// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type contentCipherFn func(context.Context, []byte) ([]byte, error)

func getOrCreateCA(
	ctx context.Context,
	provider cryptoutilCACrypto.Provider,
	randomReader io.Reader,
	authorityStore cryptoutilCAStorage.AuthorityStore,
	encryptFn contentCipherFn,
	decryptFn contentCipherFn,
) (*x509.Certificate, crypto.PrivateKey, error) {
	return getOrCreateCAWithConfig(ctx, provider, randomReader, authorityStore, encryptFn, decryptFn, DefaultPKIAuthorityConfig())
}

func getOrCreateCAWithConfig(
	ctx context.Context,
	provider cryptoutilCACrypto.Provider,
	randomReader io.Reader,
	authorityStore cryptoutilCAStorage.AuthorityStore,
	encryptFn contentCipherFn,
	decryptFn contentCipherFn,
	authorityConfig PKIAuthorityConfig,
) (*x509.Certificate, crypto.PrivateKey, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("context is required")
	}

	if provider == nil {
		return nil, nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorCryptoProviderRequired)
	}

	if authorityStore == nil {
		return nil, nil, fmt.Errorf("authority store is required")
	}

	if encryptFn == nil || decryptFn == nil {
		return nil, nil, fmt.Errorf("authority key encryption and decryption are required")
	}

	if err := authorityConfig.validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid PKI authority configuration: %w", err)
	}

	authority, err := authorityStore.Get(ctx, authorityConfig.ID)
	switch {
	case err == nil:
		return restoreCAWithConfig(ctx, authority, decryptFn, authorityConfig)
	case !errors.Is(err, cryptoutilCAStorage.ErrAuthorityNotFound):
		return nil, nil, fmt.Errorf("failed to load persisted PKI authority: %w", err)
	}

	certificate, privateKey, err := createSelfSignedCAWithConfig(provider, randomReader, authorityConfig)
	if err != nil {
		return nil, nil, err
	}

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode PKI authority private key: %w", err)
	}

	encryptedPrivateKey, err := encryptFn(ctx, privateKeyDER)
	clear(privateKeyDER)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt PKI authority private key: %w", err)
	}

	authority = &cryptoutilCAStorage.StoredAuthority{
		ID:                  authorityConfig.ID,
		CertificateDER:      certificate.Raw,
		PrivateKeyEncrypted: encryptedPrivateKey,
	}
	if err := authorityStore.Create(ctx, authority); err != nil {
		persistedAuthority, getErr := authorityStore.Get(ctx, authorityConfig.ID)
		if getErr != nil {
			return nil, nil, fmt.Errorf("failed to persist PKI authority: %w", err)
		}

		return restoreCAWithConfig(ctx, persistedAuthority, decryptFn, authorityConfig)
	}

	return certificate, privateKey, nil
}

func restoreHistoricalCA(
	ctx context.Context,
	authority *cryptoutilCAStorage.StoredAuthority,
	decryptFn contentCipherFn,
) (*x509.Certificate, crypto.PrivateKey, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("context is required")
	}

	if authority == nil || authority.ID == "" || authority.ID != strings.TrimSpace(authority.ID) {
		return nil, nil, fmt.Errorf("historical PKI authority with a non-empty ID is required")
	}

	if decryptFn == nil {
		return nil, nil, fmt.Errorf("authority key decryption is required")
	}

	certificate, err := x509.ParseCertificate(authority.CertificateDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse historical PKI authority certificate: %w", err)
	}

	if !certificate.IsCA || !certificate.BasicConstraintsValid {
		return nil, nil, fmt.Errorf("historical authority certificate must be a CA with valid basic constraints")
	}

	requiredKeyUsage := x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	if certificate.KeyUsage&requiredKeyUsage != requiredKeyUsage {
		return nil, nil, fmt.Errorf("historical authority certificate must permit certificate and CRL signing")
	}

	if len(certificate.SubjectKeyId) == 0 {
		return nil, nil, fmt.Errorf("historical authority certificate subject key identifier is required")
	}

	if err := certificate.CheckSignatureFrom(certificate); err != nil {
		return nil, nil, fmt.Errorf("historical authority certificate is not self-signed: %w", err)
	}

	if certificate.NotBefore.After(time.Now().UTC()) {
		return nil, nil, fmt.Errorf("historical authority certificate is not yet valid")
	}

	privateKeyDER, err := decryptFn(ctx, authority.PrivateKeyEncrypted)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt historical PKI authority private key: %w", err)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyDER)
	clear(privateKeyDER)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse historical PKI authority private key: %w", err)
	}

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("historical authority private key is not a signer")
	}

	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("encode historical authority certificate public key: %w", err)
	}

	privateKeyPublicKey, err := x509.MarshalPKIXPublicKey(signer.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("encode historical authority private key public component: %w", err)
	}

	if !bytes.Equal(certificatePublicKey, privateKeyPublicKey) {
		return nil, nil, fmt.Errorf("historical authority certificate and private key do not match")
	}

	return certificate, privateKey, nil
}

func restoreCAWithConfig(
	ctx context.Context,
	authority *cryptoutilCAStorage.StoredAuthority,
	decryptFn contentCipherFn,
	authorityConfig PKIAuthorityConfig,
) (*x509.Certificate, crypto.PrivateKey, error) {
	if authority == nil {
		return nil, nil, fmt.Errorf("persisted PKI authority is required")
	}

	if authority.ID != authorityConfig.ID {
		return nil, nil, fmt.Errorf("persisted PKI authority ID %q does not match configured ID %q", authority.ID, authorityConfig.ID)
	}

	certificate, err := x509.ParseCertificate(authority.CertificateDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse persisted PKI authority certificate: %w", err)
	}

	if decryptFn == nil {
		return nil, nil, fmt.Errorf("authority key decryption is required")
	}

	privateKeyDER, err := decryptFn(ctx, authority.PrivateKeyEncrypted)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt persisted PKI authority private key: %w", err)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyDER)
	clear(privateKeyDER)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse persisted PKI authority private key: %w", err)
	}

	if err := validateAuthorityMaterial(certificate, privateKey, authorityConfig); err != nil {
		return nil, nil, fmt.Errorf("persisted PKI authority failed lifecycle validation: %w", err)
	}

	return certificate, privateKey, nil
}
