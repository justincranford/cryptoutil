# Enhancements For `claude-beast-mode`

Created: 2026-05-17

## Summary

The current beast-mode agent is directionally correct, but it is too repetitive and too expansive for the amount of decision support it actually adds. The best improvements are to compress the contract, reduce duplication, and replace broad pressure statements with a smaller number of sharp execution rules.

## What To Keep

- The no-interruptions contract is valuable and should stay.
- The baseline gate and clean-worktree requirement are useful.
- The insistence on evidence-based validation is correct.
- The commit-after-task pattern is a good default for autonomous work.

## What To Modify

Items 3-5 are independent proposals, not a required implementation order. They can be applied in a different sequence if that produces a cleaner rewrite. The important constraint is conceptual consistency: the pre-edit rule, validation rule, and checklist shape should not contradict each other.

Because all 10 PS-IDs are supposed to reuse framework code heavily, the document should also avoid implying that the owning logic is usually in the package where a failure first appears. In this repo, a failing PS-ID test may be exposing a bug in shared framework code, shared test infrastructure, or a common builder path. The rewrite should therefore distinguish broad search from broad validation. It is still useful to avoid aimless searching, but it may be correct to validate at a broader scope earlier when the architecture strongly suggests shared ownership.

### 3. Add A First-Edit Hypothesis Rule

This should be reframed so that "local" means the nearest controlling abstraction, not necessarily the nearest file or the package where the failure surfaced. In a framework-heavy codebase, the first useful hypothesis may point at shared code rather than PS-ID code.

That would change the agent from broad exploratory searching to a focused, testable first move. Before the first edit, the agent would have to name one falsifiable hypothesis about where the behavior is actually controlled and one cheap check that could disconfirm it.

That changes behavior in three ways:

- stop broad searching sooner
- choose the nearest controlling abstraction, even if it is in the framework
- validate the smallest meaningful slice first

Example: if a PS-ID test fails in a handler package, the agent should not assume the handler owns the bug. It should ask which abstraction actually decides the behavior. If the handler mostly wires framework resources, the first hypothesis may be, "The shared builder or middleware stack is producing this behavior," and the first check may target that framework path rather than the handler file.

Example: if a compile error points at a type mismatch in a service package, the agent may reasonably hypothesize that a shared framework signature changed and that the local package is only the first visible breakage. The cheap check is then the shared interface or constructor, not a broad scan of every PS-ID caller.

### 4. Reduce The Weight Of Global Checklists

This would replace a long, repetitive quality checklist with a compact execution ladder that the agent can actually follow while working. The intent is not to remove quality gates; it is to move the exhaustive detail out of the main contract and keep only the steps the agent must actively execute.

The ladder should also be flexible enough to support framework-heavy validation. It should not imply that a narrow per-package test is always the correct first executable check.

The shorter ladder would be:

- build
- focused test or architecture-scoped check
- broad test
- commit
- final clean status

Example: instead of repeating full coverage and mutation rules inside the agent body, the agent would say "run the cheapest meaningful executable check first; if the owning logic is probably shared, that first check may be framework-scoped rather than package-scoped." The detailed coverage and mutation targets would live in the handbook, not in the high-pressure prose.

Example: if the agent has already proven a local fix, the checklist should not force rereading the same full validation block in multiple places. The compact ladder keeps the order obvious without repeating the same requirement in several forms.

Example: if a failure appears in one PS-ID but the same path is instantiated across all 10 PS-IDs, the efficient first broad check may be a framework package test, a shared builder test, or a compile pass across all affected services. That is still consistent with a compact ladder, because the ladder is about validation order, not about forcing package-local tests.

### 5. Make The Validation Order Explicit

This would make the post-edit workflow deterministic without hard-coding a package-local bias. After the first substantive edit, the very next step must be the cheapest executable validation that can falsify the current hypothesis. That hypothesis may be local to a framework abstraction, not local to the package where the failure appeared.

Example: if the first edit changes a shared framework parser or builder, the next step should be the smallest framework-scoped test or compile target that can falsify that change. It may be broader than a single PS-ID package and still be the cheapest correct check.

Example: if the first edit changes validation branching in code reused by all services, the next step should be a focused test that exercises the shared branch, not necessarily the first failing PS-ID test. If that test passes, the agent can widen to the next level of confidence; if it fails, the agent should repair that same slice before moving on.

Example: if architecture strongly suggests the bug is in shared test infrastructure, a broader early validation step can be more efficient than repeatedly rerunning one surface-level PS-ID test. The rule should be "cheapest discriminating check," not "smallest file-local check."

## What To Remove Or De-Emphasize

This section is a cleanup summary, not the final agent contract. It names the kinds of content that should shrink, move, or disappear in the rewrite.

- Remove repeated prose that says the same thing in different wording. Example: if the agent says "keep working" in four different ways, collapse that into one rule.
- De-emphasize token-budget style commentary. Example: lines about rate limits or token pressure should not compete with the execution rules the agent actually needs.
- Trim the giant pre-flight section to the pieces that are truly universal. Example: keep the baseline check and validation order, but move repo-specific policy references out of the main body.
- Move repository-specific CI hook policy out of the main agent body unless the task is actually about CI. Example: bulk-hook architecture belongs in the handbook or a CI-focused agent, not in a general autonomous-work contract.

## Suggested Shape For A Better Version

This is a blueprint for the rewrite, not an instruction to the current agent. It describes the structure the improved version should have, and the items can be reordered if that makes the contract read more clearly.

1. One short execution contract. Example: "Work continuously, do not interrupt, and commit each coherent unit."
2. One short pre-edit hypothesis rule. Example: "Before editing, name one controlling-abstraction hypothesis and one cheap disconfirming check."
3. One short post-edit validation rule. Example: "After the first substantive edit, run the cheapest check that can falsify the hypothesis, whether that check is package-scoped or framework-scoped."
4. One compact quality gate ladder. Example: "build -> focused or architecture-scoped check -> broad test -> commit -> final clean status."
5. One compact anti-pattern list. Example: "Do not broaden scope before the narrow check fails or passes."
6. Links to the handbook for anything longer than a paragraph. Example: put full coverage, mutation, and CI policy in the handbook instead of repeating them here.

## Net Effect

The agent would become faster to read, easier to follow, and less likely to bury the important instruction under repeated emphasis.
