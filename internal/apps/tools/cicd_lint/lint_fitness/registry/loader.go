// Copyright (c) 2025 Justin Cranford

package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultRegistryRelPath = "api/cryptosuite-registry/registry.yaml"
	// domainMigrationRangeMin is the minimum allowed migration_range_start for domain
	// migrations. Framework template migrations occupy 1001-1999; domain starts at 2001.
	domainMigrationRangeMin = 2001
)

// LoadRegistry reads the entity registry YAML file at path, unmarshals it, and runs
// structural validation. Returns a validated *RegistryFile or a descriptive error.
func LoadRegistry(path string) (*RegistryFile, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is controlled by the tooling framework
	if err != nil {
		return nil, fmt.Errorf("read registry %s: %w", path, err)
	}

	var r RegistryFile
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse registry %s: %w", path, err)
	}

	if err := validateRegistry(&r); err != nil {
		return nil, fmt.Errorf("validate registry %s: %w", path, err)
	}

	return &r, nil
}

// findRegistryYAMLPath walks up from the current working directory looking for the canonical
// registry YAML file at "api/cryptosuite-registry/registry.yaml". Returns the absolute
// path on success or an error if the file is not found.
func findRegistryYAMLPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	for {
		candidate := filepath.Join(dir, defaultRegistryRelPath)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Filesystem root reached without finding the file.
			return "", fmt.Errorf("registry YAML not found: walked up from working directory without finding %s", defaultRegistryRelPath)
		}

		dir = parent
	}
}

