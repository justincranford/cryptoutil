// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeaderToAnchor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{name: "simple h2", header: "## Document Organization", expected: "document-organization"},
		{name: "numbered section", header: "### 1.1 Vision Statement", expected: "11-vision-statement"},
		{name: "ampersand preserves double hyphen", header: "### 3.4 Port Assignments & Networking", expected: "34-port-assignments--networking"},
		{name: "dash in title keeps triple hyphen", header: "#### 11.2.8 format_go Self-Modification Protection - CRITICAL", expected: "1128-format_go-self-modification-protection---critical"},
		{name: "emoji stripped", header: "#### 🔐 Cryptographic Standards", expected: "cryptographic-standards"},
		{name: "parentheses stripped", header: "### 3.2.2 SM Instant Messenger (IM) Service", expected: "322-sm-instant-messenger-im-service"},
		{name: "h1 with special chars", header: "# cryptoutil Architecture - Single Source of Truth", expected: "cryptoutil-architecture---single-source-of-truth"},
		{name: "underscore preserved", header: "### format_go package", expected: "format_go-package"},
		{name: "trailing special chars stripped", header: "## Section ---", expected: "section"},
		{name: "empty after strip", header: "# 🔐", expected: ""},
		{name: "slash in header", header: "### OAuth 2.1 / OIDC 1.0", expected: "oauth-21--oidc-10"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := headerToAnchor(tc.header)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractAnchorsFromArchitecture(t *testing.T) {
	t.Parallel()

	content := `# Main Title
## 1. Executive Summary
### 1.1 Vision Statement
#### 1.1.1 Deep Subsection
## 2. Strategic Vision & Principles
`
	anchors := extractAnchorsFromArchitecture(content)

	require.True(t, anchors["main-title"])
	require.True(t, anchors["1-executive-summary"])
	require.True(t, anchors["11-vision-statement"])
	require.True(t, anchors["111-deep-subsection"])
	require.True(t, anchors["2-strategic-vision--principles"])
	require.False(t, anchors["nonexistent-section"])
}

func TestExtractRefsFromFile(t *testing.T) {
	t.Parallel()

	content := `# Instruction File

Some text here.

See [ARCHITECTURE.md Section 1.1](../../docs/ARCHITECTURE.md#11-vision-statement) for details.

More text.

See [ARCHITECTURE.md Section 6.4](../../docs/ARCHITECTURE.md#64-cryptographic-architecture) for crypto.

No ref on this line.
`
	refs := extractRefsFromFile("test-file.md", content)

	require.Len(t, refs, 2)
	require.Equal(t, "test-file.md", refs[0].SourceFile)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, refs[0].LineNumber)
	require.Equal(t, "11-vision-statement", refs[0].Anchor)
	require.Equal(t, "64-cryptographic-architecture", refs[1].Anchor)
}

func TestExtractRefsFromFile_MultipleOnSameLine(t *testing.T) {
	t.Parallel()

	content := `See [A](../../docs/ARCHITECTURE.md#11-vision) and [B](../../docs/ARCHITECTURE.md#12-key-chars).`

	refs := extractRefsFromFile("multi.md", content)

	require.Len(t, refs, 2)
	require.Equal(t, "11-vision", refs[0].Anchor)
	require.Equal(t, "12-key-chars", refs[1].Anchor)
}

func TestExtractRefsFromFile_NoRefs(t *testing.T) {
	t.Parallel()

	content := `# Just a plain file
No architecture references here.
`
	refs := extractRefsFromFile("plain.md", content)

	require.Empty(t, refs)
}

func TestTruncateRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "short line", input: "short", expected: "short"},
		{name: "exact cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes", input: strings.Repeat("a", cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes), expected: strings.Repeat("a", 120)},
		{name: "over cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes", input: strings.Repeat("b", 130), expected: strings.Repeat("b", 120) + "..."},
		{name: "with leading spaces", input: "  trimmed  ", expected: "trimmed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := truncateRef(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestValidatePropagation(t *testing.T) {
	t.Parallel()

	archContent := `# Main
## 1. Executive Summary
### 1.1 Vision Statement
### 1.2 Key Characteristics
## 2. Security & Principles
`
	instructionContent := `# Instructions
See [ARCHITECTURE.md Section 1.1](../../docs/ARCHITECTURE.md#11-vision-statement) for vision.
See [ARCHITECTURE.md Section 99.9](../../docs/ARCHITECTURE.md#99-nonexistent) broken.
`
	rootDir := t.TempDir()

	// Create directory structure.
	require.NoError(t, os.MkdirAll(rootDir+"/.github/instructions", 0o700))
	require.NoError(t, os.MkdirAll(rootDir+"/.github/agents", 0o700))
	require.NoError(t, os.MkdirAll(rootDir+"/docs", 0o700))
	require.NoError(t, os.WriteFile(rootDir+"/docs/ARCHITECTURE.md", []byte(archContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(rootDir+"/.github/instructions/test.instructions.md", []byte(instructionContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(rootDir+"/.github/copilot-instructions.md", []byte("No refs here."), cryptoutilSharedMagic.CacheFilePermissions))

	readFile := func(path string) ([]byte, error) {
		return os.ReadFile(rootDir + "/" + path)
	}

	result, err := ValidatePropagation(rootDir, readFile)
	require.NoError(t, err)

	// 1 valid ref (11-vision-statement), 1 broken (99-nonexistent).
	require.Len(t, result.ValidRefs, 1)
	require.Equal(t, "11-vision-statement", result.ValidRefs[0].Anchor)
	require.Len(t, result.BrokenRefs, 1)
	require.Equal(t, "99-nonexistent", result.BrokenRefs[0].Anchor)

	// Orphaned sections: 1-executive-summary, 12-key-characteristics, 2-security--principles (3 ##/### level headers not referenced).
	require.True(t, len(result.OrphanedKeys) > 0)

	// Coverage stats: 2 ## sections (1-executive-summary, 2-security--principles), 2 ### sections (11-vision-statement, 12-key-characteristics).
	// Only 11-vision-statement is referenced.
	require.Equal(t, 2, result.HighImpact.Total)
	require.Equal(t, 0, result.HighImpact.Referenced)
	require.Equal(t, 2, result.MediumImpact.Total)
	require.Equal(t, 1, result.MediumImpact.Referenced)
	require.Equal(t, 0, result.LowImpact.Total)
	require.Equal(t, 0, result.LowImpact.Referenced)
}

func TestValidatePropagation_MissingArchFile(t *testing.T) {
	t.Parallel()

	readFile := func(path string) ([]byte, error) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	result, err := ValidatePropagation(t.TempDir(), readFile)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "ARCHITECTURE.md")
}

func TestFormatPropagationResults_AllValid(t *testing.T) {
	t.Parallel()

	result := &PropagationResult{
		ValidRefs:    []PropagationRef{{Anchor: "test"}},
		BrokenRefs:   nil,
		OrphanedKeys: nil,
		TotalAnchors: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		HighImpact:   LevelCoverage{Total: 3, Referenced: 3},
		MediumImpact: LevelCoverage{Total: cryptoutilSharedMagic.GitRecentActivityDays, Referenced: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
		LowImpact:    LevelCoverage{Total: cryptoutilSharedMagic.JoseJADefaultMaxMaterials, Referenced: 2},
	}

	report := FormatPropagationResults(result)
	require.Contains(t, report, "1 valid refs, 0 broken refs")
	require.Contains(t, report, "All references resolve to valid ARCHITECTURE.md sections.")
	require.NotContains(t, report, "BROKEN")
	require.Contains(t, report, "SECTION COVERAGE:")
	require.Contains(t, report, "High   (##  ): 3/3 (100%)")
	require.Contains(t, report, "Medium (### ): 5/7 (71%)")
	require.Contains(t, report, "Low    (####): 2/10 (20%)")
	require.Contains(t, report, "Combined ##/###: 8/10 (80%)")
}

func TestFormatPropagationResults_WithBroken(t *testing.T) {
	t.Parallel()

	result := &PropagationResult{
		ValidRefs:    []PropagationRef{{Anchor: "valid"}},
		BrokenRefs:   []PropagationRef{{SourceFile: "test.md", LineNumber: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, Anchor: "broken"}},
		OrphanedKeys: []string{"orphan1"},
		TotalAnchors: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		HighImpact:   LevelCoverage{Total: 2, Referenced: 1},
		MediumImpact: LevelCoverage{Total: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, Referenced: 2},
	}

	report := FormatPropagationResults(result)
	require.Contains(t, report, "BROKEN REFERENCES (1)")
	require.Contains(t, report, "test.md:5 -> #broken")
	require.Contains(t, report, "ORPHANED SECTIONS (1 of 10")
	require.Contains(t, report, cryptoutilSharedMagic.TaskFailed)
	require.Contains(t, report, "High   (##  ): 1/2 (50%)")
	require.Contains(t, report, "Medium (### ): 2/5 (40%)")
	require.Contains(t, report, "Combined ##/###: 3/7 (42%)")
}

func TestFormatLevelCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		label    string
		lc       LevelCoverage
		expected string
	}{
		{name: "zero total", label: "Test", lc: LevelCoverage{Total: 0, Referenced: 0}, expected: "Test: 0/0 (N/A)\n"},
		{name: "full coverage", label: "High", lc: LevelCoverage{Total: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, Referenced: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries}, expected: "High: 5/5 (100%)\n"},
		{name: "partial coverage", label: "Med", lc: LevelCoverage{Total: cryptoutilSharedMagic.JoseJADefaultMaxMaterials, Referenced: 3}, expected: "Med: 3/10 (30%)\n"},
		{name: "no coverage", label: "Low", lc: LevelCoverage{Total: cryptoutilSharedMagic.IMMinPasswordLength, Referenced: 0}, expected: "Low: 0/8 (0%)\n"},
		{name: "integer truncation", label: "X", lc: LevelCoverage{Total: 3, Referenced: 1}, expected: "X: 1/3 (33%)\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := formatLevelCoverage(tc.label, tc.lc)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestValidatePropagationCommand_Integration(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := ValidatePropagationCommand(&stdout, &stderr)

	// Should succeed on the real project (0 broken refs).
	require.Equal(t, 0, exitCode, "validate-propagation should pass on real project: stdout=%s stderr=%s", stdout.String(), stderr.String())
	require.Contains(t, stdout.String(), "0 broken refs")
	require.Contains(t, stdout.String(), "All references resolve to valid ARCHITECTURE.md sections.")
	require.Contains(t, stdout.String(), "SECTION COVERAGE:")
	require.Contains(t, stdout.String(), "High   (##  ):")
	require.Contains(t, stdout.String(), "Combined ##/###:")
}

func TestValidatePropagationWithRoot_BadRoot(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := validatePropagationWithRoot("/nonexistent/path", &stdout, &stderr)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Error")
}
