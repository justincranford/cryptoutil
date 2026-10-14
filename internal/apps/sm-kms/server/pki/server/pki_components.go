// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	"time"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilCAHandler "cryptoutil/internal/apps/sm-kms/server/pki/api/handler"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gorm.io/gorm"
)

const (
	defaultCRLValidity         = cryptoutilSharedMagic.Days7
	defaultCRLNextUpdateBuffer = cryptoutilSharedMagic.Days1
	defaultOCSPValidity        = cryptoutilSharedMagic.Days1
	serialNumberBitSize        = cryptoutilSharedMagic.PKICASerialNumberBits
	defaultRootCAPathLength    = 2
)

type PKIComponents struct {
	issuer      *cryptoutilCAServiceIssuer.Issuer
	storage     cryptoutilCAStorage.Store
	crlService  *cryptoutilCAServiceRevocation.CRLService
	ocspService *cryptoutilCAServiceRevocation.OCSPService
	handler     *cryptoutilCAHandler.Handler
}

type pkiComponentFactories struct {
	provider         cryptoutilCACrypto.Provider
	storage          cryptoutilCAStorage.Store
	randomReader     io.Reader
	loadOrCreateCAFn func(cryptoutilCACrypto.Provider, io.Reader) (*x509.Certificate, crypto.PrivateKey, error)
	newIssuerFn      func(cryptoutilCACrypto.Provider, *cryptoutilCAServiceIssuer.IssuingCAConfig) (*cryptoutilCAServiceIssuer.Issuer, error)
	newCRLServiceFn  func(*cryptoutilCAServiceRevocation.CRLConfig) (*cryptoutilCAServiceRevocation.CRLService, error)
	newOCSPServiceFn func(*cryptoutilCAServiceRevocation.OCSPConfig, *cryptoutilCAServiceRevocation.CRLService) (*cryptoutilCAServiceRevocation.OCSPService, error)
	newHandlerFn     func(*cryptoutilCAServiceIssuer.Issuer, cryptoutilCAStorage.Store, map[string]*cryptoutilCAHandler.ProfileConfig) (*cryptoutilCAHandler.Handler, error)
}

func NewPKIComponents(ctx context.Context, db *gorm.DB, barrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service) (*PKIComponents, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	if barrierService == nil {
		return nil, fmt.Errorf("barrier service is required")
	}

	certificateStore, err := cryptoutilCAStorage.NewGORMStore(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create PKI certificate store: %w", err)
	}

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create PKI authority store: %w", err)
	}

	cryptoProvider := cryptoutilCACrypto.NewSoftwareProvider()

	return newPKIComponents(pkiComponentFactories{
		provider:     cryptoProvider,
		storage:      certificateStore,
		randomReader: crand.Reader,
		loadOrCreateCAFn: func(provider cryptoutilCACrypto.Provider, randomReader io.Reader) (*x509.Certificate, crypto.PrivateKey, error) {
			return getOrCreateCA(
				ctx,
				provider,
				randomReader,
				authorityStore,
				barrierService.EncryptContentWithContext,
				barrierService.DecryptContentWithContext,
			)
		},
		newIssuerFn:      cryptoutilCAServiceIssuer.NewIssuer,
		newCRLServiceFn:  cryptoutilCAServiceRevocation.NewCRLService,
		newOCSPServiceFn: cryptoutilCAServiceRevocation.NewOCSPService,
		newHandlerFn:     cryptoutilCAHandler.NewHandler,
	})
}

func NewInMemoryPKIComponents() (*PKIComponents, error) {
	cryptoProvider := cryptoutilCACrypto.NewSoftwareProvider()

	return newPKIComponents(pkiComponentFactories{
		provider:         cryptoProvider,
		storage:          cryptoutilCAStorage.NewMemoryStore(),
		randomReader:     crand.Reader,
		loadOrCreateCAFn: createSelfSignedCA,
		newIssuerFn:      cryptoutilCAServiceIssuer.NewIssuer,
		newCRLServiceFn:  cryptoutilCAServiceRevocation.NewCRLService,
		newOCSPServiceFn: cryptoutilCAServiceRevocation.NewOCSPService,
		newHandlerFn:     cryptoutilCAHandler.NewHandler,
	})
}

func newPKIComponents(factories pkiComponentFactories) (*PKIComponents, error) {
	caCert, caKey, err := factories.loadOrCreateCAFn(factories.provider, factories.randomReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA: %w", err)
	}

	issuer, err := factories.newIssuerFn(factories.provider, &cryptoutilCAServiceIssuer.IssuingCAConfig{
		Name:        cryptoutilSharedMagic.OTLPServiceSMKMS,
		Certificate: caCert,
		PrivateKey:  caKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create issuer: %w", err)
	}

	caKeySigner, ok := caKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("CA private key is not a signer")
	}

	crlService, err := factories.newCRLServiceFn(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:           caCert,
		PrivateKey:       caKeySigner,
		Provider:         factories.provider,
		Validity:         cryptoutilSharedMagic.Days7,
		NextUpdateBuffer: cryptoutilSharedMagic.Days1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create CRL service: %w", err)
	}

	ocspService, err := factories.newOCSPServiceFn(&cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       caCert,
		Responder:    caCert,
		ResponderKey: caKeySigner,
		Provider:     factories.provider,
		Validity:     cryptoutilSharedMagic.Days1,
	}, crlService)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCSP service: %w", err)
	}

	profiles := map[string]*cryptoutilCAHandler.ProfileConfig{
		cryptoutilSharedMagic.PKIProfileIDTLSServer: {
			ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
			Name:        "TLS Server Certificate",
			Description: "Standard TLS server certificate for HTTPS",
			Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
		},
		cryptoutilSharedMagic.PKIProfileIDTLSClient: {
			ID:          cryptoutilSharedMagic.PKIProfileIDTLSClient,
			Name:        "TLS Client Certificate",
			Description: "TLS client authentication certificate",
			Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
		},
	}

	handler, err := factories.newHandlerFn(issuer, factories.storage, profiles)
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	handler.SetCRLService(crlService)
	handler.SetOCSPService(ocspService)

	return &PKIComponents{
		issuer:      issuer,
		storage:     factories.storage,
		crlService:  crlService,
		ocspService: ocspService,
		handler:     handler,
	}, nil
}

func createSelfSignedCA(provider cryptoutilCACrypto.Provider, randomReader io.Reader) (*x509.Certificate, crypto.PrivateKey, error) {
	if randomReader == nil {
		return nil, nil, fmt.Errorf("random reader is required")
	}

	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilSharedMagic.ECDSA,
		ECDSACurve: "P-384",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	serialNumber, err := crand.Int(randomReader, new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.PKICASerialNumberBits))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	if serialNumber.Sign() == 0 {
		serialNumber.SetInt64(1)
	}

	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "sm-kms PKI Development CA",
			Organization: []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault},
			Country:      []string{"US"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(cryptoutilSharedMagic.TLSDefaultMaxCACertDuration),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            defaultRootCAPathLength,
	}

	signer, ok := keyPair.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("private key is not a signer")
	}

	sigAlg, err := provider.GetSignatureAlgorithm(signer.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get signature algorithm: %w", err)
	}

	template.SignatureAlgorithm = sigAlg

	certDER, err := x509.CreateCertificate(randomReader, template, template, signer.Public(), signer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	return cert, keyPair.PrivateKey, nil
}
