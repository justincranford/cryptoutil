// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestGetOrCreateCA_PersistsAuthorityAcrossRestart(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilCAStorage.StoredAuthority{}))

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	encrypt := func(_ context.Context, cleartext []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(cleartext)), nil
	}
	decrypt := func(_ context.Context, ciphertext []byte) ([]byte, error) {
		cleartext, err := base64.StdEncoding.DecodeString(string(ciphertext))
		if err != nil {
			return nil, err
		}

		return cleartext, nil
	}

	certificate, privateKey, err := getOrCreateCA(context.Background(), cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, authorityStore, encrypt, decrypt)
	require.NoError(t, err)
	require.NotNil(t, certificate)
	require.NotNil(t, privateKey)

	persisted, err := authorityStore.Get(context.Background(), cryptoutilSharedMagic.PKIDefaultAuthorityID)
	require.NoError(t, err)
	require.Equal(t, certificate.Raw, persisted.CertificateDER)
	require.NotEmpty(t, persisted.PrivateKeyEncrypted)
	require.Equal(t, cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1, persisted.PrivateKeyEnvelopeVersion)

	restartedCertificate, restartedKey, err := getOrCreateCA(context.Background(), cryptoutilCACrypto.NewSoftwareProvider(), nil, authorityStore, encrypt, decrypt)
	require.NoError(t, err)
	require.Equal(t, certificate.Raw, restartedCertificate.Raw)

	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	require.NoError(t, err)

	signer, ok := restartedKey.(crypto.Signer)
	require.True(t, ok)

	restartedPublicKey, err := x509.MarshalPKIXPublicKey(signer.Public())
	require.NoError(t, err)
	require.Equal(t, certificatePublicKey, restartedPublicKey)
}

func TestGetOrCreateCA_DistributionURLsAreDeploymentConfiguration(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilCAStorage.StoredAuthority{}))

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	config := DefaultPKIAuthorityConfig()
	config.CRLURL = "https://pki.example.test/pki/cas/" + config.ID + "/crl"
	config.OCSPURL = "https://pki.example.test/pki/ocsp/" + config.ID

	encrypt := func(_ context.Context, cleartext []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(cleartext)), nil
	}
	decrypt := func(_ context.Context, ciphertext []byte) ([]byte, error) {
		return base64.StdEncoding.DecodeString(string(ciphertext))
	}

	firstCertificate, _, err := getOrCreateCAWithConfig(
		context.Background(),
		cryptoutilCACrypto.NewSoftwareProvider(),
		crand.Reader,
		authorityStore,
		encrypt,
		decrypt,
		config,
	)
	require.NoError(t, err)

	changedConfig := config
	changedConfig.CRLURL = "https://pki-new.example.test/pki/cas/" + config.ID + "/crl"
	changedConfig.OCSPURL = "https://pki-new.example.test/pki/ocsp/" + config.ID
	restartedCertificate, _, err := getOrCreateCAWithConfig(
		context.Background(),
		cryptoutilCACrypto.NewSoftwareProvider(),
		nil,
		authorityStore,
		encrypt,
		decrypt,
		changedConfig,
	)
	require.NoError(t, err)
	require.Equal(t, firstCertificate.Raw, restartedCertificate.Raw)
}

