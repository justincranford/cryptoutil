// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExecRunnerHelper(t *testing.T) {
	t.Parallel()

	if len(os.Args) < 2 || os.Args[len(os.Args)-2] != testSeparator {
		return
	}

	switch os.Args[len(os.Args)-1] {
	case testEcho:
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			os.Exit(2)
		}

		_, _ = os.Stdout.Write(data)

		os.Exit(0)
	case testFailure:
		_, _ = fmt.Fprint(os.Stderr, "private diagnostic")

		os.Exit(1)
	case testWait:
		<-time.After(time.Minute)
		os.Exit(0)
	case testLarge:
		_, _ = os.Stdout.Write(make([]byte, maxOutputBytes+1))

		os.Exit(0)
	}
}

func TestExecRunner(t *testing.T) {
	t.Parallel()

	executable, err := os.Executable()
	require.NoError(t, err)

	for _, name := range []string{testEcho, testFailure, testWait, testLarge, testMissingExecutable, testNilContext} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := executable

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if name == testWait {
				cancel()
			}

			if name == testMissingExecutable {
				path += "-does-not-exist"
			}

			if name == testNilContext {
				ctx = nil
			}

			result, err := (ExecRunner{}).Run(ctx, path, []string{"-test.run=^TestExecRunnerHelper$", testSeparator, name}, []byte("memory-only-test-material"))
			if name == testEcho {
				require.NoError(t, err)
				require.Equal(t, []byte("memory-only-test-material"), result)
			} else {
				require.Error(t, err)
				require.Nil(t, result)
				require.NotContains(t, err.Error(), "private diagnostic")

				if name == testWait {
					require.ErrorIs(t, err, context.Canceled)
				}
			}
		})
	}
}

func TestBoundedOutput(t *testing.T) {
	t.Parallel()

	output := &boundedOutput{}
	n, err := output.Write(make([]byte, maxOutputBytes))
	require.NoError(t, err)
	require.Equal(t, maxOutputBytes, n)
	n, err = output.Write([]byte{0})
	require.Error(t, err)
	require.Zero(t, n)
	require.Equal(t, maxOutputBytes, output.Len())
}
