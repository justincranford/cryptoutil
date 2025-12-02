# Workspace and Tooling Alignment

## Overview

This document defines VS Code configuration updates required for the service group refactoring, including settings, launch configurations, tasks, and terminal auto-approval patterns.

**Cross-references:**

- [Group Directory Blueprint](./blueprint.md) - Defines target package locations
- [Import Alias Policy](./import-aliases.md) - Import alias migration strategy
- [CLI Strategy Framework](./cli-strategy.md) - New CLI command structure

---

## .vscode/settings.json Updates

### gopls Configuration

**Current Configuration:**

```jsonc
"gopls": {
  "formatting.local": "cryptoutil",
  "buildFlags": ["-tags=e2e"]
}
```

**Refactor Impact:**

- **No changes needed** - `cryptoutil` prefix covers all refactored packages
- `buildFlags` includes E2E tests regardless of package location

**Recommendation:**

- Keep existing gopls settings unchanged

---

### Terminal Auto-Approval Patterns

**Current Pattern Count:** ~50 patterns (git, docker, go, python, java, linux, windows, powershell commands)

**Refactor Impact:**

- **Medium** - New CLI commands need auto-approval patterns
- Add patterns for `cryptoutil <service-group> <subcommand>`

**Proposed Additions:**

```jsonc
"chat.tools.terminal.autoApprove": {
  // Existing patterns preserved...

  // NEW: Service Group CLI Commands
  "/^cryptoutil (kms|identity|ca) (server|client|help|version).*/": true,
  "/^cryptoutil kms (keygen|barrier|unseal|seal|rekey).*/": true,
  "/^cryptoutil identity (authz|idp|rs|spa-rp) (start|stop|status|help).*/": true,
  "/^cryptoutil ca (issue|revoke|crl|ocsp|renew|export|import).*/": true,

  // Backward compatibility (deprecated commands)
  "/^cryptoutil (server|help|version).*/": true,  // Legacy KMS commands
}
```

**Rationale:**

- Auto-approve all service group CLI commands (safe, informational, or controlled by flags)
- Includes subcommands: server start/stop, client operations, keygen, barrier management
- Backward compatibility for deprecated `kms cryptoutil server` commands

**Total Auto-Approval Patterns:** ~50 existing + 5 new = 55 patterns

---

### File/Search Exclusions

**Current Exclusions:**

**`files.exclude`:**

- Build artifacts (bin/, *.exe,*.dll, *.so,*.dylib)
- Test artifacts (coverage*, test-results/)
- Cache files (.cspellcache, .cicd/)
- Dependencies (vendor/, node_modules/)
- Python artifacts (**pycache**/, *.pyc, venv/, .pytest_cache/)
- Java artifacts (out/, .mvn/, .gradle/)

**Refactor Impact:**

- **None** - Exclusions are artifact-based, not package-based
- Post-refactor structure generates same artifacts

**Recommendation:**

- No changes needed to `files.exclude`, `search.exclude`, `files.watcherExclude`

---

### YAML Schema Validation

**Current Schemas:**

- Docker Compose: `**/*compose*.{yml,yaml}`
- GitHub Workflows: `.github/workflows/*.{yml,yaml}`
- GitHub Actions: `.github/actions/**/*.{yml,yaml}`
- OpenAPI: `api/openapi_spec*.{yaml,yml}`
- Pre-commit: `.pre-commit-config.yml`
- Prometheus, Grafana, Kustomization, Helmfile schemas

**Refactor Impact:**

- **None** - Schemas are file pattern-based
- Docker Compose and workflow files remain in same locations

**Recommendation:**

- No changes needed to `yaml.schemas`

---

## .vscode/launch.json Updates

### Current Launch Configurations (5 Total)

| Name | Type | Program | Args | Purpose |
|------|------|---------|------|---------|
| `cryptoutil sqlite` | go | `${workspaceFolder}/main.go` | `--dev=true, --log-level=INFO` | Run KMS with SQLite |
| `cryptoutil postgres` | go | `${workspaceFolder}/main.go` | `--database-url=postgres://..., --log-level=ALL` | Run KMS with PostgreSQL |
| `cryptoutil model` | go | `oapi-codegen` | OpenAPI model generation | Generate OpenAPI models |
| `cryptoutil client` | go | `oapi-codegen` | OpenAPI client generation | Generate OpenAPI client |
| `kms cryptoutil server` | go | `oapi-codegen` | OpenAPI server generation | Generate OpenAPI server |

