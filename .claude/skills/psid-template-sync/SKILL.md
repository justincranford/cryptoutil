---
name: psid-template-sync
description: "Keep stable PS-ID template-instantiated files synchronized across all 10 services using the canonical internal app templates and exact template-drift enforcement."
argument-hint: "[template path or PS-ID file family]"
---

Keep stable PS-ID template-instantiated files synchronized across all 10 services.

## Purpose

Use this skill when a change belongs in the canonical internal app templates rather than in one service only.
This applies to the stable PS-ID file families instantiated from `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/`.

## Key Rules

- Update the canonical template before editing instantiated PS-ID files.
- Keep the enforced file families byte-identical across all 10 PS-IDs after placeholder substitution.
- Apply the template change and all 10 instantiations in the same semantic commit.
- Validate with `go run ./cmd/cicd-lint lint-fitness` and require `apps-ps-id-template` to pass.
- If a file family is no longer structurally identical across all 10 PS-IDs, remove it from exact template enforcement explicitly instead of allowing silent drift.

## Enforced Canonical Template Families

The current exact-match PS-ID template families are:

- `internal/apps/__PS_ID__/__SERVICE___usage.go`
- `internal/apps/__PS_ID__/__SERVICE___cli_test.go`
- `internal/apps/__PS_ID__/client/client.go`
- `internal/apps/__PS_ID__/README.md`
- `internal/apps/__PS_ID__/testmain_test.go`
- `internal/apps/__PS_ID__/server/__SERVICE___port_conflict_test.go`

## Workflow

1. Edit the canonical template under `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/`.
2. Propagate the equivalent instantiated change to every PS-ID file in that family.
3. Confirm there are still 10 instantiated files when the family is intended to cover all services.
4. Run `go run ./cmd/cicd-lint lint-fitness`.
5. Fix any `apps-ps-id-template` mismatch before touching unrelated code.

## Anti-Patterns

- Do not add one-off service variants when the file is supposed to stay template-instantiated.
- Do not change only a subset of PS-IDs for an enforced file family.
- Do not keep obsolete template files whose instantiated counterparts are intentionally removed.
- Do not rely on shared contract-test helpers to enforce consistency; use canonical templates plus linting.

## References

Read [ENG-HANDBOOK.md Section 10.3.5](../../../docs/ENG-HANDBOOK.md#1035-cross-service-ps-id-template-instantiation-pattern) for the project rule.
Read [apps_ps_id_template.go](../../../internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template.go) for the MANIFEST-driven validation logic.
Read [apps_ps_id_template_service_template.go](../../../internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/apps_ps_id_template_service_template.go) for the exact canonical file comparisons.
