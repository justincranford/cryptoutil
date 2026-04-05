// Copyright (c) 2025 Justin Cranford

package lint_gotest_hardcoded_uuid

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheck_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, []string{})

	require.NoError(t, err)
}

func TestCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
	}{
		{
			name: "no_violation_newv7",
			fileContent: `package test

import (
	"testing"
	googleUuid "github.com/google/uuid"
)

func TestSomething(t *testing.T) {
	t.Parallel()
	id := googleUuid.NewV7()
	_ = id
}
`,
			wantErr: false,
		},
		{
			name: "no_violation_nil_uuid",
			fileContent: `package test

import (
	"testing"
	googleUuid "github.com/google/uuid"
)

func TestNilUUID(t *testing.T) {
	t.Parallel()
	id := googleUuid.UUID{}
	_ = id
}
`,
			wantErr: false,
		},
		{
			name: "violation_must_parse",
			fileContent: `package test

import (
	"testing"
	googleUuid "github.com/google/uuid"
)

func TestHardcoded(t *testing.T) {
	t.Parallel()
	id := googleUuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	_ = id
}
`,
			wantErr: true,
		},
		{
			name: "violation_aliased_prefix",
			fileContent: `package test

import (
	"testing"
	uuid "github.com/google/uuid"
)

func TestAliased(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	_ = id
}
`,
			wantErr: true,
		},
		{
			name: "no_violation_parse_variable",
			fileContent: `package test

import (
	"testing"
	googleUuid "github.com/google/uuid"
)

func TestParseVar(t *testing.T) {
	t.Parallel()
	uuidStr := "some-string"
	id, _ := googleUuid.Parse(uuidStr)
	_ = id
}
`,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "example_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = Check(logger, []string{testFile})

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckHardcodedUUIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name: "must_parse_literal",
			fileContent: `googleUuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
`,
			wantIssues: true,
		},
		{
			name: "new_v7_no_issue",
			fileContent: `googleUuid.NewV7()
`,
			wantIssues: false,
		},
		{
			name: "parse_non_literal_no_issue",
			fileContent: `googleUuid.Parse(uuidStr)
`,
			wantIssues: false,
		},
		{
			name: "multiple_violations",
			fileContent: `googleUuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
someOther := "not uuid"
uuid.MustParse("00000000-0000-0000-0000-000000000001")
`,
			wantIssues: true,
		},
		{
			name:        "empty_file",
			fileContent: ``,
			wantIssues:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "example_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			issues := checkHardcodedUUIDs(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

func TestCheckHardcodedUUIDs_ReadError(t *testing.T) {
	t.Parallel()

	issues := checkHardcodedUUIDs("/nonexistent/path/test_test.go")

	require.NotEmpty(t, issues)
	require.Contains(t, issues[0], "Error reading file")
}
