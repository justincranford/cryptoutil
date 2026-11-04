// Copyright (c) 2025-2026 Justin Cranford.
// Package cicd_checkov runs the pinned Checkov CLI with repository-safe local exclusions.
package cicd_checkov

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

const expectedCheckovVersion = "3.3.29"

var blockedCheckovConfigurationKeys = map[string]struct{}{
	"access_key":                     {},
	"api_key":                        {},
	"bc_api_key":                     {},
	"bc_api_url":                     {},
	"download_external_modules":      {},
	"external_checks_git":            {},
	"password":                       {},
	"prisma_api_url":                 {},
	"repo_id":                        {},
	"secret_key":                     {},
	checkovSupportArgument:           {},
	cryptoutilSharedMagic.ParamToken: {},
}

func canonicalRepositoryRoot(root string) (string, error) {
	if root == "" {
		root = cryptoutilSharedMagic.CurrentDirectory
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve Checkov repository root: %w", err)
	}

	canonicalRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", fmt.Errorf("resolve Checkov repository root symlinks: %w", err)
	}

	info, err := os.Stat(canonicalRoot)
	if err != nil {
		return "", fmt.Errorf("inspect Checkov repository root: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("checkov repository root is not a directory: %s", canonicalRoot)
	}

	return filepath.Clean(canonicalRoot), nil
}

func resolveCheckovPath(checkovPath string) (string, error) {
	if checkovPath == "" {
		resolvedPath, err := exec.LookPath("checkov")
		if err != nil {
			return "", fmt.Errorf("find pinned Checkov executable on PATH: %w", err)
		}

		return resolvedPath, nil
	}

	if filepath.IsAbs(checkovPath) || strings.ContainsAny(checkovPath, `/\`) {
		absolutePath, err := filepath.Abs(checkovPath)
		if err != nil {
			return "", fmt.Errorf("resolve Checkov executable path: %w", err)
		}

		return absolutePath, nil
	}

	resolvedPath, err := exec.LookPath(checkovPath)
	if err != nil {
		return "", fmt.Errorf("find Checkov executable %q on PATH: %w", checkovPath, err)
	}

	return resolvedPath, nil
}

func verifyPinnedCheckov(ctx context.Context, checkovPath, root string, environment []string) error {
	versionCommand, err := CommandContext(ctx, checkovPath, cryptoutilSharedMagic.CLIVersionFlag)
	if err != nil {
		return err
	}

	versionCommand.Dir = root
	versionCommand.Env = environment

	versionOutput, err := versionCommand.CombinedOutput()
	if err != nil {
		return fmt.Errorf("check pinned Checkov version: %w; output: %s", err, strings.TrimSpace(string(versionOutput)))
	}

	versionPattern := regexp.MustCompile(`(^|[^0-9.])` + regexp.QuoteMeta(expectedCheckovVersion) + `([^0-9.]|$)`)
	if !versionPattern.Match(versionOutput) {
		return fmt.Errorf(
			"checkov version mismatch: expected %s, got %q",
			expectedCheckovVersion,
			strings.TrimSpace(string(versionOutput)),
		)
	}

	return nil
}

// LocalCheckovEnvironment preserves only runtime essentials and isolates user-specific settings.
func LocalCheckovEnvironment(sourceEnvironment []string, homeDirectory string) []string {
	allowedVariables := map[string]struct{}{
		"COMSPEC":     {},
		"LANG":        {},
		"LC_ALL":      {},
		"LC_CTYPE":    {},
		"LC_MESSAGES": {},
		"PATH":        {},
		"PATHEXT":     {},
		"SYSTEMROOT":  {},
		"TEMP":        {},
		"TMP":         {},
		"TMPDIR":      {},
		"TZ":          {},
		"WINDIR":      {},
	}
	values := make(map[string]string, len(allowedVariables)+5)

	for _, variable := range sourceEnvironment {
		name, value, found := strings.Cut(variable, "=")
		if !found {
			continue
		}

		normalizedName := strings.ToUpper(name)
		if _, allowed := allowedVariables[normalizedName]; allowed {
			values[normalizedName] = value
		}
	}

	values["HOME"] = homeDirectory

	values["USERPROFILE"] = homeDirectory
	if runtime.GOOS == "windows" {
		homeDrive := filepath.VolumeName(homeDirectory)
		values["HOMEDRIVE"] = homeDrive
		values["HOMEPATH"] = strings.TrimPrefix(homeDirectory, homeDrive)
	}

	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}

	sort.Strings(names)

	environment := make([]string, 0, len(names))
	for _, name := range names {
		environment = append(environment, name+"="+values[name])
	}

	return environment
}

func rejectRemoteCheckovConfiguration(root string) error {
	configNames := []string{
		".checkov.yaml",
		".checkov.yml",
		"checkov.yaml",
		"checkov.yml",
	}

	for _, configName := range configNames {
		configPath := filepath.Join(root, configName)

		configBytes, err := os.ReadFile(configPath)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return fmt.Errorf("read Checkov configuration %s: %w", configName, err)
		}

		var config yaml.Node
		if err := yaml.Unmarshal(configBytes, &config); err != nil {
			return fmt.Errorf("parse Checkov configuration %s: %w", configName, err)
		}

		if key, found := findBlockedCheckovConfigurationKey(&config); found {
			return fmt.Errorf(
				"local-only Checkov refuses cloud credentials, remote policy sources, or reporting setting %q in %s",
				key,
				configName,
			)
		}
	}

	return nil
}

func findBlockedCheckovConfigurationKey(node *yaml.Node) (string, bool) {
	if node == nil {
		return "", false
	}

	if node.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(node.Content); index += 2 {
			key := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(node.Content[index].Value), "-", "_"))
			if _, blocked := blockedCheckovConfigurationKeys[key]; blocked {
				return key, true
			}
		}
	}

	for _, child := range node.Content {
		if nestedKey, found := findBlockedCheckovConfigurationKey(child); found {
			return nestedKey, true
		}
	}

	return "", false
}
