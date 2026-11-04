// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"fmt"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
)

func writeDevelopmentSmokeProtected(directory *gen.DevelopmentPrivateDirectory, name string, data []byte) error {
	if err := directory.Write(name, data); err != nil {
		return fmt.Errorf("protected development output: %w", err)
	}

	return nil
}
