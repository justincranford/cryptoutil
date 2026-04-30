// Copyright (c) 2025-2026 Justin Cranford.
// Package precommit_cicd_architecture enforces bulk cicd-lint hook architecture
// in .pre-commit-config.yaml.
package precommit_cicd_architecture

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

const (
	preCommitConfigPath = ".pre-commit-config.yaml"
	stagePreCommit      = "pre-commit"
	stagePrePush        = "pre-push"
)

type preCommitConfig struct {
	Repos []preCommitRepo `yaml:"repos"`
}

type preCommitRepo struct {
	Hooks []preCommitHook `yaml:"hooks"`
}

type preCommitHook struct {
	ID            string   `yaml:"id"`
	Entry         string   `yaml:"entry"`
	Args          []string `yaml:"args"`
	Stages        []string `yaml:"stages"`
	RequireSerial bool     `yaml:"require_serial"`
}

type bulkHookInfo struct {
	ID            string
	Stage         string
	RequireSerial bool
	Commands      []string
	kind          hookKind
}

type hookKind int

const (
	hookKindUnknown hookKind = iota
	hookKindLint
	hookKindFormat
	hookKindMixed
)

// Check validates .pre-commit-config.yaml from the current directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates .pre-commit-config.yaml from rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking pre-commit cicd bulk hook architecture...")

	configPath := filepath.Join(rootDir, preCommitConfigPath)

	cfg, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("precommit-cicd-architecture: %w", err)
	}

	bulkHooks := collectBulkHooks(cfg)
	violations := validateBulkHooks(bulkHooks)

	if len(violations) > 0 {
		sort.Strings(violations)

		return fmt.Errorf("precommit-cicd-architecture violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("precommit-cicd-architecture: bulk lint/format hook architecture is valid")

	return nil
}

func loadConfig(configPath string) (*preCommitConfig, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", preCommitConfigPath, err)
	}

	var cfg preCommitConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", preCommitConfigPath, err)
	}

	return &cfg, nil
}

func collectBulkHooks(cfg *preCommitConfig) []bulkHookInfo {
	bulkHooks := make([]bulkHookInfo, 0)

	for _, repo := range cfg.Repos {
		for _, hook := range repo.Hooks {
			commands := extractCICDCommands(hook.Args)
			if len(commands) == 0 {
				continue
			}

			stage := deriveStage(hook.Stages)
			bulkHooks = append(bulkHooks, bulkHookInfo{
				ID:            hook.ID,
				Stage:         stage,
				RequireSerial: hook.RequireSerial,
				Commands:      commands,
				kind:          classifyCommands(commands),
			})
		}
	}

	return bulkHooks
}

func deriveStage(stages []string) string {
	if len(stages) == 0 {
		return stagePreCommit
	}

	return stages[0]
}

func extractCICDCommands(args []string) []string {
	if len(args) == 0 {
		return nil
	}

	binaryIndex := -1

	for i, arg := range args {
		if arg == "cmd/cicd-lint/main.go" || arg == "./cmd/cicd-lint/main.go" {
			binaryIndex = i

			break
		}
	}

	if binaryIndex < 0 || binaryIndex+1 >= len(args) {
		return nil
	}

	commands := make([]string, 0, len(args)-(binaryIndex+1))

	for _, arg := range args[binaryIndex+1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}

		if cryptoutilSharedMagic.ValidCommands[arg] {
			commands = append(commands, arg)
		}
	}

	return commands
}

func classifyCommands(commands []string) hookKind {
	hasLint := false
	hasFormat := false

	for _, command := range commands {
		switch {
		case strings.HasPrefix(command, "lint-"):
			hasLint = true
		case strings.HasPrefix(command, "format-"):
			hasFormat = true
		}
	}

	switch {
	case hasLint && hasFormat:
		return hookKindMixed
	case hasLint:
		return hookKindLint
	case hasFormat:
		return hookKindFormat
	default:
		return hookKindUnknown
	}
}

func validateBulkHooks(bulkHooks []bulkHookInfo) []string {
	violations := make([]string, 0)

	expectedStages := []string{stagePreCommit, stagePrePush}

	lintCommandsCoverage := make(map[string]int)
	formatCommandsCoverage := make(map[string]int)
	seenStageKind := make(map[string]int)

	for _, hook := range bulkHooks {
		switch hook.kind {
		case hookKindLint, hookKindFormat:
			// Valid kinds continue to stage and serial/coverage validation below.
		case hookKindMixed:
			violations = append(violations, fmt.Sprintf("hook %q mixes lint-* and format-* commands; keep lint and format bulk calls mutually exclusive", hook.ID))

			continue
		case hookKindUnknown:
			continue
		}

		if hook.Stage != stagePreCommit && hook.Stage != stagePrePush {
			violations = append(violations, fmt.Sprintf("hook %q uses unsupported stage %q; expected pre-commit or pre-push", hook.ID, hook.Stage))

			continue
		}

		switch hook.kind {
		case hookKindLint:
			if hook.RequireSerial {
				violations = append(violations, fmt.Sprintf("hook %q stage %q is lint-only but require_serial=true; lint bulk calls must be concurrent-safe", hook.ID, hook.Stage))
			}

			seenStageKind[hook.Stage+"|lint"]++
			for _, command := range hook.Commands {
				if strings.HasPrefix(command, "lint-") {
					lintCommandsCoverage[command]++
				}
			}
		case hookKindFormat:
			if !hook.RequireSerial {
				violations = append(violations, fmt.Sprintf("hook %q stage %q is format-only but require_serial=false; format bulk calls must be serial", hook.ID, hook.Stage))
			}

			seenStageKind[hook.Stage+"|format"]++
			for _, command := range hook.Commands {
				if strings.HasPrefix(command, "format-") {
					formatCommandsCoverage[command]++
				}
			}
		case hookKindUnknown, hookKindMixed:
			// Already handled by the first kind switch.
		}
	}

	for _, stage := range expectedStages {
		switch {
		case seenStageKind[stage+"|lint"] != 1:
			violations = append(violations, fmt.Sprintf("expected exactly one lint-only bulk cicd hook for stage %q", stage))
		}

		switch {
		case seenStageKind[stage+"|format"] != 1:
			violations = append(violations, fmt.Sprintf("expected exactly one format-only bulk cicd hook for stage %q", stage))
		}
	}

	violations = append(violations, validateCommandCoverage(lintCommandsCoverage, "lint-")...)
	violations = append(violations, validateCommandCoverage(formatCommandsCoverage, "format-")...)

	return violations
}

func validateCommandCoverage(commandCoverage map[string]int, prefix string) []string {
	violations := make([]string, 0)

	expected := expectedCommandsForPrefix(prefix)
	for _, command := range expected {
		if commandCoverage[command] == 0 {
			violations = append(violations, fmt.Sprintf("command %q is not present in any %s bulk hook", command, prefix))
		}
	}

	return violations
}

func expectedCommandsForPrefix(prefix string) []string {
	commands := make([]string, 0)

	for command := range cryptoutilSharedMagic.ValidCommands {
		if strings.HasPrefix(command, prefix) {
			commands = append(commands, command)
		}
	}

	sort.Strings(commands)

	return commands
}
