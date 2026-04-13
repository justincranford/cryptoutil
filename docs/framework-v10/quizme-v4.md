# Quizme v4 — Framework v10 Follow-Up Questions

**Context**: These questions emerged from applying quizme-v3 answers and verifying plan/tasks
consistency. They address concrete implementation details that affect template file content.

---

## Question 1: Tini Availability in Dockerfile Runtime Image

**Question**: Decision 8 specifies shell-form command with tini ENTRYPOINT:
`ENTRYPOINT ["/sbin/tini", "--"]`. The Dockerfile template uses Alpine runtime.
Is `/sbin/tini` already installed in the Alpine runtime image, or does it need to
be added (e.g., `apk add --no-cache tini`)?

**A)** Tini is already installed — current Dockerfiles have `apk add tini` in the runtime stage
**B)** Tini is NOT installed — add `RUN apk add --no-cache tini` to the Dockerfile template
**C)** Use `docker-init` (built into Docker Engine) instead of tini — remove ENTRYPOINT entirely
**D)** Don't use tini at all — shell-form command handles PID 1 correctly with `exec`
**E)**

**Answer**:

**Rationale**: The shell-form command (`/bin/sh -c "exec ..."`) with `exec` replaces the shell
process, but without tini, zombie reaping doesn't work. The Dockerfile template content depends
on whether tini is explicitly installed.

---

## Question 2: Product-Level pki-init Override Mechanism

**Question**: Decision 4/Q6 says product compose "overrides PS-ID pki-init with deterministic
no-op." What exact Docker Compose mechanism should the template use to disable per-PS-ID
pki-init services when the product-level pki-init runs instead?

**A)** Override each PS-ID pki-init with `command: ["/bin/true"]` (exit 0 immediately)
**B)** Override each PS-ID pki-init with `entrypoint: ["/bin/true"]` + `command: []`
**C)** Use Docker Compose `profiles` — PS-ID pki-init in `ps-id` profile, product pki-init in `product` profile
**D)** Don't override — let both run; product pki-init generates product cert, PS-ID generates PS-ID cert (both needed)
**E)**

**Answer**:

**Rationale**: The override mechanism affects template content. Option D means both pki-init
services run (different tiers), which may actually be the correct behavior — product cert
is different from PS-ID cert.

---

## Question 3: Multiple --config= Flags in Shell-Form Command

**Question**: With Decision 16 (deployment config framework/domain split), each service needs
to load BOTH framework and domain config files. How should the shell-form command specify this?

**A)** Two `--config=` flags: `--config=/app/config/<ps-id>-app-framework-<variant>.yml --config=/app/config/<ps-id>-app-domain-<variant>.yml`
**B)** Single `--config=` with glob: `--config=/app/config/<ps-id>-app-*-<variant>.yml`
**C)** Single `--config=` directory: `--config=/app/config/` (app loads all YAML files)
**D)** Three `--config=` flags: framework deployment + domain deployment + standalone config
**E)**

**Answer**:

**Rationale**: The pflag `StringSlice` supports multiple `--config=` flags. The compose command
template needs the exact number and pattern. This also determines how `$SUITE_ARGS` interacts
(does suite add extra `--config=` flags?).

---

## Question 4: shared-postgres Follower Container Topology

**Question**: Decision 10 says "16 follower databases." How are these distributed across
PostgreSQL containers?

**A)** 1 suite-level follower container (all 10 DBs) + 5 product-level follower containers (subset of PS-ID DBs each) = 6 follower containers total
**B)** 1 single follower container with all 16 follower databases
**C)** 6 follower containers (1 suite + 5 product), but only the suite follower exists initially; product followers deferred
**D)** 3 follower containers (1 per tier: PS-ID follower, product follower, suite follower)
**E)**

**Answer**:

**Rationale**: The compose template needs to know exactly how many PostgreSQL container services
to define, their names, and which databases each follower replicates. This affects init scripts,
replication setup, and port assignments.

---

## Question 5: PostgreSQL Conf File Parameterization

**Question**: Do `postgresql-leader.conf` and `postgresql-follower.conf` need `__SUITE__`
or `__PS_ID__` substitution in their content, or are they static configuration files?

**A)** Static — same conf for all deployments (no parameterization needed), could even live outside template dir
**B)** Uses `__SUITE__` substitution for `cluster_name` or similar identification parameters
**C)** Leader is static, follower uses `__SUITE__` for `primary_conninfo` connection string
**D)** Both use `__SUITE__` for identification and `__PS_ID__` for per-database replication slots
**E)**

**Answer**:

**Rationale**: If conf files are fully static, they don't need `__KEY__` placeholders and the
linter can do simple byte comparison. If parameterized, the linter needs substitution logic.

---

## Question 6: validate_schema.go Update for static-files-path

**Question**: Quizme-v3 Q4 classified identity-spa `static-files-path` as a FRAMEWORK setting.
This means it needs to be added to the framework config schema in `validate_schema.go` and
registered as a pflag in the server builder. Is this in scope for framework-v10?

**A)** Yes — add `static-files-path` to framework schema + pflag registration in v10
**B)** No — document the gap, create a separate follow-up task outside v10 scope
**C)** Partial — add to framework schema only (viper key), defer pflag to later
**D)** Move it back to DOMAIN — identity-spa is the only PS-ID using it, not worth framework change
**E)**

**Answer**:

**Rationale**: Adding a new framework config key affects `validate_schema.go`, pflag
registration in the server builder, config validation, deployment template content, and
potentially all 10 PS-ID framework config files (as an optional key). Scope assessment needed.

---

## Question 7: `__PRODUCT_INCLUDE_LIST__` Exact YAML Format

**Question**: Product compose templates use `__PRODUCT_INCLUDE_LIST__` for the multi-line
include entries. What is the exact YAML format for the expanded value? (This determines
how `buildProductParams()` generates the substitution.)

**A)** Indented YAML list items (no `include:` key — key is in template, value is substituted):
```yaml
include:
  __PRODUCT_INCLUDE_LIST__
```
expands to:
```yaml
include:
  - path: ../sm-kms/compose.yml
  - path: ../sm-im/compose.yml
```

**B)** Full include block including key (entire `include:` section is the placeholder):
```yaml
__PRODUCT_INCLUDE_LIST__
```
expands to:
```yaml
include:
  - path: ../sm-kms/compose.yml
  - path: ../sm-im/compose.yml
```

**C)** Single-line comma-separated (linter expands to proper YAML):
`__PRODUCT_INCLUDE_LIST__` = `../sm-kms/compose.yml,../sm-im/compose.yml`

**D)** Read actual product compose files and match their exact existing format
**E)**

**Answer**:

**Rationale**: The substitution format determines both template content AND the
`buildProductParams()` implementation. Must match actual product compose file format
exactly for the compliance check to pass.

---
