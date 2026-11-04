// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testDiagnosticCancellation = "cancellation"
	testDiagnosticProcess      = "process"
	testDiagnosticOutputBound  = "output-bound"
)

func TestDevelopmentNativeDiagnosticMemoryAndRedaction(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{"", testDiagnosticProcess, testDiagnosticCancellation, testDiagnosticOutputBound} {
		t.Run("native="+failure, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			raw, err := readNativeDevelopmentLogs(ctx, "explicit-docker", "captured-immutable-id", func(command *exec.Cmd) error {
				require.Nil(t, command.Env)
				require.Nil(t, command.Stdin)
				require.Same(t, command.Stdout, command.Stderr)
				require.Equal(t, []string{"explicit-docker", developmentLogCommand, "--tail", "200", "captured-immutable-id"}, command.Args)

				if failure == testDiagnosticProcess {
					return errors.New(testDiagnosticPrivate)
				}

				if failure == testDiagnosticOutputBound {
					_, err := command.Stderr.Write([]byte(strings.Repeat("x", developmentDiagnosticBytes+1)))

					return err
				}

				if failure == testDiagnosticCancellation {
					cancel()
				}

				_, err := command.Stdout.Write([]byte("PRIVATE-token "))
				require.NoError(t, err)
				_, err = command.Stderr.Write([]byte("permission denied PRIVATE-key"))

				return err
			})
			if failure == "" {
				require.NoError(t, err)
				require.Equal(t, "bootstrap observed permission failure marker", developmentStartupMarker(raw))
			} else {
				require.Error(t, err)
				require.Nil(t, raw)
				require.NotContains(t, err.Error(), testDiagnosticPrivate)
			}
		})
	}
}