### Refactor Impact

**Phase 1 (Identity Extraction):**

- **Add:** Identity service launch configurations

**Phase 2 (KMS Extraction):**

- **Update:** KMS launch configurations to reference new paths
- **Rename:** `cryptoutil sqlite` → `KMS (SQLite)`, `cryptoutil postgres` → `KMS (PostgreSQL)`

**Phase 3 (CA Preparation):**

- **Add:** CA service launch configurations

---

### Proposed Launch Configurations Post-Refactor

#### KMS Service Group

```jsonc
{
  "name": "KMS Server (SQLite)",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "kms", "server", "start",
    "--dev=true",
    "--log-level=INFO"
  ],
  "env": {
    "CGO_ENABLED": "0"
  }
},
{
  "name": "KMS Server (PostgreSQL)",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "kms", "server", "start",
    "--config=configs/kms/production.yml",
    "--log-level=DEBUG"
  ],
  "env": {
    "CGO_ENABLED": "0"
  }
},
{
  "name": "KMS Client (Test Connection)",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "kms", "client", "status",
    "--url=https://localhost:8080"
  ]
}
```

#### Identity Service Group

```jsonc
{
  "name": "Identity AuthZ Server",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "identity", "authz", "start",
    "--config=configs/identity/development.yml",
    "--log-level=DEBUG"
  ]
},
{
  "name": "Identity IdP Server",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "identity", "idp", "start",
    "--config=configs/identity/development.yml",
    "--log-level=DEBUG"
  ]
},
{
  "name": "Identity RS Server",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "identity", "rs", "start",
    "--config=configs/identity/development.yml",
    "--log-level=DEBUG"
  ]
},
{
  "name": "Identity SPA RP Server",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "identity", "spa-rp", "start",
    "--config=configs/identity/development.yml",
    "--log-level=DEBUG"
  ]
}
```

#### CA Service Group (Placeholder)

```jsonc
{
  "name": "CA Server (Development)",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "ca", "server", "start",
    "--config=configs/ca/development.yml",
    "--log-level=DEBUG"
  ]
},
{
  "name": "CA Client (Issue Certificate)",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/cryptoutil/main.go",
  "args": [
    "ca", "client", "issue",
    "--cn=example.com",
    "--url=https://localhost:9000"
  ]
}
```

#### OpenAPI Code Generation (Unchanged)

```jsonc
{
  "name": "OpenAPI Generate Model",
  "type": "go",
  "request": "launch",
  "program": "${workspaceFolder}/../oapi-codegen/cmd/oapi-codegen/oapi-codegen.go",
  "cwd": "${workspaceFolder}/api",
  "args": [
    "--config=./openapi-gen_config_model.yaml",
    "./openapi_spec_components.yaml"
  ]
},
{
  "name": "OpenAPI Generate Client",
  "type": "go",
  "request": "launch",
  "program": "${workspaceFolder}/../oapi-codegen/cmd/oapi-codegen/oapi-codegen.go",
  "cwd": "${workspaceFolder}/api",
  "args": [
    "--config=./openapi-gen_config_client.yaml",
    "./openapi_spec_paths.yaml"
  ]
},
{
  "name": "OpenAPI Generate Server",
  "type": "go",
  "request": "launch",
  "program": "${workspaceFolder}/../oapi-codegen/cmd/oapi-codegen/oapi-codegen.go",
  "cwd": "${workspaceFolder}/api",
  "args": [
    "--config=./openapi-gen_config_server.yaml",
    "./openapi_spec_paths.yaml"
  ]
}
```

**Total Launch Configurations:** 12 (3 KMS + 4 Identity + 2 CA + 3 OpenAPI)

---

## .vscode/tasks.json Creation

**Current State:** File does not exist

**Proposed Tasks:**

### Service Group Tasks

