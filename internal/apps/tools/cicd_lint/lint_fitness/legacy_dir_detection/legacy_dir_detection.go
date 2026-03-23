// Copyright (c) 2025 Justin Cranford

// Package legacy_dir_detection enforces that legacy product/service directories
// do not exist in the repository. This prevents the old "Cipher" product name
// from persisting as obsolete directories after the rename to SM.
package legacy_dir_detection

import (
"fmt"
"os"
"path/filepath"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// legacyDirs is the list of directory paths (relative to project root) that
// must NOT exist. Each entry is checked for existence on every run.
var legacyDirs = []string{
filepath.Join("internal", "apps", "cipher"),
}

// legacyPrefixes is the list of directory name prefixes that must NOT appear
// under the specified scan directories.
var legacyPrefixes = []struct {
scanDir string
prefix  string
}{
{scanDir: "deployments", prefix: "cipher-"},
		{scanDir: cryptoutilSharedMagic.CICDConfigsDir, prefix: "cipher-"},
{scanDir: "cmd", prefix: "cipher-"},
}

// Check runs the legacy-dir-detection check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir checks rootDir for legacy directory presence.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking for legacy directory presence...")

violations, err := FindViolationsInDir(rootDir)
if err != nil {
return fmt.Errorf("failed to check legacy directories: %w", err)
}

if len(violations) > 0 {
for _, v := range violations {
fmt.Fprintf(os.Stderr, "  legacy directory found: %s\n", v)
}

return fmt.Errorf("found %d legacy directories that must be removed", len(violations))
}

logger.Log("legacy-dir-detection: no legacy directories found")

return nil
}

// FindViolationsInDir checks rootDir for legacy directories and returns paths of any found.
func FindViolationsInDir(rootDir string) ([]string, error) {
var violations []string

for _, relPath := range legacyDirs {
fullPath := filepath.Join(rootDir, relPath)

if _, err := os.Stat(fullPath); err == nil {
violations = append(violations, fullPath)
}
}

for _, entry := range legacyPrefixes {
scanPath := filepath.Join(rootDir, entry.scanDir)

if _, err := os.Stat(scanPath); os.IsNotExist(err) {
continue
}

entries, err := os.ReadDir(scanPath)
if err != nil {
return nil, fmt.Errorf("failed to read directory %s: %w", scanPath, err)
}

for _, e := range entries {
if e.IsDir() && hasPrefix(e.Name(), entry.prefix) {
violations = append(violations, filepath.Join(scanPath, e.Name()))
}
}
}

return violations, nil
}

// hasPrefix reports whether s starts with prefix.
func hasPrefix(s, prefix string) bool {
return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
