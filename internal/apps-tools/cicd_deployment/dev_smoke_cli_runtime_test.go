// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"

	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentSmokeCLIUsesExplicitBoundedNativeRuntime(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{testImageValid, testImageNativeStart, testImageTransport} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			docker := filepath.Join(root, "docker.exe")
			args := []string{
				"--repository-root=" + root, "--image=explicit-public-image",
				"--kubectl=" + filepath.Join(root, "kubectl.exe"), "--helm=" + filepath.Join(root, "helm.exe"), "--timeout=1m",
			}
			calls := 0
			lookup := func(name string) (string, error) {
				require.Equal(t, "docker", name)

				if mode == testImageNativeStart {
					return "", errors.New("PRIVATE")
				}

				return docker, nil
			}
			execute := func(ctx context.Context, input developmentSmokeInput, runner execution.Runner, _ io.Writer) error {
				calls++

				require.Equal(t, docker, input.docker)
				require.Equal(t, root, input.root)

				_, bounded := ctx.Deadline()
				require.True(t, bounded)
				require.IsType(t, execution.ExecRunner{}, runner)

				if mode == testImageTransport {
					return errors.New("public bounded runtime rejection")
				}

				return nil
			}

			var stdout, stderr bytes.Buffer

			code := developmentSmokeMainWithRuntime(args, &stdout, &stderr, lookup, execute)
			if mode == testImageValid {
				require.Zero(t, code)
				require.Equal(t, 1, calls)
			} else {
				require.Equal(t, 1, code)
				require.NotContains(t, stderr.String(), "PRIVATE")
			}
		})
	}
}
