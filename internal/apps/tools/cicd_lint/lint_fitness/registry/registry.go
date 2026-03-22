// Copyright (c) 2025 Justin Cranford

// Package registry defines the canonical entity registry for all cryptoutil products,
// product-services, and suites. This is the single source of truth for structural
// conventions. Fitness linters use this registry to detect drift.
package registry

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Product represents a cryptoutil product (e.g. "sm", "jose").
type Product struct {
	// ID is the canonical product identifier (e.g. "sm").
	ID string
	// DisplayName is the human-readable name (e.g. "Secrets Manager").
	DisplayName string
	// InternalAppsDir is the path under internal/apps/ (e.g. "sm/").
	InternalAppsDir string
	// CmdDir is the sub-path under cmd/ (e.g. "sm/").
	CmdDir string
}

// ProductService represents a product-service pair (e.g. "sm-kms").
type ProductService struct {
	// PSID is the canonical PS identifier (e.g. "sm-kms").
	PSID string
	// Product is the product name component (e.g. "sm").
	Product string
	// Service is the service name component (e.g. "kms").
	Service string
	// DisplayName is the human-readable name (e.g. "Secrets Manager Key Management").
	DisplayName string
	// InternalAppsDir is the path under internal/apps/ (e.g. "sm/kms/").
	InternalAppsDir string
	// MagicFile is the filename of the primary magic constants file
	// under internal/shared/magic/ (e.g. "magic_sm.go").
	MagicFile string
}

// Suite represents the cryptoutil top-level suite deployment.
type Suite struct {
	// ID is the canonical suite identifier (e.g. "cryptoutil").
	ID string
	// DisplayName is the human-readable name.
	DisplayName string
	// CmdDir is the sub-path under cmd/ (e.g. "cryptoutil/").
	CmdDir string
}

// allProducts is the canonical registry of all 5 cryptoutil products.
var allProducts = []Product{
	{
		ID:              cryptoutilSharedMagic.IdentityProductName,
		DisplayName:     "Identity",
		InternalAppsDir: cryptoutilSharedMagic.IdentityProductName + "/",
		CmdDir:          cryptoutilSharedMagic.IdentityProductName + "/",
	},
	{
		ID:              cryptoutilSharedMagic.JoseProductName,
		DisplayName:     "JOSE",
		InternalAppsDir: cryptoutilSharedMagic.JoseProductName + "/",
		CmdDir:          cryptoutilSharedMagic.JoseProductName + "/",
	},
	{
		ID:              cryptoutilSharedMagic.PKIProductName,
		DisplayName:     "PKI",
		InternalAppsDir: cryptoutilSharedMagic.PKIProductName + "/",
		CmdDir:          cryptoutilSharedMagic.PKIProductName + "/",
	},
	{
		ID:              cryptoutilSharedMagic.SkeletonProductName,
		DisplayName:     cryptoutilSharedMagic.SkeletonProductNameTitleCase,
		InternalAppsDir: cryptoutilSharedMagic.SkeletonProductName + "/",
		CmdDir:          cryptoutilSharedMagic.SkeletonProductName + "/",
	},
	{
		ID:              cryptoutilSharedMagic.SMProductName,
		DisplayName:     "Secrets Manager",
		InternalAppsDir: cryptoutilSharedMagic.SMProductName + "/",
		CmdDir:          cryptoutilSharedMagic.SMProductName + "/",
	},
}

