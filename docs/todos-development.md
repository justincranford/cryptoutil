# Cryptoutil Development Workflow & Configuration TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 16, 2025
**Status**: Development workflow enhancements planned for ongoing maintenance - Pre-commit automation analysis added

---

## 🟢 LOW - Development Workflow & Configuration

### Task DW1: Implement 12-Factor App Standards Compliance
- **Description**: Ensure application follows 12-factor app methodology for cloud-native deployment
- **12-Factor Requirements**:
  - **I. Codebase**: One codebase tracked in revision control, many deploys - **Status: ✅ IMPLEMENTED** (Single Git repository with clear versioning)
  - **II. Dependencies**: Explicitly declare and isolate dependencies - **Status: ✅ IMPLEMENTED** (Go modules with explicit dependency management)
  - **III. Config**: Store config in the environment - **Status: ✅ IMPLEMENTED** (YAML configs + environment variables for secrets)
  - **IV. Backing services**: Treat backing services as attached resources - **Status: ✅ IMPLEMENTED** (Database via connection strings)
  - **V. Build, release, run**: Strictly separate build and run stages - **Status: ✅ IMPLEMENTED** (Dockerfile with distinct build/validation/runtime stages)
  - **VI. Processes**: Execute the app as one or more stateless processes - **Status: ❓ PARTIALLY IMPLEMENTED** (Appears stateless but needs verification)
  - **VII. Port binding**: Export services via port binding - **Status: ✅ IMPLEMENTED** (Binds to configurable ports 8080/9090)
  - **VIII. Concurrency**: Scale out via the process model - **Status: ❓ NEEDS AUDIT** (Horizontal scaling capability needs verification)
  - **IX. Disposability**: Maximize robustness with fast startup and graceful shutdown - **Status: ✅ IMPLEMENTED** (Signal handling + health checks)
  - **X. Dev/prod parity**: Keep development, staging, and production as similar as possible - **Status: ✅ IMPLEMENTED** (Docker compose environments)
  - **XI. Logs**: Treat logs as event streams - **Status: ✅ IMPLEMENTED** (Structured slog logging as event streams)
  - **XII. Admin processes**: Run admin/management tasks as one-off processes - **Status: ❓ NEEDS AUDIT** (Admin task separation needs verification)
- **Current State**: 8/12 factors fully implemented, 2 partially implemented, 2 need audit
- **Action Items**:
  - Audit Factor VI (stateless processes) - verify no local file storage or in-memory state
  - Audit Factor VIII (concurrency) - verify horizontal scaling capability with multiple instances
  - Audit Factor XII (admin processes) - verify admin tasks run as separate processes
  - Document final 12-factor compliance status
  - Update deployment configurations for any missing factors
- **Files**: Docker configs, deployment files, application architecture
- **Expected Outcome**: Cloud-native, scalable application following industry best practices
- **Priority**: LOW - Best practices alignment
- **Timeline**: Ongoing maintenance

### Task DW2: Implement Hot Config File Reload
- **Description**: Add ability to reload configuration files without restarting the server
- **Current State**: Configuration loaded only at startup
- **Action Items**:
  - Add file watcher for config files (development mode only)
  - Implement graceful config reload with validation
  - Add reload endpoint for runtime config updates
  - Handle config reload failures gracefully
  - Add configuration versioning/checksum validation
- **Files**: `internal/common/config/config.go`, server startup code
- **Expected Outcome**: Development workflow improvement with live config reloading
- **Priority**: LOW - Developer experience enhancement
- **Timeline**: Q1 2026

### Task DW4: Implement Parallel Step Execution
- **Context**: Currently all setup steps run sequentially, but some can run in parallel
- **Action Items**:
  - Run directory creation in background (`mkdir -p configs/test & mkdir -p ./dast-reports &`)
  - Parallelize config file creation with other setup tasks
  - Optimize application startup sequence
