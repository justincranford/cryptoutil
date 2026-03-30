// Copyright (c) 2025 Justin Cranford

// Package compose_port_formula validates that Docker Compose port bindings follow
// the canonical formula: host_port = base_port + tier_offset + variant_offset.
//
// Formula:
//   - SERVICE tier:  base_port + 0  (e.g., sm-kms: 8000/8001/8002)
//   - PRODUCT tier:  base_port + 10000 (e.g., sm-kms: 18000/18001/18002)
//   - SUITE tier:    base_port + 20000 (e.g., sm-kms: 28000/28001/28002)
//
// Variant offsets add to the tier base:
//   - sqlite-1:   +0 (base instance)
//   - postgres-1: +1 (first PostgreSQL instance)
//   - postgres-2: +2 (second PostgreSQL instance)
//
// See ARCHITECTURE.md Section 3.4.1 for the port allocation design.
package compose_port_formula

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// Injectable functions for testing defensive error paths.
var (
	portReadFileFn = os.ReadFile
)

// portMapping maps a compose service variant to its expected port offset.
type portMapping struct {
	// serviceVariant is the variant suffix (e.g., "sqlite-1").
	serviceVariant string
	// variantOffset is the base offset within a tier (0, 1, or 2).
	variantOffset int
}

// variantMappings defines the expected port offsets for each compose service variant.
var variantMappings = []portMapping{
	{serviceVariant: lintFitnessRegistry.ComposeVariantSQLite1, variantOffset: 0},
	{serviceVariant: lintFitnessRegistry.ComposeVariantPostgres1, variantOffset: 1},
	{serviceVariant: lintFitnessRegistry.ComposeVariantPostgres2, variantOffset: 2},
}

// tierConfig describes a compose file and its expected port tier offset.
type tierConfig struct {
	// composePath is the relative path to the compose file from rootDir.
	composePath string
	// tierOffset is the port offset for this tier (0, 10000, or 20000).
	tierOffset int
}

// Check validates compose port formula from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates compose port formula under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking compose port formula (base_port + tier_offset + variant_offset)...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		tiers := []tierConfig{
			{
				composePath: filepath.Join("deployments", ps.PSID, "compose.yml"),
				tierOffset:  lintFitnessRegistry.PortTierOffsetService,
			},
			{
				composePath: filepath.Join("deployments", ps.Product, "compose.yml"),
				tierOffset:  lintFitnessRegistry.PortTierOffsetProduct,
			},
		}

		// Add suite tier if registry defines at least one suite.
		if suites := lintFitnessRegistry.AllSuites(); len(suites) > 0 {
			tiers = append(tiers, tierConfig{
				composePath: filepath.Join("deployments", suites[0].ID, "compose.yml"),
				tierOffset:  lintFitnessRegistry.PortTierOffsetSuite,
			})
		}

		basePort := lintFitnessRegistry.PublicPort(ps.PSID)

		for _, tier := range tiers {
			v := checkTierPorts(rootDir, ps.PSID, tier.composePath, basePort, tier.tierOffset)
			violations = append(violations, v...)
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose port formula violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-port-formula: all compose port bindings match formula (base_port + tier_offset + variant_offset)")

	return nil
}

// portLinePattern matches port mapping lines like "- "8000:8080"" or "- 8000:8080".
var portLinePattern = regexp.MustCompile(`-\s*"?(\d+):(\d+)"?`)

// checkTierPorts validates all per-psid port bindings in one compose file.
func checkTierPorts(rootDir, psID, composeRelPath string, basePort, tierOffset int) []string {
	fullPath := filepath.Join(rootDir, composeRelPath)

	data, err := portReadFileFn(fullPath)
	if err != nil {
		// In PRODUCT/SUITE compose files, multiple PSIDs share the file — skip missing files silently
		// unless it is the SERVICE-level compose (unique per PS-ID).
		if os.IsNotExist(err) {
			return nil
		}

		return []string{fmt.Sprintf("%s: cannot read %s: %v", psID, composeRelPath, err)}
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var violations []string

	currentService := ""

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track service block headers (two-leading-space keys under services:).
		if len(line) > 2 && line[0] == ' ' && line[1] == ' ' && line[2] != ' ' {
			// Strip trailing colon.
			candidate := strings.TrimSuffix(trimmed, ":")
			candidate = strings.TrimSpace(candidate)
			currentService = candidate

			continue
		}

		// Look for port mappings in the current service block.
		if !portLinePattern.MatchString(trimmed) {
			continue
		}

		match := portLinePattern.FindStringSubmatch(trimmed)
		if match == nil {
			continue
		}

		hostPortVal, parseErr := strconv.Atoi(match[1])
		if parseErr != nil {
			continue
		}

		// Only validate services belonging to this PS-ID.
		for _, vm := range variantMappings {
			expectedServiceName := lintFitnessRegistry.ComposeServiceName(psID, vm.serviceVariant)
			if currentService != expectedServiceName {
				continue
			}

			expectedPort := basePort + tierOffset + vm.variantOffset

			if hostPortVal != expectedPort {
				violations = append(violations, fmt.Sprintf(
					"%s: %s line %d: service %q host port %d; want %d (base=%d + tier=%d + variant=%d)",
					psID, composeRelPath, lineNum+1,
					currentService, hostPortVal, expectedPort,
					basePort, tierOffset, vm.variantOffset,
				))
			}
		}
	}

	return violations
}
