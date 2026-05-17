# Enhancements For `claude-beast-mode`

Created: 2026-05-17

## Summary

The current beast-mode agent is directionally correct, but it is too repetitive and too expansive for the amount of decision support it actually adds. After reviewing the actual `copilot-beast-mode` file, the deeper problem is not only repetition. The contract also mixes core autonomy rules, repository policy, language-specific implementation details, testing doctrine, and tool-behavior commands into one very large body. Some of those instructions reinforce each other, but others compete or create ambiguity about what the agent should optimize for first.

The best improvements are to compress the contract, reduce duplication, separate core execution rules from repository-specific detail, and replace broad pressure statements with a smaller number of sharp execution rules that resolve conflicts instead of adding more prose.

## What To Keep

- The no-interruptions contract is valuable and should stay.
- The baseline gate and clean-worktree requirement are useful.
- The insistence on evidence-based validation is correct.
- The commit-after-task pattern is a good default for autonomous work.

## What The Current Draft Missed

- The actual agent already separates some repository-specific material into a trailing reference section, so the problem is broader than CI hook policy alone. Repo-specific and Go-specific execution detail still appears throughout the main body.
- The agent tells the model both to plan extensively and to keep zero text between tools. Those instructions compete with each other and should be reconciled, not merely shortened.
- The agent tells the model to read 2000+ lines before editing, which directly weakens any proposed first-edit hypothesis rule unless the rewrite explicitly narrows or replaces that mandate.
- The agent contains several absolute workflow slogans such as "every discrete work unit -> commit" and "zero text between tools" that are stronger than the surrounding execution logic and may be counterproductive in practice.
- The current enhancement draft over-focuses on sections 3-5. Those sections still matter, but they are not the only places where the agent needs restructuring.

## What To Modify

Before the specific proposed edits below, the rewrite should explicitly separate three layers that are currently blended together:

1. Core autonomy contract.
2. Default execution heuristics.
3. Repository-specific validation and policy references.

Without that separation, improvements to sections 3-5 will help, but the file will still feel overloaded and internally competitive.

Item 5 remains the active proposal. Items 3-4 are complete and documented in [docs/ENHANCEMENTS-BEAST-MODE-AGENT-COMPLETED.md](docs/ENHANCEMENTS-BEAST-MODE-AGENT-COMPLETED.md).

Because all 10 PS-IDs are supposed to reuse framework code heavily, the document should also avoid implying that the owning logic is usually in the package where a failure first appears. In this repo, a failing PS-ID test may be exposing a bug in shared framework code, shared test infrastructure, or a common builder path. The rewrite should therefore distinguish broad search from broad validation. It is still useful to avoid aimless searching, but it may be correct to validate at a broader scope earlier when the architecture strongly suggests shared ownership.

The rewrite should also account for concurrency-heavy test execution. Unit, integration, and e2e tests are expected to run with high concurrency, packages reuse TestMain to amortize expensive setup, and tests are supposed to stay independent by using non-conflicting data and other isolated resources. In that environment, one failing test may be the first visible symptom of shared-state conflicts, non-unique test data, port collisions, hostname collisions, noisy-neighbour effects, or other concurrency defects. The draft should therefore avoid implying that the first failing test is automatically the best or only place to look.

The handbook adds several related strategies that should also shape the rewrite. Tests are expected to run with `t.Parallel()`, shuffle, and race detection, so some failures are schedule-sensitive rather than input-sensitive. A small number of tests are allowed to be sequential, but only when they intentionally mutate package-level state and are marked with a `// Sequential:` reason. Integration tests rely on dynamic ports, localhost-only bindings, `TLSProvisionMode=auto`, and `DisableKeepAlives: true` for real-server clients, which means some apparent feature failures are actually setup, teardown, transport, or environment-parity defects. E2E coverage also validates cross-database behavior, so a failing path may be specific to shared PostgreSQL versus isolated SQLite instances rather than the API logic itself.

### 5. Make The Validation Order Explicit

This would make the post-edit workflow deterministic without hard-coding a package-local bias. After the first substantive edit, the very next step must be the cheapest executable validation that can falsify the current hypothesis. That hypothesis may be local to a framework abstraction, a shared fixture pattern, or a concurrency failure mode rather than local to the package where the failure appeared.

