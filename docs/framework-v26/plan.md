# Framework v26 — In-Process PKI CA Configuration Reload

**Status**: Parked for future planning; not implementation-ready.
**Created**: 2026-10-21
**Purpose**: Define and implement safe in-process activation of supported PKI CA
configuration changes without restarting an `sm-kms` process.

## Scope

The current PKI component graph is assembled at startup. It includes the active issuer,
CRL and OCSP services, request handler, and public revocation endpoint map. See
[`pki_components.go`](../../internal/apps/sm-kms/server/pki/server/pki_components.go).
Authority and distribution settings are read from configuration files at startup; see
[`pki_authority_config.go`](../../internal/apps/sm-kms/server/pki/server/pki_authority_config.go).
The existing supported activation path for CA changes is coordinated process replacement.

This plan is for a future, explicitly triggered, in-process reload capability. It does
not assume that every PKI setting or CA key can be changed at runtime. The first phase
must define the reloadable setting allowlist and distinguish distribution/profile changes
from issuer certificate or signing-key rollover. Changes that conflict with persisted
authority identity or key material must continue to fail closed unless a separate,
reviewed rollover contract authorizes them.

The reload trigger is deliberately undecided. Phase 1 must compare the repository's
supported configuration and administration patterns before selecting a trigger. This
plan does not promise an atomic update across separate `sm-kms` replicas; any cluster
coordination contract must be explicit.

## Safety Invariants

- Parse and validate a complete candidate configuration and component graph before activation.
- A failed reload leaves the last known-good active generation serving requests.
- Each request uses one internally consistent generation; issuer, CRL, OCSP, handler, and
  endpoint routing must not be mixed across generations.
- Historical authorities needed to verify or serve existing certificates remain available
  after an active-authority change.
- Reload outcomes are observable without logging private keys, credentials, or other secrets.
- Existing startup validation and coordinated process replacement remain supported.

## Phases

### Phase 1: Define Reload Contract

Research existing configuration lifecycle, PKI component ownership, route registration,
and administrative operations. Specify reloadable fields, trigger, authorization,
validation, activation, rollback, in-flight request behavior, replica semantics, and
historical-authority retention before implementation begins.

### Phase 2: Implement Safe Per-Process Reload

Build and validate a replacement component generation off the request path, then activate
it as one coherent generation. Preserve the active generation on any parse, validation,
key-load, or component-construction failure. Provide an explicit and observable reload
result using the trigger selected in Phase 1.

### Phase 3: Verify and Document Operations

Test successful activation, rejected configuration, rollback, historical certificate
verification/revocation, and concurrent requests during reload. Document the selected
trigger, supported settings, per-replica behavior, failure recovery, and the continued
process-replacement fallback.

### Phase 4: Propagate Lessons

Update the engineering handbook and relevant PKI/configuration guidance if implementation
establishes a reusable project-wide pattern. Record phase post-mortems and completion
evidence before archiving this plan.

## Acceptance Criteria

- The reloadable setting allowlist and explicit non-reloadable settings are documented.
- A failed candidate cannot partially replace live issuer, CRL, OCSP, handler, or endpoint state.
- In-flight requests observe one complete generation, and concurrent reload/request tests pass
  under the race detector.
- Existing and historical authority behavior remains correct after activation.
- Reload success/failure is actionable to operators and does not disclose secret material.
- Normal and tagged builds, focused tests, repository validators, and applicable CI quality gates
  pass; operations documentation describes the supported activation and fallback paths.

## References

- [PKI component construction](../../internal/apps/sm-kms/server/pki/server/pki_components.go)
- [PKI authority configuration](../../internal/apps/sm-kms/server/pki/server/pki_authority_config.go)
- [TLS structure and rotation constraints](../tls-structure.md)
- [ENG-HANDBOOK.md §6 Security Architecture](../ENG-HANDBOOK.md#6-security-architecture)
- [ENG-HANDBOOK.md §8 API Architecture](../ENG-HANDBOOK.md#8-api-architecture)
- [ENG-HANDBOOK.md §10 Testing Architecture](../ENG-HANDBOOK.md#10-testing-architecture)
- [ENG-HANDBOOK.md §11 Quality Strategy](../ENG-HANDBOOK.md#11-quality-strategy)
- [ENG-HANDBOOK.md §14 Coding Standards](../ENG-HANDBOOK.md#14-development-practices)
