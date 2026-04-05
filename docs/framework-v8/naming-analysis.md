# ARCHITECTURE.md — Naming Analysis

## What the Document Actually Covers

The current document title is **"cryptoutil Architecture — Single Source of Truth"**. At 4,179 lines across 15 sections + 3 appendices, its scope far exceeds what "Architecture" implies.

### Full Topic Inventory (15 domains)

| Domain | Key Terms | Primary Section |
|--------|-----------|----------------|
| Product suite | PKI, JOSE, Secrets Manager, Identity, Skeleton; 1 Suite → 5 Products → 10 Services | §3 |
| Service architecture | Dual HTTPS, health checks, graceful shutdown, service framework, builder pattern | §5 |
| Security / cryptography | FIPS 140-3, barrier/encryption-at-rest, elastic key ring, JOSE/JWK/JWS/JWE, PKI/X.509, CA/Browser Forum, mTLS, secrets detection | §6 |
| Data architecture | GORM, SQLite/PostgreSQL dual-DB strategy, multi-tenancy, schema isolation, migrations | §7 |
| API design | OpenAPI 3.0.3, oapi-codegen, REST conventions, rate limiting, dual path pattern (`/service/**` + `/browser/**`) | §8 |
| Infrastructure / DevOps | CLI patterns, config management, OTLP observability, Docker / Kubernetes, CI/CD workflows, pre-commit hooks | §9 |
| Testing | Unit, integration, E2E, mutation (gremlins), load (Gatling), fuzz, benchmark, race, SAST, DAST, contract, seam injection | §10 |
| Quality gates | Coverage ≥95%/98%, golangci-lint v2, linter catalog, file size limits, CGO ban | §11 |
| Deployment | Multi-stage Docker, Docker secrets, multi-level hierarchy (Suite/Product/Service), Kubernetes, graceful shutdown | §12 |
| Deployment tooling | 8 deployment validators, config schema, secrets enforcement, documentation propagation (`@propagate`) | §13 |
| Development practices | Coding standards, conventional commits, branching, opportunistic quality, scripting policy, archive policy | §14 |
| AI agent orchestration | Copilot agents/skills/instructions, Claude Code agents/skills, dual canonical format, handoff flows, lint-agent-drift, agentskills.io | §2.1 |
| Architecture fitness functions | 18+ fitness sub-linters (parallel-tests, file-size, test-patterns, entity-registry-completeness, banned-product-names, …) | §9.11 |
| Developer inner-loop tooling | `cicd-lint` (13 linters, 2 formatters, 1 script), `format-go`, `format-go-test`, fitness runner | §9.10 |
| Autonomous execution protocol | Beast-mode, end-of-turn commit gate, evidence-based completion, mandatory review passes | §14.11 |

### Terms Covered But Previously Absent from Executive Summary / Title

The following terms are extensively documented inside the body but were not mentioned in §1.2 Key Architectural Characteristics or §1.3 Core Principles (gaps corrected in this framework-v8 session):

1. **AI Agent Orchestration** — §2.1 comprises ~600 lines on Copilot/Claude agent types, dual canonical strategy, lint-agent-drift, agentskills.io open standard
2. **Service Framework / Builder Pattern** — §5.1–5.2 describes a shared framework eliminating 48,000+ lines of boilerplate per service
3. **Architecture Fitness Functions** — §9.11 defines 18+ linters that programmatically enforce invariants on every commit
4. **Documentation Propagation System** — §13.4 @source/@propagate marker system keeping all artifacts byte-for-byte in sync
5. **Developer Inner-Loop Tooling** — §9.10 `cicd-lint` local enforcement with 13 linters, 2 formatters
6. **Autonomous Execution Protocol** — §14.11 beast-mode, pre-commit quality gates, end-of-turn commit discipline
7. **CGO-Free Compilation** — §11.1.2 CGO_ENABLED=0 mandate for fully static, cross-compilable binaries
8. **Spec-Driven Development (SDD)** — mentioned in §1.1 vision but not surfaced in characteristics/principles
9. **Elastic Key Ring** — §6.6/6.4.5 per-operation key rotation strategy unique to this platform
10. **Barrier / Encryption-at-Rest Layer** — §6.4.2 multi-layer key hierarchy distinct from TLS
11. **CA/Browser Forum Compliance** — §6.5 PKI standards beyond FIPS 140-3
12. **Dual Database Strategy** — §7.3 SQLite (dev/test) + PostgreSQL (prod/E2E) with unit tests NEVER using PostgreSQL

---

## Top 20 Better Names

The current name misleads on scope. "Architecture" typically implies component diagrams and service topology, not testing protocols, AI agent orchestration, pre-commit hooks, or documentation propagation. The document is the engineering equivalent of a constitution + handbook + runbook combined.

Evaluated against three criteria:
- **Scope signaling** — does the name hint at the breadth beyond mere architecture?
- **Authority signaling** — does it convey this is the normative, binding reference?
- **Discoverability** — would a new team member look for this document under this name?