```jsonc
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build All",
      "type": "shell",
      "command": "go build -v -ldflags=\"-s -extldflags '-static'\" -o bin/cryptoutil ./cmd/cryptoutil",
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "problemMatcher": ["$go"]
    },
    {
      "label": "Test All",
      "type": "shell",
      "command": "go test ./... -count=1 -timeout=10m",
      "group": {
        "kind": "test",
        "isDefault": true
      },
      "problemMatcher": ["$go"]
    },
    {
      "label": "Test KMS",
      "type": "shell",
      "command": "go test ./internal/kms/... ./internal/server/... -count=1 -timeout=10m",
      "group": "test",
      "problemMatcher": ["$go"]
    },
    {
      "label": "Test Identity",
      "type": "shell",
      "command": "go test ./internal/identity/... -count=1 -timeout=10m",
      "group": "test",
      "problemMatcher": ["$go"]
    },
    {
      "label": "Test CA",
      "type": "shell",
      "command": "go test ./internal/ca/... -count=1 -timeout=10m",
      "group": "test",
      "problemMatcher": ["$go"]
    },
    {
      "label": "Lint All",
      "type": "shell",
      "command": "golangci-lint run ./...",
      "group": "test",
      "problemMatcher": ["$go"]
    },
    {
      "label": "Format All",
      "type": "shell",
      "command": "golangci-lint run --fix ./...",
      "group": "none",
      "problemMatcher": ["$go"]
    },
    {
      "label": "Docker Compose Up",
      "type": "shell",
      "command": "docker compose -f ./deployments/compose/compose.yml up -d",
      "group": "none",
      "problemMatcher": []
    },
    {
      "label": "Docker Compose Down",
      "type": "shell",
      "command": "docker compose -f ./deployments/compose/compose.yml down -v",
      "group": "none",
      "problemMatcher": []
    },
    {
      "label": "Run E2E Tests",
      "type": "shell",
      "command": "go test -tags=e2e -v -timeout=30m ./internal/test/e2e/",
      "group": "test",
      "problemMatcher": ["$go"]
    },
    {
      "label": "Run DAST Tests (Quick)",
      "type": "shell",
      "command": "go run ./cmd/workflow -workflows=dast -inputs=\"scan_profile=quick\"",
      "group": "test",
      "problemMatcher": []
    },
    {
      "label": "Run All Workflows",
      "type": "shell",
      "command": "go run ./cmd/workflow -workflows=all",
      "group": "test",
      "problemMatcher": []
    },
    {
      "label": "CICD - Enforce Test Patterns",
      "type": "shell",
      "command": "go run ./cmd/cicd go-enforce-test-patterns",
      "group": "test",
      "problemMatcher": []
    },
    {
      "label": "CICD - Enforce Any (Remove interface{})",
      "type": "shell",
      "command": "go run ./cmd/cicd go-enforce-any",
      "group": "test",
      "problemMatcher": []
    },
    {
      "label": "CICD - Update Direct Dependencies",
      "type": "shell",
      "command": "go run ./cmd/cicd go-update-direct-dependencies",
      "group": "none",
      "problemMatcher": []
    },
    {
      "label": "Generate OpenAPI Code",
      "type": "shell",
      "command": "go generate ./api",
      "group": "build",
      "problemMatcher": ["$go"]
    }
  ]
}
```

**Task Categories:**

- **Build:** `Build All`, `Generate OpenAPI Code`
- **Test:** `Test All`, `Test KMS`, `Test Identity`, `Test CA`, `Run E2E Tests`, `Run DAST Tests`, `Run All Workflows`
- **Lint:** `Lint All`, `Format All`, `CICD - Enforce Test Patterns`, `CICD - Enforce Any`
- **Docker:** `Docker Compose Up`, `Docker Compose Down`
- **Maintenance:** `CICD - Update Direct Dependencies`

**Default Tasks:**

- **Build (default):** `Build All`
- **Test (default):** `Test All`

---

## Extension Recommendations

### Current .vscode/extensions.json

**Expected Extensions:**

- **Go:** `golang.go`
- **YAML:** `redhat.vscode-yaml`
- **Docker:** `ms-azuretools.vscode-docker`
- **Python:** `ms-python.python`
- **Copilot:** `GitHub.copilot`, `GitHub.copilot-chat`
- **Markdown:** `yzhang.markdown-all-in-one`
- **GitLens:** `eamodio.gitlens`
- **REST Client:** `humao.rest-client` (for API testing)
- **OpenAPI:** `42Crunch.vscode-openapi` (for OpenAPI editing)

