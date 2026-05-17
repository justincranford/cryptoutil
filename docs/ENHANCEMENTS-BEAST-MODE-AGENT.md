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

### 3. Add A First-Edit Hypothesis Rule

This would change the agent from broad exploratory searching to a focused, testable first move. Before the first edit, the agent would have to name one falsifiable local hypothesis and one cheap check that could disconfirm it.

That changes behavior in three ways:

- stop broad searching sooner
- choose the nearest controlling abstraction
- validate the smallest meaningful slice first

Example: if a test fails in a handler package, the agent should not immediately scan the whole repository. It should say something like, "The failure is likely in the input-to-model mapping in this handler," then run the nearest focused test or inspect the nearest constructor before widening scope.

Example: if a compile error points at a type mismatch, the agent should hypothesize the mismatch is caused by a nearby signature change, check that file first, and only then search callers.

### 4. Reduce The Weight Of Global Checklists

This would replace a long, repetitive quality checklist with a compact execution ladder that the agent can actually follow while working. The intent is not to remove quality gates; it is to move the exhaustive detail out of the main contract and keep only the steps the agent must actively execute.

The shorter ladder would be:

- build
- narrow test
- broad test
- commit
- final clean status

Example: instead of repeating full coverage and mutation rules inside the agent body, the agent would say "run a narrow test first; if the hypothesis survives, widen to the package or suite." The detailed coverage and mutation targets would live in the handbook, not in the high-pressure prose.

Example: if the agent has already proven a local fix, the checklist should not force rereading the same full validation block in multiple places. The compact ladder keeps the order obvious without repeating the same requirement in several forms.

### 5. Make The Validation Order Explicit

This would make the post-edit workflow deterministic. After the first substantive edit, the very next step must be the cheapest executable validation that can falsify the current hypothesis. That is the control rule that prevents speculative widening and unnecessary reruns.

Example: if the first edit changes a parser, the next step should be the narrowest parser test or the smallest compile target that can fail for the right reason. It should not jump straight to a full suite unless the narrow check is unavailable.

Example: if the first edit changes validation branching, the next step should be a focused test that exercises that branch. If the test passes, the agent can widen to the next level of confidence; if it fails, the agent should repair that same slice before moving on.

## What To Remove Or De-Emphasize

This section is a cleanup summary, not the final agent contract. It names the kinds of content that should shrink, move, or disappear in the rewrite.

- Remove repeated prose that says the same thing in different wording. Example: if the agent says "keep working" in four different ways, collapse that into one rule.
- De-emphasize token-budget style commentary. Example: lines about rate limits or token pressure should not compete with the execution rules the agent actually needs.
- Trim the giant pre-flight section to the pieces that are truly universal. Example: keep the baseline check and validation order, but move repo-specific policy references out of the main body.
- Move repository-specific CI hook policy out of the main agent body unless the task is actually about CI. Example: bulk-hook architecture belongs in the handbook or a CI-focused agent, not in a general autonomous-work contract.

## Suggested Shape For A Better Version

This is a blueprint for the rewrite, not an instruction to the current agent. It describes the structure the improved version should have.

1. One short execution contract. Example: "Work continuously, do not interrupt, and commit each coherent unit."
2. One short pre-edit hypothesis rule. Example: "Before editing, name one local hypothesis and one cheap disconfirming check."
3. One short post-edit validation rule. Example: "After the first substantive edit, run the cheapest check that can falsify the hypothesis."
4. One compact quality gate ladder. Example: "build -> narrow test -> broad test -> commit -> final clean status."
5. One compact anti-pattern list. Example: "Do not broaden scope before the narrow check fails or passes."
6. Links to the handbook for anything longer than a paragraph. Example: put full coverage, mutation, and CI policy in the handbook instead of repeating them here.

## Net Effect

The agent would become faster to read, easier to follow, and less likely to bury the important instruction under repeated emphasis.
