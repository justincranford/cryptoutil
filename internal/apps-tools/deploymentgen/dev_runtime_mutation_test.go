// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentCommandHasNoMutationFlags(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{developmentRuntimePlanCommand, testDevelopmentHelp}, {developmentRuntimePlanCommand, "--write"}, {developmentRuntimePlanCommand, "--launch"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := Main(args, &stdout, &stderr, func(string) ([]byte, error) {
				t.Fatal("read before validated flags")

				return nil, nil
			}, func(string, []byte, os.FileMode) error {
				t.Fatal("development preflight must not write")

				return nil
			})
			if args[1] == testDevelopmentHelp {
				require.Zero(t, code)
			} else {
				require.Equal(t, 2, code)
			}
		})
	}
}
