// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"fmt"
	"io"
	"math/big"
	"time"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilCAHandler "cryptoutil/internal/apps/sm-kms/server/pki/api/handler"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
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
	issuer                    *cryptoutilCAServiceIssuer.Issuer
	storage                   cryptoutilCAStorage.Store
	crlService                *cryptoutilCAServiceRevocation.CRLService
	ocspService               *cryptoutilCAServiceRevocation.OCSPService
	handler                   *cryptoutilCAHandler.Handler
	publicRevocationEndpoints map[string]publicRevocationEndpoint
}

type publicRevocationEndpoint struct {
	handler    *cryptoutilCAHandler.Handler
	issuerName string
}

// RegisterPublicRevocationEndpoints exposes standard revocation endpoints without requiring
// an application session or tenant context.
func (components *PKIComponents) RegisterPublicRevocationEndpoints(app *fiber.App) error {
	if components == nil || components.handler == nil {
		return fmt.Errorf("PKI handler is required")
	}

	if app == nil {
		return fmt.Errorf("fiber application is required")
	}

	app.Get("/pki/cas/:caID/crl", func(c *fiber.Ctx) error {
		endpoint, exists := components.publicRevocationEndpoints[c.Params("caID")]
		if !exists {
			return c.SendStatus(fiber.StatusNotFound)
		}

		var params cryptoutilApiCaServer.GetCRLParams

		if format := c.Query("format"); format != "" {
			formatValue := cryptoutilApiCaServer.GetCRLParamsFormat(format)
			params.Format = &formatValue
		}

		return endpoint.handler.GetCRL(c, endpoint.issuerName, params)
	})
	app.Post("/pki/ocsp", components.handler.HandleOCSP)
	app.Post("/pki/ocsp/:caID", func(c *fiber.Ctx) error {
		endpoint, exists := components.publicRevocationEndpoints[c.Params("caID")]
		if !exists {
			return c.SendStatus(fiber.StatusNotFound)
		}

		return endpoint.handler.HandleOCSP(c)
	})

	return nil
}

type pkiComponentFactories struct {
	provider              cryptoutilCACrypto.Provider
	storage               cryptoutilCAStorage.Store
	randomReader          io.Reader
	ctx                   context.Context
	authorityID           string
	crlURL                string
	ocspURL               string
	listAuthoritiesFn     func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error)
	decryptAuthorityKeyFn contentCipherFn
	loadOrCreateCAFn      func(cryptoutilCACrypto.Provider, io.Reader) (*x509.Certificate, crypto.PrivateKey, error)
	newIssuerFn           func(cryptoutilCACrypto.Provider, *cryptoutilCAServiceIssuer.IssuingCAConfig) (*cryptoutilCAServiceIssuer.Issuer, error)
	newCRLServiceFn       func(*cryptoutilCAServiceRevocation.CRLConfig) (*cryptoutilCAServiceRevocation.CRLService, error)
	newOCSPServiceFn      func(*cryptoutilCAServiceRevocation.OCSPConfig, *cryptoutilCAServiceRevocation.CRLService) (*cryptoutilCAServiceRevocation.OCSPService, error)
	newHandlerFn          func(*cryptoutilCAServiceIssuer.Issuer, cryptoutilCAStorage.Store, map[string]*cryptoutilCAHandler.ProfileConfig) (*cryptoutilCAHandler.Handler, error)
}

func NewPKIComponents(ctx context.Context, db *gorm.DB, barrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service) (*PKIComponents, error) {
	return NewPKIComponentsWithAuthorityConfig(ctx, db, barrierService, DefaultPKIAuthorityConfig())
}

