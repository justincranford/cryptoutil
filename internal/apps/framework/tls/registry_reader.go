// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// realmEntry describes a single authentication realm for a PS-ID.
type realmEntry struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	Type     string `yaml:"type"`
}

// registryPSID is a minimal struct for parsing a PS-ID entry from registry.yaml.
type registryPSID struct {
	PSID   string       `yaml:"ps_id"`
	Realms []realmEntry `yaml:"realms"`
}

// registryDoc is the top-level registry.yaml structure (only fields needed by pki-init).
type registryDoc struct {
	ProductServices []registryPSID `yaml:"product_services"`
}

// defaultRegistryPath is the well-known location of registry.yaml within the module.
const defaultRegistryPath = "api/cryptosuite-registry/registry.yaml"

// defaultRealms returns the standard two-realm fallback when no overrides are present.
func defaultRealms() []string {
	return []string{"file", "db"}
}

// readRealmsForPSID reads the realm names for a given PS-ID from the registry file.
// Returns an error if the PS-ID is not found or has no realms configured.
func readRealmsForPSID(registryPath, psID string) ([]string, error) {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file %q: %w", registryPath, err)
	}

	var doc registryDoc

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse registry file %q: %w", registryPath, err)
	}

	for _, ps := range doc.ProductServices {
		if ps.PSID != psID {
			continue
		}

		if len(ps.Realms) == 0 {
			return nil, fmt.Errorf("PS-ID %q has no realms configured in registry.yaml", psID)
		}

		names := make([]string, len(ps.Realms))

		for i, r := range ps.Realms {
			names[i] = r.Name
		}

		return names, nil
	}

	return nil, fmt.Errorf("PS-ID %q not found in registry.yaml", psID)
}
