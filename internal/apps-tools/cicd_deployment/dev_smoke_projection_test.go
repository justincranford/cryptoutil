// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"

	"github.com/stretchr/testify/require"
)

const (
	testProjectionReadInput   = "read-input"
	testProjectionModel       = "model-identity"
	testProjectionStorage     = "storage-identity"
	testProjectionRuntime     = "runtime-image"
	testProjectionPublication = "publication-json"
)

func TestDevelopmentSmokeProjectionBoundaries(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{testProjectionReadInput, testProjectionModel, testProjectionStorage, testProjectionRuntime, testProjectionPublication} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()

			root, err := filepath.Abs(filepath.Join("..", "..", ".."))
			require.NoError(t, err)

			read := func(relative string) ([]byte, error) {
				raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
				if err != nil {
					return nil, errors.New("public projection input unavailable")
				}

				return raw, nil
			}
			raw, err := read(testDevModelPath)
			require.NoError(t, err)
			model, err := gen.DecodeModel(raw)
			require.NoError(t, err)
			raw, err = read(testDevLockPath)
			require.NoError(t, err)
			lock, err := gen.DecodeToolchainLock(raw)
			require.NoError(t, err)
			plan, err := gen.BuildDevelopmentRuntimePlan(model, lock, gen.DevelopmentSmokeRun, gen.DevelopmentSmokeSelector, lock.Images.Postgres.Reference)
			require.NoError(t, err)
			raw, err = read(testDevDomainPath)
			require.NoError(t, err)
			plan, err = gen.ConfigureDevelopmentSmoke(plan, raw)
			require.NoError(t, err)

			switch failure {
			case testProjectionReadInput:
				read = func(string) ([]byte, error) { return nil, errors.New("fixed-read-failure") }
			case testProjectionModel:
				plan.Model.Run = "bad"
			case testProjectionStorage:
				plan.RuntimeInput.Storage.ClassName = ""
			case testProjectionRuntime:
				plan.RuntimeInput.Images = nil
			}

			input, err := developmentSmokePreflight(plan, lock, read)
			if failure == testProjectionPublication {
				require.NoError(t, err)

				input.RuntimeValues[0].Workloads[0].MaterializerConfig = "invalid-protected-json"
				_, err = developmentPublication(input.RuntimeValues)
			}

			require.Error(t, err)
		})
	}
}
