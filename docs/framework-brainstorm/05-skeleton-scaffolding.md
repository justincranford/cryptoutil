# Skeleton Stereotype and Scaffolding

## The Problem

The skeleton-template today is 21 files, mostly stubs. A developer looking at it cannot answer:

- Where do I put my database models?
- How do I add a new API endpoint from spec to handler?
- How do I write integration tests for my repository?
- How do I add a new migration?
- How do I write a client that calls this service?

Every new service migration requires reading 50-60 files in sm-im/jose-ja to extract conventions.
That knowledge must be re-derived every time, costing days per service.

## Spring Boot Initializr Model (Reference)

Spring Initializr generates a runnable project with:
- A real Entity + Repository + Service + Controller wired together
- A working test suite (unit + integration)
- A complete build file with all declared dependencies
The generated project compiles, tests pass, and it answers WHERE to put things.

## What skeleton-template Should Become

### Level 1: Reference Implementation (immediate goal)

A skeleton-template that compiles, tests pass, and contains a real Widget CRUD domain:

```ninternal/apps/skeleton/template/
  domain/
    widget.go              # Domain model (no GORM tags)
    widget_service.go      # Business logic interface + implementation
  repository/
    widget_repository.go   # GORM model + interface + implementation
    widget_repository_test.go  # Integration test with SQLite in-memory
    migrations/
      2001_init.up.sql
      2001_init.down.sql
  server/
    server.go              # NewServer() wiring builder
    apis/
      widgets.go           # Strict server implementation
      widgets_test.go      # Handler test with app.Test()
  client/
    client.go              # Generated client + wrapper
    client_test.go         # Integration test calling real server
  e2e/
    e2e_test.go            # Full stack test with Docker Compose
  testing/
    helpers.go             # Shared test utilities
```n
The widget domain is intentionally trivial: ID, name, description, created_at.
The structure, not the domain, is what skeleton teaches.


### Level 2: Code Generation (next goal)

A CLI command that reads skeleton-template and generates a new service:

```bash
go run ./cmd/cicd new-service \
  --product pki \
  --service ca \
  --domain Certificate \
  --port-public 8100 \
  --port-admin 9100
```n
Output: internal/apps/pki/ca/ with all skeleton files renamed and Widget -> Certificate.

### Level 3: Conformance Checking (ongoing)

```bash
go run ./cmd/cicd diff-skeleton --service pki-ca
```n
Diffs a real service against skeleton structure. Reports:
- Missing files
- Structural divergence
- Outdated patterns (skeleton updated, service not yet synced)

## Skeleton File Annotations

Each skeleton file gets a header comment indicating what to change vs. keep:

```go
// SCAFFOLD: Rename Widget -> Certificate (your domain model).
// SCAFFOLD: Update table name in migration 2001_init.up.sql.
// KEEP: All gorm:\"type:text\" annotations (cross-DB compat).
// KEEP: All t.Parallel() calls (required by testing standards).
```n
In generated files, these are stripped. In the skeleton source, they document
the intent and make code review of generated services faster.


## cicd new-service Design

The new-service scaffolding tool would:

1. Read skeleton-template as a Go embed.FS
2. Walk all files in the directory tree
3. Apply substitutions: skeleton->product, template->service, Widget->DomainEntity
4. Remove SCAFFOLD: lines from file headers
5. Write output to internal/apps/{product}/{service}/
6. Register cmd entry point, update compose files, create config dirs
7. Run go build to verify clean output

This is the same approach used by:
- Spring Initializr (HTTP API, generates zip of template files)
- Buffalo (buffalo new) -- Go framework generator
- NestJS CLI (nest generate module/service/controller)
- Rails generators (rails generate scaffold Name field:type)

## Prototype Implementation Path

### Step 1: Promote skeleton-template to full CRUD reference (2 weeks)
- Add widget domain model, service, repository, migrations
- Add widget server handlers implementing the OpenAPI spec
- Add widget client with integration test
- Add e2e test using Docker Compose subset
- Confirm: go test ./internal/apps/skeleton/... -v

### Step 2: Add SCAFFOLD annotations (2 days)
- Headers on files explaining what to customize vs keep
- Decision table in README.md with common customizations

### Step 3: new-service scaffolding tool (3 weeks)
- Parse skeleton file tree
- Text substitution engine (simple string replace)
- Output validation (go build of generated code)
- Integration into cicd command suite
- Test: generate pki-ca service, verify it builds

### Step 4: diff-skeleton conformance checker (1 week)
- Walk skeleton + target service trees in parallel
- Generate structural diff report
- Add to CI/CD as non-blocking warning
- Elevate to error after adoption period

## Open Questions

1. Should skeleton generate a single service or a product (multiple services)?
   Lean: single service per invocation; compose product from services.

2. How to handle identity services which share common auth patterns?
   Generate each separately; rely on shared internal packages for common code.

3. Should skeleton also encompass OpenAPI spec generation?
   Aspirational: generate spec first (intent-first), generate code from spec.
   Practical: generate spec scaffold with example Widget CRUD endpoints.

4. How to keep generated code in sync as skeleton evolves?
   diff-skeleton tool identifies drift; developer migrates with guidance.
   Long-term: regenerate boilerplate files, keep domain files untouched.
