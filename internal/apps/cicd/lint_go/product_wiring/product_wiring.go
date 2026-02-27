// Copyright (c) 2025 Justin Cranford

// Package product_wiring validates cmd/ entry points exist for all products and services.
package product_wiring

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// productServicePair defines a product and its expected service.
type productServicePair struct {
	product string
	service string
}

// knownProducts lists all products that must have cmd/PRODUCT/main.go.
var knownProducts = []string{"identity", "jose", "pki", "skeleton", "sm"}

// knownServices lists all product-service pairs that must have cmd/PRODUCT-SERVICE/main.go.
var knownServices = []productServicePair{
	{product: "skeleton", service: "template"},
	{product: "pki", service: "ca"},
	{product: "jose", service: "ja"},
	{product: "sm", service: "im"},
	{product: "sm", service: "kms"},
	{product: "identity", service: "authz"},
	{product: "identity", service: "idp"},
	{product: "identity", service: "rp"},
	{product: "identity", service: "rs"},
	{product: "identity", service: "spa"},
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
