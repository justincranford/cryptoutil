// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	"errors"
	"io"
	http "net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentProductionWrappersRejectBeforeMutation(t *testing.T) {
	t.Parallel()

	for mode := range 4 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			input := developmentSmokeInput{
				root: root, image: testDevSyntheticImage, docker: filepath.Join(root, "docker.exe"),
				kubectl: filepath.Join(root, "kubectl.exe"), helm: filepath.Join(root, "helm.exe"), timeout: time.Minute,
			}
			calls := 0
			runner := rejectedPublicationRunner{calls: &calls}

			switch mode {
			case 0:
				require.Error(t, executeDevelopmentSmoke(t.Context(), input, runner, io.Discard))
				require.Zero(t, calls)
			case 1:
				require.Error(t, verifyDevelopmentPublicAPI(t.Context(), gen.BootstrapPreflightInput{}, nil, execution.KubectlConfig{}))
			case 2:
				require.Error(t, deployDevelopmentSmoke(t.Context(), input, gen.BootstrapPreflightInput{}, gen.LocalK3sPlan{},
					execution.Request{}, execution.KubectlConfig{}, runner, io.Discard))
				require.Equal(t, 1, calls)
			case testDevThirdCase:
				var stdout, stderr bytes.Buffer

				args := []string{"--repository-root", root, "--image", testDevSyntheticImage, "--kubectl", input.kubectl, "--helm", input.helm, "--timeout", "1m"}
				require.Equal(t, 1, Main(append([]string{devSmokeCommand}, args...), nil, &stdout, &stderr))
				require.Zero(t, stdout.Len())
			}
		})
	}
}

func TestDevelopmentNativeForwardHasBoundedClosure(t *testing.T) {
	t.Parallel()

	for mode := range 2 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()
			binary := filepath.Join(t.TempDir(), "absent-native-command.exe")

			if mode == 0 {
				path, err := exec.LookPath("go")
				require.NoError(t, err)
				binary, err = filepath.Abs(path)
				require.NoError(t, err)
			}

			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()

			closeForward, err := startDevelopmentForward(ctx, binary, []string{cryptoutilSharedMagic.CLIVersionCommand})
			if mode == 0 {
				require.NoError(t, err)
				closeForward()
			} else {
				require.Error(t, err)
				require.Nil(t, closeForward)
			}
		})
	}
}

func TestDevelopmentWaitAndAuthenticationRejectInvalidTransport(t *testing.T) {
	t.Parallel()

	for mode := range 4 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			switch mode {
			case 0:
				cancel()
				require.ErrorIs(t, waitDevelopment(ctx, func() error {
					require.FailNow(t, "canceled wait executed")

					return nil
				}), context.Canceled)
			case 1:
				calls := 0

				require.NoError(t, waitDevelopment(ctx, func() error {
					calls++
					if calls == 1 {
						return errors.New("synthetic retry")
					}

					return nil
				}))
				require.Equal(t, 2, calls)
			case 2:
				require.Error(t, verifyDevelopmentAuthentication(ctx, &http.Client{}, "http://["))
			case testDevThirdCase:
				require.Error(t, verifyDevelopmentAuthentication(ctx, &http.Client{}, "unsupported://owned"))
			}
		})
	}
}
