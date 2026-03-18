// Copyright (c) 2025 Justin Cranford

// Package require_api_dir validates that every registered service has a corresponding
// api/<product>-<service>/ directory with required OpenAPI generation files.
package require_api_dir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// apiServiceDef maps a service to its api/ directory name.
type apiServiceDef struct {
	Product string
	Service string
}

// knownAPIServices lists all product-service pairs that must have a corresponding api/ directory.
var knownAPIServices = []apiServiceDef{
	{Product: cryptoutilSharedMagic.IdentityProductName, Service: cryptoutilSharedMagic.AuthzServiceName},
	{Product: cryptoutilSharedMagic.IdentityProductName, Service: cryptoutilSharedMagic.IDPServiceName},
	{Product: cryptoutilSharedMagic.IdentityProductName, Service: cryptoutilSharedMagic.RPServiceName},
	{Product: cryptoutilSharedMagic.IdentityProductName, Service: cryptoutilSharedMagic.RSServiceName},
	{Product: cryptoutilSharedMagic.IdentityProductName, Service: cryptoutilSharedMagic.SPAServiceName},
	{Product: cryptoutilSharedMagic.JoseProductName, Service: cryptoutilSharedMagic.JoseJAServiceName},
	{Product: cryptoutilSharedMagic.PKIProductName, Service: cryptoutilSharedMagic.PKICAServiceName},
	{Product: cryptoutilSharedMagic.SkeletonProductName, Service: cryptoutilSharedMagic.SkeletonTemplateServiceName},
	{Product: cryptoutilSharedMagic.SMProductName, Service: cryptoutilSharedMagic.IMServiceName},
	{Product: cryptoutilSharedMagic.SMProductName, Service: cryptoutilSharedMagic.KMSServiceName},
}

// requiredFiles are files that every api/<product>-<service>/ directory must have.
var requiredFiles = []string{
	"generate.go",
}

// Check validates api/ directory presence from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates api/ directories under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking api/ directory presence for registered services...")

	apiDir := filepath.Join(rootDir, "api")
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		return fmt.Errorf("api/ directory not found at %s", apiDir)
	}

	var violations []string

	for _, svc := range knownAPIServices {
		apiName := svc.Product + "-" + svc.Service
		svcAPIDir := filepath.Join(apiDir, apiName)

		if errs := checkAPIDir(svcAPIDir, apiName); len(errs) > 0 {
			violations = append(violations, errs...)
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("require-api-dir violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("require-api-dir: all services have required api/ directories")

	return nil
}

// checkAPIDir validates a single api/<product>-<service>/ directory.
func checkAPIDir(svcAPIDir, apiName string) []string {
	var errors []string

	if _, err := os.Stat(svcAPIDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("api/%s: directory missing", apiName)}
	}

	for _, f := range requiredFiles {
		fullPath := filepath.Join(svcAPIDir, f)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("api/%s: missing required file %s", apiName, f))
		}
	}

	return errors
}