This proposal should also explicitly overrule weaker or competing slogans in the current agent, especially instructions like "zero text between tools" and "commit after each discrete work unit" when those would interrupt a tighter edit-then-validate loop. The validation-order rule is only useful if it is clearly higher priority than generic momentum rules.

Example: if the first edit changes a shared framework parser or builder, the next step should be the smallest framework-scoped test or compile target that can falsify that change. It may be broader than a single PS-ID package and still be the cheapest correct check.

Example: if the first edit changes validation branching in code reused by all services, the next step should be a focused test that exercises the shared branch, not necessarily the first failing PS-ID test. If that test passes, the agent can widen to the next level of confidence; if it fails, the agent should repair that same slice before moving on.

Example: if architecture strongly suggests the bug is in shared test infrastructure, a broader early validation step can be more efficient than repeatedly rerunning one surface-level PS-ID test. The rule should be "cheapest discriminating check," not "smallest file-local check."

Example: if the first edit changes test data generation, fixture allocation, port assignment, or hostname isolation, the next step should preserve the concurrency conditions that exposed the problem. A single isolated rerun that removes the conflict may hide the bug rather than falsify the hypothesis.

Example: if the hypothesis is that a test should actually be marked `// Sequential:` because it mutates package-level state, the next validation step should check that specific global-state interaction. The agent should not force a parallel-first assumption when the handbook explicitly allows a documented sequential exemption.

Example: if the hypothesis is that a race-only or shuffle-only failure was fixed, the next validation step should include the same stress mode that exposed it. A plain happy-path rerun is weaker evidence than a targeted probabilistic check.

## What To Remove Or De-Emphasize

This section is a cleanup summary, not the final agent contract. It names the kinds of content that should shrink, move, or disappear in the rewrite.

- Remove repeated prose that says the same thing in different wording. Example: if the agent says "keep working" in four different ways, collapse that into one rule.
- De-emphasize token-budget style commentary. Example: lines about rate limits or token pressure should not compete with the execution rules the agent actually needs.
- Trim the giant pre-flight section to the pieces that are truly universal. Example: keep the baseline check and validation order, but move repo-specific policy references out of the main body.
- Remove or downgrade overly rigid slogans that compete with better local judgment. Example: "zero text between tools" and "commit after each discrete work unit" are too absolute compared with the more useful goal of maintaining momentum with evidence.
- Move repository-specific operational detail out of the core contract unless it directly changes autonomous behavior. Example: bulk-hook architecture, line-ending recovery, and detailed Go command matrices belong in handbook references or task-specific addenda, not in the main autonomy rules.
- De-emphasize broad mandatory-reading language when a narrower routing rule would do better. Example: a local hypothesis rule is more actionable than a blanket "read 2000+ lines before editing" directive.

## Suggested Shape For A Better Version

This is a blueprint for the rewrite, not an instruction to the current agent. It describes the structure the improved version should have, and the items can be reordered if that makes the contract read more clearly.

1. One short execution contract. Example: "Work continuously, do not interrupt, and commit each coherent unit."
2. One short pre-edit hypothesis rule. Example: "Before editing, name one controlling-abstraction or shared-fixture hypothesis and one cheap disconfirming check."
3. One short post-edit validation rule. Example: "After the first substantive edit, run the cheapest check that can falsify the hypothesis, whether that check is package-scoped, framework-scoped, or concurrency-scoped."
4. One compact quality gate ladder. Example: "build -> focused, architecture-scoped, or concurrency-scoped check -> broad test -> commit -> final clean status."
5. One short precedence rule for conflicts. Example: "When momentum rules conflict with falsification or validation rules, validation wins."
6. One compact anti-pattern list. Example: "Do not broaden scope before the narrow check fails or passes."
7. Links to the handbook for anything longer than a paragraph. Example: put full coverage, mutation, and CI policy in the handbook instead of repeating them here.

The anti-pattern examples should explicitly include: treating the first failing test as the owner by default, converting a concurrency failure into a sequential rerun too early, stripping away TestMain or environment-parity conditions before validating a hypothesis, ignoring cross-database or transport-specific failure modes, and letting generic workflow slogans outrank better local falsification logic.

## Net Effect

The agent would become faster to read, easier to follow, and less likely to bury the important instruction under repeated emphasis. More importantly, it would stop forcing the model to choose between competing absolutes such as broad reading, zero-text momentum, immediate commits, exhaustive checklists, and local validation. The rewrite would give the agent a clearer order of operations instead of more volume.
