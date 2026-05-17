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

### 2. Separate Contract From Policy

The agent currently mixes three different layers:

- execution behavior
- quality gates
- repository-specific CI policy

Those should be separated more cleanly. The agent should define the contract, then link to the repo handbook for the details. The current bulk-hook architecture block is a good example of content that belongs in the handbook or a dedicated CI reference, not in the main autonomy file.

### 3. Add A First-Edit Hypothesis Rule

The strongest missing rule is the one already implied by the developer instructions: before the first edit, state one falsifiable local hypothesis and one cheap check that could disconfirm it.

That would improve speed and quality at the same time because it forces the agent to:

- stop broad searching sooner
- choose the nearest controlling abstraction
- validate the smallest meaningful slice first

### 4. Reduce The Weight Of Global Checklists

The current file contains a very large quality checklist. The intent is good, but a long checklist often becomes noise in practice.

Prefer a shorter ladder:

- build
- narrow test
- broad test
- commit
- final clean status

Then link to the handbook for the full coverage/mutation rules.

### 5. Make The Validation Order Explicit

The agent should say, in one place, that after the first substantive edit the next step must be the cheapest executable validation that can falsify the current hypothesis. That is the best defense against speculative widening and unnecessary reruns.

## What To Remove Or De-Emphasize

- Remove repeated prose that says the same thing in different wording.
- De-emphasize token-budget style commentary. It does not help the actual work and can dilute focus.
- Trim the giant pre-flight section to the pieces that are truly universal.
- Move repository-specific CI hook policy out of the main agent body unless the task is actually about CI.

## Suggested Shape For A Better Version

1. One short execution contract.
2. One short pre-edit hypothesis rule.
3. One short post-edit validation rule.
4. One compact quality gate ladder.
5. One compact anti-pattern list.
6. Links to the handbook for anything longer than a paragraph.

## Net Effect

The agent would become faster to read, easier to follow, and less likely to bury the important instruction under repeated emphasis.
