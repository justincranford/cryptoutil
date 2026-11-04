// Copyright (c) 2025-2026 Justin Cranford.
// Package cicd_checkov runs the pinned Checkov CLI with repository-safe local exclusions.
package cicd_checkov

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedRepositoryScan "cryptoutil/internal/shared/util/repositoryscan"
)

const (
	maxNestedSkipPathPatternSize = 16 * 1024
	checkovSupportArgument       = "support"
)

var blockedCheckovArguments = map[string]struct{}{
	"bc-api-key":                {},
	"bc-api-url":                {},
	"config":                    {},
	"config-file":               {},
	"config-file-path":          {},
	"download-external-modules": {},
	"external-checks-git":       {},
	"prisma-api-url":            {},
	"repo-id":                   {},
	checkovSupportArgument:      {},
}

// Runner configures one local Checkov directory-mode invocation.
type Runner struct {
	Root        string
	CheckovPath string
	Args        []string
	Environment []string
	Stdout      io.Writer
	Stderr      io.Writer
}

// Run scans the complete repository directory with local-only Checkov settings.
func (runner Runner) Run(ctx context.Context) (runErr error) {
	if ctx == nil {
		return errors.New("checkov runner context must not be nil")
	}

	args, err := normalizeCheckovArguments(runner.Args)
	if err != nil {
		return err
	}

	root, err := canonicalRepositoryRoot(runner.Root)
	if err != nil {
		return err
	}

	if err := rejectRemoteCheckovConfiguration(root); err != nil {
		return err
	}

	checkovPath, err := resolveCheckovPath(runner.CheckovPath)
	if err != nil {
		return err
	}

	homeDirectory, err := os.MkdirTemp("", "cryptoutil-checkov-home-")
	if err != nil {
		return fmt.Errorf("create isolated Checkov home directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(homeDirectory); cleanupErr != nil {
			runErr = errors.Join(runErr, fmt.Errorf("remove isolated Checkov home directory: %w", cleanupErr))
		}
	}()

	sourceEnvironment := runner.Environment
	if sourceEnvironment == nil {
		sourceEnvironment = os.Environ()
	}

	environment := LocalCheckovEnvironment(sourceEnvironment, homeDirectory)
	if err := verifyPinnedCheckov(ctx, checkovPath, root, environment); err != nil {
		return err
	}

	nestedGitRoots, err := DiscoverNestedGitRoots(root, args)
	if err != nil {
		return err
	}

	args, err = appendLocalOnlyArguments(args, root, nestedGitRoots)
	if err != nil {
		return err
	}

	command, err := pinnedCommandContext(ctx, checkovPath, args...)
	if err != nil {
		return err
	}

	command.Dir = root
	command.Env = environment

	command.Stdout = runner.Stdout
	if command.Stdout == nil {
		command.Stdout = os.Stdout
	}

	command.Stderr = runner.Stderr
	if command.Stderr == nil {
		command.Stderr = os.Stderr
	}

	if err := command.Run(); err != nil {
		return fmt.Errorf("run local Checkov scan: %w", err)
	}

	return nil
}

// DiscoverNestedGitRoots returns nested directories containing a .git marker.
// Checkov skip-path arguments are honored before descending into excluded directories.
func DiscoverNestedGitRoots(root string, checkovArguments []string) ([]string, error) {
	root, err := canonicalRepositoryRoot(root)
	if err != nil {
		return nil, err
	}

	skipPathValues, err := checkovSkipPathValues(checkovArguments)
	if err != nil {
		return nil, err
	}

	skipPathPatterns, err := compileCheckovSkipPathPatterns(skipPathValues)
	if err != nil {
		return nil, err
	}

	nestedGitRoots := make([]string, 0)

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("discover nested Git checkouts below %s: %w", root, walkErr)
		}

		if path == root || !entry.IsDir() {
			return nil
		}

		relativePath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return fmt.Errorf("resolve nested Git checkout path below %s: %w", root, relErr)
		}

		if isRootGitMetadataDirectory(relativePath) || matchesCheckovSkipPath(root, relativePath, skipPathPatterns) {
			return fs.SkipDir
		}

		skipDirectory, skipErr := cryptoutilSharedRepositoryScan.ShouldSkipDir(
			root,
			path,
			func(string) bool { return false },
		)
		if skipErr != nil {
			return fmt.Errorf("inspect nested Git checkout at %s: %w", relativePath, skipErr)
		}

		if skipDirectory {
			nestedGitRoots = append(nestedGitRoots, filepath.Clean(relativePath))

			return fs.SkipDir
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk nested Git checkouts: %w", err)
	}

	sort.Strings(nestedGitRoots)

	return nestedGitRoots, nil
}