| Rank | Proposed Name | Scope | Authority | Discoverability | Notes |
|------|--------------|-------|-----------|----------------|-------|
| 1 | **cryptoutil Engineering Handbook** | ★★★★★ | ★★★★☆ | ★★★★★ | "Handbook" universally implies comprehensive, practical, binding reference |
| 2 | **cryptoutil Platform Engineering Reference** | ★★★★★ | ★★★★★ | ★★★★☆ | "Platform Engineering" names the discipline explicitly; "Reference" signals canonical |
| 3 | **cryptoutil Engineering Codex** | ★★★★★ | ★★★★★ | ★★★☆☆ | "Codex" = authoritative binding document; less familiar outside large orgs |
| 4 | **cryptoutil Engineering Constitution** | ★★★★★ | ★★★★★ | ★★★★☆ | Signals the document governs all decisions; may feel over-formal for a solo/small project |
| 5 | **cryptoutil Engineering Playbook** | ★★★★★ | ★★★★☆ | ★★★★★ | "Playbook" implies actionable patterns; widely used in engineering culture |
| 6 | **cryptoutil Technical Standards & Practices** | ★★★★★ | ★★★★★ | ★★★★☆ | Precise and formal; mirrors ISO/IEC document naming |
| 7 | **cryptoutil Developer Codex** | ★★★★☆ | ★★★★★ | ★★★★☆ | "Developer" scopes to implementers; "Codex" = binding rulebook |
| 8 | **cryptoutil Engineering Doctrine** | ★★★★★ | ★★★★★ | ★★★☆☆ | Strong authority signal; may sound militaristic |
| 9 | **cryptoutil Platform Architecture Manual** | ★★★★☆ | ★★★★★ | ★★★★☆ | "Manual" implies procedural depth; "Platform" widens scope |
| 10 | **cryptoutil Systems Engineering Reference** | ★★★★★ | ★★★★★ | ★★★★☆ | "Systems Engineering" encompasses architecture + testing + tooling + deployment |
| 11 | **cryptoutil Engineering Excellence Guide** | ★★★★☆ | ★★★★☆ | ★★★★☆ | Google-SRE style naming; pairs well with quality-first theme |
| 12 | **cryptoutil Technical Compendium** | ★★★★★ | ★★★★☆ | ★★★☆☆ | "Compendium" signals exhaustive coverage; less common in engineering |
| 13 | **cryptoutil Engineering Design System** | ★★★★☆ | ★★★★☆ | ★★★☆☆ | Adapts "design system" concept from UI to platform engineering |
| 14 | **cryptoutil Developer Intelligence** | ★★★★☆ | ★★★☆☆ | ★★★☆☆ | Hints at AI-augmented DX; unique but unfamiliar |
| 15 | **cryptoutil Platform Governance Guide** | ★★★★☆ | ★★★★★ | ★★★★☆ | "Governance" signals binding rules; fits the compliance + fitness-function angle |
| 16 | **cryptoutil AI-Augmented Engineering Reference** | ★★★★★ | ★★★★☆ | ★★★★☆ | Explicitly surfaces AI agent orchestration as a first-class topic |
| 17 | **cryptoutil DevSecOps Handbook** | ★★★★☆ | ★★★★☆ | ★★★★★ | Industry-recognized term (Dev + Sec + Ops); widely searchable |
| 18 | **cryptoutil Software Engineering Bible** | ★★★★★ | ★★★★★ | ★★★★★ | Maximally authoritative; informal superlative (common in team culture) |
| 19 | **cryptoutil Engineering Manifesto** | ★★★★☆ | ★★★★★ | ★★★☆☆ | Signals strong opinionated principles; implies evolution path |
| 20 | **cryptoutil Platform Rulebook** | ★★★★★ | ★★★★★ | ★★★★★ | Simplest, clearest signal of binding authority + breadth |

### Recommendation

**Top 3 finalists:**

1. **cryptoutil Engineering Handbook** — most universally understood, signals practical depth, zero ambiguity. Best for new contributors. Recommended if the goal is clarity.

2. **cryptoutil Platform Engineering Reference** — explicitly names the discipline (Platform Engineering) and signals a canonical normative document. Best if the project positions itself as platform engineering.

3. **cryptoutil Engineering Codex** — signals "the binding rulebook". Best for emphasizing that this document governs ALL decisions with no exceptions. Recommended if the goal is authority.

### Why Not Keep "Architecture - Single Source of Truth"?

- **"Architecture"** undersells scope: the document covers testing, AI agents, pre-commit hooks, documentation propagation, and developer workflow — none of which are "architecture" in the classic sense.
- **"Single Source of Truth"** is a meta-claim every internal document makes; it adds no searchability or scope signal.
- The subtitle invites readers to skip sections ("I just need the architecture, not the testing rules") when in fact ALL sections are binding.

---

## Proposed Title Change (Optional)

If renaming is desired, recommended replacement:

```
# cryptoutil Engineering Handbook — Normative Reference
```

Or keeping the platform engineering framing:

```
# cryptoutil Platform Engineering Reference
```

The subtitle "Normative Reference" replaces "Single Source of Truth" with a formal term that signals binding authority without redundancy.
