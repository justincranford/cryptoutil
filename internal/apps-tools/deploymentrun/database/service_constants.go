// Copyright (c) 2025-2026 Justin Cranford.

package database

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
	psidAuthz    = cryptoutilSharedMagic.OTLPServiceIdentityAuthz
	psidIDP      = cryptoutilSharedMagic.OTLPServiceIdentityIDP
	psidRP       = cryptoutilSharedMagic.OTLPServiceIdentityRP
	psidRS       = cryptoutilSharedMagic.OTLPServiceIdentityRS
	psidSPA      = cryptoutilSharedMagic.OTLPServiceIdentitySPA
	psidSkeleton = cryptoutilSharedMagic.OTLPServiceSkeletonTemplate
	psidKMS      = cryptoutilSharedMagic.IMJWTIssuer
)
