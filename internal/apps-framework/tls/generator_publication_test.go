// Copyright (c) 2025-2026 Justin Cranford.
//

package tls_test

import (
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestGenerate_RenameFailureRollsBackPublishedOutputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		failOnRename int32
	}{
		{name: "certificate tree published before config failure", failOnRename: 2},
		{name: "certificate tree and config published before bundle failure", failOnRename: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			outputDir := t.TempDir()
			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				os.MkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
				stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
			)

			var renameCount atomic.Int32

			gen.ExportedSetRenameFn(func(source, destination string) error {
				if renameCount.Add(1) == tc.failOnRename {
					return errors.New("injected rename failure")
				}

				return os.Rename(source, destination)
			})

			err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, outputDir)
			require.ErrorContains(t, err, "injected rename failure")
			require.NoDirExists(t, filepath.Join(outputDir, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate))
			require.NoFileExists(t, filepath.Join(outputDir, "tls-config.yml"))
			require.NoFileExists(t, filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitAdminCABundleFilename))

			publishedFiles, err := os.ReadDir(outputDir)
			require.NoError(t, err)
			require.Empty(t, publishedFiles)
		})
	}
}
