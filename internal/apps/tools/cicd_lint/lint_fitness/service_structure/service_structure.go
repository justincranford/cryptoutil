// Copyright (c) 2025 Justin Cranford

// Package service_structure validates service directory conventions.
package service_structure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ServiceDef defines a service and its required files.
type ServiceDef struct {
	PSID     string
	Service  string
	Required []string // Optional override; nil means use defaultRequiredFiles.
}

// knownServices lists all product-service pairs that must follow structural conventions.
// identity/authz and identity/idp are intentionally excluded (legacy, don't follow service template pattern).
var knownServices = []ServiceDef{
	{PSID: cryptoutilSharedMagic.OTLPServiceSMKMS, Service: cryptoutilSharedMagic.KMSServiceName, Required: kmsRequiredFiles},
	{PSID: cryptoutilSharedMagic.OTLPServiceSMIM, Service: cryptoutilSharedMagic.IMServiceName},
	{PSID: cryptoutilSharedMagic.OTLPServiceJoseJA, Service: cryptoutilSharedMagic.JoseJAServiceName},
	{PSID: cryptoutilSharedMagic.OTLPServicePKICA, Service: cryptoutilSharedMagic.PKICAServiceName},
	{PSID: cryptoutilSharedMagic.OTLPServiceIdentityRS, Service: cryptoutilSharedMagic.RSServiceName},
	{PSID: cryptoutilSharedMagic.OTLPServiceIdentityRP, Service: cryptoutilSharedMagic.RPServiceName},
	{PSID: cryptoutilSharedMagic.OTLPServiceIdentitySPA, Service: cryptoutilSharedMagic.SPAServiceName},
	{PSID: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, Service: cryptoutilSharedMagic.SkeletonTemplateServiceName},
}

// defaultRequiredFiles are files that every service must have (relative to service dir).
var defaultRequiredFiles = []string{
	"{SERVICE}.go",
	"{SERVICE}_usage.go",
	"server/server.go",
	"server/config/config.go",
}

// kmsRequiredFiles are files required for sm-kms (legacy service, no server/config/config.go).
var kmsRequiredFiles = []string{
	"{SERVICE}.go",
	"{SERVICE}_usage.go",
	"server/server.go",
}

// Check validates service structure from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates service directories under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var errors []string

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	for _, svc := range knownServices {
		serviceDir := filepath.Join(appsDir, svc.PSID)

		required := svc.Required
		if required == nil {
			required = defaultRequiredFiles
		}

		if errs := checkServiceDir(serviceDir, svc.Service, required); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("service structure violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("service-structure: all services pass structural validation")

	return nil
}

// checkServiceDir validates a single service directory for required files.
func checkServiceDir(serviceDir, service string, required []string) []string {
	var errors []string

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: service directory missing", serviceDir)}
	}

	for _, tmpl := range required {
		relPath := strings.ReplaceAll(tmpl, "{SERVICE}", service)
		fullPath := filepath.Join(serviceDir, relPath)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("%s: missing required file %s", serviceDir, relPath))
		}
	}

	return errors
}
