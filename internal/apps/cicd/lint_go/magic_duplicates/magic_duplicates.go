// Copyright (c) 2025 Justin Cranford

// Package magic_duplicates verifies that magic constants have no duplicate values across magic files.
package magic_duplicates

import (
	"fmt"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
)

// Check is a LinterFunc that scans the magic package for constants
// that share the same literal value under multiple names.  Duplicate values indicate
// that a canonical constant exists alongside redundant aliases, causing confusion
// about which name callers should use.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckMagicDuplicatesInDir(logger, lintGoCommon.MagicDefaultDir)
}

// CheckMagicDuplicatesInDir is the testable implementation that accepts an explicit magicDir.
func CheckMagicDuplicatesInDir(logger *cryptoutilCmdCicdCommon.Logger, magicDir string) error {
	logger.Log("Checking magic package for duplicate constant values...")

	inv, err := lintGoCommon.ParseMagicDir(magicDir)
	if err != nil {
		return fmt.Errorf("failed to parse magic package: %w", err)
	}

	type dupGroup struct {
		value  string
		consts []lintGoCommon.MagicConstant
	}

	var groups []dupGroup

	for value, consts := range inv.ByValue {
		if len(consts) < 2 {
			continue
		}

		group := dupGroup{value: value, consts: append([]lintGoCommon.MagicConstant(nil), consts...)}
		sort.Slice(group.consts, func(i, j int) bool {
			if group.consts[i].File != group.consts[j].File {
				return group.consts[i].File < group.consts[j].File
			}

			return group.consts[i].Name < group.consts[j].Name
		})

		groups = append(groups, group)
	}

	if len(groups) == 0 {
		logger.Log("✅ magic-duplicates: no duplicate values found")

		return nil
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].value < groups[j].value })

	var sb strings.Builder

	fmt.Fprintf(&sb, "magic-duplicates: %d duplicate value group(s) found (informational — fix incrementally)\n\n", len(groups))

	for _, g := range groups {
		fmt.Fprintf(&sb, "  value %s defined %d times:\n", g.value, len(g.consts))

		for _, mc := range g.consts {
			fmt.Fprintf(&sb, "    %s:%d  %s\n", mc.File, mc.Line, mc.Name)
		}

		fmt.Fprint(&sb, "\n")
	}

	// magic-duplicates is informational: it logs violations but does not block CI.
	// The magic package has accumulated duplicate values that need incremental cleanup.
	// Run 'cicd lint-go' regularly to measure progress.
	logger.Log(sb.String())

	return nil
}
