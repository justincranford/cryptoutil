# Skills Enhancements Review

Created: 2026-05-17
Last Updated: 2026-05-17

## Executive Summary

1. [Copilot Customization](#copilot-customization): merged `agent-scaffold`, `instruction-scaffold`, and `skill-scaffold` into one repo-local customization workflow and removed the redundant legacy directories.
2. [Test Table Driven](#test-table-driven): tightened the handler-versus-lifecycle boundary, added suite-flake triage guidance, and corrected the heavyweight `TestMain` example toward shared SQLite helpers.
3. [Test Benchmark Gen](#test-benchmark-gen): removed benchmark-harness noise from the timed example and added baseline-versus-current regression reading guidance.
4. [Propagation Check](#propagation-check): replaced an invalid `lint-docs validate-propagation` command with the real `lint-docs` entrypoint and removed Python-specific inspection guidance.
5. [Sync Copilot Claude](#sync-copilot-claude): expanded the skill from body-drift repair to include catalog discoverability checks, overlap retirement, and same-commit cleanup expectations.
6. [Coverage Analysis](#coverage-analysis): removed shell-heavy coverage steps, narrowed seam guidance to the repository's allowed exception pattern, and clarified that follow-on test authoring belongs in `test-table-driven`.
7. [Test Fuzz Gen](#test-fuzz-gen): clarified that fuzz build tags are optional and should not be presented as the default pattern, while keeping the skill distinct from deterministic test generation.
8. [Fitness Function Gen](#fitness-function-gen): removed stale linter-count and registry references, repaired malformed code fences, and tightened the scope boundary versus `psid-template-sync`.
9. [New Service](#new-service): corrected the service port catalog, removed Bash-only copy and rename snippets, and clarified where migration, OpenAPI, and customization work belongs in adjacent skills.
10. [FIPS Audit](#fips-audit): reviewed for handbook compliance and overlap; no edits were required because the skill already maps cleanly to the FIPS policy surface.
11. [Migration Create](#migration-create): reviewed for handbook compliance and overlap; no edits were required because the skill stays focused on migration-file mechanics rather than whole-service setup.
12. [OpenAPI Codegen](#openapi-codegen): reviewed for handbook compliance and overlap; no edits were required because the skill remains a focused OpenAPI/config generation helper.
13. [PSID Template Sync](#psid-template-sync): reviewed for handbook compliance and overlap; no edits were required because the skill already has a tight boundary around exact-match template families.

## Copilot Customization

- Compliance review: The old scaffold family was structurally correct in isolation, but the catalog violated the repo's own simplification goal because three near-identical creation helpers fragmented one workflow.
- Overlap review: `agent-scaffold`, `instruction-scaffold`, and `skill-scaffold` all scaffolded repo-local customization artifacts and competed with `sync-copilot-claude` on when to create mirrored Claude files.
- Issues found: redundant catalog entries, duplicated creation logic, and unnecessary decision overhead in the README and command tables.
- Fixes applied: created `copilot-customization`, removed the three legacy skill pairs, updated the README, `CLAUDE.md`, `.github/copilot-instructions.md`, `docs/ENG-HANDBOOK.md`, and `docs/target-structure.md`, and added a "most-used skills" section to de-emphasize long flat indexes.
- Final scope boundary: `copilot-customization` creates new repo-local customization artifacts; `sync-copilot-claude` audits or repairs existing dual-canonical drift.

## Test Table Driven

- Compliance review: The skill matched the testing handbook at a high level, but one example still normalized heavyweight container setup inside a generic `TestMain` pattern and did not explicitly encode suite-flake triage.
- Overlap review: The skill risked overlap with `test-fuzz-gen` and infrastructure-oriented lifecycle testing because it did not clearly separate handler-only coverage from transport-level coverage.
- Issues found: missing guidance on when real listeners are justified, no explicit package-versus-isolated rerun advice for flaky tests, and an example that leaned toward heavyweight setup instead of the mandated in-memory SQLite defaults.
- Fixes applied: added the handler-versus-listener decision boundary, added bounded-timeout and suite-flake triage guidance, and replaced the illustrative `TestMain` pattern with the shared in-memory SQLite helper flow.
- Final scope boundary: use this skill for deterministic Go test structure and validation strategy; move to `test-fuzz-gen` for fuzzing and to package-specific lifecycle tests only when `app.Test()` cannot cover the behavior.

## Test Benchmark Gen

- Compliance review: The benchmark skill was structurally aligned with the handbook, but one example measured UUID generation inside the timed loop and therefore encouraged harness noise rather than the target operation.
- Overlap review: The skill did not overlap heavily with other skills, but it under-served the user's requested performance-triage angle by not explaining how to interpret regressions.
- Issues found: noisy example benchmark, no baseline-versus-current comparison guidance, and no reminder about common benchmark-noise sources.
- Fixes applied: removed UUID creation from the timed example, added an explicit "Reading Regressions" section, and documented TLS setup, fixture generation, GC pressure, and mismatched benchmark scopes as noise sources.
- Final scope boundary: use this skill to write and interpret benchmarks; use test-performance tooling or direct benchmark runs to collect the raw numbers.

## Propagation Check

- Compliance review: The skill referenced a non-existent `lint-docs validate-propagation` subcommand even though the repo's validator is the aggregate `lint-docs` command.
- Overlap review: The skill partially duplicated `sync-copilot-claude` by wandering into general multi-file coordination rather than staying anchored on propagation markers.
- Issues found: invalid command guidance and Python-specific manual parsing that was both unnecessary and off-policy relative to the repository's preference for Go-first tooling guidance.
- Fixes applied: replaced the command guidance with `go run ./cmd/cicd-lint lint-docs`, removed the Python inspection block, and clarified that `sync-copilot-claude` is the adjacent skill when a propagation change also affects mirrored agent or skill files.
- Final scope boundary: use this skill for `@propagate` and `@source` integrity; use `sync-copilot-claude` for pair drift that extends beyond propagation markers.

## Sync Copilot Claude

- Compliance review: The skill already satisfied the basic dual-canonical rule, but it under-described the repo's catalog maintenance requirements after a sync or merge.
- Overlap review: Before the merge, it overlapped weakly with the scaffold skills by partly describing missing-file creation without describing when catalog entries or redundant skills had to be removed.
- Issues found: shell- and Python-heavy examples, no explicit catalog discoverability review, and no instruction to retire redundant skills in the same change after a merge.
- Fixes applied: removed the platform-specific audit snippets, added catalog review requirements, added overlap-retirement guidance, and made same-commit cleanup explicit.
- Final scope boundary: use this skill to synchronize existing pairs and surrounding catalog references after edits; use `copilot-customization` to create a new skill or agent pair from scratch.

## Coverage Analysis

- Compliance review: The skill aligned with handbook coverage targets, but its workflow used shell-centric commands and its seam advice drifted toward broader package-level seam patterns than the repo usually permits.
- Overlap review: It blurred into `test-table-driven` by implying it would also supply the test-authoring pattern rather than focusing on analysis.
- Issues found: shell-specific setup steps, overly broad seam guidance, and no explicit handoff to the deterministic test-authoring skill.
- Fixes applied: simplified the workflow to core Go coverage commands, narrowed seam guidance to the restricted `osExit` exception plus function-parameter or `export_test.go` seams, and explicitly routed follow-on test authoring to `test-table-driven`.
- Final scope boundary: use this skill to rank and classify coverage gaps; use `test-table-driven` to implement the tests that close them.

## Test Fuzz Gen

- Compliance review: The skill correctly cited the 15-second minimum fuzz time and unique-name rule, but it presented `//go:build fuzz` as the default template even though the handbook treats tagged fuzz-only files as a special-case pattern.
- Overlap review: Without a clearer boundary, the skill could be confused with `test-table-driven` for deterministic parser examples.
- Issues found: default template implied a fuzz build tag was the standard case and the purpose section did not explain the relationship to deterministic example coverage.
- Fixes applied: changed the guidance so build tags are opt-in rather than default, removed the build tag from the baseline example, and clarified that deterministic example coverage belongs in `test-table-driven`.
- Final scope boundary: use this skill for fuzz-specific corpus and invariant design; use `test-table-driven` for ordinary example-driven test coverage.

## Fitness Function Gen

- Compliance review: The skill contained stale numerical claims, a stale registry-file reference, malformed fenced code blocks, and examples that no longer matched the repo's emphasis on testable filesystem seams.
- Overlap review: It needed a sharper line against `psid-template-sync`, because not every cross-service structural change requires a new fitness linter.
- Issues found: stale checker-count language, stale `lint-fitness-registry.yaml` mention, broken markdown fences, and imprecise seam guidance.
- Fixes applied: removed the stale count and registry-file references, repaired fenced examples, updated the seam guidance to prefer `fs.FS`, `io.Reader`, or explicit function parameters, and added a direct scope note pointing template-only work to `psid-template-sync`.
- Final scope boundary: use this skill when the repo needs a new enforced invariant; use `psid-template-sync` when the work is only a canonical-template propagation problem.

## New Service

- Compliance review: The skill contained stale service-port examples and several Bash-only copy or rename snippets that were a poor fit for the repo's cross-platform guidance.
- Overlap review: The skill mixed whole-service rollout steps with migration, OpenAPI, and customization work that each already has a focused adjacent skill.
- Issues found: incorrect service-catalog rows, shell-centric cloning instructions, and weak routing to adjacent specialist skills.
- Fixes applied: corrected the JOSE, PKI, and Identity port examples, removed Bash-only command blocks in favor of repo-aware procedural steps, and explicitly routed migration, OpenAPI, and customization work to `migration-create`, `openapi-codegen`, and `copilot-customization`.
- Final scope boundary: use this skill for end-to-end service instantiation from `skeleton-template`; use the narrower skills when the task is only one slice of that rollout.

## FIPS Audit

- Compliance review: The skill already matched the handbook's FIPS-approved and banned-algorithm policy, TLS minimums, and algorithm-agility requirements.
- Overlap review: The skill stays distinct from `fips`-style linting because it explains how to reason about violations and fixes rather than only enforcing static checks.
- Issues found: none that required a content change.
- Fixes applied: none.
- Final scope boundary: use this skill for cryptographic policy review and remediation guidance, not for general code review or test generation.

## Migration Create

- Compliance review: The skill already matched the handbook's migration-range, up/down pairing, cross-database SQL, and builder-registration guidance.
- Overlap review: The skill remains cleanly narrower than `new-service`, because it covers migration-file mechanics rather than full service rollout.
- Issues found: none that required a content change.
- Fixes applied: none.
- Final scope boundary: use this skill when the task is a migration artifact or migration-numbering change; use `new-service` when the migration is only one step inside a larger service bootstrap.

## OpenAPI Codegen

- Compliance review: The skill already matched the handbook on OpenAPI 3.0.3, strict server generation, dual `/service/` and `/browser/` paths, and standard pagination or error-schema conventions.
- Overlap review: The skill stays distinct from `new-service` by focusing on API/config generation rather than overall service setup.
- Issues found: none that required a content change.
- Fixes applied: none.
- Final scope boundary: use this skill for OpenAPI specs and codegen configuration, not for whole-service provisioning.

## PSID Template Sync

- Compliance review: The skill already matched the handbook's exact-template-family enforcement and same-commit propagation rules.
- Overlap review: The skill stays distinct from `fitness-function-gen` because it assumes the invariant already exists and the work is to keep instantiations aligned with the canonical template.
- Issues found: none that required a content change.
- Fixes applied: none.
- Final scope boundary: use this skill for template-instantiated file families already governed by lint-fitness, not for designing a new invariant.

## Document Review

- Completeness check: this review covers every current Copilot skill in `.github/skills/` after the scaffold merge.
- Correctness check: every recorded fix maps to a file change made in this session or an explicit no-change review result.
- Overlap check: each section records the neighboring skill boundary so the catalog does not collapse back into broad, redundant helpers.
- Consistency check: the merged scaffold inventory is described as 13 current skills, matching the updated catalog and target-structure document.
