# Enhancements For Copilot Skills

Created: 2026-05-17

## Summary

The skills catalog is useful, but it is missing one important capability and several existing skills are too narrowly focused on generation while under-serving triage and performance analysis.

## What The Catalog Already Does Well

- The skills are well grouped by task type.
- The `README.md` makes discovery straightforward.
- The testing skills are specific enough to be useful.
- The Copilot/Claude sync skill reduces drift risk.

## Missing Capability

The recent test-performance work exposed a gap: there is no dedicated skill for analyzing a slow or flaky test suite from logs and turning it into a ranked hotspot list.

Add a skill such as:

- `test-performance-analysis`

Suggested scope:

- parse `go test` output and `-json` logs
- rank packages, test families, and individual tests by runtime
- identify timeout-driven failures versus logic failures
- suggest the smallest validation command to confirm the hypothesis

That would make performance triage repeatable instead of ad hoc.

## Existing Skills That Should Be Improved

### `test-table-driven`

This is one of the most valuable skills, but it should do more to prevent slow or brittle test patterns.

Suggested additions:

- explicitly prefer in-memory handler tests over real listeners when possible
- flag real network listeners as a last resort for handler-only coverage
- mention bounded timeouts for lifecycle tests so suite-only flakiness is less likely
- call out suite flake triage when a test passes alone but fails in the full run

### `test-benchmark-gen`

The benchmark skill is solid, but it could be more useful if it also told users how to read benchmark regressions.

Suggested additions:

- note how to compare baseline versus current benchmark output
- mention common noise sources such as TLS setup, fixture generation, and GC pressure
- add a short reminder to benchmark only the code under test, not the setup path

### `sync-copilot-claude`

This skill is practical, but it should do more than body-drift checks.

Suggested additions:

- validate that new skills and agents are discoverable in the README index
- flag missing removal candidates when a skill has become redundant
- surface overlap between skills that could be merged or simplified

## Skills That Could Be Consolidated Or Trimmed

The catalog has several small scaffold-style helpers. They are useful, but the catalog should be reviewed for overlap:

- `agent-scaffold`
- `instruction-scaffold`
- `skill-scaffold`

These are distinct today, but the README could make it clearer which one to use when the user wants a new pattern versus a new executable agent.

## What To Remove Or De-Emphasize

- Remove duplicated explanation text from the catalog when the skill name and purpose already imply the behavior.
- De-emphasize long template blocks in the index; move them into the skill body where they are easier to maintain.
- Avoid treating every skill as equally important in the README. A short "most-used skills" section would help users find the practical entry points faster.

## Suggested Additions To The Catalog

1. `test-performance-analysis` for suite triage and log mining.
2. `test-isolation-debug` for tests that pass alone but fail in suite.
3. `fixture-minimization` for replacing heavyweight setup with smaller, shared fixtures.

## Net Effect

The skills ecosystem would cover the actual failure modes seen in this repo: slow suites, flaky lifecycle tests, repeated server setup, and drift across paired agent files.