// NewPKIComponentsWithAuthorityConfig constructs production PKI components using the
// supplied persisted authority identity and lifecycle constraints.
func NewPKIComponentsWithAuthorityConfig(
	ctx context.Context,
	db *gorm.DB,
	barrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service,
	authorityConfig PKIAuthorityConfig,
) (*PKIComponents, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	if barrierService == nil {
		return nil, fmt.Errorf("barrier service is required")
	}

	if err := authorityConfig.validate(); err != nil {
		return nil, fmt.Errorf("invalid PKI authority configuration: %w", err)
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
		provider:              cryptoProvider,
		storage:               certificateStore,
		randomReader:          crand.Reader,
		ctx:                   ctx,
		authorityID:           authorityConfig.ID,
		crlURL:                authorityConfig.CRLURL,
		ocspURL:               authorityConfig.OCSPURL,
		listAuthoritiesFn:     authorityStore.List,
		decryptAuthorityKeyFn: barrierService.DecryptContentWithContext,
		loadOrCreateCAFn: func(provider cryptoutilCACrypto.Provider, randomReader io.Reader) (*x509.Certificate, crypto.PrivateKey, error) {
			return getOrCreateCAWithConfig(
				ctx,
				provider,
				randomReader,
				authorityStore,
				barrierService.EncryptContentWithContext,
				barrierService.DecryptContentWithContext,
				authorityConfig,
			)
		},
		newIssuerFn:      cryptoutilCAServiceIssuer.NewIssuer,
		newCRLServiceFn:  cryptoutilCAServiceRevocation.NewCRLService,
		newOCSPServiceFn: cryptoutilCAServiceRevocation.NewOCSPService,
		newHandlerFn:     cryptoutilCAHandler.NewHandler,
	})
}

func NewInMemoryPKIComponents() (*PKIComponents, error) {
	return NewInMemoryPKIComponentsWithAuthorityConfig(DefaultPKIAuthorityConfig())
}

