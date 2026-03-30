// Copyright (c) 2025 Justin Cranford

// Package test_file_suffix_structure validates that Go test files follow the project's
// suffix naming convention: benchmark functions belong in _bench_test.go files,
// fuzz functions in _fuzz_test.go files, property tests in _property_test.go files,
// and integration tests in _integration_test.go files.
//
// Rules are loaded from test-file-suffix-rules.yaml. Two categories:
//
//   - suffix_rules: files matching a specific suffix must/must-not contain certain patterns.
//   - content_rules: files containing certain patterns must end with the correct suffix.
package test_file_suffix_structure

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

// SuffixRule defines required and forbidden content patterns for a specific file suffix.
type SuffixRule struct {
	Suffix                   string   `yaml:"suffix"`
	Description              string   `yaml:"description"`
	RequiredContentPatterns  []string `yaml:"required_content_patterns"`
	ForbiddenContentPatterns []string `yaml:"forbidden_content_patterns"`
	RequiredBuildTags        []string `yaml:"required_build_tags"`
}

// ContentRule defines a mapping from a content pattern to a required file suffix.
type ContentRule struct {
	ContentPattern string `yaml:"content_pattern"`
	RequiredSuffix string `yaml:"required_suffix"`
	Description    string `yaml:"description"`
}

// SuffixRules is the top-level structure of test-file-suffix-rules.yaml.
type SuffixRules struct {
	SuffixRules  []SuffixRule  `yaml:"suffix_rules"`
	ContentRules []ContentRule `yaml:"content_rules"`
}

// testFileSuffixReadFileFn is a seam for testing.
var testFileSuffixReadFileFn = os.ReadFile

// testFileSuffixWalkDirFn is a seam for testing.
var testFileSuffixWalkDirFn = filepath.WalkDir

// testFileSuffixGetwdFn is a seam for testing.
var testFileSuffixGetwdFn = os.Getwd

// findTestFileSuffixProjectRootFn is a seam for testing.
var findTestFileSuffixProjectRootFn = findTestFileSuffixProjectRoot

// findTestFileSuffixProjectRoot walks up from cwd to find the directory containing go.mod.
func findTestFileSuffixProjectRoot() (string, error) {
	dir, err := testFileSuffixGetwdFn()
	if err != nil {
		return "", fmt.Errorf("getwd failed: %w", err)
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}

		dir = parent
	}
}

// LoadSuffixRules loads and parses the test-file-suffix-rules.yaml manifest.
func LoadSuffixRules(rootDir string) (*SuffixRules, error) {
	path := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDTestFileSuffixRulesFile))

	data, err := testFileSuffixReadFileFn(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", cryptoutilSharedMagic.CICDTestFileSuffixRulesFile, err)
	}

	var rules SuffixRules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", cryptoutilSharedMagic.CICDTestFileSuffixRulesFile, err)
	}

	return &rules, nil
}

// compiledSuffixRule holds the pre-compiled regexes for a suffix rule.
type compiledSuffixRule struct {
	Suffix                  string
	Description             string
	RequiredContentRegexes  []*regexp.Regexp
	ForbiddenContentRegexes []*regexp.Regexp
	RequiredBuildTags       []string
}

// compiledContentRule holds the pre-compiled regex for a content rule.
type compiledContentRule struct {
	ContentRegex   *regexp.Regexp
	RequiredSuffix string
	Description    string
}

// compileRules compiles regexes in the loaded rules.
func compileRules(rules *SuffixRules) ([]compiledSuffixRule, []compiledContentRule, error) {
	suffixRules := make([]compiledSuffixRule, 0, len(rules.SuffixRules))

	for _, r := range rules.SuffixRules {
		compiled := compiledSuffixRule{
			Suffix:            r.Suffix,
			Description:       r.Description,
			RequiredBuildTags: r.RequiredBuildTags,
		}

		for _, pat := range r.RequiredContentPatterns {
			re, err := regexp.Compile(pat)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid required_content_pattern %q in rule %q: %w", pat, r.Suffix, err)
			}

			compiled.RequiredContentRegexes = append(compiled.RequiredContentRegexes, re)
		}

		for _, pat := range r.ForbiddenContentPatterns {
			re, err := regexp.Compile(pat)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid forbidden_content_pattern %q in rule %q: %w", pat, r.Suffix, err)
			}

			compiled.ForbiddenContentRegexes = append(compiled.ForbiddenContentRegexes, re)
		}

		suffixRules = append(suffixRules, compiled)
	}

	contentRules := make([]compiledContentRule, 0, len(rules.ContentRules))

	for _, r := range rules.ContentRules {
		re, err := regexp.Compile(r.ContentPattern)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid content_pattern %q in content rule: %w", r.ContentPattern, err)
		}

		contentRules = append(contentRules, compiledContentRule{
			ContentRegex:   re,
			RequiredSuffix: r.RequiredSuffix,
			Description:    r.Description,
		})
	}

	return suffixRules, contentRules, nil
}

// buildTagPattern matches //go:build lines.
var buildTagPattern = regexp.MustCompile(`^//go:build\s+(.+)$`)

