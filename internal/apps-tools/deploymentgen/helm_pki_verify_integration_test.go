// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	testVerifyForgedWrite = "verify-only forged write"
	testVerifyForgedToken = "verify-only forged token"
)

func TestHelmPKIVerifyOnlyHasNoTokenOrWriteAndRejectsForgedInput(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testVerifySupported, testVerifyMissingSource, testVerifyForgedWrite, testVerifyForgedToken} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
			values, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)

			selected := ""

			for index := range values.Jobs {
				job := &values.Jobs[index]
				if !job.APIToken {
					continue
				}

				for sourceIndex := range values.Secrets {
					source := &values.Secrets[sourceIndex]
					if source.Namespace != job.Namespace {
						continue
					}

					source.Files[testVerifyInputFile] = testVerifyProtectedFile
					job.Args = []string{lifecycleVerifyCommand, "--input=/run/cryptoutil/secrets/" + source.Name + "/" + testVerifyInputFile}
					job.SecretNames = []string{source.Name}

					for boundaryIndex := range values.Namespaces {
						if values.Namespaces[boundaryIndex].Name == job.Namespace {
							values.Namespaces[boundaryIndex].ReadSecretNames = append(values.Namespaces[boundaryIndex].ReadSecretNames, source.Name)
						}
					}

					job.WriteSecretNames = []string{}
					job.APIToken = false
					selected = job.Name

					if test == testVerifyMissingSource {
						delete(source.Files, testVerifyInputFile)
					}

					if test == testVerifyForgedWrite {
						job.WriteSecretNames = []string{source.Name}
					}

					if test == testVerifyForgedToken {
						job.APIToken = true
					}

					break
				}

				break
			}

			require.NotEmpty(t, selected)

			raw, err := marshalCanonicalJSON(values)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			command := exec.CommandContext(ctx, requirePinnedHelm(t), cryptoutilSharedMagic.SkeletonTemplateServiceName, "cu-lifecycle", filepath.Join("..", "..", "..", canonicalChartSourceRoot, "cryptoutil-lifecycle"), "--values", "-", "--kube-version", "1.37.1")
			command.Stdin = bytes.NewReader(raw)

			output, err := command.CombinedOutput()
			if test != testVerifySupported {
				require.Error(t, err)

				return
			}

			require.NoError(t, err, string(output))

			decoder := yaml.NewDecoder(bytes.NewReader(output))
			seen := false

			for {
				var document map[string]any

				err := decoder.Decode(&document)
				if errors.Is(err, io.EOF) {
					break
				}

				require.NoError(t, err)

				metadata, _ := document["metadata"].(map[string]any)
				if document["kind"] != "Job" || metadata[cryptoutilSharedMagic.ClaimName] != selected {
					continue
				}

				seen = true

				encoded, err := marshalCanonicalJSON(document)
				require.NoError(t, err)
				require.NotContains(t, string(encoded), "bootstrap-api-token")
				require.NotContains(t, string(encoded), "serviceAccountToken")
				require.Contains(t, string(encoded), lifecycleVerifyCommand)
			}

			require.True(t, seen)
		})
	}
}
