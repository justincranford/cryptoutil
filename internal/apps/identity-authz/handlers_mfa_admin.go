// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// EnrollMFARequest represents a request to enroll an MFA factor.
type EnrollMFARequest struct {
	UserID     string `json:"user_id" validate:"required"`
	FactorType string `json:"factor_type" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Required   bool   `json:"required"`
}

// EnrollMFAResponse represents a response after enrolling an MFA factor.
type EnrollMFAResponse struct {
	ID         string `json:"id"`
	FactorType string `json:"factor_type"`
	Name       string `json:"name"`
	Required   bool   `json:"required"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  string `json:"created_at"`
}

// ListMFAFactorsResponse represents a response listing MFA factors.
type ListMFAFactorsResponse struct {
	Factors []MFAFactorSummary `json:"factors"`
}

// MFAFactorSummary represents a summary of an MFA factor.
type MFAFactorSummary struct {
	ID         string `json:"id"`
	FactorType string `json:"factor_type"`
	Name       string `json:"name"`
	Required   bool   `json:"required"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  string `json:"created_at"`
}

// handleEnrollMFA handles POST /oidc/v1/mfa/enroll.
func (s *Service) handleEnrollMFA(c *fiber.Ctx) error {
	var req EnrollMFARequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "invalid request body",
		})
	}

	// Parse user_id.
	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "invalid user_id format",
		})
	}

	// Verify user exists.
	ctx := c.Context()

	user, err := s.repoFactory.UserRepository().GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "user_not_found",
			"error_description":               fmt.Sprintf("user not found: %v", err),
		})
	}

	// Validate factor_type.
	factorType := cryptoutilIdentityDomain.MFAFactorType(req.FactorType)
	if !isValidMFAFactorType(factorType) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               fmt.Sprintf("invalid factor_type: %s", req.FactorType),
		})
	}

	// Get or create default auth profile for user.
	// NOTE: Using user-specific profile name as workaround since AuthProfile lacks UserID field.
	authProfileRepo := s.repoFactory.AuthProfileRepository()
	profileName := fmt.Sprintf("user_%s_default", user.ID.String())

	authProfile, err := authProfileRepo.GetByName(ctx, profileName)
	if err != nil {
		// Create default auth profile for user.
		authProfile = &cryptoutilIdentityDomain.AuthProfile{
			ID:          googleUuid.New(),
			Name:        profileName,
			Description: fmt.Sprintf("Default auth profile for user %s", user.ID.String()),
			ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
			RequireMFA:  false,
			Enabled:     true,
		}
		if err := authProfileRepo.Create(ctx, authProfile); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
				"error_description":               fmt.Sprintf("failed to create auth profile: %v", err),
			})
		}
	}

	// Create MFA factor.
	mfaFactor := &cryptoutilIdentityDomain.MFAFactor{
		ID:            googleUuid.New(),
		Name:          req.Name,
		FactorType:    factorType,
		AuthProfileID: authProfile.ID,
		Enabled:       true,
	}

	// Set required field (IntBool type stores as 0/1 in database).
	mfaFactor.Required = cryptoutilIdentityDomain.IntBool(req.Required)

	mfaFactorRepo := s.repoFactory.MFAFactorRepository()
	if err := mfaFactorRepo.Create(ctx, mfaFactor); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               fmt.Sprintf("failed to create MFA factor: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(EnrollMFAResponse{
		ID:         mfaFactor.ID.String(),
		FactorType: string(mfaFactor.FactorType),
		Name:       mfaFactor.Name,
		Required:   mfaFactor.Required.Bool(),
		Enabled:    mfaFactor.Enabled,
		CreatedAt:  mfaFactor.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// handleListMFAFactors handles GET /oidc/v1/mfa/factors.
func (s *Service) handleListMFAFactors(c *fiber.Ctx) error {
	// Get user_id from query params.
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "missing user_id query parameter",
		})
	}

	userID, err := googleUuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "invalid user_id format",
		})
	}

	// Verify user exists.
	ctx := c.Context()

	user, err := s.repoFactory.UserRepository().GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "user_not_found",
			"error_description":               fmt.Sprintf("user not found: %v", err),
		})
	}

	// Get user-specific auth profile (using naming convention).
	authProfileRepo := s.repoFactory.AuthProfileRepository()
	profileName := fmt.Sprintf("user_%s_default", user.ID.String())

	authProfile, err := authProfileRepo.GetByName(ctx, profileName)
	if err != nil {
		// User has no auth profile yet, return empty list.
		return c.Status(fiber.StatusOK).JSON(ListMFAFactorsResponse{
			Factors: []MFAFactorSummary{},
		})
	}

	// Get MFA factors for user's auth profile.
	mfaFactorRepo := s.repoFactory.MFAFactorRepository()

	factors, err := mfaFactorRepo.GetByAuthProfileID(ctx, authProfile.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               fmt.Sprintf("failed to get MFA factors: %v", err),
		})
	}

	allFactors := make([]MFAFactorSummary, 0, len(factors))
	for _, factor := range factors {
		allFactors = append(allFactors, MFAFactorSummary{
			ID:         factor.ID.String(),
			FactorType: string(factor.FactorType),
			Name:       factor.Name,
			Required:   factor.Required.Bool(),
			Enabled:    factor.Enabled,
			CreatedAt:  factor.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ListMFAFactorsResponse{
		Factors: allFactors,
	})
}

// handleDeleteMFAFactor handles DELETE /oidc/v1/mfa/factors/{id}.
func (s *Service) handleDeleteMFAFactor(c *fiber.Ctx) error {
	// Get factor_id from path params.
	factorIDStr := c.Params("id")
	if factorIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "missing factor id in path",
		})
	}

	factorID, err := googleUuid.Parse(factorIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "invalid factor id format",
		})
	}

	// Get user_id from query params for authorization check.
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "missing user_id query parameter",
		})
	}

	userID, err := googleUuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "invalid user_id format",
		})
	}

	ctx := c.Context()

	// Verify user exists.
	user, err := s.repoFactory.UserRepository().GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "user_not_found",
			"error_description":               fmt.Sprintf("user not found: %v", err),
		})
	}

	// Verify factor exists and belongs to user.
	mfaFactorRepo := s.repoFactory.MFAFactorRepository()

	factor, err := mfaFactorRepo.GetByID(ctx, factorID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "factor_not_found",
			"error_description":               fmt.Sprintf("MFA factor not found: %v", err),
		})
	}

	// Verify factor belongs to user's auth profile (via naming convention).
	authProfileRepo := s.repoFactory.AuthProfileRepository()

	authProfile, err := authProfileRepo.GetByID(ctx, factor.AuthProfileID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               fmt.Sprintf("failed to get auth profile: %v", err),
		})
	}

	// Verify auth profile belongs to user (check naming convention).
	expectedProfileName := fmt.Sprintf("user_%s_default", user.ID.String())
	if authProfile.Name != expectedProfileName {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: "unauthorized",
			"error_description":               "MFA factor does not belong to specified user",
		})
	}

	// Soft delete the factor.
	if err := mfaFactorRepo.Delete(ctx, factorID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               fmt.Sprintf("failed to delete MFA factor: %v", err),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// isValidMFAFactorType checks if the factor type is valid.
func isValidMFAFactorType(factorType cryptoutilIdentityDomain.MFAFactorType) bool {
	switch factorType {
	case cryptoutilIdentityDomain.MFAFactorTypePassword,
		cryptoutilIdentityDomain.MFAFactorTypeEmailOTP,
		cryptoutilIdentityDomain.MFAFactorTypeSMSOTP,
		cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		cryptoutilIdentityDomain.MFAFactorTypeHOTP,
		cryptoutilIdentityDomain.MFAFactorTypePasskey,
		cryptoutilIdentityDomain.MFAFactorTypeMagicLink,
		cryptoutilIdentityDomain.MFAFactorTypeMTLS,
		cryptoutilIdentityDomain.MFAFactorTypeHardwareToken,
		cryptoutilIdentityDomain.MFAFactorTypeBiometric:
		return true
	default:
		return false
	}
}
