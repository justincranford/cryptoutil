// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

const (
	canonicalChartSourceRoot         = "api/cryptosuite-registry/templates/kubernetes"
	chartTemplatesFormat             = "charts"
	unexpectedChartMirrorFileMessage = "unexpected generated chart mirror file"
)

func canonicalChartArtifacts(source fs.FS) ([]generatedArtifact, error) {
	charts := []string{"cryptoutil-infra", "cryptoutil-runtime", "cryptoutil-lifecycle"}
	artifacts := make([]generatedArtifact, 0)

	for _, chart := range charts {
		sourceRoot := path.Join(canonicalChartSourceRoot, chart)

		walkErr := fs.WalkDir(source, sourceRoot, func(sourcePath string, entry fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("walk canonical chart source %s: %w", sourcePath, err)
			}

			if entry.IsDir() {
				if entry.Name() == "fixtures" {
					return fs.SkipDir
				}

				return nil
			}

			if entry.Type()&fs.ModeSymlink != 0 || !entry.Type().IsRegular() {
				return fmt.Errorf("canonical chart source must be a regular file: %s", sourcePath)
			}

			data, err := fs.ReadFile(source, sourcePath)
			if err != nil {
				return fmt.Errorf("read canonical chart source %s: %w", sourcePath, err)
			}

			relativePath := strings.TrimPrefix(sourcePath, sourceRoot+"/")
			targetPath := path.Join("deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "charts", chart, relativePath)
			artifacts = append(artifacts, generatedArtifact{path: targetPath, data: data})

			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("walk canonical chart source %s: %w", sourceRoot, walkErr)
		}
	}

	return artifacts, nil
}

func validateCanonicalChartMirror(source fs.FS, artifacts []generatedArtifact) error {
	expected := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		expected[artifact.path] = struct{}{}
	}

	const mirrorRoot = "deployments/k8s/charts"

	walkErr := fs.WalkDir(source, mirrorRoot, func(mirrorPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk generated chart mirror %s: %w", mirrorPath, err)
		}

		if entry.IsDir() {
			return nil
		}

		if entry.Type()&fs.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fmt.Errorf("generated chart mirror must contain only regular files: %s", mirrorPath)
		}

		if _, ok := expected[mirrorPath]; !ok {
			return fmt.Errorf("%s: %s", unexpectedChartMirrorFileMessage, mirrorPath)
		}

		delete(expected, mirrorPath)

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("walk generated chart mirror %s: %w", mirrorRoot, walkErr)
	}

	if len(expected) != 0 {
		missing := make([]string, 0, len(expected))
		for mirrorPath := range expected {
			missing = append(missing, mirrorPath)
		}

		sort.Strings(missing)

		return fmt.Errorf("generated chart mirror is missing expected files: %s", strings.Join(missing, ", "))
	}

	return nil
}