- **Files**: `.github/workflows/dast.yml` (Start application step)
- **Expected Savings**: ~10-15 seconds per run (minor optimization)
- **Priority**: Low - workflow already runs efficiently with scan profiles

### Task DW5: Evaluate and Configure Go Extension Settings
- **Description**: Systematically evaluate and configure VS Code Go extension settings for optimal development experience
- **Current State**: Basic settings configured, comprehensive evaluation needed
- **Action Items**: Review and decide on each setting below

#### 🟢 **RECOMMENDED TO INCLUDE** (High-value, safe settings)
- **`go.terminal.activateEnvironment: true`** - Default: `true`, Recommended: ✅ INCLUDE
  - Ensures Go environment variables are available in VS Code integrated terminals
  - Improves development workflow consistency
- **`go.testExplorer.enable: true`** - Default: `true`, Recommended: ✅ INCLUDE
  - Enables built-in test explorer for better test management
  - Provides visual test execution and results
- **`go.toolsManagement.autoUpdate: true`** - Default: `false`, Recommended: ✅ INCLUDE
  - Automatically keeps Go tools (gopls, etc.) updated
  - Ensures latest features and bug fixes
- **`go.tasks.provideDefault: true`** - Default: `true`, Recommended: ✅ INCLUDE
  - Provides default Go build/test tasks
  - Enables standard VS Code task integration
- **`go.enableCodeLens: true`** - Default: `true`, Recommended: ✅ INCLUDE
  - Shows run/debug test buttons above test functions
  - Improves test execution workflow
- **`go.editorContextMenuCommands.testAtCursor: true`** - Default: `true`, Recommended: ✅ INCLUDE
  - Adds "Run Test at Cursor" to editor context menu
  - Quick access to test execution
- **`go.editorContextMenuCommands.debugTestAtCursor: true`** - Default: `false`, Recommended: ✅ INCLUDE
  - Adds "Debug Test at Cursor" to editor context menu
  - Enables debugging individual tests
- **`go.inlayHints.parameterNames: true`** - Default: `false`, Recommended: ✅ INCLUDE
  - Shows parameter names in function calls
  - Improves code readability
- **`go.inlayHints.assignVariableTypes: true`** - Default: `false`, Recommended: ✅ INCLUDE
  - Shows variable types in assignments
  - Reduces need to hover for type information

#### 🟡 **RECOMMENDED TO EXCLUDE** (Performance/safety concerns)
- **`go.coverOnSave: false`** - Default: `false`, Recommended: ❌ EXCLUDE
  - Would run test coverage on every file save
  - Performance impact, especially with large test suites
- **`go.vetOnSave: "off"`** - Default: `"package"`, Recommended: ❌ EXCLUDE
  - Would run `go vet` on save (redundant with golangci-lint)
  - Performance impact, duplicate analysis
- **`go.lintOnSave: "workspace"`** - Default: `"package"`, Recommended: ❌ EXCLUDE
  - Would lint entire workspace on save
  - Significant performance impact on large projects

#### 🤔 **NEEDS EVALUATION** (Project-specific decision needed)
- **`go.buildOnSave: "off"`** - Default: `"package"`, Recommended: ❓ EVALUATE
  - Compiles code on save using 'go build'
  - May be useful for immediate feedback but can be slow
- **`go.testOnSave: false`** - Default: `false`, Recommended: ❓ EVALUATE
  - Runs tests on save for current package
  - Useful for TDD but may be disruptive with auto-save
- **`go.coverOnSingleTest: false`** - Default: `false`, Recommended: ❓ EVALUATE
  - Shows coverage when running individual tests
  - Useful for focused coverage analysis
- **`go.experiments.testExplorer: true`** - Default: `true`, Recommended: ❓ EVALUATE
  - Uses experimental test explorer (already enabled)
  - May have new features but could be unstable
- **`go.survey.prompt: true`** - Default: `true`, Recommended: ❓ EVALUATE
  - Prompts for Go developer surveys
  - Contributes to Go ecosystem but may be annoying