func TestGetOrCreateCA_EncryptionFailureDoesNotPersistKey(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilCAStorage.StoredAuthority{}))

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	encryptErr := errors.New("injected encryption failure")
	encrypt := func(context.Context, []byte) ([]byte, error) {
		return nil, encryptErr
	}
	decrypt := func(context.Context, []byte) ([]byte, error) {
		return nil, errors.New("unexpected decrypt")
	}

	_, _, err = getOrCreateCA(context.Background(), cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, authorityStore, encrypt, decrypt)
	require.ErrorIs(t, err, encryptErr)

	var count int64
	require.NoError(t, db.Model(&cryptoutilCAStorage.StoredAuthority{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestGetOrCreateCA_PersistenceFailureDoesNotReturnUnpersistedAuthority(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilCAStorage.StoredAuthority{}))

	baseStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	persistErr := errors.New("injected persistence failure")
	authorityStore := authorityStoreCreateWrapper{
		AuthorityStore: baseStore,
		createFn: func(context.Context, *cryptoutilCAStorage.StoredAuthority) error {
			return persistErr
		},
	}
	encrypt := func(_ context.Context, cleartext []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(cleartext)), nil
	}
	decrypt := func(context.Context, []byte) ([]byte, error) {
		return nil, errors.New("unexpected decrypt")
	}

	_, _, err = getOrCreateCA(context.Background(), cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, authorityStore, encrypt, decrypt)
	require.ErrorIs(t, err, persistErr)

	var count int64
	require.NoError(t, db.Model(&cryptoutilCAStorage.StoredAuthority{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestGetOrCreateCA_ConcurrentCreationUsesPersistedAuthority(t *testing.T) {
	t.Parallel()

	authorityStore := newAuthorityRaceStore()
	encrypt := func(_ context.Context, cleartext []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(cleartext)), nil
	}
	decrypt := func(_ context.Context, ciphertext []byte) ([]byte, error) {
		return base64.StdEncoding.DecodeString(string(ciphertext))
	}

	type result struct {
		certificate *x509.Certificate
		privateKey  crypto.PrivateKey
		err         error
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results := make(chan result, 2)

	for range 2 {
		go func() {
			certificate, privateKey, err := getOrCreateCA(ctx, cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, authorityStore, encrypt, decrypt)
			results <- result{certificate: certificate, privateKey: privateKey, err: err}
		}()
	}

	for range 2 {
		select {
		case <-authorityStore.initialGetReady:
		case <-ctx.Done():
			require.FailNow(t, "concurrent authority creation did not reach persistence", ctx.Err().Error())
		}
	}

	close(authorityStore.releaseInitialGet)

	first := <-results
	second := <-results

	require.NoError(t, first.err)
	require.NoError(t, second.err)
	require.Equal(t, first.certificate.Raw, second.certificate.Raw)

	firstSigner, ok := first.privateKey.(crypto.Signer)
	require.True(t, ok)
	secondSigner, ok := second.privateKey.(crypto.Signer)
	require.True(t, ok)

	firstPublicKey, err := x509.MarshalPKIXPublicKey(firstSigner.Public())
	require.NoError(t, err)
	secondPublicKey, err := x509.MarshalPKIXPublicKey(secondSigner.Public())
	require.NoError(t, err)
	require.Equal(t, firstPublicKey, secondPublicKey)
}

func TestGetOrCreateCA_ConfigurableIdentitySupportsRollover(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilCAStorage.StoredAuthority{}))

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	encrypt := func(_ context.Context, cleartext []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(cleartext)), nil
	}
	decrypt := func(_ context.Context, ciphertext []byte) ([]byte, error) {
		return base64.StdEncoding.DecodeString(string(ciphertext))
	}

	config := DefaultPKIAuthorityConfig()
	config.ID = "sm-kms-pki-test-authority"
	config.Subject.CommonName = "Test PKI Authority"
	config.Subject.Organization = []string{"cryptoutil test"}
	config.KeySpec.ECDSACurve = pkiAuthorityCurveP256
	config.Validity = 365 * 24 * time.Hour
	config.MaxPathLength = 0

	certificate, privateKey, err := getOrCreateCAWithConfig(
		context.Background(),
		cryptoutilCACrypto.NewSoftwareProvider(),
		crand.Reader,
		authorityStore,
		encrypt,
		decrypt,
		config,
	)
	require.NoError(t, err)
	require.Equal(t, config.Subject.CommonName, certificate.Subject.CommonName)
	require.True(t, certificate.MaxPathLenZero)
	require.NotEmpty(t, certificate.SubjectKeyId)
	require.NotNil(t, privateKey)

	persisted, err := authorityStore.Get(context.Background(), config.ID)
	require.NoError(t, err)

	restartedCertificate, restartedKey, err := getOrCreateCAWithConfig(
		context.Background(),
		cryptoutilCACrypto.NewSoftwareProvider(),
		nil,
		authorityStore,
		encrypt,
		decrypt,
		config,
	)
	require.NoError(t, err)
	require.Equal(t, certificate.Raw, restartedCertificate.Raw)

	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	require.NoError(t, err)

	restartedSigner, ok := restartedKey.(crypto.Signer)
	require.True(t, ok)

	restartedPublicKey, err := x509.MarshalPKIXPublicKey(restartedSigner.Public())
	require.NoError(t, err)
	require.Equal(t, certificatePublicKey, restartedPublicKey)

	mismatchedConfig := config
	mismatchedConfig.Subject.CommonName = "Unexpected replacement authority"

	_, _, err = getOrCreateCAWithConfig(
		context.Background(),
		cryptoutilCACrypto.NewSoftwareProvider(),
		nil,
		authorityStore,
		encrypt,
		decrypt,
		mismatchedConfig,
	)
	require.ErrorContains(t, err, "subject does not match configured identity")
	require.Equal(t, certificate.Raw, persisted.CertificateDER)

	rotatedIdentity := config
	rotatedIdentity.ID = "replacement-pki-authority"
	rotatedCertificate, _, err := getOrCreateCAWithConfig(
		context.Background(),
		cryptoutilCACrypto.NewSoftwareProvider(),
		crand.Reader,
		authorityStore,
		encrypt,
		decrypt,
		rotatedIdentity,
	)
	require.NoError(t, err)
	require.NotEqual(t, certificate.Raw, rotatedCertificate.Raw)

	var authorityCount int64
	require.NoError(t, db.Model(&cryptoutilCAStorage.StoredAuthority{}).Count(&authorityCount).Error)
	require.Equal(t, int64(2), authorityCount)

	restoredOriginalCertificate, _, err := getOrCreateCAWithConfig(
		context.Background(),
		cryptoutilCACrypto.NewSoftwareProvider(),
		nil,
		authorityStore,
		encrypt,
		decrypt,
		config,
	)
	require.NoError(t, err)
	require.Equal(t, certificate.Raw, restoredOriginalCertificate.Raw)
}

func TestRestoreCA_RejectsUnsupportedPrivateKeyEnvelopeVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		restore func(*cryptoutilCAStorage.StoredAuthority) error
	}{
		{
			name: "configured authority",
			restore: func(authority *cryptoutilCAStorage.StoredAuthority) error {
				_, _, err := restoreCAWithConfig(
					context.Background(),
					authority,
					func(context.Context, []byte) ([]byte, error) {
						return nil, errors.New("unexpected decrypt")
					},
					DefaultPKIAuthorityConfig(),
				)

				return err
			},
		},
		{
			name: "historical authority",
			restore: func(authority *cryptoutilCAStorage.StoredAuthority) error {
				_, _, err := restoreHistoricalCA(
					context.Background(),
					authority,
					func(context.Context, []byte) ([]byte, error) {
						return nil, errors.New("unexpected decrypt")
					},
				)

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authority := &cryptoutilCAStorage.StoredAuthority{
				ID:                        cryptoutilSharedMagic.PKIDefaultAuthorityID,
				PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1 + 1,
			}
			require.ErrorContains(t, tc.restore(authority), "unsupported PKI authority private key envelope version")
		})
	}
}

