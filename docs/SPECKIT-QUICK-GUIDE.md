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

## Speckit Implement Tracking

When making frequent adjustments during /speckit.implement, I ask LLM Agents to track 3 things:

- DETAILED.md:
  - Section 1: Copy of tasks.md content, as a checklist, to track out-of-order task completion
  - Section 2: Append-only timeline of in-order task completions, useful as an audit trail of decisions
- EXECUTIVE.md
  - Summary of latest completed progress, issues, lessons learned, workarounds, mitigations, risks, clarifications, etc

## Speckit Methodology Challenges

I occasionally ask LLM Agents to bubble content from EXECUTIVE.md and DETAILED.md to earlier speckit steps. For example,
apply updates to:

1. /speckit.constitution: .specify\memory\constitution.md
2. /speckit.specify: specs\002-cryptoutil\spec.md
3. /speckit.clarify: specs\002-cryptoutil\clarify.md and specs\002-cryptoutil\CLARIFY-QUIZME.md
4. /speckit.plan: specs\002-cryptoutil\plan.md

My goal is to be able to copy the current speckit directory, truncate it after step 2 or 4, and start immediately after the truncation point.

Everytime I have tried that, critically important content and context is lost from EXECUTIVE.md, DETAILED.md, and tasks are lost. The issues, lessons learned, workarounds, mitigations, risks, clarifications, etc are not preserved. I would
have expected all of that content and context to be preserved in the reused constitution.md and spec.md, and optionally the reused clarify.md and plan.md if I keep them. However, it needs seems to work.

I don't understand if/how feedback loops work in speckit methodology.

I also don't understand if/what I need to do differently during or after /speckit.implement step, to preserve
content and context in constitution.md and spec.md, and optionally the reused clarify.md and plan.md.