#### 🔧 **CONFIGURATION SETTINGS** (Set as needed)
- **`go.buildFlags: []`** - Default: `[]`, Recommended: ❓ CONFIGURE IF NEEDED
  - Additional flags for `go build`/`go test`
  - Use for custom build requirements (e.g., `["-ldflags='-s'"]`)
- **`go.buildTags: ""`** - Default: `""`, Recommended: ❓ CONFIGURE IF NEEDED
  - Build tags for conditional compilation
  - Use for platform-specific or feature-gated code
- **`go.testFlags: []`** - Default: `[]`, Recommended: ❓ CONFIGURE IF NEEDED
  - Additional flags for `go test`
  - Use for custom test configuration
- **`go.testTimeout: "30s"`** - Default: `"30s"`, Recommended: ❓ CONFIGURE IF NEEDED
  - Timeout for test execution
  - Increase for slow integration tests
- **`go.lintFlags: []`** - Default: `[]`, Recommended: ❓ CONFIGURE IF NEEDED
  - Additional flags for linter
  - Configure golangci-lint behavior
- **`go.toolsEnvVars: {}`** - Default: `{}`, Recommended: ❓ CONFIGURE IF NEEDED
  - Environment variables for Go tools
  - Use for CGO or custom tool configuration

#### 📊 **VISUALIZATION SETTINGS** (Personal preference)
- **`go.coverageDecorator: {type: "highlight", ...}`** - Default: complex object, Recommended: ❓ PERSONAL PREFERENCE
  - Configures test coverage visualization
  - Choose between highlight or gutter display
- **`go.coverMode: "default"`** - Default: `"default"`, Recommended: ❓ PERSONAL PREFERENCE
  - Code coverage mode (set/count/atomic)
  - "atomic" for concurrent programs
- **`go.showWelcome: true`** - Default: `true`, Recommended: ❓ PERSONAL PREFERENCE
  - Shows welcome screen on first install
  - Disable if you find it annoying

#### 🧪 **ADVANCED/EXPERIMENTAL SETTINGS** (Use with caution)
- **`go.delveConfig: {}`** - Default: complex object, Recommended: ❌ AVOID UNLESS NEEDED
  - Advanced debugger configuration
  - Only modify if you have specific debugging requirements
- **`go.trace.server: "off"`** - Default: `"off"`, Recommended: ❌ AVOID UNLESS DEBUGGING
  - Traces communication between VS Code and language server
  - Only enable for troubleshooting LSP issues
- **`go.languageServerFlags: []`** - Default: `[]`, Recommended: ❌ AVOID UNLESS NEEDED
  - Additional flags for gopls language server
  - Only modify for advanced LSP configuration

- **Files**: `.vscode/settings.json`
- **Expected Outcome**: Optimized VS Code Go development experience with appropriate settings
- **Priority**: LOW - Developer experience enhancement
- **Timeline**: Q4 2025

---

## 🟢 LOW - Documentation & API Management

### Task DOC1: API Versioning Strategy Documentation
- **Description**: Document comprehensive API versioning strategy and deprecation policy
- **Current State**: Basic API versioning exists but not formally documented
- **Action Items**:
  - Document API versioning conventions (URL-based, header-based, etc.)
  - Create API deprecation policy and timeline
  - Document backward compatibility guarantees
  - Create migration guides for API changes
- **Files**: `docs/api-versioning.md`, OpenAPI specifications
- **Expected Outcome**: Clear API evolution and compatibility guidelines
- **Priority**: Low - API management

### Task DOC2: Performance Benchmarks Documentation
- **Description**: Create comprehensive performance benchmarks and documentation
- **Current State**: Performance testing exists but not documented
- **Action Items**:
  - Document performance benchmarks for key operations
  - Create performance comparison charts and metrics
  - Document performance testing methodology
  - Add performance expectations to API documentation
- **Files**: `docs/performance-benchmarks.md`, benchmark results
- **Expected Outcome**: Performance transparency and expectations
- **Priority**: Low - Documentation enhancement
