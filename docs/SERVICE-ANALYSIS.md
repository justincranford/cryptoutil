# Service Analysis: Consolidation Feasibility vs sm-kms

Date: 2026-05-24

## Executive Summary

- [Bottom line](#bottom-line)
- [sm-im vs sm-kms](#sm-im-vs-sm-kms-deep-analysis)
- [jose-ja vs sm-kms](#jose-ja-vs-sm-kms-deep-analysis)
- [pki-ca vs sm-kms](#pki-ca-vs-sm-kms-deep-analysis)
- [Maintenance cost reality](#maintenance-cost-reality)
- [Recommendation and path forward](#recommendation-and-path-forward)

## Bottom line

1. `sm-kms` is **not** a superset of `sm-im` today.
2. `sm-kms` is **close to** a superset of `jose-ja`, but not complete.
3. `sm-kms` is **not** a superset of `pki-ca` and should not replace PKI CA responsibilities.
4. If your primary goal is reducing alignment/lint drift pain, the highest ROI is likely:
   - Keep `pki-ca` separate.
   - Strongly evaluate folding `jose-ja` into `sm-kms`.
   - Treat `sm-im` as a product/API decision (messaging semantics), not as a pure crypto overlap question.

## Scope and method

This analysis uses exact OpenAPI method+path comparison and service code/test footprint metrics:

- `sm-kms`: [api/sm-kms/openapi_spec.yaml](../api/sm-kms/openapi_spec.yaml)
- `sm-im`: [api/sm-im/openapi_spec.yaml](../api/sm-im/openapi_spec.yaml)
- `jose-ja`: [api/jose-ja/openapi_spec.yaml](../api/jose-ja/openapi_spec.yaml)
- `pki-ca`: [api/pki-ca/openapi_spec_enrollment.yaml](../api/pki-ca/openapi_spec_enrollment.yaml)
- Service code roots:
  - [internal/apps/sm-kms](../internal/apps/sm-kms)
  - [internal/apps/sm-im](../internal/apps/sm-im)
  - [internal/apps/jose-ja](../internal/apps/jose-ja)
  - [internal/apps/pki-ca](../internal/apps/pki-ca)

Comparison rule: exact HTTP method + path equality within each service's declared server base path.

## sm-im vs sm-kms (Deep Analysis)

### API overlap result

- `sm-im` operations: 5
- Shared with `sm-kms`: 0
- Unique in `sm-im` (not in `sm-kms`): 5

### APIs unique to sm-im (not in sm-kms)

- `GET /messages`
- `GET /messages/{messageID}`
- `DELETE /messages/{messageID}`
- `POST /messages/send`
- `GET /messages/receive`

### Interpretation

`sm-im` is a messaging domain API (mailbox/message lifecycle + recipient semantics). `sm-kms` exposes key/container and crypto primitive APIs. Even where cryptography overlaps conceptually, the API contract and domain objects do not.

### Consolidation implication

- `sm-kms` cannot drop-in replace `sm-im` clients today.
- You can consolidate only by **building a messaging facade/domain in sm-kms** (or another service) that recreates these message APIs and behaviors.
- This is not just endpoint renaming; it is domain relocation.

## jose-ja vs sm-kms (Deep Analysis)

### API overlap result

- `jose-ja` operations: 16
- Shared with `sm-kms`: 13
- Unique in `jose-ja` (not in `sm-kms`): 3

### APIs unique to jose-ja (not in sm-kms)

- `GET /elastic-keys/{elasticKeyID}/material-keys/active`
- `POST /elastic-keys/{elasticKeyID}/rotate`
- `GET /jwks`

### Additional notable delta (where sm-kms has APIs jose-ja does not)

- `PUT /elastic-keys/{elasticKeyID}`
- `POST /elastic-keys/{elasticKeyID}/import`
- `DELETE /elastic-keys/{elasticKeyID}/material-keys/{materialKeyID}`
- `POST /elastic-keys/{elasticKeyID}/material-keys/{materialKeyID}/revoke`

### Interpretation

`jose-ja` and `sm-kms` substantially overlap on elastic/material key and crypto operations. `jose-ja` adds explicit JWKS publication and rotation/active-key convenience endpoints.

### Consolidation implication

- This is the strongest candidate for merge into `sm-kms`.
- Minimal API gap to close in `sm-kms`: active key endpoint, rotate endpoint, JWKS endpoint.
- Main risk is compatibility for existing JOSE clients (payload schema/status code/response shape differences), not core crypto capability.

## pki-ca vs sm-kms (Deep Analysis)

### API overlap result

- `pki-ca` operations: 18
- Shared with `sm-kms`: 0
- Unique in `pki-ca` (not in `sm-kms`): 18

### APIs unique to pki-ca (not in sm-kms)

- `GET /cas`
- `GET /cas/{caID}`
- `GET /cas/{caID}/crl`
- `POST /enrollments`
- `GET /enrollments/{requestID}`
- `GET /certificates`
- `GET /certificates/{serialNumber}`
- `GET /certificates/{serialNumber}/chain`
- `POST /certificates/{serialNumber}/revoke`
- `GET /profiles`
- `GET /profiles/{profileID}`
- `POST /ocsp`
- `GET /est/cacerts`
- `POST /est/simpleenroll`
- `POST /est/simplereenroll`
- `POST /est/serverkeygen`
- `GET /est/csrattrs`
- `POST /tsa/timestamp`

### Interpretation

`pki-ca` is a CA/PKI protocol service (enrollment, revocation, CRL, OCSP, EST, TSA). `sm-kms` is not designed as an X.509 CA protocol endpoint service.

### Consolidation implication

- Collapsing `pki-ca` into `sm-kms` would be a major architectural expansion, not an API merge.
- It would likely increase complexity/risk in `sm-kms` rather than reduce total system complexity.

## Maintenance cost reality

Code and test footprint under `internal/apps/*`:

| Service | Go files | Prod files | Test files | Total LOC | Prod LOC | Test LOC |
|---|---:|---:|---:|---:|---:|---:|
| sm-kms | 71 | 20 | 51 | 12059 | 2853 | 9206 |
| sm-im | 60 | 18 | 42 | 7934 | 1580 | 6354 |
| jose-ja | 75 | 22 | 53 | 14112 | 3424 | 10688 |
| pki-ca | 119 | 37 | 82 | 25841 | 9169 | 16672 |

Observations:

1. Test maintenance dominates all four services.
2. `pki-ca` is by far the largest maintenance surface.
3. Consolidation can reduce duplicate scaffolding/tests, but only where API/domain overlap is real.

## Recommendation and path forward

### Recommended consolidation strategy

1. **Phase 1 (highest ROI):** Merge `jose-ja` capability into `sm-kms` first.
2. **Phase 2 (optional):** Reassess `sm-im` based on product need for message lifecycle APIs.
3. **Do not collapse `pki-ca` into `sm-kms`** unless you explicitly want `sm-kms` to become a full CA protocol service.

### Practical next step design

If your goal is reducing drift and duplicated maintenance quickly:

1. Add the 3 missing JOSE endpoints to `sm-kms` (`/jwks`, `/rotate`, `/material-keys/active`).
2. Provide a compatibility shim/deprecation layer for `jose-ja` clients.
3. Keep `sm-im` and `pki-ca` separate until there is a deliberate domain migration plan (not just code alignment pressure).

### Direct answer to your core question

- Is `sm-kms` currently a superset of `sm-im`? **No.**
- Is `sm-kms` currently a superset of `jose-ja`? **Almost, but not yet.**
- Is `sm-kms` currently a superset of `pki-ca`? **No.**

Your instinct that maintaining separate services is expensive is accurate. The evidence supports starting consolidation with `jose-ja`, not with `sm-im` or `pki-ca`.