// validateRegistry enforces structural invariants on the registry.
// All violations are gathered and returned in a single error.
func validateRegistry(r *RegistryFile) error {
	var errs []string //nolint:prealloc // size determined by validation

	errs = append(errs, validateSuites(r.Suites)...)
	errs = append(errs, validateProducts(r.Products)...)
	errs = append(errs, validateProductServices(r.ProductServices)...)

	if len(errs) > 0 {
		return fmt.Errorf("registry validation errors:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

func validateSuites(suites []RegistrySuite) []string {
	var errs []string

	if len(suites) == 0 {
		errs = append(errs, "suites: must contain at least one entry")
	}

	seen := make(map[string]bool, len(suites))

	for i, s := range suites {
		prefix := fmt.Sprintf("suites[%d]", i)

		switch {
		case s.ID == "":
			errs = append(errs, prefix+".id: must not be empty")
		case seen[s.ID]:
			errs = append(errs, prefix+".id: duplicate id "+s.ID)
		default:
			seen[s.ID] = true
		}

		if s.DisplayName == "" {
			errs = append(errs, prefix+".display_name: must not be empty")
		}

		if s.CmdDir == "" {
			errs = append(errs, prefix+".cmd_dir: must not be empty")
		} else if !strings.HasSuffix(s.CmdDir, "/") {
			errs = append(errs, prefix+".cmd_dir: must end with '/' (got "+s.CmdDir+")")
		}
	}

	return errs
}

func validateProducts(products []RegistryProduct) []string {
	var errs []string

	if len(products) == 0 {
		errs = append(errs, "products: must contain at least one entry")
	}

	seen := make(map[string]bool, len(products))

	for i, p := range products {
		prefix := fmt.Sprintf("products[%d]", i)

		switch {
		case p.ID == "":
			errs = append(errs, prefix+".id: must not be empty")
		case seen[p.ID]:
			errs = append(errs, prefix+".id: duplicate id "+p.ID)
		default:
			seen[p.ID] = true
		}

		if p.DisplayName == "" {
			errs = append(errs, prefix+".display_name: must not be empty")
		}

		if p.InternalAppsDir == "" {
			errs = append(errs, prefix+".internal_apps_dir: must not be empty")
		} else if !strings.HasSuffix(p.InternalAppsDir, "/") {
			errs = append(errs, prefix+".internal_apps_dir: must end with '/' (got "+p.InternalAppsDir+")")
		}

		if p.CmdDir == "" {
			errs = append(errs, prefix+".cmd_dir: must not be empty")
		} else if !strings.HasSuffix(p.CmdDir, "/") {
			errs = append(errs, prefix+".cmd_dir: must end with '/' (got "+p.CmdDir+")")
		}
	}

	return errs
}

func validateProductServices(pss []RegistryProductService) []string {
	var errs []string

	if len(pss) == 0 {
		errs = append(errs, "product_services: must contain at least one entry")
	}

	seenPSID := make(map[string]bool, len(pss))
	seenBasePort := make(map[int]string, len(pss))
	seenPGPort := make(map[int]string, len(pss))

	// Store validated ranges for overlap detection.
	type rangeEntry struct {
		psid  string
		start int
		end   int
	}

	validatedRanges := make([]rangeEntry, 0, len(pss))

	for i := range pss {
		ps := &pss[i]
		prefix := fmt.Sprintf("product_services[%d] (%s)", i, ps.PSID)

		// Required string fields.
		if ps.PSID == "" {
			errs = append(errs, prefix+".ps_id: must not be empty")
		}

		if ps.Product == "" {
			errs = append(errs, prefix+".product: must not be empty")
		}

		if ps.Service == "" {
			errs = append(errs, prefix+".service: must not be empty")
		}

		if ps.DisplayName == "" {
			errs = append(errs, prefix+".display_name: must not be empty")
		}

		if ps.InternalAppsDir == "" {
			errs = append(errs, prefix+".internal_apps_dir: must not be empty")
		} else if !strings.HasSuffix(ps.InternalAppsDir, "/") {
			errs = append(errs, prefix+".internal_apps_dir: must end with '/' (got "+ps.InternalAppsDir+")")
		}

		if ps.MagicFile == "" {
			errs = append(errs, prefix+".magic_file: must not be empty")
		}

		// PSID == product + "-" + service invariant.
		if ps.PSID != "" && ps.Product != "" && ps.Service != "" {
			expected := ps.Product + "-" + ps.Service
			if ps.PSID != expected {
				errs = append(errs, fmt.Sprintf("%s.ps_id: expected %q (product-service) but got %q", prefix, expected, ps.PSID))
			}
		}

		// InternalAppsDir == psid + "/" invariant.
		if ps.PSID != "" && ps.InternalAppsDir != "" {
			expected := ps.PSID + "/"
			if ps.InternalAppsDir != expected {
				errs = append(errs, fmt.Sprintf("%s.internal_apps_dir: expected %q but got %q", prefix, expected, ps.InternalAppsDir))
			}
		}

		if ps.PSID != "" {
			if seenPSID[ps.PSID] {
				errs = append(errs, prefix+".ps_id: duplicate ps_id "+ps.PSID)
			} else {
				seenPSID[ps.PSID] = true
			}
		}

		// Ports must be positive and unique.
		if ps.BasePort <= 0 {
			errs = append(errs, fmt.Sprintf("%s.base_port: must be positive (got %d)", prefix, ps.BasePort))
		} else if existing, conflict := seenBasePort[ps.BasePort]; conflict {
			errs = append(errs, fmt.Sprintf("%s.base_port: %d already assigned to %s", prefix, ps.BasePort, existing))
		} else {
			seenBasePort[ps.BasePort] = ps.PSID
		}

		if ps.PGHostPort <= 0 {
			errs = append(errs, fmt.Sprintf("%s.pg_host_port: must be positive (got %d)", prefix, ps.PGHostPort))
		} else if existing, conflict := seenPGPort[ps.PGHostPort]; conflict {
			errs = append(errs, fmt.Sprintf("%s.pg_host_port: %d already assigned to %s", prefix, ps.PGHostPort, existing))
		} else {
			seenPGPort[ps.PGHostPort] = ps.PSID
		}

		// Migration range: [start, end] where start >= domainMigrationRangeMin and end > start.
		rangeValid := true

		if ps.MigrationRangeStart < domainMigrationRangeMin {
			errs = append(errs, fmt.Sprintf("%s.migration_range_start: must be ≥ %d (got %d)", prefix, domainMigrationRangeMin, ps.MigrationRangeStart))
			rangeValid = false
		}

		if ps.MigrationRangeEnd <= ps.MigrationRangeStart {
			errs = append(errs, fmt.Sprintf("%s.migration_range_end: must be > migration_range_start %d (got %d)", prefix, ps.MigrationRangeStart, ps.MigrationRangeEnd))
			rangeValid = false
		}

		if rangeValid && ps.PSID != "" {
			// Check for overlap with already-validated ranges.
			for _, existing := range validatedRanges {
				if ps.MigrationRangeStart <= existing.end && existing.start <= ps.MigrationRangeEnd {
					errs = append(errs, fmt.Sprintf("%s.migration_range [%d,%d] overlaps with %s [%d,%d]",
						prefix, ps.MigrationRangeStart, ps.MigrationRangeEnd,
						existing.psid, existing.start, existing.end))
				}
			}

			validatedRanges = append(validatedRanges, rangeEntry{
				psid:  ps.PSID,
				start: ps.MigrationRangeStart,
				end:   ps.MigrationRangeEnd,
			})
		}
	}

	return errs
}

// ToSuites converts the YAML suites to the Go registry Suite type.
func (r *RegistryFile) ToSuites() []Suite {
	suites := make([]Suite, len(r.Suites))

	for i, s := range r.Suites {
		suites[i] = Suite(s)
	}

	return suites
}

// ToProducts converts the YAML products to the Go registry Product type.
func (r *RegistryFile) ToProducts() []Product {
	products := make([]Product, len(r.Products))

	for i, p := range r.Products {
		products[i] = Product(p)
	}

	return products
}

// ToProductServices converts the YAML product-services to the Go registry ProductService type.
func (r *RegistryFile) ToProductServices() []ProductService {
	pss := make([]ProductService, len(r.ProductServices))

	for i, ps := range r.ProductServices {
		pss[i] = ProductService{
			PSID:            ps.PSID,
			Product:         ps.Product,
			Service:         ps.Service,
			DisplayName:     ps.DisplayName,
			InternalAppsDir: ps.InternalAppsDir,
			MagicFile:       ps.MagicFile,
		}
	}

	return pss
}
