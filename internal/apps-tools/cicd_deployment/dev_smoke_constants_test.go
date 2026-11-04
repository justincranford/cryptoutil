// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
	testDevelopmentHelp             = cryptoutilSharedMagic.CLIHelpFlag
	testDevelopmentActualIndex      = "actual index"
	testDevelopmentForgedIndex      = "forged index"
	testDevelopmentAbsentIndex      = "absent index"
	testDevelopmentInvalidReference = "invalid reference"
	testDevelopmentOwnedContext     = "owned context"
	testDevelopmentInvalidYAML      = "invalid yaml"
	testDevelopmentWrongContext     = "incorrect kubecontext"
	testDevelopmentAbsentClusters   = "absent clusters"
	testDevelopmentWrongServer      = "incorrect Kubernetes server"
	testDevelopmentAbsentTrust      = "absent trust"
	testDevelopmentAuthComplete     = "complete authentication"
	testDevelopmentAuthRegistration = "registration failure"
	testDevelopmentAuthCredentials  = "missing credentials"
	testDevelopmentAuthPassword     = "bad-password accepted"
	testDevelopmentAuthToken        = "missing token"
	testDevelopmentAuthDenial       = "authenticated denial"
	testDevelopmentTimeout          = "--timeout=1m"
	testDevelopmentClusterField     = "cluster"
)
