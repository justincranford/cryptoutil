// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"

	fiber "github.com/gofiber/fiber/v2"
)

// KeyStatusResponse represents the status of a specific barrier key.
type KeyStatusResponse struct {
	UUID      string `json:"uuid"`       // Key UUID
	CreatedAt int64  `json:"created_at"` // Unix epoch milliseconds
	UpdatedAt int64  `json:"updated_at"` // Unix epoch milliseconds
}

// KeysStatusResponse represents the current status of all barrier key layers.
type KeysStatusResponse struct {
	RootKey         *KeyStatusResponse `json:"root_key"`         // Latest root key (nil if none)
	IntermediateKey *KeyStatusResponse `json:"intermediate_key"` // Latest intermediate key (nil if none)
}

// StatusService provides barrier key status query operations.
type StatusService struct {
	repo Repository
}

// NewStatusService creates a new StatusService instance.
func NewStatusService(repo Repository) (*StatusService, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository must be non-nil")
	}

	return &StatusService{repo: repo}, nil
}

// GetBarrierKeysStatus retrieves the current status of barrier keys.
// Returns the latest root and intermediate keys.
// Content keys are not included (elastic key rotation - no "latest" concept).
func (s *StatusService) GetBarrierKeysStatus(ctx context.Context) (*KeysStatusResponse, error) {
	var response KeysStatusResponse

	err := s.repo.WithTransaction(ctx, func(tx Transaction) error {
		// Get latest root key.
		rootKey, err := tx.GetRootKeyLatest()
		if err != nil {
			return fmt.Errorf("failed to get latest root key: %w", err)
		}

		if rootKey != nil {
			response.RootKey = &KeyStatusResponse{
				UUID:      rootKey.UUID.String(),
				CreatedAt: rootKey.CreatedAt,
				UpdatedAt: rootKey.UpdatedAt,
			}
		}

		// Get latest intermediate key.
		intermediateKey, err := tx.GetIntermediateKeyLatest()
		if err != nil {
			return fmt.Errorf("failed to get latest intermediate key: %w", err)
		}

		if intermediateKey != nil {
			response.IntermediateKey = &KeyStatusResponse{
				UUID:      intermediateKey.UUID.String(),
				CreatedAt: intermediateKey.CreatedAt,
				UpdatedAt: intermediateKey.UpdatedAt,
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get barrier keys status: %w", err)
	}

	return &response, nil
}

// HandleGetBarrierKeysStatus handles GET /admin/api/v1/barrier/keys/status requests.
// Returns the current status of barrier keys (latest root and intermediate keys).
func HandleGetBarrierKeysStatus(statusService *StatusService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		// Get barrier keys status.
		status, err := statusService.GetBarrierKeysStatus(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Failed to retrieve barrier keys status",
			})
		}

		return c.Status(fiber.StatusOK).JSON(status)
	}
}

// RegisterStatusRoutes registers the barrier keys status HTTP endpoints.
// Routes:
//   - GET /admin/api/v1/barrier/keys/status: Get current barrier keys status
//
// This function should be called during server initialization to wire up
// status endpoints to the admin server's fiber.App.
func RegisterStatusRoutes(app *fiber.App, statusService *StatusService) {
	app.Get("/admin/api/v1/barrier/keys/status", HandleGetBarrierKeysStatus(statusService))
}
