package magic

// DevSetup tool types for dependency categorization.
const (
	DevSetupToolTypeGo           = "go"
	DevSetupToolTypeGoModule     = "go-module"
	DevSetupToolTypePython       = "python"
	DevSetupToolTypePythonModule = "python-module"
	DevSetupToolTypeSystemBinary = "system-binary"
	DevSetupToolTypeUV           = "uv"
)

// DevSetup tool check/install commands.
const (
	DevSetupPythonCheckCmd      = "python --version"
	DevSetupPythonInstallCmd    = "python -m venv .venv --upgrade-deps"
	DevSetupUVCheckCmd          = "uv --version"
	DevSetupUVInstallCmd        = "curl -LsSf https://astral.sh/uv/install.sh | sh"
	DevSetupPreCommitCheckCmd   = "pre-commit --version"
	DevSetupPreCommitInstallCmd = "uv pip install pre-commit"
	DevSetupSemgrepCheckCmd     = "semgrep --version"
	DevSetupSemgrepInstallCmd   = "uv pip install semgrep"
	DevSetupYamllintCheckCmd    = "yamllint --version"
	DevSetupYamllintInstallCmd  = "uv pip install yamllint"
	DevSetupGoCheckCmd          = "go version"
	DevSetupGoInstallCmd        = "curl -LsSf https://go.dev/dl/go1.26.1.linux-amd64.tar.gz | tar -C $HOME -xzf -"
	DevSetupGolangciCheckCmd    = "golangci-lint --version"
	DevSetupGolangciInstallCmd  = "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
)

// DevSetup status codes for dependency checks.
const (
	DevSetupStatusInstalled     = "installed"
	DevSetupStatusCheckFailed   = "check-failed"
	DevSetupStatusInstallFailed = "install-failed"
	DevSetupStatusVerifyFailed  = "verify-failed"
	DevSetupStatusIconSuccess   = "✓"
	DevSetupStatusIconFailure   = "✗"
)

// DevSetup test constants for group names.
const (
	DevSetupTestGroupPythonRuntime      = "Python Runtime"
	DevSetupTestGroupUVPackageManager   = "UV Package Manager"
	DevSetupTestGroupPythonDependencies = "Python Dependencies"
	DevSetupTestGroupGoToolchain        = "Go Toolchain"
	DevSetupTestGroupGoLinting          = "Go Linting"
	DevSetupTestGroupPreCommitHooks     = "Pre-commit Hooks"
)

// DevSetup test constants for checker tests.
const (
	DevSetupTestPythonCmd           = "python --version"
	DevSetupTestPythonVersion       = "Python 3.14.0"
	DevSetupTestPythonVersionParsed = "3.14.0"
	DevSetupTestGoCmd               = "go version"
	DevSetupTestGoVersionOutput     = "go version go1.26.1 linux/amd64"
	DevSetupTestGoVersionParsed     = "1.26.1"
	DevSetupTestGolangciCmd         = "golangci-lint --version"
	DevSetupTestGolangciOutput      = "golangci-lint has version v2.12.2"
	DevSetupTestGolangciVersion     = "2.12.2"
	DevSetupTestInvalidName         = "invalid"
	DevSetupTestVersion123          = "1.2.3"
	DevSetupTestNoVersionHere       = "no version here"
)

// DevSetup test constants for installer tests.
const (
	DevSetupTestInvalidToolName = "invalid"
	DevSetupTestSomeStderr      = "some stderr"
)

// DevSetup test constants for reporter tests.
const (
	DevSetupTestGroupNamePython        = "Python"
	DevSetupTestToolNamePython         = "python"
	DevSetupTestVersionPythonRequired  = "3.14"
	DevSetupTestVersionPythonActual    = "3.14.0"
	DevSetupTestGroupNameGo            = "Go"
	DevSetupTestToolNameGo             = "go"
	DevSetupTestVersionGoRequired      = "1.26.1"
	DevSetupTestVersionGoActual        = "1.26.1"
	DevSetupTestGroupNamePythonRuntime = "Python Runtime"
	DevSetupTestVersionPythonNext      = "24.0"
	DevSetupTestGroupNameGoToolchain   = "Go Toolchain"
	DevSetupTestVersionGoNext          = "1.27.0"
)
