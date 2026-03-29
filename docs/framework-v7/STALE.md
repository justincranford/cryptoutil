# STALE.md — Stale Items in docs/ARCHITECTURE.md

Items identified as outdated, inconsistent, or needing cleanup.

## Skeleton-Template Status Contradiction

1. Section 3.2 Implementation Status table shows skeleton-template as "Complete | ~95%" but Section 3.2.5.1 detailed description marks it "Not Started". Reconcile to reflect actual implementation state.

## Migration Priority Inconsistency

1. Section 2.2 (line ~264) lists migration priority as `sm-im -> jose-ja -> sm-kms -> skeleton-template -> pki-ca -> identity services`. Section 5.1.3 (line ~1291) omits skeleton-template from the list. Reconcile both sections to use the same migration priority order.

## Outdated Phase References Without Context

1. Multiple sections reference Phase numbers (Phase 0, Phase 2C, Phase 7, Phase 8) without a phase numbering scheme introduction. Either add a phase overview section or remove phase references in favor of descriptive timelines.
2. Section 3.2 references "Phase 8 reintegration" for pki-ca and all identity services (0% completion, archived domains). Verify whether Phase 8 is current or stale, and update status accordingly.
3. Section 6.2 references "Phase 2C (deferred)" and "Future (Phase 2C)" for mTLS authentication. Clarify whether Phase 2C is still planned or has been superseded.

## Archived Domain Directory References

1. Section 3.2 Implementation Status references `_ca-archived/`, `_authz-archived/`, `_idp-archived/`, `_rs-archived/`, `_rp-archived/`, `_spa-archived/` directories. The `archive-detector` fitness linter rejects archived directories. Remove these references or update to reflect current directory state.

## Demo Deferral Without Timeline

1. Section 13.1.6 states "Demo orchestration is deferred until a solid E2E orchestration foundation exists" with no date or status. Update with current E2E orchestration status and whether demo support is still deferred.

## Port Assignment Order vs Canonical Service Order

1. Section 3.4 Port Assignments table lists ports in an order that does not match the canonical service order (sm-kms=8000, pki-ca=8100, identity-authz=8200, identity-idp=8300, identity-rs=8400, identity-rp=8500, identity-spa=8600, sm-im=8700, jose-ja=8800, skeleton-template=8900). The port numbers themselves are a separate concern (see PORT-REORDERING.md), but the table row ordering should match canonical order for consistency.

## Document Version Date

1. Document header shows version 2.0 dated February 8, 2026. Verify this is intentional and update if the document has been significantly revised since that date.
