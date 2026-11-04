// Copyright (c) 2025-2026 Justin Cranford.
// Package golangci_runner runs golangci-lint on explicitly enumerated repository packages.
package golangci_runner

import (
	"context"
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf16"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

const (
	golangciExecutable           = "golangci-lint"
	buildTagsFlag                = "--build-tags"
	timeoutFlag                  = "--timeout"
	newFromRevFlag               = "--new-from-rev"
	maxPackagesPerBatch          = 64
	maxEstimatedCommandLineUTF16 = 24_000
)

type (
	commandRunner func(context.Context, string, []string, string) error
	lookPathFunc  func(string) (string, error)
	readDirFunc   func(string) ([]os.DirEntry, error)
)

// Run invokes golangci-lint with the provided supported flags and explicit
// repository-safe Go package patterns. It never accepts caller package patterns.
func Run(ctx context.Context, rootDir string, args []string) error {
	return runWithFunctions(ctx, rootDir, args, os.Lstat, os.ReadDir, build.Default, exec.LookPath, execute)
}

func runWithFunctions(
	ctx context.Context,
	rootDir string,
	args []string,
	lstatFn func(string) (os.FileInfo, error),
	readDir readDirFunc,
	buildContext build.Context,
	lookPath lookPathFunc,
	runCommand commandRunner,
) error {
	if ctx == nil {
		return errors.New("missing command context")
	}

	if strings.TrimSpace(rootDir) == "" {
		return errors.New("missing repository root")
	}

	if lookPath == nil {
		return errors.New("missing executable lookup function")
	}

	if runCommand == nil {
		return errors.New("missing command runner")
	}

	validatedArgs, err := validateArgs(args)
	if err != nil {
		return err
	}

	packages, err := cryptoutilSharedUtilFiles.ListRepositoryGoPackagesWithLstat(rootDir, lstatFn)
	if err != nil {
		return fmt.Errorf("failed to enumerate repository Go packages: %w", err)
	}

	buildContext.BuildTags = appendUnique(buildContext.BuildTags, buildTagsFromArgs(validatedArgs)...)

	eligiblePackages, err := filterEligiblePackages(rootDir, packages.PackagePatterns, buildContext, readDir)
	if err != nil {
		return fmt.Errorf("failed to determine eligible repository Go packages: %w", err)
	}

	if len(eligiblePackages) == 0 {
		return fmt.Errorf("no eligible repository Go packages found under %s", rootDir)
	}

	executable, err := lookPath(golangciExecutable)
	if err != nil {
		return fmt.Errorf("failed to locate %s: %w", golangciExecutable, err)
	}

	commandPrefix := make([]string, 0, 1+len(validatedArgs))
	commandPrefix = append(commandPrefix, cryptoutilSharedMagic.WorkflowSubCmdRun)
	commandPrefix = append(commandPrefix, validatedArgs...)

	batches, err := packageBatches(executable, commandPrefix, eligiblePackages, packageBatchLimits{
		maxPackages:         maxPackagesPerBatch,
		maxCommandLineUTF16: maxEstimatedCommandLineUTF16,
	})
	if err != nil {
		return err
	}

	var batchErrors []error

	for batchIndex, packageBatch := range batches {
		commandArgs := append(append([]string(nil), commandPrefix...), packageBatch...)

		runErr := runCommand(ctx, executable, commandArgs, rootDir)
		if runErr == nil {
			continue
		}

		batchErrors = append(batchErrors, fmt.Errorf("batch %d/%d: %w", batchIndex+1, len(batches), runErr))

		if ctx.Err() != nil {
			break
		}
	}

	if len(batchErrors) > 0 {
		return fmt.Errorf(
			"golangci-lint failed in %d of %d package batches: %w",
			len(batchErrors),
			len(batches),
			errors.Join(batchErrors...),
		)
	}

	return nil
}

func validateArgs(args []string) ([]string, error) {
	validated := make([]string, 0, len(args))
	seen := make(map[string]bool)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--fix":
			if seen[arg] {
				return nil, fmt.Errorf("duplicate golangci-lint flag %q", arg)
			}

			seen[arg] = true
			validated = append(validated, arg)
		case arg == timeoutFlag || arg == newFromRevFlag || arg == buildTagsFlag:
			if seen[arg] {
				return nil, fmt.Errorf("duplicate golangci-lint flag %q", arg)
			}

			if i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "-") {
				return nil, fmt.Errorf("golangci-lint flag %q requires a value", arg)
			}

			value := args[i+1]
			if arg == timeoutFlag {
				duration, durationErr := time.ParseDuration(value)
				if durationErr != nil {
					return nil, fmt.Errorf("invalid golangci-lint timeout %q: %w", value, durationErr)
				}

				if duration <= 0 {
					return nil, fmt.Errorf("invalid golangci-lint timeout %q: must be positive", value)
				}
			}

			if arg == newFromRevFlag && !validRevisionValue(value) {
				return nil, fmt.Errorf("invalid golangci-lint revision %q", value)
			}

			if arg == buildTagsFlag {
				if _, tagErr := parseBuildTags(value); tagErr != nil {
					return nil, tagErr
				}
			}

			seen[arg] = true
			validated = append(validated, arg, value)
			i++
		case strings.HasPrefix(arg, "--timeout="):
			if seen[timeoutFlag] {
				return nil, errors.New("duplicate golangci-lint flag \"--timeout\"")
			}

			value := strings.TrimPrefix(arg, "--timeout=")

			duration, durationErr := time.ParseDuration(value)
			if durationErr != nil {
				return nil, fmt.Errorf("invalid golangci-lint timeout %q: %w", value, durationErr)
			}

			if duration <= 0 {
				return nil, fmt.Errorf("invalid golangci-lint timeout %q: must be positive", value)
			}

			seen[timeoutFlag] = true

			validated = append(validated, arg)
		case strings.HasPrefix(arg, "--new-from-rev="):
			if seen[newFromRevFlag] {
				return nil, errors.New("duplicate golangci-lint flag \"--new-from-rev\"")
			}

			value := strings.TrimPrefix(arg, "--new-from-rev=")
			if !validRevisionValue(value) {
				return nil, fmt.Errorf("invalid golangci-lint revision %q", value)
			}

			seen[newFromRevFlag] = true

			validated = append(validated, arg)
		case strings.HasPrefix(arg, "--build-tags="):
			if seen[buildTagsFlag] {
				return nil, errors.New("duplicate golangci-lint flag \"--build-tags\"")
			}

			value := strings.TrimPrefix(arg, "--build-tags=")
			if _, tagErr := parseBuildTags(value); tagErr != nil {
				return nil, tagErr
			}

			seen[buildTagsFlag] = true

			validated = append(validated, arg)
		default:
			if !strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("package argument %q is not allowed; package patterns are discovered by the wrapper", arg)
			}

			return nil, fmt.Errorf("unsupported golangci-lint flag %q", arg)
		}
	}

	return validated, nil
}

