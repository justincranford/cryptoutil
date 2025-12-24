# SpecKit Methodology - Complete Specification

**Version**: 1.0.0
**Last Updated**: 2025-12-24
**Referenced By**: `.github/instructions/01-03.speckit.instructions.md`

## Overview

SpecKit (Specification Kit) is a spec-driven development (SDD) design and validation toolset that allows LLM Agents to implement designs autonomously.

**Seven Steps**: 1. Constitution, 2. Clarify (optional), 3. Spec, 4. Plan, 5. Tasks, 6. Analyze (optional), 7. Implement

**When to Use**:
- ✅ Design before implementing code
- ✅ Catch design issues early
- ✅ Direct LLM Agents to converge on working implementations
- ✅ Straight-line or spiral workflows

**When NOT to Use**:
- ❌ Human SDLC (lacks iterative feedback loops)
- ❌ Validating implementation, business domain rules, complex state transitions
- ❌ Spiking or quick iteration (feels heavy, slow, overly restrictive)

## Customizations

### Living Documents Pattern

**Treat constitution, spec, and plan as evolving through implementation feedback** - NOT static prerequisites.

**MANDATORY**: Update immediately when discovering:
- Constraints
- Contradictions
- Insights

### Clarify and Analyze as MANDATORY

**Original**: Clarify and Analyze steps are optional
**Customization**: Treat both as MANDATORY

**Rationale**: Clarify helps identify and fix issues in constitution and spec before proceeding to plan.

### CLARIFY-QUIZME-##.md Format

**MANDATORY**: CLARIFY-QUIZME-##.md MUST only contain UNKNOWN answers requiring user input.

**Format**: Multiple choice questions with insightful A-D answers and blank E write-in answer.

**NEVER include**:
- ❌ Questions with answers from codebase/constitution/spec/copilot instructions
- ❌ Pre-filled answers or example answers in Write-in questions

**ALWAYS**:
- ✅ Search codebase/docs before adding questions to CLARIFY-QUIZME-##.md
- ✅ Merge CLARIFY-QUIZME-##.md answers into clarify.md, refactor structure/content to accommodate
- ✅ Use updated clarify.md to backport clarifications into constitution.md and spec.md
- ✅ Re-analyze constitution.md, spec.md, clarify.md, generate another CLARIFY-QUIZME-##.md automatically
- ✅ Prompt user to review CLARIFY-QUIZME-##.md if new questions, or automatically run /specify.plan and notify user

## Workflow Gates - MANDATORY

| Gate | Evidence Required | Validation |
|------|-------------------|------------|
| **Constitution** | No TBD placeholders, terminology defined, quality gates, alignment with copilot instructions without conflicts | `grep -i "TBD\|TODO\|FIXME" constitution.md` = 0 |
| **Specification** | Functional/non-functional requirements, architecture, API contracts, services overview, databases overview, security requirements | Verify all constitution requirements mapped |
| **Clarification** | Topical organization, clarify.md for knowns, CLARIFY-QUIZME-##.md for unknowns | Verify no pending questions |
| **Plan** | Phases with task breakdowns, task dependencies, completion criteria per phase, risk assessment, aligns with constitution/spec/copilot | Compare plan vs constitution/spec/clarify |
| **Tasks** | Task checklist, completion criteria per task, phase assignments, dependencies, effort estimates (S/M/L) | Verify tasks.md matches plan.md phases |
| **Analysis** | Complexity analysis, risk assessment, dependency graph, resource requirements (optional) | Verify analyze.md exists if needed |
| **Implementation** | Track in DETAILED.md Sections 1+2, EXECUTIVE.md, issues found, risks discovered, lessons learned, progress evidence (tests pass, coverage ≥95%/98%, mutation ≥85%/98%, timing <15s/<120s unit, <45s/<240s E2E) | `go test ./... -coverprofile`, `gremlins unleash` |

## Evidence-Based Completion - MANDATORY

**NEVER mark tasks complete during Implementation without objective evidence**:

### Code Evidence
- `go build ./...` clean
- `golangci-lint run` clean
- No new TODOs
- Coverage ≥95% (production code) or ≥98% (infrastructure/utility code)

### Test Evidence
- `go test ./...` passes
- No skips without tracking
- Coverage reports generated

### Mutation Evidence
- `gremlins unleash` passes
- Quality analysis shows ≥85% during early phases
- Quality gate jumps to ≥98% during later phases

### Git Evidence
- Conventional commit format
- Clean working tree
- Changes align with task

### Git Hooks Evidence
- Pre-commit checks pass
- Pre-push checks pass

## Feedback Loop Patterns - MANDATORY

**When discovering constraints/contradictions/lessons during Implementation**:

1. Document in DETAILED.md Section 2 timeline, mark as needing user review for copilot anti-patterns instructions
2. Document in EXECUTIVE.md, mark as needing user review for copilot anti-patterns instructions
3. Update constitution/spec/clarify immediately
4. Commit with traceable reference to source

**CRITICAL**: DON'T PROMPT THE USER TO REVIEW DETAILED.md and EXECUTIVE.md. The user already implicitly knows to asynchronously review DETAILED.md and EXECUTIVE.md.

## DETAILED.md Structure

### Section 1: Task Checklist

Maintain from tasks.md with:
- Status: ❌ (not started), ⚠️ (in progress/blocked), ✅ (complete)
- Blockers
- Notes
- Coverage metrics
- Commit references

### Section 2: Append-Only Timeline

Chronological entries with:
- **Date**: Task description
- **Work completed**: Summary of implementation
- **Coverage/quality metrics**: Before/after numbers
- **Lessons learned**: Insights discovered
- **Constraints discovered**: Add to constitution.md
- **Requirements discovered**: Add to spec.md
- **Related commits**: Git commit hashes

## EXECUTIVE.md Structure

### Stakeholder Overview
- Current phase
- Progress percentage
- Coverage metrics
- Mutation score
- Blockers

### Customer Demonstrability
- Docker Compose commands
- E2E demo scenarios
- Video demonstrations (optional)

### Risk Tracking
- Known issues with severity
- Impact assessment
- Workarounds
- Root cause analysis
- Resolution plan
- Status

### Post-Mortem Lessons
- Lesson learned
- Prevention pattern
- Where applied
- Reference documentation

## Key Takeaways

1. **Living Documents**: Constitution, spec, plan evolve through implementation feedback
2. **Clarify/Analyze Mandatory**: Treat optional steps as mandatory for quality
3. **QUIZME Format**: Multiple choice for efficiency, only UNKNOWN answers
4. **Evidence-Based**: Never mark complete without objective evidence (tests, coverage, mutations, git)
5. **Feedback Loops**: Update constitution/spec/clarify immediately when discovering new information
6. **Async Review**: User reviews DETAILED.md/EXECUTIVE.md asynchronously, don't prompt
