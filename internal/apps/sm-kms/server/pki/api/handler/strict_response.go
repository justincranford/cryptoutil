// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"errors"
	"fmt"
	"io"
	http "net/http"
	"strings"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
)

type strictResponse struct {
	statusCode  int
	contentType string
	body        any
	headers     [][2]string
}

func (response strictResponse) write(ctx *fiber.Ctx) error {
	ctx.Status(response.statusCode)

	if response.contentType != "" {
		ctx.Set("Content-Type", response.contentType)
	}

	for _, header := range response.headers {
		ctx.Set(header[0], header[1])
	}

	switch body := response.body.(type) {
	case nil:
		if err := ctx.SendStatus(response.statusCode); err != nil {
			return fmt.Errorf("failed to send status response: %w", err)
		}

		return nil
	case []byte:
		if err := ctx.Send(body); err != nil {
			return fmt.Errorf("failed to send binary response: %w", err)
		}

		return nil
	case io.Reader:
		if _, err := io.Copy(ctx.Response().BodyWriter(), body); err != nil {
			return fmt.Errorf("failed to write binary response: %w", err)
		}

		return nil
	default:
		if err := ctx.JSON(body); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}

		return nil
	}
}

func strictJSON(statusCode int, body any) strictResponse {
	return strictResponse{statusCode: statusCode, contentType: "application/json", body: body}
}

func strictBinary(statusCode int, contentType string, body []byte, headers ...[2]string) strictResponse {
	return strictResponse{statusCode: statusCode, contentType: contentType, body: body, headers: headers}
}

func strictStatus(statusCode int, contentType string) strictResponse {
	return strictResponse{statusCode: statusCode, contentType: contentType}
}

func strictError(statusCode int, errorCode, message string) strictResponse {
	return strictJSON(statusCode, cryptoutilApiCaServer.ErrorResponse{
		Error:   errorCode,
		Message: &message,
	})
}

func strictResponseResult[T any](response strictResponse) (T, error) {
	typedResponse, ok := any(response).(T)
	if !ok {
		var zero T

		return zero, fmt.Errorf("strict response does not implement generated response type")
	}

	return typedResponse, nil
}

func strictHandledResponse(response strictResponse) (strictResponse, error) {
	return response, nil
}

// StrictRequestBodyErrorMiddleware preserves PKI JSON parser error responses.
func StrictRequestBodyErrorMiddleware(ctx *fiber.Ctx) error {
	if clientCert := estClientCertificateFromFiberContext(ctx); clientCert != nil {
		ctx.SetUserContext(withESTClientCertificate(ctx.UserContext(), clientCert))
	}

	if err := ctx.Next(); err != nil {
		if isStrictRequestBodyParseError(ctx, err) {
			message := cryptoutilSharedMagic.PKIMessageFailedToParseRequestBody

			if err := ctx.Status(fiber.StatusBadRequest).JSON(cryptoutilApiCaServer.ErrorResponse{
				Error:   cryptoutilSharedMagic.ErrorInvalidRequest,
				Message: &message,
			}); err != nil {
				return fmt.Errorf("failed to write strict request error response: %w", err)
			}

			return nil
		}

		return fmt.Errorf("strict request middleware failed: %w", err)
	}

	return nil
}

func isStrictRequestBodyParseError(ctx *fiber.Ctx, err error) bool {
	var fiberError *fiber.Error
	if !errors.As(err, &fiberError) || fiberError.Code != fiber.StatusBadRequest || ctx.Method() != http.MethodPost {
		return false
	}

	routePath := ctx.Route().Path

	return strings.Contains(routePath, "/pki/") &&
		(strings.HasSuffix(routePath, "/enrollments") ||
			strings.HasSuffix(routePath, "/certificates/:serialNumber/revoke"))
}

func (response strictResponse) VisitListCAsResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitGetCAResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitGetCRLResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitListCertificatesResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitGetCertificateResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitGetCertificateChainResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitRevokeCertificateResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitSubmitEnrollmentResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitGetEnrollmentStatusResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitESTCACertsResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitESTCSRAttrsResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitESTServerKeyGenResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitESTSimpleEnrollResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitESTSimpleReenrollResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitHandleOCSPResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitListProfilesResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitGetProfileResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}

func (response strictResponse) VisitTsaTimestampResponse(ctx *fiber.Ctx) error {
	return response.write(ctx)
}