func validRevisionValue(value string) bool {
	if value == "" || strings.TrimSpace(value) != value || strings.HasPrefix(value, "-") {
		return false
	}

	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}

	return true
}

func parseBuildTags(value string) ([]string, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return nil, errors.New("golangci-lint flag \"--build-tags\" requires comma-separated tag identifiers")
	}

	tags := strings.Split(value, ",")
	for _, tag := range tags {
		if tag == "" {
			return nil, fmt.Errorf("invalid empty build tag in %q", value)
		}

		for characterIndex, character := range tag {
			letter := character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z'
			if characterIndex == 0 {
				if !letter && character != '_' {
					return nil, fmt.Errorf("invalid build tag %q", tag)
				}

				continue
			}

			if !letter && (character < '0' || character > '9') && character != '_' && character != '.' {
				return nil, fmt.Errorf("invalid build tag %q", tag)
			}
		}
	}

	return tags, nil
}

func buildTagsFromArgs(args []string) []string {
	var tags []string

	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == buildTagsFlag:
			if i+1 < len(args) {
				parsedTags, _ := parseBuildTags(args[i+1])
				tags = append(tags, parsedTags...)
				i++
			}
		case strings.HasPrefix(args[i], "--build-tags="):
			parsedTags, _ := parseBuildTags(strings.TrimPrefix(args[i], "--build-tags="))
			tags = append(tags, parsedTags...)
		}
	}

	return tags
}

