# Speckit Quick Guide

## Background

<https://github.com/github/spec-kit>

## Steps

| Step | Output | Notes |
|------|--------|-------|
| 1. /speckit.constitution | .specify\memory\constitution.md | |
| 2. /speckit.specify | specs\002-cryptoutil\spec.md | |
| 3. /speckit.clarify | specs\002-cryptoutil\clarify.md and specs\002-cryptoutil\CLARIFY-QUIZME.md | (optional: after specify, before plan) |
| 4. /speckit.plan | specs\002-cryptoutil\plan.md | |
| 5. /speckit.tasks | specs\002-cryptoutil\tasks.md | |
| 6. /speckit.analyze | specs\002-cryptoutil\analyze.md | (optional: after tasks, before implement) |
| 7. /speckit.implement | (e.g., implement/DETAILED.md and implement/EXECUTIVE.md) | |

## How It's Going

While working through speckit, I spend the vast majority of my time in step 7, /speckit.implement.

During /speckit.implement, I have to frequently make adjustments because LLM Agents diverge from my desired outcomes:

- small adjustments: frequent
- medium adjustments: occasionally
- large adjustments: sometimes but rare

## Assumptions

### LLM Agent Limitations

I assume that LLM Agents diverge from my desired outputs because of many factors:

- lack of domain expertise, especially:
  - LLM was trained on data up to 2023||2024, which predates modern speckit methodology in 2025||2026
  - LLM is general purpose, instead of specializing in agentic coding, software design, and specific programming languages
  - insufficient understanding of my project's architecture
- omissions (e.g. incomplete or outdated context)
- ambiguities
- misinterpretations
- hallucinations (e.g. generating plausible but incorrect information)
- context window truncation (e.g. token limits lead to truncation of long documents and lost details)
- over-reliance on familiar patterns (e.g. defaulting to familiar code structures instead of your specific requirements)
- bias toward completion (e.g. prioritizing finishing tasks over accuracy)

### Human Software Methodologies

Retrospectives and post mortems are very useful steps for capturing:

- what went well
- what needs improvement
- what to keep doing
- what to stop doing

Retrospectives capture that important after implement increments (e.g. Agile/SCRUM sprint). Post Mortems capture additional important content at milestone increments (e.g. product release, SDLC program increment).

Both sources of feedback allow continuous improvement of earlier, higher-level steps in SDLC process.
Improvements and hard lessens learned are applied, with the goal to avoid or mitigate problems in subsequent SDLC increments.

## Speckit Methodology Challenges

I'm working with the Speckit methodology for AI-assisted software development (available at <https://github.com/github/spec-kit>). Speckit involves sequential steps as shown in the table below:

| Step | Output | Notes |
|------|--------|-------|
| 1. /speckit.constitution | .specify\memory\constitution.md | |
| 2. /speckit.specify | specs\002-cryptoutil\spec.md | |
| 3. /speckit.clarify | specs\002-cryptoutil\clarify.md and specs\002-cryptoutil\CLARIFY-QUIZME.md | (optional: after specify, before plan) |
| 4. /speckit.plan | specs\002-cryptoutil\plan.md | |
| 5. /speckit.tasks | specs\002-cryptoutil\tasks.md | |
| 6. /speckit.analyze | specs\002-cryptoutil\analyze.md | (optional: after tasks, before implement) |
| 7. /speckit.implement | (e.g., implement/DETAILED.md and implement/EXECUTIVE.md) | |

During implementation (step 7), I frequently adjust for LLM agent divergences (e.g., small tweaks often, medium occasionally, large rarely). I track progress in DETAILED.md (Section 1: task checklist from tasks.md; Section 2: append-only timeline of completions) and EXECUTIVE.md (summaries of progress, issues, lessons learned, workarounds, mitigations, risks, and clarifications).

**The Problem:** I want to implement feedback loops by "bubbling" insights from EXECUTIVE.md and DETAILED.md back to earlier steps (e.g., updating constitution.md, spec.md, clarify.md, or plan.md). My goal is to copy the current Speckit directory, truncate it after step 2 or 4, and restart from that point while preserving critical content like issues, lessons learned, workarounds, mitigations, risks, and clarifications.

However, every attempt fails—important context is lost, and the restarted process lacks the depth from implementation. I expected reused files (e.g., constitution.md and spec.md) to retain this feedback, but they don't. This breaks continuity and forces rework.

**Questions:**

1. How do feedback loops work in Speckit methodology? Are there established patterns or best practices?
2. What should I do differently during or after step 7 (implement) to ensure content preservation?
3. How can I effectively integrate EXECUTIVE.md and DETAILED.md insights into earlier files like constitution.md and spec.md?
4. Provide examples of modifications to my process or file structures to enable successful restarts.

**Key Insight: Real Development is Iterative**
Speckit presents a sequential workflow, but real software development is inherently iterative. Constitution, spec, and plan are never truly "complete"—they evolve through feedback loops from implementation mistakes, lessons learned, and applied mitigations. This creates tension: Speckit assumes prerequisites are finalized, but iteration requires revisiting and improving them. The challenge is balancing Speckit's structure with practical iteration.

Please structure your response with: 1) Analysis of the issue, 2) Concrete recommendations, 3) Examples or templates, and 4) Potential pitfalls to avoid.
