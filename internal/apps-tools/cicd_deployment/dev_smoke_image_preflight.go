// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"errors"
)

const developmentImageInventoryCount = 3

func verifyNativeDevelopmentImages(ctx context.Context, input developmentSmokeInput, references []string) error {
	return verifyDevelopmentImagesWithFactory(ctx, input, references, nativeDevelopmentImageProcess)
}

func verifyDevelopmentImagesWithFactory(ctx context.Context, input developmentSmokeInput, references []string, factory developmentImageFactory) error {
	if len(references) != developmentImageInventoryCount {
		return errors.New("explicit complete disposable image inventory required")
	}

	for _, reference := range references {
		if err := streamDevelopmentImage(ctx, input.docker, reference, nil, factory); err != nil {
			return errors.New("disposable immutable image archive admission rejected")
		}
	}

	return nil
}