func normalizeCheckovArguments(arguments []string) ([]string, error) {
	args := append([]string(nil), arguments...)
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if len(args) == 0 {
		return nil, errors.New("checkov runner requires directory-mode arguments")
	}

	for _, argument := range args {
		if !strings.HasPrefix(argument, "-") {
			continue
		}

		flagName, _, _ := strings.Cut(strings.TrimLeft(argument, "-"), "=")

		flagName = strings.ToLower(strings.ReplaceAll(flagName, "_", "-"))
		if _, blocked := blockedCheckovArguments[flagName]; blocked {
			return nil, fmt.Errorf("checkov argument --%s is prohibited in local-only mode", flagName)
		}
	}

	if err := requireRepositoryDirectoryMode(args); err != nil {
		return nil, err
	}

	return args, nil
}

func requireRepositoryDirectoryMode(args []string) error {
	directoryCount := 0

	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == "-f" || argument == "--file" || strings.HasPrefix(argument, "--file=") {
			return errors.New("checkov runner requires a full -d . directory scan, not file mode")
		}

		if argument == "-d" || argument == "--directory" {
			if index+1 >= len(args) {
				return fmt.Errorf("checkov directory argument %s is missing its path", argument)
			}

			directoryCount++

			if args[index+1] != cryptoutilSharedMagic.CurrentDirectory {
				return fmt.Errorf("checkov directory argument must be ., got %q", args[index+1])
			}

			index++

			continue
		}

		if strings.HasPrefix(argument, "--directory=") {
			directoryCount++

			if strings.TrimPrefix(argument, "--directory=") != cryptoutilSharedMagic.CurrentDirectory {
				return fmt.Errorf("checkov directory argument must be ., got %q", strings.TrimPrefix(argument, "--directory="))
			}

			continue
		}

		if strings.HasPrefix(argument, "-d=") {
			directoryCount++

			if strings.TrimPrefix(argument, "-d=") != cryptoutilSharedMagic.CurrentDirectory {
				return fmt.Errorf("checkov directory argument must be ., got %q", strings.TrimPrefix(argument, "-d="))
			}
		}
	}

	if directoryCount != 1 {
		return fmt.Errorf("checkov runner requires exactly one -d . directory scan, found %d", directoryCount)
	}

	return nil
}

func appendLocalOnlyArguments(args []string, root string, nestedGitRoots []string) ([]string, error) {
	result := append([]string(nil), args...)

	for _, requiredFlag := range []string{
		"--skip-results-upload",
		"--skip-download",
		"--include-all-checkov-policies",
	} {
		found := false

		for _, argument := range result {
			if argument == requiredFlag {
				found = true

				break
			}

			if strings.HasPrefix(argument, requiredFlag+"=") {
				return nil, fmt.Errorf(
					"checkov argument %s must be enabled without an override in local-only mode",
					requiredFlag,
				)
			}
		}

		if !found {
			result = append(result, requiredFlag)
		}
	}

	if len(nestedGitRoots) == 0 {
		return result, nil
	}

	pattern, err := buildNestedGitSkipPathPattern(root, nestedGitRoots)
	if err != nil {
		return nil, err
	}

	result = append(result, "--skip-path", pattern)

	return result, nil
}