// allProductServices is the canonical registry of all 10 cryptoutil product-services.
var allProductServices = []ProductService{
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		Product:         cryptoutilSharedMagic.IdentityProductName,
		Service:         cryptoutilSharedMagic.AuthzServiceName,
		DisplayName:     "Identity Authorization Server",
		InternalAppsDir: cryptoutilSharedMagic.IdentityProductName + "/" + cryptoutilSharedMagic.AuthzServiceName + "/",
		MagicFile:       "magic_identity.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		Product:         cryptoutilSharedMagic.IdentityProductName,
		Service:         cryptoutilSharedMagic.IDPServiceName,
		DisplayName:     "Identity Provider",
		InternalAppsDir: cryptoutilSharedMagic.IdentityProductName + "/" + cryptoutilSharedMagic.IDPServiceName + "/",
		MagicFile:       "magic_identity.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceIdentityRP,
		Product:         cryptoutilSharedMagic.IdentityProductName,
		Service:         cryptoutilSharedMagic.RPServiceName,
		DisplayName:     "Identity Relying Party",
		InternalAppsDir: cryptoutilSharedMagic.IdentityProductName + "/" + cryptoutilSharedMagic.RPServiceName + "/",
		MagicFile:       "magic_identity.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceIdentityRS,
		Product:         cryptoutilSharedMagic.IdentityProductName,
		Service:         cryptoutilSharedMagic.RSServiceName,
		DisplayName:     "Identity Resource Server",
		InternalAppsDir: cryptoutilSharedMagic.IdentityProductName + "/" + cryptoutilSharedMagic.RSServiceName + "/",
		MagicFile:       "magic_identity.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		Product:         cryptoutilSharedMagic.IdentityProductName,
		Service:         cryptoutilSharedMagic.SPAServiceName,
		DisplayName:     "Identity Single Page App",
		InternalAppsDir: cryptoutilSharedMagic.IdentityProductName + "/" + cryptoutilSharedMagic.SPAServiceName + "/",
		MagicFile:       "magic_identity.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceJoseJA,
		Product:         cryptoutilSharedMagic.JoseProductName,
		Service:         cryptoutilSharedMagic.JoseJAServiceName,
		DisplayName:     "JOSE JWK Authority",
		InternalAppsDir: cryptoutilSharedMagic.JoseProductName + "/" + cryptoutilSharedMagic.JoseJAServiceName + "/",
		MagicFile:       "magic_jose.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServicePKICA,
		Product:         cryptoutilSharedMagic.PKIProductName,
		Service:         cryptoutilSharedMagic.PKICAServiceName,
		DisplayName:     "PKI Certificate Authority",
		InternalAppsDir: cryptoutilSharedMagic.PKIProductName + "/" + cryptoutilSharedMagic.PKICAServiceName + "/",
		MagicFile:       "magic_pki.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
		Product:         cryptoutilSharedMagic.SkeletonProductName,
		Service:         cryptoutilSharedMagic.SkeletonTemplateServiceName,
		DisplayName:     "Skeleton Template",
		InternalAppsDir: cryptoutilSharedMagic.SkeletonProductName + "/" + cryptoutilSharedMagic.SkeletonTemplateServiceName + "/",
		MagicFile:       "magic_skeleton.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceSMIM,
		Product:         cryptoutilSharedMagic.SMProductName,
		Service:         cryptoutilSharedMagic.IMServiceName,
		DisplayName:     "Secrets Manager Instant Messenger",
		InternalAppsDir: cryptoutilSharedMagic.SMProductName + "/" + cryptoutilSharedMagic.IMServiceName + "/",
		MagicFile:       "magic_sm_im.go",
	},
	{
		PSID:            cryptoutilSharedMagic.OTLPServiceSMKMS,
		Product:         cryptoutilSharedMagic.SMProductName,
		Service:         cryptoutilSharedMagic.KMSServiceName,
		DisplayName:     "Secrets Manager Key Management",
		InternalAppsDir: cryptoutilSharedMagic.SMProductName + "/" + cryptoutilSharedMagic.KMSServiceName + "/",
		MagicFile:       "magic_sm.go",
	},
}

// allSuites is the canonical registry of the cryptoutil suite.
var allSuites = []Suite{
	{
		ID:          cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		DisplayName: cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		CmdDir:      cryptoutilSharedMagic.DefaultOTLPServiceDefault + "/",
	},
}

// AllProducts returns the canonical list of all 5 products.
func AllProducts() []Product {
	result := make([]Product, len(allProducts))
	copy(result, allProducts)

	return result
}

// AllProductServices returns the canonical list of all 10 product-services.
func AllProductServices() []ProductService {
	result := make([]ProductService, len(allProductServices))
	copy(result, allProductServices)

	return result
}

// AllSuites returns the canonical list of all suites (currently 1).
func AllSuites() []Suite {
	result := make([]Suite, len(allSuites))
	copy(result, allSuites)

	return result
}
