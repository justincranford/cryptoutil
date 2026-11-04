// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

const lifecycleVerifyCommand = "pki-verify"

func validatePKIVerificationCommand(job LifecycleJobInput, sources []LifecycleSecretSource) error {
	if len(job.Args) != 2 || !strings.HasPrefix(job.Args[1], "--input=") {
		return fmt.Errorf("verify-only PKI requires exact protected input argument")
	}

	file := strings.TrimPrefix(job.Args[1], "--input=")

	parts := strings.Split(strings.TrimPrefix(file, "/run/cryptoutil/secrets/"), "/")
	if !strings.HasPrefix(file, "/run/cryptoutil/secrets/") || path.Clean(file) != file || len(parts) != 2 || parts[1] != "pki-verification.json" ||
		!slices.Contains(job.SecretNames, parts[0]) {
		return fmt.Errorf("verify-only PKI manifest requires an owned restricted Secret mount")
	}

	for _, source := range sources {
		if source.Namespace == job.Namespace && source.Name == parts[0] && source.Files[parts[1]] != "" {
			return nil
		}
	}

	return fmt.Errorf("verify-only PKI manifest source missing")
}