// hasBuildTag checks whether the file content includes any of the required build tags.
func hasBuildTag(content []byte, required []string) bool {
	for _, line := range strings.Split(string(content), "\n") {
		m := buildTagPattern.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}

		tagExpr := m[1]
		for _, tag := range required {
			if strings.Contains(tagExpr, tag) {
				return true
			}
		}
	}

	return false
}

// checkFileAgainstSuffixRule validates that a file's content satisfies a suffix rule.
func checkFileAgainstSuffixRule(filePath string, content []byte, rule compiledSuffixRule) []string {
	var violations []string

	lines := strings.Split(string(content), "\n")

	// Check required build tags.
	if len(rule.RequiredBuildTags) > 0 && !hasBuildTag(content, rule.RequiredBuildTags) {
		violations = append(violations, fmt.Sprintf(
			"%s: missing required build tag (one of: %s) — %s",
			filePath, strings.Join(rule.RequiredBuildTags, " | "), rule.Description))
	}

	// Check required content (at least one pattern must match any line).
	for _, re := range rule.RequiredContentRegexes {
		found := false

		for _, line := range lines {
			if re.MatchString(line) {
				found = true

				break
			}
		}

		if !found {
			violations = append(violations, fmt.Sprintf(
				"%s: missing required pattern %q — %s",
				filePath, re.String(), rule.Description))
		}
	}

	// Check forbidden content (no pattern may match any line).
	for _, re := range rule.ForbiddenContentRegexes {
		for _, line := range lines {
			if re.MatchString(line) {
				violations = append(violations, fmt.Sprintf(
					"%s: contains forbidden pattern %q — %s",
					filePath, re.String(), rule.Description))

				break
			}
		}
	}

	return violations
}

// ExcludedDirs are directories skipped during the walk.
var ExcludedDirs = map[string]bool{
	cryptoutilSharedMagic.CICDExcludeDirVendor: true,
	cryptoutilSharedMagic.CICDExcludeDirGit:    true,
}

// isExcludedFromContentRules returns true for test files that deliberately contain
// benchmark or fuzz pattern strings as test fixture content (cicd tooling).
func isExcludedFromContentRules(filePath string) bool {
	normalised := filepath.ToSlash(filePath)

	return strings.Contains(normalised, "cicd_lint/format_gotest") ||
		strings.Contains(normalised, "cicd_lint/lint_fitness") ||
		strings.Contains(normalised, "cicd_lint/lint_gotest")
}

// CheckFiles validates all provided test files against the loaded suffix rules.
func CheckFiles(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string, rules *SuffixRules) error {
	logger.Log("Enforcing test file suffix structure rules...")

	suffixRules, contentRules, err := compileRules(rules)
	if err != nil {
		return fmt.Errorf("test-file-suffix-structure: %w", err)
	}

	var allViolations []string

	for _, filePath := range testFiles {
		content, readErr := testFileSuffixReadFileFn(filePath)
		if readErr != nil {
			return fmt.Errorf("test-file-suffix-structure: failed to read %s: %w", filePath, readErr)
		}

		// Check suffix-driven rules.
		for _, rule := range suffixRules {
			if strings.HasSuffix(filePath, rule.Suffix) {
				allViolations = append(allViolations, checkFileAgainstSuffixRule(filePath, content, rule)...)
			}
		}

		// Check content-driven rules (skip cicd test fixtures that deliberately contain special patterns).
		if !isExcludedFromContentRules(filePath) {
			lines := strings.Split(string(content), "\n")

			for _, rule := range contentRules {
				for _, line := range lines {
					if rule.ContentRegex.MatchString(line) && !strings.HasSuffix(filePath, rule.RequiredSuffix) {
						allViolations = append(allViolations, fmt.Sprintf(
							"%s: contains %q but does not end with %q — %s",
							filePath, rule.ContentRegex.String(), rule.RequiredSuffix, rule.Description))

						break
					}
				}
			}
		}
	}

	if len(allViolations) > 0 {
		for _, v := range allViolations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("test-file-suffix-structure: found %d violation(s)", len(allViolations))
	}

	logger.Log(fmt.Sprintf("test-file-suffix-structure: all %d test files pass suffix rules", len(testFiles)))

	return nil
}

// Check runs the linter by discovering all _test.go files in the repository.
// Returns an error if any violations are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	rootDir, err := findTestFileSuffixProjectRootFn()
	if err != nil {
		return fmt.Errorf("test-file-suffix-structure: %w", err)
	}

	return CheckInDir(logger, rootDir)
}

// CheckInDir runs the linter from a specified root directory.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	rules, err := LoadSuffixRules(rootDir)
	if err != nil {
		return fmt.Errorf("test-file-suffix-structure: %w", err)
	}

	var testFiles []string

	walkErr := testFileSuffixWalkDirFn(rootDir, func(path string, d fs.DirEntry, walkFileErr error) error {
		if walkFileErr != nil {
			return walkFileErr
		}

		if d.IsDir() {
			if ExcludedDirs[d.Name()] {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("test-file-suffix-structure: failed to walk %s: %w", rootDir, walkErr)
	}

	return CheckFiles(logger, testFiles, rules)
}
