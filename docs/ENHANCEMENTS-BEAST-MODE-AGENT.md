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

Items 3-5 are independent proposals, not a required implementation order. They can be applied in a different sequence if that produces a cleaner rewrite. The important constraint is conceptual consistency: the pre-edit rule, validation rule, and checklist shape should not contradict each other.

Because all 10 PS-IDs are supposed to reuse framework code heavily, the document should also avoid implying that the owning logic is usually in the package where a failure first appears. In this repo, a failing PS-ID test may be exposing a bug in shared framework code, shared test infrastructure, or a common builder path. The rewrite should therefore distinguish broad search from broad validation. It is still useful to avoid aimless searching, but it may be correct to validate at a broader scope earlier when the architecture strongly suggests shared ownership.

The rewrite should also account for concurrency-heavy test execution. Unit, integration, and e2e tests are expected to run with high concurrency, packages reuse TestMain to amortize expensive setup, and tests are supposed to stay independent by using non-conflicting data and other isolated resources. In that environment, one failing test may be the first visible symptom of shared-state conflicts, non-unique test data, port collisions, hostname collisions, noisy-neighbour effects, or other concurrency defects. The draft should therefore avoid implying that the first failing test is automatically the best or only place to look.

The handbook adds several related strategies that should also shape the rewrite. Tests are expected to run with `t.Parallel()`, shuffle, and race detection, so some failures are schedule-sensitive rather than input-sensitive. A small number of tests are allowed to be sequential, but only when they intentionally mutate package-level state and are marked with a `// Sequential:` reason. Integration tests rely on dynamic ports, localhost-only bindings, `TLSProvisionMode=auto`, and `DisableKeepAlives: true` for real-server clients, which means some apparent feature failures are actually setup, teardown, transport, or environment-parity defects. E2E coverage also validates cross-database behavior, so a failing path may be specific to shared PostgreSQL versus isolated SQLite instances rather than the API logic itself.

### 3. Add A First-Edit Hypothesis Rule

This should be reframed so that "local" means the nearest controlling abstraction, not necessarily the nearest file, the package where the failure surfaced, or the individual test case that failed first. In a framework-heavy and concurrency-heavy codebase, the first useful hypothesis may point at shared code, shared fixtures, or conflicting test data rather than PS-ID code.

This proposal should also explicitly replace or narrow the current "Read 2000+ lines before editing" style of instruction. If the broad-read mandate stays unchanged, the first-edit hypothesis rule will be undermined by a conflicting instruction that rewards over-exploration before action.

That would change the agent from broad exploratory searching to a focused, testable first move. Before the first edit, the agent would have to name one falsifiable hypothesis about where the behavior is actually controlled and one cheap check that could disconfirm it.

That changes behavior in three ways:

- stop broad searching sooner
- choose the nearest controlling abstraction, even if it is in the framework
- validate the smallest meaningful slice first, which may be a shared-fixture or concurrency check rather than a single-test rerun

Example: if a PS-ID test fails in a handler package, the agent should not assume the handler owns the bug. It should ask which abstraction actually decides the behavior. If the handler mostly wires framework resources, the first hypothesis may be, "The shared builder or middleware stack is producing this behavior," and the first check may target that framework path rather than the handler file.

Example: if a compile error points at a type mismatch in a service package, the agent may reasonably hypothesize that a shared framework signature changed and that the local package is only the first visible breakage. The cheap check is then the shared interface or constructor, not a broad scan of every PS-ID caller.

Example: if one integration test fails under concurrent package execution and the package uses TestMain to share a server or database, the first hypothesis may be, "This is a shared-fixture collision caused by non-orthogonal test data," not "this single test's logic is wrong." The cheap check may be to rerun the package, inspect whether multiple tests touch the same logical records, or verify that ports and hostnames are uniquely allocated before drilling into one assertion.

Example: if a failure appears only under `-shuffle=on` or `-race`, the first hypothesis may be that the defect is schedule-sensitive. The cheap check is then one that preserves ordering or concurrency pressure rather than an isolated rerun that removes it.

Example: if an integration binary hangs in teardown, the first hypothesis may be transport misuse such as missing `DisableKeepAlives: true`, not application logic. The cheap check should target teardown and client transport behavior before changing business code.

### 4. Reduce The Weight Of Global Checklists

This would replace a long, repetitive quality checklist with a compact execution ladder that the agent can actually follow while working. The intent is not to remove quality gates; it is to move the exhaustive detail out of the main contract and keep only the steps the agent must actively execute.

After reviewing the actual beast-mode file, this proposal should be broadened. The problem is not only the formal checklist section. The agent repeats checklist behavior across pre-flight rules, completion checklists, quality gates, blocker handling, work discovery, and review-pass language. The rewrite should therefore collapse duplicate validation obligations across the whole file, not just shorten one local checklist block.

The ladder should also be flexible enough to support framework-heavy and concurrency-heavy validation. It should not imply that a narrow per-package test or a single-test rerun is always the correct first executable check.

The shorter ladder would be:

- build
- focused test or architecture-scoped check
- broad test
- commit
- final clean status

Example: instead of repeating full coverage and mutation rules inside the agent body, the agent would say "run the cheapest meaningful executable check first; if the owning logic is probably shared, that first check may be framework-scoped rather than package-scoped." The detailed coverage and mutation targets would live in the handbook, not in the high-pressure prose.

Example: if the agent has already proven a local fix, the checklist should not force rereading the same full validation block in multiple places. The compact ladder keeps the order obvious without repeating the same requirement in several forms.

Example: if a failure appears in one PS-ID but the same path is instantiated across all 10 PS-IDs, the efficient first broad check may be a framework package test, a shared builder test, or a compile pass across all affected services. That is still consistent with a compact ladder, because the ladder is about validation order, not about forcing package-local tests.

Example: if a failure occurs in a package that uses TestMain and heavy parallel execution, the efficient first focused check may be a package-level rerun under the same concurrency conditions, or a targeted review of shared test data and fixture allocation, instead of immediately zooming into one failing test function. That is still a focused check because it is aimed at the most plausible failure class.

Example: if the failure only occurs in integration or e2e tests, the first focused check may need to preserve dynamic ports, TLS auto-provisioning, localhost binding, or the shared TestMain server. Replacing that with a simpler unit-style rerun may answer a different question than the one the failing test exposed.

Example: if the path is supposed to be cross-database compatible, the focused check may need to compare SQLite and PostgreSQL behavior rather than validating only the first backend that failed.

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
