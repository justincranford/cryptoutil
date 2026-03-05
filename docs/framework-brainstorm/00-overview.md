# Framework Brainstorm — Overview and Navigation

**Date**: March 2026
**Status**: Brainstorm / Ideas (not prescriptive decisions)

---

## Purpose

This brainstorm examines how to evolve cryptoutil from its current state — a
monorepo of 10 services with a shared template — toward a scalable, enforceable
framework that a single developer can use to build and maintain all 10 services
efficiently.

Two distinct roles are articulated:

| Role | Current Name | Desired Behavior |
|------|--------------|-----------------|
| Application Framework | internal/apps/template/service/ | Equivalent to Spring Boot @SpringBootApplication - provides infrastructure, auto-configuration, extension points |
| Service Stereotype | skeleton-template | Equivalent to a Spring Boot Initializr project - copy-paste starting point that enforces conventions, domain business logic goes here |

---

## Files in This Brainstorm

- 01-current-state-analysis.md   What works, what hurts, root cause analysis
- 02-cross-language-frameworks.md  Java/Python/JS/Rust/Go framework lessons
- 03-go-framework-patterns.md   Go-specific DI, plugins, code generation
- 04-framework-design.md    Concrete redesign proposals
- 05-skeleton-scaffolding.md   Skeleton + scaffolding tool design
- 06-solo-developer-scaling.md  10x and 100x ideas for a solo developer
- 07-fitness-functions.md   Automated architectural enforcement
- 08-recommendations.md    Prioritized action list

---

## Current State Summary

cryptoutil has 1795 Go files total:
- template:  316 files (framework)
- identity:  550 files (5 services, partially migrated)
- sm:        181 files (2 services, done)
- jose:       78 files (1 service, done)
- pki:       130 files (1 service, TODO)
- skeleton:   21 files (stub only)
- cicd:      188 files (tooling)

3 of 10 services migrated. Each migration harder than the last.

---

## TL;DR Recommendations

Priority 0 (do first):
- ServiceContract interface enforced at compile time
- cicd new-service scaffolding tool
- Promote skeleton to full reference implementation

Priority 1:
- Module/plugin system to decompose ServerBuilder
- Cross-service contract test suite
- cicd diff-skeleton conformance tool

Priority 2:
- Fitness functions in CI
- OpenAPI to full CRUD generation pipeline

Priority 3 (big investment):
- Extract framework to separate Go module