type authorityStoreCreateWrapper struct {
	cryptoutilCAStorage.AuthorityStore
	createFn func(context.Context, *cryptoutilCAStorage.StoredAuthority) error
}

func (s authorityStoreCreateWrapper) Create(ctx context.Context, authority *cryptoutilCAStorage.StoredAuthority) error {
	return s.createFn(ctx, authority)
}

type authorityRaceStore struct {
	mu                sync.Mutex
	authorities       map[string]*cryptoutilCAStorage.StoredAuthority
	initialGetCount   int
	initialGetReady   chan struct{}
	releaseInitialGet chan struct{}
}

func newAuthorityRaceStore() *authorityRaceStore {
	return &authorityRaceStore{
		authorities:       make(map[string]*cryptoutilCAStorage.StoredAuthority),
		initialGetReady:   make(chan struct{}, 2),
		releaseInitialGet: make(chan struct{}),
	}
}

func (s *authorityRaceStore) Get(ctx context.Context, id string) (*cryptoutilCAStorage.StoredAuthority, error) {
	s.mu.Lock()

	authority, exists := s.authorities[id]
	if !exists && s.initialGetCount < 2 {
		s.initialGetCount++
		s.mu.Unlock()

		s.initialGetReady <- struct{}{}

		select {
		case <-s.releaseInitialGet:
			return nil, cryptoutilCAStorage.ErrAuthorityNotFound
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	s.mu.Unlock()

	if !exists {
		return nil, cryptoutilCAStorage.ErrAuthorityNotFound
	}

	return authority, nil
}

func (s *authorityRaceStore) GetAny(context.Context) (*cryptoutilCAStorage.StoredAuthority, error) {
	return nil, cryptoutilCAStorage.ErrAuthorityNotFound
}

func (s *authorityRaceStore) List(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
	return nil, nil
}

func (s *authorityRaceStore) Create(_ context.Context, authority *cryptoutilCAStorage.StoredAuthority) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.authorities[authority.ID]; exists {
		return errors.New("authority already exists")
	}

	s.authorities[authority.ID] = authority

	return nil
}