func checkovSkipPathValues(args []string) ([]string, error) {
	skipPathValues := make([]string, 0)

	for index := 0; index < len(args); index++ {
		argument := args[index]
		if strings.HasPrefix(argument, "--skip-path=") {
			value := strings.TrimPrefix(argument, "--skip-path=")
			if value == "" {
				return nil, errors.New("checkov --skip-path value must not be empty")
			}

			skipPathValues = append(skipPathValues, value)

			continue
		}

		if argument != "--skip-path" {
			continue
		}

		if index+1 >= len(args) || args[index+1] == "" {
			return nil, errors.New("checkov --skip-path argument is missing its pattern")
		}

		skipPathValues = append(skipPathValues, args[index+1])
		index++
	}

	return skipPathValues, nil
}

func compileCheckovSkipPathPatterns(values []string) ([]*regexp.Regexp, error) {
	patterns := make([]*regexp.Regexp, 0, len(values))
	for _, value := range values {
		// Checkov uses Python's non-capturing groups; Go's RE2-compatible regexp
		// package uses ordinary groups with the same matching behavior here.
		compatiblePattern := strings.ReplaceAll(value, "(?:", "(")

		pattern, err := regexp.Compile(compatiblePattern)
		if err != nil {
			return nil, fmt.Errorf("compile Checkov --skip-path pattern for safe discovery: %w", err)
		}

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

func matchesCheckovSkipPath(root, relativePath string, patterns []*regexp.Regexp) bool {
	if len(patterns) == 0 {
		return false
	}

	relativeWithDot := "." + string(os.PathSeparator) + relativePath
	candidates := []string{
		relativePath,
		filepath.ToSlash(relativePath),
		relativeWithDot,
		filepath.ToSlash(relativeWithDot),
		filepath.Join(root, relativePath),
		filepath.ToSlash(filepath.Join(root, relativePath)),
	}

	for _, pattern := range patterns {
		for _, candidate := range candidates {
			if pattern.MatchString(candidate) {
				return true
			}
		}
	}

	return false
}

func isRootGitMetadataDirectory(relativePath string) bool {
	return filepath.Dir(relativePath) == cryptoutilSharedMagic.CurrentDirectory && strings.EqualFold(filepath.Base(relativePath), cryptoutilSharedMagic.CICDExcludeDirGit)
}

func buildNestedGitSkipPathPattern(root string, nestedGitRoots []string) (string, error) {
	relativePatterns := make([]string, 0, len(nestedGitRoots))
	for _, nestedGitRoot := range nestedGitRoots {
		cleanedPath := filepath.Clean(nestedGitRoot)
		if filepath.IsAbs(cleanedPath) || cleanedPath == cryptoutilSharedMagic.CurrentDirectory || cleanedPath == ".." ||
			strings.HasPrefix(cleanedPath, ".."+string(os.PathSeparator)) {
			return "", fmt.Errorf("nested Git checkout path must be relative to the repository root: %q", nestedGitRoot)
		}

		relativePatterns = append(relativePatterns, filesystemPathPattern(cleanedPath))
	}

	sort.Strings(relativePatterns)

	rootPattern := filesystemPathPattern(root)

	pattern := `^(?:` + rootPattern + `[\\/]|(?:\.[\\/])?)` +
		`(?:` + strings.Join(relativePatterns, "|") + `)(?:[\\/]|$)`
	if len(pattern) > maxNestedSkipPathPatternSize {
		return "", fmt.Errorf(
			"nested Git checkout exclusions exceed the safe Checkov argument limit of %d bytes",
			maxNestedSkipPathPatternSize,
		)
	}

	return pattern, nil
}

func filesystemPathPattern(path string) string {
	normalizedPath := filepath.ToSlash(filepath.Clean(path))
	prefix := ""

	switch {
	case strings.HasPrefix(normalizedPath, "//"):
		prefix = `[\\/]{2}`
		normalizedPath = strings.TrimLeft(normalizedPath, "/")
	case strings.HasPrefix(normalizedPath, "/"):
		prefix = `[\\/]`
		normalizedPath = strings.TrimPrefix(normalizedPath, "/")
	}

	components := strings.Split(normalizedPath, "/")

	escapedComponents := make([]string, 0, len(components))
	for _, component := range components {
		if component == "" {
			continue
		}

		escapedComponents = append(escapedComponents, regexp.QuoteMeta(component))
	}

	return prefix + strings.Join(escapedComponents, `[\\/]`)
}
