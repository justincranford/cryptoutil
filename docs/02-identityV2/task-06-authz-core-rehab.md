# Task 06 – OAuth 2.1 Authorization Server Core Rehab

## Objective

Bring the authorization server back into alignment with OAuth 2.1 draft 15, addressing incomplete flows, error handling gaps, and missing telemetry introduced since commit `0418528`.

## Historical Context

- The original Task 4 implementation landed in `0418528`, but subsequent commits flagged missing scopes, PKCE edge cases, and logging inconsistencies.
- Partial coverage identified during the SPA and e2e work indicates the need for a full rehab.

## Scope

- Audit authorization code, refresh token, and client credentials flows for spec compliance.
- Harden error handling and logging (structured, contextual) to support observability requirements.
- Update documentation to reflect supported grant types and configuration toggles.

## Deliverables

- Updated server code with comprehensive unit and integration tests.
- Spec conformance matrix documenting grant type handling, PKCE enforcement, and error responses.
- Enhanced logs and telemetry wiring, including trace/span propagation.

## Validation

- Automated test suites covering all grant types and error branches.
- Optional Postman or Go-based integration suite demonstrating end-to-end flows.
- Verification against requirements registry entries related to authorization flows.

## Dependencies

- Utilizes configuration templates from Task 03 and requirement IDs from Task 02.
- Provides foundation for Tasks 07, 09, and 10.

## Risks & Notes

- Ensure backward compatibility with existing clients; document any breaking changes explicitly.
- Coordinate token schema changes with Task 08 to avoid duplicate work.
