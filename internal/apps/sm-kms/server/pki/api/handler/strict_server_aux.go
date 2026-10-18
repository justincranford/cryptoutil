// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"fmt"
	"io"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceTimestamp "cryptoutil/internal/apps/sm-kms/server/pki/service/timestamp"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
)

func (s *StrictServer) ListProfiles(_ context.Context, _ cryptoutilApiCaServer.ListProfilesRequestObject) (cryptoutilApiCaServer.ListProfilesResponseObject, error) {
	h := s.handler

	h.mu.RLock()
	defer h.mu.RUnlock()

	profiles := make([]cryptoutilApiCaServer.ProfileSummary, 0, len(h.profiles))
	for _, profile := range h.profiles {
		category := h.mapCategory(profile.Category)
		profiles = append(profiles, cryptoutilApiCaServer.ProfileSummary{
			ID: profile.ID, Name: profile.Name, Description: &profile.Description, Category: &category,
		})
	}

	return strictJSON(fiber.StatusOK, cryptoutilApiCaServer.ProfileListResponse{Profiles: profiles}), nil
}

func (s *StrictServer) GetProfile(_ context.Context, request cryptoutilApiCaServer.GetProfileRequestObject) (cryptoutilApiCaServer.GetProfileResponseObject, error) {
	s.handler.mu.RLock()
	profile, exists := s.handler.profiles[request.ProfileID]
	s.handler.mu.RUnlock()

	if !exists {
		return strictError(fiber.StatusNotFound, "not_found", "profile not found"), nil
	}

	return strictJSON(fiber.StatusOK, s.handler.buildProfileResponse(profile)), nil
}

func (s *StrictServer) TsaTimestamp(_ context.Context, request cryptoutilApiCaServer.TsaTimestampRequestObject) (cryptoutilApiCaServer.TsaTimestampResponseObject, error) {
	h := s.handler
	h.mu.RLock()
	tsaService := h.tsaService
	h.mu.RUnlock()

	if tsaService == nil {
		return strictError(fiber.StatusServiceUnavailable, "service_unavailable", "TSA service not configured"), nil
	}

	if request.Body == nil {
		return strictError(fiber.StatusBadRequest, "bad_request", "empty timestamp request"), nil
	}

	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		return strictError(fiber.StatusBadRequest, cryptoutilSharedMagic.ErrorInvalidRequest, fmt.Sprintf("failed to parse timestamp request: %v", err)), nil
	}

	if len(requestBody) == 0 {
		return strictError(fiber.StatusBadRequest, "bad_request", "empty timestamp request"), nil
	}

	tsReq, err := cryptoutilCAServiceTimestamp.ParseTimestampRequest(requestBody)
	if err != nil {
		return strictError(fiber.StatusBadRequest, cryptoutilSharedMagic.ErrorInvalidRequest, fmt.Sprintf("failed to parse timestamp request: %v", err)), nil
	}

	tsResp, err := tsaService.CreateTimestamp(tsReq)
	if err != nil {
		return strictError(fiber.StatusInternalServerError, "timestamp_error", err.Error()), nil
	}

	respDER, err := cryptoutilCAServiceTimestamp.SerializeTimestampResponse(tsResp)
	if err != nil {
		return strictError(fiber.StatusInternalServerError, "serialization_error", err.Error()), nil
	}

	return strictBinary(fiber.StatusOK, cryptoutilSharedMagic.PKIContentTypeTimestampReply, respDER), nil
}