func appendUnique(existing []string, additions ...string) []string {
	seen := make(map[string]struct{}, len(existing)+len(additions))
	result := make([]string, 0, len(existing)+len(additions))
	combined := make([]string, 0, len(existing)+len(additions))
	combined = append(combined, existing...)
	combined = append(combined, additions...)

	for _, value := range combined {
		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func filterEligiblePackages(
	rootDir string,
	packagePatterns []string,
	buildContext build.Context,
	readDir readDirFunc,
) ([]string, error) {
	if readDir == nil {
		return nil, errors.New("missing directory reader")
	}

	if buildContext.GOOS == "" || buildContext.GOARCH == "" {
		return nil, fmt.Errorf("incomplete Go build context: GOOS=%q GOARCH=%q", buildContext.GOOS, buildContext.GOARCH)
	}

	eligible := make([]string, 0, len(packagePatterns))
	for _, packagePattern := range packagePatterns {
		packageDir := rootDir

		if packagePattern != "." {
			if !strings.HasPrefix(packagePattern, "./") {
				return nil, fmt.Errorf("unexpected discovered package pattern %q", packagePattern)
			}

			relativePackageDir := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(packagePattern, "./")))
			if filepath.IsAbs(relativePackageDir) || relativePackageDir == ".." ||
				strings.HasPrefix(relativePackageDir, ".."+string(filepath.Separator)) {
				return nil, fmt.Errorf("discovered package pattern escapes repository root: %q", packagePattern)
			}

			packageDir = filepath.Join(rootDir, relativePackageDir)
		}

		entries, err := readDir(packageDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read Go package directory %s: %w", packageDir, err)
		}

		hasEligibleGoFile := false

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}

			matchesBuild, matchErr := buildContext.MatchFile(packageDir, entry.Name())
			if matchErr != nil {
				return nil, fmt.Errorf("failed to evaluate Go build constraints for %s: %w", filepath.Join(packageDir, entry.Name()), matchErr)
			}

			if matchesBuild {
				hasEligibleGoFile = true

				break
			}
		}

		if hasEligibleGoFile {
			eligible = append(eligible, packagePattern)
		}
	}

	return eligible, nil
}

type packageBatchLimits struct {
	maxPackages         int
	maxCommandLineUTF16 int
}

func packageBatches(
	executable string,
	commandPrefix []string,
	packagePatterns []string,
	limits packageBatchLimits,
) ([][]string, error) {
	if limits.maxPackages < 1 {
		return nil, errors.New("maximum packages per batch must be positive")
	}

	if limits.maxCommandLineUTF16 < 1 {
		return nil, errors.New("maximum command-line length must be positive")
	}

	if len(packagePatterns) == 0 {
		return nil, errors.New("cannot create package batches from an empty package set")
	}

	batches := make([][]string, 0, (len(packagePatterns)+limits.maxPackages-1)/limits.maxPackages)
	current := make([]string, 0, limits.maxPackages)

	for _, packagePattern := range packagePatterns {
		candidate := append(append([]string(nil), current...), packagePattern)
		if current != nil && (len(current) >= limits.maxPackages ||
			estimatedCommandLineUTF16(executable, commandPrefix, candidate) > limits.maxCommandLineUTF16) {
			batches = append(batches, current)
			current = nil
			candidate = []string{packagePattern}
		}

		if estimatedCommandLineUTF16(executable, commandPrefix, candidate) > limits.maxCommandLineUTF16 {
			return nil, fmt.Errorf("package pattern %q exceeds the safe command-line length limit", packagePattern)
		}

		current = append(current, packagePattern)
	}

	if len(current) > 0 {
		batches = append(batches, current)
	}

	return batches, nil
}

func estimatedCommandLineUTF16(executable string, commandPrefix, packagePatterns []string) int {
	total := estimatedArgumentUTF16(executable)
	for _, argument := range commandPrefix {
		total += estimatedArgumentUTF16(argument)
	}

	for _, packagePattern := range packagePatterns {
		total += estimatedArgumentUTF16(packagePattern)
	}

	return total
}

func estimatedArgumentUTF16(argument string) int {
	// Double each code unit to conservatively account for Windows quoting and
	// backslash escaping, then reserve delimiters and the terminating NUL.
	return 2*len(utf16.Encode([]rune(argument))) + 2
}

func execute(ctx context.Context, executable string, args []string, rootDir string) error {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute golangci-lint: %w", err)
	}

	return nil
}
