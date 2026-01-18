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

While working through speckit, I see LLM Agents spending the vast majority of time in step 7, /speckit.implement.

I have to make frequent adjustments during step 7, because LLM Agents diverge from my desired outcomes:

- small adjustments: frequent
- medium adjustments: occasionally
- large adjustments: sometimes but rare

I also track progress in two add-on documents:

- implement/DETAILED.md: section 1) task checklist from tasks.md, and section 2) append-only timeline
- implement/EXECUTIVE.md: up-to-date summary, as well as issues, lessons learned, workarounds, mitigations, risks, and clarifications.

**The Problem:** Speckit is forward biased towards the /speckit.implement step. This is unlike human SDLC which has checkpoints for feedback loops.  I want to apply feedback based on progress in EXECUTIVE.md and DETAILED.md to earlier steps, and tell the LLM Agent to start at the next step. For example, I ask it to update constitution and spec, and restart at clarify. Important context is rarely preserved or applied correctly, so by the time the LLM Agent gets to implementation step 7 again, it quickly goes off in unwanted directions again.

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

Retrospectives capture that important feedback after implement increments (e.g. Agile/SCRUM sprint). Post Mortems capture additional important content at milestone increments (e.g. product release, SDLC program increment).

Both sources of feedback allow continuous improvement of earlier, higher-level steps in SDLC process.
Improvements and hard lessens learned are applied, with the goal to avoid or mitigate problems in subsequent SDLC increments.

## Speckit Methodology Challenges

**Questions:**

1. How do feedback loops work in Speckit methodology? Are there established patterns or best practices?
2. What should I do differently during or after step 7 (implement) to ensure content preservation?
3. How can I effectively integrate EXECUTIVE.md and DETAILED.md insights into earlier files like constitution.md and spec.md?
4. Provide examples of modifications to my process or file structures to enable successful restarts.

**Key Insight: Real Development is Iterative**
Speckit presents a sequential workflow, but real software development is inherently iterative. Constitution, spec, and plan are never truly "complete"—they evolve through feedback loops from implementation mistakes, lessons learned, and applied mitigations. This creates tension: Speckit assumes prerequisites are finalized, but iteration requires revisiting and improving them. The challenge is balancing Speckit's structure with practical iteration.

Please structure your response with: 1) Analysis of the issue, 2) Concrete recommendations, 3) Examples or templates, and 4) Potential pitfalls to avoid.
