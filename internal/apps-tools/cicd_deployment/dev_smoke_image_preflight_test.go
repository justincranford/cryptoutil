// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentImagePreflightNeverStartsCompute(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{testImageValid, testImageCount, testImageReference, testImageNativeStart} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			raw, reference := developmentStreamArchive(t, false)

			references := []string{reference, reference, reference}
			if mode == testImageCount {
				references = nil
			}

			if mode == testImageReference {
				references[1] = "unbound-image"
			}

			calls := 0
			factory := func(_ context.Context, _ string, args []string, input io.Reader, output io.Writer) developmentImageProcess {
				require.Nil(t, input)
				require.Equal(t, []string{developmentImage, developmentImageSave, reference}, args)

				calls++

				return &developmentStreamFixtureProcess{input: input, output: output, archive: raw}
			}
			input := developmentSmokeInput{docker: "absent-development-image-executable"}

			var err error
			if mode == testImageNativeStart {
				err = verifyNativeDevelopmentImages(t.Context(), input, references)
			} else {
				err = verifyDevelopmentImagesWithFactory(t.Context(), input, references, factory)
			}

			if mode == testImageValid {
				require.NoError(t, err)
				require.Equal(t, developmentImageInventoryCount, calls)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")
			}
		})
	}
}