**Refactor Impact:**

- **None** - Extensions are language/tooling-based, not package-based

**Recommendation:**

- Verify `.vscode/extensions.json` includes recommended extensions
- No updates needed for refactor

---

## VS Code Workspace Settings Validation

### gopls Diagnostic Validation

**After Refactor:**

1. Open VS Code
2. Navigate to refactored packages (e.g., `internal/kms/`, `pkg/crypto/`)
3. Verify gopls diagnostics work correctly:
   - Unused imports highlighted
   - Unused variables highlighted
   - Type inference shows correct types
   - Inlay hints display for parameters, types, constants

**Expected Behavior:**

- No gopls errors related to package paths
- Code completion works for new import paths
- Go to Definition works across service group boundaries

### Terminal Auto-Approval Validation

**Test Commands:**

```bash
# Should auto-approve (new patterns)
cryptoutil kms server start --help
cryptoutil identity authz start --help
cryptoutil ca server start --help

# Should auto-approve (existing patterns)
go test ./internal/kms/... -v
docker compose -f ./deployments/compose/compose.yml ps
git status

# Should require manual approval (destructive)
git push origin main
docker compose exec cryptoutil-sqlite sh
rm -rf ./test-output
```

**Validation:**

- Auto-approved commands execute immediately
- Manual approval commands prompt user

---

## Migration Checklist

### Phase 1: Identity Extraction

- [ ] No `.vscode/settings.json` changes needed
- [ ] Add Identity launch configurations to `.vscode/launch.json`
- [ ] Add auto-approval patterns for `cryptoutil identity` commands
- [ ] Create `.vscode/tasks.json` with service group tasks
- [ ] Verify gopls diagnostics work for Identity packages
- [ ] Test terminal auto-approval for Identity CLI commands

### Phase 2: KMS Extraction

- [ ] Update KMS launch configurations in `.vscode/launch.json`
- [ ] Add auto-approval patterns for `cryptoutil kms` commands
- [ ] Update tasks in `.vscode/tasks.json` (Test KMS, Build KMS)
- [ ] Verify gopls diagnostics work for KMS packages
- [ ] Test terminal auto-approval for KMS CLI commands
- [ ] Validate backward compatibility patterns (`kms cryptoutil server`)

### Phase 3: CA Preparation

- [ ] Add CA launch configurations to `.vscode/launch.json`
- [ ] Add auto-approval patterns for `cryptoutil ca` commands
- [ ] Update tasks in `.vscode/tasks.json` (Test CA, Build CA)
- [ ] Verify gopls diagnostics work for CA packages
- [ ] Test terminal auto-approval for CA CLI commands

### Post-Refactor Validation

- [ ] Run all VS Code launch configurations
- [ ] Execute all tasks from Command Palette
- [ ] Verify gopls provides correct diagnostics across all packages
- [ ] Test terminal auto-approval for all service group commands
- [ ] Confirm no gopls errors in Problems panel
- [ ] Validate inlay hints display correctly

---

## Cross-References

- **Group Directory Blueprint:** [docs/01-refactor/blueprint.md](./blueprint.md)
- **Import Alias Policy:** [docs/01-refactor/import-aliases.md](./import-aliases.md)
- **CLI Strategy Framework:** [docs/01-refactor/cli-strategy.md](./cli-strategy.md)
- **Pipeline Impact Assessment:** [docs/01-refactor/pipeline-impact.md](./pipeline-impact.md)

---

## Notes

- **gopls settings are stable** - `cryptoutil` prefix covers all packages post-refactor
- **Terminal auto-approval patterns** - Add 5 new patterns for service group CLI commands
- **Launch configurations** - 12 total post-refactor (3 KMS + 4 Identity + 2 CA + 3 OpenAPI)
- **tasks.json** - Create new file with 16 tasks (build, test, lint, docker, cicd, openapi)
- **File exclusions are path-agnostic** - No changes needed for `files.exclude`, `search.exclude`
- **YAML schemas are pattern-based** - No changes needed for workflow/compose file validation
- **Extension recommendations unchanged** - Language/tooling-based, not package-based
- **Validation is critical** - Test gopls, terminal auto-approval, and launch configurations after each phase
