// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentImageCancellationAtSuccessfulProcessExitCannotBeAccepted(t *testing.T) {
	t.Parallel()
	raw, reference := developmentStreamArchive(t, false)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	factory := func(_ context.Context, _ string, _ []string, input io.Reader, output io.Writer) developmentImageProcess {
		return &developmentStreamFixtureProcess{input: input, output: output, archive: raw, onWait: cancel}
	}
	err := streamDevelopmentImage(ctx, "explicit-test-docker", reference, nil, factory)
	require.ErrorIs(t, err, context.Canceled)
}
