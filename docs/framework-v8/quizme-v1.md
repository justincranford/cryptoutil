# Framework-v8 Quizme — v1

**Instructions**: For each question, write your answer (A, B, C, D, or E) next to `**Answer:**`.
Use E for a custom answer — write it after E). All three questions must be answered before
Phases 2–6 can begin. Return this file when complete; the agent will merge answers into
plan.md/tasks.md and delete this file.

---

## Question 1: How should each deployment tier expose PostgreSQL to the Docker host?

**Context**: The per-PS-ID PostgreSQL service currently uses `profiles: ["postgres"]` in the
PS-ID compose file, exposing a unique host port (e.g., sm-im at `127.0.0.1:54321:5432`).
The shared-postgres compose defines a leader/follower pair but exposes NO host ports.

You said: "PostgreSQL and Telemetry compose.yml files MUST be imported by all levels."

The design question is HOW each tier should expose PostgreSQL to developers on the Docker host:

**A)** Full `shared-postgres` include at all tiers. Per-PS-ID postgres DB service → `profiles: ["standalone"]`
(opt-in for exclusive per-PS-ID use by individual developers). Each PS-ID compose REDEFINES
`postgres-leader` with its specific host port (Approach C). PRODUCT and SUITE compose files
ALSO redefine `postgres-leader` on new PRODUCT/SUITE-level host ports (see Q2). Shared-postgres
is the single postgres for all-service runs at PRODUCT and SUITE.

**B)** `shared-postgres` included at PRODUCT and SUITE tiers only. SERVICE (PS-ID) tier keeps its
dedicated per-PS-ID postgres in `profiles: ["postgres"]` (current behavior, no include). When
the user said "all levels," they meant PRODUCT-level-and-above.

**C)** `shared-postgres` included at all tiers. Per-PS-ID postgres DB service REMOVED entirely
(one less service per PS-ID compose). Developers use `docker exec postgres-leader psql` for
direct database access. No host port exposure for postgres at any tier.

**D)** `shared-postgres` included at PRODUCT and SUITE only. PS-ID level runs a SINGLE postgres
(not leader/follower) using a new dedicated `single-node-postgres/compose.yml` file. SERVICE tier
exposes host ports 54320–54329 per PS-ID. PRODUCT/SUITE use shared-postgres.

**E)**

**Answer:**

**Why this matters**: This decision determines how many compose files are modified (10 PS-ID
files vs only 5 PRODUCT files), how postgres ports are structured at PRODUCT/SUITE, and whether
existing `profiles: ["postgres"]` workflows break.

---

## Question 2: What PostgreSQL host port scheme should PRODUCT and SUITE tiers use?

**Context**: Current SERVICE-tier postgres ports are 54320–54329 (one per PS-ID).

The standard offset scheme (+10000 for PRODUCT, +20000 for SUITE) **cannot be applied** to
postgres ports because 54320 + 20000 = 74,320, which exceeds the TCP maximum of 65,535.

This only applies if Answer Q1 = A or D (PRODUCT/SUITE expose postgres to host).
If Q1 = B or C, skip this question and write `N/A`.

**A)** **New compact range**: Define separate PRODUCT postgres port ranges under 60,000.
- SERVICE: 54320–54329 (10 ports, unchanged)
- PRODUCT: 54420–54424 (5 ports: sm=54420, jose=54421, pki=54422, identity=54423, skeleton=54424)
- SUITE: leader=54530, follower=54531

Document these in ENG-HANDBOOK.md Section 3.4.

**B)** **No host port exposure above SERVICE level.** PRODUCT and SUITE postgres services
run container-only (no `ports:` binding). Developers needing PRODUCT/SUITE postgres access use:
`docker compose -f deployments/sm/compose.yml exec postgres-leader psql -U ...`
Only the SERVICE tier exposes postgres on 54320–54329.

**C)** **Reuse 54320–54329 at all tiers.** Add a documented constraint to ENG-HANDBOOK.md:
"Only one deployment tier may run at a time on a developer machine."
This is operationally simple but disallows running SERVICE and PRODUCT simultaneously.

**D)** **Dynamic host ports at PRODUCT/SUITE.** Bind postgres as `"127.0.0.1::5432"` (no fixed
host port). Docker assigns an ephemeral port. Developers query the live port with:
`docker compose port postgres-leader 5432`

**E)**

**Answer:**

**Why this matters**: This determines whether ENG-HANDBOOK.md Section 3.4 needs a new postgres
port table, whether developers can run SERVICE and PRODUCT simultaneously (impacts CI/CD
isolation), and whether compose files at PRODUCT/SUITE have postgres `ports:` entries.

---

## Question 3: What should happen with Product-level Dockerfiles (Carryover Item 2)?

**Context**: Framework-v7 carryover Item 2 (priority: HIGH) called for creating Dockerfiles
at the PRODUCT level (`deployments/sm/Dockerfile`, `deployments/jose/Dockerfile`, etc.).

The recursive-include design in framework-v8 RENDERS THIS UNNECESSARY: PRODUCT compose files
include PS-ID compose files, which include PS-ID Dockerfiles. Each PS-ID image is built by its
own `builder-{PS-ID}` service. No separate PRODUCT-level build is needed.

However, `validate_structure.go` currently REQUIRES a Dockerfile at every PRODUCT deployment
directory—and will error without one.

**A)** **Defer / superseded**: Update `validate_structure.go` to remove the Dockerfile requirement
from `DeploymentTypeProduct`. Note in carryover.md that the requirement is superseded by the
recursive-include architecture. No product Dockerfiles created.

**B)** **Create lightweight product Dockerfiles**: Each PRODUCT gets a thin `FROM cryptoutil:dev`
wrapper with product-specific OCI labels (name, vendor). Validator keeps its Dockerfile requirement.
These Dockerfiles exist for CI/CD identification purposes only; they do not produce new binaries.

**C)** **Create product binary Dockerfiles**: Each product Dockerfile builds a single multi-service
binary (`sm-server`, `jose-server`, etc.) that bundles the PS-ID binaries as subcommands. This
requires new `cmd/sm/main.go`, `cmd/jose/main.go`, etc. implementations — significant scope expansion.

**D)** **Mark as permanently infeasible**: Update carryover.md to mark Item 2 as CANCELLED with
rationale. Update ENG-HANDBOOK.md to document that product deployments use PS-ID Dockerfiles
transitively and do not require product-level Dockerfiles. Remove the Dockerfile requirement from
`validate_structure.go`.

**E)**

**Answer:**

**Why this matters**: This controls whether `validate_structure.go` is loosened (A/D) or
product Dockerfiles are created (B/C). Option C significantly expands scope beyond framework-v8.
Option B creates files that may cause confusion (do they actually build anything?).
