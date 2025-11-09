# Task 07 – Client Authentication Enhancements

## Objective

Stabilize and extend client authentication methods (basic, JWT-based, bearer, mTLS) with policy-driven controls, certificate validation, and developer-friendly ergonomics.

## Historical Context

- Tasks 5 and 6 (commits `ca597cd`, `35fde63`) delivered initial implementations but lacked cohesive policy management and CLI guidance.
- Later fixes exposed gaps in certificate validation and fallback logic.

## Scope

- Review all client authentication handlers for security posture, error clarity, and configuration hooks.
- Introduce policy abstractions under `internal/identity/security` to govern allowed methods per client profile.
- Enhance CLI tooling and documentation for configuring and testing client auth modes.

## Deliverables

- Updated client authentication modules with policy enforcement and improved certificate handling.
- CLI samples and documentation demonstrating configuration flows for each auth method.
- Integration tests (including mTLS) exercising success and failure scenarios.

## Validation

- Run `go test ./internal/identity/auth/...` with high coverage.
- Execute mTLS end-to-end tests, ideally reusing the deterministic mock services added in `5c04e44`.
- Map results to requirements registry entries covering client authentication.

## Dependencies

- Builds upon Task 06 (authorization server rehab) and Task 03 (configuration normalization).
- Provides prerequisites for Tasks 08, 11, and 20.

## Risks & Notes

- Handle certificate rotation gracefully to avoid breaking long-lived clients.
- Document policy defaults to prevent accidental lockouts.