// NewInMemoryPKIComponentsWithAuthorityConfig constructs test or development components
// using the supplied ephemeral authority configuration.
func NewInMemoryPKIComponentsWithAuthorityConfig(authorityConfig PKIAuthorityConfig) (*PKIComponents, error) {
	if err := authorityConfig.validate(); err != nil {
		return nil, fmt.Errorf("invalid PKI authority configuration: %w", err)
	}

	cryptoProvider := cryptoutilCACrypto.NewSoftwareProvider()

	return newPKIComponents(pkiComponentFactories{
		provider:     cryptoProvider,
		storage:      cryptoutilCAStorage.NewMemoryStore(),
		randomReader: crand.Reader,
		ctx:          context.Background(),
		authorityID:  authorityConfig.ID,
		loadOrCreateCAFn: func(provider cryptoutilCACrypto.Provider, randomReader io.Reader) (*x509.Certificate, crypto.PrivateKey, error) {
			return createSelfSignedCAWithConfig(provider, randomReader, authorityConfig)
		},
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
		Name:                  cryptoutilSharedMagic.OTLPServiceSMKMS,
		AuthorityID:           factories.authorityID,
		Certificate:           caCert,
		PrivateKey:            caKey,
		CRLDistributionPoints: configuredDistributionURL(factories.crlURL),
		OCSPServer:            configuredDistributionURL(factories.ocspURL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create issuer: %w", err)
	}

	caKeySigner, ok := caKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("CA private key is not a signer")
	}

	crlService, err := factories.newCRLServiceFn(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:            caCert,
		PrivateKey:        caKeySigner,
		Provider:          factories.provider,
		Validity:          cryptoutilSharedMagic.Days7,
		NextUpdateBuffer:  cryptoutilSharedMagic.Days1,
		AllocateCRLNumber: factories.storage.AllocateCRLNumber,
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

	activeAuthorityID := factories.authorityID
	if activeAuthorityID == "" {
		activeAuthorityID = issuer.GetCAConfig().Name
	}

	publicRevocationEndpoints := make(map[string]publicRevocationEndpoint)

	activeEndpoint := publicRevocationEndpoint{handler: handler, issuerName: issuer.GetCAConfig().Name}
	if err := addPublicRevocationEndpoint(publicRevocationEndpoints, activeAuthorityID, activeEndpoint); err != nil {
		return nil, err
	}

	if err := addPublicRevocationEndpoint(publicRevocationEndpoints, issuer.GetCAConfig().Name, activeEndpoint); err != nil {
		return nil, err
	}

	if factories.listAuthoritiesFn != nil {
		componentCtx := factories.ctx
		if componentCtx == nil {
			componentCtx = context.Background()
		}

		authorities, err := factories.listAuthoritiesFn(componentCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to list persisted PKI authorities: %w", err)
		}

		for _, authority := range authorities {
			if authority == nil {
				return nil, fmt.Errorf("persisted PKI authority list contains a nil entry")
			}

			if authority.ID == activeAuthorityID {
				continue
			}

			historicalCertificate, historicalPrivateKey, err := restoreHistoricalCA(
				componentCtx,
				authority,
				factories.decryptAuthorityKeyFn,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to restore historical PKI authority %q: %w", authority.ID, err)
			}

			historicalIssuer, err := factories.newIssuerFn(factories.provider, &cryptoutilCAServiceIssuer.IssuingCAConfig{
				Name:        authority.ID,
				AuthorityID: authority.ID,
				Certificate: historicalCertificate,
				PrivateKey:  historicalPrivateKey,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create historical PKI issuer %q: %w", authority.ID, err)
			}

			historicalSigner, ok := historicalPrivateKey.(crypto.Signer)
			if !ok {
				return nil, fmt.Errorf("historical PKI authority %q private key is not a signer", authority.ID)
			}

			historicalCRLService, err := factories.newCRLServiceFn(&cryptoutilCAServiceRevocation.CRLConfig{
				Issuer:            historicalCertificate,
				PrivateKey:        historicalSigner,
				Provider:          factories.provider,
				Validity:          cryptoutilSharedMagic.Days7,
				NextUpdateBuffer:  cryptoutilSharedMagic.Days1,
				AllocateCRLNumber: factories.storage.AllocateCRLNumber,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create historical PKI CRL service %q: %w", authority.ID, err)
			}

			historicalOCSPService, err := factories.newOCSPServiceFn(&cryptoutilCAServiceRevocation.OCSPConfig{
				Issuer:       historicalCertificate,
				Responder:    historicalCertificate,
				ResponderKey: historicalSigner,
				Provider:     factories.provider,
				Validity:     cryptoutilSharedMagic.Days1,
			}, historicalCRLService)
			if err != nil {
				return nil, fmt.Errorf("failed to create historical PKI OCSP service %q: %w", authority.ID, err)
			}

			historicalHandler, err := factories.newHandlerFn(historicalIssuer, factories.storage, profiles)
			if err != nil {
				return nil, fmt.Errorf("failed to create historical PKI handler %q: %w", authority.ID, err)
			}

			historicalHandler.SetCRLService(historicalCRLService)
			historicalHandler.SetOCSPService(historicalOCSPService)

			if err := addPublicRevocationEndpoint(publicRevocationEndpoints, authority.ID, publicRevocationEndpoint{
				handler:    historicalHandler,
				issuerName: authority.ID,
			}); err != nil {
				return nil, err
			}
		}
	}

	return &PKIComponents{
		issuer:                    issuer,
		storage:                   factories.storage,
		crlService:                crlService,
		ocspService:               ocspService,
		handler:                   handler,
		publicRevocationEndpoints: publicRevocationEndpoints,
	}, nil
}

func configuredDistributionURL(value string) []string {
	if value == "" {
		return nil
	}

	return []string{value}
}

func addPublicRevocationEndpoint(
	endpoints map[string]publicRevocationEndpoint,
	authorityID string,
	endpoint publicRevocationEndpoint,
) error {
	if authorityID == "" {
		return fmt.Errorf("public revocation endpoint authority ID is required")
	}

	if existing, exists := endpoints[authorityID]; exists {
		if existing.handler == endpoint.handler {
			return nil
		}

		return fmt.Errorf("duplicate public revocation endpoint authority ID %q", authorityID)
	}

	endpoints[authorityID] = endpoint

	return nil
}

func createSelfSignedCA(provider cryptoutilCACrypto.Provider, randomReader io.Reader) (*x509.Certificate, crypto.PrivateKey, error) {
	return createSelfSignedCAWithConfig(provider, randomReader, DefaultPKIAuthorityConfig())
}

func createSelfSignedCAWithConfig(
	provider cryptoutilCACrypto.Provider,
	randomReader io.Reader,
	authorityConfig PKIAuthorityConfig,
) (*x509.Certificate, crypto.PrivateKey, error) {
	if provider == nil {
		return nil, nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorCryptoProviderRequired)
	}

	if randomReader == nil {
		return nil, nil, fmt.Errorf("random reader is required")
	}

	if err := authorityConfig.validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid PKI authority configuration: %w", err)
	}

	keyPair, err := provider.GenerateKeyPair(authorityConfig.KeySpec)
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
		SerialNumber:          serialNumber,
		Subject:               authorityConfig.Subject,
		NotBefore:             now,
		NotAfter:              now.Add(authorityConfig.Validity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            authorityConfig.MaxPathLength,
		MaxPathLenZero:        authorityConfig.MaxPathLength == 0,
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

	if err := validateAuthorityMaterial(cert, keyPair.PrivateKey, authorityConfig); err != nil {
		return nil, nil, fmt.Errorf("generated CA failed lifecycle validation: %w", err)
	}

	return cert, keyPair.PrivateKey, nil
}
