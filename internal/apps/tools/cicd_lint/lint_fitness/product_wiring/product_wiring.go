// Copyright (c) 2025 Justin Cranford

// Package product_wiring validates cmd/ entry points exist for all products and services.
package product_wiring

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// productServicePair defines a product and its expected service.
type productServicePair struct {
	product string
	service string
}

// knownProducts lists all products that must have cmd/PRODUCT/main.go.
var knownProducts = []string{"sm", cryptoutilSharedMagic.JoseProductName, cryptoutilSharedMagic.PKIProductName, cryptoutilSharedMagic.IdentityProductName, cryptoutilSharedMagic.SkeletonProductName}

// knownServices lists all product-service pairs that must have cmd/PRODUCT-SERVICE/main.go.
var knownServices = []productServicePair{
	{product: "sm", service: cryptoutilSharedMagic.KMSServiceName},
	{product: "sm", service: "im"},
	{product: cryptoutilSharedMagic.JoseProductName, service: "ja"},
	{product: cryptoutilSharedMagic.PKIProductName, service: "ca"},
	{product: cryptoutilSharedMagic.IdentityProductName, service: cryptoutilSharedMagic.AuthzServiceName},
	{product: cryptoutilSharedMagic.IdentityProductName, service: cryptoutilSharedMagic.IDPServiceName},
	{product: cryptoutilSharedMagic.IdentityProductName, service: "rs"},
	{product: cryptoutilSharedMagic.IdentityProductName, service: "rp"},
	{product: cryptoutilSharedMagic.IdentityProductName, service: cryptoutilSharedMagic.SPAServiceName},
	{product: cryptoutilSharedMagic.SkeletonProductName, service: cryptoutilSharedMagic.SkeletonTemplateServiceName},
}

// Check validates product wiring from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates cmd/ entry points under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var errors []string

	cmdDir := filepath.Join(rootDir, "cmd")
	if _, statErr := os.Stat(cmdDir); os.IsNotExist(statErr) {
		return fmt.Errorf("cmd directory not found at %s", cmdDir)
	}

	// Check product binaries: cmd/PRODUCT/main.go.
	for _, product := range knownProducts {
		mainFile := filepath.Join(cmdDir, product, "main.go")
		if _, err := os.Stat(mainFile); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("cmd/%s/main.go: missing product entry point", product))
		}
	}

	// Check service binaries: cmd/PRODUCT-SERVICE/main.go.
	for _, pair := range knownServices {
		dirName := pair.product + "-" + pair.service
		mainFile := filepath.Join(cmdDir, dirName, "main.go")

		if _, err := os.Stat(mainFile); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("cmd/%s/main.go: missing service entry point", dirName))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("product wiring violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("product-wiring: all cmd/ entry points present")

	return nil
}
