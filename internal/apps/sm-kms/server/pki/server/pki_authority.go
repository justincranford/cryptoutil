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

	authority, err := authorityStore.Get(ctx, cryptoutilSharedMagic.PKIDefaultAuthorityID)
	switch {
	case err == nil:
		return restoreCA(ctx, authority, decryptFn)
	case !errors.Is(err, cryptoutilCAStorage.ErrAuthorityNotFound):
		return nil, nil, fmt.Errorf("failed to load persisted PKI authority: %w", err)
	}

	certificate, privateKey, err := createSelfSignedCA(provider, randomReader)
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
		ID:                  cryptoutilSharedMagic.PKIDefaultAuthorityID,
		CertificateDER:      certificate.Raw,
		PrivateKeyEncrypted: encryptedPrivateKey,
	}
	if err := authorityStore.Create(ctx, authority); err != nil {
		persistedAuthority, getErr := authorityStore.Get(ctx, cryptoutilSharedMagic.PKIDefaultAuthorityID)
		if getErr != nil {
			return nil, nil, fmt.Errorf("failed to persist PKI authority: %w", err)
		}

		return restoreCA(ctx, persistedAuthority, decryptFn)
	}

	return certificate, privateKey, nil
}

func restoreCA(ctx context.Context, authority *cryptoutilCAStorage.StoredAuthority, decryptFn contentCipherFn) (*x509.Certificate, crypto.PrivateKey, error) {
	if authority == nil {
		return nil, nil, fmt.Errorf("persisted PKI authority is required")
	}

	certificate, err := x509.ParseCertificate(authority.CertificateDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse persisted PKI authority certificate: %w", err)
	}

	if err := certificate.CheckSignatureFrom(certificate); err != nil {
		return nil, nil, fmt.Errorf("persisted PKI authority certificate is not self-signed: %w", err)
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

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("persisted PKI authority private key is not a signer")
	}

	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode persisted PKI authority public key: %w", err)
	}

	privateKeyPublicKey, err := x509.MarshalPKIXPublicKey(signer.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode persisted PKI authority private key public component: %w", err)
	}

	if !bytes.Equal(certificatePublicKey, privateKeyPublicKey) {
		return nil, nil, fmt.Errorf("persisted PKI authority certificate and private key do not match")
	}

	return certificate, privateKey, nil
}
